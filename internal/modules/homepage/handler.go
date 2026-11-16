package homepage

import (
	"errors"
	"net/http"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/httpx"
	"modl-go/internal/models"
	"modl-go/internal/realtime"

	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var PanelPolicy = models.AccessPolicy{
	Rule:   models.RuleRequirePermission,
	View:   "admin.settings.view.content",
	Modify: "admin.settings.modify.content",
}

type Handler struct {
	svc *Service
	hub *realtime.Hub
}

func NewHandler(svc *Service, hub *realtime.Hub) *Handler {
	return &Handler{svc: svc, hub: hub}
}

func (h *Handler) PanelRouter(perms models.PermissionChecker) chi.Router {
	r := chi.NewRouter()
	h.RegisterPanel(r, perms)
	return r
}

func (h *Handler) RegisterPanel(r chi.Router, perms models.PermissionChecker) {
	r.Group(func(r chi.Router) {
		r.Use(httpx.RequirePanelAuth)
		if perms != nil {
			r.Use(httpx.RequirePanelPolicy(perms, PanelPolicy))
		}

		r.Get("/", h.panelGetCards)
		r.Post("/", h.panelCreateCard)
		r.Put("/reorder", h.panelReorderCards)
		r.Put("/{id}", h.panelUpdateCard)
		r.Delete("/{id}", h.panelDeleteCard)
	})
}

func (h *Handler) PublicRouter() chi.Router {
	r := chi.NewRouter()
	h.RegisterPublic(r)
	return r
}

func (h *Handler) RegisterPublic(r chi.Router) {
	r.Get("/", h.publicGetCards)
}

func (h *Handler) panelGetCards(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	cards, err := h.svc.GetAllCardsEnriched(r.Context(), srv.ID)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load homepage cards")
		return
	}
	resp := &modlv1.PanelHomepageCardsResponse{
		Cards: make([]*modlv1.PanelHomepageCardResponse, 0, len(cards)),
	}
	for _, c := range cards {
		pb := toPanelCardProto(c.Card)
		if c.Category != nil {
			pb.Category = &modlv1.PanelHomepageCardResponse_EmbeddedCategory{
				Id:   c.Category.ID,
				Name: c.Category.Name,
				Slug: c.Category.Slug,
			}
		}
		resp.Cards = append(resp.Cards, pb)
	}
	httpx.WriteProto(w, r, http.StatusOK, resp)
}

func (h *Handler) panelCreateCard(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	var req modlv1.CreateCardRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	card, err := h.svc.CreateCard(r.Context(), srv.ID, req.Title, CardInput{
		Description:      req.Description,
		Icon:             req.Icon,
		IconColor:        req.IconColor,
		ActionType:       req.ActionType,
		ActionURL:        req.ActionUrl,
		ActionButtonText: req.ActionButtonText,
		CategoryID:       req.CategoryId,
		BackgroundColor:  req.BackgroundColor,
		IsEnabled:        req.IsEnabled,
	})
	if errors.Is(err, ErrInvalidURL) {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to create homepage card")
		return
	}
	h.invalidate(srv.ID, card.ID)
	httpx.WriteProto(w, r, http.StatusCreated, toPanelCardProto(*card))
}

func (h *Handler) panelUpdateCard(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	id := chi.URLParam(r, "id")
	var req modlv1.UpdateCardRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	card, err := h.svc.UpdateCard(r.Context(), srv.ID, id, CardInput{
		Title:            req.Title,
		Description:      req.Description,
		Icon:             req.Icon,
		IconColor:        req.IconColor,
		ActionType:       req.ActionType,
		ActionURL:        req.ActionUrl,
		ActionButtonText: req.ActionButtonText,
		CategoryID:       req.CategoryId,
		BackgroundColor:  req.BackgroundColor,
		IsEnabled:        req.IsEnabled,
	})
	if errors.Is(err, ErrInvalidURL) {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, ErrNotFound) {
		httpx.WriteError(w, r, http.StatusNotFound, "Card not found")
		return
	}
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to update homepage card")
		return
	}
	h.invalidate(srv.ID, card.ID)
	httpx.WriteProto(w, r, http.StatusOK, toPanelCardProto(*card))
}

