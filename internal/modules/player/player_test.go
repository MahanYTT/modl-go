package player

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/config"
	"modl-go/internal/db"
	"modl-go/internal/httpx"
	"modl-go/internal/models"

	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type testPerms struct {
	superAdmin bool
	roleID     string
	perms      map[string]bool
}

func (p testPerms) IsSuperAdmin(_ *models.Server, _ string) bool { return p.superAdmin }
func (p testPerms) IsAuthorizedEmail(_ context.Context, _ *models.Server, _ string) bool {
	return true
}
func (p testPerms) PanelRoleID(_ context.Context, _ *models.Server, _ string) string {
	if p.roleID != "" {
		return p.roleID
	}
	return "mod"
}
func (p testPerms) HasPermission(_ context.Context, _ *models.Server, _, perm string) bool {
	if p.superAdmin {
		return true
	}
	return p.perms[perm]
}
func (p testPerms) HasAnyWithPrefix(_ context.Context, _ *models.Server, _, prefix string) bool {
	if p.superAdmin {
		return true
	}
	for k, ok := range p.perms {
		if ok && strings.HasPrefix(k, prefix) {
			return true
		}
	}
	return false
}

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	now := time.Now().UnixMilli()
	_, err = database.Exec(`
		INSERT INTO servers (id, server_name, custom_domain, admin_email, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		"srv-1", "TestNet", "panel.testnet.gg", "owner@testnet.gg", now, now,
	)
	if err != nil {
		t.Fatalf("insert test server: %v", err)
	}
	return database
}

func testServer() *models.Server {
	return &models.Server{
		ID:           "srv-1",
		ServerName:   "TestNet",
		CustomDomain: "panel.testnet.gg",
		AdminEmail:   "owner@testnet.gg",
	}
}

func mountTestRouter(h *Handler, email string) chi.Router {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := httpx.WithServer(r.Context(), testServer())
			if email != "" {
				c = httpx.WithSession(c, &models.AuthSession{ID: "sess-1", ServerID: "srv-1", Email: email})
			}
			next.ServeHTTP(w, r.WithContext(c))
		})
	})
	r.Route("/v1/panel", h.RegisterPanelRoutes)
	r.Route("/v1/public", h.RegisterPublicRoutes)
	r.Route("/v1/minecraft", h.RegisterMinecraftV1Routes)
	r.Route("/v2/minecraft", h.RegisterMinecraftV2Routes)
	r.Route("/v3/minecraft", h.RegisterMinecraftV3Routes)
	return r
}

func doJSONReq(t *testing.T, r chi.Router, method, target string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode json: %v", err)
		}
	}
	req := httptest.NewRequest(method, target, &buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func doProtoJSONReq(t *testing.T, r chi.Router, method, target string, msg proto.Message) *httptest.ResponseRecorder {
	t.Helper()
	var body []byte
	if msg != nil {
		var err error
		body, err = protojson.Marshal(msg)
		if err != nil {
			t.Fatalf("marshal protojson: %v", err)
		}
	}
	req := httptest.NewRequest(method, target, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func doProtoBinReq(t *testing.T, r chi.Router, method, target string, msg proto.Message) *httptest.ResponseRecorder {
	t.Helper()
	var body []byte
	if msg != nil {
		var err error
		body, err = proto.Marshal(msg)
		if err != nil {
			t.Fatalf("marshal proto: %v", err)
		}
	}
	req := httptest.NewRequest(method, target, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("Accept", "application/x-protobuf")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestPanelPlayerAndPunishmentLifecycle(t *testing.T) {
	database := setupTestDB(t)
	perms := testPerms{
		perms: map[string]bool{
			"punishment.view":      true,
			"punishment.modify":    true,
			"punishment.apply.ban": true,
		},
	}
	h := NewHandler(database, &config.Config{}, nil, perms)
	r := mountTestRouter(h, "mod@testnet.gg")

	pUUID := "11111111-1111-1111-1111-111111111111"

	// Create player
	rec := doProtoJSONReq(t, r, http.MethodPost, "/v1/panel/players", &modlv1.CreatePlayerRequest{
		MinecraftUuid: pUUID,
		Username:      "Notch",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create player got %d: %s", rec.Code, rec.Body.String())
	}

	// Duplicate player -> 409
	rec = doProtoJSONReq(t, r, http.MethodPost, "/v1/panel/players", &modlv1.CreatePlayerRequest{
		MinecraftUuid: pUUID,
		Username:      "Notch",
	})
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 conflict, got %d", rec.Code)
	}

	// Search player
	rec = doJSONReq(t, r, http.MethodGet, "/v1/panel/players?search=Notch", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("search players got %d: %s", rec.Code, rec.Body.String())
	}
	var searchResp modlv1.PlayerSearchResultsResponse
	if err := protojson.Unmarshal(rec.Body.Bytes(), &searchResp); err != nil || len(searchResp.Items) != 1 {
		t.Fatalf("expected 1 search result, got %v err=%v", len(searchResp.Items), err)
	}

	// Add username, note, IP
	if rec = doProtoJSONReq(t, r, http.MethodPost, "/v1/panel/players/"+pUUID+"/usernames", &modlv1.AddPlayerUsernameRequest{Username: "NotchNew"}); rec.Code != 200 {
		t.Fatalf("add username got %d: %s", rec.Code, rec.Body.String())
	}
	in := "Moderator"
	if rec = doProtoJSONReq(t, r, http.MethodPost, "/v1/panel/players/"+pUUID+"/notes", &modlv1.CreatePlayerNoteRequest{Text: "Suspicious activity", IssuerName: &in}); rec.Code != 200 {
		t.Fatalf("add note got %d: %s", rec.Code, rec.Body.String())
	}
	if rec = doProtoJSONReq(t, r, http.MethodPost, "/v1/panel/players/"+pUUID+"/ips", &modlv1.AddPlayerIpRequest{IpAddress: "203.0.113.5"}); rec.Code != 200 {
		t.Fatalf("add ip got %d: %s", rec.Code, rec.Body.String())
	}

	// Preview punishment (Ban = ordinal 3)
	rec = doJSONReq(t, r, http.MethodGet, "/v1/panel/players/"+pUUID+"/punishments/preview?typeOrdinal=3", nil)
	if rec.Code != 200 {
		t.Fatalf("preview got %d: %s", rec.Code, rec.Body.String())
	}
	var prev modlv1.PunishmentPreviewResponse
	if err := protojson.Unmarshal(rec.Body.Bytes(), &prev); err != nil || !prev.Success || prev.Regular == nil || !prev.CanBeAltBlocking {
		t.Fatalf("unexpected preview: %+v err=%v", &prev, err)
	}

	// Permission check: Mute (ordinal 2) not granted -> 403
	sev := "regular"
	reason := "Griefing"
	rec = doProtoJSONReq(t, r, http.MethodPost, "/v1/panel/players/"+pUUID+"/punishments", &modlv1.PanelCreatePunishmentRequest{
		TypeOrdinal: 2,
		Severity:    &sev,
		Reason:      &reason,
		IssuerName:  &in,
	})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for unpermitted mute, got %d: %s", rec.Code, rec.Body.String())
	}

	// Create Ban (ordinal 3) -> 200
	rec = doProtoJSONReq(t, r, http.MethodPost, "/v1/panel/players/"+pUUID+"/punishments", &modlv1.PanelCreatePunishmentRequest{
		TypeOrdinal: 3,
		Severity:    &sev,
		Reason:      &reason,
		IssuerName:  &in,
	})
	if rec.Code != 200 {
		t.Fatalf("create ban got %d: %s", rec.Code, rec.Body.String())
	}

	// Create second Ban -> should be queued as unstarted
	reason2 := "Second offense"
	rec = doProtoJSONReq(t, r, http.MethodPost, "/v1/panel/players/"+pUUID+"/punishments", &modlv1.PanelCreatePunishmentRequest{
		TypeOrdinal: 3,
		Severity:    &sev,
		Reason:      &reason2,
		IssuerName:  &in,
	})
	if rec.Code != 200 {
		t.Fatalf("create queued ban got %d: %s", rec.Code, rec.Body.String())
	}

	// Active punishments should only include 1st ban
	rec = doJSONReq(t, r, http.MethodGet, "/v1/panel/players/"+pUUID+"/punishments/active", nil)
	if rec.Code != 200 {
		t.Fatalf("active punishments got %d", rec.Code)
	}
	var activeResp modlv1.ActivePunishmentsResponse
	if err := protojson.Unmarshal(rec.Body.Bytes(), &activeResp); err != nil || len(activeResp.Items) != 1 {
		t.Fatalf("expected 1 active ban, got %d err=%v", len(activeResp.Items), err)
	}
	ban1ID := activeResp.Items[0].Id

	// Get player details (verify 2 punishments, 1 note, 2 usernames, 1 IP, gameplay points = 10 because 2nd ban is unstarted)
	rec = doJSONReq(t, r, http.MethodGet, "/v1/panel/players/"+pUUID, nil)
	if rec.Code != 200 {
		t.Fatalf("get player got %d: %s", rec.Code, rec.Body.String())
	}
	var detail modlv1.PlayerDetailResponse
	if err := protojson.Unmarshal(rec.Body.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal detail: %v", err)
	}
	if len(detail.Punishments) != 2 || len(detail.Usernames) != 2 || len(detail.Notes) != 1 || len(detail.IpAddresses) != 1 {
		t.Fatalf("unexpected detail counts: puns=%d users=%d notes=%d ips=%d",
			len(detail.Punishments), len(detail.Usernames), len(detail.Notes), len(detail.IpAddresses))
	}
	if detail.GameplayPoints != 10 {
		t.Fatalf("expected gameplayPoints=10 (only active ban counted), got %d", detail.GameplayPoints)
	}

	// Public appeal info BEFORE started -> 400
	rec = doJSONReq(t, r, http.MethodGet, "/v1/public/punishment/"+ban1ID+"/appeal-info", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unacknowledged appeal-info, got %d: %s", rec.Code, rec.Body.String())
	}

	// Acknowledge punishment start via Minecraft V1
	rec = doJSONReq(t, r, http.MethodPost, "/v1/minecraft/punishments/acknowledge", map[string]any{
		"punishmentId": ban1ID,
		"playerUuid":   pUUID,
		"success":      true,
	})
	if rec.Code != 200 {
		t.Fatalf("acknowledge got %d: %s", rec.Code, rec.Body.String())
	}

	// Seed an appeal ticket and verify Public appeal info -> 200
	nowMs := time.Now().UnixMilli()
	_, err := database.Exec(`
		INSERT INTO tickets (id, server_id, type, subject, status, punishment_id, created_at, updated_at)
		VALUES (?, 'srv-1', 'appeal', 'Appeal Ban', 'OPEN', ?, ?, ?)`,
		"app-1", ban1ID, nowMs, nowMs,
	)
	if err != nil {
		t.Fatalf("insert appeal ticket: %v", err)
	}
	rec = doJSONReq(t, r, http.MethodGet, "/v1/public/punishment/"+ban1ID+"/appeal-info", nil)
	if rec.Code != 200 {
		t.Fatalf("expected 200 appeal-info, got %d: %s", rec.Code, rec.Body.String())
	}
	var appealInfo modlv1.PublicPunishmentAppealInfoResponse
	if err := protojson.Unmarshal(rec.Body.Bytes(), &appealInfo); err != nil || appealInfo.ExistingAppeal == nil || !appealInfo.Appealable {
		t.Fatalf("unexpected appeal info: %+v err=%v", &appealInfo, err)
	}

	// Search punishments
	rec = doJSONReq(t, r, http.MethodGet, "/v1/panel/players/punishments/search?q=Grief&activeOnly=true", nil)
	if rec.Code != 200 {
		t.Fatalf("search punishments got %d", rec.Code)
	}
	var pSearch []PunishmentSearchResult
	if err := json.Unmarshal(rec.Body.Bytes(), &pSearch); err != nil || len(pSearch) != 1 || pSearch[0].ID != ban1ID {
		t.Fatalf("unexpected punishment search: %+v err=%v", pSearch, err)
	}

	// Add punishment note, evidence, and APPEAL_PARDON modification
	if rec = doProtoJSONReq(t, r, http.MethodPost, "/v1/panel/players/"+pUUID+"/punishments/"+ban1ID+"/notes", &modlv1.PanelAddPunishmentNoteRequest{
		Text:       "Reviewed clip",
		IssuerName: &in,
	}); rec.Code != 200 {
		t.Fatalf("add punishment note got %d", rec.Code)
	}
	evURL := "https://example.com/clip.mp4"
	if rec = doProtoJSONReq(t, r, http.MethodPost, "/v1/panel/players/"+pUUID+"/punishments/"+ban1ID+"/evidence", &modlv1.PanelAddEvidenceRequest{
		Type: "url",
		Url:  &evURL,
	}); rec.Code != 200 {
		t.Fatalf("add evidence got %d: %s", rec.Code, rec.Body.String())
	}
	appID := "app-1"
	appReason := "Accepted appeal"
	if rec = doProtoJSONReq(t, r, http.MethodPost, "/v1/panel/players/"+pUUID+"/punishments/"+ban1ID+"/modifications", &modlv1.PanelAddModificationRequest{
		Type:           ModAppealPardon,
		IssuerName:     &in,
		Reason:         &appReason,
		AppealTicketId: &appID,
	}); rec.Code != 200 {
		t.Fatalf("add modification got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify appeal ticket was closed and approved
	var wfStatus, tStatus string
	if err := database.QueryRow(`SELECT appeal_workflow_status, status FROM tickets WHERE id = 'app-1'`).Scan(&wfStatus, &tStatus); err != nil {
		t.Fatalf("query ticket: %v", err)
	}
	if wfStatus != "approved" || tStatus != "closed" {
		t.Fatalf("expected ticket approved/closed, got %s/%s", wfStatus, tStatus)
	}
}

func TestAltBlockingLinkedAccountsCascade(t *testing.T) {
	database := setupTestDB(t)
	h := NewHandler(database, &config.Config{}, nil, testPerms{superAdmin: true})
	r := mountTestRouter(h, "owner@testnet.gg")

	u1 := "22222222-2222-2222-2222-222222222222"
	u2 := "33333333-3333-3333-3333-333333333333"
	sharedIP := "198.51.100.42"

	// Both players log in from shared non-proxy IP
	for _, tc := range []struct{ u, name string }{{u1, "MainAcc"}, {u2, "AltAcc"}} {
		rec := doJSONReq(t, r, http.MethodPost, "/v1/minecraft/players/login", map[string]any{
			"minecraftUUID": tc.u,
			"username":      tc.name,
			"ip":            sharedIP,
			"ipInfo":        map[string]any{"country": "US", "proxy": false},
		})
		if rec.Code != 201 {
			t.Fatalf("login %s got %d: %s", tc.name, rec.Code, rec.Body.String())
		}
	}

	// Verify linked accounts panel endpoint
	rec := doJSONReq(t, r, http.MethodGet, "/v1/panel/players/"+u1+"/linked", nil)
	if rec.Code != 200 {
		t.Fatalf("get linked got %d", rec.Code)
	}
	var linkedResp modlv1.PanelLinkedAccountsResponse
	if err := protojson.Unmarshal(rec.Body.Bytes(), &linkedResp); err != nil || len(linkedResp.LinkedAccounts) != 1 {
		t.Fatalf("expected 1 linked account, got %d err=%v", len(linkedResp.LinkedAccounts), err)
	}

	// Issue alt-blocking ban on u1 via V1 dynamic create
	rec = doJSONReq(t, r, http.MethodPost, "/v1/minecraft/punishments/dynamic", map[string]any{
		"targetUuid":  u1,
		"issuerName":  "Admin",
		"typeOrdinal": 3,
		"severity":    "regular",
		"reason":      "Cheating",
		"data":        map[string]any{"altBlocking": true},
	})
	if rec.Code != 200 {
		t.Fatalf("create alt-blocking ban got %d: %s", rec.Code, rec.Body.String())
	}
	var dyn map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &dyn)
	parentBanID, _ := dyn["punishmentId"].(string)

	// Acknowledge u1 ban so it is started & active
	doJSONReq(t, r, http.MethodPost, "/v1/minecraft/punishments/acknowledge", map[string]any{
		"punishmentId": parentBanID,
		"playerUuid":   u1,
		"success":      true,
	})

	// Alt account u2 logs in -> should get auto-issued Linked Ban!
	rec = doJSONReq(t, r, http.MethodPost, "/v1/minecraft/players/login", map[string]any{
		"minecraftUUID": u2,
		"username":      "AltAcc",
		"ip":            sharedIP,
	})
	if rec.Code != 200 && rec.Code != 201 {
		t.Fatalf("alt login got %d: %s", rec.Code, rec.Body.String())
	}
	var loginBody struct {
		ActivePunishments []SimplePunishmentView `json:"activePunishments"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &loginBody)
	if len(loginBody.ActivePunishments) != 1 || loginBody.ActivePunishments[0].TypeOrdinal != LinkedBanTypeOrdinal {
		t.Fatalf("expected linked ban on alt login, got %+v", loginBody.ActivePunishments)
	}

	// Check panel linked-bans endpoint
	rec = doJSONReq(t, r, http.MethodGet, "/v1/panel/players/punishments/"+parentBanID+"/linked-bans", nil)
	if rec.Code != 200 {
		t.Fatalf("linked-bans got %d", rec.Code)
	}
	var lbResp modlv1.PanelLinkedBansResponse
	_ = protojson.Unmarshal(rec.Body.Bytes(), &lbResp)
	if len(lbResp.LinkedBans) != 1 || !lbResp.LinkedBans[0].Active {
		t.Fatalf("expected 1 active linked ban, got %+v", lbResp.LinkedBans)
	}

	// Change duration on parent -> cascades
	rec = doJSONReq(t, r, http.MethodPost, "/v1/minecraft/punishments/"+parentBanID+"/duration", map[string]any{
		"issuerName":  "Admin",
		"newDuration": int64(3600000),
	})
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "cascaded") {
		t.Fatalf("expected cascaded duration message, got %d: %s", rec.Code, rec.Body.String())
	}

	// Pardon parent -> cascades auto-pardon to linked ban
	rec = doJSONReq(t, r, http.MethodPost, "/v1/minecraft/punishments/"+parentBanID+"/pardon", map[string]any{
		"issuerName": "Admin",
		"reason":     "False positive",
	})
	if rec.Code != 200 {
		t.Fatalf("pardon parent got %d: %s", rec.Code, rec.Body.String())
	}
	rec = doJSONReq(t, r, http.MethodGet, "/v1/panel/players/punishments/"+parentBanID+"/linked-bans", nil)
	_ = protojson.Unmarshal(rec.Body.Bytes(), &lbResp)
	if len(lbResp.LinkedBans) != 1 || lbResp.LinkedBans[0].Active {
		t.Fatalf("expected linked ban to be inactive after parent pardon, got %+v", lbResp.LinkedBans)
	}
}

