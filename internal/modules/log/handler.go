package log

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/httpx"
	"modl-go/internal/models"

	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	defaultLogLimit = 100
	maxLogLimit     = 500
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) LogAction(ctx context.Context, serverID, description, level, source string) error {
	if serverID == "" {
		return nil
	}
	id := httpx.NewID()
	now := time.Now().UnixMilli()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO server_logs (id, server_id, description, level, source, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		id, serverID, description, level, source, now,
	)
	return err
}

func (s *Service) GetLogs(ctx context.Context, serverID string, limit int) ([]*modlv1.PanelLogResponse, error) {
	if limit < 1 {
		limit = 1
	} else if limit > maxLogLimit {
		limit = maxLogLimit
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, description, level, source, created_at
		 FROM server_logs
		 WHERE server_id = ?
		 ORDER BY created_at DESC
		 LIMIT ?`,
		serverID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]*modlv1.PanelLogResponse, 0)
	for rows.Next() {
		var id, desc, level, source string
		var createdAt int64
		if err := rows.Scan(&id, &desc, &level, &source, &createdAt); err != nil {
			return nil, err
		}
		logs = append(logs, &modlv1.PanelLogResponse{
			Id:          id,
			Description: desc,
			Level:       level,
			Source:      source,
			Created:     timestamppb.New(time.UnixMilli(createdAt)),
		})
	}
	return logs, rows.Err()
}

type Handler struct {
	service *Service
	perms   models.PermissionChecker
}

func NewHandler(db *sql.DB, perms models.PermissionChecker) *Handler {
	return &Handler{
		service: NewService(db),
		perms:   perms,
	}
}

func (h *Handler) Service() *Service {
	return h.service
}

func (h *Handler) RegisterPanelRoutes(r chi.Router) {
	r.Route("/logs", func(r chi.Router) {
		r.Use(httpx.RequirePanelPolicy(h.perms, models.AccessPolicy{
			Rule:   models.RuleRequirePermission,
			View:   "admin.audit.view.logs",
			Modify: "admin.audit.view.logs",
		}))
		r.Get("/", h.getLogs)
	})
}

func (h *Handler) getLogs(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Server context required")
		return
	}
	limit := httpx.QueryInt(r, "limit", defaultLogLimit, 1, maxLogLimit)
	logs, err := h.service.GetLogs(r.Context(), srv.ID, limit)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load logs")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PanelLogsResponse{Logs: logs})
}
