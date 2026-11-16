package homepage_test

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
	"modl-go/internal/modules/homepage"
	"modl-go/internal/modules/knowledgebase"

	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/encoding/protojson"
)

type stubPerms struct{}

func (s *stubPerms) IsSuperAdmin(srv *models.Server, email string) bool {
	return email == "admin@example.com"
}
func (s *stubPerms) IsAuthorizedEmail(ctx context.Context, srv *models.Server, email string) bool {
	return true
}
func (s *stubPerms) PanelRoleID(ctx context.Context, srv *models.Server, email string) string {
	return email
}
func (s *stubPerms) HasPermission(ctx context.Context, srv *models.Server, roleID, perm string) bool {
	if roleID == "viewer@example.com" {
		return perm == "admin.settings.view.content"
	}
	return false
}
func (s *stubPerms) HasAnyWithPrefix(ctx context.Context, srv *models.Server, roleID, prefix string) bool {
	return false
}

func setupHomepage(t *testing.T) (chi.Router, *knowledgebase.Service) {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	kbSvc := knowledgebase.NewService(database)
	hpSvc := homepage.NewService(database, kbSvc)
	hpHandler := homepage.NewHandler(hpSvc, nil)

	r := chi.NewRouter()
	r.Mount("/v1/panel/homepage-cards", hpHandler.PanelRouter(&stubPerms{}))
	r.Mount("/v1/public/homepage-cards", hpHandler.PublicRouter())
	return r, kbSvc
}

func req(r http.Handler, method, path string, body any, srv *models.Server, email string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		switch v := body.(type) {
		case string:
			buf.WriteString(v)
		default:
			_ = json.NewEncoder(&buf).Encode(v)
		}
	}
	request := httptest.NewRequest(method, path, &buf)
	request.Header.Set("Content-Type", "application/json")
	ctx := request.Context()
	if srv != nil {
		ctx = httpx.WithServer(ctx, srv)
	}
	if email != "" {
		ctx = httpx.WithSession(ctx, &models.AuthSession{ID: "s1", ServerID: srv.ID, Email: email})
	}
	request = request.WithContext(ctx)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, request)
	return rr
}

