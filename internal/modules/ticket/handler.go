package ticket

import (
	"database/sql"
	"errors"
	"math"
	"net/http"
	"strings"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/ai"
	"modl-go/internal/config"
	"modl-go/internal/email"
	"modl-go/internal/httpx"
	"modl-go/internal/models"
	"modl-go/internal/realtime"

	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler struct {
	store   *Store
	service *Service
	roles   models.PermissionChecker
	hub     *realtime.Hub
}

func NewHandler(
	db *sql.DB,
	cfg *config.Config,
	mailer *email.Sender,
	hub *realtime.Hub,
	roles models.PermissionChecker,
	aiClient *ai.Client,
) *Handler {
	store := NewStore(db)
	svc := NewService(store, cfg, mailer, hub, aiClient)
	return &Handler{
		store:   store,
		service: svc,
		roles:   roles,
		hub:     hub,
	}
}

func (h *Handler) Service() *Service { return h.service }
func (h *Handler) Store() *Store     { return h.store }

func (h *Handler) RegisterPanelRoutes(r chi.Router) {
	r.Route("/tickets", func(r chi.Router) {
		r.Use(httpx.RequirePanelPolicy(h.roles, models.AccessPolicy{
			Rule:   models.RuleRequirePermission,
			View:   "ticket.view.all",
			Modify: "ticket.reply.all",
		}))
		r.Get("/", h.getPanelTickets)
		r.Get("/counts", h.getPanelTicketCounts)
		r.With(httpx.RequirePanelPolicy(h.roles, models.AccessPolicy{
			Rule:   models.RuleRequirePermission,
			View:   "ticket.view.all",
			Modify: "ticket.close.all",
		})).Post("/bulk", h.bulkUpdateTickets)
		r.Get("/player/{uuid}", h.getTicketsByPlayer)
		r.Get("/tag/{tag}", h.getTicketsByTag)
		r.Get("/{id}", h.getPanelTicketByID)
		r.Post("/", h.createPanelTicket)
		r.Patch("/{id}", h.updatePanelTicket)
		r.Post("/{id}/notes", h.addPanelNote)
		r.Post("/{id}/replies", h.addPanelReply)
		r.Post("/{id}/tags", h.addPanelTag)
		r.Delete("/{id}/tags/{tag}", h.removePanelTag)
		r.Post("/{id}/quick-response", h.processQuickResponse)
	})

	r.Route("/ticket-subscriptions", func(r chi.Router) {
		r.Use(httpx.RequirePanelPolicy(h.roles, models.AccessPolicy{
			Rule:   models.RuleRequirePermission,
			View:   "ticket.view.all",
			Modify: "ticket.reply.all",
		}))
		r.Get("/", h.getSubscriptions)
		r.Delete("/{ticketId}", h.unsubscribe)
		r.Get("/updates", h.getSubscriptionUpdates)
		r.Post("/updates/{updateId}/read", h.markSubscriptionUpdateRead)
		r.Post("/tickets/{ticketId}/read", h.markTicketSubscriptionRead)
		r.Get("/assigned-updates", h.getAssignedUpdates)
	})
}

func (h *Handler) RegisterPublicRoutes(r chi.Router) {
	r.Route("/tickets", func(r chi.Router) {
		r.Post("/", h.createPublicTicket)
		r.Post("/unfinished", h.createUnfinishedPublicTicket)
		r.Get("/{id}", h.getPublicTicket)
		r.Get("/{id}/status", h.getPublicTicketStatus)
		r.Post("/{id}/replies", h.addPublicReply)
		r.Post("/{id}/submit", h.submitPublicTicketForm)
		r.Post("/{id}/request-verification", h.requestPublicVerification)
		r.Post("/{id}/verify", h.verifyPublicCode)
	})
}

func (h *Handler) RegisterMinecraftV1Routes(r chi.Router) {
	r.Route("/tickets", func(r chi.Router) {
		r.Post("/", h.createMinecraftTicketV1)
		r.Post("/unfinished", h.createUnfinishedMinecraftTicketV1)
		r.Get("/", h.getMinecraftTicketsV1)
		r.Get("/player/{uuid}", h.getPlayerTicketsV1)
		r.Post("/by-ids", h.getTicketsByIdsV1)
		r.Get("/{id}", h.getMinecraftTicketDetailV1)
		r.Post("/{id}/claim", h.claimMinecraftTicketV1)
	})
	r.Route("/reports", func(r chi.Router) {
		r.Get("/", h.getMinecraftReportsV1)
		r.Get("/player/{uuid}", h.getMinecraftPlayerReportsV1)
		r.Post("/{id}/assign", h.assignMinecraftReportV1)
		r.Post("/{id}/dismiss", h.dismissMinecraftReportV1)
		r.Post("/{id}/resolve", h.resolveMinecraftReportV1)
	})
}

func (h *Handler) RegisterMinecraftV3Routes(r chi.Router) {
	r.Route("/tickets", func(r chi.Router) {
		r.Post("/", h.createMinecraftTicketV3)
		r.Post("/unfinished", h.createUnfinishedMinecraftTicketV3)
		r.Get("/", h.getMinecraftTicketsV3)
		r.Get("/player/{uuid}", h.getPlayerTicketsV3)
		r.Post("/by-ids", h.getTicketsByIdsV3)
		r.Get("/{id}", h.getMinecraftTicketDetailV3)
		r.Post("/{id}/claim", h.claimMinecraftTicketV3)
	})
	r.Route("/reports", func(r chi.Router) {
		r.Get("/", h.getMinecraftReportsV3)
		r.Get("/player/{uuid}", h.getMinecraftPlayerReportsV3)
		r.Post("/{id}/assign", h.assignMinecraftReportV3)
		r.Post("/{id}/dismiss", h.dismissMinecraftReportV3)
		r.Post("/{id}/resolve", h.resolveMinecraftReportV3)
	})
}

func writeServiceErr(w http.ResponseWriter, r *http.Request, err error) {
	switch e := err.(type) {
	case BadRequestError:
		httpx.WriteError(w, r, http.StatusBadRequest, e.Msg)
	case ConflictError:
		httpx.WriteError(w, r, http.StatusConflict, e.Msg)
	case UnauthorizedError:
		httpx.WriteError(w, r, http.StatusUnauthorized, e.Msg)
	case ForbiddenError:
		httpx.WriteError(w, r, http.StatusForbidden, e.Msg)
	default:
		if errors.Is(err, ErrNotFound) {
			httpx.WriteError(w, r, http.StatusNotFound, "Ticket not found")
			return
		}
		httpx.WriteError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}

func queryMulti(r *http.Request, key string) []string {
	vals := r.URL.Query()[key]
	if len(vals) == 0 {
		return nil
	}
	var out []string
	for _, v := range vals {
		for _, part := range strings.Split(v, ",") {
			p := strings.TrimSpace(part)
			if p != "" {
				out = append(out, p)
			}
		}
	}
	return out
}

// --- Panel Ticket Endpoints ---

func (h *Handler) getPanelTickets(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	page := httpx.QueryInt(r, "page", 1, 1, 100000)
	limit := httpx.QueryInt(r, "limit", 10, 1, 100)
	q := r.URL.Query()
	filter := TicketSearchFilter{
		Search:    q.Get("search"),
		Status:    q.Get("status"),
		Types:     queryMulti(r, "type"),
		Author:    q.Get("author"),
		Labels:    queryMulti(r, "labels"),
		Assignees: queryMulti(r, "assignee"),
	}
	sortOpt := q.Get("sort")
	if sortOpt == "" {
		sortOpt = "newest"
	}

	tickets, total, err := h.store.SearchTickets(r.Context(), srv.ID, filter, sortOpt, page, limit)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	totalPages := int32(math.Ceil(float64(total) / float64(limit)))
	items := make([]*modlv1.TicketListItemResponse, 0, len(tickets))
	for _, t := range tickets {
		items = append(items, ToProtoTicketListItem(t))
	}
	resp := &modlv1.PaginatedTicketsResponse{
		Tickets: items,
		Pagination: &modlv1.PaginatedTicketsResponse_PaginatedTicketsPaginationInfo{
			Current:      int32(page),
			Total:        totalPages,
			Limit:        int32(limit),
			TotalTickets: total,
			HasNext:      int32(page) < totalPages,
			HasPrev:      page > 1,
		},
		Filters: &modlv1.PaginatedTicketsResponse_PaginatedTicketsFiltersInfo{
			Search: filter.Search,
			Status: filter.Status,
			Types:  filter.Types,
		},
	}
	httpx.WriteProto(w, r, http.StatusOK, resp)
}

func (h *Handler) getPanelTicketCounts(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	q := r.URL.Query()
	filter := TicketSearchFilter{
		Search:    q.Get("search"),
		Types:     queryMulti(r, "type"),
		Author:    q.Get("author"),
		Labels:    queryMulti(r, "labels"),
		Assignees: queryMulti(r, "assignee"),
	}
	counts, err := h.store.CountTickets(r.Context(), srv.ID, filter)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.TicketCountsResponse{
		Open:   counts.Open,
		Closed: counts.Closed,
	})
}

func (h *Handler) bulkUpdateTickets(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.BulkTicketUpdateRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	updated, err := h.service.BulkUpdate(r.Context(), srv, &req, httpx.GetSessionEmail(r))
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.BulkTicketUpdateResponse{
		Updated: int32(updated),
		Message: "Tickets updated successfully",
	})
}

