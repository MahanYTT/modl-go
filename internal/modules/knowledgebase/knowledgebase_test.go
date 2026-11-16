package knowledgebase_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/db"
	"modl-go/internal/httpx"
	"modl-go/internal/models"
	"modl-go/internal/modules/knowledgebase"

	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type stubPerms struct {
	superAdmins map[string]bool
	roles       map[string]string
	perms       map[string]map[string]bool
}

func (s *stubPerms) IsSuperAdmin(srv *models.Server, email string) bool {
	return s.superAdmins[email]
}
func (s *stubPerms) IsAuthorizedEmail(ctx context.Context, srv *models.Server, email string) bool {
	return true
}
func (s *stubPerms) PanelRoleID(ctx context.Context, srv *models.Server, email string) string {
	return s.roles[email]
}
func (s *stubPerms) HasPermission(ctx context.Context, srv *models.Server, roleID, perm string) bool {
	return s.perms[roleID][perm]
}
func (s *stubPerms) HasAnyWithPrefix(ctx context.Context, srv *models.Server, roleID, prefix string) bool {
	return false
}

func setupKBRouter(t *testing.T) (chi.Router, *knowledgebase.Service) {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	svc := knowledgebase.NewService(database)
	h := knowledgebase.NewHandler(svc, nil)

	perms := &stubPerms{
		superAdmins: map[string]bool{"admin@example.com": true},
		roles: map[string]string{
			"viewer@example.com":   "role-view",
			"editor@example.com":   "role-edit",
			"nobody@example.com":   "role-none",
		},
		perms: map[string]map[string]bool{
			"role-view": {"admin.settings.view.content": true},
			"role-edit": {"admin.settings.view.content": true, "admin.settings.modify.content": true},
			"role-none": {},
		},
	}

	r := chi.NewRouter()
	r.Mount("/v1/panel/knowledgebase", h.PanelRouter(perms))
	r.Mount("/v1/public/knowledgebase", h.PublicRouter())
	return r, svc
}