func TestHomepageCardsCRUDAndEnrichment(t *testing.T) {
	r, kbSvc := setupHomepage(t)
	srv := &models.Server{ID: "srv-1", ServerName: "Test"}

	// Create KB category + visible and hidden articles
	cat, err := kbSvc.CreateCategory(context.Background(), srv.ID, "Appeals", nil)
	if err != nil {
		t.Fatal(err)
	}
	visArt, err := kbSvc.CreateArticle(context.Background(), srv.ID, cat.ID, "How to appeal", "Steps", nil)
	if err != nil {
		t.Fatal(err)
	}
	hidden := false
	_, err = kbSvc.CreateArticle(context.Background(), srv.ID, cat.ID, "Internal", "Secret", &hidden)
	if err != nil {
		t.Fatal(err)
	}

	// RBAC: viewer can GET (both without and with trailing slash), cannot POST
	if rr := req(r, http.MethodGet, "/v1/panel/homepage-cards", nil, srv, "viewer@example.com"); rr.Code != http.StatusOK {
		t.Fatalf("expected 200 without trailing slash, got %d", rr.Code)
	}
	if rr := req(r, http.MethodGet, "/v1/panel/homepage-cards/", nil, srv, "viewer@example.com"); rr.Code != http.StatusOK {
		t.Fatalf("expected 200 with trailing slash, got %d", rr.Code)
	}
	if rr := req(r, http.MethodPost, "/v1/panel/homepage-cards", `{"title":"Card"}`, srv, "viewer@example.com"); rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for viewer POST, got %d", rr.Code)
	}

	// Unsafe URLs rejected
	for _, bad := range []string{"javascript:alert(1)", "//evil.com", `/\evil.com`, "data:text/html,hi"} {
		rr := req(r, http.MethodPost, "/v1/panel/homepage-cards", map[string]any{"title": "Bad", "actionUrl": bad}, srv, "admin@example.com")
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for bad URL %q, got %d", bad, rr.Code)
		}
	}

	// Create dropdown card (enabled) and link card (disabled)
	rr := req(r, http.MethodPost, "/v1/panel/homepage-cards", map[string]any{
		"title":      "Ban Appeals",
		"actionType": "category_dropdown",
		"categoryId": cat.ID,
		"actionUrl":  "/appeals",
	}, srv, "admin@example.com")
	if rr.Code != http.StatusCreated {
		t.Fatalf("create card 1 failed: %d %s", rr.Code, rr.Body.String())
	}
	var card1 modlv1.PanelHomepageCardResponse
	_ = protojson.Unmarshal(rr.Body.Bytes(), &card1)

	rr = req(r, http.MethodPost, "/v1/panel/homepage-cards", map[string]any{
		"title":      "Store",
		"actionType": "link",
		"actionUrl":  "https://store.example.com",
		"isEnabled":  false,
	}, srv, "admin@example.com")
	if rr.Code != http.StatusCreated {
		t.Fatalf("create card 2 failed: %d %s", rr.Code, rr.Body.String())
	}
	var card2 modlv1.PanelHomepageCardResponse
	_ = protojson.Unmarshal(rr.Body.Bytes(), &card2)

	// Panel lists both cards with embedded category on card1
	rr = req(r, http.MethodGet, "/v1/panel/homepage-cards", nil, srv, "admin@example.com")
	if rr.Code != http.StatusOK {
		t.Fatalf("panel get failed: %d", rr.Code)
	}
	var panelResp modlv1.PanelHomepageCardsResponse
	_ = protojson.Unmarshal(rr.Body.Bytes(), &panelResp)
	if len(panelResp.Cards) != 2 || panelResp.Cards[0].Category == nil || panelResp.Cards[0].Category.Slug != "appeals" {
		t.Fatalf("unexpected panel cards: %+v", panelResp.Cards)
	}

	// Public lists only enabled card1, enriched with visible articles
	rr = req(r, http.MethodGet, "/v1/public/homepage-cards", nil, srv, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("public get failed: %d", rr.Code)
	}
	var pubResp modlv1.PublicHomepageCardsResponse
	_ = protojson.Unmarshal(rr.Body.Bytes(), &pubResp)
	if len(pubResp.Cards) != 1 || pubResp.Cards[0].Id != card1.Id || pubResp.Cards[0].Category == nil || len(pubResp.Cards[0].Category.Articles) != 1 || pubResp.Cards[0].Category.Articles[0].Id != visArt.ID {
		t.Fatalf("unexpected public cards: %+v", pubResp.Cards)
	}

	// Update + reorder + delete
	rr = req(r, http.MethodPut, "/v1/panel/homepage-cards/"+card2.Id, map[string]any{"isEnabled": true}, srv, "admin@example.com")
	if rr.Code != http.StatusOK {
		t.Fatalf("update failed: %d", rr.Code)
	}
	rr = req(r, http.MethodPut, "/v1/panel/homepage-cards/reorder", map[string]any{"ids": []string{card2.Id, card1.Id}}, srv, "admin@example.com")
	if rr.Code != http.StatusOK {
		t.Fatalf("reorder failed: %d", rr.Code)
	}
	rr = req(r, http.MethodGet, "/v1/public/homepage-cards", nil, srv, "")
	_ = protojson.Unmarshal(rr.Body.Bytes(), &pubResp)
	if len(pubResp.Cards) != 2 || pubResp.Cards[0].Id != card2.Id || pubResp.Cards[1].Id != card1.Id {
		t.Fatalf("unexpected order after reorder: %+v", pubResp.Cards)
	}
	if rr := req(r, http.MethodDelete, "/v1/panel/homepage-cards/"+card1.Id, nil, srv, "admin@example.com"); rr.Code != http.StatusOK {
		t.Fatalf("delete failed: %d", rr.Code)
	}
}
