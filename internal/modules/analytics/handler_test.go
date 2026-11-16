package analytics

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
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

func TestAnalyticsEndpoints(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer database.Close()

	now := time.Now().UnixMilli()
	srv := &models.Server{ID: "srv-1", AdminEmail: "owner@example.com"}

	// Seed staff
	_, err = database.Exec(`INSERT INTO staff (id, server_id, email, username, role_id, created_at, updated_at)
		VALUES ('st-1', 'srv-1', 'mod@example.com', 'ModUser', 'mod', ?, ?)`, now, now)
	if err != nil {
		t.Fatalf("seed staff: %v", err)
	}

	// Seed players with IPs
	ip1 := `[{"country":"US","proxy":true,"hosting":false,"firstLogin":` + strconvFormat(now-2*dayMs) + `,"logins":[` + strconvFormat(now-dayMs) + `,` + strconvFormat(now-1000) + `]}]`
	ip2 := `[{"country":"DE","proxy":false,"hosting":true,"firstLogin":` + strconvFormat(now-40*dayMs) + `,"logins":[` + strconvFormat(now-40*dayMs) + `]}]`
	_, err = database.Exec(`INSERT INTO players (id, server_id, minecraft_uuid, current_username, ip_addresses_json, first_seen, last_seen)
		VALUES ('p1', 'srv-1', 'u1', 'P1', ?, ?, ?),
		       ('p2', 'srv-1', 'u2', 'P2', ?, ?, ?)`,
		ip1, now-2*dayMs, now, ip2, now-40*dayMs, now)
	if err != nil {
		t.Fatalf("seed players: %v", err)
	}

	// Seed tickets (1 open, 1 closed after 2h, 1 unfinished)
	_, err = database.Exec(`INSERT INTO tickets (id, server_id, type, subject, status, created_at, updated_at)
		VALUES ('t1', 'srv-1', 'support', 'Help', 'open', ?, ?),
		       ('t2', 'srv-1', 'bug', 'Bug', 'closed', ?, ?),
		       ('t3', 'srv-1', 'appeal', 'Draft', 'unfinished', ?, ?)`,
		now-2*dayMs, now-2*dayMs,
		now-3*dayMs, now-3*dayMs+2*60*60*1000,
		now-1*dayMs, now-1*dayMs)
	if err != nil {
		t.Fatalf("seed tickets: %v", err)
	}

	// Seed punishments
	_, err = database.Exec(`INSERT INTO punishments (id, server_id, player_id, player_uuid, type, punishment_type_id, issuer_uuid, issuer_name, issued_at, data_json, updated_at)
		VALUES ('pn1', 'srv-1', 'p1', 'u1', '14', '14', 'st-1', 'ModUser', ?, '{"typeOrdinal":14}', ?),
		       ('pn2', 'srv-1', 'p1', 'u1', '1', '1', '', 'Console', ?, '{"typeOrdinal":1}', ?)`,
		now-dayMs, now, now-2*dayMs, now)
	if err != nil {
		t.Fatalf("seed punishments: %v", err)
	}

	// Seed audit logs
	_, err = database.Exec(`INSERT INTO audit_logs (id, server_id, created_at, level, source, description)
		VALUES ('al1', 'srv-1', ?, 'moderation', 'ModUser', 'Banned P1'),
		       ('al2', 'srv-1', ?, 'info', 'system', 'Sync complete')`,
		now-1000, now-500)
	if err != nil {
		t.Fatalf("seed audit logs: %v", err)
	}

	h := NewHandler(database, stubPerms{allow: true})
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := httpx.WithServer(r.Context(), srv)
			c = httpx.WithSession(c, &models.AuthSession{ID: "s1", ServerID: srv.ID, Email: "mod@example.com"})
			next.ServeHTTP(w, r.WithContext(c))
		})
	})
	h.RegisterPanelRoutes(r)

	// Overview
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/analytics/overview", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("overview status=%d body=%s", rec.Code, rec.Body.String())
	}
	var ov struct {
		Overview map[string]struct {
			TotalTickets string `json:"totalTickets"`
			TotalPlayers string `json:"totalPlayers"`
			TotalStaff   string `json:"totalStaff"`
			TicketChange int32  `json:"ticketChange"`
		} `json:"overview"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &ov)
	if ov.Overview["30d"].TotalTickets != "3" || ov.Overview["30d"].TotalPlayers != "2" || ov.Overview["30d"].TotalStaff != "2" || ov.Overview["30d"].TicketChange != 100 {
		t.Fatalf("unexpected overview: %+v", ov.Overview)
	}

	// Tickets
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/analytics/tickets?period=30d", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("tickets status=%d body=%s", rec.Code, rec.Body.String())
	}
	var tk struct {
		ByStatus []struct {
			Status string `json:"status"`
			Count  int32  `json:"count"`
		} `json:"byStatus"`
		AvgResolutionByCategory []struct {
			Category string  `json:"category"`
			AvgHours float64 `json:"avgHours"`
		} `json:"avgResolutionByCategory"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &tk)
	if len(tk.ByStatus) != 2 || len(tk.AvgResolutionByCategory) != 1 || tk.AvgResolutionByCategory[0].AvgHours != 2.0 {
		t.Fatalf("unexpected ticket analytics: %+v", tk)
	}

	// Punishments
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/analytics/punishments?period=30d", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("punishments status=%d", rec.Code)
	}
	var pn struct {
		ByType []struct {
			Type  string `json:"type"`
			Count int32  `json:"count"`
		} `json:"byType"`
		ByStaff []struct {
			Username string `json:"username"`
			Count    int32  `json:"count"`
		} `json:"byStaff"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &pn)
	if len(pn.ByType) != 2 || len(pn.ByStaff) != 2 {
		t.Fatalf("unexpected punishment analytics: %+v", pn)
	}

	// Audit logs
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/analytics/audit-logs?period=30d", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("audit-logs status=%d", rec.Code)
	}
	var al struct {
		ByLevel []struct {
			Level string `json:"level"`
			Count int32  `json:"count"`
		} `json:"byLevel"`
		HourlyTrend []struct {
			Hour  string `json:"hour"`
			Count int32  `json:"count"`
		} `json:"hourlyTrend"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &al)
	if len(al.ByLevel) != 2 || len(al.HourlyTrend) != 1 || al.HourlyTrend[0].Count != 2 {
		t.Fatalf("unexpected audit log analytics: %+v", al)
	}

	// Player activity
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/analytics/player-activity?period=30d", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("player-activity status=%d", rec.Code)
	}
	var pa struct {
		LoginsByCountry []struct {
			Country string `json:"country"`
			Count   int32  `json:"count"`
		} `json:"loginsByCountry"`
		SuspiciousActivity struct {
			ProxyCount   int32 `json:"proxyCount"`
			HostingCount int32 `json:"hostingCount"`
		} `json:"suspiciousActivity"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &pa)
	if len(pa.LoginsByCountry) != 1 || pa.LoginsByCountry[0].Country != "US" || pa.LoginsByCountry[0].Count != 2 ||
		pa.SuspiciousActivity.ProxyCount != 1 || pa.SuspiciousActivity.HostingCount != 0 {
		t.Fatalf("unexpected player activity: %+v", pa)
	}
}

func strconvFormat(v int64) string {
	return strconv.FormatInt(v, 10)
}
