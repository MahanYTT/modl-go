package knowledgebase

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

		r.Get("/categories", h.panelGetCategories)
		r.Post("/categories", h.panelCreateCategory)
		r.Put("/categories/reorder", h.panelReorderCategories)
		r.Put("/categories/{id}", h.panelUpdateCategory)
		r.Delete("/categories/{id}", h.panelDeleteCategory)

		r.Get("/categories/{categoryId}/articles", h.panelGetArticles)
		r.Post("/categories/{categoryId}/articles", h.panelCreateArticle)
		r.Put("/categories/{categoryId}/articles/reorder", h.panelReorderArticles)
		r.Get("/categories/{categoryId}/articles/{articleId}", h.panelGetArticle)
		r.Put("/categories/{categoryId}/articles/{articleId}", h.panelUpdateArticle)
		r.Delete("/categories/{categoryId}/articles/{articleId}", h.panelDeleteArticle)
	})
}

func (h *Handler) PublicRouter() chi.Router {
	r := chi.NewRouter()
	h.RegisterPublic(r)
	return r
}

func (h *Handler) RegisterPublic(r chi.Router) {
	r.Get("/categories", h.publicGetCategories)
	r.Get("/categories/{categoryId}/articles", h.publicGetArticles)
	r.Get("/articles/{idOrSlug}", h.publicGetArticle)
	r.Get("/search", h.publicSearchArticles)
}

func (h *Handler) panelGetCategories(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	categories, err := h.svc.GetAllCategories(r.Context(), srv.ID)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load categories")
		return
	}
	articlesByCat, err := h.svc.GetAllArticlesGroupedByCategory(r.Context(), srv.ID)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load articles")
		return
	}

	resp := &modlv1.PanelKnowledgebaseCategoriesResponse{
		Categories: make([]*modlv1.PanelKnowledgebaseCategoryWithArticlesResponse, 0, len(categories)),
	}
	for _, c := range categories {
		arts := articlesByCat[c.ID]
		pbArts := make([]*modlv1.KnowledgebaseArticleResponse, 0, len(arts))
		for _, a := range arts {
			pbArts = append(pbArts, toArticleProto(a))
		}
		resp.Categories = append(resp.Categories, &modlv1.PanelKnowledgebaseCategoryWithArticlesResponse{
			Id:          c.ID,
			Name:        c.Name,
			Slug:        c.Slug,
			Description: c.Description,
			Ordinal:     c.Ordinal,
			IsVisible:   c.IsVisible,
			Articles:    pbArts,
		})
	}
	httpx.WriteProto(w, r, http.StatusOK, resp)
}

func (h *Handler) panelCreateCategory(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	var req modlv1.CreateCategoryRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	cat, err := h.svc.CreateCategory(r.Context(), srv.ID, req.Name, req.Description)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to create category")
		return
	}
	h.invalidate(srv.ID, cat.ID)
	httpx.WriteProto(w, r, http.StatusCreated, toCategoryProto(*cat))
}

func (h *Handler) panelUpdateCategory(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	id := chi.URLParam(r, "id")
	var req modlv1.UpdateCategoryRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	cat, err := h.svc.UpdateCategory(r.Context(), srv.ID, id, req.Name, req.Description, req.IsVisible)
	if errors.Is(err, ErrNotFound) {
		httpx.WriteError(w, r, http.StatusNotFound, "Category not found")
		return
	}
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to update category")
		return
	}
	h.invalidate(srv.ID, cat.ID)
	httpx.WriteProto(w, r, http.StatusOK, toCategoryProto(*cat))
}

func (h *Handler) panelDeleteCategory(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	id := chi.URLParam(r, "id")
	deleted, err := h.svc.DeleteCategory(r.Context(), srv.ID, id)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to delete category")
		return
	}
	if !deleted {
		httpx.WriteError(w, r, http.StatusNotFound, "Category not found")
		return
	}
	h.invalidate(srv.ID, id)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.KnowledgebaseMessageResponse{Message: "Category deleted"})
}

func (h *Handler) panelReorderCategories(w http.ResponseWriter, r *http.Request) {
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
	if err := h.svc.ReorderCategories(r.Context(), srv.ID, req.Ids); err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to reorder categories")
		return
	}
	h.invalidate(srv.ID, "")
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.KnowledgebaseMessageResponse{Message: "Categories reordered"})
}

func (h *Handler) panelGetArticles(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	categoryID := chi.URLParam(r, "categoryId")
	articles, err := h.svc.GetArticlesByCategory(r.Context(), srv.ID, categoryID)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load articles")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, toArticlesProto(articles))
}

func (h *Handler) panelGetArticle(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	articleID := chi.URLParam(r, "articleId")
	article, err := h.svc.GetArticleByID(r.Context(), srv.ID, articleID)
	if errors.Is(err, ErrNotFound) {
		httpx.WriteError(w, r, http.StatusNotFound, "Article not found")
		return
	}
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load article")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, toArticleProto(*article))
}

func (h *Handler) panelCreateArticle(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	categoryID := chi.URLParam(r, "categoryId")
	var req modlv1.CreateArticleRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	article, err := h.svc.CreateArticle(r.Context(), srv.ID, categoryID, req.Title, req.Content, req.IsVisible)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to create article")
		return
	}
	h.invalidate(srv.ID, article.ID)
	httpx.WriteProto(w, r, http.StatusCreated, toArticleProto(*article))
}

func (h *Handler) panelUpdateArticle(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	articleID := chi.URLParam(r, "articleId")
	var req modlv1.UpdateArticleRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	article, err := h.svc.UpdateArticle(r.Context(), srv.ID, articleID, req.Title, req.Content, req.IsVisible)
	if errors.Is(err, ErrNotFound) {
		httpx.WriteError(w, r, http.StatusNotFound, "Article not found")
		return
	}
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to update article")
		return
	}
	h.invalidate(srv.ID, article.ID)
	httpx.WriteProto(w, r, http.StatusOK, toArticleProto(*article))
}

