package ticket_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/config"
	"modl-go/internal/db"
	"modl-go/internal/email"
	"modl-go/internal/httpx"
	"modl-go/internal/models"
	"modl-go/internal/modules/ticket"

	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
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

func setupTestEnv(t *testing.T, perms models.PermissionChecker) (*sql.DB, *ticket.Handler, *models.Server, *config.Config) {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open memory db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	srv := &models.Server{
		ID:           "srv-1",
		ServerName:   "TestNetwork",
		CustomDomain: "test.modl.gg",
		AdminEmail:   "admin@test.gg",
		APIKey:       "test-api-key",
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
		INSERT INTO staff (id, server_id, email, username, role_id, assigned_minecraft_uuid, assigned_minecraft_username, created_at, updated_at)
		VALUES ('st-1', ?, 'mod@test.gg', 'ModAlice', 'mod', '11111111-1111-1111-1111-111111111111', 'ModAlice', 1000, 1000)`, srv.ID)
	if err != nil {
		t.Fatalf("insert staff 1: %v", err)
	}
	_, err = database.Exec(`
		INSERT INTO staff (id, server_id, email, username, role_id, created_at, updated_at)
		VALUES ('st-2', ?, 'helper@test.gg', 'HelperBob', 'mod', 1000, 1000)`, srv.ID)
	if err != nil {
		t.Fatalf("insert staff 2: %v", err)
	}

	cfg := &config.Config{
		AuthCodeHashSecret: "unit-test-secret",
		ModlDomain:         "modl.gg",
	}
	mailer := email.NewSender(cfg)
	h := ticket.NewHandler(database, cfg, mailer, nil, perms, nil)
	return database, h, srv, cfg
}

func doReq(t *testing.T, router http.Handler, srv *models.Server, sessionEmail, method, path string, body proto.Message, rawJSON any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	contentType := "application/json"
	if body != nil {
		b, err := protojson.Marshal(body)
		if err != nil {
			t.Fatalf("marshal proto: %v", err)
		}
		buf.Write(b)
	} else if rawJSON != nil {
		if err := json.NewEncoder(&buf).Encode(rawJSON); err != nil {
			t.Fatalf("encode json: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", contentType)
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

func TestPanelTicketsCRUDAndRBAC(t *testing.T) {
	perms := &stubPerms{perms: map[string]bool{
		"ticket.view.all":  true,
		"ticket.reply.all": true,
		"ticket.close.all": true,
	}}
	database, h, srv, _ := setupTestEnv(t, perms)

	// Seed quick responses setting
	qr := &modlv1.QuickResponseSettings{
		Categories: []*modlv1.QuickResponseCategory{{
			Id:   "cat-1",
			Name: "General",
			Actions: []*modlv1.QuickResponseAction{{
				Id:          "act-1",
				Name:        "Resolved",
				Message:     "Your issue has been resolved!",
				CloseTicket: proto.Bool(true),
			}},
		}},
	}
	qrBytes, _ := protojson.Marshal(qr)
	_, err := database.Exec(`INSERT INTO settings (id, server_id, type, data_json, version, updated_at) VALUES ('set-1', ?, 'quickResponses', ?, 1, 1000)`, srv.ID, string(qrBytes))
	if err != nil {
		t.Fatalf("insert quickResponses: %v", err)
	}

	r := chi.NewRouter()
	h.RegisterPanelRoutes(r)

	// 1. Create panel ticket
	subj := "Cannot join lobby"
	desc := "Getting kicked on join"
	cname := "PlayerOne"
	cuuid := "22222222-2222-2222-2222-222222222222"
	rec := doReq(t, r, srv, "mod@test.gg", http.MethodPost, "/tickets", &modlv1.CreateTicketRequest{
		Type:        "bug",
		Subject:     &subj,
		Description: &desc,
		CreatorName: &cname,
		CreatorUuid: &cuuid,
		Tags:        []string{"lobby"},
	}, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var created modlv1.TicketResponse
	if err := protojson.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal created: %v", err)
	}
	if created.Id == "" || created.Type != "bug" || len(created.Messages) != 1 {
		t.Fatalf("unexpected created ticket: %+v", &created)
	}

	// 2. Search & Counts
	rec = doReq(t, r, srv, "mod@test.gg", http.MethodGet, "/tickets?search=lobby&type=bug&status=open", nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("search failed: %d %s", rec.Code, rec.Body.String())
	}
	var paged modlv1.PaginatedTicketsResponse
	_ = protojson.Unmarshal(rec.Body.Bytes(), &paged)
	if len(paged.Tickets) != 1 || paged.Pagination.TotalTickets != 1 {
		t.Fatalf("expected 1 ticket in search, got %d", len(paged.Tickets))
	}

	rec = doReq(t, r, srv, "mod@test.gg", http.MethodGet, "/tickets/counts", nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("counts failed: %d", rec.Code)
	}
	var counts modlv1.TicketCountsResponse
	_ = protojson.Unmarshal(rec.Body.Bytes(), &counts)
	if counts.Open != 1 || counts.Closed != 0 {
		t.Fatalf("unexpected counts: %+v", &counts)
	}

	// 3. Add note, tag, reply
	rec = doReq(t, r, srv, "mod@test.gg", http.MethodPost, "/tickets/"+created.Id+"/notes", &modlv1.AddNoteRequest{
		Text:       "Checked proxy logs",
		IssuerName: "Ignored",
	}, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("add note failed: %d %s", rec.Code, rec.Body.String())
	}
	var note modlv1.TicketNote
	_ = protojson.Unmarshal(rec.Body.Bytes(), &note)
	if note.IssuerName != "ModAlice" {
		t.Fatalf("expected IssuerName ModAlice, got %q", note.IssuerName)
	}

	rec = doReq(t, r, srv, "mod@test.gg", http.MethodPost, "/tickets/"+created.Id+"/tags", &modlv1.AddTagRequest{Tag: "urgent"}, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("add tag failed: %d", rec.Code)
	}
	rec = doReq(t, r, srv, "mod@test.gg", http.MethodDelete, "/tickets/"+created.Id+"/tags/lobby", nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("remove tag failed: %d", rec.Code)
	}

	rec = doReq(t, r, srv, "mod@test.gg", http.MethodPost, "/tickets/"+created.Id+"/replies", &modlv1.AddReplyRequest{
		Name:    "ModAlice",
		Content: "Restarted lobby 2",
		Staff:   true,
	}, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("add reply failed: %d %s", rec.Code, rec.Body.String())
	}

	// 4. Quick response closes ticket
	rec = doReq(t, r, srv, "mod@test.gg", http.MethodPost, "/tickets/"+created.Id+"/quick-response", &modlv1.QuickResponseRequest{
		CategoryId: "cat-1",
		ActionId:   "act-1",
	}, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("quick response failed: %d %s", rec.Code, rec.Body.String())
	}
	var qrRes modlv1.QuickResponseResult
	_ = protojson.Unmarshal(rec.Body.Bytes(), &qrRes)
	if !qrRes.Success || !qrRes.TicketClosed {
		t.Fatalf("expected quick response to close ticket, got %+v", &qrRes)
	}

	// 5. Reply to locked ticket returns 409
	rec = doReq(t, r, srv, "mod@test.gg", http.MethodPost, "/tickets/"+created.Id+"/replies", &modlv1.AddReplyRequest{
		Name:    "ModAlice",
		Content: "Should fail",
		Staff:   true,
	}, nil)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 on locked ticket reply, got %d", rec.Code)
	}

	// 6. Bulk reopen + assign
	rec = doReq(t, r, srv, "mod@test.gg", http.MethodPost, "/tickets/bulk", &modlv1.BulkTicketUpdateRequest{
		TicketIds: []string{created.Id},
		Locked:    proto.Bool(false),
		AssignTo:  proto.String("ModAlice"),
	}, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("bulk update failed: %d %s", rec.Code, rec.Body.String())
	}

	// 7. RBAC check: missing ticket.close.all -> 403 on /bulk
	delete(perms.perms, "ticket.close.all")
	rec = doReq(t, r, srv, "mod@test.gg", http.MethodPost, "/tickets/bulk", &modlv1.BulkTicketUpdateRequest{
		TicketIds: []string{created.Id},
	}, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without ticket.close.all, got %d", rec.Code)
	}
}

func TestTicketSubscriptions(t *testing.T) {
	perms := &stubPerms{superAdmin: true}
	_, h, srv, _ := setupTestEnv(t, perms)
	r := chi.NewRouter()
	h.RegisterPanelRoutes(r)

	subj := "Sub test"
	desc := "Need help"
	rec := doReq(t, r, srv, "mod@test.gg", http.MethodPost, "/tickets", &modlv1.CreateTicketRequest{
		Type:        "support",
		Subject:     &subj,
		Description: &desc,
	}, nil)
	var created modlv1.TicketResponse
	_ = protojson.Unmarshal(rec.Body.Bytes(), &created)

	// ModAlice replies -> auto-subscribes ModAlice
	rec = doReq(t, r, srv, "mod@test.gg", http.MethodPost, "/tickets/"+created.Id+"/replies", &modlv1.AddReplyRequest{
		Name:    "ModAlice",
		Content: "Looking into it",
		Staff:   true,
	}, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("reply failed: %d", rec.Code)
	}

	// HelperBob replies -> generates subscription update for ModAlice
	rec = doReq(t, r, srv, "helper@test.gg", http.MethodPost, "/tickets/"+created.Id+"/replies", &modlv1.AddReplyRequest{
		Name:    "HelperBob",
		Content: "Fixed it!",
		Staff:   true,
	}, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("helper reply failed: %d", rec.Code)
	}

	rec = doReq(t, r, srv, "mod@test.gg", http.MethodGet, "/ticket-subscriptions", nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get subscriptions failed: %d", rec.Code)
	}
	var subs modlv1.TicketSubscriptionsResponse
	_ = protojson.Unmarshal(rec.Body.Bytes(), &subs)
	if len(subs.Subscriptions) != 1 || subs.Subscriptions[0].TicketId != created.Id {
		t.Fatalf("expected 1 subscription, got %+v", subs.Subscriptions)
	}

	rec = doReq(t, r, srv, "mod@test.gg", http.MethodGet, "/ticket-subscriptions/updates", nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get updates failed: %d", rec.Code)
	}
	var upd modlv1.SubscriptionUpdatesResponse
	_ = protojson.Unmarshal(rec.Body.Bytes(), &upd)
	if len(upd.Updates) != 1 || upd.Updates[0].IsRead {
		t.Fatalf("expected 1 unread update, got %+v", upd.Updates)
	}

	rec = doReq(t, r, srv, "mod@test.gg", http.MethodPost, "/ticket-subscriptions/updates/"+upd.Updates[0].Id+"/read", nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("mark read failed: %d", rec.Code)
	}

	rec = doReq(t, r, srv, "mod@test.gg", http.MethodDelete, "/ticket-subscriptions/"+created.Id, nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("unsubscribe failed: %d", rec.Code)
	}
}

func TestPublicTicketsAndEmailVerification(t *testing.T) {
	database, h, srv, cfg := setupTestEnv(t, &stubPerms{superAdmin: true})
	r := chi.NewRouter()
	h.RegisterPublicRoutes(r)

	// 1. Create unfinished ticket + submit form
	rec := doReq(t, r, srv, "", http.MethodPost, "/tickets/unfinished", &modlv1.CreateTicketRequest{
		Type: "support",
	}, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create unfinished failed: %d %s", rec.Code, rec.Body.String())
	}
	var unf modlv1.CreateTicketResponse
	_ = protojson.Unmarshal(rec.Body.Bytes(), &unf)
	if unf.Ticket == nil || unf.Ticket.GetStatus() != "unfinished" {
		t.Fatalf("expected unfinished status, got %+v", &unf)
	}

	formStruct, _ := structpb.NewStruct(map[string]any{"issueDescription": "Lost my rank"})
	email := "Player@Example.com"
	rec = doReq(t, r, srv, "", http.MethodPost, "/tickets/"+unf.GetTicketId()+"/submit", &modlv1.SubmitTicketFormRequest{
		Subject:      "Missing rank",
		CreatorEmail: &email,
		FormData:     formStruct,
	}, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("submit failed: %d %s", rec.Code, rec.Body.String())
	}

	// 2. Unauthenticated GET requires email verification (403)
	rec = doReq(t, r, srv, "", http.MethodGet, "/tickets/"+unf.GetTicketId(), nil, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 verification required, got %d: %s", rec.Code, rec.Body.String())
	}
	var reqVer modlv1.TicketVerificationRequiredResponse
	_ = protojson.Unmarshal(rec.Body.Bytes(), &reqVer)
	if !reqVer.RequiresVerification || reqVer.EmailHint == "" {
		t.Fatalf("unexpected verification required payload: %+v", &reqVer)
	}

	// 3. Request verification code, replace hash with known code, verify
	rec = doReq(t, r, srv, "", http.MethodPost, "/tickets/"+unf.GetTicketId()+"/request-verification", nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("request-verification failed: %d %s", rec.Code, rec.Body.String())
	}
	knownHash := httpx.HashOneTimeCode("654321", cfg.AuthCodeHashSecret)
	_, err := database.Exec(`UPDATE ticket_verifications SET code_hash = ? WHERE server_id = ? AND ticket_id = ?`,
		knownHash, srv.ID, unf.GetTicketId())
	if err != nil {
		t.Fatalf("update verification code hash: %v", err)
	}

	// Wrong code returns 401
	rec = doReq(t, r, srv, "", http.MethodPost, "/tickets/"+unf.GetTicketId()+"/verify", &modlv1.VerifyTicketCodeRequest{Code: "000000"}, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for wrong code, got %d", rec.Code)
	}

	// Correct code returns token
	rec = doReq(t, r, srv, "", http.MethodPost, "/tickets/"+unf.GetTicketId()+"/verify", &modlv1.VerifyTicketCodeRequest{Code: "654321"}, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("verify code failed: %d %s", rec.Code, rec.Body.String())
	}
	var verResp modlv1.PublicTicketVerificationResponse
	_ = protojson.Unmarshal(rec.Body.Bytes(), &verResp)
	if !verResp.Success || verResp.Token == "" {
		t.Fatalf("expected token in verify response, got %+v", &verResp)
	}

	// 4. Access ticket + status + reply with token
	rec = doReq(t, r, srv, "", http.MethodGet, "/tickets/"+unf.GetTicketId()+"?token="+verResp.Token, nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get public ticket with token failed: %d %s", rec.Code, rec.Body.String())
	}
	var pubTicket modlv1.PublicTicketResponse
	_ = protojson.Unmarshal(rec.Body.Bytes(), &pubTicket)
	if _, hasEmail := pubTicket.Data.AsMap()["creatorEmail"]; hasEmail {
		t.Fatalf("creatorEmail must be redacted from public response!")
	}

	rec = doReq(t, r, srv, "", http.MethodPost, "/tickets/"+unf.GetTicketId()+"/replies?token="+verResp.Token, &modlv1.AddReplyRequest{
		Name:    "Player",
		Content: "Here is my receipt",
	}, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("public reply failed: %d %s", rec.Code, rec.Body.String())
	}
}

func TestMinecraftTicketsAndReportsV1AndV3(t *testing.T) {
	_, h, srv, _ := setupTestEnv(t, &stubPerms{superAdmin: true})

	rV1 := chi.NewRouter()
	h.RegisterMinecraftV1Routes(rV1)
	rV3 := chi.NewRouter()
	h.RegisterMinecraftV3Routes(rV3)

	rPub := chi.NewRouter()
	h.RegisterPublicRoutes(rPub)

	// Create an anonymous public ticket and claim it via MC V1
	subj := "Claim me"
	rec := doReq(t, rPub, srv, "", http.MethodPost, "/tickets", &modlv1.CreateTicketRequest{
		Type:    "support",
		Subject: &subj,
	}, nil)
	var pubCreated modlv1.CreateTicketResponse
	_ = protojson.Unmarshal(rec.Body.Bytes(), &pubCreated)

	playerUUID := "33333333-3333-3333-3333-333333333333"
	rec = doReq(t, rV1, srv, "", http.MethodPost, "/tickets/"+pubCreated.GetTicketId()+"/claim", nil, map[string]any{
		"playerUuid": playerUUID,
		"playerName": "Steve",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("claim failed: %d %s", rec.Code, rec.Body.String())
	}

	// Claiming by another player returns 409
	rec = doReq(t, rV1, srv, "", http.MethodPost, "/tickets/"+pubCreated.GetTicketId()+"/claim", nil, map[string]any{
		"playerUuid": "44444444-4444-4444-4444-444444444444",
		"playerName": "Alex",
	})
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 on second claim, got %d", rec.Code)
	}

	// Create chat report via MC V3
	reportedUUID := "55555555-5555-5555-5555-555555555555"
	reportedName := "Griefer"
	repSubj := "Toxic chat"
	rec = doReq(t, rV3, srv, "", http.MethodPost, "/tickets", &modlv1.MinecraftCreateTicketRequest{
		CreatorUuid:        playerUUID,
		CreatorName:        proto.String("Steve"),
		Type:               "chat",
		Subject:            &repSubj,
		Description:        proto.String("Saying bad words"),
		ReportedPlayerUuid: &reportedUUID,
		ReportedPlayerName: &reportedName,
		ChatMessages:       []string{"[12:00:00] Griefer: bad message"},
	}, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("v3 create report failed: %d %s", rec.Code, rec.Body.String())
	}
	var mcCreated modlv1.MinecraftCreateTicketResponse
	_ = protojson.Unmarshal(rec.Body.Bytes(), &mcCreated)

	// List V3 reports, assign, and resolve
	rec = doReq(t, rV3, srv, "", http.MethodGet, "/reports", nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("v3 list reports failed: %d", rec.Code)
	}
	var reportsResp modlv1.ReportsResponse
	_ = protojson.Unmarshal(rec.Body.Bytes(), &reportsResp)
	if len(reportsResp.Reports) != 1 || reportsResp.Reports[0].Id != mcCreated.TicketId {
		t.Fatalf("expected 1 report, got %+v", reportsResp.Reports)
	}

	rec = doReq(t, rV1, srv, "", http.MethodPost, "/reports/"+mcCreated.TicketId+"/assign", nil, map[string]any{"assignee": "ModAlice"})
	if rec.Code != http.StatusOK {
		t.Fatalf("v1 assign failed: %d", rec.Code)
	}
	rec = doReq(t, rV3, srv, "", http.MethodPost, "/reports/"+mcCreated.TicketId+"/resolve", &modlv1.ResolveReportRequest{
		ResolvedBy:   "ModAlice",
		Resolution:   "Muted",
		PunishmentId: "pun-1",
	}, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("v3 resolve failed: %d", rec.Code)
	}

	// Verify V1 ticket detail shows closed + staff reply
	rec = doReq(t, rV1, srv, "", http.MethodGet, "/tickets/"+mcCreated.TicketId, nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("v1 get detail failed: %d", rec.Code)
	}
	var detail ticket.MinecraftTicketDetailJSON
	_ = json.Unmarshal(rec.Body.Bytes(), &detail)
	if detail.Ticket.Status != "closed" || !detail.Ticket.Locked {
		t.Fatalf("expected closed locked report, got %+v", detail.Ticket)
	}
}
