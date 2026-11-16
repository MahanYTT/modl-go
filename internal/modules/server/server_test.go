package server_test

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
	"modl-go/internal/modules/role"
	"modl-go/internal/modules/server"

	"github.com/go-chi/chi/v5"
)

func TestBootstrapAndRegistrationFlow(t *testing.T) {
	sqlDB, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer sqlDB.Close()

	cfg := &config.Config{
		ModlDomain:          "modl.local",
		BootstrapServerName: "CraftTest",
		BootstrapDomain:     "play.crafttest.net",
		BootstrapAdminEmail: "owner@crafttest.net",
		AuthCodeHashSecret:  "test-secret",
		SessionCookieName:   "MODL_SESSION",
		SessionDurationSec:  14 * 24 * 3600,
	}
	mailer := email.NewSender(cfg)
	roles := role.NewHandler(sqlDB, nil)
	srvH := server.NewHandler(sqlDB, cfg, mailer, roles)

	boot, err := srvH.BootstrapIfConfigured(context.Background())
	if err != nil || boot == nil {
		t.Fatalf("bootstrap failed: %v", err)
	}
	if boot.ServerName != "CraftTest" || boot.CustomDomain != "play.crafttest.net" {
		t.Fatalf("unexpected bootstrap server: %+v", boot)
	}

	resolved, err := srvH.GetByAPIKey(context.Background(), boot.APIKey)
	if err != nil || resolved.ID != boot.ID {
		t.Fatalf("GetByAPIKey failed: %v", err)
	}

	r := chi.NewRouter()
	r.Route("/v1/public", func(r chi.Router) {
		srvH.RegisterPublicRoutes(r)
	})
	r.Route("/v1/panel", func(r chi.Router) {
		r.Use(httpx.ServerHeaderMiddleware(cfg, srvH))
		srvH.RegisterPanelRoutes(r)
	})

	// CLI registration flow
	body := `{"email":"cli@example.com","serverName":"CliServer","customDomain":"cliserver"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/public/registration/cli", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("cli register status %d: %s", rec.Code, rec.Body.String())
	}
	var cliResp struct {
		Success    bool   `json:"success"`
		SetupToken string `json:"setupToken"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &cliResp); err != nil || !cliResp.Success || cliResp.SetupToken == "" {
		t.Fatalf("unexpected cli response: %s", rec.Body.String())
	}

	// Check CLI setup status returns API key (auto-verified when no SMTP)
	statusBody := `{"token":"` + cliResp.SetupToken + `"}`
	req2 := httptest.NewRequest(http.MethodPost, "/v1/public/registration/cli/status", bytes.NewBufferString(statusBody))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("cli status %d: %s", rec2.Code, rec2.Body.String())
	}
	var stResp struct {
		Success bool   `json:"success"`
		APIKey  string `json:"apiKey"`
	}
	_ = json.Unmarshal(rec2.Body.Bytes(), &stResp)
	if !stResp.Success || stResp.APIKey == "" {
		t.Fatalf("expected apiKey in cli status: %s", rec2.Body.String())
	}
}