func TestMinecraftV2V3StartupSyncLogsAndUploadToken(t *testing.T) {
	database := setupTestDB(t)
	h := NewHandler(database, &config.Config{
		RealtimeEnabled: true,
		ModlDomain:      "rt.testnet.gg",
	}, nil, testPerms{superAdmin: true})
	r := mountTestRouter(h, "owner@testnet.gg")

	pUUID := "44444444-4444-4444-4444-444444444444"

	// Startup V2
	rec := doJSONReq(t, r, http.MethodPost, "/v2/minecraft/startup", map[string]any{
		"serverVersion": "1.21.4",
		"platformType":  "PAPER",
		"pluginVersion": "3.2.0",
		"maxPlayers":    100,
		"serverName":    "lobby-1",
	})
	if rec.Code != 200 {
		t.Fatalf("startup got %d: %s", rec.Code, rec.Body.String())
	}
	var startup map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &startup)
	if startup["realtimeEnabled"] != true || startup["panelUrl"] != "https://panel.testnet.gg" {
		t.Fatalf("unexpected startup payload: %+v", startup)
	}

	// V3 Binary Protobuf Login
	ipStr := "198.51.100.77"
	srvName := "lobby-1"
	rec = doProtoBinReq(t, r, http.MethodPost, "/v3/minecraft/players/login", &modlv1.PlayerLoginRequest{
		MinecraftUuid: pUUID,
		Username:      "Alex",
		IpAddress:     &ipStr,
		ServerName:    &srvName,
	})
	if rec.Code != 201 {
		t.Fatalf("v3 login got %d: %s", rec.Code, rec.Body.String())
	}
	var loginResp modlv1.PlayerLoginResponse
	if err := proto.Unmarshal(rec.Body.Bytes(), &loginResp); err != nil || loginResp.Status != 201 {
		t.Fatalf("unmarshal v3 login: %v status=%d", err, loginResp.Status)
	}

	// V3 Binary Create Punishment Dynamic (Ban with wipeAfterExpiry)
	sev := "low"
	reason := "Xray"
	st, _ := mapToStruct(map[string]any{"wipeAfterExpiry": true}), error(nil)
	rec = doProtoBinReq(t, r, http.MethodPost, "/v3/minecraft/punishments/dynamic", &modlv1.CreatePunishmentRequest{
		TargetUuid:  pUUID,
		TypeOrdinal: 3,
		Severity:    &sev,
		Reason:      &reason,
		Data:        st,
	})
	if rec.Code != 200 {
		t.Fatalf("v3 dynamic create got %d: %s", rec.Code, rec.Body.String())
	}
	var createResp modlv1.PunishmentCreateResponse
	if err := proto.Unmarshal(rec.Body.Bytes(), &createResp); err != nil || createResp.PunishmentId == "" {
		t.Fatalf("unmarshal v3 create: %v", err)
	}
	pid := createResp.PunishmentId

	// V3 Upload Token & verify SHA-256 hash in SQLite
	rec = doProtoBinReq(t, r, http.MethodPost, "/v3/minecraft/punishments/"+pid+"/upload-token", &modlv1.CreateEvidenceUploadTokenRequest{})
	if rec.Code != 200 {
		t.Fatalf("v3 upload token got %d", rec.Code)
	}
	var tokResp modlv1.EvidenceUploadTokenResponse
	if err := proto.Unmarshal(rec.Body.Bytes(), &tokResp); err != nil || tokResp.Token == "" {
		t.Fatalf("unmarshal token: %v", err)
	}
	sum := sha256.Sum256([]byte(tokResp.Token))
	var storedPID string
	if err := database.QueryRow(`SELECT punishment_id FROM evidence_upload_tokens WHERE token_hash = ?`, hex.EncodeToString(sum[:])).Scan(&storedPID); err != nil || storedPID != pid {
		t.Fatalf("expected stored token for %s, got %s err=%v", pid, storedPID, err)
	}

	// V3 Stat Wipe Acknowledge
	rec = doProtoBinReq(t, r, http.MethodPost, "/v3/minecraft/punishments/"+pid+"/stat-wipe-acknowledge", &modlv1.StatWipeAcknowledgeRequest{
		PunishmentId: pid,
		Success:      true,
	})
	if rec.Code != 200 {
		t.Fatalf("stat wipe ack got %d", rec.Code)
	}
	var swResp modlv1.StatWipeAcknowledgeResponse
	if err := proto.Unmarshal(rec.Body.Bytes(), &swResp); err != nil || !swResp.Success {
		t.Fatalf("expected stat wipe success, got %+v err=%v", &swResp, err)
	}

	// V3 Sync with online players and chat logs
	nowMs := time.Now().UnixMilli()
	rec = doProtoBinReq(t, r, http.MethodPost, "/v3/minecraft/players/sync", &modlv1.SyncRequest{
		ServerName: &srvName,
		OnlinePlayers: []*modlv1.SyncOnlinePlayer{{
			Uuid:      pUUID,
			Username:  "Alex",
			IpAddress: ipStr,
		}},
		ChatLogs: []*modlv1.SyncChatLogEntry{{
			Uuid:      pUUID,
			Username:  "Alex",
			Message:   "hello world",
			Server:    "lobby-1",
			Timestamp: nowMs,
		}},
	})
	if rec.Code != 200 {
		t.Fatalf("v3 sync got %d: %s", rec.Code, rec.Body.String())
	}
	var syncResp modlv1.SyncResponse
	if err := proto.Unmarshal(rec.Body.Bytes(), &syncResp); err != nil || syncResp.Data == nil {
		t.Fatalf("unmarshal v3 sync: %v", err)
	}
	if len(syncResp.Data.PendingPunishments) != 1 || syncResp.Data.PendingPunishments[0].Punishment.Id != pid {
		t.Fatalf("expected pending punishment %s in sync, got %+v", pid, syncResp.Data.PendingPunishments)
	}

	// Verify chat log persisted & queryable via V1
	rec = doJSONReq(t, r, http.MethodGet, "/v1/minecraft/players/"+pUUID+"/chat-logs?limit=10", nil)
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "hello world") {
		t.Fatalf("expected chat log entry, got %d: %s", rec.Code, rec.Body.String())
	}

	// V1 Pardon player by name
	rec = doJSONReq(t, r, http.MethodPost, "/v1/minecraft/players/pardon", map[string]any{
		"playerName": "Alex",
		"issuerName": "Console",
		"reason":     "Clearing",
	})
	if rec.Code != 200 {
		t.Fatalf("pardon player got %d: %s", rec.Code, rec.Body.String())
	}
	var pardonBody map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &pardonBody)
	if pardonBody["success"] != true {
		t.Fatalf("expected pardon success=true, got %+v", pardonBody)
	}
}

