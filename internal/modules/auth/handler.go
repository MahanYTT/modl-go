package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/config"
	"modl-go/internal/email"
	"modl-go/internal/httpx"
	"modl-go/internal/models"
	"modl-go/internal/modules/role"

	"github.com/go-chi/chi/v5"
	"github.com/go-webauthn/webauthn/webauthn"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler struct {
	db     *sql.DB
	cfg    *config.Config
	mailer *email.Sender
	roles  *role.Handler
}

func NewHandler(db *sql.DB, cfg *config.Config, mailer *email.Sender, roles *role.Handler) *Handler {
	return &Handler{db: db, cfg: cfg, mailer: mailer, roles: roles}
}

func (h *Handler) RegisterPanelAuthRoutes(r chi.Router) {
	r.Post("/send-email-code", h.sendEmailCode)
	r.Post("/verify-email-code", h.verifyEmailCode)
	r.Post("/logout", h.logout)
	r.Get("/me", h.getCurrentUser)
	r.Patch("/profile", h.updateProfile)
	r.Post("/email/send-code", h.sendEmailChangeCode)
	r.Patch("/email", h.updateEmail)
	r.Get("/sessions", h.getSessions)
	r.Delete("/sessions", h.revokeAllSessions)
	r.Delete("/sessions/{publicId}", h.revokeSession)
	r.Get("/permissions", h.getUserPermissions)

	r.Route("/webauthn", func(r chi.Router) {
		r.Post("/register/options", h.webauthnRegisterOptions)
		r.Post("/register/verify", h.webauthnRegisterVerify)
		r.Get("/credentials", h.webauthnListCredentials)
		r.Patch("/credentials/{id}", h.webauthnRenameCredential)
		r.Delete("/credentials/{id}", h.webauthnDeleteCredential)
		r.Post("/login/start", h.webauthnLoginStart)
		r.Post("/login/options", h.webauthnLoginOptions)
		r.Post("/login/verify", h.webauthnLoginVerify)
	})
}

