package role_test

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

	"github.com/go-chi/chi/v5"
)

type stubServerResolver struct{ srv *models.Server }

func (s stubServerResolver) GetByDomain(context.Context, string) (*models.Server, error) {
	return s.srv, nil
}
func (s stubServerResolver) GetByAPIKey(context.Context, string) (*models.Server, error) {
	return s.srv, nil
}
func (s stubServerResolver) GetDefault(context.Context) (*models.Server, error) {
	return s.srv, nil
}

func TestRolesAndPermissions(t *testing.T) {
	sqlDB, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer sqlDB.Close()

	srv := &models.Server{ID: "srv1", ServerName: "Test", CustomDomain: "test.local", AdminEmail: "admin@test.local"}
	cfg := &config.Config{SingleTenant: true}
	rh := role.NewHandler(sqlDB, nil)

	r := chi.NewRouter()
	r.Route("/v1/panel", func(r chi.Router) {
		r.Use(httpx.ServerHeaderMiddleware(cfg, stubServerResolver{srv}))
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				req = req.WithContext(httpx.WithSession(req.Context(), &models.AuthSession{ID: "s1", ServerID: srv.ID, Email: srv.AdminEmail}))
				next.ServeHTTP(w, req)
			})
		})
		rh.RegisterPanelRoutes(r)
	})

	// List seeded default roles
	req := httptest.NewRequest(http.MethodGet, "/v1/panel/roles", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET roles failed %d: %s", rec.Code, rec.Body.String())
	}
	var listResp struct {
		Roles []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"roles"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &listResp)
	if len(listResp.Roles) < 3 {
		t.Fatalf("expected 3 default roles, got %d", len(listResp.Roles))
	}

	// Create custom role (superAdminOnly perms filtered out)
	createBody := `{"name":"TrialMod","description":"Trial","permissions":["punishment.view","admin.audit.rollback"]}`
	req2 := httptest.NewRequest(http.MethodPost, "/v1/panel/roles", bytes.NewBufferString(createBody))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusCreated {
		t.Fatalf("POST role failed %d: %s", rec2.Code, rec2.Body.String())
	}
	var mutResp struct {
		Role struct {
			ID          string   `json:"id"`
			Permissions []string `json:"permissions"`
		} `json:"role"`
	}
	_ = json.Unmarshal(rec2.Body.Bytes(), &mutResp)
	if len(mutResp.Role.Permissions) != 1 || mutResp.Role.Permissions[0] != "punishment.view" {
		t.Fatalf("expected super-admin-only perm filtered out, got %+v", mutResp.Role.Permissions)
	}
}
