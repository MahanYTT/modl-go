package health

import (
	"net/http"

	"modl-go/internal/httpx"

	"github.com/go-chi/chi/v5"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/v1/health", h.health)
	r.Head("/v1/health", h.healthHead)
	r.Head("/v1", h.healthHead)
	r.Get("/actuator/health", h.health)
	r.Get("/actuator/health/liveness", h.health)
	r.Get("/actuator/health/readiness", h.health)
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"timestamp": httpx.NowMs(),
	})
}

func (h *Handler) healthHead(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
