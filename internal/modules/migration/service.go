package migration

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/config"
	"modl-go/internal/httpx"
	"modl-go/internal/models"
	"modl-go/internal/realtime"

	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	cooldownMs              = 60 * 60 * 1000
	staleMigrationMs        = 30 * 60 * 1000
	maxMessageLength        = 1000
	defaultFileSizeLimit    = 50 * 1024 * 1024
	maxConcurrentMigrations = 2
)

var (
	validTypes = map[string]bool{
		"litebans": true,
	}
	validStatuses = map[string]bool{
		"idle":            true,
		"building_json":   true,
		"uploading_json":  true,
		"processing_data": true,
		"completed":       true,
		"failed":          true,
	}

	errNoActiveMigration = errors.New("no active migration found")
	errInvalidStatus     = errors.New("invalid status value")
)

type progressJSON struct {
	Message          string `json:"message"`
	RecordsProcessed *int32 `json:"recordsProcessed,omitempty"`
	RecordsSkipped   *int32 `json:"recordsSkipped,omitempty"`
	TotalRecords     *int32 `json:"totalRecords,omitempty"`
	CooldownExempt   bool   `json:"cooldownExempt,omitempty"`
}

type migrationRow struct {
	ID          string
	TaskID      string
	Type        string
	Status      string
	Progress    progressJSON
	StartedAt   sql.NullInt64
	CompletedAt sql.NullInt64
	Error       string
}

type Service struct {
	db        *sql.DB
	cfg       *config.Config
	hub       *realtime.Hub
	uploadDir string
	sem       chan struct{}
}

func NewService(db *sql.DB, cfg *config.Config, hub *realtime.Hub) *Service {
	uploadDir := filepath.Join(os.TempDir(), "modl-migrations")
	if cfg != nil && strings.TrimSpace(cfg.StorageDir) != "" {
		uploadDir = filepath.Join(cfg.StorageDir, "migrations")
	}
	return &Service{
		db:        db,
		cfg:       cfg,
		hub:       hub,
		uploadDir: uploadDir,
		sem:       make(chan struct{}, maxConcurrentMigrations),
	}
}

func (s *Service) FileSizeLimit() int64 {
	return defaultFileSizeLimit
}

func (s *Service) failStaleMigrations(ctx context.Context, serverID string, nowMs int64) error {
	staleBefore := nowMs - staleMigrationMs
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, task_id, progress_json FROM migrations
		 WHERE server_id = ? AND status IN ('building_json', 'uploading_json', 'processing_data')
		   AND started_at IS NOT NULL AND started_at < ?`,
		serverID, staleBefore,
	)
	if err != nil {
		return err
	}
	type staleItem struct {
		id       string
		taskID   string
		progress progressJSON
	}
	var items []staleItem
	for rows.Next() {
		var id, taskID, pJSON string
		if err := rows.Scan(&id, &taskID, &pJSON); err != nil {
			rows.Close()
			return err
		}
		var p progressJSON
		_ = json.Unmarshal([]byte(pJSON), &p)
		items = append(items, staleItem{id: id, taskID: taskID, progress: p})
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	for _, it := range items {
		it.progress.Message = "Migration timed out and was automatically cancelled."
		pBytes, _ := json.Marshal(it.progress)
		_, err := s.db.ExecContext(ctx,
			`UPDATE migrations SET status = 'failed', error = ?, completed_at = ?, progress_json = ? WHERE id = ? AND server_id = ?`,
			"Migration timed out and was automatically cancelled.", nowMs, string(pBytes), it.id, serverID,
		)
		if err != nil {
			return err
		}
		s.publishStatusChange(serverID, it.taskID, "failed")
	}
	return nil
}

func (s *Service) scanMigrationRow(row *sql.Row) (*migrationRow, error) {
	var r migrationRow
	var pJSON string
	err := row.Scan(&r.ID, &r.TaskID, &r.Type, &r.Status, &pJSON, &r.StartedAt, &r.CompletedAt, &r.Error)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(pJSON) != "" {
		_ = json.Unmarshal([]byte(pJSON), &r.Progress)
	}
	return &r, nil
}

func (s *Service) GetLatestMigration(ctx context.Context, serverID string) (*migrationRow, error) {
	nowMs := time.Now().UnixMilli()
	if err := s.failStaleMigrations(ctx, serverID, nowMs); err != nil {
		return nil, err
	}
	row := s.db.QueryRowContext(ctx,
		`SELECT id, task_id, type, status, progress_json, started_at, completed_at, error
		 FROM migrations WHERE server_id = ? ORDER BY COALESCE(started_at, 0) DESC, rowid DESC LIMIT 1`,
		serverID,
	)
	m, err := s.scanMigrationRow(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *Service) findActiveMigration(ctx context.Context, serverID string) (*migrationRow, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, task_id, type, status, progress_json, started_at, completed_at, error
		 FROM migrations
		 WHERE server_id = ? AND status IN ('building_json', 'uploading_json', 'processing_data')
		 ORDER BY COALESCE(started_at, 0) DESC, rowid DESC LIMIT 1`,
		serverID,
	)
	m, err := s.scanMigrationRow(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return m, err
}