func (h *Handler) getPanelTicketByID(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	t, err := h.store.FindByID(r.Context(), srv.ID, id)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	if email := httpx.GetSessionEmail(r); email != "" {
		h.service.MarkTicketAsRead(r.Context(), srv.ID, email, id)
	}
	httpx.WriteProto(w, r, http.StatusOK, ToProtoTicketResponse(r.Context(), h.store, srv.ID, t))
}

func (h *Handler) createPanelTicket(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.CreateTicketRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	t, err := h.service.CreateTicket(r.Context(), srv, &req, false, false)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	httpx.WriteProto(w, r, http.StatusCreated, ToProtoTicketResponse(r.Context(), h.store, srv.ID, t))
}

func (h *Handler) updatePanelTicket(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	var req modlv1.UpdateTicketRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	t, err := h.service.UpdateTicket(r.Context(), srv, id, &req, httpx.GetSessionEmail(r))
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, ToProtoTicketResponse(r.Context(), h.store, srv.ID, t))
}

func (h *Handler) addPanelNote(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	var req modlv1.AddNoteRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	note, err := h.service.AddNote(r.Context(), srv, id, &req, httpx.GetSessionEmail(r))
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	httpx.WriteProto(w, r, http.StatusCreated, ToProtoTicketNote(note))
}

