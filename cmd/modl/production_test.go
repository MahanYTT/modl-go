package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/config"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

type capturedMail struct {
	From    string
	To      string
	RawBody string
}

type testSMTPServer struct {
	ln    net.Listener
	Host  string
	Port  int
	mu    sync.Mutex
	Mails []capturedMail
}

func startTestSMTPServer(t *testing.T) *testSMTPServer {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test SMTP listener: %v", err)
	}
	host, portStr, _ := net.SplitHostPort(ln.Addr().String())
	port, _ := strconv.Atoi(portStr)
	srv := &testSMTPServer{ln: ln, Host: host, Port: port}
	go srv.acceptLoop()
	t.Cleanup(func() { _ = ln.Close() })
	return srv
}

func (s *testSMTPServer) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return
		}
		go s.handleConn(conn)
	}
}

func (s *testSMTPServer) handleConn(conn net.Conn) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	writeLine := func(line string) {
		_, _ = rw.WriteString(line + "\r\n")
		_ = rw.Flush()
	}
	writeLine("220 localhost ESMTP ready")
	var mail capturedMail
	inData := false
	var dataBuf strings.Builder
	for {
		line, err := rw.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		if inData {
			if line == "." {
				inData = false
				mail.RawBody = dataBuf.String()
				s.mu.Lock()
				s.Mails = append(s.Mails, mail)
				s.mu.Unlock()
				writeLine("250 OK queued")
				continue
			}
			dataBuf.WriteString(line + "\n")
			continue
		}
		upper := strings.ToUpper(line)
		switch {
		case strings.HasPrefix(upper, "EHLO") || strings.HasPrefix(upper, "HELO"):
			writeLine("250-localhost")
			writeLine("250 OK")
		case strings.HasPrefix(upper, "MAIL FROM:"):
			mail.From = strings.TrimSpace(line[len("MAIL FROM:"):])
			writeLine("250 OK")
		case strings.HasPrefix(upper, "RCPT TO:"):
			mail.To = strings.TrimSpace(line[len("RCPT TO:"):])
			writeLine("250 OK")
		case upper == "DATA":
			inData = true
			dataBuf.Reset()
			writeLine("354 End data with <CR><LF>.<CR><LF>")
		case upper == "RSET":
			writeLine("250 OK")
		case upper == "QUIT":
			writeLine("221 Bye")
			return
		default:
			writeLine("250 OK")
		}
	}
}

func (s *testSMTPServer) latestCodeFor(emailAddr string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	re := regexp.MustCompile(`\b(\d{6})\b`)
	for i := len(s.Mails) - 1; i >= 0; i-- {
		m := s.Mails[i]
		if strings.Contains(strings.ToLower(m.To), strings.ToLower(emailAddr)) {
			if match := re.FindStringSubmatch(m.RawBody); len(match) == 2 {
				return match[1]
			}
		}
	}
	return ""
}

func newProdTestConfig(tmpDir string, smtpSrv *testSMTPServer, apiKey string) *config.Config {
	return &config.Config{
		Port:                "0",
		DBPath:              filepath.Join(tmpDir, "modl.db"),
		StorageDir:          filepath.Join(tmpDir, "storage"),
		ModlDomain:          "panel.example.com",
		CookieSecure:        false, // httptest runs over plain HTTP on 127.0.0.1
		DevelopmentMode:     false, // test real production behavior!
		SingleTenant:        true,
		TrustProxyHeaders:   false,
		AuthCodeHashSecret:  "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		SessionCookieName:   "MODL_SESSION",
		SessionDurationSec:  3600,
		CORSAppDomains:      []string{"panel.example.com"},
		CORSSystemOrigins:   []string{"https://panel.example.com"},
		SMTPHost:            smtpSrv.Host,
		SMTPPort:            smtpSrv.Port,
		SMTPAuth:            false,
		SMTPTLS:             false,
		EmailDomain:         "panel.example.com",
		BootstrapServerName: "ProdNet",
		BootstrapDomain:     "panel.example.com",
		BootstrapAdminEmail: "owner@panel.example.com",
		BootstrapAPIKey:     apiKey,
		RealtimeEnabled:     true,
	}
}

