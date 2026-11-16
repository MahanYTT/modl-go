package migration

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/config"
	"modl-go/internal/httpx"
	"modl-go/internal/models"
	"modl-go/internal/realtime"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
	perms   models.PermissionChecker
}

func NewHandler(db *sql.DB, cfg *config.Config, hub *realtime.Hub, perms models.PermissionChecker) *Handler {
	return &Handler{
		service: NewService(db, cfg, hub),
		perms:   perms,
	}
}

func (h *Handler) Service() *Service {
	return h.service
}

func (h *Handler) RegisterPanelRoutes(r chi.Router) {
	r.Route("/migration", func(r chi.Router) {
		r.Use(httpx.RequirePanelPolicy(h.perms, models.AccessPolicy{
			Rule:   models.RuleRequirePermission,
			View:   "admin.settings.view.migration",
			Modify: "admin.settings.modify.migration",
		}))
		r.Get("/status", h.getStatus)
		r.Post("/start", h.startMigration)
		r.Post("/cancel", h.cancelMigration)
	})
}

func (h *Handler) RegisterMinecraftV1Routes(r chi.Router) {
	r.Route("/migration", func(r chi.Router) {
		r.Post("/upload", h.uploadMigrationFile)
		r.Post("/progress", h.updateProgress)
	})
}

func (h *Handler) getStatus(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Server context required")
		return
	}
	resp, err := h.service.GetStatus(r.Context(), srv)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load migration status")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, resp)
}

func (h *Handler) startMigration(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Server context required")
		return
	}
	var req modlv1.StartMigrationRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid start migration request")
		return
	}
	resp, err := h.service.StartMigration(r.Context(), srv, req.MigrationType)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to start migration")
		return
	}
	status := http.StatusOK
	if !resp.Success {
		status = http.StatusBadRequest
	}
	httpx.WriteProto(w, r, status, resp)
}

func (h *Handler) cancelMigration(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Server context required")
		return
	}
	resp, err := h.service.CancelMigration(r.Context(), srv)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to cancel migration")
		return
	}
	status := http.StatusOK
	if !resp.Success {
		status = http.StatusBadRequest
	}
	httpx.WriteProto(w, r, status, resp)
}

func (h *Handler) uploadMigrationFile(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Server context required")
		return
	}

	if !h.service.IsActiveMigrationPresent(r.Context(), srv.ID) {
		httpx.WriteError(w, r, http.StatusNotFound, "No active migration found for upload")
		return
	}

	limit := h.service.FileSizeLimit()
	r.Body = http.MaxBytesReader(w, r.Body, limit+1024*1024)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid multipart upload")
		return
	}

	file, header, err := r.FormFile("migrationFile")
	if err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "migrationFile is required")
		return
	}
	defer file.Close()

	if header.Size > limit {
		_, _ = h.service.UpdateProgress(r.Context(), srv, "failed", "Migration file exceeds size limit", int32Ptr(0), int32Ptr(0), nil)
		fileSizeMB := float64(header.Size) / (1024.0 * 1024.0)
		limitMB := float64(limit) / (1024.0 * 1024.0)
		httpx.WriteJSON(w, http.StatusRequestEntityTooLarge, map[string]any{
			"error":    "Migration file exceeds size limit",
			"message":  fmt.Sprintf("File size (%.2fMB) exceeds the limit of %.2fMB.", fileSizeMB, limitMB),
			"fileSize": header.Size,
			"limit":    limit,
		})
		return
	}

	if !h.service.TryAcquireProcessor() {
		httpx.WriteJSON(w, http.StatusServiceUnavailable, map[string]any{
			"error":   "Migration processing is busy",
			"message": "The server is processing other migrations. Please try again shortly.",
		})
		return
	}

	filePath, savedSize, err := h.service.SaveUploadedFile(file)
	if err != nil {
		h.service.ReleaseProcessor()
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to save migration file")
		return
	}

	if savedSize > limit {
		_ = os.Remove(filePath)
		h.service.ReleaseProcessor()
		_, _ = h.service.UpdateProgress(r.Context(), srv, "failed", "Migration file exceeds size limit", int32Ptr(0), int32Ptr(0), nil)
		fileSizeMB := float64(savedSize) / (1024.0 * 1024.0)
		limitMB := float64(limit) / (1024.0 * 1024.0)
		httpx.WriteJSON(w, http.StatusRequestEntityTooLarge, map[string]any{
			"error":    "Migration file exceeds size limit",
			"message":  fmt.Sprintf("File size (%.2fMB) exceeds the limit of %.2fMB.", fileSizeMB, limitMB),
			"fileSize": savedSize,
			"limit":    limit,
		})
		return
	}

	go func() {
		defer h.service.ReleaseProcessor()
		h.service.ProcessFile(context.Background(), srv, filePath)
	}()

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"success":  true,
		"message":  "Migration file uploaded successfully. Processing started.",
		"fileSize": savedSize,
	})
}

type updateProgressPayload struct {
	Status           string `json:"status"`
	Message          string `json:"message"`
	RecordsProcessed *int32 `json:"recordsProcessed"`
	RecordsSkipped   *int32 `json:"recordsSkipped"`
	TotalRecords     *int32 `json:"totalRecords"`
}

func (h *Handler) updateProgress(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Server context required")
		return
	}
	var payload updateProgressPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid progress payload")
		return
	}
	if payload.Status == "" || !validStatuses[payload.Status] {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid status value")
		return
	}
	if payload.Message == "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "Message is required")
		return
	}
	if payload.RecordsProcessed != nil && *payload.RecordsProcessed < 0 {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid recordsProcessed value")
		return
	}
	if payload.TotalRecords != nil && *payload.TotalRecords < 0 {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid totalRecords value")
		return
	}

	_, err := h.service.UpdateProgress(r.Context(), srv, payload.Status, payload.Message, payload.RecordsProcessed, payload.RecordsSkipped, payload.TotalRecords)
	if errors.Is(err, errNoActiveMigration) {
		httpx.WriteError(w, r, http.StatusNotFound, "No active migration found")
		return
	}
	if errors.Is(err, errInvalidStatus) {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid status value")
		return
	}
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to update migration progress")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"success": true})
}