func (h *Handler) addPanelReply(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	var req modlv1.AddReplyRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	reply, err := h.service.AddReply(r.Context(), srv, id, &req, httpx.GetSessionEmail(r), true)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	httpx.WriteProto(w, r, http.StatusCreated, &modlv1.AddTicketReplyResponse{
		Success: true,
		Reply:   ToProtoPublicTicketReply(reply),
	})
}

func (h *Handler) addPanelTag(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	var req modlv1.AddTagRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	tags, err := h.service.AddTag(r.Context(), srv, id, req.Tag)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.TicketTagsResponse{Tags: tags})
}

func (h *Handler) removePanelTag(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	tag := chi.URLParam(r, "tag")
	tags, err := h.service.RemoveTag(r.Context(), srv, id, tag)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.TicketTagsResponse{Tags: tags})
}

func (h *Handler) getTicketsByPlayer(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	uuid := chi.URLParam(r, "uuid")
	tickets, err := h.store.FindByPlayer(r.Context(), srv.ID, uuid)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	out := make([]PlayerTicketResponse, 0, len(tickets))
	for _, t := range tickets {
		out = append(out, ToPlayerTicketResponse(t))
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) getTicketsByTag(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	tag := chi.URLParam(r, "tag")
	tickets, err := h.store.FindByTag(r.Context(), srv.ID, tag)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	out := make([]PlayerTicketResponse, 0, len(tickets))
	for _, t := range tickets {
		out = append(out, ToPlayerTicketResponse(t))
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) processQuickResponse(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	var req modlv1.QuickResponseRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	res, err := h.service.ProcessQuickResponse(r.Context(), srv, id, &req, httpx.GetSessionEmail(r))
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	status := http.StatusOK
	if !res.Success {
		status = http.StatusBadRequest
	}
	httpx.WriteProto(w, r, status, res)
}

// --- Ticket Subscription Endpoints ---

func (h *Handler) getSubscriptions(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	email := httpx.GetSessionEmail(r)
	if email == "" {
		httpx.WriteError(w, r, http.StatusUnauthorized, "Not authenticated")
		return
	}
	resp, err := h.service.GetSubscriptions(r.Context(), srv.ID, email)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, resp)
}

func (h *Handler) unsubscribe(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	email := httpx.GetSessionEmail(r)
	if email == "" {
		httpx.WriteError(w, r, http.StatusUnauthorized, "Not authenticated")
		return
	}
	ticketID := chi.URLParam(r, "ticketId")
	ok, err := h.service.Unsubscribe(r.Context(), srv.ID, email, ticketID)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	if !ok {
		httpx.WriteError(w, r, http.StatusNotFound, "Subscription not found")
		return
	}
	h.service.InvalidateNotifications(srv.ID)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.DeleteTicketSubscriptionResponse{
		Message: "Successfully unsubscribed from ticket",
	})
}

func (h *Handler) getSubscriptionUpdates(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	email := httpx.GetSessionEmail(r)
	if email == "" {
		httpx.WriteError(w, r, http.StatusUnauthorized, "Not authenticated")
		return
	}
	limit := httpx.QueryInt(r, "limit", 10, 1, 100)
	resp, err := h.service.GetSubscriptionUpdates(r.Context(), srv.ID, email, limit)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, resp)
}

func (h *Handler) markSubscriptionUpdateRead(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	email := httpx.GetSessionEmail(r)
	if email == "" {
		httpx.WriteError(w, r, http.StatusUnauthorized, "Not authenticated")
		return
	}
	updateID := chi.URLParam(r, "updateId")
	ok, err := h.store.MarkSubscriptionUpdateRead(r.Context(), srv.ID, email, updateID)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	h.service.InvalidateNotifications(srv.ID)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.MarkSubscriptionUpdateReadResponse{
		Message:  "Update marked as read",
		Modified: ok,
	})
}

func (h *Handler) markTicketSubscriptionRead(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	email := httpx.GetSessionEmail(r)
	if email == "" {
		httpx.WriteError(w, r, http.StatusUnauthorized, "Not authenticated")
		return
	}
	ticketID := chi.URLParam(r, "ticketId")
	h.service.MarkTicketAsRead(r.Context(), srv.ID, email, ticketID)
	if h.hub != nil {
		h.hub.InvalidatePanel(srv.ID, modlv1.PanelResource_PANEL_RESOURCE_NOTIFICATIONS, ticketID)
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.MarkTicketSubscriptionReadResponse{
		Message: "All updates for ticket marked as read",
	})
}

func (h *Handler) getAssignedUpdates(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	email := httpx.GetSessionEmail(r)
	if email == "" {
		httpx.WriteError(w, r, http.StatusUnauthorized, "Not authenticated")
		return
	}
	limit := httpx.QueryInt(r, "limit", 10, 1, 100)
	resp, err := h.service.GetAssignedTicketUpdates(r.Context(), srv.ID, email, limit)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, resp)
}

// --- Public Ticket Endpoints ---

