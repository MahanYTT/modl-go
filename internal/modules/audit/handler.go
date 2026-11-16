package audit

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/httpx"
	"modl-go/internal/models"
	"modl-go/internal/realtime"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
	hub     *realtime.Hub
	perms   models.PermissionChecker
}

func NewHandler(db *sql.DB, hub *realtime.Hub, perms models.PermissionChecker) *Handler {
	return &Handler{
		service: NewService(db),
		hub:     hub,
		perms:   perms,
	}
}

func (h *Handler) Service() *Service {
	return h.service
}

func (h *Handler) RegisterPanelRoutes(r chi.Router) {
	r.Route("/audit", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(httpx.RequirePanelPolicy(h.perms, models.AccessPolicy{
				Rule:   models.RuleRequirePermission,
				View:   "admin.audit.view.logs",
				Modify: "admin.audit.view.logs",
			}))
			r.Get("/staff-performance", h.getStaffPerformance)
			r.Get("/staff/{username}/details", h.getStaffDetails)
			r.Get("/punishments/active", h.getActivePunishments)
			r.Get("/punishments", h.getPunishments)
		})

		r.Group(func(r chi.Router) {
			r.Use(httpx.RequirePanelPolicy(h.perms, models.AccessPolicy{
				Rule: models.RuleSuperAdmin,
			}))
			r.Get("/database/{table}", h.getDatabaseTable)
		})

		r.Group(func(r chi.Router) {
			r.Use(httpx.RequirePanelPolicy(h.perms, models.AccessPolicy{
				Rule:                  models.RuleSuperAdmin,
				SupersedesPermissions: []string{"admin.audit.rollback"},
			}))
			r.Post("/punishments/{id}/rollback", h.rollbackPunishment)
			r.Post("/staff/{username}/rollback-all", h.rollbackAllByStaff)
			r.Post("/staff/{username}/rollback-date-range", h.rollbackByDateRange)
			r.Post("/punishments/bulk-pardon", h.bulkPardon)
			r.Post("/punishments/bulk-set-expiration", h.bulkSetExpiration)
		})
	})
}

func (h *Handler) invalidateAudit(serverID string) {
	if h.hub == nil {
		return
	}
	h.hub.InvalidatePanel(serverID, modlv1.PanelResource_PANEL_RESOURCE_AUDIT, "")
	h.hub.InvalidatePanel(serverID, modlv1.PanelResource_PANEL_RESOURCE_PUNISHMENTS, "")
	h.hub.InvalidatePanel(serverID, modlv1.PanelResource_PANEL_RESOURCE_DASHBOARD, "")
}

func (h *Handler) currentPerformer(r *http.Request) string {
	email := httpx.GetSessionEmail(r)
	if email == "" {
		return "Admin"
	}
	srv := httpx.GetServer(r)
	if srv != nil {
		var username string
		if err := h.service.db.QueryRowContext(r.Context(),
			`SELECT username FROM staff WHERE server_id = ? AND LOWER(email) = LOWER(?) LIMIT 1`,
			srv.ID, email,
		).Scan(&username); err == nil && strings.TrimSpace(username) != "" {
			return username
		}
	}
	return email
}

func (h *Handler) getStaffPerformance(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Server context required")
		return
	}
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "30d"
	}
	items, err := h.service.GetStaffPerformance(r.Context(), srv.ID, period)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load staff performance")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.StaffPerformanceListResponse{Staff: items})
}

func (h *Handler) getStaffDetails(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Server context required")
		return
	}
	username := chi.URLParam(r, "username")
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "30d"
	}
	resp, err := h.service.GetStaffDetails(r.Context(), srv.ID, username, period)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load staff details")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, resp)
}

func (h *Handler) getActivePunishments(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Server context required")
		return
	}
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "active"
	}
	items, err := h.service.GetPunishmentsList(r.Context(), srv.ID, status)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load punishments")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.ActivePunishmentsAuditResponse{Punishments: items})
}

func (h *Handler) getPunishments(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Server context required")
		return
	}
	limit := httpx.QueryInt(r, "limit", 50, 1, 100)
	canRollback := strings.EqualFold(r.URL.Query().Get("canRollback"), "true")
	items, err := h.service.GetPunishmentAuditLogs(r.Context(), srv.ID, limit, canRollback)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load punishment audit logs")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PunishmentAuditListResponse{Punishments: items})
}

func (h *Handler) getDatabaseTable(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Server context required")
		return
	}
	table := chi.URLParam(r, "table")
	limit := httpx.QueryInt(r, "limit", 100, 1, 100)
	skip := httpx.QueryInt(r, "skip", 0, 0, 1_000_000)

	resp, err := h.service.GetDatabaseTable(r.Context(), srv.ID, table, limit, skip)
	if errors.Is(err, errInvalidTable) {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid table name")
		return
	}
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to read database table")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, resp)
}

