package appeal_test

import (
	"bytes"
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/config"
	"modl-go/internal/db"
	"modl-go/internal/email"
	"modl-go/internal/httpx"
	"modl-go/internal/models"
	"modl-go/internal/modules/appeal"
	"modl-go/internal/modules/ticket"

	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type stubPerms struct {
	superAdmin bool
	perms      map[string]bool
}

func (s *stubPerms) IsSuperAdmin(_ *models.Server, _ string) bool { return s.superAdmin }
func (s *stubPerms) IsAuthorizedEmail(_ context.Context, _ *models.Server, _ string) bool {
	return true
}
func (s *stubPerms) PanelRoleID(_ context.Context, _ *models.Server, _ string) string {
	return "mod"
}
func (s *stubPerms) HasPermission(_ context.Context, _ *models.Server, _ string, perm string) bool {
	return s.superAdmin || s.perms[perm]
}
func (s *stubPerms) HasAnyWithPrefix(_ context.Context, _ *models.Server, _ string, _ string) bool {
	return s.superAdmin
}

func setupAppealEnv(t *testing.T, perms models.PermissionChecker) (*sql.DB, *appeal.Handler, *models.Server, *config.Config) {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open memory db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	srv := &models.Server{
		ID:           "srv-1",
		ServerName:   "AppealNet",
		CustomDomain: "appeal.modl.gg",
		AdminEmail:   "admin@appeal.gg",
		APIKey:       "api-key",
		CreatedAt:    1000,
		UpdatedAt:    1000,
	}
	_, err = database.Exec(`
		INSERT INTO servers (id, server_name, custom_domain, admin_email, api_key, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 1000, 1000)`,
		srv.ID, srv.ServerName, srv.CustomDomain, srv.AdminEmail, srv.APIKey)
	if err != nil {
		t.Fatalf("insert server: %v", err)
	}
	_, err = database.Exec(`
		INSERT INTO staff (id, server_id, email, username, role_id, created_at, updated_at)
		VALUES ('st-1', ?, 'mod@appeal.gg', 'ModBob', 'mod', 1000, 1000)`, srv.ID)
	if err != nil {
		t.Fatalf("insert staff: %v", err)
	}

	playerUUID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	_, err = database.Exec(`
		INSERT INTO players (id, server_id, minecraft_uuid, current_username, usernames_json, first_seen, last_seen)
		VALUES ('pl-1', ?, ?, 'BannedPlayer', '[{"username":"BannedPlayer","timestamp":1000}]', 1000, 1000)`,
		srv.ID, playerUUID)
	if err != nil {
		t.Fatalf("insert player: %v", err)
	}

	// Main ban with block_alts = 1 and a linked alt ban
	_, err = database.Exec(`
		INSERT INTO punishments (id, server_id, player_id, player_uuid, player_username, type, reason, issuer_name, issued_at, pardoned, block_alts, notes_json, data_json, updated_at)
		VALUES ('pun-1', ?, 'pl-1', ?, 'BannedPlayer', 'BAN', 'Unfair ban', 'Admin', 1000, 0, 1, '[]', '{}', 1000)`,
		srv.ID, playerUUID)
	if err != nil {
		t.Fatalf("insert punishment: %v", err)
	}
	_, err = database.Exec(`
		INSERT INTO punishments (id, server_id, player_id, player_uuid, player_username, type, reason, issuer_name, issued_at, pardoned, block_alts, notes_json, data_json, updated_at)
		VALUES ('pun-alt', ?, 'pl-2', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'AltAccount', 'BAN', 'Alt ban', 'System', 1000, 0, 0, '[]', '{"linkedBanId":"pun-1"}', 1000)`,
		srv.ID)
	if err != nil {
		t.Fatalf("insert linked alt punishment: %v", err)
	}

	// Second punishment already pardoned
	_, err = database.Exec(`
		INSERT INTO punishments (id, server_id, player_id, player_uuid, player_username, type, reason, issuer_name, issued_at, pardoned, block_alts, notes_json, data_json, updated_at)
		VALUES ('pun-pardoned', ?, 'pl-1', ?, 'BannedPlayer', 'MUTE', 'Old mute', 'Admin', 1000, 1, 0, '[]', '{}', 1000)`,
		srv.ID, playerUUID)
	if err != nil {
		t.Fatalf("insert pardoned punishment: %v", err)
	}

	cfg := &config.Config{AuthCodeHashSecret: "appeal-test-secret"}
	mailer := email.NewSender(cfg)
	tHandler := ticket.NewHandler(database, cfg, mailer, nil, perms, nil)
	aHandler := appeal.NewHandler(database, cfg, mailer, nil, perms, tHandler)
	return database, aHandler, srv, cfg
}

func doProtoReq(t *testing.T, router http.Handler, srv *models.Server, sessionEmail, method, path string, body proto.Message) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		b, err := protojson.Marshal(body)
		if err != nil {
			t.Fatalf("marshal proto: %v", err)
		}
		buf.Write(b)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	ctx := httpx.WithServer(req.Context(), srv)
	ctx = httpx.WithServerDomain(ctx, srv.CustomDomain)
	if sessionEmail != "" {
		ctx = httpx.WithSession(ctx, &models.AuthSession{ID: "sess-1", ServerID: srv.ID, Email: sessionEmail})
	}
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestAppealFullLifecycleAndRBAC(t *testing.T) {
	perms := &stubPerms{perms: map[string]bool{
		"ticket.view.all": true,
		"appeal.modify":   true,
	}}
	database, h, srv, cfg := setupAppealEnv(t, perms)

	rPub := chi.NewRouter()
	h.RegisterPublicRoutes(rPub)
	rPanel := chi.NewRouter()
	h.RegisterPanelRoutes(rPanel)

	playerUUID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	reason := "It was my brother on my account"

	// 1. Create appeal via public endpoint
	rec := doProtoReq(t, rPub, srv, "", http.MethodPost, "/appeals", &modlv1.CreateAppealRequest{
		PunishmentId: "pun-1",
		PlayerUuid:   playerUUID,
		Email:        "player@appeal.gg",
		Reason:       &reason,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create appeal failed: %d %s", rec.Code, rec.Body.String())
	}
	var created modlv1.CreatePublicAppealResponse
	_ = protojson.Unmarshal(rec.Body.Bytes(), &created)
	if !created.Success || created.AppealId == "" {
		t.Fatalf("unexpected create appeal response: %+v", &created)
	}

	// Duplicate active appeal -> 409 Conflict
	rec = doProtoReq(t, rPub, srv, "", http.MethodPost, "/appeals", &modlv1.CreateAppealRequest{
		PunishmentId: "pun-1",
		PlayerUuid:   playerUUID,
		Email:        "player@appeal.gg",
		Reason:       &reason,
	})
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 on duplicate appeal, got %d", rec.Code)
	}

	// Already pardoned punishment -> 409 Conflict
	rec = doProtoReq(t, rPub, srv, "", http.MethodPost, "/appeals", &modlv1.CreateAppealRequest{
		PunishmentId: "pun-pardoned",
		PlayerUuid:   playerUUID,
		Email:        "player@appeal.gg",
		Reason:       &reason,
	})
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 on pardoned punishment, got %d", rec.Code)
	}

	// Verify punishment notes & attachedTicketIds updated
	var notesJSON, dataJSON string
	_ = database.QueryRow(`SELECT notes_json, data_json FROM punishments WHERE id = 'pun-1'`).Scan(&notesJSON, &dataJSON)
	if !strings.Contains(notesJSON, created.AppealId) || !strings.Contains(dataJSON, created.AppealId) {
		t.Fatalf("punishment not linked to appeal %s: notes=%s data=%s", created.AppealId, notesJSON, dataJSON)
	}

	// 2. Public GET requires verification (403), then verify & reply
	rec = doProtoReq(t, rPub, srv, "", http.MethodGet, "/appeals/"+created.AppealId, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 before verification, got %d", rec.Code)
	}
	rec = doProtoReq(t, rPub, srv, "", http.MethodPost, "/appeals/"+created.AppealId+"/request-verification", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("request-verification failed: %d %s", rec.Code, rec.Body.String())
	}
	_, _ = database.Exec(`UPDATE ticket_verifications SET code_hash = ? WHERE ticket_id = ?`,
		httpx.HashOneTimeCode("123456", cfg.AuthCodeHashSecret), created.AppealId)
	rec = doProtoReq(t, rPub, srv, "", http.MethodPost, "/appeals/"+created.AppealId+"/verify", &modlv1.VerifyTicketCodeRequest{Code: "123456"})
	if rec.Code != http.StatusOK {
		t.Fatalf("verify failed: %d %s", rec.Code, rec.Body.String())
	}
	var ver modlv1.PublicTicketVerificationResponse
	_ = protojson.Unmarshal(rec.Body.Bytes(), &ver)

	rec = doProtoReq(t, rPub, srv, "", http.MethodGet, "/appeals/"+created.AppealId+"?token="+ver.Token, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get public appeal with token failed: %d %s", rec.Code, rec.Body.String())
	}
	var pubAppeal modlv1.PublicAppealResponse
	_ = protojson.Unmarshal(rec.Body.Bytes(), &pubAppeal)
	if _, hasEmail := pubAppeal.Data.AsMap()["contactEmail"]; hasEmail {
		t.Fatalf("contactEmail must be redacted from public appeal!")
	}

	rec = doProtoReq(t, rPub, srv, "", http.MethodPost, "/appeals/"+created.AppealId+"/replies?token="+ver.Token, &modlv1.AddAppealReplyRequest{
		Name:    "BannedPlayer",
		Content: "Added more proof",
		Type:    "user",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("public reply failed: %d %s", rec.Code, rec.Body.String())
	}

	// 3. Panel GET by punishment & panel reply
	rec = doProtoReq(t, rPanel, srv, "mod@appeal.gg", http.MethodGet, "/appeals/punishment/pun-1", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get appeals by punishment failed: %d %s", rec.Code, rec.Body.String())
	}
	var byPun modlv1.AppealTicketsResponse
	_ = protojson.Unmarshal(rec.Body.Bytes(), &byPun)
	if len(byPun.Tickets) != 1 || byPun.Tickets[0].Id != created.AppealId {
		t.Fatalf("unexpected appeals by punishment: %+v", byPun.Tickets)
	}

	rec = doProtoReq(t, rPanel, srv, "mod@appeal.gg", http.MethodPost, "/appeals/"+created.AppealId+"/replies", &modlv1.AddAppealReplyRequest{
		Name:    "ModBob",
		Content: "Reviewing your appeal now",
		Type:    "staff",
		Staff:   true,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("panel reply failed: %d %s", rec.Code, rec.Body.String())
	}

	// 4. Approve appeal -> pardons pun-1 and cascades to pun-alt
	rec = doProtoReq(t, rPanel, srv, "mod@appeal.gg", http.MethodPatch, "/appeals/"+created.AppealId+"/status", &modlv1.UpdateAppealStatusRequest{
		Status:     "approved",
		Resolution: proto.String("Accepted"),
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("approve appeal failed: %d %s", rec.Code, rec.Body.String())
	}
	var approved modlv1.TicketResponse
	_ = protojson.Unmarshal(rec.Body.Bytes(), &approved)
	if approved.Status != "closed" || !approved.Locked || approved.AppealWorkflowStatus != "approved" {
		t.Fatalf("unexpected approved state: %+v", &approved)
	}

	var pardoned1, pardonedAlt int
	_ = database.QueryRow(`SELECT pardoned FROM punishments WHERE id = 'pun-1'`).Scan(&pardoned1)
	_ = database.QueryRow(`SELECT pardoned FROM punishments WHERE id = 'pun-alt'`).Scan(&pardonedAlt)
	if pardoned1 != 1 || pardonedAlt != 1 {
		t.Fatalf("expected both pun-1 and pun-alt pardoned, got %d and %d", pardoned1, pardonedAlt)
	}

	// Direct terminal-to-terminal transition returns 409
	rec = doProtoReq(t, rPanel, srv, "mod@appeal.gg", http.MethodPatch, "/appeals/"+created.AppealId+"/status", &modlv1.UpdateAppealStatusRequest{
		Status: "rejected",
	})
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 on direct approved->rejected transition, got %d", rec.Code)
	}

	// Reopen then reject
	rec = doProtoReq(t, rPanel, srv, "mod@appeal.gg", http.MethodPatch, "/appeals/"+created.AppealId+"/status", &modlv1.UpdateAppealStatusRequest{
		Status: "open",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("reopen failed: %d %s", rec.Code, rec.Body.String())
	}
	rec = doProtoReq(t, rPanel, srv, "mod@appeal.gg", http.MethodPatch, "/appeals/"+created.AppealId+"/status", &modlv1.UpdateAppealStatusRequest{
		Status: "rejected",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("reject after reopen failed: %d %s", rec.Code, rec.Body.String())
	}

	// RBAC check: without appeal.modify or ticket.reply.all -> 403 on reply & status
	delete(perms.perms, "appeal.modify")
	rec = doProtoReq(t, rPanel, srv, "mod@appeal.gg", http.MethodPost, "/appeals/"+created.AppealId+"/replies", &modlv1.AddAppealReplyRequest{
		Name:    "ModBob",
		Content: "Denied",
		Type:    "staff",
		Staff:   true,
	})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without appeal.modify/ticket.reply.all, got %d", rec.Code)
	}
}
