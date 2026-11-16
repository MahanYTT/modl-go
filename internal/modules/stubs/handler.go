package stubs

import (
	"net/http"
	"time"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/httpx"

	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler struct{}

func NewHandler() *Handler { return &Handler{} }

func (h *Handler) RegisterPanelRoutes(r chi.Router) {
	r.Route("/billing", func(r chi.Router) {
		r.Get("/status", h.billingStatus)
		r.Get("/usage", h.billingUsage)
		r.Post("/usage-settings", h.usageSettings)
		r.Patch("/usage-settings", h.usageSettings)
		r.Post("/overage-limits", h.overageLimits)
		r.Patch("/overage-limits", h.overageLimits)
		r.Post("/cancel", h.cancel)
		r.Post("/resubscribe", h.resubscribe)
		r.Post("/checkout-session", h.checkout)
		r.Post("/portal-session", h.portal)
	})
	r.Get("/players/{uuid}/replays", h.emptyReplays)
	r.Get("/replays", h.emptyReplays)
	r.Get("/replays/{id}", h.notFoundReplay)
	r.Get("/settings/ai-punishment-types", h.emptyAIPunishmentTypes)
	r.Post("/settings/test-database", h.testDatabase)
}

func (h *Handler) billingStatus(w http.ResponseWriter, r *http.Request) {
	plan := "PRO"
	st := "ACTIVE"
	gf := true
	maxBytes := int64(100 * 1024 * 1024 * 1024)
	maxAI := int64(10000)
	now := time.Now().UTC()
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.BillingStatusResponse{
		Plan:                      plan,
		SubscriptionStatus:        st,
		CurrentPeriodStart:        timestamppb.New(now.Add(-24 * time.Hour)),
		CurrentPeriodEnd:          timestamppb.New(now.Add(365 * 24 * time.Hour)),
		CustomDomainGrandfathered: &gf,
		MaxStorageLimitBytes:      &maxBytes,
		MaxAiOverageRequests:      &maxAI,
		PlanValue:                 &plan,
		SubscriptionStatusValue:   &st,
	})
}

func (h *Handler) billingUsage(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.UsageResponse{
		Period: &modlv1.UsageResponse_Period{
			Start: timestamppb.New(now.Add(-24 * time.Hour)),
			End:   timestamppb.New(now.Add(30 * 24 * time.Hour)),
		},
		Cdn: &modlv1.UsageResponse_UsageMetric{Used: 0, Limit: 100},
		Ai:  &modlv1.UsageResponse_UsageMetric{Used: 0, Limit: 10000},
	})
}

func (h *Handler) usageSettings(w http.ResponseWriter, r *http.Request) {
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.UsageBillingSettingsResponse{
		Success: true,
		Message: "Self-hosted mode has no usage billing.",
	})
}

func (h *Handler) overageLimits(w http.ResponseWriter, r *http.Request) {
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.UpdateOverageLimitsResponse{
		Success:              true,
		MaxStorageLimitBytes: 100 * 1024 * 1024 * 1024,
		MaxAiOverageRequests: 10000,
	})
}

func (h *Handler) cancel(w http.ResponseWriter, r *http.Request) {
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.CancelResponse{
		Success: false,
		Message: "Self-hosted instance.",
	})
}

func (h *Handler) resubscribe(w http.ResponseWriter, r *http.Request) {
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.ResubscribeResponse{
		Success: true,
		Message: "Self-hosted instance is always active.",
	})
}

func (h *Handler) checkout(w http.ResponseWriter, r *http.Request) {
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.CheckoutSessionResponse{Url: "/settings"})
}

func (h *Handler) portal(w http.ResponseWriter, r *http.Request) {
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PortalSessionResponse{Url: "/settings"})
}

func (h *Handler) emptyReplays(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"replays": []any{},
		"total":   0,
		"page":    1,
		"limit":   20,
	})
}

func (h *Handler) notFoundReplay(w http.ResponseWriter, r *http.Request) {
	httpx.WriteError(w, r, http.StatusNotFound, "Replay feature is disabled in self-hosted mode")
}

func (h *Handler) emptyAIPunishmentTypes(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"types": []any{}})
}

func (h *Handler) testDatabase(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "message": "SQLite OK"})
}
