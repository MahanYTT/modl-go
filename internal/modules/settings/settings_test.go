package settings_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"modl-go/internal/config"
	"modl-go/internal/db"
	"modl-go/internal/httpx"
	"modl-go/internal/models"
	"modl-go/internal/modules/role"
	"modl-go/internal/modules/settings"
	"modl-go/internal/storage"

	"github.com/go-chi/chi/v5"
)

type stubResolver struct{ srv *models.Server }

func (s stubResolver) GetByDomain(context.Context, string) (*models.Server, error) {
	return s.srv, nil
}
func (s stubResolver) GetByAPIKey(context.Context, string) (*models.Server, error) {
	return s.srv, nil
}
func (s stubResolver) GetDefault(context.Context) (*models.Server, error) { return s.srv, nil }

func TestSettingsAndPunishmentTypes(t *testing.T) {
	sqlDB, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer sqlDB.Close()

	store, err := storage.NewLocalStore(sqlDB, t.TempDir())
	if err != nil {
		t.Fatalf("local store: %v", err)
	}
	srv := &models.Server{ID: "srv1", ServerName: "TestSMP", CustomDomain: "testsmp.local", AdminEmail: "admin@test.local", APIKey: "modl_test123"}
	cfg := &config.Config{SingleTenant: true}
	rh := role.NewHandler(sqlDB, nil)
	sh := settings.NewHandler(sqlDB, cfg, nil, rh, store)

	r := chi.NewRouter()
	r.Route("/v1/panel", func(r chi.Router) {
		r.Use(httpx.ServerHeaderMiddleware(cfg, stubResolver{srv}))
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				req = req.WithContext(httpx.WithSession(req.Context(), &models.AuthSession{ID: "s1", ServerID: srv.ID, Email: srv.AdminEmail}))
				next.ServeHTTP(w, req)
			})
		})
		sh.RegisterPanelRoutes(r)
	})
	r.Route("/v1/public", func(r chi.Router) {
		r.Use(httpx.ServerHeaderMiddleware(cfg, stubResolver{srv}))
		sh.RegisterPublicRoutes(r)
	})
	r.Route("/v3/minecraft", func(r chi.Router) {
		r.Use(httpx.ApiKeyMiddleware(stubResolver{srv}))
		sh.RegisterMinecraftV3Routes(r)
	})

	// Patch general settings
	patch := `{"expectedVersion":1,"serverDisplayName":"My Awesome SMP","defaultLanguage":"en"}`
	req := httptest.NewRequest(http.MethodPatch, "/v1/panel/settings/general", bytes.NewBufferString(patch))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("patch general %d: %s", rec.Code, rec.Body.String())
	}

	// Verify public settings reflects updated display name
	req2 := httptest.NewRequest(http.MethodGet, "/v1/public/settings", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("get public settings %d: %s", rec2.Code, rec2.Body.String())
	}
	var pub map[string]any
	_ = json.Unmarshal(rec2.Body.Bytes(), &pub)
	if pub["serverDisplayName"] != "My Awesome SMP" {
		t.Fatalf("unexpected display name: %v", pub["serverDisplayName"])
	}

	// Verify v3 protobuf punishment types endpoint returns binary protobuf
	req3 := httptest.NewRequest(http.MethodGet, "/v3/minecraft/punishments/types", nil)
	req3.Header.Set("X-API-Key", "modl_test123")
	rec3 := httptest.NewRecorder()
	r.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK || rec3.Header().Get("Content-Type") != "application/x-protobuf" {
		t.Fatalf("v3 punishment types failed %d ct=%s", rec3.Code, rec3.Header().Get("Content-Type"))
	}
}