func toCreateTicketResponsePB(t *Ticket, msg string) *modlv1.CreateTicketResponse {
	id := t.ID
	subj := nonEmpty(t.Subject, t.Type.DisplayName())
	st := string(t.Status)
	created := ""
	if t.CreatedAtMs > 0 {
		created = MsToTime(t.CreatedAtMs).Format("2006-01-02T15:04:05.000Z")
	}
	return &modlv1.CreateTicketResponse{
		Success:  true,
		TicketId: &id,
		Message:  &msg,
		Ticket: &modlv1.CreateTicketResponse_CreateTicketInfo{
			Id:      t.ID,
			Type:    string(t.Type),
			Subject: &subj,
			Status:  &st,
			Created: &created,
		},
	}
}

func (h *Handler) createPublicTicket(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.CreateTicketRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	t, err := h.service.CreateTicket(r.Context(), srv, &req, true, false)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	httpx.WriteProto(w, r, http.StatusCreated, toCreateTicketResponsePB(t, "Ticket created successfully"))
}

func (h *Handler) createUnfinishedPublicTicket(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.CreateTicketRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	t, err := h.service.CreateTicket(r.Context(), srv, &req, true, true)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	httpx.WriteProto(w, r, http.StatusCreated, toCreateTicketResponsePB(t, "Ticket created successfully (Unfinished)"))
}

func (h *Handler) getPublicTicket(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	token := r.URL.Query().Get("token")
	t, err := h.store.FindByID(r.Context(), srv.ID, id)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	access := h.service.AuthorizePublicAccess(r.Context(), srv, t, token)
	if access.NotFound {
		httpx.WriteError(w, r, http.StatusNotFound, "Ticket not found")
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
	allowlist := h.service.BuildFormFieldAllowlist(r.Context(), srv.ID, t.Type)
	httpx.WriteProto(w, r, http.StatusOK, ToPublicTicketResponse(r.Context(), h.store, srv.ID, t, allowlist))
}

func (h *Handler) getPublicTicketStatus(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	token := r.URL.Query().Get("token")
	t, err := h.store.FindByID(r.Context(), srv.ID, id)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	access := h.service.AuthorizePublicAccess(r.Context(), srv, t, token)
	if access.NotFound {
		httpx.WriteError(w, r, http.StatusNotFound, "Ticket not found")
		return
	}
	if !access.Allowed {
		httpx.WriteError(w, r, http.StatusForbidden, "Email verification required")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PublicTicketStatusResponse{
		Id:      t.ID,
		Type:    string(t.Type),
		Subject: nonEmpty(t.Subject, "No Subject"),
		Status:  string(t.Status),
		Created: MsToTimestamp(t.CreatedAtMs),
		Locked:  t.Locked,
	})
}

func (h *Handler) addPublicReply(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	token := r.URL.Query().Get("token")
	var req modlv1.AddReplyRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	t, err := h.store.FindByID(r.Context(), srv.ID, id)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	access := h.service.AuthorizePublicAccess(r.Context(), srv, t, token)
	if access.NotFound {
		httpx.WriteError(w, r, http.StatusNotFound, "Ticket not found")
		return
	}
	if !access.Allowed {
		httpx.WriteError(w, r, http.StatusForbidden, "Email verification required")
		return
	}
	reply, err := h.service.AddReply(r.Context(), srv, id, &req, "", false)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	httpx.WriteProto(w, r, http.StatusCreated, &modlv1.AddTicketReplyResponse{
		Success: true,
		Message: "Reply added successfully",
		Reply:   ToProtoPublicTicketReply(reply),
	})
}

func (h *Handler) submitPublicTicketForm(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	token := r.URL.Query().Get("token")
	var req modlv1.SubmitTicketFormRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	t, err := h.store.FindByID(r.Context(), srv.ID, id)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	access := h.service.AuthorizePublicSubmission(r.Context(), srv, t, token)
	if access.NotFound {
		httpx.WriteError(w, r, http.StatusNotFound, "Ticket not found")
		return
	}
	if !access.Allowed {
		httpx.WriteError(w, r, http.StatusForbidden, "Email verification required")
		return
	}
	if t.Locked || t.Status.IsTerminal() {
		httpx.WriteError(w, r, http.StatusForbidden, "Ticket is closed and cannot be resubmitted")
		return
	}
	saved, err := h.service.SubmitExistingTicketForm(r.Context(), srv, t, &req, access.Verified)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.SubmitPublicTicketResponse{
		Success: true,
		Message: "Ticket submitted successfully",
		Ticket: &modlv1.SubmitPublicTicketResponse_SubmitPublicTicketInfo{
			Id:      saved.ID,
			Subject: nonEmpty(saved.Subject, "No Subject"),
			Status:  string(saved.Status),
		},
	})
}

func (h *Handler) requestPublicVerification(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	emailHint, err := h.service.RequestVerificationCode(r.Context(), srv, id)
	if err != nil {
		writeServiceErr(w, r, err)
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
	var req modlv1.VerifyTicketCodeRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	token, err := h.service.VerifyCode(r.Context(), srv, id, req.Code)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PublicTicketVerificationResponse{
		Success: true,
		Token:   token,
		Message: "Verification successful",
	})
}

// --- Minecraft Tickets V1 & V3 ---

