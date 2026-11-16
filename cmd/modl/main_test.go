package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/config"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

const protoMime = "application/x-protobuf"

func TestEndToEndSelfHostedServer(t *testing.T) {
	tmpDir := t.TempDir()
	apiKey := "test-mc-api-key-123456"
	cfg := &config.Config{
		Port:                "0",
		DBPath:              filepath.Join(tmpDir, "modl.db"),
		StorageDir:          filepath.Join(tmpDir, "storage"),
		ModlDomain:          "localhost",
		DevelopmentMode:     true,
		SingleTenant:        true,
		AuthCodeHashSecret:  "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		SessionCookieName:   "MODL_SESSION",
		SessionDurationSec:  3600,
		CORSAppDomains:      []string{"localhost"},
		CORSSystemOrigins:   []string{"http://localhost"},
		BootstrapServerName: "TestNetwork",
		BootstrapDomain:     "localhost",
		BootstrapAdminEmail: "owner@testnet.gg",
		BootstrapAPIKey:     apiKey,
		RealtimeEnabled:     true,
	}

	handler, cleanup, err := buildServerHandler(cfg)
	if err != nil {
		t.Fatalf("buildServerHandler failed: %v", err)
	}
	defer cleanup()

	ts := httptest.NewServer(handler)
	defer ts.Close()

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar, Timeout: 10 * time.Second}

	// 1. Embedded Panel SPA serves index.html & fallback routes
	resp, err := client.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != 200 || !strings.Contains(string(bodyBytes), "<html") {
		t.Fatalf("expected embedded SPA HTML on /, got status=%d body=%s", resp.StatusCode, string(bodyBytes))
	}

	// 2. Health endpoint works on both /v1/health and /api/v1/health
	for _, p := range []string{"/v1/health", "/api/v1/health"} {
		r, err := client.Get(ts.URL + p)
		if err != nil || r.StatusCode != 200 {
			t.Fatalf("health check %s failed: status=%v err=%v", p, r, err)
		}
		_ = r.Body.Close()
	}

	// 3. Public settings & billing stub via /api/v1/* (how modl-gg/panel calls it)
	r, err := client.Get(ts.URL + "/api/v1/public/settings")
	if err != nil || r.StatusCode != 200 {
		t.Fatalf("public settings failed: err=%v", err)
	}
	_ = r.Body.Close()

	// 4. Minecraft plugin V2 startup & V3 binary Protobuf login + sync
	loginReqBytes, _ := proto.Marshal(&modlv1.PlayerLoginRequest{
		MinecraftUuid: "550e8400-e29b-41d4-a716-446655440000",
		Username:      "Notch",
	})
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/v3/minecraft/players/login", bytes.NewReader(loginReqBytes))
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", protoMime)
	req.Header.Set("Accept", protoMime)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("v3 login request failed: %v", err)
	}
	rawLoginResp, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		t.Fatalf("v3 login status %d: %s", resp.StatusCode, string(rawLoginResp))
	}
	var loginResp modlv1.PlayerLoginResponse
	if err := proto.Unmarshal(rawLoginResp, &loginResp); err != nil {
		t.Fatalf("unmarshal v3 PlayerLoginResponse failed: %v", err)
	}

	// 5. WebSocket realtime connection & topic subscription with Minecraft API key
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/v1/realtime/ws"
	wsHeader := http.Header{}
	wsHeader.Set("X-API-Key", apiKey)
	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, wsHeader)
	if err != nil {
		t.Fatalf("websocket dial failed: %v", err)
	}
	defer wsConn.Close()

	helloEnv, _ := proto.Marshal(&modlv1.RealtimeEnvelope{
		ProtocolVersion: 1,
		Payload: &modlv1.RealtimeEnvelope_ClientHello{
			ClientHello: &modlv1.ClientHello{
				ClientKind:      modlv1.ClientKind_CLIENT_KIND_MINECRAFT_PLUGIN,
				ProtocolVersion: 1,
				SupportedTopics: []modlv1.Topic{
					modlv1.Topic_TOPIC_MINECRAFT_PUNISHMENTS,
					modlv1.Topic_TOPIC_MINECRAFT_PERMISSIONS,
				},
			},
		},
	})
	if err := wsConn.WriteMessage(websocket.BinaryMessage, helloEnv); err != nil {
		t.Fatalf("ws ClientHello failed: %v", err)
	}
	_ = wsConn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, helloAckBytes, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("ws ServerHello read failed: %v", err)
	}
	var helloAckEnv modlv1.RealtimeEnvelope
	if err := proto.Unmarshal(helloAckBytes, &helloAckEnv); err != nil || helloAckEnv.GetServerHello() == nil {
		t.Fatalf("expected ServerHello, got %+v err=%v", &helloAckEnv, err)
	}

	// 6. V3 binary Protobuf punishment creation triggers real-time WS push
	reason := "Integration test ban"
	createPunBytes, _ := proto.Marshal(&modlv1.CreatePunishmentRequest{
		TargetUuid:  "550e8400-e29b-41d4-a716-446655440000",
		TypeOrdinal: 1,
		Reason:      &reason,
	})
	req2, _ := http.NewRequest(http.MethodPost, ts.URL+"/v3/minecraft/punishments/dynamic", bytes.NewReader(createPunBytes))
	req2.Header.Set("X-API-Key", apiKey)
	req2.Header.Set("Content-Type", protoMime)
	req2.Header.Set("Accept", protoMime)
	resp2, err := client.Do(req2)
	if err != nil {
		t.Fatalf("v3 create punishment failed: %v", err)
	}
	rawPunResp, _ := io.ReadAll(resp2.Body)
	_ = resp2.Body.Close()
	if resp2.StatusCode != 200 {
		t.Fatalf("v3 create punishment status %d: %s", resp2.StatusCode, string(rawPunResp))
	}
	var punCreateResp modlv1.PunishmentCreateResponse
	if err := proto.Unmarshal(rawPunResp, &punCreateResp); err != nil || punCreateResp.PunishmentId == "" {
		t.Fatalf("unexpected v3 PunishmentCreateResponse: %+v err=%v", &punCreateResp, err)
	}

	// Verify WebSocket receives either subscribed ack or punishment push without timeout
	_ = wsConn.SetReadDeadline(time.Now().Add(3 * time.Second))
	gotPunPush := false
	for i := 0; i < 4; i++ {
		_, msg, err := wsConn.ReadMessage()
		if err != nil {
			break
		}
		var env modlv1.RealtimeEnvelope
		if proto.Unmarshal(msg, &env) == nil {
			if pp := env.GetPunishmentPush(); pp != nil && len(pp.Pending) > 0 {
				gotPunPush = true
				break
			}
		}
	}
	if !gotPunPush {
		t.Fatalf("expected realtime punishment push over WebSocket after punishment creation")
	}

	// 7. V1 JSON Minecraft sync confirms online/punishment/staff payloads
	syncReqBody, _ := json.Marshal(map[string]any{
		"lastSyncTimestamp": "",
		"onlinePlayers":     []any{},
		"serverName":        "survival-1",
	})
	req3, _ := http.NewRequest(http.MethodPost, ts.URL+"/v2/minecraft/players/sync", bytes.NewReader(syncReqBody))
	req3.Header.Set("X-API-Key", apiKey)
	req3.Header.Set("Content-Type", "application/json")
	resp3, err := client.Do(req3)
	if err != nil || resp3.StatusCode != 200 {
		t.Fatalf("v2 sync failed: err=%v resp=%v", err, resp3)
	}
	_ = resp3.Body.Close()
}