func (h *Handler) CreateSession(ctx context.Context, serverID, emailAddr, ip, ua string) (*models.AuthSession, error) {
	now := httpx.NowMs()
	sess := &models.AuthSession{
		ID:        httpx.RandomToken(32),
		ServerID:  serverID,
		Email:     strings.ToLower(strings.TrimSpace(emailAddr)),
		CreatedAt: now,
		ExpiresAt: now + h.cfg.SessionDurationSec*1000,
		IPAddress: ip,
		UserAgent: ua,
	}
	_, err := h.db.ExecContext(ctx, `
		INSERT INTO sessions (id, server_id, email, created_at, expires_at, ip_address, user_agent)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		sess.ID, sess.ServerID, sess.Email, sess.CreatedAt, sess.ExpiresAt, sess.IPAddress, sess.UserAgent,
	)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func (h *Handler) FindAndRefresh(ctx context.Context, serverID, token string) (*models.AuthSession, error) {
	now := httpx.NowMs()
	var s models.AuthSession
	err := h.db.QueryRowContext(ctx, `
		SELECT id, server_id, email, created_at, expires_at, ip_address, user_agent
		FROM sessions WHERE id = ? AND server_id = ? AND expires_at > ? LIMIT 1`,
		token, serverID, now,
	).Scan(&s.ID, &s.ServerID, &s.Email, &s.CreatedAt, &s.ExpiresAt, &s.IPAddress, &s.UserAgent)
	if err != nil {
		return nil, err
	}
	newExp := now + h.cfg.SessionDurationSec*1000
	_, _ = h.db.ExecContext(ctx, `UPDATE sessions SET expires_at = ? WHERE id = ?`, newExp, s.ID)
	s.ExpiresAt = newExp
	return &s, nil
}

func (h *Handler) sendEmailCode(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.PanelSendEmailCodeRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	em := strings.ToLower(strings.TrimSpace(req.Email))
	if h.roles.IsAuthorizedEmail(r.Context(), srv, em) {
		code := httpx.GenerateNumericCode(6)
		hash := httpx.HashOneTimeCode(code, h.cfg.AuthCodeHashSecret)
		exp := time.Now().Add(5 * time.Minute).UnixMilli()
		_, _ = h.db.ExecContext(r.Context(), `
			INSERT INTO auth_codes (server_id, email, code_hash, failed_attempts, expires_at)
			VALUES (?, ?, ?, 0, ?)
			ON CONFLICT(server_id, email) DO UPDATE SET code_hash=excluded.code_hash, failed_attempts=0, expires_at=excluded.expires_at`,
			srv.ID, em, hash, exp,
		)
		_ = h.mailer.SendAuthCode(em, srv.ServerName, code)
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PanelAuthResponse{
		Success: true,
		Message: "If that email is registered, a verification code has been sent.",
	})
}

func (h *Handler) verifyEmailCode(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.PanelVerifyEmailCodeRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	em := strings.ToLower(strings.TrimSpace(req.Email))
	var hash string
	var attempts int
	var exp int64
	err := h.db.QueryRowContext(r.Context(),
		`SELECT code_hash, failed_attempts, expires_at FROM auth_codes WHERE server_id = ? AND email = ? LIMIT 1`,
		srv.ID, em).Scan(&hash, &attempts, &exp)
	if err != nil || exp < httpx.NowMs() || attempts >= 5 {
		httpx.WriteProto(w, r, http.StatusUnauthorized, &modlv1.PanelAuthResponse{Success: false, Message: "Invalid or expired verification code"})
		return
	}
	if !httpx.VerifyOneTimeCode(strings.TrimSpace(req.Code), hash, h.cfg.AuthCodeHashSecret) {
		_, _ = h.db.ExecContext(r.Context(), `UPDATE auth_codes SET failed_attempts = failed_attempts + 1 WHERE server_id = ? AND email = ?`, srv.ID, em)
		httpx.WriteProto(w, r, http.StatusUnauthorized, &modlv1.PanelAuthResponse{Success: false, Message: "Invalid verification code"})
		return
	}
	_, _ = h.db.ExecContext(r.Context(), `DELETE FROM auth_codes WHERE server_id = ? AND email = ?`, srv.ID, em)

	sess, err := h.CreateSession(r.Context(), srv.ID, em, httpx.GetClientIP(r), r.UserAgent())
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to create session")
		return
	}
	http.SetCookie(w, httpx.CreateSessionCookie(h.cfg, sess.ID))
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PanelAuthResponse{
		Success: true,
		Message: "Authenticated successfully",
	})
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	for _, c := range r.Cookies() {
		if c.Name == h.cfg.SessionCookieName && c.Value != "" && srv != nil {
			_, _ = h.db.ExecContext(r.Context(), `DELETE FROM sessions WHERE server_id = ? AND id = ?`, srv.ID, c.Value)
		}
	}
	httpx.WriteExpiredSessionCookies(w, h.cfg)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PanelAuthResponse{Success: true, Message: "Logged out successfully"})
}

func (h *Handler) getCurrentUser(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	emailAddr := httpx.GetSessionEmail(r)
	if emailAddr == "" {
		httpx.WriteProto(w, r, http.StatusUnauthorized, &modlv1.PanelAuthResponse{Success: false, Message: "Not authenticated"})
		return
	}
	var id, username, roleID, mcUser, lang, df string
	err := h.db.QueryRowContext(r.Context(), `
		SELECT id, username, role_id, assigned_minecraft_username, language, date_format
		FROM staff WHERE server_id = ? AND LOWER(email) = ? LIMIT 1`,
		srv.ID, strings.ToLower(emailAddr)).Scan(&id, &username, &roleID, &mcUser, &lang, &df)
	if err != nil {
		if h.roles.IsSuperAdmin(srv, emailAddr) {
			id = role.SuperAdminRoleID
			username = strings.Split(emailAddr, "@")[0]
			roleID = "Super Admin"
			lang = "en"
			df = "MM/dd/yyyy"
		} else {
			httpx.WriteProto(w, r, http.StatusUnauthorized, &modlv1.PanelAuthResponse{Success: false, Message: "Not authenticated"})
			return
		}
	} else {
		var rName string
		if h.db.QueryRowContext(r.Context(), `SELECT name FROM staff_roles WHERE server_id = ? AND id = ?`, srv.ID, roleID).Scan(&rName) == nil && rName != "" {
			roleID = rName
		}
		if h.roles.IsSuperAdmin(srv, emailAddr) {
			roleID = "Super Admin"
		}
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PanelProfileResponse{
		Id:                &id,
		Email:             emailAddr,
		Username:          username,
		Role:              roleID,
		MinecraftUsername: mcUser,
		Language:          lang,
		DateFormat:        df,
	})
}

func (h *Handler) updateProfile(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	emailAddr := httpx.GetSessionEmail(r)
	if emailAddr == "" {
		httpx.WriteProto(w, r, http.StatusUnauthorized, &modlv1.PanelAuthResponse{Success: false, Message: "Not authenticated"})
		return
	}
	var req modlv1.PanelUpdateProfileRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if req.Username != nil {
		_, _ = h.db.ExecContext(r.Context(), `UPDATE staff SET username = ?, updated_at = ? WHERE server_id = ? AND LOWER(email) = ?`,
			strings.TrimSpace(*req.Username), httpx.NowMs(), srv.ID, strings.ToLower(emailAddr))
	}
	if req.Language != nil {
		_, _ = h.db.ExecContext(r.Context(), `UPDATE staff SET language = ?, updated_at = ? WHERE server_id = ? AND LOWER(email) = ?`,
			strings.TrimSpace(*req.Language), httpx.NowMs(), srv.ID, strings.ToLower(emailAddr))
	}
	if req.DateFormat != nil {
		_, _ = h.db.ExecContext(r.Context(), `UPDATE staff SET date_format = ?, updated_at = ? WHERE server_id = ? AND LOWER(email) = ?`,
			strings.TrimSpace(*req.DateFormat), httpx.NowMs(), srv.ID, strings.ToLower(emailAddr))
	}
	h.getCurrentUser(w, r)
}

func (h *Handler) sendEmailChangeCode(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	curEmail := httpx.GetSessionEmail(r)
	if curEmail == "" {
		httpx.WriteProto(w, r, http.StatusUnauthorized, &modlv1.PanelAuthResponse{Success: false, Message: "Not authenticated"})
		return
	}
	var req modlv1.PanelUpdateEmailRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	newEmail := strings.ToLower(strings.TrimSpace(req.NewEmail))
	code := httpx.GenerateNumericCode(6)
	hash := httpx.HashOneTimeCode(code, h.cfg.AuthCodeHashSecret)
	exp := time.Now().Add(5 * time.Minute).UnixMilli()
	_, _ = h.db.ExecContext(r.Context(), `
		INSERT INTO auth_codes (server_id, email, code_hash, failed_attempts, expires_at)
		VALUES (?, ?, ?, 0, ?)
		ON CONFLICT(server_id, email) DO UPDATE SET code_hash=excluded.code_hash, failed_attempts=0, expires_at=excluded.expires_at`,
		srv.ID, "change:"+newEmail, hash, exp,
	)
	_ = h.mailer.SendAuthCode(newEmail, srv.ServerName, code)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PanelAuthResponse{Success: true, Message: "Verification code sent"})
}

func (h *Handler) updateEmail(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	curEmail := httpx.GetSessionEmail(r)
	if curEmail == "" {
		httpx.WriteProto(w, r, http.StatusUnauthorized, &modlv1.PanelAuthResponse{Success: false, Message: "Not authenticated"})
		return
	}
	var req modlv1.PanelUpdateEmailWithCodeRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	newEmail := strings.ToLower(strings.TrimSpace(req.NewEmail))
	var hash string
	var exp int64
	err := h.db.QueryRowContext(r.Context(), `SELECT code_hash, expires_at FROM auth_codes WHERE server_id = ? AND email = ?`, srv.ID, "change:"+newEmail).Scan(&hash, &exp)
	if err != nil || exp < httpx.NowMs() || !httpx.VerifyOneTimeCode(strings.TrimSpace(req.Code), hash, h.cfg.AuthCodeHashSecret) {
		httpx.WriteProto(w, r, http.StatusBadRequest, &modlv1.PanelAuthResponse{Success: false, Message: "Invalid or expired code"})
		return
	}
	_, _ = h.db.ExecContext(r.Context(), `DELETE FROM auth_codes WHERE server_id = ? AND email = ?`, srv.ID, "change:"+newEmail)
	_, _ = h.db.ExecContext(r.Context(), `UPDATE staff SET email = ?, updated_at = ? WHERE server_id = ? AND LOWER(email) = ?`, newEmail, httpx.NowMs(), srv.ID, strings.ToLower(curEmail))
	if strings.EqualFold(srv.AdminEmail, curEmail) {
		_, _ = h.db.ExecContext(r.Context(), `UPDATE servers SET admin_email = ?, updated_at = ? WHERE id = ?`, newEmail, httpx.NowMs(), srv.ID)
	}
	_, _ = h.db.ExecContext(r.Context(), `UPDATE sessions SET email = ? WHERE server_id = ? AND LOWER(email) = ?`, newEmail, srv.ID, strings.ToLower(curEmail))
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PanelAuthResponse{Success: true, Message: "Email updated successfully"})
}

func (h *Handler) getSessions(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	curSess := httpx.GetSession(r)
	if curSess == nil {
		httpx.WriteProto(w, r, http.StatusUnauthorized, &modlv1.PanelAuthResponse{Success: false, Message: "Not authenticated"})
		return
	}
	rows, err := h.db.QueryContext(r.Context(), `
		SELECT id, ip_address, user_agent, created_at, expires_at
		FROM sessions WHERE server_id = ? AND LOWER(email) = ? AND expires_at > ? ORDER BY created_at DESC`,
		srv.ID, strings.ToLower(curSess.Email), httpx.NowMs())
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to list sessions")
		return
	}
	defer rows.Close()
	var list []*modlv1.PanelSessionInfoResponse
	for rows.Next() {
		var id, ip, ua string
		var created, exp int64
		if rows.Scan(&id, &ip, &ua, &created, &exp) == nil {
			cStr := time.UnixMilli(created).UTC().Format(time.RFC3339)
			eStr := time.UnixMilli(exp).UTC().Format(time.RFC3339)
			pubID := id
			if len(pubID) > 12 {
				pubID = pubID[:12]
			}
			list = append(list, &modlv1.PanelSessionInfoResponse{
				Id:        pubID,
				IpAddress: ip,
				UserAgent: ua,
				CreatedAt: &cStr,
				ExpiresAt: &eStr,
				IsCurrent: id == curSess.ID,
			})
		}
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PanelSessionsResponse{Sessions: list})
}

func (h *Handler) revokeSession(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	emailAddr := httpx.GetSessionEmail(r)
	if emailAddr == "" {
		httpx.WriteProto(w, r, http.StatusUnauthorized, &modlv1.PanelAuthResponse{Success: false, Message: "Not authenticated"})
		return
	}
	pubID := chi.URLParam(r, "publicId")
	_, _ = h.db.ExecContext(r.Context(), `DELETE FROM sessions WHERE server_id = ? AND LOWER(email) = ? AND id LIKE ?`, srv.ID, strings.ToLower(emailAddr), pubID+"%")
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PanelAuthResponse{Success: true, Message: "Session revoked"})
}

func (h *Handler) revokeAllSessions(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	emailAddr := httpx.GetSessionEmail(r)
	if emailAddr == "" {
		httpx.WriteProto(w, r, http.StatusUnauthorized, &modlv1.PanelAuthResponse{Success: false, Message: "Not authenticated"})
		return
	}
	_, _ = h.db.ExecContext(r.Context(), `DELETE FROM sessions WHERE server_id = ? AND LOWER(email) = ?`, srv.ID, strings.ToLower(emailAddr))
	httpx.WriteExpiredSessionCookies(w, h.cfg)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PanelAuthResponse{Success: true, Message: "All sessions revoked"})
}

func (h *Handler) getUserPermissions(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	emailAddr := httpx.GetSessionEmail(r)
	if emailAddr == "" {
		httpx.WriteProto(w, r, http.StatusUnauthorized, &modlv1.PanelPermissionsResponse{})
		return
	}
	perms := h.roles.GetEffectivePermissions(r.Context(), srv, emailAddr)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PanelPermissionsResponse{Permissions: perms})
}

// --- WebAuthn Passkeys ---

type webAuthnUser struct {
	email string
	creds []webauthn.Credential
}

func (u *webAuthnUser) WebAuthnID() []byte                         { return []byte(u.email) }
func (u *webAuthnUser) WebAuthnName() string                       { return u.email }
func (u *webAuthnUser) WebAuthnDisplayName() string                { return u.email }
func (u *webAuthnUser) WebAuthnCredentials() []webauthn.Credential { return u.creds }

func (h *Handler) newWebAuthn(r *http.Request, srv *models.Server) (*webauthn.WebAuthn, error) {
	rpID := httpx.NormalizeServerDomain(r.Host)
	if origin := r.Header.Get("Origin"); origin != "" {
		if oh := httpx.ExtractOriginHost(origin); oh != "" {
			rpID = oh
		}
	}
	if rpID == "" {
		rpID = srv.CustomDomain
	}
	origins := []string{"https://" + rpID, "http://" + rpID}
	if orig := r.Header.Get("Origin"); orig != "" {
		origins = append(origins, orig)
	}
	return webauthn.New(&webauthn.Config{
		RPDisplayName: srv.ServerName,
		RPID:          rpID,
		RPOrigins:     origins,
	})
}

func (h *Handler) loadUserCreds(ctx context.Context, serverID, emailAddr string) []webauthn.Credential {
	rows, err := h.db.QueryContext(ctx, `SELECT raw_json FROM webauthn_credentials WHERE server_id = ? AND LOWER(email) = ?`, serverID, strings.ToLower(emailAddr))
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []webauthn.Credential
	for rows.Next() {
		var raw string
		if rows.Scan(&raw) == nil && raw != "" {
			var c webauthn.Credential
			if json.Unmarshal([]byte(raw), &c) == nil {
				out = append(out, c)
			}
		}
	}
	return out
}

func (h *Handler) webauthnRegisterOptions(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	em := httpx.GetSessionEmail(r)
	if em == "" {
		httpx.WriteError(w, r, http.StatusUnauthorized, "Authentication required")
		return
	}
	wa, err := h.newWebAuthn(r, srv)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	user := &webAuthnUser{email: em, creds: h.loadUserCreds(r.Context(), srv.ID, em)}
	opts, sessData, err := wa.BeginRegistration(user)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	challengeID := httpx.NewID()
	sessBytes, _ := json.Marshal(sessData)
	_, _ = h.db.ExecContext(r.Context(), `INSERT INTO webauthn_challenges (id, server_id, challenge_json, email, expires_at) VALUES (?, ?, ?, ?, ?)`,
		challengeID, srv.ID, string(sessBytes), em, time.Now().Add(5*time.Minute).UnixMilli())
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"challengeId": challengeID,
		"options":     opts.Response,
	})
}

func (h *Handler) webauthnRegisterVerify(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	em := httpx.GetSessionEmail(r)
	if em == "" {
		httpx.WriteError(w, r, http.StatusUnauthorized, "Authentication required")
		return
	}
	var body struct {
		ChallengeID string          `json:"challengeId"`
		Name        string          `json:"name"`
		Credential  json.RawMessage `json:"credential"`
	}
	if err := httpx.ReadJSON(r, &body); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid request")
		return
	}
	var challJSON string
	var exp int64
	err := h.db.QueryRowContext(r.Context(), `SELECT challenge_json, expires_at FROM webauthn_challenges WHERE id = ? AND server_id = ?`, body.ChallengeID, srv.ID).Scan(&challJSON, &exp)
	if err != nil || exp < httpx.NowMs() {
		httpx.WriteError(w, r, http.StatusBadRequest, "Challenge expired or not found")
		return
	}
	_, _ = h.db.ExecContext(r.Context(), `DELETE FROM webauthn_challenges WHERE id = ?`, body.ChallengeID)

	var sessData webauthn.SessionData
	_ = json.Unmarshal([]byte(challJSON), &sessData)
	wa, err := h.newWebAuthn(r, srv)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	user := &webAuthnUser{email: em, creds: h.loadUserCreds(r.Context(), srv.ID, em)}
	fakeReq, _ := http.NewRequest(http.MethodPost, r.URL.String(), strings.NewReader(string(body.Credential)))
	cred, err := wa.FinishRegistration(user, sessData, fakeReq)
	if err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, fmt.Sprintf("WebAuthn verification failed: %v", err))
		return
	}
	rawCred, _ := json.Marshal(cred)
	credName := strings.TrimSpace(body.Name)
	if credName == "" {
		credName = "Passkey"
	}
	_, err = h.db.ExecContext(r.Context(), `
		INSERT INTO webauthn_credentials (id, server_id, email, credential_id, public_key_cose, signature_count, user_handle, name, raw_json, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		httpx.NewID(), srv.ID, em, base64.RawURLEncoding.EncodeToString(cred.ID), cred.PublicKey, int64(cred.Authenticator.SignCount), em, credName, string(rawCred), httpx.NowMs(),
	)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to save passkey")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (h *Handler) webauthnListCredentials(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	em := httpx.GetSessionEmail(r)
	if em == "" {
		httpx.WriteError(w, r, http.StatusUnauthorized, "Authentication required")
		return
	}
	rows, err := h.db.QueryContext(r.Context(), `SELECT id, name, created_at, COALESCE(last_used_at, 0) FROM webauthn_credentials WHERE server_id = ? AND LOWER(email) = ? ORDER BY created_at DESC`, srv.ID, strings.ToLower(em))
	if err != nil {
		httpx.WriteProto(w, r, http.StatusOK, &modlv1.WebAuthnCredentialsResponse{})
		return
	}
	defer rows.Close()
	var list []*modlv1.WebAuthnCredentialMetadata
	for rows.Next() {
		var id, name string
		var created, used int64
		if rows.Scan(&id, &name, &created, &used) == nil {
			item := &modlv1.WebAuthnCredentialMetadata{
				Id:        id,
				Name:      name,
				CreatedAt: timestamppb.New(time.UnixMilli(created)),
			}
			if used > 0 {
				item.LastUsedAt = timestamppb.New(time.UnixMilli(used))
			}
			list = append(list, item)
		}
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.WebAuthnCredentialsResponse{Credentials: list})
}

func (h *Handler) webauthnRenameCredential(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	em := httpx.GetSessionEmail(r)
	if em == "" {
		httpx.WriteError(w, r, http.StatusUnauthorized, "Authentication required")
		return
	}
	var req modlv1.RenameWebAuthnCredentialRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	res, _ := h.db.ExecContext(r.Context(), `UPDATE webauthn_credentials SET name = ? WHERE id = ? AND server_id = ? AND LOWER(email) = ?`,
		strings.TrimSpace(req.Name), chi.URLParam(r, "id"), srv.ID, strings.ToLower(em))
	n, _ := res.RowsAffected()
	if n == 0 {
		httpx.WriteError(w, r, http.StatusNotFound, "Credential not found")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.WebAuthnCredentialMutationResponse{Success: true})
}

func (h *Handler) webauthnDeleteCredential(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	em := httpx.GetSessionEmail(r)
	if em == "" {
		httpx.WriteError(w, r, http.StatusUnauthorized, "Authentication required")
		return
	}
	res, _ := h.db.ExecContext(r.Context(), `DELETE FROM webauthn_credentials WHERE id = ? AND server_id = ? AND LOWER(email) = ?`,
		chi.URLParam(r, "id"), srv.ID, strings.ToLower(em))
	n, _ := res.RowsAffected()
	if n == 0 {
		httpx.WriteError(w, r, http.StatusNotFound, "Credential not found")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.WebAuthnCredentialMutationResponse{Success: true})
}

func (h *Handler) webauthnLoginStart(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	wa, err := h.newWebAuthn(r, srv)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	opts, sessData, err := wa.BeginDiscoverableLogin()
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	challengeID := httpx.NewID()
	b, _ := json.Marshal(sessData)
	_, _ = h.db.ExecContext(r.Context(), `INSERT INTO webauthn_challenges (id, server_id, challenge_json, email, expires_at) VALUES (?, ?, ?, '', ?)`,
		challengeID, srv.ID, string(b), time.Now().Add(5*time.Minute).UnixMilli())
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"challengeId": challengeID,
		"options":     opts.Response,
	})
}

