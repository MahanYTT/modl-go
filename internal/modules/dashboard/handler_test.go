package dashboard

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/db"
	"modl-go/internal/httpx"
	"modl-go/internal/models"

	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/proto"
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

func TestDashboardEndpointsAndAlerts(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer database.Close()

	now := time.Now().UnixMilli()
	srv := &models.Server{ID: "srv-1", AdminEmail: "owner@example.com"}

	// Seed staff
	_, err = database.Exec(`INSERT INTO staff (id, server_id, email, username, role_id, assigned_minecraft_uuid, created_at, updated_at)
		VALUES ('st-1', 'srv-1', 'mod@example.com', 'ModUser', 'mod', 'uuid-1', ?, ?)`, now, now)
	if err != nil {
		t.Fatalf("seed staff: %v", err)
	}

	// Seed players (1 online staff, 1 offline player)
	_, err = database.Exec(`INSERT INTO players (id, server_id, minecraft_uuid, current_username, data_json, first_seen, last_seen)
		VALUES ('pl-1', 'srv-1', 'uuid-1', 'ModUser', '{"isOnline":true}', ?, ?),
		       ('pl-2', 'srv-1', 'uuid-2', 'BadPlayer', '{"isOnline":false}', ?, ?)`,
		now-2*dayMs, now, now-10*dayMs, now)
	if err != nil {
		t.Fatalf("seed players: %v", err)
	}

	// Seed tickets (1 open report, 1 open support, 1 unfinished)
	_, err = database.Exec(`INSERT INTO tickets (id, server_id, type, subject, status, creator_name, replies_json, created_at, updated_at)
		VALUES ('tk-1', 'srv-1', 'player', 'Hacking report', 'open', 'PlayerA', '[{"name":"PlayerA","content":"Help!"}]', ?, ?),
		       ('tk-2', 'srv-1', 'support', 'Need help', 'open', 'ModUser', '[{"name":"ModUser","content":"Initial","created":?}]', ?, ?),
		       ('tk-3', 'srv-1', 'bug', 'Draft', 'unfinished', 'PlayerB', '[]', ?, ?)`,
		now-1*dayMs, now, now-1*dayMs, now-1*dayMs, now, now-1*dayMs, now)
	if err != nil {
		t.Fatalf("seed tickets: %v", err)
	}

	// Seed punishments (1 active ban, 1 active mute, 1 pardoned ban, 1 kick)
	_, err = database.Exec(`INSERT INTO punishments (id, server_id, player_id, player_uuid, player_username, type, punishment_type_id, reason, issuer_name, issued_at, expires_at, pardoned, data_json, updated_at)
		VALUES ('pn-1', 'srv-1', 'pl-2', 'uuid-2', 'BadPlayer', '14', '14', 'Cheating', 'ModUser', ?, ?, 0, '{"typeOrdinal":14}', ?),
		       ('pn-2', 'srv-1', 'pl-2', 'uuid-2', 'BadPlayer', '1', '1', 'Spam', 'ModUser', ?, ?, 0, '{"typeOrdinal":1}', ?),
		       ('pn-3', 'srv-1', 'pl-2', 'uuid-2', 'BadPlayer', '2', '2', 'Old ban', 'ModUser', ?, NULL, 1, '{"typeOrdinal":2}', ?),
		       ('pn-4', 'srv-1', 'pl-2', 'uuid-2', 'BadPlayer', '0', '0', 'Kicked', 'ModUser', ?, NULL, 0, '{"typeOrdinal":0}', ?)`,
		now-1*dayMs, now+dayMs, now,
		now-2*dayMs, now+dayMs, now,
		now-3*dayMs, now,
		now-1*dayMs, now,
	)
	if err != nil {
		t.Fatalf("seed punishments: %v", err)
	}

	// Seed system alerts (1 general, 1 super-admin only, 1 expired)
	_, err = database.Exec(`INSERT INTO system_alerts (id, message, severity, audience, expires_at, created_at, updated_at, created_by, updated_by)
		VALUES ('al-1', 'General alert', 'INFO', 'ALL_PANEL_USERS', ?, ?, ?, 'admin', 'admin'),
		       ('al-2', 'SuperAdmin alert', 'WARNING', 'SUPER_ADMINS_ONLY', ?, ?, ?, 'admin', 'admin'),
		       ('al-3', 'Expired alert', 'CRITICAL', 'ALL_PANEL_USERS', ?, ?, ?, 'admin', 'admin')`,
		now+dayMs, now, now,
		now+dayMs, now, now,
		now-dayMs, now-2*dayMs, now-2*dayMs,
	)
	if err != nil {
		t.Fatalf("seed alerts: %v", err)
	}

	// Regular staff with dashboard view permission
	h := NewHandler(database, stubPerms{superAdmin: false, allow: true})
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := httpx.WithServer(r.Context(), srv)
			c = httpx.WithSession(c, &models.AuthSession{ID: "s1", ServerID: srv.ID, Email: "mod@example.com"})
			next.ServeHTTP(w, r.WithContext(c))
		})
	})
	h.RegisterPanelRoutes(r)
	r.Route("/v1/minecraft", func(r chi.Router) { h.RegisterMinecraftV1Routes(r) })
	r.Route("/v3/minecraft", func(r chi.Router) { h.RegisterMinecraftV3Routes(r) })

	// 1. Metrics
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/dashboard/metrics?period=7d", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("metrics status=%d body=%s", rec.Code, rec.Body.String())
	}
	var metrics map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &metrics)
	if metrics["totalTickets"] != "3" || metrics["openTickets"] != "2" || metrics["totalPlayers"] != "2" || metrics["activePunishments"] != "2" || metrics["totalStaff"] != "2" {
		t.Fatalf("unexpected metrics: %+v", metrics)
	}

	// 2. Recent tickets (excludes unfinished)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/dashboard/recent-tickets?limit=10", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("recent-tickets status=%d body=%s", rec.Code, rec.Body.String())
	}
	var rtResp struct {
		Tickets []struct {
			ID string `json:"id"`
		} `json:"tickets"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &rtResp)
	if len(rtResp.Tickets) != 2 {
		t.Fatalf("expected 2 non-unfinished tickets, got %d", len(rtResp.Tickets))
	}

	// 3. Recent punishments
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/dashboard/recent-punishments?limit=10", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("recent-punishments status=%d body=%s", rec.Code, rec.Body.String())
	}
	var rpResp struct {
		Punishments []struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		} `json:"punishments"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &rpResp)
	if len(rpResp.Punishments) != 4 || rpResp.Punishments[0].Type != "Cheating" {
		t.Fatalf("unexpected recent punishments: %+v", rpResp.Punishments)
	}

	// 4. Recent activity
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/dashboard/activity/recent?limit=10&days=7", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("activity status=%d body=%s", rec.Code, rec.Body.String())
	}
	var actResp struct {
		Activity []struct {
			ID string `json:"id"`
		} `json:"activity"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &actResp)
	if len(actResp.Activity) == 0 {
		t.Fatalf("expected activities, got 0")
	}

	// 5. Alerts (non-superadmin sees only al-1)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/dashboard/alerts", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("alerts status=%d body=%s", rec.Code, rec.Body.String())
	}
	var alResp struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &alResp)
	if len(alResp.Items) != 1 || alResp.Items[0].ID != "al-1" {
		t.Fatalf("expected only al-1 for non-superadmin, got %+v", alResp.Items)
	}

	// 6. Minecraft V1 stats (JSON)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/minecraft/dashboard/stats", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("mc v1 status=%d body=%s", rec.Code, rec.Body.String())
	}
	var mcV1 struct {
		Status int            `json:"status"`
		Stats  MinecraftStats `json:"stats"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &mcV1)
	if mcV1.Stats.UnresolvedReports != 1 || mcV1.Stats.UnresolvedTickets != 2 || mcV1.Stats.OnlineStaff != 1 ||
		mcV1.Stats.OnlinePlayers != 1 || mcV1.Stats.ActiveBans != 1 || mcV1.Stats.ActiveMutes != 1 ||
		mcV1.Stats.TotalActivePunishments != 2 || mcV1.Stats.TotalPlayers != 2 {
		t.Fatalf("unexpected mc v1 stats: %+v", mcV1.Stats)
	}

	// 7. Minecraft V3 stats (Protobuf)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v3/minecraft/dashboard/stats", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("mc v3 status=%d", rec.Code)
	}
	var mcV3 modlv1.MinecraftDashboardResponse
	if err := proto.Unmarshal(rec.Body.Bytes(), &mcV3); err != nil {
		t.Fatalf("unmarshal v3 proto: %v", err)
	}
	if mcV3.GetStatus() != 200 || mcV3.GetStats().GetActiveBans() != 1 || mcV3.GetStats().GetActiveMutes() != 1 {
		t.Fatalf("unexpected mc v3 proto: %+v", &mcV3)
	}
}