func toMinecraftListItemJSON(t *Ticket) MinecraftTicketListItemJSON {
	hasStaff := false
	for _, r := range t.Replies {
		if r.Staff {
			hasStaff = true
			break
		}
	}
	assigned := t.AssignedTo
	if assigned == nil {
		assigned = []string{}
	}
	return MinecraftTicketListItemJSON{
		ID:               t.ID,
		Type:             string(t.Type),
		Category:         t.Type.DisplayName(),
		Subject:          nonEmpty(t.Subject, "No Subject"),
		Status:           string(t.Status),
		PlayerName:       nonEmpty(t.CreatorName, "Unknown"),
		PlayerUUID:       t.CreatorUUID,
		Priority:         string(t.Priority),
		AssignedTo:       assigned,
		CreatedAt:        MsToTime(t.CreatedAtMs),
		UpdatedAt:        MsToTime(t.UpdatedAtMs),
		HasStaffResponse: hasStaff,
		ReplyCount:       len(t.Replies),
		Locked:           t.Locked,
	}
}

func toMinecraftListItemPB(t *Ticket) *modlv1.MinecraftTicketListItem {
	hasStaff := false
	for _, r := range t.Replies {
		if r.Staff {
			hasStaff = true
			break
		}
	}
	var updPtr *int64
	if t.UpdatedAtMs > 0 {
		u := t.UpdatedAtMs
		updPtr = &u
	}
	return &modlv1.MinecraftTicketListItem{
		Id:               t.ID,
		Type:             string(t.Type),
		Category:         t.Type.DisplayName(),
		Subject:          nonEmpty(t.Subject, "No Subject"),
		Status:           string(t.Status),
		PlayerName:       nonEmpty(t.CreatorName, "Unknown"),
		PlayerUuid:       t.CreatorUUID,
		Priority:         string(t.Priority),
		AssignedTo:       t.AssignedTo,
		CreatedAt:        t.CreatedAtMs,
		UpdatedAt:        updPtr,
		HasStaffResponse: hasStaff,
		Locked:           t.Locked,
		ReplyCount:       int32(len(t.Replies)),
	}
}

func (h *Handler) createMinecraftTicketV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req MinecraftCreateTicketRequestJSON
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid JSON request")
		return
	}
	t, err := h.service.CreateMinecraftTicket(r.Context(), srv, req, false)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, MinecraftTicketCreatedJSON{
		Status:   200,
		Success:  true,
		TicketID: t.ID,
		Message:  "Ticket created successfully",
	})
}

func (h *Handler) createUnfinishedMinecraftTicketV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req MinecraftCreateTicketRequestJSON
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid JSON request")
		return
	}
	t, err := h.service.CreateMinecraftTicket(r.Context(), srv, req, true)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, MinecraftTicketCreatedJSON{
		Status:   200,
		Success:  true,
		TicketID: t.ID,
		Message:  "Ticket draft created - complete the form on the panel",
	})
}

func (h *Handler) getMinecraftTicketsV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	q := r.URL.Query()
	limit := httpx.QueryInt(r, "limit", 50, 1, 100)
	tickets, err := h.store.FindMinecraftTickets(r.Context(), srv.ID, q.Get("status"), q.Get("type"), limit)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	items := make([]MinecraftTicketListItemJSON, 0, len(tickets))
	for _, t := range tickets {
		items = append(items, toMinecraftListItemJSON(t))
	}
	httpx.WriteJSON(w, http.StatusOK, MinecraftTicketListJSON{Status: 200, Tickets: items})
}

func (h *Handler) getMinecraftTicketDetailV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	t, err := h.store.FindByID(r.Context(), srv.ID, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteJSON(w, http.StatusNotFound, MinecraftNotFoundJSON{Status: 404, Message: "Ticket not found"})
			return
		}
		writeServiceErr(w, r, err)
		return
	}
	replies := make([]MinecraftTicketDetailReplyJSON, 0, len(t.Replies))
	for _, rep := range t.Replies {
		replies = append(replies, MinecraftTicketDetailReplyJSON{
			ID:         rep.ID,
			Content:    rep.Content,
			AuthorName: nonEmpty(rep.Name, "Unknown"),
			AuthorID:   rep.CreatorIdentifier,
			IsStaff:    rep.Staff,
			CreatedAt:  MsToTime(rep.CreatedMs),
		})
	}
	var chats []MinecraftChatMessageJSON
	for _, cm := range t.ChatMessages {
		chats = append(chats, MinecraftChatMessageJSON{
			Content:   cm.Content,
			Timestamp: MsToTime(cm.TimestampMs),
			Sender:    cm.Sender,
		})
	}
	assigned := t.AssignedTo
	if assigned == nil {
		assigned = []string{}
	}
	httpx.WriteJSON(w, http.StatusOK, MinecraftTicketDetailJSON{
		Status: 200,
		Ticket: MinecraftTicketDetailViewJSON{
			ID:           t.ID,
			Type:         string(t.Type),
			Category:     t.Type.DisplayName(),
			Subject:      nonEmpty(t.Subject, "No Subject"),
			Status:       string(t.Status),
			PlayerName:   nonEmpty(t.CreatorName, "Unknown"),
			PlayerUUID:   t.CreatorUUID,
			Priority:     string(t.Priority),
			AssignedTo:   assigned,
			CreatedAt:    MsToTime(t.CreatedAtMs),
			UpdatedAt:    MsToTime(t.UpdatedAtMs),
			Locked:       t.Locked,
			Replies:      replies,
			ChatMessages: chats,
		},
	})
}

