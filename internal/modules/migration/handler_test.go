package migration

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"modl-go/internal/config"
	"modl-go/internal/db"
	"modl-go/internal/httpx"
	"modl-go/internal/models"

	"github.com/go-chi/chi/v5"
)

type stubPerms struct {
	superAdmin bool
	view       bool
	modify     bool
}

func (p stubPerms) IsSuperAdmin(_ *models.Server, _ string) bool { return p.superAdmin }
func (p stubPerms) IsAuthorizedEmail(_ context.Context, _ *models.Server, _ string) bool {
	return true
}
func (p stubPerms) PanelRoleID(_ context.Context, _ *models.Server, _ string) string { return "mod" }
func (p stubPerms) HasPermission(_ context.Context, _ *models.Server, _, perm string) bool {
	if perm == "admin.settings.view.migration" {
		return p.view
	}
	if perm == "admin.settings.modify.migration" {
		return p.modify
	}
	return false
}
func (p stubPerms) HasAnyWithPrefix(_ context.Context, _ *models.Server, _, _ string) bool {
	return false
}

func TestMigrationLifecycleUploadAndPermissions(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer database.Close()

	tmpDir := t.TempDir()
	cfg := &config.Config{StorageDir: tmpDir}
	srv := &models.Server{ID: "srv-1", AdminEmail: "owner@example.com"}

	// Viewer cannot start migration (modify permission missing)
	hView := NewHandler(database, cfg, nil, stubPerms{view: true, modify: false})
	rView := chi.NewRouter()
	rView.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := httpx.WithServer(r.Context(), srv)
			c = httpx.WithSession(c, &models.AuthSession{ID: "s1", ServerID: srv.ID, Email: "viewer@example.com"})
			next.ServeHTTP(w, r.WithContext(c))
		})
	})
	hView.RegisterPanelRoutes(rView)

	rec := httptest.NewRecorder()
	rView.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/migration/status", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("viewer status=%d", rec.Code)
	}
	rec = httptest.NewRecorder()
	rView.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/migration/start", bytes.NewBufferString(`{"migrationType":"litebans"}`)))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for viewer start, got %d", rec.Code)
	}

	// Admin router + minecraft router
	h := NewHandler(database, cfg, nil, stubPerms{superAdmin: true})
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := httpx.WithServer(r.Context(), srv)
			c = httpx.WithSession(c, &models.AuthSession{ID: "s1", ServerID: srv.ID, Email: "owner@example.com"})
			next.ServeHTTP(w, r.WithContext(c))
		})
	})
	h.RegisterPanelRoutes(r)
	r.Route("/v1/minecraft", func(r chi.Router) { h.RegisterMinecraftV1Routes(r) })

	// 1. Start invalid type -> 400
	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/migration/start", bytes.NewBufferString(`{"migrationType":"unknown"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid type, got %d", rec.Code)
	}

	// 2. Start litebans -> 200, then cancel from building_json (cooldown-exempt) -> 200
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/migration/start", bytes.NewBufferString(`{"migrationType":"litebans"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("start status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/migration/cancel", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("cancel status=%d body=%s", rec.Code, rec.Body.String())
	}

	// 3. Can start again immediately because cancel from building_json is cooldown-exempt
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/migration/start", bytes.NewBufferString(`{"migrationType":"litebans"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("restart after exempt cancel status=%d body=%s", rec.Code, rec.Body.String())
	}

	// 4. Minecraft progress update
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/v1/minecraft/migration/progress",
		bytes.NewBufferString(`{"status":"uploading_json","message":"Uploading...","recordsProcessed":0,"totalRecords":1}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("progress status=%d body=%s", rec.Code, rec.Body.String())
	}

	// 5. Minecraft multipart upload
	migrationJSON := `{
	  "metadata": {"playerCount": 2},
	  "players": [
	    {
	      "minecraftUuid": "123e4567e89b12d3a456426614174000",
	      "usernames": [{"username": "MigratedPlayer", "date": "2026-01-01T00:00:00Z"}],
	      "ipAddresses": [{"ipAddress": "1.2.3.4", "country": "US"}],
	      "punishments": [{"id": "mig-p1", "typeOrdinal": 14, "reason": "Cheating", "issuerName": "Console", "issued": "2026-01-01T00:00:00Z", "duration": 3600000}]
	    },
	    {
	      "minecraftUuid": "not-a-uuid"
	    }
	  ]
	}`

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("migrationFile", "migration.json")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	_, _ = fw.Write([]byte(migrationJSON))
	mw.Close()

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/v1/minecraft/migration/upload", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("upload status=%d body=%s", rec.Code, rec.Body.String())
	}

	// Poll status until completed (fast in-memory sqlite)
	deadline := time.Now().Add(3 * time.Second)
	var finalStatus string
	for time.Now().Before(deadline) {
		rec = httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/migration/status", nil))
		var st struct {
			CurrentMigration *struct {
				Status   string `json:"status"`
				Progress struct {
					RecordsProcessed int32 `json:"recordsProcessed"`
					RecordsSkipped   int32 `json:"recordsSkipped"`
				} `json:"progress"`
			} `json:"currentMigration"`
			Cooldown struct {
				OnCooldown bool `json:"onCooldown"`
			} `json:"cooldown"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &st)
		if st.CurrentMigration != nil && (st.CurrentMigration.Status == "completed" || st.CurrentMigration.Status == "failed") {
			finalStatus = st.CurrentMigration.Status
			if finalStatus != "completed" || st.CurrentMigration.Progress.RecordsProcessed != 1 || st.CurrentMigration.Progress.RecordsSkipped != 1 || !st.Cooldown.OnCooldown {
				t.Fatalf("unexpected final status: %+v cooldown=%+v", st.CurrentMigration, st.Cooldown)
			}
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if finalStatus != "completed" {
		t.Fatalf("expected migration to complete, got %q", finalStatus)
	}

	// Verify player & punishment persisted in SQLite
	var pCount, pnCount int
	_ = database.QueryRow(`SELECT COUNT(*) FROM players WHERE server_id = 'srv-1' AND minecraft_uuid = '123e4567-e89b-12d3-a456-426614174000'`).Scan(&pCount)
	_ = database.QueryRow(`SELECT COUNT(*) FROM punishments WHERE server_id = 'srv-1' AND id = 'mig-p1'`).Scan(&pnCount)
	if pCount != 1 || pnCount != 1 {
		t.Fatalf("expected player and punishment persisted, got players=%d punishments=%d", pCount, pnCount)
	}
}
