package appeal

import (
	"database/sql"
	"errors"
	"net/http"
	"os"
	"strings"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/config"
	"modl-go/internal/email"
	"modl-go/internal/httpx"
	"modl-go/internal/models"
	"modl-go/internal/modules/ticket"
	"modl-go/internal/realtime"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	store               *Store
	ticketStore         *ticket.Store
	ticketSvc           *ticket.Service
	service             *Service
	roles               models.PermissionChecker
	hub                 *realtime.Hub
	appealTokenRequired bool
}

func NewHandler(
	db *sql.DB,
	cfg *config.Config,
	mailer *email.Sender,
	hub *realtime.Hub,
	roles models.PermissionChecker,
	tickets *ticket.Handler,
) *Handler {
	var tStore *ticket.Store
	var tSvc *ticket.Service
	if tickets != nil {
		tStore = tickets.Store()
		tSvc = tickets.Service()
	} else {
		tStore = ticket.NewStore(db)
		tSvc = ticket.NewService(tStore, cfg, mailer, hub, nil)
	}
	store := NewStore(db, tStore)
	svc := NewService(store, tSvc, hub)
	return &Handler{
		store:               store,
		ticketStore:         tStore,
		ticketSvc:           tSvc,
		service:             svc,
		roles:               roles,
		hub:                 hub,
		appealTokenRequired: strings.EqualFold(os.Getenv("MODL_PUBLIC_ACCESS_APPEAL_TOKEN_REQUIRED"), "true"),
	}
}

func (h *Handler) RegisterPanelRoutes(r chi.Router) {
	r.Route("/appeals", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(httpx.RequirePanelPolicy(h.roles, models.AccessPolicy{
				Rule:   models.RuleRequirePermission,
				View:   "ticket.view.all",
				Modify: "appeal.modify",
			}))
			r.Get("/punishment/{punishmentId}", h.getAppealsByPunishment)
			r.Get("/{id}", h.getPanelAppealByID)
			r.Patch("/{id}/status", h.updateAppealStatus)
		})
		r.With(httpx.RequirePanelPolicy(h.roles, models.AccessPolicy{
			Rule: models.RuleAppealReply,
		})).Post("/{id}/replies", h.addPanelReply)
	})
}

func (h *Handler) RegisterPublicRoutes(r chi.Router) {
	r.Route("/appeals", func(r chi.Router) {
		r.Get("/{id}", h.getPublicAppeal)
		r.Post("/", h.createPublicAppeal)
		r.Post("/{id}/replies", h.addPublicReply)
		r.Post("/{id}/request-verification", h.requestPublicVerification)
		r.Post("/{id}/verify", h.verifyPublicCode)
	})
}