func (h *Handler) getPlayerTicketsV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	uuid := chi.URLParam(r, "uuid")
	tickets, err := h.store.FindRecentByCreator(r.Context(), srv.ID, uuid, 50)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	items := make([]MinecraftPlayerTicketJSON, 0, len(tickets))
	for _, t := range tickets {
		items = append(items, MinecraftPlayerTicketJSON{
			ID:        t.ID,
			Type:      string(t.Type),
			Category:  t.Type.DisplayName(),
			Subject:   nonEmpty(t.Subject, "No Subject"),
			Status:    string(t.Status),
			CreatedAt: MsToTime(t.CreatedAtMs),
		})
	}
	httpx.WriteJSON(w, http.StatusOK, MinecraftPlayerTicketListJSON{Status: 200, Tickets: items})
}

func (h *Handler) claimMinecraftTicketV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	var req MinecraftClaimTicketRequestJSON
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid JSON request")
		return
	}
	t, err := h.service.ClaimTicket(r.Context(), srv, id, req.PlayerUUID, req.PlayerName)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteJSON(w, http.StatusNotFound, MinecraftTicketClaimJSON{Status: 404, Success: false, Message: "Ticket not found"})
			return
		}
		if _, ok := err.(ConflictError); ok {
			httpx.WriteJSON(w, http.StatusConflict, MinecraftTicketClaimJSON{Status: 409, Success: false, Message: "Ticket is already linked to a Minecraft account"})
			return
		}
		writeServiceErr(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, MinecraftTicketClaimJSON{
		Status:   200,
		Success:  true,
		Message:  "Ticket successfully linked to your account",
		TicketID: t.ID,
		Subject:  t.Subject,
	})
}

func (h *Handler) getTicketsByIdsV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid JSON request")
		return
	}
	if len(req.IDs) == 0 {
		httpx.WriteJSON(w, http.StatusOK, MinecraftTicketLookupListJSON{Status: 200, Tickets: []MinecraftTicketLookupJSON{}})
		return
	}
	tickets, err := h.store.FindByIDs(r.Context(), srv.ID, req.IDs)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	items := make([]MinecraftTicketLookupJSON, 0, len(tickets))
	for _, t := range tickets {
		var firstReply *string
		if len(t.Replies) > 0 {
			c := t.Replies[0].Content
			firstReply = &c
		}
		items = append(items, MinecraftTicketLookupJSON{
			ID:                t.ID,
			Type:              string(t.Type),
			Category:          t.Type.DisplayName(),
			Subject:           nonEmpty(t.Subject, "No Subject"),
			Status:            string(t.Status),
			PlayerName:        nonEmpty(t.CreatorName, "Unknown"),
			PlayerUUID:        t.CreatorUUID,
			CreatedAt:         MsToTime(t.CreatedAtMs),
			FirstReplyContent: firstReply,
		})
	}
	httpx.WriteJSON(w, http.StatusOK, MinecraftTicketLookupListJSON{Status: 200, Tickets: items})
}

// --- Minecraft Tickets V3 (Protobuf) ---

func mcProtoToJSONReq(req *modlv1.MinecraftCreateTicketRequest) MinecraftCreateTicketRequestJSON {
	return MinecraftCreateTicketRequestJSON{
		CreatorUUID:        req.CreatorUuid,
		CreatorName:        req.GetCreatorName(),
		Type:               req.Type,
		Subject:            req.GetSubject(),
		Description:        req.GetDescription(),
		ReportedPlayerUUID: req.GetReportedPlayerUuid(),
		ReportedPlayerName: req.GetReportedPlayerName(),
		ChatMessages:       req.ChatMessages,
		Tags:               req.Tags,
		Priority:           req.GetPriority(),
		CreatedServer:      req.GetCreatedServer(),
		ReplayURL:          req.GetReplayUrl(),
	}
}

func (h *Handler) createMinecraftTicketV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.MinecraftCreateTicketRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	t, err := h.service.CreateMinecraftTicket(r.Context(), srv, mcProtoToJSONReq(&req), false)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.MinecraftCreateTicketResponse{
		Status:   200,
		Success:  true,
		TicketId: t.ID,
		Message:  "Ticket created successfully",
	})
}

func (h *Handler) createUnfinishedMinecraftTicketV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.MinecraftCreateTicketRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	t, err := h.service.CreateMinecraftTicket(r.Context(), srv, mcProtoToJSONReq(&req), true)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.MinecraftCreateTicketResponse{
		Status:   200,
		Success:  true,
		TicketId: t.ID,
		Message:  "Ticket draft created - complete the form on the panel",
	})
}

func (h *Handler) getMinecraftTicketsV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	q := r.URL.Query()
	limit := httpx.QueryInt(r, "limit", 50, 1, 100)
	tickets, err := h.store.FindMinecraftTickets(r.Context(), srv.ID, q.Get("status"), q.Get("type"), limit)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	items := make([]*modlv1.MinecraftTicketListItem, 0, len(tickets))
	for _, t := range tickets {
		items = append(items, toMinecraftListItemPB(t))
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.TicketsResponse{
		Status:  200,
		Tickets: items,
	})
}

