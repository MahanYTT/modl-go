package media

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"modl-go/internal/config"
	"modl-go/internal/httpx"
	"modl-go/internal/storage"

	"github.com/google/uuid"
)

var (
	ErrNotFound            = errors.New("not found")
	ErrForbidden           = errors.New("access denied")
	ErrInvalidInput        = errors.New("invalid input")
	ErrQuotaExceeded       = errors.New("storage quota exceeded")
	ErrTempLimitExceeded   = errors.New("temporary upload limit reached")
	ErrUploadNotFound      = errors.New("upload not found")
	ErrStorageNotConfigured = errors.New("file storage is not configured")
	ErrInvalidToken        = errors.New("invalid or expired upload token")
	ErrInvalidEvidenceURL  = errors.New("invalid evidence URL")
	ErrPunishmentNotFound  = errors.New("punishment not found")
)

const (
	presignUploadDuration  = 15 * time.Minute
	evidenceTokenTTL       = 30 * time.Minute
	storageOverageRate     = 0.08
)

type PresignResult struct {
	PresignedURL    string
	Key             string
	ExpiresAt       time.Time
	Method          string
	RequiredHeaders map[string]string
}

type UploadResult struct {
	ID          string
	Key         string
	URL         string
	FileName    string
	Size        int64
	ContentType string
}

type QuotaInfo struct {
	UsedBytes          int64
	MaxBytes           int64
	UsedPercentage     float64
	UsedFormatted      string
	MaxFormatted       string
	ByType             map[string]int64
	IsPremium          bool
	StorageOverageRate float64
}

type EvidenceToken struct {
	TokenHash    string
	ServerID     string
	PunishmentID string
	PlayerUUID   string
	IssuerName   string
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

type EvidenceTokenInfo struct {
	PunishmentID string
	PlayerName   string
	IssuerName   string
}

type EvidenceSubmitItem struct {
	URL      string
	FileName string
	FileType string
	FileSize int64
}

type pendingUpload struct {
	ServerID    string
	FileName    string
	ContentType string
	Size        int64
	ExpiresAt   time.Time
}

type Service struct {
	db                 *sql.DB
	store              *storage.LocalStore
	cfg                *config.Config
	maxStorageBytes    int64
	tempUploadCapBytes int64
	tempUploadWindow   time.Duration

	pendingMu sync.RWMutex
	pending   map[string]pendingUpload
}

func NewService(db *sql.DB, store *storage.LocalStore, cfg *config.Config) *Service {
	return &Service{
		db:                 db,
		store:              store,
		cfg:                cfg,
		maxStorageBytes:    DefaultFreeStorageBytes,
		tempUploadCapBytes: DefaultTempUploadCapBytes,
		tempUploadWindow:   6 * time.Hour,
		pending:            make(map[string]pendingUpload),
	}
}

func (s *Service) IsConfigured() bool {
	return s.store != nil
}

func (s *Service) SetMaxStorageBytes(maxBytes int64) {
	if maxBytes > 0 {
		s.maxStorageBytes = maxBytes
	}
}

func (s *Service) SetTempUploadCapBytes(capBytes int64) {
	if capBytes > 0 {
		s.tempUploadCapBytes = capBytes
	}
}

func (s *Service) PresignUpload(
	ctx context.Context,
	serverID, uploadType, fileName, contentType string,
	fileSize int64,
	entityID string,
	isPremium, enforceTempCap bool,
	uploadURLBuilder func(key string, exp int64, sig string) string,
) (*PresignResult, error) {
	if !s.IsConfigured() {
		return nil, ErrStorageNotConfigured
	}
	normalizedType := NormalizeUploadType(uploadType)
	if ok, msg := ValidateMetadata(fileName, contentType, fileSize, normalizedType, isPremium); !ok {
		return nil, fmt.Errorf("%w: %s", ErrInvalidInput, msg)
	}
	if enforceTempCap {
		exceeds, err := s.exceedsTempBudget(ctx, serverID, fileSize)
		if err != nil {
			return nil, err
		}
		if exceeds {
			return nil, ErrTempLimitExceeded
		}
	}
	can, err := s.canUpload(ctx, serverID, fileSize)
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, ErrQuotaExceeded
	}

	key := BuildKey(serverID, normalizedType, fileName, entityID)
	expiresAt := time.Now().UTC().Add(presignUploadDuration)
	expUnix := expiresAt.Unix()
	baseCT := normalizeContentType(contentType)
	sig := s.signPresign(key, expUnix, baseCT, fileSize)

	s.pendingMu.Lock()
	s.pending[key] = pendingUpload{
		ServerID:    serverID,
		FileName:    sanitizeFileName(fileName),
		ContentType: baseCT,
		Size:        fileSize,
		ExpiresAt:   expiresAt,
	}
	s.pendingMu.Unlock()

	return &PresignResult{
		PresignedURL: uploadURLBuilder(key, expUnix, sig),
		Key:          key,
		ExpiresAt:    expiresAt,
		Method:       http.MethodPut,
		RequiredHeaders: map[string]string{
			"Content-Type": baseCT,
		},
	}, nil
}

