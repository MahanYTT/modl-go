package log

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"modl-go/internal/db"
	"modl-go/internal/httpx"
	"modl-go/internal/models"

	"github.com/go-chi/chi/v5"
)

type stubPerms struct {
	superAdmin bool
	allow      bool
}

func (p stubPerms) IsSuperAdmin(_ *models.Server, _ string) bool { return p.superAdmin }
func (p stubPerms) IsAuthorizedEmail(_ context.Context, _ *models.Server, _ string) bool {
	return p.allow
}
func (p stubPerms) PanelRoleID(_ context.Context, _ *models.Server, _ string) string { return "mod" }
func (p stubPerms) HasPermission(_ context.Context, _ *models.Server, _, _ string) bool {
	return p.allow
}
func (p stubPerms) HasAnyWithPrefix(_ context.Context, _ *models.Server, _, _ string) bool {
	return p.allow
}

func TestGetLogsAndClamp(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer database.Close()

	svc := NewService(database)
	ctx := context.Background()

	if err := svc.LogAction(ctx, "srv-1", "First action", "info", "Admin"); err != nil {
		t.Fatalf("LogAction 1: %v", err)
	}
	// Ensure distinct timestamps for deterministic order
	_, err = database.ExecContext(ctx,
		`INSERT INTO server_logs (id, server_id, description, level, source, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		"log-2", "srv-1", "Second action", "moderation", "Mod", time.Now().Add(time.Second).UnixMilli(),
	)
	if err != nil {
		t.Fatalf("insert log-2: %v", err)
	}
	if err := svc.LogAction(ctx, "srv-other", "Other server", "info", "Admin"); err != nil {
		t.Fatalf("LogAction other: %v", err)
	}

	h := NewHandler(database, stubPerms{superAdmin: true})
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := httpx.WithServer(r.Context(), &models.Server{ID: "srv-1", AdminEmail: "admin@example.com"})
			c = httpx.WithSession(c, &models.AuthSession{ID: "s1", ServerID: "srv-1", Email: "admin@example.com"})
			next.ServeHTTP(w, r.WithContext(c))
		})
	})
	h.RegisterPanelRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/logs?limit=1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Logs []struct {
			ID          string `json:"id"`
			Description string `json:"description"`
			Level       string `json:"level"`
			Source      string `json:"source"`
		} `json:"logs"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(resp.Logs) != 1 || resp.Logs[0].ID != "log-2" || resp.Logs[0].Description != "Second action" {
		t.Fatalf("unexpected logs payload: %+v", resp.Logs)
	}

	// Verify permission denial
	hDeny := NewHandler(database, stubPerms{superAdmin: false, allow: false})
	rDeny := chi.NewRouter()
	rDeny.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := httpx.WithServer(r.Context(), &models.Server{ID: "srv-1"})
			c = httpx.WithSession(c, &models.AuthSession{ID: "s1", ServerID: "srv-1", Email: "user@example.com"})
			next.ServeHTTP(w, r.WithContext(c))
		})
	})
	hDeny.RegisterPanelRoutes(rDeny)
	reqDeny := httptest.NewRequest(http.MethodGet, "/logs/", nil)
	recDeny := httptest.NewRecorder()
	rDeny.ServeHTTP(recDeny, reqDeny)
	if recDeny.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recDeny.Code)
	}
}
