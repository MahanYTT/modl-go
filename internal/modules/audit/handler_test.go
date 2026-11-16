package audit

import (
	"bytes"
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
	perms      map[string]bool
}

func (p stubPerms) IsSuperAdmin(_ *models.Server, _ string) bool { return p.superAdmin }
func (p stubPerms) IsAuthorizedEmail(_ context.Context, _ *models.Server, _ string) bool {
	return true
}
func (p stubPerms) PanelRoleID(_ context.Context, _ *models.Server, _ string) string { return "mod" }
func (p stubPerms) HasPermission(_ context.Context, _ *models.Server, _, perm string) bool {
	return p.perms[perm]
}
func (p stubPerms) HasAnyWithPrefix(_ context.Context, _ *models.Server, _, _ string) bool {
	return false
}

func TestAuditEndpointsAndRBAC(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer database.Close()

	now := time.Now().UnixMilli()
	srv := &models.Server{ID: "srv-1", AdminEmail: "owner@example.com"}

	// Seed staff & role
	_, err = database.Exec(`INSERT INTO staff_roles (id, server_id, name, permissions_json, created_at, updated_at)
		VALUES ('mod', 'srv-1', 'Moderator', '["admin.audit.view.logs"]', ?, ?)`, now, now)
	if err != nil {
		t.Fatalf("seed role: %v", err)
	}
	_, err = database.Exec(`INSERT INTO staff (id, server_id, email, username, role_id, created_at, updated_at)
		VALUES ('st-1', 'srv-1', 'mod@example.com', 'ModUser', 'mod', ?, ?)`, now, now)
	if err != nil {
		t.Fatalf("seed staff: %v", err)
	}

	// Seed player + punishments
	_, err = database.Exec(`INSERT INTO players (id, server_id, minecraft_uuid, current_username, first_seen, last_seen)
		VALUES ('pl-1', 'srv-1', 'u1', 'BadUser', ?, ?)`, now-dayMs, now)
	if err != nil {
		t.Fatalf("seed player: %v", err)
	}
	_, err = database.Exec(`INSERT INTO punishments (id, server_id, player_id, player_uuid, player_username, type, punishment_type_id, reason, issuer_uuid, issuer_name, issued_at, expires_at, duration_ms, pardoned, data_json, updated_at)
		VALUES ('pn-1', 'srv-1', 'pl-1', 'u1', 'BadUser', '14', '14', 'Fly hacks', 'st-1', 'ModUser', ?, ?, 86400000, 0, '{"typeOrdinal":14}', ?),
		       ('pn-2', 'srv-1', 'pl-1', 'u1', 'BadUser', '14', '14', 'Speed hacks', 'st-1', 'ModUser', ?, ?, 86400000, 0, '{"typeOrdinal":14}', ?),
		       ('pn-3', 'srv-1', 'pl-1', 'u1', 'BadUser', '1', '1', 'Spamming', 'st-1', 'ModUser', ?, ?, 3600000, 0, '{"typeOrdinal":1}', ?)`,
		now-2*dayMs, now+dayMs, now,
		now-1*dayMs, now+dayMs, now,
		now-1*dayMs, now+dayMs, now,
	)
	if err != nil {
		t.Fatalf("seed punishments: %v", err)
	}

	// Seed tickets, settings (with secret & unsafe type), and audit log
	_, err = database.Exec(`INSERT INTO tickets (id, server_id, type, subject, status, replies_json, created_at, updated_at)
		VALUES ('tk-1', 'srv-1', 'support', 'Help', 'open', ?, ?, ?)`,
		`[{"name":"ModUser","content":"On it","created":`+jsonInt(now-dayMs+60000)+`}]`, now-dayMs, now)
	if err != nil {
		t.Fatalf("seed tickets: %v", err)
	}
	_, err = database.Exec(`INSERT INTO settings (id, server_id, type, data_json, updated_at)
		VALUES ('set-1', 'srv-1', 'general', '{"siteName":"Modl","webhookUrl":"https://secret"}', ?),
		       ('set-2', 'srv-1', 'discord', '{"token":"secret-token"}', ?)`, now, now)
	if err != nil {
		t.Fatalf("seed settings: %v", err)
	}
	_, err = database.Exec(`INSERT INTO audit_logs (id, server_id, created_at, level, source, description, metadata_json)
		VALUES ('al-1', 'srv-1', ?, 'moderation', 'ModUser', 'Banned BadUser', '{"playerId":"pl-1","playerName":"BadUser"}'),
		       ('al-2', 'srv-1', ?, 'info', 'ModUser', 'Uploaded evidence for pn-1', '{}')`,
		now-1000, now-500)
	if err != nil {
		t.Fatalf("seed audit logs: %v", err)
	}

	newRouter := func(p models.PermissionChecker) http.Handler {
		h := NewHandler(database, nil, p)
		r := chi.NewRouter()
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c := httpx.WithServer(r.Context(), srv)
				c = httpx.WithSession(c, &models.AuthSession{ID: "s1", ServerID: srv.ID, Email: "owner@example.com"})
				next.ServeHTTP(w, r.WithContext(c))
			})
		})
		h.RegisterPanelRoutes(r)
		return r
	}

	// 1. Non-superadmin with view.logs can read performance/details/punishments, cannot browse database
	modRouter := newRouter(stubPerms{superAdmin: false, perms: map[string]bool{"admin.audit.view.logs": true}})
	rec := httptest.NewRecorder()
	modRouter.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/audit/staff-performance?period=30d", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("staff-performance status=%d body=%s", rec.Code, rec.Body.String())
	}
	var perf struct {
		Staff []struct {
			Username          string `json:"username"`
			PunishmentsIssued int32  `json:"punishmentsIssued"`
			TicketResponses   int32  `json:"ticketResponses"`
		} `json:"staff"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &perf)
	if len(perf.Staff) != 1 || perf.Staff[0].PunishmentsIssued != 3 || perf.Staff[0].TicketResponses != 1 {
		t.Fatalf("unexpected staff perf: %+v", perf.Staff)
	}

	rec = httptest.NewRecorder()
	modRouter.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/audit/staff/ModUser/details?period=30d", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("staff details status=%d body=%s", rec.Code, rec.Body.String())
	}
	var det struct {
		EvidenceUploads int32 `json:"evidenceUploads"`
		Summary         struct {
			TotalPunishments int32 `json:"totalPunishments"`
			TotalTickets     int32 `json:"totalTickets"`
		} `json:"summary"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &det)
	if det.EvidenceUploads != 1 || det.Summary.TotalPunishments != 3 || det.Summary.TotalTickets != 1 {
		t.Fatalf("unexpected staff details: %+v", det)
	}

	rec = httptest.NewRecorder()
	modRouter.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/audit/punishments/active?status=active", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("active punishments status=%d", rec.Code)
	}

	rec = httptest.NewRecorder()
	modRouter.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/audit/punishments?limit=10", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("punishment audit status=%d", rec.Code)
	}

	// Database browser forbidden for non-superadmin
	rec = httptest.NewRecorder()
	modRouter.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/audit/database/settings", nil))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for database browser, got %d", rec.Code)
	}

	// 2. Superadmin operations
	adminRouter := newRouter(stubPerms{superAdmin: true})

	// Database browser redacts secrets & unsafe settings types
	rec = httptest.NewRecorder()
	adminRouter.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/audit/database/settings?limit=10&skip=0", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("db settings status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("[REDACTED]")) {
		t.Fatalf("expected redacted fields in settings browser: %s", rec.Body.String())
	}

	// Invalid table returns 400
	rec = httptest.NewRecorder()
	adminRouter.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/audit/database/users_secret", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid table, got %d", rec.Code)
	}

	// Single rollback pn-1
	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/audit/punishments/pn-1/rollback", bytes.NewBufferString(`{"reason":"Mistake"}`))
	req.Header.Set("Content-Type", "application/json")
	adminRouter.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("rollback status=%d body=%s", rec.Code, rec.Body.String())
	}

	// Double rollback fails with 400
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/audit/punishments/pn-1/rollback", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	adminRouter.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 on second rollback, got %d", rec.Code)
	}

	// Bulk set expiration on type 14 (pn-2 is still active; pn-1 is pardoned)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/audit/punishments/bulk-set-expiration", bytes.NewBufferString(`{"typeOrdinals":[14],"reason":"Reduce","newDurationMs":3600000}`))
	req.Header.Set("Content-Type", "application/json")
	adminRouter.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("bulk set expiration status=%d body=%s", rec.Code, rec.Body.String())
	}
	var bulkResp struct {
		Count int32 `json:"count"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &bulkResp)
	if bulkResp.Count != 1 {
		t.Fatalf("expected bulk set expiration count=1, got %d", bulkResp.Count)
	}

	// Bulk pardon on type 1 (pn-3)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/audit/punishments/bulk-pardon", bytes.NewBufferString(`{"typeOrdinals":[1],"reason":"Amnesty"}`))
	req.Header.Set("Content-Type", "application/json")
	adminRouter.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("bulk pardon status=%d body=%s", rec.Code, rec.Body.String())
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &bulkResp)
	if bulkResp.Count != 1 {
		t.Fatalf("expected bulk pardon count=1, got %d", bulkResp.Count)
	}

	// Rollback by date range for ModUser (pn-2 and pn-3 not yet rolled back)
	startISO := time.UnixMilli(now - 3*dayMs).UTC().Format(time.RFC3339)
	endISO := time.UnixMilli(now).UTC().Format(time.RFC3339)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/audit/staff/ModUser/rollback-date-range",
		bytes.NewBufferString(`{"startDate":"`+startISO+`","endDate":"`+endISO+`","reason":"Cleanup"}`))
	req.Header.Set("Content-Type", "application/json")
	adminRouter.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("rollback date range status=%d body=%s", rec.Code, rec.Body.String())
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &bulkResp)
	if bulkResp.Count != 2 {
		t.Fatalf("expected rollback-date-range count=2, got %d", bulkResp.Count)
	}
}

func jsonInt(v int64) string {
	b, _ := json.Marshal(v)
	return string(b)
}