func (s *Service) IsActiveMigrationPresent(ctx context.Context, serverID string) bool {
	m, err := s.findActiveMigration(ctx, serverID)
	return err == nil && m != nil
}

func (s *Service) CheckCooldown(ctx context.Context, serverID string) (bool, *int64, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT progress_json, completed_at FROM migrations
		 WHERE server_id = ? AND status IN ('completed', 'failed') AND completed_at IS NOT NULL
		 ORDER BY completed_at DESC LIMIT 10`,
		serverID,
	)
	if err != nil {
		return false, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var pJSON string
		var completedAt int64
		if err := rows.Scan(&pJSON, &completedAt); err != nil {
			return false, nil, err
		}
		var p progressJSON
		_ = json.Unmarshal([]byte(pJSON), &p)
		if p.CooldownExempt {
			continue
		}
		elapsed := time.Now().UnixMilli() - completedAt
		if elapsed < cooldownMs {
			rem := cooldownMs - elapsed
			return true, &rem, nil
		}
		return false, nil, nil
	}
	return false, nil, rows.Err()
}

func (s *Service) GetStatus(ctx context.Context, srv *models.Server) (*modlv1.MigrationStatusResponse, error) {
	latest, err := s.GetLatestMigration(ctx, srv.ID)
	if err != nil {
		return nil, err
	}
	onCooldown, rem, err := s.CheckCooldown(ctx, srv.ID)
	if err != nil {
		return nil, err
	}

	resp := &modlv1.MigrationStatusResponse{
		Cooldown: &modlv1.MigrationStatusResponse_CooldownInfo{
			OnCooldown:    onCooldown,
			RemainingTime: rem,
		},
	}
	if latest != nil {
		cm := &modlv1.MigrationStatusResponse_CurrentMigration{
			TaskId: latest.TaskID,
			Type:   latest.Type,
			Status: latest.Status,
			Progress: &modlv1.MigrationStatusResponse_MigrationProgress{
				Message:          latest.Progress.Message,
				RecordsProcessed: latest.Progress.RecordsProcessed,
				RecordsSkipped:   latest.Progress.RecordsSkipped,
				TotalRecords:     latest.Progress.TotalRecords,
			},
			Error: latest.Error,
		}
		if latest.StartedAt.Valid && latest.StartedAt.Int64 > 0 {
			cm.StartedAt = timestamppb.New(time.UnixMilli(latest.StartedAt.Int64))
		}
		if latest.CompletedAt.Valid && latest.CompletedAt.Int64 > 0 {
			cm.CompletedAt = timestamppb.New(time.UnixMilli(latest.CompletedAt.Int64))
		}
		resp.CurrentMigration = cm
	}
	return resp, nil
}

func (s *Service) StartMigration(ctx context.Context, srv *models.Server, migrationType string) (*modlv1.MigrationOperationResponse, error) {
	mType := strings.ToLower(strings.TrimSpace(migrationType))
	if !validTypes[mType] {
		errMsg := "Invalid migration type"
		return &modlv1.MigrationOperationResponse{Success: false, Error: &errMsg}, nil
	}

	nowMs := time.Now().UnixMilli()
	if err := s.failStaleMigrations(ctx, srv.ID, nowMs); err != nil {
		return nil, err
	}
	if active, err := s.findActiveMigration(ctx, srv.ID); err != nil {
		return nil, err
	} else if active != nil {
		errMsg := "A migration is already in progress"
		return &modlv1.MigrationOperationResponse{Success: false, Error: &errMsg}, nil
	}

	onCooldown, _, err := s.CheckCooldown(ctx, srv.ID)
	if err != nil {
		return nil, err
	}
	if onCooldown {
		errMsg := "Migration on cooldown. Please wait before starting another migration."
		return &modlv1.MigrationOperationResponse{Success: false, Error: &errMsg}, nil
	}

	taskID := httpx.NewID()
	rowID := httpx.NewID()
	p := progressJSON{
		Message:          "Waiting for Minecraft server to build migration file...",
		RecordsProcessed: int32Ptr(0),
		RecordsSkipped:   int32Ptr(0),
	}
	pBytes, _ := json.Marshal(p)

	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO migrations (id, server_id, task_id, type, status, progress_json, started_at, error)
		 VALUES (?, ?, ?, ?, 'building_json', ?, ?, '')`,
		rowID, srv.ID, taskID, mType, string(pBytes), nowMs,
	); err != nil {
		return nil, err
	}

	if s.hub != nil {
		s.hub.PushMigrationTask(srv.ID, &modlv1.SyncMigrationTask{
			TaskId: taskID,
			Type:   mType,
		})
	}
	s.publishStatusChange(srv.ID, taskID, "building_json")

	msg := "Migration task initiated. Waiting for Minecraft server to process."
	return &modlv1.MigrationOperationResponse{
		Success: true,
		TaskId:  &taskID,
		Message: &msg,
	}, nil
}