func (h *Handler) rollbackPunishment(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Server context required")
		return
	}
	id := chi.URLParam(r, "id")
	var req modlv1.RollbackRequest
	if r.ContentLength > 0 {
		if err := httpx.ReadProto(r, &req); err != nil {
			httpx.WriteError(w, r, http.StatusBadRequest, "Invalid rollback request")
			return
		}
	}
	reason := "Admin rollback"
	if req.Reason != nil && strings.TrimSpace(*req.Reason) != "" {
		reason = strings.TrimSpace(*req.Reason)
	}
	ok, err := h.service.RollbackPunishment(r.Context(), srv.ID, id, reason, h.currentPerformer(r))
	if errors.Is(err, errCannotRollback) {
		httpx.WriteError(w, r, http.StatusBadRequest, "This punishment cannot be rolled back")
		return
	}
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to rollback punishment")
		return
	}
	if !ok {
		httpx.WriteError(w, r, http.StatusNotFound, "Punishment not found")
		return
	}
	h.invalidateAudit(srv.ID)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.AuditRollbackResponse{
		Success: true,
		Message: "Punishment rolled back successfully",
	})
}

func (h *Handler) rollbackAllByStaff(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Server context required")
		return
	}
	username := chi.URLParam(r, "username")
	var req modlv1.RollbackRequest
	if r.ContentLength > 0 {
		if err := httpx.ReadProto(r, &req); err != nil {
			httpx.WriteError(w, r, http.StatusBadRequest, "Invalid rollback request")
			return
		}
	}
	reason := "Bulk rollback by admin"
	if req.Reason != nil && strings.TrimSpace(*req.Reason) != "" {
		reason = strings.TrimSpace(*req.Reason)
	}
	count, err := h.service.RollbackByStaff(r.Context(), srv.ID, username, nil, nil, reason, h.currentPerformer(r))
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to rollback punishments")
		return
	}
	h.invalidateAudit(srv.ID)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.AuditBulkOperationResponse{
		Success: true,
		Count:   count,
		Message: fmt.Sprintf("Successfully rolled back %d punishments", count),
	})
}

func (h *Handler) rollbackByDateRange(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Server context required")
		return
	}
	username := chi.URLParam(r, "username")
	var req modlv1.DateRangeRollbackRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid date range rollback request")
		return
	}
	if req.StartDate == nil || req.EndDate == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Start date and end date are required")
		return
	}
	startMs := req.StartDate.AsTime().UnixMilli()
	endMs := req.EndDate.AsTime().UnixMilli()
	reason := "Bulk rollback by admin"
	if req.Reason != nil && strings.TrimSpace(*req.Reason) != "" {
		reason = strings.TrimSpace(*req.Reason)
	}
	count, err := h.service.RollbackByStaff(r.Context(), srv.ID, username, &startMs, &endMs, reason, h.currentPerformer(r))
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to rollback punishments")
		return
	}
	h.invalidateAudit(srv.ID)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.AuditBulkOperationResponse{
		Success: true,
		Count:   count,
		Message: fmt.Sprintf("Successfully rolled back %d punishments", count),
	})
}

func (h *Handler) bulkPardon(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Server context required")
		return
	}
	var req modlv1.BulkPunishmentActionRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid bulk pardon request")
		return
	}
	count, err := h.service.BulkPardonByType(r.Context(), srv.ID, req.TypeOrdinals, req.Reason, h.currentPerformer(r))
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to bulk pardon punishments")
		return
	}
	h.invalidateAudit(srv.ID)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.AuditBulkOperationResponse{
		Success: true,
		Count:   count,
		Message: fmt.Sprintf("Successfully pardoned %d punishments", count),
	})
}

func (h *Handler) bulkSetExpiration(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Server context required")
		return
	}
	var req modlv1.BulkPunishmentActionRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid bulk set expiration request")
		return
	}
	if req.NewDurationMs == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "newDurationMs is required for set-expiration")
		return
	}
	count, err := h.service.BulkSetExpirationByType(r.Context(), srv.ID, req.TypeOrdinals, *req.NewDurationMs, req.Reason, h.currentPerformer(r))
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to update punishment expirations")
		return
	}
	h.invalidateAudit(srv.ID)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.AuditBulkOperationResponse{
		Success: true,
		Count:   count,
		Message: fmt.Sprintf("Successfully updated expiration for %d punishments", count),
	})
}
