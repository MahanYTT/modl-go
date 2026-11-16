package staff_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"modl-go/internal/config"
	"modl-go/internal/db"
	"modl-go/internal/email"
	"modl-go/internal/httpx"
	"modl-go/internal/models"
	"modl-go/internal/modules/role"
	"modl-go/internal/modules/staff"

	"github.com/go-chi/chi/v5"
)

type stubRes struct{ srv *models.Server }

func (s stubRes) GetByDomain(context.Context, string) (*models.Server, error) { return s.srv, nil }
func (s stubRes) GetByAPIKey(context.Context, string) (*models.Server, error) { return s.srv, nil }
func (s stubRes) GetDefault(context.Context) (*models.Server, error)          { return s.srv, nil }

func TestStaffCRUDAnd2FA(t *testing.T) {
	sqlDB, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer sqlDB.Close()

	srv := &models.Server{ID: "srv1", ServerName: "TestSMP", CustomDomain: "test.local", AdminEmail: "owner@test.local", APIKey: "modl_k1"}
	cfg := &config.Config{SingleTenant: true}
	rh := role.NewHandler(sqlDB, nil)
	sh := staff.NewHandler(sqlDB, cfg, email.NewSender(cfg), nil, rh)

	r := chi.NewRouter()
	r.Route("/v1/panel", func(r chi.Router) {
		r.Use(httpx.ServerHeaderMiddleware(cfg, stubRes{srv}))
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				req = req.WithContext(httpx.WithSession(req.Context(), &models.AuthSession{ID: "s1", ServerID: srv.ID, Email: srv.AdminEmail}))
				next.ServeHTTP(w, req)
			})
		})
		sh.RegisterPanelRoutes(r)
	})
	r.Route("/v1/public", func(r chi.Router) {
		r.Use(httpx.ServerHeaderMiddleware(cfg, stubRes{srv}))
		sh.RegisterPublicRoutes(r)
	})
	r.Route("/v1/minecraft", func(r chi.Router) {
		r.Use(httpx.ApiKeyMiddleware(stubRes{srv}))
		sh.RegisterMinecraftV1Routes(r)
	})

	// Create staff member
	req := httptest.NewRequest(http.MethodPost, "/v1/panel/staff", bytes.NewBufferString(`{"email":"mod@test.local","username":"ModUser","role":"Moderator"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create staff %d: %s", rec.Code, rec.Body.String())
	}

	// Assign Minecraft UUID
	uuid := "123e4567-e89b-12d3-a456-426614174000"
	req2 := httptest.NewRequest(http.MethodPatch, "/v1/panel/staff/mod@test.local/minecraft-player", bytes.NewBufferString(`{"minecraft_uuid":"`+uuid+`","minecraft_username":"ModMC"}`))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("assign mc player %d: %s", rec2.Code, rec2.Body.String())
	}

	// Generate & verify 2FA token
	req3 := httptest.NewRequest(http.MethodPost, "/v1/minecraft/staff/2fa/generate", bytes.NewBufferString(`{"minecraftUuid":"`+uuid+`","ip":"1.2.3.4"}`))
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("X-API-Key", "modl_k1")
	rec3 := httptest.NewRecorder()
	r.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Fatalf("generate 2fa %d: %s", rec3.Code, rec3.Body.String())
	}
	var genResp struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(rec3.Body.Bytes(), &genResp)

	req4 := httptest.NewRequest(http.MethodPost, "/v1/public/staff/2fa/verify/"+genResp.Token, nil)
	rec4 := httptest.NewRecorder()
	r.ServeHTTP(rec4, req4)
	if rec4.Code != http.StatusOK {
		t.Fatalf("verify 2fa %d: %s", rec4.Code, rec4.Body.String())
	}
}