func (s *Service) HandlePresignedPut(key, expStr, sig, contentType string, maxDeclaredSize int64, body io.Reader) error {
	if !s.IsConfigured() {
		return ErrStorageNotConfigured
	}
	expUnix, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil || time.Now().UTC().Unix() > expUnix {
		return fmt.Errorf("%w: presigned URL expired", ErrForbidden)
	}

	// Verify signature against either the exact pending upload or signed metadata
	s.pendingMu.RLock()
	pend, hasPending := s.pending[key]
	s.pendingMu.RUnlock()

	maxBytes := DefaultMaxFileSize
	ct := normalizeContentType(contentType)
	if hasPending {
		if !hmac.Equal([]byte(sig), []byte(s.signPresign(key, expUnix, pend.ContentType, pend.Size))) {
			return fmt.Errorf("%w: invalid upload signature", ErrForbidden)
		}
		maxBytes = pend.Size
		if ct == "" {
			ct = pend.ContentType
		}
	} else {
		if maxDeclaredSize <= 0 {
			maxDeclaredSize = 100 * 1024 * 1024
		}
		if ct == "" {
			ct = "application/octet-stream"
		}
		if !hmac.Equal([]byte(sig), []byte(s.signPresign(key, expUnix, ct, maxDeclaredSize))) {
			return fmt.Errorf("%w: invalid upload signature", ErrForbidden)
		}
		maxBytes = maxDeclaredSize
	}

	if hasPathTraversal(key) {
		return ErrForbidden
	}

	diskPath := s.store.DiskPath(key)
	if err := os.MkdirAll(filepath.Dir(diskPath), 0o755); err != nil {
		return err
	}
	f, err := os.Create(diskPath)
	if err != nil {
		return err
	}
	n, copyErr := io.Copy(f, io.LimitReader(body, maxBytes+1))
	closeErr := f.Close()
	if copyErr != nil {
		_ = os.Remove(diskPath)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(diskPath)
		return closeErr
	}
	if n > maxBytes {
		_ = os.Remove(diskPath)
		return fmt.Errorf("%w: uploaded payload exceeds expected size", ErrInvalidInput)
	}

	s.pendingMu.Lock()
	s.pending[key] = pendingUpload{
		ServerID:    strings.Split(key, "/")[0],
		FileName:    ExtractFileName(key),
		ContentType: ct,
		Size:        n,
		ExpiresAt:   time.Unix(expUnix, 0).UTC(),
	}
	s.pendingMu.Unlock()
	return nil
}