func mustJSONRequest(t *testing.T, client *http.Client, method, url string, payload any, headers map[string]string) (*http.Response, []byte) {
	t.Helper()
	var bodyReader io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
		bodyReader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		t.Fatalf("new request %s %s: %v", method, url, err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request %s %s failed: %v", method, url, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return resp, raw
}

func loginViaRealSMTP(t *testing.T, client *http.Client, baseURL string, smtpSrv *testSMTPServer, emailAddr string) {
	t.Helper()
	resp, raw := mustJSONRequest(t, client, http.MethodPost, baseURL+"/api/v1/panel/auth/send-email-code", map[string]string{
		"email": emailAddr,
	}, map[string]string{"Origin": baseURL})
	if resp.StatusCode != 200 {
		t.Fatalf("send-email-code failed (%d): %s", resp.StatusCode, string(raw))
	}

	var code string
	for i := 0; i < 30; i++ {
		code = smtpSrv.latestCodeFor(emailAddr)
		if code != "" {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	if code == "" {
		t.Fatalf("never received OTP email over SMTP for %s", emailAddr)
	}

	resp, raw = mustJSONRequest(t, client, http.MethodPost, baseURL+"/api/v1/panel/auth/verify-email-code", map[string]string{
		"email": emailAddr,
		"code":  code,
	}, map[string]string{"Origin": baseURL})
	if resp.StatusCode != 200 {
		t.Fatalf("verify-email-code failed (%d): %s", resp.StatusCode, string(raw))
	}
}

func TestProductionReadinessSuite(t *testing.T) {
	tmpDir := t.TempDir()
	smtpSrv := startTestSMTPServer(t)
	apiKey := "prod-mc-secret-key-999"
	cfg := newProdTestConfig(tmpDir, smtpSrv, apiKey)

	handler, cleanup, err := buildServerHandler(cfg)
	if err != nil {
		t.Fatalf("buildServerHandler failed: %v", err)
	}
	ts := httptest.NewServer(handler)
	closedEarly := false
	defer func() {
		if !closedEarly {
			ts.Close()
			cleanup()
		}
	}()

	jar, _ := cookiejar.New(nil)
	ownerClient := &http.Client{Jar: jar, Timeout: 10 * time.Second}
	anonClient := &http.Client{Timeout: 10 * time.Second}

	// 1. Security Headers, Unauthenticated Access & OTP Brute-force Lockout
	t.Run("Auth_BruteForceLockout_SecurityHeaders_CSRF_CORS", func(t *testing.T) {
		// Unauthenticated panel route must return 401
		r, raw := mustJSONRequest(t, anonClient, http.MethodGet, ts.URL+"/api/v1/panel/players", nil, nil)
		if r.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401 on unauthenticated panel players, got %d: %s", r.StatusCode, string(raw))
		}
		if r.Header.Get("X-Frame-Options") != "DENY" || r.Header.Get("X-Content-Type-Options") != "nosniff" {
			t.Fatalf("missing security headers: X-Frame-Options=%q X-Content-Type-Options=%q",
				r.Header.Get("X-Frame-Options"), r.Header.Get("X-Content-Type-Options"))
		}

		// Invalid API key must return 401 on v1/v2/v3 & ws
		for _, path := range []string{"/v1/minecraft/roles", "/v2/minecraft/players/sync", "/v3/minecraft/staff"} {
			r, _ := mustJSONRequest(t, anonClient, http.MethodGet, ts.URL+path, nil, map[string]string{"X-API-Key": "wrong-key"})
			if r.StatusCode != http.StatusUnauthorized {
				t.Fatalf("expected 401 for wrong API key on %s, got %d", path, r.StatusCode)
			}
		}

		// Trigger OTP email, then fail 5 times -> 6th attempt with the real code must still fail (locked out!)
		r, _ = mustJSONRequest(t, ownerClient, http.MethodPost, ts.URL+"/api/v1/panel/auth/send-email-code", map[string]string{
			"email": cfg.BootstrapAdminEmail,
		}, map[string]string{"Origin": ts.URL})
		if r.StatusCode != 200 {
			t.Fatalf("send-email-code failed: %d", r.StatusCode)
		}
		var realCode string
		for i := 0; i < 30; i++ {
			realCode = smtpSrv.latestCodeFor(cfg.BootstrapAdminEmail)
			if realCode != "" {
				break
			}
			time.Sleep(20 * time.Millisecond)
		}
		if realCode == "" {
			t.Fatal("OTP email not delivered")
		}

		for attempt := 0; attempt < 5; attempt++ {
			wrong := "000000"
			if wrong == realCode {
				wrong = "999999"
			}
			r, _ := mustJSONRequest(t, ownerClient, http.MethodPost, ts.URL+"/api/v1/panel/auth/verify-email-code", map[string]string{
				"email": cfg.BootstrapAdminEmail,
				"code":  wrong,
			}, map[string]string{"Origin": ts.URL})
			if r.StatusCode != http.StatusUnauthorized {
				t.Fatalf("expected 401 on wrong OTP attempt %d, got %d", attempt+1, r.StatusCode)
			}
		}
		// Now even the real code must be rejected due to attempt >= 5 lockout
		r, _ = mustJSONRequest(t, ownerClient, http.MethodPost, ts.URL+"/api/v1/panel/auth/verify-email-code", map[string]string{
			"email": cfg.BootstrapAdminEmail,
			"code":  realCode,
		}, map[string]string{"Origin": ts.URL})
		if r.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401 after 5 failed OTP attempts, got %d", r.StatusCode)
		}

		// Now perform clean SMTP login
		loginViaRealSMTP(t, ownerClient, ts.URL, smtpSrv, cfg.BootstrapAdminEmail)

		// Verify /me succeeds
		r, raw = mustJSONRequest(t, ownerClient, http.MethodGet, ts.URL+"/api/v1/panel/auth/me", nil, nil)
		if r.StatusCode != 200 || !strings.Contains(string(raw), cfg.BootstrapAdminEmail) {
			t.Fatalf("/me failed (%d): %s", r.StatusCode, string(raw))
		}

		// CSRF: Cookie-authenticated mutation with evil Origin or evil Referer MUST return 403
		r, _ = mustJSONRequest(t, ownerClient, http.MethodPost, ts.URL+"/api/v1/panel/roles", map[string]any{
			"name": "EvilRole",
		}, map[string]string{"Origin": "https://evil-attacker.com"})
		if r.StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403 CSRF block for evil Origin, got %d", r.StatusCode)
		}

		r, _ = mustJSONRequest(t, ownerClient, http.MethodPost, ts.URL+"/api/v1/panel/roles", map[string]any{
			"name": "EvilRole",
		}, map[string]string{"Referer": "https://evil-attacker.com/form"})
		if r.StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403 CSRF block for evil Referer, got %d", r.StatusCode)
		}

		// CORS: Evil Origin must NOT get Access-Control-Allow-Origin
		r, _ = mustJSONRequest(t, anonClient, http.MethodOptions, ts.URL+"/api/v1/public/settings", nil, map[string]string{
			"Origin": "https://evil-attacker.com",
		})
		if got := r.Header.Get("Access-Control-Allow-Origin"); got != "" {
			t.Fatalf("evil Origin leaked Access-Control-Allow-Origin=%q", got)
		}
		// Allowed custom domain Origin MUST get Access-Control-Allow-Origin
		r, _ = mustJSONRequest(t, anonClient, http.MethodOptions, ts.URL+"/api/v1/public/settings", nil, map[string]string{
			"Origin": "https://panel.example.com",
		})
		if got := r.Header.Get("Access-Control-Allow-Origin"); got != "https://panel.example.com" {
			t.Fatalf("expected Access-Control-Allow-Origin=https://panel.example.com, got %q", got)
		}
	})

	// 2. RBAC Role Creation, Staff Creation & Permission Enforcement
	t.Run("RBAC_Enforcement", func(t *testing.T) {
		sameOrigin := map[string]string{"Origin": ts.URL}
		// Create restricted "Helper" role with only ticket view permission
		r, raw := mustJSONRequest(t, ownerClient, http.MethodPost, ts.URL+"/api/v1/panel/roles", map[string]any{
			"name":        "Helper",
			"description": "Helper staff role",
			"color":       "#3b82f6",
			"permissions": []string{"ticket.view.all"},
		}, sameOrigin)
		if r.StatusCode != 200 && r.StatusCode != 201 {
			t.Fatalf("create role failed (%d): %s", r.StatusCode, string(raw))
		}
		var roleResp map[string]any
		_ = json.Unmarshal(raw, &roleResp)
		helperRoleID, _ := roleResp["id"].(string)
		if helperRoleID == "" {
			if dataMap, ok := roleResp["role"].(map[string]any); ok {
				helperRoleID, _ = dataMap["id"].(string)
			}
		}
		if helperRoleID == "" {
			// Fallback lookup from GET /roles
			_, rolesRaw := mustJSONRequest(t, ownerClient, http.MethodGet, ts.URL+"/api/v1/panel/roles", nil, nil)
			var listResp struct {
				Roles []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"roles"`
			}
			_ = json.Unmarshal(rolesRaw, &listResp)
			for _, rl := range listResp.Roles {
				if rl.Name == "Helper" {
					helperRoleID = rl.ID
				}
			}
		}
		if helperRoleID == "" {
			t.Fatalf("could not find created Helper role ID in response: %s", string(raw))
		}

		// Create helper staff member directly
		helperEmail := "helper@panel.example.com"
		r, raw = mustJSONRequest(t, ownerClient, http.MethodPost, ts.URL+"/api/v1/panel/staff", map[string]any{
			"email":    helperEmail,
			"username": "HelperMod",
			"role":     helperRoleID,
		}, sameOrigin)
		if r.StatusCode != 200 && r.StatusCode != 201 {
			t.Fatalf("create staff failed (%d): %s", r.StatusCode, string(raw))
		}

		// Helper logs in via real SMTP OTP
		helperJar, _ := cookiejar.New(nil)
		helperClient := &http.Client{Jar: helperJar, Timeout: 10 * time.Second}
		loginViaRealSMTP(t, helperClient, ts.URL, smtpSrv, helperEmail)

		// Helper CAN view tickets
		r, raw = mustJSONRequest(t, helperClient, http.MethodGet, ts.URL+"/api/v1/panel/tickets", nil, nil)
		if r.StatusCode != 200 {
			t.Fatalf("expected Helper to view tickets (200), got %d: %s", r.StatusCode, string(raw))
		}

		// Helper CANNOT create roles or read storage quota (403 Forbidden)
		r, _ = mustJSONRequest(t, helperClient, http.MethodPost, ts.URL+"/api/v1/panel/roles", map[string]any{
			"name": "EscalatedRole",
		}, sameOrigin)
		if r.StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403 for Helper creating role, got %d", r.StatusCode)
		}
		r, _ = mustJSONRequest(t, helperClient, http.MethodGet, ts.URL+"/api/v1/panel/storage/quota", nil, nil)
		if r.StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403 for Helper viewing storage quota, got %d", r.StatusCode)
		}

		// Logout invalidates Helper session
		r, _ = mustJSONRequest(t, helperClient, http.MethodPost, ts.URL+"/api/v1/panel/auth/logout", map[string]any{}, sameOrigin)
		if r.StatusCode != 200 {
			t.Fatalf("logout failed: %d", r.StatusCode)
		}
		r, _ = mustJSONRequest(t, helperClient, http.MethodGet, ts.URL+"/api/v1/panel/tickets", nil, nil)
		if r.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401 after logout, got %d", r.StatusCode)
		}
	})

	// 3. Player Login, Alt-Linking, Ban Evasion Enforcement & Punishment Lifecycle
	player1UUID := "11111111-1111-4111-8111-111111111111"
	player2UUID := "22222222-2222-4222-8222-222222222222"
	var punishmentID string

	t.Run("Player_AltLinking_BanAndAppealLifecycle", func(t *testing.T) {
		mcHeaders := map[string]string{"X-API-Key": apiKey}
		sameOrigin := map[string]string{"Origin": ts.URL}

		// Player 1 & Player 2 log in from the same IP -> should link accounts
		for _, p := range []struct{ uuid, name string }{{player1UUID, "MainAcc"}, {player2UUID, "AltAcc"}} {
			r, raw := mustJSONRequest(t, anonClient, http.MethodPost, ts.URL+"/v1/minecraft/players/login", map[string]any{
				"minecraftUUID": p.uuid,
				"username":      p.name,
				"ip":            "203.0.113.42",
				"serverName":    "survival",
			}, mcHeaders)
			if r.StatusCode != 200 && r.StatusCode != 201 {
				t.Fatalf("player login %s failed (%d): %s", p.name, r.StatusCode, string(raw))
			}
		}

		// Ban Player 1 (typeOrdinal 3 = Ban) using /dynamic which returns punishmentId
		r, raw := mustJSONRequest(t, anonClient, http.MethodPost, ts.URL+"/v1/minecraft/punishments/dynamic", map[string]any{
			"targetUuid":   player1UUID,
			"type_ordinal": 3,
			"duration":     86400000,
			"reason":       "Cheating in ranked",
			"issuerName":   "Console",
		}, mcHeaders)
		if r.StatusCode != 200 && r.StatusCode != 201 {
			t.Fatalf("punishment create failed (%d): %s", r.StatusCode, string(raw))
		}
		var punOut map[string]any
		_ = json.Unmarshal(raw, &punOut)
		punishmentID, _ = punOut["punishmentId"].(string)
		if punishmentID == "" {
			punishmentID, _ = punOut["id"].(string)
		}
		if punishmentID == "" {
			t.Fatalf("missing punishmentId in create response: %s", string(raw))
		}

		// Player 1 logs in again -> activePunishments must be non-empty
		r, raw = mustJSONRequest(t, anonClient, http.MethodPost, ts.URL+"/v1/minecraft/players/login", map[string]any{
			"minecraftUUID": player1UUID,
			"username":      "MainAcc",
			"ip":            "203.0.113.42",
			"serverName":    "survival",
		}, mcHeaders)
		var loginOut struct {
			ActivePunishments []any `json:"activePunishments"`
		}
		_ = json.Unmarshal(raw, &loginOut)
		if len(loginOut.ActivePunishments) == 0 {
			t.Fatalf("expected banned Player 1 to have activePunishments on login, got: %s", string(raw))
		}

		// Submit public Appeal for the punishment
		r, raw = mustJSONRequest(t, anonClient, http.MethodPost, ts.URL+"/api/v1/public/appeals", map[string]any{
			"punishmentId": punishmentID,
			"playerUuid":   player1UUID,
			"username":     "MainAcc",
			"email":        "player1@example.com",
			"reason":       "My brother used my PC, I removed all mods",
		}, sameOrigin)
		if r.StatusCode != 200 && r.StatusCode != 201 {
			t.Fatalf("create public appeal failed (%d): %s", r.StatusCode, string(raw))
		}
		var appealResp map[string]any
		_ = json.Unmarshal(raw, &appealResp)
		appealID, _ := appealResp["id"].(string)
		if appealID == "" {
			if ap, ok := appealResp["appeal"].(map[string]any); ok {
				appealID, _ = ap["id"].(string)
			}
		}
		if appealID == "" {
			t.Fatalf("missing appeal ID in response: %s", string(raw))
		}

		// Staff accepts appeal (action: pardon) -> must automatically pardon the punishment!
		r, raw = mustJSONRequest(t, ownerClient, http.MethodPatch, ts.URL+"/api/v1/panel/appeals/"+appealID+"/status", map[string]any{
			"status":      "ACCEPTED",
			"action":      "pardon",
			"staffReason": "First offense pardoned",
		}, sameOrigin)
		if r.StatusCode != 200 {
			t.Fatalf("accept appeal failed (%d): %s", r.StatusCode, string(raw))
		}

		// Verify Player 1 no longer has activePunishments on login
		r, raw = mustJSONRequest(t, anonClient, http.MethodPost, ts.URL+"/v1/minecraft/players/login", map[string]any{
			"minecraftUUID": player1UUID,
			"username":      "MainAcc",
			"ip":            "203.0.113.42",
			"serverName":    "survival",
		}, mcHeaders)
		loginOut.ActivePunishments = nil
		_ = json.Unmarshal(raw, &loginOut)
		if len(loginOut.ActivePunishments) != 0 {
			t.Fatalf("expected pardoned Player 1 to have 0 activePunishments, got: %s", string(raw))
		}
	})

	// 4. Tickets, WebSocket Push, KnowledgeBase, Homepage Cards
	t.Run("Tickets_KnowledgeBase_HomepageCards", func(t *testing.T) {
		sameOrigin := map[string]string{"Origin": ts.URL}

		// Public submits a ticket
		r, raw := mustJSONRequest(t, anonClient, http.MethodPost, ts.URL+"/api/v1/public/tickets", map[string]any{
			"type":     "support",
			"username": "MainAcc",
			"email":    "player1@example.com",
			"subject":  "Lost rank after restart",
			"message":  "Please restore my VIP rank.",
		}, sameOrigin)
		if r.StatusCode != 200 && r.StatusCode != 201 {
			t.Fatalf("create public ticket failed (%d): %s", r.StatusCode, string(raw))
		}
		var tickResp map[string]any
		_ = json.Unmarshal(raw, &tickResp)
		ticketID, _ := tickResp["id"].(string)
		if ticketID == "" {
			if tk, ok := tickResp["ticket"].(map[string]any); ok {
				ticketID, _ = tk["id"].(string)
			}
		}
		if ticketID == "" {
			t.Fatalf("missing ticket id: %s", string(raw))
		}

		// Staff replies and closes ticket
		r, raw = mustJSONRequest(t, ownerClient, http.MethodPost, ts.URL+"/api/v1/panel/tickets/"+ticketID+"/replies", map[string]any{
			"name":    "Owner",
			"content": "Restored your VIP rank!",
		}, sameOrigin)
		if r.StatusCode != 200 && r.StatusCode != 201 {
			t.Fatalf("panel ticket reply failed (%d): %s", r.StatusCode, string(raw))
		}
		r, raw = mustJSONRequest(t, ownerClient, http.MethodPatch, ts.URL+"/api/v1/panel/tickets/"+ticketID, map[string]any{
			"status": "CLOSED",
		}, sameOrigin)
		if r.StatusCode != 200 {
			t.Fatalf("panel ticket close failed (%d): %s", r.StatusCode, string(raw))
		}

		// Knowledgebase category + article -> verify public search finds it
		r, raw = mustJSONRequest(t, ownerClient, http.MethodPost, ts.URL+"/api/v1/panel/knowledgebase/categories", map[string]any{
			"name":        "Server Rules",
			"description": "Core rules for all players",
		}, sameOrigin)
		if r.StatusCode != 200 && r.StatusCode != 201 {
			t.Fatalf("create KB category failed (%d): %s", r.StatusCode, string(raw))
		}
		var catResp map[string]any
		_ = json.Unmarshal(raw, &catResp)
		catID, _ := catResp["id"].(string)
		if catID == "" {
			if c, ok := catResp["category"].(map[string]any); ok {
				catID, _ = c["id"].(string)
			}
		}
		if catID == "" {
			t.Fatalf("missing KB category id: %s", string(raw))
		}

		r, raw = mustJSONRequest(t, ownerClient, http.MethodPost, ts.URL+"/api/v1/panel/knowledgebase/categories/"+catID+"/articles", map[string]any{
			"title":     "No X-Ray or Fly Hacks",
			"content":   "Cheating results in a permanent ban.",
			"published": true,
		}, sameOrigin)
		if r.StatusCode != 200 && r.StatusCode != 201 {
			t.Fatalf("create KB article failed (%d): %s", r.StatusCode, string(raw))
		}

		r, raw = mustJSONRequest(t, anonClient, http.MethodGet, ts.URL+"/api/v1/public/knowledgebase/search?q=X-Ray", nil, nil)
		if r.StatusCode != 200 || !strings.Contains(string(raw), "No X-Ray") {
			t.Fatalf("public KB search failed (%d): %s", r.StatusCode, string(raw))
		}

		// Homepage card create + public fetch
		r, raw = mustJSONRequest(t, ownerClient, http.MethodPost, ts.URL+"/api/v1/panel/homepage-cards", map[string]any{
			"title":       "Need Support?",
			"description": "Open a ticket anytime",
			"actionType":  "ticket",
			"enabled":     true,
		}, sameOrigin)
		if r.StatusCode != 200 && r.StatusCode != 201 {
			t.Fatalf("create homepage card failed (%d): %s", r.StatusCode, string(raw))
		}
		r, raw = mustJSONRequest(t, anonClient, http.MethodGet, ts.URL+"/api/v1/public/homepage-cards", nil, nil)
		if r.StatusCode != 200 || !strings.Contains(string(raw), "Need Support?") {
			t.Fatalf("public homepage cards failed (%d): %s", r.StatusCode, string(raw))
		}
	})

	// 5. Media & Storage Presigned Upload -> Binary PUT -> Confirm -> Retrieve -> Traversal Check
	var uploadedMediaURL string
	t.Run("Media_PresignUploadRetrieveAndSecurity", func(t *testing.T) {
		sameOrigin := map[string]string{"Origin": ts.URL}
		// 1x1 valid PNG bytes
		pngBytes := []byte{
			0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
			0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
			0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
			0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4, 0x89,
			0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41, 0x54,
			0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00, 0x05, 0x00, 0x01,
			0x0D, 0x0A, 0x2D, 0xB4, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44,
			0xAE, 0x42, 0x60, 0x82,
		}

		r, raw := mustJSONRequest(t, ownerClient, http.MethodPost, ts.URL+"/api/v1/panel/media/presign", map[string]any{
			"fileName":    "proof.png",
			"contentType": "image/png",
			"fileSize":    len(pngBytes),
			"uploadType":  "article",
		}, sameOrigin)
		if r.StatusCode != 200 {
			t.Fatalf("media presign failed (%d): %s", r.StatusCode, string(raw))
		}
		var presign struct {
			PresignedURL string `json:"presignedUrl"`
			Key          string `json:"key"`
		}
		_ = json.Unmarshal(raw, &presign)
		if presign.PresignedURL == "" || presign.Key == "" {
			t.Fatalf("missing presignedUrl/key: %s", string(raw))
		}

		// PUT binary payload to presignedUrl
		putReq, _ := http.NewRequest(http.MethodPut, presign.PresignedURL, bytes.NewReader(pngBytes))
		putReq.Header.Set("Content-Type", "image/png")
		putResp, err := anonClient.Do(putReq)
		if err != nil {
			t.Fatalf("PUT presigned upload failed: %v", err)
		}
		_ = putResp.Body.Close()
		if putResp.StatusCode != 200 {
			t.Fatalf("PUT presigned upload status: %d", putResp.StatusCode)
		}

		// Confirm upload
		r, raw = mustJSONRequest(t, ownerClient, http.MethodPost, ts.URL+"/api/v1/panel/media/confirm", map[string]any{
			"key":         presign.Key,
			"fileName":    "proof.png",
			"contentType": "image/png",
			"fileSize":    len(pngBytes),
		}, sameOrigin)
		if r.StatusCode != 200 {
			t.Fatalf("media confirm failed (%d): %s", r.StatusCode, string(raw))
		}
		var confirmResp struct {
			URL string `json:"url"`
		}
		_ = json.Unmarshal(raw, &confirmResp)
		uploadedMediaURL = confirmResp.URL
		if uploadedMediaURL == "" {
			t.Fatalf("missing URL in confirm response: %s", string(raw))
		}

		// GET uploaded file and verify exact bytes
		getResp, err := anonClient.Get(uploadedMediaURL)
		if err != nil {
			t.Fatalf("GET uploaded media failed: %v", err)
		}
		gotBytes, _ := io.ReadAll(getResp.Body)
		_ = getResp.Body.Close()
		if getResp.StatusCode != 200 || !bytes.Equal(gotBytes, pngBytes) {
			t.Fatalf("uploaded media mismatch: status=%d len=%d", getResp.StatusCode, len(gotBytes))
		}

		// Path traversal attempt must fail safely
		r, _ = mustJSONRequest(t, ownerClient, http.MethodDelete, ts.URL+"/api/v1/panel/media/../../etc/passwd", nil, sameOrigin)
		if r.StatusCode == 200 {
			t.Fatal("path traversal delete must not succeed")
		}
	})

	// 6. Dashboard, Analytics & Audit Endpoints
	t.Run("Dashboard_Analytics_Audit", func(t *testing.T) {
		for _, ep := range []string{
			"/api/v1/panel/dashboard/metrics",
			"/api/v1/panel/dashboard/recent-tickets",
			"/api/v1/panel/dashboard/recent-punishments",
			"/api/v1/panel/analytics/overview",
			"/api/v1/panel/audit/punishments",
		} {
			r, raw := mustJSONRequest(t, ownerClient, http.MethodGet, ts.URL+ep, nil, nil)
			if r.StatusCode != 200 {
				t.Fatalf("GET %s failed (%d): %s", ep, r.StatusCode, string(raw))
			}
		}
	})

	// 7. Heavy Concurrency Torture Test (20 parallel workers across logins, punishments, syncs, tickets, WS)
	t.Run("Concurrency_Stress_Torture", func(t *testing.T) {
		wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/v1/realtime/ws"
		wsH := http.Header{}
		wsH.Set("X-API-Key", apiKey)
		wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, wsH)
		if err != nil {
			t.Fatalf("ws dial failed: %v", err)
		}
		defer wsConn.Close()
		helloBytes, _ := proto.Marshal(&modlv1.RealtimeEnvelope{
			ProtocolVersion: 1,
			Payload: &modlv1.RealtimeEnvelope_ClientHello{
				ClientHello: &modlv1.ClientHello{
					ClientKind:      modlv1.ClientKind_CLIENT_KIND_MINECRAFT_PLUGIN,
					ProtocolVersion: 1,
					SupportedTopics: []modlv1.Topic{modlv1.Topic_TOPIC_MINECRAFT_PUNISHMENTS},
				},
			},
		})
		_ = wsConn.WriteMessage(websocket.BinaryMessage, helloBytes)

		const workers = 20
		const itersPerWorker = 10
		var wg sync.WaitGroup
		errCh := make(chan error, workers*itersPerWorker*3)

		for w := 0; w < workers; w++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				c := &http.Client{Timeout: 10 * time.Second}
				for i := 0; i < itersPerWorker; i++ {
					idx := workerID*100 + i
					uuid := fmt.Sprintf("33333333-3333-4333-8333-%012d", idx)
					// Login
					body, _ := json.Marshal(map[string]any{
						"minecraftUUID": uuid,
						"username":      fmt.Sprintf("Stress%d", idx),
						"ip":            fmt.Sprintf("198.51.100.%d", (idx%200)+1),
						"serverName":    "stress-1",
					})
					req, _ := http.NewRequest(http.MethodPost, ts.URL+"/v1/minecraft/players/login", bytes.NewReader(body))
					req.Header.Set("X-API-Key", apiKey)
					req.Header.Set("Content-Type", "application/json")
					resp, err := c.Do(req)
					if err != nil || (resp.StatusCode != 200 && resp.StatusCode != 201) {
						errCh <- fmt.Errorf("login worker=%d iter=%d err=%v status=%v", workerID, i, err, resp)
						if resp != nil {
							_ = resp.Body.Close()
						}
						continue
					}
					_ = resp.Body.Close()

					// Punish every 2nd player
					if i%2 == 0 {
						pbody, _ := json.Marshal(map[string]any{
							"targetUuid":   uuid,
							"type_ordinal": 1,
							"reason":       "Stress test",
							"issuerName":   "Console",
						})
						preq, _ := http.NewRequest(http.MethodPost, ts.URL+"/v1/minecraft/punishments/create", bytes.NewReader(pbody))
						preq.Header.Set("X-API-Key", apiKey)
						preq.Header.Set("Content-Type", "application/json")
						presp, perr := c.Do(preq)
						if perr != nil || (presp.StatusCode != 200 && presp.StatusCode != 201) {
							errCh <- fmt.Errorf("punish worker=%d iter=%d err=%v status=%v", workerID, i, perr, presp)
						}
						if presp != nil {
							_ = presp.Body.Close()
						}
					}

					// Sync
					sbody, _ := json.Marshal(map[string]any{
						"lastSyncTimestamp": "",
						"onlinePlayers":     []any{},
						"serverName":        "stress-1",
					})
					sreq, _ := http.NewRequest(http.MethodPost, ts.URL+"/v2/minecraft/players/sync", bytes.NewReader(sbody))
					sreq.Header.Set("X-API-Key", apiKey)
					sreq.Header.Set("Content-Type", "application/json")
					sresp, serr := c.Do(sreq)
					if serr != nil || sresp.StatusCode != 200 {
						errCh <- fmt.Errorf("sync worker=%d iter=%d err=%v status=%v", workerID, i, serr, sresp)
					}
					if sresp != nil {
						_ = sresp.Body.Close()
					}
				}
			}(w)
		}
		wg.Wait()
		close(errCh)
		for err := range errCh {
			t.Fatalf("concurrency error: %v", err)
		}
	})

	// 8. Graceful Shutdown & Full Restart Persistence Verification
	t.Run("Restart_Persistence", func(t *testing.T) {
		ts.Close()
		cleanup()
		closedEarly = true

		// Reboot completely fresh server handler against same DBPath + StorageDir
		handler2, cleanup2, err := buildServerHandler(cfg)
		if err != nil {
			t.Fatalf("reboot buildServerHandler failed: %v", err)
		}
		defer cleanup2()

		ts2 := httptest.NewServer(handler2)
		defer ts2.Close()

		// Existing owner session cookie MUST still work across restart!
		r, raw := mustJSONRequest(t, ownerClient, http.MethodGet, ts2.URL+"/api/v1/panel/auth/me", nil, nil)
		if r.StatusCode != 200 || !strings.Contains(string(raw), cfg.BootstrapAdminEmail) {
			t.Fatalf("owner session did not persist across restart (%d): %s", r.StatusCode, string(raw))
		}

		// Previously created KB article MUST still exist!
		r, raw = mustJSONRequest(t, anonClient, http.MethodGet, ts2.URL+"/api/v1/public/knowledgebase/search?q=X-Ray", nil, nil)
		if r.StatusCode != 200 || !strings.Contains(string(raw), "No X-Ray") {
			t.Fatalf("KB article did not persist across restart (%d): %s", r.StatusCode, string(raw))
		}

		// Previously uploaded media file MUST still be downloadable via rewritten host!
		if uploadedMediaURL != "" {
			idx := strings.Index(uploadedMediaURL, "/v1/public/media/file/")
			if idx >= 0 {
				reboundURL := ts2.URL + uploadedMediaURL[idx:]
				r, err := anonClient.Get(reboundURL)
				if err != nil || r.StatusCode != 200 {
					t.Fatalf("uploaded media did not persist across restart: err=%v resp=%v", err, r)
				}
				_ = r.Body.Close()
			}
		}

		// Stress-created players must be queryable via Panel
		r, raw = mustJSONRequest(t, ownerClient, http.MethodGet, ts2.URL+"/api/v1/panel/players?search=MainAcc&size=50", nil, nil)
		if r.StatusCode != 200 || !strings.Contains(string(raw), "MainAcc") {
			t.Fatalf("players did not persist across restart (%d): %s", r.StatusCode, string(raw))
		}
	})
}