func (h *Handler) webauthnLoginOptions(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var body struct {
		Email string `json:"email"`
	}
	if err := httpx.ReadJSON(r, &body); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid request")
		return
	}
	em := strings.ToLower(strings.TrimSpace(body.Email))
	if !h.roles.IsAuthorizedEmail(r.Context(), srv, em) {
		httpx.WriteError(w, r, http.StatusUnauthorized, "Invalid credentials")
		return
	}
	wa, err := h.newWebAuthn(r, srv)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	user := &webAuthnUser{email: em, creds: h.loadUserCreds(r.Context(), srv.ID, em)}
	opts, sessData, err := wa.BeginLogin(user)
	if err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	challengeID := httpx.NewID()
	b, _ := json.Marshal(sessData)
	_, _ = h.db.ExecContext(r.Context(), `INSERT INTO webauthn_challenges (id, server_id, challenge_json, email, expires_at) VALUES (?, ?, ?, ?, ?)`,
		challengeID, srv.ID, string(b), em, time.Now().Add(5*time.Minute).UnixMilli())
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"challengeId": challengeID,
		"options":     opts.Response,
	})
}

func (h *Handler) webauthnLoginVerify(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var body struct {
		ChallengeID string          `json:"challengeId"`
		Credential  json.RawMessage `json:"credential"`
	}
	if err := httpx.ReadJSON(r, &body); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid request")
		return
	}
	var challJSON, emailAddr string
	var exp int64
	err := h.db.QueryRowContext(r.Context(), `SELECT challenge_json, email, expires_at FROM webauthn_challenges WHERE id = ? AND server_id = ?`, body.ChallengeID, srv.ID).Scan(&challJSON, &emailAddr, &exp)
	if err != nil || exp < httpx.NowMs() {
		httpx.WriteError(w, r, http.StatusBadRequest, "Challenge expired or not found")
		return
	}
	_, _ = h.db.ExecContext(r.Context(), `DELETE FROM webauthn_challenges WHERE id = ?`, body.ChallengeID)

	var sessData webauthn.SessionData
	_ = json.Unmarshal([]byte(challJSON), &sessData)
	wa, err := h.newWebAuthn(r, srv)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	fakeReq, _ := http.NewRequest(http.MethodPost, r.URL.String(), strings.NewReader(string(body.Credential)))
	var matchedEmail string
	if emailAddr != "" {
		user := &webAuthnUser{email: emailAddr, creds: h.loadUserCreds(r.Context(), srv.ID, emailAddr)}
		cred, err := wa.FinishLogin(user, sessData, fakeReq)
		if err != nil {
			httpx.WriteError(w, r, http.StatusUnauthorized, "Passkey authentication failed")
			return
		}
		matchedEmail = emailAddr
		_, _ = h.db.ExecContext(r.Context(), `UPDATE webauthn_credentials SET last_used_at = ?, signature_count = ? WHERE server_id = ? AND credential_id = ?`,
			httpx.NowMs(), int64(cred.Authenticator.SignCount), srv.ID, base64.RawURLEncoding.EncodeToString(cred.ID))
	} else {
		_, cred, err := wa.FinishPasskeyLogin(func(rawID, userHandle []byte) (webauthn.User, error) {
			em := string(userHandle)
			if em == "" {
				_ = h.db.QueryRowContext(r.Context(), `SELECT email FROM webauthn_credentials WHERE server_id = ? AND credential_id = ?`, srv.ID, base64.RawURLEncoding.EncodeToString(rawID)).Scan(&em)
			}
			matchedEmail = em
			return &webAuthnUser{email: em, creds: h.loadUserCreds(r.Context(), srv.ID, em)}, nil
		}, sessData, fakeReq)
		if err != nil || matchedEmail == "" {
			httpx.WriteError(w, r, http.StatusUnauthorized, "Passkey authentication failed")
			return
		}
		_, _ = h.db.ExecContext(r.Context(), `UPDATE webauthn_credentials SET last_used_at = ?, signature_count = ? WHERE server_id = ? AND credential_id = ?`,
			httpx.NowMs(), int64(cred.Authenticator.SignCount), srv.ID, base64.RawURLEncoding.EncodeToString(cred.ID))
	}
	sess, err := h.CreateSession(r.Context(), srv.ID, matchedEmail, httpx.GetClientIP(r), r.UserAgent())
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to create session")
		return
	}
	http.SetCookie(w, httpx.CreateSessionCookie(h.cfg, sess.ID))
	_, _ = rand.Read(make([]byte, 1))
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "email": matchedEmail})
}