func (h *Handler) getMinecraftTicketDetailV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	t, err := h.store.FindByID(r.Context(), srv.ID, id)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	replies := make([]*modlv1.MinecraftTicketDetailReply, 0, len(t.Replies))
	for _, rep := range t.Replies {
		replies = append(replies, &modlv1.MinecraftTicketDetailReply{
			Id:         rep.ID,
			Content:    rep.Content,
			AuthorName: nonEmpty(rep.Name, "Unknown"),
			AuthorId:   rep.CreatorIdentifier,
			IsStaff:    rep.Staff,
			CreatedAt:  MsToTimestamp(rep.CreatedMs),
		})
	}
	var chats []*structpb.Struct
	for _, cm := range t.ChatMessages {
		m := map[string]any{
			"content":   cm.Content,
			"timestamp": cm.TimestampMs,
		}
		if cm.Sender != "" {
			m["sender"] = cm.Sender
		}
		chats = append(chats, MapToStruct(m))
	}
	var updTS *timestamppb.Timestamp
	if t.UpdatedAtMs > 0 {
		updTS = MsToTimestamp(t.UpdatedAtMs)
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.MinecraftTicketDetailResponse{
		Status: 200,
		Ticket: &modlv1.MinecraftTicketDetail{
			Id:           t.ID,
			Type:         string(t.Type),
			Category:     t.Type.DisplayName(),
			Subject:      nonEmpty(t.Subject, "No Subject"),
			Status:       string(t.Status),
			PlayerName:   nonEmpty(t.CreatorName, "Unknown"),
			PlayerUuid:   t.CreatorUUID,
			Priority:     string(t.Priority),
			AssignedTo:   t.AssignedTo,
			CreatedAt:    MsToTimestamp(t.CreatedAtMs),
			UpdatedAt:    updTS,
			Locked:       t.Locked,
			Replies:      replies,
			ChatMessages: chats,
		},
	})
}

func (h *Handler) getPlayerTicketsV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	uuid := chi.URLParam(r, "uuid")
	tickets, err := h.store.FindRecentByCreator(r.Context(), srv.ID, uuid, 50)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	items := make([]*modlv1.MinecraftTicketListItem, 0, len(tickets))
	for _, t := range tickets {
		items = append(items, &modlv1.MinecraftTicketListItem{
			Id:        t.ID,
			Type:      string(t.Type),
			Category:  t.Type.DisplayName(),
			Subject:   nonEmpty(t.Subject, "No Subject"),
			Status:    string(t.Status),
			CreatedAt: t.CreatedAtMs,
		})
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.TicketsResponse{
		Status:  200,
		Tickets: items,
	})
}

func (h *Handler) claimMinecraftTicketV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	var req modlv1.MinecraftClaimTicketRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	t, err := h.service.ClaimTicket(r.Context(), srv, id, req.PlayerUuid, req.PlayerName)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.ClaimTicketResponse{
		Status:   200,
		Success:  true,
		Message:  "Ticket successfully linked to your account",
		TicketId: t.ID,
		Subject:  t.Subject,
	})
}

func (h *Handler) getTicketsByIdsV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.MinecraftTicketsByIdsRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	tickets, err := h.store.FindByIDs(r.Context(), srv.ID, req.Ids)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	items := make([]*modlv1.MinecraftTicketListItem, 0, len(tickets))
	for _, t := range tickets {
		var firstReply *string
		if len(t.Replies) > 0 {
			c := t.Replies[0].Content
			firstReply = &c
		}
		items = append(items, &modlv1.MinecraftTicketListItem{
			Id:                t.ID,
			Type:              string(t.Type),
			Category:          t.Type.DisplayName(),
			Subject:           nonEmpty(t.Subject, "No Subject"),
			Status:            string(t.Status),
			PlayerName:        nonEmpty(t.CreatorName, "Unknown"),
			PlayerUuid:        t.CreatorUUID,
			CreatedAt:         t.CreatedAtMs,
			FirstReplyContent: firstReply,
		})
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.TicketsResponse{
		Status:  200,
		Tickets: items,
	})
}

// --- Minecraft Reports V1 & V3 ---

func toMinecraftReportViewJSON(t *Ticket) MinecraftReportViewJSON {
	content := ""
	if len(t.Replies) > 0 {
		content = t.Replies[0].Content
	}
	var chats []MinecraftChatMessageJSON
	for _, cm := range t.ChatMessages {
		chats = append(chats, MinecraftChatMessageJSON{
			Content:   cm.Content,
			Timestamp: MsToTime(cm.TimestampMs),
			Sender:    cm.Sender,
		})
	}
	assigned := t.AssignedTo
	if assigned == nil {
		assigned = []string{}
	}
	return MinecraftReportViewJSON{
		ID:                 t.ID,
		Type:               string(t.Type),
		ReporterName:       nonEmpty(t.CreatorName, "Unknown"),
		ReporterUUID:       t.CreatorUUID,
		ReportedPlayerUUID: t.ReportedPlayerUUID,
		ReportedPlayerName: t.ReportedPlayer,
		Subject:            nonEmpty(t.Subject, "No Subject"),
		Content:            content,
		Status:             string(t.Status),
		Priority:           string(t.Priority),
		CreatedAt:          MsToTime(t.CreatedAtMs),
		AssignedTo:         assigned,
		ChatMessages:       chats,
	}
}

func toReportEntryPB(t *Ticket) *modlv1.ReportEntry {
	content := ""
	if len(t.Replies) > 0 {
		content = t.Replies[0].Content
	}
	var chats []*structpb.Struct
	for _, cm := range t.ChatMessages {
		m := map[string]any{
			"content":   cm.Content,
			"timestamp": cm.TimestampMs,
		}
		if cm.Sender != "" {
			m["sender"] = cm.Sender
		}
		chats = append(chats, MapToStruct(m))
	}
	return &modlv1.ReportEntry{
		Id:                 t.ID,
		Type:               string(t.Type),
		Category:           string(t.Type),
		ReporterName:       nonEmpty(t.CreatorName, "Unknown"),
		ReporterUuid:       t.CreatorUUID,
		ReportedPlayerUuid: t.ReportedPlayerUUID,
		ReportedPlayerName: t.ReportedPlayer,
		Subject:            nonEmpty(t.Subject, "No Subject"),
		Content:            content,
		Status:             string(t.Status),
		Priority:           string(t.Priority),
		CreatedAt:          t.CreatedAtMs,
		AssignedTo:         t.AssignedTo,
		ChatMessages:       chats,
	}
}

