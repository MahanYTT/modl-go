package media_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/config"
	"modl-go/internal/db"
	"modl-go/internal/httpx"
	"modl-go/internal/models"
	"modl-go/internal/modules/media"
	"modl-go/internal/storage"

	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/encoding/protojson"
)

type stubPerms struct{}

func (s *stubPerms) IsSuperAdmin(srv *models.Server, email string) bool {
	return email == "super@example.com"
}
func (s *stubPerms) IsAuthorizedEmail(ctx context.Context, srv *models.Server, email string) bool {
	return true
}
func (s *stubPerms) PanelRoleID(ctx context.Context, srv *models.Server, email string) string {
	return email
}
func (s *stubPerms) HasPermission(ctx context.Context, srv *models.Server, roleID, perm string) bool {
	switch roleID {
	case "content@example.com":
		return perm == "admin.settings.view.content" || perm == "admin.settings.modify.content"
	case "storage@example.com":
		return perm == "admin.settings.view.storage" || perm == "admin.settings.modify.storage"
	case "viewonly@example.com":
		return perm == "admin.settings.view.content" || perm == "admin.settings.view.storage"
	}
	return false
}
func (s *stubPerms) HasAnyWithPrefix(ctx context.Context, srv *models.Server, roleID, prefix string) bool {
	return false
}

func setupMedia(t *testing.T) (chi.Router, *media.Service, *storage.LocalStore, *sql.DB) {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	tmpDir := t.TempDir()
	store, err := storage.NewLocalStore(database, tmpDir)
	if err != nil {
		t.Fatalf("new local store: %v", err)
	}
	cfg := &config.Config{AuthCodeHashSecret: "test-secret-key"}
	svc := media.NewService(database, store, cfg)
	h := media.NewHandler(svc)
	perms := &stubPerms{}

	r := chi.NewRouter()
	r.Mount("/v1/panel/media", h.PanelMediaRouter(perms))
	r.Mount("/v1/public/media", h.PublicMediaRouter())
	r.Mount("/v1/panel/storage", h.PanelStorageRouter(perms))
	r.Mount("/v1/public/evidence-upload", h.EvidenceUploadRouter())
	return r, svc, store, database
}