func (s *Service) CancelMigration(ctx context.Context, srv *models.Server) (*modlv1.MigrationOperationResponse, error) {
	active, err := s.findActiveMigration(ctx, srv.ID)
	if err != nil {
		return nil, err
	}
	if active == nil {
		errMsg := "No active migration to cancel"
		return &modlv1.MigrationOperationResponse{Success: false, Error: &errMsg}, nil
	}

	nowMs := time.Now().UnixMilli()
	p := active.Progress
	p.Message = "Migration cancelled by administrator"
	p.CooldownExempt = active.Status == "building_json"
	pBytes, _ := json.Marshal(p)

	if _, err := s.db.ExecContext(ctx,
		`UPDATE migrations SET status = 'failed', error = ?, completed_at = ?, progress_json = ? WHERE id = ? AND server_id = ?`,
		"Cancelled by administrator", nowMs, string(pBytes), active.ID, srv.ID,
	); err != nil {
		return nil, err
	}
	s.publishStatusChange(srv.ID, active.TaskID, "failed")

	msg := "Migration cancelled successfully"
	return &modlv1.MigrationOperationResponse{
		Success: true,
		Message: &msg,
	}, nil
}

func (s *Service) UpdateProgress(ctx context.Context, srv *models.Server, status, message string, processed, skipped, total *int32) (*migrationRow, error) {
	st := strings.TrimSpace(status)
	if !validStatuses[st] {
		return nil, errInvalidStatus
	}
	active, err := s.findActiveMigration(ctx, srv.ID)
	if err != nil {
		return nil, err
	}
	if active == nil {
		return nil, errNoActiveMigration
	}

	p := active.Progress
	p.Message = truncate(message, maxMessageLength)
	if processed != nil {
		p.RecordsProcessed = processed
	}
	if skipped != nil {
		p.RecordsSkipped = skipped
	}
	if total != nil {
		p.TotalRecords = total
	}
	pBytes, _ := json.Marshal(p)

	var completedAt sql.NullInt64
	errCol := active.Error
	if st == "completed" || st == "failed" {
		completedAt = sql.NullInt64{Int64: time.Now().UnixMilli(), Valid: true}
		if st == "failed" && errCol == "" {
			errCol = p.Message
		}
	}

	if _, err := s.db.ExecContext(ctx,
		`UPDATE migrations SET status = ?, progress_json = ?, completed_at = ?, error = ? WHERE id = ? AND server_id = ?`,
		st, string(pBytes), completedAt, errCol, active.ID, srv.ID,
	); err != nil {
		return nil, err
	}

	s.publishStatusChange(srv.ID, active.TaskID, st)
	active.Status = st
	active.Progress = p
	return active, nil
}

func (s *Service) SaveUploadedFile(r io.Reader) (string, int64, error) {
	if err := os.MkdirAll(s.uploadDir, 0o755); err != nil {
		return "", 0, err
	}
	filename := fmt.Sprintf("migration-%s.json", httpx.NewID())
	fullPath := filepath.Join(s.uploadDir, filename)
	f, err := os.Create(fullPath)
	if err != nil {
		return "", 0, err
	}
	n, copyErr := io.Copy(f, r)
	closeErr := f.Close()
	if copyErr != nil {
		_ = os.Remove(fullPath)
		return "", 0, copyErr
	}
	if closeErr != nil {
		_ = os.Remove(fullPath)
		return "", 0, closeErr
	}
	return fullPath, n, nil
}

func (s *Service) TryAcquireProcessor() bool {
	select {
	case s.sem <- struct{}{}:
		return true
	default:
		return false
	}
}

func (s *Service) ReleaseProcessor() {
	select {
	case <-s.sem:
	default:
	}
}

func (s *Service) publishStatusChange(serverID, taskID, status string) {
	if s.hub == nil {
		return
	}
	s.hub.Publish(serverID, modlv1.Topic_TOPIC_PANEL_MIGRATIONS, &modlv1.RealtimeEnvelope{
		Payload: &modlv1.RealtimeEnvelope_MigrationStatusChanged{
			MigrationStatusChanged: &modlv1.MigrationStatusChangedEvent{
				MigrationId: taskID,
				Status:      &status,
				TenantId:    &serverID,
			},
		},
	})
}