func (h *Handler) panelDeleteArticle(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	articleID := chi.URLParam(r, "articleId")
	deleted, err := h.svc.DeleteArticle(r.Context(), srv.ID, articleID)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to delete article")
		return
	}
	if !deleted {
		httpx.WriteError(w, r, http.StatusNotFound, "Article not found")
		return
	}
	h.invalidate(srv.ID, articleID)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.KnowledgebaseMessageResponse{Message: "Article deleted"})
}

func (h *Handler) panelReorderArticles(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	categoryID := chi.URLParam(r, "categoryId")
	var req modlv1.ReorderRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.svc.ReorderArticles(r.Context(), srv.ID, categoryID, req.Ids); err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to reorder articles")
		return
	}
	h.invalidate(srv.ID, "")
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.KnowledgebaseMessageResponse{Message: "Articles reordered"})
}

func (h *Handler) publicGetCategories(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	categories, err := h.svc.GetVisibleCategories(r.Context(), srv.ID)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load categories")
		return
	}
	catIDs := make([]string, 0, len(categories))
	for _, c := range categories {
		catIDs = append(catIDs, c.ID)
	}
	articlesByCat, err := h.svc.GetVisibleArticlesGroupedByCategoryIDs(r.Context(), srv.ID, catIDs)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load articles")
		return
	}

	resp := &modlv1.PublicKnowledgebaseCategoriesResponse{
		Categories: make([]*modlv1.PublicKnowledgebaseCategoryWithArticlesResponse, 0, len(categories)),
	}
	for _, c := range categories {
		arts := articlesByCat[c.ID]
		stubs := make([]*modlv1.KnowledgebaseArticleStubResponse, 0, len(arts))
		for _, a := range arts {
			stubs = append(stubs, &modlv1.KnowledgebaseArticleStubResponse{
				Id:      a.ID,
				Title:   a.Title,
				Slug:    a.Slug,
				Ordinal: a.Ordinal,
			})
		}
		resp.Categories = append(resp.Categories, &modlv1.PublicKnowledgebaseCategoryWithArticlesResponse{
			Id:          c.ID,
			Name:        c.Name,
			Slug:        c.Slug,
			Description: c.Description,
			Ordinal:     c.Ordinal,
			Articles:    stubs,
		})
	}
	httpx.WriteProto(w, r, http.StatusOK, resp)
}

func (h *Handler) publicGetArticles(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	categoryID := chi.URLParam(r, "categoryId")
	articles, err := h.svc.GetVisibleArticlesByCategory(r.Context(), srv.ID, categoryID)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load articles")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, toArticlesProto(articles))
}

func (h *Handler) publicGetArticle(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	idOrSlug := chi.URLParam(r, "idOrSlug")
	article, err := h.svc.GetArticleByID(r.Context(), srv.ID, idOrSlug)
	if errors.Is(err, ErrNotFound) {
		article, err = h.svc.GetArticleBySlug(r.Context(), srv.ID, idOrSlug)
	}
	if errors.Is(err, ErrNotFound) || (err == nil && !article.IsVisible) {
		httpx.WriteError(w, r, http.StatusNotFound, "Article not found")
		return
	}
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load article")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, toArticleProto(*article))
}

func (h *Handler) publicSearchArticles(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	if !r.URL.Query().Has("q") {
		httpx.WriteError(w, r, http.StatusBadRequest, "Missing query parameter q")
		return
	}
	q := r.URL.Query().Get("q")
	articles, err := h.svc.SearchArticles(r.Context(), srv.ID, q)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to search articles")
		return
	}
	pbArts := make([]*modlv1.KnowledgebaseArticleResponse, 0, len(articles))
	for _, a := range articles {
		pbArts = append(pbArts, toArticleProto(a))
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.KnowledgebaseSearchResponse{Articles: pbArts})
}

func (h *Handler) invalidate(serverID, resourceID string) {
	if h.hub != nil {
		h.hub.InvalidatePanel(serverID, modlv1.PanelResource_PANEL_RESOURCE_KNOWLEDGEBASE, resourceID)
	}
}

func toCategoryProto(c Category) *modlv1.KnowledgebaseCategoryResponse {
	return &modlv1.KnowledgebaseCategoryResponse{
		Id:          c.ID,
		Name:        c.Name,
		Slug:        c.Slug,
		Description: c.Description,
		Ordinal:     c.Ordinal,
		IsVisible:   c.IsVisible,
		CreatedAt:   timestamppb.New(c.CreatedAt),
		UpdatedAt:   timestamppb.New(c.UpdatedAt),
	}
}

func toArticleProto(a Article) *modlv1.KnowledgebaseArticleResponse {
	return &modlv1.KnowledgebaseArticleResponse{
		Id:         a.ID,
		Title:      a.Title,
		Slug:       a.Slug,
		Content:    a.Content,
		CategoryId: a.CategoryID,
		Ordinal:    a.Ordinal,
		IsVisible:  a.IsVisible,
		CreatedAt:  timestamppb.New(a.CreatedAt),
		UpdatedAt:  timestamppb.New(a.UpdatedAt),
	}
}

func toArticlesProto(articles []Article) *modlv1.KnowledgebaseArticlesResponse {
	pb := make([]*modlv1.KnowledgebaseArticleResponse, 0, len(articles))
	for _, a := range articles {
		pb = append(pb, toArticleProto(a))
	}
	return &modlv1.KnowledgebaseArticlesResponse{Articles: pb}
}