func doReq(r http.Handler, method, path string, body any, srv *models.Server, email string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		switch v := body.(type) {
		case proto.Message:
			b, _ := protojson.Marshal(v)
			buf.Write(b)
		case string:
			buf.WriteString(v)
		default:
			_ = json.NewEncoder(&buf).Encode(v)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	ctx := req.Context()
	if srv != nil {
		ctx = httpx.WithServer(ctx, srv)
	}
	if email != "" {
		ctx = httpx.WithSession(ctx, &models.AuthSession{ID: "sess1", ServerID: srv.ID, Email: email})
	}
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

func TestKnowledgebaseRBACAndCRUD(t *testing.T) {
	r, _ := setupKBRouter(t)
	srv1 := &models.Server{ID: "srv-1", ServerName: "Server 1"}
	srv2 := &models.Server{ID: "srv-2", ServerName: "Server 2"}

	// Unauthenticated -> 401
	if rr := doReq(r, http.MethodGet, "/v1/panel/knowledgebase/categories", nil, srv1, ""); rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
	// No permission -> 403
	if rr := doReq(r, http.MethodGet, "/v1/panel/knowledgebase/categories", nil, srv1, "nobody@example.com"); rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
	// View-only can GET, cannot POST
	if rr := doReq(r, http.MethodGet, "/v1/panel/knowledgebase/categories", nil, srv1, "viewer@example.com"); rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if rr := doReq(r, http.MethodPost, "/v1/panel/knowledgebase/categories", `{"name":"Rules"}`, srv1, "viewer@example.com"); rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for viewer POST, got %d", rr.Code)
	}

	// Create categories with duplicate names -> unique slugs
	rr := doReq(r, http.MethodPost, "/v1/panel/knowledgebase/categories", `{"name":"Getting Started","description":"Start here"}`, srv1, "editor@example.com")
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	var cat1 modlv1.KnowledgebaseCategoryResponse
	if err := protojson.Unmarshal(rr.Body.Bytes(), &cat1); err != nil {
		t.Fatal(err)
	}
	if cat1.Slug != "getting-started" || cat1.Ordinal != 0 || !cat1.IsVisible {
		t.Fatalf("unexpected cat1: %+v", &cat1)
	}

	rr = doReq(r, http.MethodPost, "/v1/panel/knowledgebase/categories", `{"name":"Getting Started!"}`, srv1, "editor@example.com")
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	var cat2 modlv1.KnowledgebaseCategoryResponse
	_ = protojson.Unmarshal(rr.Body.Bytes(), &cat2)
	if cat2.Slug != "getting-started-1" || cat2.Ordinal != 1 {
		t.Fatalf("unexpected cat2: %+v", &cat2)
	}

	// Reorder categories
	rr = doReq(r, http.MethodPut, "/v1/panel/knowledgebase/categories/reorder", map[string]any{"ids": []string{cat2.Id, cat1.Id}}, srv1, "editor@example.com")
	if rr.Code != http.StatusOK {
		t.Fatalf("reorder failed: %d %s", rr.Code, rr.Body.String())
	}

	// Create visible & hidden articles
	rr = doReq(r, http.MethodPost, "/v1/panel/knowledgebase/categories/"+cat1.Id+"/articles", `{"title":"Welcome Guide","content":"Read all rules 100% carefully!","isVisible":true}`, srv1, "editor@example.com")
	if rr.Code != http.StatusCreated {
		t.Fatalf("create article 1 failed: %d %s", rr.Code, rr.Body.String())
	}
	var art1 modlv1.KnowledgebaseArticleResponse
	_ = protojson.Unmarshal(rr.Body.Bytes(), &art1)

	rr = doReq(r, http.MethodPost, "/v1/panel/knowledgebase/categories/"+cat1.Id+"/articles", `{"title":"Welcome Guide","content":"Draft guide","isVisible":false}`, srv1, "editor@example.com")
	if rr.Code != http.StatusCreated {
		t.Fatalf("create article 2 failed: %d %s", rr.Code, rr.Body.String())
	}
	var art2 modlv1.KnowledgebaseArticleResponse
	_ = protojson.Unmarshal(rr.Body.Bytes(), &art2)
	if art2.Slug != "welcome-guide-1" || art2.IsVisible {
		t.Fatalf("unexpected art2: %+v", &art2)
	}

	// Public categories hides hidden articles
	rr = doReq(r, http.MethodGet, "/v1/public/knowledgebase/categories", nil, srv1, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("public categories failed: %d", rr.Code)
	}
	var pubCats modlv1.PublicKnowledgebaseCategoriesResponse
	_ = protojson.Unmarshal(rr.Body.Bytes(), &pubCats)
	if len(pubCats.Categories) != 2 || pubCats.Categories[0].Id != cat2.Id || pubCats.Categories[1].Id != cat1.Id {
		t.Fatalf("unexpected public categories order: %+v", pubCats.Categories)
	}
	if len(pubCats.Categories[1].Articles) != 1 || pubCats.Categories[1].Articles[0].Id != art1.Id {
		t.Fatalf("expected only visible article in cat1, got %+v", pubCats.Categories[1].Articles)
	}

	// Public get article by ID and slug; hidden returns 404; other server returns 404
	if rr := doReq(r, http.MethodGet, "/v1/public/knowledgebase/articles/"+art1.Slug, nil, srv1, ""); rr.Code != http.StatusOK {
		t.Fatalf("get by slug failed: %d", rr.Code)
	}
	if rr := doReq(r, http.MethodGet, "/v1/public/knowledgebase/articles/"+art2.Id, nil, srv1, ""); rr.Code != http.StatusNotFound {
		t.Fatalf("hidden article should be 404, got %d", rr.Code)
	}
	if rr := doReq(r, http.MethodGet, "/v1/public/knowledgebase/articles/"+art1.Id, nil, srv2, ""); rr.Code != http.StatusNotFound {
		t.Fatalf("cross-tenant article should be 404, got %d", rr.Code)
	}

	// Search escapes '%' and is case-insensitive
	rr = doReq(r, http.MethodGet, "/v1/public/knowledgebase/search?q=100%25", nil, srv1, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("search failed: %d", rr.Code)
	}
	var searchResp modlv1.KnowledgebaseSearchResponse
	_ = protojson.Unmarshal(rr.Body.Bytes(), &searchResp)
	if len(searchResp.Articles) != 1 || searchResp.Articles[0].Id != art1.Id {
		t.Fatalf("expected 1 match for '100%%', got %+v", searchResp.Articles)
	}
	if rr := doReq(r, http.MethodGet, "/v1/public/knowledgebase/search", nil, srv1, ""); rr.Code != http.StatusBadRequest {
		t.Fatalf("missing q should be 400, got %d", rr.Code)
	}

	// Delete category cascades articles
	if rr := doReq(r, http.MethodDelete, "/v1/panel/knowledgebase/categories/"+cat1.Id, nil, srv1, "admin@example.com"); rr.Code != http.StatusOK {
		t.Fatalf("delete category failed: %d", rr.Code)
	}
	if rr := doReq(r, http.MethodGet, "/v1/panel/knowledgebase/categories/"+cat1.Id+"/articles/"+art1.Id, nil, srv1, "admin@example.com"); rr.Code != http.StatusNotFound {
		t.Fatalf("expected cascaded article to be 404, got %d", rr.Code)
	}
}
