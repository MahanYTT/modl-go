package auth_test

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
	"modl-go/internal/modules/auth"
	"modl-go/internal/modules/role"

	"github.com/go-chi/chi/v5"
)

type stubRes struct{ srv *models.Server }

func (s stubRes) GetByDomain(context.Context, string) (*models.Server, error) { return s.srv, nil }
func (s stubRes) GetByAPIKey(context.Context, string) (*models.Server, error) { return s.srv, nil }
func (s stubRes) GetDefault(context.Context) (*models.Server, error)          { return s.srv, nil }

func TestAuthEmailCodeAndSessionLifecycle(t *testing.T) {
	sqlDB, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer sqlDB.Close()

	srv := &models.Server{ID: "srv1", ServerName: "TestSMP", CustomDomain: "test.local", AdminEmail: "owner@test.local"}
	cfg := &config.Config{
		SingleTenant:       true,
		AuthCodeHashSecret: "secret-123",
		SessionCookieName:  "MODL_SESSION",
		SessionDurationSec: 3600,
	}
	rh := role.NewHandler(sqlDB, nil)
	ah := auth.NewHandler(sqlDB, cfg, email.NewSender(cfg), rh)

	r := chi.NewRouter()
	r.Route("/v1/panel/auth", func(r chi.Router) {
		r.Use(httpx.ServerHeaderMiddleware(cfg, stubRes{srv}))
		r.Use(httpx.SessionAuthMiddleware(cfg, ah))
		ah.RegisterPanelAuthRoutes(r)
	})

	// Request login code for super admin
	req := httptest.NewRequest(http.MethodPost, "/v1/panel/auth/send-email-code", bytes.NewBufferString(`{"email":"owner@test.local"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("send-email-code %d: %s", rec.Code, rec.Body.String())
	}

	// Replace stored auth code with known hash for verification test
	knownHash := httpx.HashOneTimeCode("123456", cfg.AuthCodeHashSecret)
	_, err = sqlDB.Exec(`UPDATE auth_codes SET code_hash = ? WHERE server_id = ? AND email = ?`, knownHash, srv.ID, "owner@test.local")
	if err != nil {
		t.Fatalf("update auth code: %v", err)
	}

	// Verify code -> receives session cookie
	req2 := httptest.NewRequest(http.MethodPost, "/v1/panel/auth/verify-email-code", bytes.NewBufferString(`{"email":"owner@test.local","code":"123456"}`))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("verify-email-code %d: %s", rec2.Code, rec2.Body.String())
	}
	cookies := rec2.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != cfg.SessionCookieName {
		t.Fatalf("expected session cookie, got %+v", cookies)
	}

	// GET /me with session cookie
	req3 := httptest.NewRequest(http.MethodGet, "/v1/panel/auth/me", nil)
	req3.AddCookie(cookies[0])
	rec3 := httptest.NewRecorder()
	r.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Fatalf("GET /me failed %d: %s", rec3.Code, rec3.Body.String())
	}
	var me map[string]any
	_ = json.Unmarshal(rec3.Body.Bytes(), &me)
	if me["email"] != "owner@test.local" || me["role"] != "Super Admin" {
		t.Fatalf("unexpected /me: %s", rec3.Body.String())
	}
}
