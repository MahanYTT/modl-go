package analytics

import (
	"database/sql"
	"net/http"

	"modl-go/internal/httpx"
	"modl-go/internal/models"

	"github.com/go-chi/chi/v5"
)

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
	r.Route("/analytics", func(r chi.Router) {
		r.Use(httpx.RequirePanelPolicy(h.perms, models.AccessPolicy{
			Rule:   models.RuleRequirePermission,
			View:   "admin.audit.view.analytics",
			Modify: "admin.audit.view.analytics",
		}))
		r.Get("/overview", h.getOverview)
		r.Get("/tickets", h.getTicketAnalytics)
		r.Get("/punishments", h.getPunishmentAnalytics)
		r.Get("/audit-logs", h.getAuditLogAnalytics)
		r.Get("/player-activity", h.getPlayerActivity)
	})
}

func (h *Handler) getOverview(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Server context required")
		return
	}
	resp, err := h.service.GetOverview(r.Context(), srv)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load overview analytics")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, resp)
}

func (h *Handler) getTicketAnalytics(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Server context required")
		return
	}
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "30d"
	}
	resp, err := h.service.GetTicketAnalytics(r.Context(), srv, period)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load ticket analytics")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, resp)
}

func (h *Handler) getPunishmentAnalytics(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Server context required")
		return
	}
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "30d"
	}
	resp, err := h.service.GetPunishmentAnalytics(r.Context(), srv, period)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load punishment analytics")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, resp)
}

func (h *Handler) getAuditLogAnalytics(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Server context required")
		return
	}
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "30d"
	}
	resp, err := h.service.GetAuditLogAnalytics(r.Context(), srv, period)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load audit log analytics")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, resp)
}

func (h *Handler) getPlayerActivity(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Server context required")
		return
	}
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "30d"
	}
	resp, err := h.service.GetPlayerActivity(r.Context(), srv, period)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load player activity analytics")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, resp)
}