func TestRestrictionAutoPardonOnUsernameChange(t *testing.T) {
	database := setupTestDB(t)
	// Seed custom punishment type ordinal 5 with permanentUntilUsernameChange=true
	customTypes := &modlv1.PanelPunishmentTypesResponse{
		PunishmentTypes: []*modlv1.PunishmentType{
			{
				Id:                           ptrInt32(5),
				Ordinal:                      ptrInt32(5),
				Name:                         "NameBan",
				Category:                     "SOCIAL",
				PermanentUntilUsernameChange: ptrBool(true),
			},
		},
	}
	b, _ := protojson.Marshal(customTypes)
	_, err := database.Exec(`
		INSERT INTO settings (id, server_id, type, data_json, version, updated_at)
		VALUES ('set-1', 'srv-1', 'punishmentTypes', ?, 1, ?)`,
		string(b), time.Now().UnixMilli(),
	)
	if err != nil {
		t.Fatalf("insert settings: %v", err)
	}

	h := NewHandler(database, &config.Config{}, nil, testPerms{superAdmin: true})
	r := mountTestRouter(h, "owner@testnet.gg")

	pUUID := "55555555-5555-5555-5555-555555555555"
	doJSONReq(t, r, http.MethodPost, "/v1/minecraft/players/login", map[string]any{
		"minecraftUUID": pUUID,
		"username":      "BadName",
		"ip":            "198.51.100.99",
	})
	rec := doJSONReq(t, r, http.MethodPost, "/v1/minecraft/punishments/dynamic", map[string]any{
		"targetUuid":  pUUID,
		"issuerName":  "Mod",
		"typeOrdinal": 5,
		"reason":      "Inappropriate name",
	})
	if rec.Code != 200 {
		t.Fatalf("create NameBan got %d: %s", rec.Code, rec.Body.String())
	}

	// Player logs back in with new clean username -> auto-pardoned!
	rec = doJSONReq(t, r, http.MethodPost, "/v1/minecraft/players/login", map[string]any{
		"minecraftUUID": pUUID,
		"username":      "GoodName",
		"ip":            "198.51.100.99",
	})
	if rec.Code != 200 {
		t.Fatalf("re-login got %d", rec.Code)
	}
	var loginBody struct {
		ActivePunishments []SimplePunishmentView `json:"activePunishments"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &loginBody)
	if len(loginBody.ActivePunishments) != 0 {
		t.Fatalf("expected NameBan to be auto-pardoned on username change, got %+v", loginBody.ActivePunishments)
	}
}

func ptrInt32(v int32) *int32 { return &v }
func ptrBool(v bool) *bool    { return &v }

func TestPermissionsAndRemainingEndpoints(t *testing.T) {
	database := setupTestDB(t)
	// Moderator who ONLY has punishment.apply.ban (no punishment.modify)
	modOnlyApply := testPerms{
		perms: map[string]bool{
			"punishment.apply.ban": true,
		},
	}
	h := NewHandler(database, &config.Config{}, nil, modOnlyApply)
	r := mountTestRouter(h, "helper@testnet.gg")

	pUUID := "66666666-6666-6666-6666-666666666666"
	// Login player first
	rec := doJSONReq(t, r, http.MethodPost, "/v1/minecraft/players/login", map[string]any{
		"minecraftUUID": pUUID,
		"username":      "TargetGuy",
		"ip":            "203.0.113.88",
		"serverName":    "survival-1",
	})
	if rec.Code != 201 {
		t.Fatalf("login got %d", rec.Code)
	}

	// Mod with only punishment.apply.ban CAN issue a Ban
	sev := "low"
	reason := "Rule break"
	rec = doProtoJSONReq(t, r, http.MethodPost, "/v1/panel/players/"+pUUID+"/punishments", &modlv1.PanelCreatePunishmentRequest{
		TypeOrdinal: 3,
		Severity:    &sev,
		Reason:      &reason,
	})
	if rec.Code != 200 {
		t.Fatalf("expected 200 for punishment.apply.ban, got %d: %s", rec.Code, rec.Body.String())
	}

	// Mod without punishment.modify CANNOT add player note via panel
	rec = doProtoJSONReq(t, r, http.MethodPost, "/v1/panel/players/"+pUUID+"/notes", &modlv1.CreatePlayerNoteRequest{
		Text: "Blocked note",
	})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for add note without punishment.modify, got %d", rec.Code)
	}

	// Online players, update-server, submit-ip-info, disconnect
	if rec = doJSONReq(t, r, http.MethodGet, "/v1/minecraft/players/online", nil); rec.Code != 200 || !strings.Contains(rec.Body.String(), "TargetGuy") {
		t.Fatalf("expected TargetGuy online, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec = doJSONReq(t, r, http.MethodPost, "/v1/minecraft/players/update-server", map[string]any{
		"minecraftUuid": pUUID, "serverName": "skyblock-1",
	}); rec.Code != 200 {
		t.Fatalf("update-server got %d", rec.Code)
	}
	if rec = doJSONReq(t, r, http.MethodPost, "/v1/minecraft/players/submit-ip-info", map[string]any{
		"minecraftUUID": pUUID, "ip": "203.0.113.88", "country": "CA", "region": "ON", "asn": "AS1234",
	}); rec.Code != 200 {
		t.Fatalf("submit-ip-info got %d", rec.Code)
	}
	if rec = doJSONReq(t, r, http.MethodPost, "/v1/minecraft/players/disconnect", map[string]any{
		"minecraftUuid": pUUID, "sessionDurationMs": 120000,
	}); rec.Code != 200 {
		t.Fatalf("disconnect got %d", rec.Code)
	}

	// Paginated punishments, notes, reports
	if rec = doJSONReq(t, r, http.MethodGet, "/v1/minecraft/players/"+pUUID+"/punishments?page=1&limit=5", nil); rec.Code != 200 || !strings.Contains(rec.Body.String(), "Rule break") {
		t.Fatalf("get punishments v1 got %d: %s", rec.Code, rec.Body.String())
	}
	if rec = doJSONReq(t, r, http.MethodGet, "/v1/minecraft/players/"+pUUID+"/notes?page=1&limit=5", nil); rec.Code != 200 {
		t.Fatalf("get notes v1 got %d", rec.Code)
	}
	if rec = doJSONReq(t, r, http.MethodGet, "/v1/minecraft/players/"+pUUID+"/reports", nil); rec.Code != 200 {
		t.Fatalf("get reports v1 got %d", rec.Code)
	}

	// Seed notification & acknowledge via V3
	nowMs := time.Now().UnixMilli()
	_, err := database.Exec(`
		INSERT INTO player_notifications (id, server_id, player_uuid, type, message, created_at)
		VALUES ('notif-1', 'srv-1', ?, 'INFO', 'Hello', ?)`, pUUID, nowMs)
	if err != nil {
		t.Fatalf("insert notif: %v", err)
	}
	rec = doProtoBinReq(t, r, http.MethodPost, "/v3/minecraft/notifications/acknowledge", &modlv1.AcknowledgeNotificationsRequest{
		PlayerUuid:      pUUID,
		NotificationIds: []string{"notif-1"},
	})
	if rec.Code != 200 {
		t.Fatalf("v3 ack notifications got %d: %s", rec.Code, rec.Body.String())
	}
	var delivered int
	if err := database.QueryRow(`SELECT delivered FROM player_notifications WHERE id = 'notif-1'`).Scan(&delivered); err != nil || delivered != 1 {
		t.Fatalf("expected notification delivered=1, got %d err=%v", delivered, err)
	}

	// V2 sync with recent ticket checks relative ticketUrl conversion
	_, _ = database.Exec(`
		INSERT INTO tickets (id, server_id, type, subject, status, creator_name, created_at, updated_at)
		VALUES ('t-1', 'srv-1', 'support', 'Help', 'open', 'TargetGuy', ?, ?)`, nowMs+1000, nowMs+1000)
	rec = doJSONReq(t, r, http.MethodPost, "/v2/minecraft/players/sync", map[string]any{
		"lastSyncTimestamp": time.UnixMilli(nowMs - 5000).UTC().Format(time.RFC3339Nano),
		"onlinePlayers":     []any{},
		"serverName":        "skyblock-1",
	})
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"/ticket/t-1"`) {
		t.Fatalf("expected v2 sync relative ticketUrl /ticket/t-1, got %d: %s", rec.Code, rec.Body.String())
	}

	// Public avatar invalid UUID -> 400
	if rec = doJSONReq(t, r, http.MethodGet, "/v1/public/players/avatar/not-a-uuid", nil); rec.Code != 400 {
		t.Fatalf("expected 400 for invalid avatar uuid, got %d", rec.Code)
	}
}

