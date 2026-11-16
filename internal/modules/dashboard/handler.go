package dashboard

import (
	"database/sql"
	"net/http"

	modlv1 "modl-go/gen/modl/v1"
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
	r.Route("/dashboard", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(httpx.RequirePanelPolicy(h.perms, models.AccessPolicy{
				Rule:   models.RuleRequirePermission,
				View:   "admin.audit.view.dashboard",
				Modify: "admin.audit.view.dashboard",
			}))
			r.Get("/metrics", h.getMetrics)
			r.Get("/recent-tickets", h.getRecentTickets)
			r.Get("/recent-punishments", h.getRecentPunishments)
			r.Get("/activity/recent", h.getRecentActivity)
		})

		r.Group(func(r chi.Router) {
			r.Use(httpx.RequirePanelPolicy(h.perms, models.AccessPolicy{
				Rule: models.RulePermitAll,
			}))
			r.Get("/alerts", h.getVisibleAlerts)
		})
	})
}

func (h *Handler) RegisterMinecraftV1Routes(r chi.Router) {
	r.Route("/dashboard", func(r chi.Router) {
		r.Get("/stats", h.getMinecraftStatsV1)
	})
}

func (h *Handler) RegisterMinecraftV3Routes(r chi.Router) {
	r.Route("/dashboard", func(r chi.Router) {
		r.Get("/stats", h.getMinecraftStatsV3)
	})
}

func (h *Handler) getMetrics(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Server context required")
		return
	}
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "7d"
	}
	resp, err := h.service.GetMetrics(r.Context(), srv, period)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load dashboard metrics")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, resp)
}

func (h *Handler) getRecentTickets(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Server context required")
		return
	}
	limit := httpx.QueryInt(r, "limit", 10, 1, 100)
	tickets, err := h.service.GetRecentTickets(r.Context(), srv, limit)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load recent tickets")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.DashboardRecentTicketsResponse{Tickets: tickets})
}

func (h *Handler) getRecentPunishments(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Server context required")
		return
	}
	limit := httpx.QueryInt(r, "limit", 10, 1, 100)
	punishments, err := h.service.GetRecentPunishments(r.Context(), srv, limit)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load recent punishments")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.DashboardRecentPunishmentsResponse{Punishments: punishments})
}

func (h *Handler) getRecentActivity(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Server context required")
		return
	}
	email := httpx.GetSessionEmail(r)
	if email == "" {
		httpx.WriteErrorCode(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}
	limit := httpx.QueryInt(r, "limit", 20, 1, 100)
	days := httpx.QueryInt(r, "days", 7, 1, maxActivityDays)
	items, err := h.service.GetRecentActivity(r.Context(), srv, email, limit, days)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load recent activity")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.DashboardActivityResponse{Activity: items})
}

func (h *Handler) getVisibleAlerts(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	email := httpx.GetSessionEmail(r)
	if email == "" {
		httpx.WriteErrorCode(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}
	superAdmin := srv != nil && h.perms != nil && h.perms.IsSuperAdmin(srv, email)
	items, err := h.service.ListVisibleAlerts(r.Context(), superAdmin)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load system alerts")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PanelSystemAlertsResponse{Items: items})
}

func (h *Handler) getMinecraftStatsV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Server context required")
		return
	}
	stats, err := h.service.GetMinecraftStats(r.Context(), srv)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load minecraft dashboard stats")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status": 200,
		"stats":  stats,
	})
}

func (h *Handler) getMinecraftStatsV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Server context required")
		return
	}
	stats, err := h.service.GetMinecraftStats(r.Context(), srv)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load minecraft dashboard stats")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.MinecraftDashboardResponse{
		Status: 200,
		Stats: &modlv1.MinecraftDashboardStatsResponse{
			UnresolvedReports:      stats.UnresolvedReports,
			UnresolvedTickets:      stats.UnresolvedTickets,
			OnlineStaff:            stats.OnlineStaff,
			OnlinePlayers:          stats.OnlinePlayers,
			ActiveBans:             stats.ActiveBans,
			ActiveMutes:            stats.ActiveMutes,
			TotalActivePunishments: stats.TotalActivePunishments,
			TotalPlayers:           stats.TotalPlayers,
		},
	})
}