func (s *Service) ConfirmUpload(
	ctx context.Context,
	serverID, key string,
	enforceTempCap bool,
	fileURLBuilder func(id string) string,
) (*UploadResult, error) {
	if !s.IsConfigured() {
		return nil, ErrStorageNotConfigured
	}
	if !IsKeyOwnedByServer(key, serverID) {
		return nil, ErrForbidden
	}

	if existing, err := s.store.GetByKey(ctx, serverID, key); err == nil && existing != nil {
		return &UploadResult{
			ID:          existing.ID,
			Key:         existing.Key,
			URL:         fileURLBuilder(existing.ID),
			FileName:    existing.FileName,
			Size:        existing.Size,
			ContentType: existing.ContentType,
		}, nil
	}

	diskPath := s.store.DiskPath(key)
	info, err := os.Stat(diskPath)
	if err != nil || info.IsDir() {
		return nil, ErrUploadNotFound
	}
	size := info.Size()

	s.pendingMu.RLock()
	pend, hasPending := s.pending[key]
	s.pendingMu.RUnlock()

	contentType := "application/octet-stream"
	if hasPending && pend.ContentType != "" {
		contentType = pend.ContentType
	} else if detected := detectFileContentType(diskPath); detected != "" {
		contentType = detected
	}

	if enforceTempCap {
		exceeds, err := s.exceedsTempBudget(ctx, serverID, size)
		if err != nil {
			_ = os.Remove(diskPath)
			return nil, err
		}
		if exceeds {
			_ = os.Remove(diskPath)
			return nil, ErrTempLimitExceeded
		}
	}

	can, err := s.canUpload(ctx, serverID, size)
	if err != nil {
		_ = os.Remove(diskPath)
		return nil, err
	}
	if !can {
		_ = os.Remove(diskPath)
		return nil, ErrQuotaExceeded
	}

	rec := &storage.FileRecord{
		ID:          httpx.NewID(),
		ServerID:    serverID,
		Key:         key,
		FileName:    ExtractFileName(key),
		Size:        size,
		ContentType: contentType,
		Category:    CategorizeFile(key),
		CreatedAt:   httpx.NowMs(),
	}
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO storage_files (id, server_id, key, file_name, size, content_type, category, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		rec.ID, rec.ServerID, rec.Key, rec.FileName, rec.Size, rec.ContentType, rec.Category, rec.CreatedAt,
	); err != nil {
		// Handle possible race if already inserted
		if existing, getErr := s.store.GetByKey(ctx, serverID, key); getErr == nil && existing != nil {
			return &UploadResult{
				ID:          existing.ID,
				Key:         existing.Key,
				URL:         fileURLBuilder(existing.ID),
				FileName:    existing.FileName,
				Size:        existing.Size,
				ContentType: existing.ContentType,
			}, nil
		}
		_ = os.Remove(diskPath)
		return nil, err
	}

	s.pendingMu.Lock()
	delete(s.pending, key)
	s.pendingMu.Unlock()

	return &UploadResult{
		ID:          rec.ID,
		Key:         rec.Key,
		URL:         fileURLBuilder(rec.ID),
		FileName:    rec.FileName,
		Size:        rec.Size,
		ContentType: rec.ContentType,
	}, nil
}

func (s *Service) DirectUpload(
	ctx context.Context,
	serverID, uploadType, fileName, contentType string,
	fileSize int64,
	entityID string,
	isPremium, enforceTempCap bool,
	r io.Reader,
	fileURLBuilder func(id string) string,
) (*UploadResult, error) {
	if !s.IsConfigured() {
		return nil, ErrStorageNotConfigured
	}
	normalizedType := NormalizeUploadType(uploadType)
	baseCT := normalizeContentType(contentType)
	if baseCT == "" {
		baseCT = "application/octet-stream"
	}
	if ok, msg := ValidateMetadata(fileName, baseCT, fileSize, normalizedType, isPremium); !ok {
		return nil, fmt.Errorf("%w: %s", ErrInvalidInput, msg)
	}
	if enforceTempCap {
		exceeds, err := s.exceedsTempBudget(ctx, serverID, fileSize)
		if err != nil {
			return nil, err
		}
		if exceeds {
			return nil, ErrTempLimitExceeded
		}
	}
	can, err := s.canUpload(ctx, serverID, fileSize)
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, ErrQuotaExceeded
	}

	key := BuildKey(serverID, normalizedType, fileName, entityID)
	diskPath := s.store.DiskPath(key)
	if err := os.MkdirAll(filepath.Dir(diskPath), 0o755); err != nil {
		return nil, err
	}
	maxAllowed := MaxSizeForType(normalizedType, isPremium)
	f, err := os.Create(diskPath)
	if err != nil {
		return nil, err
	}
	n, copyErr := io.Copy(f, io.LimitReader(r, maxAllowed+1))
	closeErr := f.Close()
	if copyErr != nil {
		_ = os.Remove(diskPath)
		return nil, copyErr
	}
	if closeErr != nil {
		_ = os.Remove(diskPath)
		return nil, closeErr
	}
	if n > maxAllowed || n <= 0 {
		_ = os.Remove(diskPath)
		return nil, fmt.Errorf("%w: invalid file size", ErrInvalidInput)
	}

	canAfter, err := s.canUpload(ctx, serverID, n)
	if err != nil || !canAfter {
		_ = os.Remove(diskPath)
		if err != nil {
			return nil, err
		}
		return nil, ErrQuotaExceeded
	}

	rec := &storage.FileRecord{
		ID:          httpx.NewID(),
		ServerID:    serverID,
		Key:         key,
		FileName:    ExtractFileName(key),
		Size:        n,
		ContentType: baseCT,
		Category:    CategorizeFile(key),
		CreatedAt:   httpx.NowMs(),
	}
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO storage_files (id, server_id, key, file_name, size, content_type, category, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		rec.ID, rec.ServerID, rec.Key, rec.FileName, rec.Size, rec.ContentType, rec.Category, rec.CreatedAt,
	); err != nil {
		_ = os.Remove(diskPath)
		return nil, err
	}

	return &UploadResult{
		ID:          rec.ID,
		Key:         rec.Key,
		URL:         fileURLBuilder(rec.ID),
		FileName:    rec.FileName,
		Size:        rec.Size,
		ContentType: rec.ContentType,
	}, nil
}