func (h *Handler) getMinecraftReportsV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	q := r.URL.Query()
	limit := httpx.QueryInt(r, "limit", 50, 1, 100)
	tickets, err := h.store.FindReports(r.Context(), srv.ID, q.Get("status"), q.Get("playerUuid"), limit, true)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	reports := make([]MinecraftReportViewJSON, 0, len(tickets))
	for _, t := range tickets {
		reports = append(reports, toMinecraftReportViewJSON(t))
	}
	httpx.WriteJSON(w, http.StatusOK, MinecraftReportListJSON{Status: 200, Reports: reports})
}

func (h *Handler) getMinecraftPlayerReportsV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	uuid := chi.URLParam(r, "uuid")
	q := r.URL.Query()
	limit := httpx.QueryInt(r, "limit", 50, 1, 100)
	tickets, err := h.store.FindReports(r.Context(), srv.ID, q.Get("status"), uuid, limit, false)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	reports := make([]MinecraftReportViewJSON, 0, len(tickets))
	for _, t := range tickets {
		reports = append(reports, toMinecraftReportViewJSON(t))
	}
	httpx.WriteJSON(w, http.StatusOK, MinecraftReportListJSON{Status: 200, Reports: reports})
}

func (h *Handler) assignMinecraftReportV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	var req AssignReportRequestJSON
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid JSON request")
		return
	}
	_, err := h.service.AssignReport(r.Context(), srv, id, req.Assignee)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteJSON(w, http.StatusNotFound, MinecraftNotFoundJSON{Status: 404, Message: "Report not found"})
			return
		}
		writeServiceErr(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, MinecraftReportOperationJSON{Status: 200, Success: true, Message: "Report assigned"})
}

func (h *Handler) dismissMinecraftReportV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	var req DismissReportRequestJSON
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid JSON request")
		return
	}
	_, err := h.service.DismissReport(r.Context(), srv, id, req.DismissedBy, req.Reason)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteJSON(w, http.StatusNotFound, MinecraftNotFoundJSON{Status: 404, Message: "Report not found"})
			return
		}
		writeServiceErr(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, MinecraftReportOperationJSON{Status: 200, Success: true, Message: "Report dismissed"})
}

func (h *Handler) resolveMinecraftReportV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	var req ResolveReportRequestJSON
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid JSON request")
		return
	}
	_, err := h.service.ResolveReport(r.Context(), srv, id, req.ResolvedBy, req.Resolution, req.PunishmentID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteJSON(w, http.StatusNotFound, MinecraftNotFoundJSON{Status: 404, Message: "Report not found"})
			return
		}
		writeServiceErr(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, MinecraftReportOperationJSON{Status: 200, Success: true, Message: "Report resolved"})
}

func (h *Handler) getMinecraftReportsV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	q := r.URL.Query()
	limit := httpx.QueryInt(r, "limit", 50, 1, 100)
	tickets, err := h.store.FindReports(r.Context(), srv.ID, q.Get("status"), q.Get("playerUuid"), limit, true)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	entries := make([]*modlv1.ReportEntry, 0, len(tickets))
	for _, t := range tickets {
		entries = append(entries, toReportEntryPB(t))
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.ReportsResponse{Status: 200, Reports: entries})
}

func (h *Handler) getMinecraftPlayerReportsV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	uuid := chi.URLParam(r, "uuid")
	q := r.URL.Query()
	limit := httpx.QueryInt(r, "limit", 50, 1, 100)
	tickets, err := h.store.FindReports(r.Context(), srv.ID, q.Get("status"), uuid, limit, false)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	entries := make([]*modlv1.ReportEntry, 0, len(tickets))
	for _, t := range tickets {
		entries = append(entries, toReportEntryPB(t))
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.ReportsResponse{Status: 200, Reports: entries})
}

func (h *Handler) assignMinecraftReportV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	var req modlv1.AssignReportRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	_, err := h.service.AssignReport(r.Context(), srv, id, req.Assignee)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.MinecraftReportOperationResponse{
		Status:  200,
		Success: true,
		Message: "Report assigned",
	})
}

func (h *Handler) dismissMinecraftReportV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	var req modlv1.DismissReportRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	_, err := h.service.DismissReport(r.Context(), srv, id, req.DismissedBy, req.GetReason())
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.MinecraftReportOperationResponse{
		Status:  200,
		Success: true,
		Message: "Report dismissed",
	})
}

func (h *Handler) resolveMinecraftReportV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	var req modlv1.ResolveReportRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	_, err := h.service.ResolveReport(r.Context(), srv, id, req.ResolvedBy, req.Resolution, req.PunishmentId)
	if err != nil {
		writeServiceErr(w, r, err)
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.MinecraftReportOperationResponse{
		Status:  200,
		Success: true,
		Message: "Report resolved",
	})
}