func doJSON(r http.Handler, method, path string, body any, srv *models.Server, email string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		switch v := body.(type) {
		case string:
			buf.WriteString(v)
		default:
			_ = json.NewEncoder(&buf).Encode(v)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	ctx := req.Context()
	if srv != nil {
		ctx = httpx.WithServer(ctx, srv)
	}
	if email != "" {
		ctx = httpx.WithSession(ctx, &models.AuthSession{ID: "s1", ServerID: srv.ID, Email: email})
	}
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

func doPutRaw(r http.Handler, targetURL, contentType string, data []byte) *httptest.ResponseRecorder {
	u, _ := url.Parse(targetURL)
	pathAndQuery := u.RequestURI()
	req := httptest.NewRequest(http.MethodPut, pathAndQuery, bytes.NewReader(data))
	req.Header.Set("Content-Type", contentType)
	req.ContentLength = int64(len(data))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

func doMultipart(r http.Handler, path string, fields map[string]string, fileName, contentType string, data []byte, srv *models.Server, email string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for k, v := range fields {
		_ = mw.WriteField(k, v)
	}
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="`+fileName+`"`)
	h.Set("Content-Type", contentType)
	part, _ := mw.CreatePart(h)
	_, _ = part.Write(data)
	_ = mw.Close()

	req := httptest.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	ctx := req.Context()
	if srv != nil {
		ctx = httpx.WithServer(ctx, srv)
	}
	if email != "" {
		ctx = httpx.WithSession(ctx, &models.AuthSession{ID: "s1", ServerID: srv.ID, Email: email})
	}
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

func TestPanelMediaPresignPutConfirmServeDeleteAndStorage(t *testing.T) {
	r, _, store, _ := setupMedia(t)
	srv := &models.Server{ID: "srv-1", ServerName: "Server 1"}

	// Config check
	rr := doJSON(r, http.MethodGet, "/v1/panel/media/config", nil, srv, "viewonly@example.com")
	if rr.Code != http.StatusOK {
		t.Fatalf("config failed: %d", rr.Code)
	}
	var cfgResp modlv1.MediaConfigResponse
	_ = protojson.Unmarshal(rr.Body.Bytes(), &cfgResp)
	if !cfgResp.BackblazeConfigured {
		t.Fatal("expected backblazeConfigured=true in local storage mode")
	}

	// Dangerous extension blocked
	rr = doJSON(r, http.MethodPost, "/v1/panel/media/presign", map[string]any{
		"uploadType": "article", "fileName": "bad.exe", "fileSize": 100, "contentType": "image/png",
	}, srv, "content@example.com")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for .exe, got %d", rr.Code)
	}

	// Presign -> PUT -> Confirm -> GET /v1/public/media/file/{id}
	payload := []byte("fake-png-bytes-hello")
	rr = doJSON(r, http.MethodPost, "/v1/panel/media/presign", map[string]any{
		"uploadType": "article", "fileName": "banner.png", "fileSize": len(payload), "contentType": "image/png",
	}, srv, "content@example.com")
	if rr.Code != http.StatusOK {
		t.Fatalf("presign failed: %d %s", rr.Code, rr.Body.String())
	}
	var presign modlv1.PresignUploadResponse
	_ = protojson.Unmarshal(rr.Body.Bytes(), &presign)
	if presign.Method != http.MethodPut || !strings.Contains(presign.PresignedUrl, "/v1/public/media/upload") {
		t.Fatalf("unexpected presign: %+v", &presign)
	}

	// Tampered signature fails
	if putBad := doPutRaw(r, presign.PresignedUrl+"tamper", "image/png", payload); putBad.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for tampered sig, got %d", putBad.Code)
	}

	// Valid PUT
	if putOK := doPutRaw(r, presign.PresignedUrl, "image/png", payload); putOK.Code != http.StatusOK {
		t.Fatalf("presigned PUT failed: %d %s", putOK.Code, putOK.Body.String())
	}

	// Confirm
	rr = doJSON(r, http.MethodPost, "/v1/panel/media/confirm", map[string]any{"key": presign.Key}, srv, "content@example.com")
	if rr.Code != http.StatusOK {
		t.Fatalf("confirm failed: %d %s", rr.Code, rr.Body.String())
	}
	var uploaded modlv1.UploadResponse
	_ = protojson.Unmarshal(rr.Body.Bytes(), &uploaded)
	if !strings.Contains(uploaded.Url, "/v1/public/media/file/") || uploaded.Size != int64(len(payload)) {
		t.Fatalf("unexpected confirm response: %+v", &uploaded)
	}

	// Serve file directly
	fileURL, _ := url.Parse(uploaded.Url)
	rr = doJSON(r, http.MethodGet, fileURL.Path, nil, srv, "")
	if rr.Code != http.StatusOK || !bytes.Equal(rr.Body.Bytes(), payload) {
		t.Fatalf("serve file failed: %d %q", rr.Code, rr.Body.String())
	}

	// Direct multipart upload
	rr = doMultipart(r, "/v1/panel/media/upload", map[string]string{"uploadType": "article"}, "note.txt", "text/plain", []byte("hello world"), srv, "content@example.com")
	if rr.Code != http.StatusOK {
		t.Fatalf("direct upload failed: %d %s", rr.Code, rr.Body.String())
	}
	var directResp modlv1.UploadResponse
	_ = protojson.Unmarshal(rr.Body.Bytes(), &directResp)

	// Storage quota, files, download URL
	rr = doJSON(r, http.MethodGet, "/v1/panel/storage/quota", nil, srv, "storage@example.com")
	if rr.Code != http.StatusOK {
		t.Fatalf("quota failed: %d", rr.Code)
	}
	var quota modlv1.StorageQuotaResponse
	_ = protojson.Unmarshal(rr.Body.Bytes(), &quota)
	if quota.UsedBytes != int64(len(payload)+len("hello world")) {
		t.Fatalf("unexpected usedBytes: %d", quota.UsedBytes)
	}

	rr = doJSON(r, http.MethodGet, "/v1/panel/storage/files?prefix=article/", nil, srv, "storage@example.com")
	if rr.Code != http.StatusOK {
		t.Fatalf("list files failed: %d", rr.Code)
	}
	var filesResp modlv1.StorageFilesResponse
	_ = protojson.Unmarshal(rr.Body.Bytes(), &filesResp)
	if len(filesResp.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(filesResp.Files))
	}

	rr = doJSON(r, http.MethodGet, "/v1/panel/storage/download/"+presign.Key, nil, srv, "storage@example.com")
	if rr.Code != http.StatusOK {
		t.Fatalf("download url failed: %d %s", rr.Code, rr.Body.String())
	}

	// Sync requires SUPER_ADMIN
	if rr := doJSON(r, http.MethodPost, "/v1/panel/storage/sync", nil, srv, "storage@example.com"); rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-superadmin sync, got %d", rr.Code)
	}
	// Write an untracked file on disk and sync as superadmin
	orphanKey := srv.ID + "/logs/orphan.txt"
	orphanPath := store.DiskPath(orphanKey)
	_ = os.MkdirAll(filepath.Dir(orphanPath), 0o755)
	_ = os.WriteFile(orphanPath, []byte("abc"), 0o644)
	rr = doJSON(r, http.MethodPost, "/v1/panel/storage/sync", nil, srv, "super@example.com")
	if rr.Code != http.StatusOK {
		t.Fatalf("sync failed: %d", rr.Code)
	}
	var syncResp modlv1.StorageSyncResponse
	_ = protojson.Unmarshal(rr.Body.Bytes(), &syncResp)
	if syncResp.Synced != 3 {
		t.Fatalf("expected 3 synced files, got %d", syncResp.Synced)
	}

	// Delete via PanelMedia and BulkDelete via PanelStorage
	if rr := doJSON(r, http.MethodDelete, "/v1/panel/media/"+presign.Key, nil, srv, "content@example.com"); rr.Code != http.StatusOK {
		t.Fatalf("delete file failed: %d %s", rr.Code, rr.Body.String())
	}
	rr = doJSON(r, http.MethodPost, "/v1/panel/storage/bulk-delete", map[string]any{"keys": []string{directResp.Key, orphanKey}}, srv, "storage@example.com")
	if rr.Code != http.StatusOK {
		t.Fatalf("bulk delete failed: %d %s", rr.Code, rr.Body.String())
	}
	var bdResp modlv1.StorageBulkDeleteResponse
	_ = protojson.Unmarshal(rr.Body.Bytes(), &bdResp)
	if bdResp.Deleted != 2 {
		t.Fatalf("expected 2 deleted, got %d", bdResp.Deleted)
	}
}

func TestPublicMediaTicketAccessControl(t *testing.T) {
	r, _, _, database := setupMedia(t)
	srv := &models.Server{ID: "srv-1", ServerName: "Server 1"}
	now := time.Now().UnixMilli()

	// Seed ticket with email auth enabled
	_, err := database.Exec(
		`INSERT INTO tickets (id, server_id, type, subject, status, email_auth_enabled, hidden, created_at, updated_at)
		 VALUES ('t-1', 'srv-1', 'REPORT', 'Cheater', 'OPEN', 1, 0, ?, ?)`,
		now, now,
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = database.Exec(
		`INSERT INTO ticket_verifications (id, server_id, ticket_id, token, expires_at)
		 VALUES ('v-1', 'srv-1', 't-1', 'valid-ticket-tok', ?)`,
		now+600000,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Disallowed type for public
	rr := doJSON(r, http.MethodPost, "/v1/public/media/presign", map[string]any{
		"uploadType": "article", "fileName": "a.png", "fileSize": 10, "contentType": "image/png", "entityId": "new",
	}, srv, "")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for article on public, got %d", rr.Code)
	}

	// Missing email verification token for t-1 -> 403
	rr = doJSON(r, http.MethodPost, "/v1/public/media/presign", map[string]any{
		"uploadType": "ticket", "fileName": "a.png", "fileSize": 10, "contentType": "image/png", "entityId": "t-1",
	}, srv, "")
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without token, got %d", rr.Code)
	}

	// Wrong type (appeal on REPORT ticket) -> 403
	rr = doJSON(r, http.MethodPost, "/v1/public/media/presign", map[string]any{
		"uploadType": "appeal", "fileName": "a.png", "fileSize": 10, "contentType": "image/png", "entityId": "t-1", "accessToken": "valid-ticket-tok",
	}, srv, "")
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for appeal on REPORT ticket, got %d", rr.Code)
	}

	// Valid ticket upload with token
	rr = doJSON(r, http.MethodPost, "/v1/public/media/presign", map[string]any{
		"uploadType": "ticket", "fileName": "proof.png", "fileSize": 5, "contentType": "image/png", "entityId": "t-1", "accessToken": "valid-ticket-tok",
	}, srv, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d %s", rr.Code, rr.Body.String())
	}
	var presign modlv1.PresignUploadResponse
	_ = protojson.Unmarshal(rr.Body.Bytes(), &presign)
	_ = doPutRaw(r, presign.PresignedUrl, "image/png", []byte("12345"))
	rr = doJSON(r, http.MethodPost, "/v1/public/media/confirm", map[string]any{
		"key": presign.Key, "accessToken": "valid-ticket-tok",
	}, srv, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("confirm public upload failed: %d %s", rr.Code, rr.Body.String())
	}
}

func TestEvidenceUploadFullFlow(t *testing.T) {
	r, svc, _, database := setupMedia(t)
	now := time.Now().UnixMilli()

	// Seed player + punishment
	_, err := database.Exec(
		`INSERT INTO players (id, server_id, minecraft_uuid, current_username, first_seen, last_seen)
		 VALUES ('p1', 'srv-1', '12345678-1234-1234-1234-1234567890ab', 'Notch', ?, ?)`,
		now, now,
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = database.Exec(
		`INSERT INTO punishments (id, server_id, player_id, player_uuid, player_username, type, reason, issued_at, updated_at)
		 VALUES ('pun-1', 'srv-1', 'p1', '12345678-1234-1234-1234-1234567890ab', 'Notch', 'BAN', 'Flying', ?, ?)`,
		now, now,
	)
	if err != nil {
		t.Fatal(err)
	}

	token, err := svc.CreateEvidenceUploadToken(context.Background(), "srv-1", "pun-1", "123456781234123412341234567890ab", "ModeratorBob")
	if err != nil {
		t.Fatal(err)
	}

	// Validate token
	rr := doJSON(r, http.MethodGet, "/v1/public/evidence-upload/"+token, nil, nil, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("validate token failed: %d %s", rr.Code, rr.Body.String())
	}
	var valResp map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &valResp)
	if valResp["playerName"] != "Notch" || valResp["punishmentId"] != "pun-1" || valResp["issuerName"] != "ModeratorBob" {
		t.Fatalf("unexpected token info: %+v", valResp)
	}

	// Presign -> PUT -> Confirm -> Serve -> Submit
	rr = doJSON(r, http.MethodPost, "/v1/public/evidence-upload/"+token+"/presign", map[string]any{
		"fileName": "clip.png", "contentType": "image/png", "fileSize": 9,
	}, nil, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("evidence presign failed: %d %s", rr.Code, rr.Body.String())
	}
	var presignResp struct {
		Status       int    `json:"status"`
		PresignedUrl string `json:"presignedUrl"`
		Key          string `json:"key"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &presignResp)

	if putRR := doPutRaw(r, presignResp.PresignedUrl, "image/png", []byte("png-bytes")); putRR.Code != http.StatusOK {
		t.Fatalf("evidence PUT failed: %d %s", putRR.Code, putRR.Body.String())
	}

	rr = doJSON(r, http.MethodPost, "/v1/public/evidence-upload/"+token+"/confirm", map[string]any{
		"key": presignResp.Key,
	}, nil, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("evidence confirm failed: %d %s", rr.Code, rr.Body.String())
	}
	var confirmResp struct {
		Status int    `json:"status"`
		URL    string `json:"url"`
		Key    string `json:"key"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &confirmResp)
	if !strings.Contains(confirmResp.URL, "/v1/public/evidence-upload/file/") {
		t.Fatalf("unexpected evidence file URL: %s", confirmResp.URL)
	}

	// Serve evidence file directly
	u, _ := url.Parse(confirmResp.URL)
	getRR := httptest.NewRequest(http.MethodGet, u.Path, nil)
	recRR := httptest.NewRecorder()
	r.ServeHTTP(recRR, getRR)
	if recRR.Code != http.StatusOK {
		t.Fatalf("serve evidence file failed: %d", recRR.Code)
	}
	bodyBytes, _ := io.ReadAll(recRR.Body)
	if string(bodyBytes) != "png-bytes" {
		t.Fatalf("unexpected evidence bytes: %q", string(bodyBytes))
	}

	// Submit evidence
	rr = doJSON(r, http.MethodPost, "/v1/public/evidence-upload/"+token+"/submit", map[string]any{
		"evidence": []map[string]any{
			{"url": confirmResp.URL, "fileName": "clip.png", "fileType": "image/png", "fileSize": 9},
		},
	}, nil, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("evidence submit failed: %d %s", rr.Code, rr.Body.String())
	}

	// Verify DB updated & token invalidated
	var evJSON, notesJSON string
	if err := database.QueryRow(`SELECT evidence_json, notes_json FROM punishments WHERE id = 'pun-1'`).Scan(&evJSON, &notesJSON); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(evJSON, "clip.png") || !strings.Contains(notesJSON, "uploaded 1 evidence file(s)") {
		t.Fatalf("unexpected punishment evidence/notes: %s | %s", evJSON, notesJSON)
	}
	if rr := doJSON(r, http.MethodGet, "/v1/public/evidence-upload/"+token, nil, nil, ""); rr.Code != http.StatusNotFound {
		t.Fatalf("expected invalidated token to return 404, got %d", rr.Code)
	}
}