func (s *Service) DeleteByKey(ctx context.Context, serverID, key string) (bool, error) {
	if !IsKeyOwnedByServer(key, serverID) {
		return false, ErrForbidden
	}
	rec, err := s.store.GetByKey(ctx, serverID, key)
	if err == nil && rec != nil {
		if err := s.store.Delete(ctx, serverID, rec.ID); err != nil {
			return false, err
		}
		return true, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}
	// Check if orphan exists on disk
	if s.store != nil {
		diskPath := s.store.DiskPath(key)
		if _, statErr := os.Stat(diskPath); statErr == nil {
			if rmErr := os.Remove(diskPath); rmErr != nil {
				return false, rmErr
			}
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) BulkDelete(ctx context.Context, serverID string, keys []string) (int, error) {
	if len(keys) > StorageBulkDeleteMaxKeys {
		return 0, fmt.Errorf("%w: too many keys in bulk delete request", ErrInvalidInput)
	}
	for _, key := range keys {
		if !IsKeyOwnedByServer(key, serverID) {
			return 0, ErrForbidden
		}
	}
	deleted := 0
	for _, key := range keys {
		ok, err := s.DeleteByKey(ctx, serverID, key)
		if err != nil {
			return deleted, err
		}
		if ok {
			deleted++
		}
	}
	return deleted, nil
}

func (s *Service) ListFiles(ctx context.Context, serverID, prefix string) ([]storage.FileRecord, error) {
	fullPrefix := serverID + "/" + prefix
	escaped := escapeLike(fullPrefix) + "%"
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, server_id, key, file_name, size, content_type, category, created_at
		 FROM storage_files
		 WHERE server_id = ? AND key LIKE ? ESCAPE '\'
		 ORDER BY created_at DESC
		 LIMIT 1000`,
		serverID, escaped,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []storage.FileRecord
	for rows.Next() {
		var rec storage.FileRecord
		if err := rows.Scan(&rec.ID, &rec.ServerID, &rec.Key, &rec.FileName, &rec.Size, &rec.ContentType, &rec.Category, &rec.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

func (s *Service) GetDownloadRecord(ctx context.Context, serverID, key string) (*storage.FileRecord, error) {
	if !IsKeyOwnedByServer(key, serverID) {
		return nil, ErrForbidden
	}
	rec, err := s.store.GetByKey(ctx, serverID, key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if _, statErr := os.Stat(s.store.DiskPath(rec.Key)); statErr != nil {
		return nil, ErrNotFound
	}
	return rec, nil
}

func (s *Service) GetFileByID(ctx context.Context, serverID, id string) (*storage.FileRecord, string, error) {
	if !s.IsConfigured() {
		return nil, "", ErrNotFound
	}
	var row *sql.Row
	if serverID != "" {
		row = s.db.QueryRowContext(ctx,
			`SELECT id, server_id, key, file_name, size, content_type, category, created_at
			 FROM storage_files WHERE server_id = ? AND id = ?`,
			serverID, id,
		)
	} else {
		row = s.db.QueryRowContext(ctx,
			`SELECT id, server_id, key, file_name, size, content_type, category, created_at
			 FROM storage_files WHERE id = ?`,
			id,
		)
	}
	var rec storage.FileRecord
	if err := row.Scan(&rec.ID, &rec.ServerID, &rec.Key, &rec.FileName, &rec.Size, &rec.ContentType, &rec.Category, &rec.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", ErrNotFound
		}
		return nil, "", err
	}
	diskPath := s.store.DiskPath(rec.Key)
	if info, err := os.Stat(diskPath); err != nil || info.IsDir() {
		return nil, "", ErrNotFound
	}
	return &rec, diskPath, nil
}

func (s *Service) GetEvidenceFileByID(ctx context.Context, id string) (*storage.FileRecord, string, error) {
	if !s.IsConfigured() {
		return nil, "", ErrNotFound
	}
	row := s.db.QueryRowContext(ctx,
		`SELECT id, server_id, key, file_name, size, content_type, category, created_at
		 FROM storage_files WHERE id = ? AND (category = 'evidence' OR key LIKE '%/evidence/%')`,
		id,
	)
	var rec storage.FileRecord
	if err := row.Scan(&rec.ID, &rec.ServerID, &rec.Key, &rec.FileName, &rec.Size, &rec.ContentType, &rec.Category, &rec.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", ErrNotFound
		}
		return nil, "", err
	}
	diskPath := s.store.DiskPath(rec.Key)
	if info, err := os.Stat(diskPath); err != nil || info.IsDir() {
		return nil, "", ErrNotFound
	}
	return &rec, diskPath, nil
}

func (s *Service) SyncFiles(ctx context.Context, serverID string) (int, error) {
	if !s.IsConfigured() {
		return 0, nil
	}
	rootDir := s.store.DiskPath(serverID)
	diskFiles := make(map[string]os.FileInfo)
	if _, err := os.Stat(rootDir); err == nil {
		baseRoot := filepath.Clean(s.store.DiskPath(""))
		err = filepath.Walk(rootDir, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if info.IsDir() {
				return nil
			}
			rel, relErr := filepath.Rel(baseRoot, path)
			if relErr != nil {
				return relErr
			}
			key := filepath.ToSlash(rel)
			diskFiles[key] = info
			return nil
		})
		if err != nil {
			return 0, err
		}
	}

	existing, err := s.store.List(ctx, serverID, "")
	if err != nil {
		return 0, err
	}
	existingByKey := make(map[string]storage.FileRecord, len(existing))
	for _, rec := range existing {
		existingByKey[rec.Key] = rec
	}

	for key, info := range diskFiles {
		if rec, ok := existingByKey[key]; ok {
			if rec.Size != info.Size() {
				_, _ = s.db.ExecContext(ctx, `UPDATE storage_files SET size = ? WHERE id = ?`, info.Size(), rec.ID)
			}
			continue
		}
		ct := detectFileContentType(s.store.DiskPath(key))
		if ct == "" {
			ct = "application/octet-stream"
		}
		_, err := s.db.ExecContext(ctx,
			`INSERT INTO storage_files (id, server_id, key, file_name, size, content_type, category, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			httpx.NewID(), serverID, key, ExtractFileName(key), info.Size(), ct, CategorizeFile(key), info.ModTime().UnixMilli(),
		)
		if err != nil {
			return 0, err
		}
	}

	for _, rec := range existing {
		if _, ok := diskFiles[rec.Key]; !ok {
			_, _ = s.db.ExecContext(ctx, `DELETE FROM storage_files WHERE server_id = ? AND id = ?`, serverID, rec.ID)
		}
	}

	return len(diskFiles), nil
}

func (s *Service) GetQuota(ctx context.Context, serverID string) (*QuotaInfo, error) {
	byType := map[string]int64{
		"ticket":   0,
		"evidence": 0,
		"logs":     0,
		"backup":   0,
		"replay":   0,
		"other":    0,
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT category, COALESCE(SUM(size), 0) FROM storage_files WHERE server_id = ? GROUP BY category`,
		serverID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usedBytes int64
	for rows.Next() {
		var cat string
		var sz int64
		if err := rows.Scan(&cat, &sz); err != nil {
			return nil, err
		}
		if _, known := byType[cat]; !known {
			byType["other"] += sz
		} else {
			byType[cat] = sz
		}
		usedBytes += sz
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	maxBytes := s.maxStorageBytes
	var pct float64
	if maxBytes > 0 {
		pct = float64(int64((float64(usedBytes)/float64(maxBytes)*100)*100+0.5)) / 100.0
	}

	return &QuotaInfo{
		UsedBytes:          usedBytes,
		MaxBytes:           maxBytes,
		UsedPercentage:     pct,
		UsedFormatted:      FormatBytes(usedBytes),
		MaxFormatted:       FormatBytes(maxBytes),
		ByType:             byType,
		IsPremium:          false,
		StorageOverageRate: storageOverageRate,
	}, nil
}

// ValidatePublicUploadAccess checks whether a public upload for ticket/appeal is authorized.
func (s *Service) ValidatePublicUploadAccess(ctx context.Context, serverID, uploadType, entityID, accessToken string) (int, string) {
	if strings.TrimSpace(entityID) == "" {
		return http.StatusForbidden, "entityId is required for public uploads"
	}
	normalizedEntityID := strings.TrimSpace(entityID)
	normalizedType := NormalizeUploadType(uploadType)

	if IsAnonymousTempEntity(normalizedEntityID) {
		if normalizedType == "ticket" || normalizedType == "appeal" {
			return http.StatusOK, ""
		}
		return http.StatusForbidden, "Temporary uploads are only allowed for ticket and appeal types"
	}

	var ticketType string
	var hidden, emailAuthEnabled int
	err := s.db.QueryRowContext(ctx,
		`SELECT type, hidden, email_auth_enabled FROM tickets WHERE server_id = ? AND id = ?`,
		serverID, normalizedEntityID,
	).Scan(&ticketType, &hidden, &emailAuthEnabled)
	if errors.Is(err, sql.ErrNoRows) || hidden != 0 {
		return http.StatusNotFound, "Not found"
	}
	if err != nil {
		return http.StatusInternalServerError, "Failed to validate ticket access"
	}

	isAppeal := strings.EqualFold(ticketType, "APPEAL")
	if normalizedType == "appeal" && !isAppeal {
		return http.StatusForbidden, "Entity is not an appeal ticket"
	}
	if normalizedType == "ticket" && isAppeal {
		return http.StatusForbidden, "Appeal uploads must use uploadType=appeal"
	}

	if emailAuthEnabled != 0 {
		if strings.TrimSpace(accessToken) == "" {
			return http.StatusForbidden, "Email verification token required for this ticket"
		}
		var exists int
		err := s.db.QueryRowContext(ctx,
			`SELECT 1 FROM ticket_verifications WHERE server_id = ? AND ticket_id = ? AND token = ? AND expires_at > ? LIMIT 1`,
			serverID, normalizedEntityID, accessToken, httpx.NowMs(),
		).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			return http.StatusForbidden, "Email verification token required for this ticket"
		}
		if err != nil {
			return http.StatusInternalServerError, "Failed to validate ticket token"
		}
	}

	return http.StatusOK, ""
}

// Evidence upload token methods

func (s *Service) CreateEvidenceUploadToken(ctx context.Context, serverID, punishmentID, playerUUID, issuerName string) (string, error) {
	rawToken := uuid.NewString()
	hash := sha256Hex(rawToken)
	now := time.Now().UTC()
	exp := now.Add(evidenceTokenTTL)

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO evidence_upload_tokens (token_hash, server_id, punishment_id, player_uuid, issuer_name, created_at, expires_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		hash, serverID, punishmentID, playerUUID, issuerName, now.UnixMilli(), exp.UnixMilli(),
	)
	if err != nil {
		return "", err
	}
	return rawToken, nil
}

func (s *Service) ValidateEvidenceToken(ctx context.Context, rawToken string) (*EvidenceToken, error) {
	if strings.TrimSpace(rawToken) == "" {
		return nil, ErrInvalidToken
	}
	hash := sha256Hex(rawToken)
	var tok EvidenceToken
	var createdMs, expiresMs int64
	err := s.db.QueryRowContext(ctx,
		`SELECT token_hash, server_id, punishment_id, player_uuid, issuer_name, created_at, expires_at
		 FROM evidence_upload_tokens WHERE token_hash = ?`,
		hash,
	).Scan(&tok.TokenHash, &tok.ServerID, &tok.PunishmentID, &tok.PlayerUUID, &tok.IssuerName, &createdMs, &expiresMs)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalidToken
	}
	if err != nil {
		return nil, err
	}
	tok.CreatedAt = time.UnixMilli(createdMs).UTC()
	tok.ExpiresAt = time.UnixMilli(expiresMs).UTC()
	if time.Now().UTC().After(tok.ExpiresAt) {
		_, _ = s.db.ExecContext(ctx, `DELETE FROM evidence_upload_tokens WHERE token_hash = ?`, hash)
		return nil, ErrInvalidToken
	}
	return &tok, nil
}

func (s *Service) GetEvidenceTokenInfo(ctx context.Context, rawToken string) (*EvidenceTokenInfo, error) {
	tok, err := s.ValidateEvidenceToken(ctx, rawToken)
	if err != nil {
		return nil, err
	}
	playerName := s.lookupPlayerName(ctx, tok.ServerID, tok.PlayerUUID)
	return &EvidenceTokenInfo{
		PunishmentID: tok.PunishmentID,
		PlayerName:   playerName,
		IssuerName:   tok.IssuerName,
	}, nil
}

func (s *Service) PresignEvidenceUpload(
	ctx context.Context,
	rawToken, fileName, contentType string,
	fileSize int64,
	uploadURLBuilder func(key string, exp int64, sig string) string,
) (*PresignResult, error) {
	tok, err := s.ValidateEvidenceToken(ctx, rawToken)
	if err != nil {
		return nil, err
	}
	return s.PresignUpload(ctx, tok.ServerID, "evidence", fileName, contentType, fileSize, tok.PunishmentID, false, false, uploadURLBuilder)
}

func (s *Service) ConfirmEvidenceUpload(
	ctx context.Context,
	rawToken, key string,
	fileURLBuilder func(id string) string,
) (*UploadResult, error) {
	tok, err := s.ValidateEvidenceToken(ctx, rawToken)
	if err != nil {
		return nil, err
	}
	expectedPrefix := tok.ServerID + "/evidence/" + tok.PunishmentID + "/"
	if !IsKeyOwnedByServer(key, tok.ServerID) || !strings.HasPrefix(key, expectedPrefix) {
		return nil, ErrForbidden
	}
	return s.ConfirmUpload(ctx, tok.ServerID, key, false, fileURLBuilder)
}

func (s *Service) DirectEvidenceUpload(
	ctx context.Context,
	rawToken, fileName, contentType string,
	fileSize int64,
	r io.Reader,
	fileURLBuilder func(id string) string,
) (*UploadResult, error) {
	tok, err := s.ValidateEvidenceToken(ctx, rawToken)
	if err != nil {
		return nil, err
	}
	return s.DirectUpload(ctx, tok.ServerID, "evidence", fileName, contentType, fileSize, tok.PunishmentID, false, false, r, fileURLBuilder)
}

func (s *Service) SubmitEvidence(ctx context.Context, rawToken string, items []EvidenceSubmitItem) error {
	tok, err := s.ValidateEvidenceToken(ctx, rawToken)
	if err != nil {
		return err
	}
	if len(items) == 0 || len(items) > EvidenceMaxItems {
		return fmt.Errorf("%w: invalid evidence item count", ErrInvalidInput)
	}

	expectedPrefix := tok.ServerID + "/evidence/" + tok.PunishmentID + "/"
	resolvedDocs := make([]*storage.FileRecord, 0, len(items))
	for _, item := range items {
		rec, err := s.resolveEvidenceRecord(ctx, tok.ServerID, item.URL)
		if err != nil || rec == nil || !strings.HasPrefix(rec.Key, expectedPrefix) {
			return ErrInvalidEvidenceURL
		}
		resolvedDocs = append(resolvedDocs, rec)
	}

	var evidenceJSON, notesJSON string
	err = s.db.QueryRowContext(ctx,
		`SELECT evidence_json, notes_json FROM punishments WHERE server_id = ? AND id = ?`,
		tok.ServerID, tok.PunishmentID,
	).Scan(&evidenceJSON, &notesJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrPunishmentNotFound
	}
	if err != nil {
		return err
	}

	var evidenceList []map[string]any
	if strings.TrimSpace(evidenceJSON) != "" {
		_ = json.Unmarshal([]byte(evidenceJSON), &evidenceList)
	}
	var notesList []map[string]any
	if strings.TrimSpace(notesJSON) != "" {
		_ = json.Unmarshal([]byte(notesJSON), &notesList)
	}

	now := time.Now().UTC()
	nowMs := now.UnixMilli()
	issuer := tok.IssuerName
	if issuer == "" {
		issuer = "System"
	}

	for i, item := range items {
		doc := resolvedDocs[i]
		evidenceList = append(evidenceList, map[string]any{
			"text":         nil,
			"url":          item.URL,
			"type":         "file",
			"uploadedBy":   issuer,
			"uploadedById": nil,
			"uploadedAt":   now,
			"fileName":     item.FileName,
			"fileType":     doc.ContentType,
			"fileSize":     doc.Size,
		})
	}
	notesList = append(notesList, map[string]any{
		"id":         httpx.NewID(),
		"text":       fmt.Sprintf("uploaded %d evidence file(s)", len(items)),
		"date":       now,
		"issuerName": issuer,
		"issuerId":   nil,
	})

	newEvidenceBytes, err := json.Marshal(evidenceList)
	if err != nil {
		return err
	}
	newNotesBytes, err := json.Marshal(notesList)
	if err != nil {
		return err
	}

	res, err := s.db.ExecContext(ctx,
		`UPDATE punishments SET evidence_json = ?, notes_json = ?, updated_at = ? WHERE server_id = ? AND id = ?`,
		string(newEvidenceBytes), string(newNotesBytes), nowMs, tok.ServerID, tok.PunishmentID,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrPunishmentNotFound
	}

	_, _ = s.db.ExecContext(ctx, `DELETE FROM evidence_upload_tokens WHERE token_hash = ?`, tok.TokenHash)
	return nil
}

func (s *Service) resolveEvidenceRecord(ctx context.Context, serverID, rawURL string) (*storage.FileRecord, error) {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return nil, ErrInvalidEvidenceURL
	}
	path := trimmed
	if strings.Contains(trimmed, "://") {
		u, err := url.Parse(trimmed)
		if err != nil || (!strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, "https")) {
			return nil, ErrInvalidEvidenceURL
		}
		path = u.Path
	}

	for _, prefix := range []string{"/v1/public/evidence-upload/file/", "/v1/public/media/file/"} {
		if idx := strings.Index(path, prefix); idx >= 0 {
			fileID := strings.Trim(path[idx+len(prefix):], "/")
			if fileID != "" {
				rec, err := s.store.GetByID(ctx, serverID, fileID)
				if err == nil && rec != nil {
					return rec, nil
				}
			}
		}
	}

	key := strings.TrimPrefix(path, "/")
	if key != "" {
		rec, err := s.store.GetByKey(ctx, serverID, key)
		if err == nil && rec != nil {
			return rec, nil
		}
	}
	return nil, ErrInvalidEvidenceURL
}

func (s *Service) lookupPlayerName(ctx context.Context, serverID, playerUUID string) string {
	normalized := normalizeUUID(playerUUID)
	undashed := strings.ReplaceAll(normalized, "-", "")
	var currentUsername, usernamesJSON string
	err := s.db.QueryRowContext(ctx,
		`SELECT current_username, usernames_json FROM players
		 WHERE server_id = ? AND (LOWER(minecraft_uuid) = ? OR LOWER(REPLACE(minecraft_uuid, '-', '')) = ?)
		 LIMIT 1`,
		serverID, normalized, undashed,
	).Scan(&currentUsername, &usernamesJSON)
	if err != nil {
		return "Unknown"
	}
	if strings.TrimSpace(currentUsername) != "" {
		return currentUsername
	}
	var entries []map[string]any
	if err := json.Unmarshal([]byte(usernamesJSON), &entries); err == nil && len(entries) > 0 {
		for _, key := range []string{"username", "name"} {
			if v, ok := entries[len(entries)-1][key].(string); ok && strings.TrimSpace(v) != "" {
				return v
			}
		}
	}
	var strEntries []string
	if err := json.Unmarshal([]byte(usernamesJSON), &strEntries); err == nil && len(strEntries) > 0 {
		if strings.TrimSpace(strEntries[len(strEntries)-1]) != "" {
			return strEntries[len(strEntries)-1]
		}
	}
	return "Unknown"
}

func (s *Service) canUpload(ctx context.Context, serverID string, fileSize int64) (bool, error) {
	if fileSize < 0 {
		return false, nil
	}
	used, _, err := s.store.TotalSize(ctx, serverID)
	if err != nil {
		return false, err
	}
	return used+fileSize <= s.maxStorageBytes, nil
}

func (s *Service) exceedsTempBudget(ctx context.Context, serverID string, additionalBytes int64) (bool, error) {
	if additionalBytes < 0 {
		return true, nil
	}
	cutoffMs := time.Now().UTC().Add(-s.tempUploadWindow).UnixMilli()
	prefixes := TempUploadPrefixes(serverID)
	var used int64
	err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(size), 0) FROM storage_files
		 WHERE server_id = ? AND created_at >= ?
		   AND (key LIKE ? ESCAPE '\' OR key LIKE ? ESCAPE '\')`,
		serverID, cutoffMs, escapeLike(prefixes[0])+"%", escapeLike(prefixes[1])+"%",
	).Scan(&used)
	if err != nil {
		return false, err
	}
	return used+additionalBytes > s.tempUploadCapBytes, nil
}

func (s *Service) signPresign(key string, expUnix int64, contentType string, fileSize int64) string {
	secret := "local-dev-secret"
	if s.cfg != nil && s.cfg.AuthCodeHashSecret != "" {
		secret = s.cfg.AuthCodeHashSecret
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%s|%d|%s|%d", key, expUnix, contentType, fileSize)))
	return hex.EncodeToString(mac.Sum(nil))
}

func sha256Hex(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func hasPathTraversal(key string) bool {
	for _, seg := range strings.Split(key, "/") {
		if seg == ".." {
			return true
		}
	}
	return false
}

func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}

func normalizeUUID(u string) string {
	s := strings.ToLower(strings.TrimSpace(u))
	if len(s) == 32 && !strings.Contains(s, "-") {
		return fmt.Sprintf("%s-%s-%s-%s-%s", s[0:8], s[8:12], s[12:16], s[16:20], s[20:32])
	}
	return s
}

func detectFileContentType(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	if n > 0 {
		return normalizeContentType(http.DetectContentType(buf[:n]))
	}
	return ""
}