func (h *Handler) panelDeleteCard(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	id := chi.URLParam(r, "id")
	deleted, err := h.svc.DeleteCard(r.Context(), srv.ID, id)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to delete homepage card")
		return
	}
	if !deleted {
		httpx.WriteError(w, r, http.StatusNotFound, "Card not found")
		return
	}
	h.invalidate(srv.ID, id)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.HomepageCardMutationResponse{Message: "Card deleted"})
}

func (h *Handler) panelReorderCards(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	var req modlv1.ReorderRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.svc.ReorderCards(r.Context(), srv.ID, req.Ids); err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to reorder homepage cards")
		return
	}
	h.invalidate(srv.ID, "")
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.HomepageCardMutationResponse{Message: "Cards reordered"})
}

func (h *Handler) publicGetCards(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	cards, err := h.svc.GetVisibleCardsEnrichedWithArticles(r.Context(), srv.ID)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load homepage cards")
		return
	}
	resp := &modlv1.PublicHomepageCardsResponse{
		Cards: make([]*modlv1.HomepageCardResponse, 0, len(cards)),
	}
	for _, item := range cards {
		c := item.Card
		pb := &modlv1.HomepageCardResponse{
			Id:               c.ID,
			Title:            c.Title,
			Description:      c.Description,
			Icon:             c.Icon,
			IconColor:        c.IconColor,
			ActionType:       c.ActionType,
			ActionUrl:        c.ActionURL,
			ActionButtonText: c.ActionButtonText,
			CategoryId:       c.CategoryID,
			BackgroundColor:  c.BackgroundColor,
			Ordinal:          c.Ordinal,
			IsEnabled:        c.IsEnabled,
			CreatedAt:        timestamppb.New(c.CreatedAt),
			UpdatedAt:        timestamppb.New(c.UpdatedAt),
		}
		if item.Category != nil {
			stubs := make([]*modlv1.HomepageCardResponse_ArticleStub, 0, len(item.Articles))
			for _, a := range item.Articles {
				stubs = append(stubs, &modlv1.HomepageCardResponse_ArticleStub{
					Id:      a.ID,
					Title:   a.Title,
					Slug:    a.Slug,
					Ordinal: a.Ordinal,
				})
			}
			pb.Category = &modlv1.HomepageCardResponse_EmbeddedCategory{
				Id:          item.Category.ID,
				Name:        item.Category.Name,
				Slug:        item.Category.Slug,
				Description: item.Category.Description,
				Articles:    stubs,
			}
		}
		resp.Cards = append(resp.Cards, pb)
	}
	httpx.WriteProto(w, r, http.StatusOK, resp)
}

func (h *Handler) invalidate(serverID, resourceID string) {
	if h.hub != nil {
		h.hub.InvalidatePanel(serverID, modlv1.PanelResource_PANEL_RESOURCE_HOMEPAGE, resourceID)
	}
}

func toPanelCardProto(c Card) *modlv1.PanelHomepageCardResponse {
	return &modlv1.PanelHomepageCardResponse{
		Id:               c.ID,
		Title:            c.Title,
		Description:      c.Description,
		Icon:             c.Icon,
		IconColor:        c.IconColor,
		ActionType:       c.ActionType,
		ActionUrl:        c.ActionURL,
		ActionButtonText: c.ActionButtonText,
		CategoryId:       c.CategoryID,
		BackgroundColor:  c.BackgroundColor,
		IsEnabled:        c.IsEnabled,
		Ordinal:          c.Ordinal,
		CreatedAt:        timestamppb.New(c.CreatedAt),
		UpdatedAt:        timestamppb.New(c.UpdatedAt),
	}
}