func writeAppealErr(w http.ResponseWriter, r *http.Request, err error) {
	switch e := err.(type) {
	case ticket.BadRequestError:
		httpx.WriteError(w, r, http.StatusBadRequest, e.Msg)
	case ticket.ConflictError:
		httpx.WriteError(w, r, http.StatusConflict, e.Msg)
	case ticket.UnauthorizedError:
		httpx.WriteError(w, r, http.StatusUnauthorized, e.Msg)
	case ticket.ForbiddenError:
		httpx.WriteError(w, r, http.StatusForbidden, e.Msg)
	default:
		if errors.Is(err, ticket.ErrNotFound) {
			httpx.WriteError(w, r, http.StatusNotFound, "Appeal not found")
			return
		}
		httpx.WriteError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}

// --- Panel Appeal Endpoints ---

func (h *Handler) getAppealsByPunishment(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	punishmentID := chi.URLParam(r, "punishmentId")
	appeals, err := h.service.GetAppealsByPunishment(r.Context(), srv.ID, punishmentID)
	if err != nil {
		writeAppealErr(w, r, err)
		return
	}
	if len(appeals) == 0 {
		httpx.WriteError(w, r, http.StatusNotFound, "No appeals found for punishment")
		return
	}
	pb := &modlv1.AppealTicketsResponse{
		Tickets: make([]*modlv1.TicketResponse, 0, len(appeals)),
	}
	for _, a := range appeals {
		pb.Tickets = append(pb.Tickets, ToProtoAppealTicketResponse(r.Context(), h.ticketStore, srv.ID, a))
	}
	httpx.WriteProto(w, r, http.StatusOK, pb)
}

func (h *Handler) getPanelAppealByID(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	a, err := h.service.GetAppealRaw(r.Context(), srv.ID, id)
	if err != nil {
		writeAppealErr(w, r, err)
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, ToProtoAppealTicketResponse(r.Context(), h.ticketStore, srv.ID, a))
}

func (h *Handler) addPanelReply(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	var req modlv1.AddAppealReplyRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	reply, err := h.service.AddPanelReply(r.Context(), srv, id, &req, httpx.GetSessionEmail(r))
	if err != nil {
		writeAppealErr(w, r, err)
		return
	}
	httpx.WriteProto(w, r, http.StatusCreated, &modlv1.AddTicketReplyResponse{
		Success: true,
		Reply:   ticket.ToProtoPublicTicketReply(reply),
	})
}

func (h *Handler) updateAppealStatus(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	var req modlv1.UpdateAppealStatusRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	actorEmail := httpx.GetSessionEmail(r)
	a, err := h.service.UpdateStatus(r.Context(), srv, id, &req, actorEmail)
	if err != nil {
		writeAppealErr(w, r, err)
		return
	}
	actor := h.ticketStore.ResolvePanelActor(r.Context(), srv.ID, actorEmail)
	if strings.TrimSpace(req.Status) != "" {
		desc := "Appeal " + id + " " + string(a.Status)
		if req.Resolution != nil && strings.TrimSpace(*req.Resolution) != "" {
			desc += " (" + *req.Resolution + ")"
		}
		h.ticketStore.RecordAuditLog(r.Context(), srv.ID, "moderation", actor, desc)
	} else {
		h.ticketStore.RecordAuditLog(r.Context(), srv.ID, "staff", actor, "Updated appeal "+id)
	}
	httpx.WriteProto(w, r, http.StatusOK, ToProtoAppealTicketResponse(r.Context(), h.ticketStore, srv.ID, a))
}

// --- Public Appeal Endpoints ---

func (h *Handler) getPublicAppeal(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	token := r.URL.Query().Get("token")
	a, err := h.service.GetAppealRaw(r.Context(), srv.ID, id)
	if err != nil {
		writeAppealErr(w, r, err)
		return
	}
	access := h.ticketSvc.AuthorizePublicAccess(r.Context(), srv, a, token)
	if access.NotFound {
		httpx.WriteError(w, r, http.StatusNotFound, "Appeal not found")
		return
	}
	if !access.Allowed {
		httpx.WriteProto(w, r, http.StatusForbidden, &modlv1.TicketVerificationRequiredResponse{
			RequiresVerification: true,
			EmailHint:            access.EmailHint,
			TicketId:             id,
		})
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, ToPublicAppealResponse(a))
}

func (h *Handler) createPublicAppeal(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	token := r.URL.Query().Get("token")
	var req modlv1.CreateAppealRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	a, err := h.service.CreateAppeal(r.Context(), srv, &req, token, h.appealTokenRequired)
	if err != nil {
		writeAppealErr(w, r, err)
		return
	}
	httpx.WriteProto(w, r, http.StatusCreated, ToCreatePublicAppealResponse(a))
}

func (h *Handler) addPublicReply(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	token := r.URL.Query().Get("token")
	var req modlv1.AddAppealReplyRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	a, err := h.service.GetAppealRaw(r.Context(), srv.ID, id)
	if err != nil {
		writeAppealErr(w, r, err)
		return
	}
	access := h.ticketSvc.AuthorizePublicAccess(r.Context(), srv, a, token)
	if access.NotFound {
		httpx.WriteError(w, r, http.StatusNotFound, "Appeal not found")
		return
	}
	if !access.Allowed {
		httpx.WriteProto(w, r, http.StatusForbidden, &modlv1.TicketVerificationRequiredResponse{
			RequiresVerification: true,
			EmailHint:            access.EmailHint,
			TicketId:             id,
		})
		return
	}
	reply, err := h.service.AddPublicReplyFromProto(r.Context(), srv, id, &req)
	if err != nil {
		writeAppealErr(w, r, err)
		return
	}
	httpx.WriteProto(w, r, http.StatusCreated, ToAddPublicAppealReplyResponse(reply))
}

func (h *Handler) requestPublicVerification(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	a, err := h.service.GetAppealRaw(r.Context(), srv.ID, id)
	if err != nil || a.Hidden {
		httpx.WriteError(w, r, http.StatusNotFound, "Appeal not found")
		return
	}
	emailHint, err := h.ticketSvc.RequestVerificationCode(r.Context(), srv, id)
	if err != nil {
		writeAppealErr(w, r, err)
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PublicTicketVerificationRequestResponse{
		Success:   true,
		Message:   "Verification code sent",
		EmailHint: emailHint,
	})
}

func (h *Handler) verifyPublicCode(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	a, err := h.service.GetAppealRaw(r.Context(), srv.ID, id)
	if err != nil || a.Hidden {
		httpx.WriteError(w, r, http.StatusNotFound, "Appeal not found")
		return
	}
	var req modlv1.VerifyTicketCodeRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	token, err := h.ticketSvc.VerifyCode(r.Context(), srv, id, req.Code)
	if err != nil {
		writeAppealErr(w, r, err)
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PublicTicketVerificationResponse{
		Success: true,
		Token:   token,
		Message: "Verification successful",
	})
}
