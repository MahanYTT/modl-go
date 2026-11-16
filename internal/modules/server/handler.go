package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/config"
	"modl-go/internal/email"
	"modl-go/internal/httpx"
	"modl-go/internal/models"

	"github.com/go-chi/chi/v5"
)

var subdomainRe = regexp.MustCompile(`^[a-z0-9-]+$`)

type SessionCreator interface {
	CreateSession(ctx context.Context, serverID, email, ip, ua string) (*models.AuthSession, error)
}

type Handler struct {
	db       *sql.DB
	cfg      *config.Config
	mailer   *email.Sender
	sessions SessionCreator
	perms    models.PermissionChecker
}

func NewHandler(db *sql.DB, cfg *config.Config, mailer *email.Sender, perms models.PermissionChecker) *Handler {
	return &Handler{db: db, cfg: cfg, mailer: mailer, perms: perms}
}

func (h *Handler) SetSessionCreator(sc SessionCreator) {
	h.sessions = sc
}

func (h *Handler) RegisterPanelRoutes(r chi.Router) {
	r.Route("/server", func(r chi.Router) {
		r.Use(httpx.RequirePanelPolicy(h.perms, models.AccessPolicy{
			Rule:   models.RuleRequirePermission,
			View:   "admin.settings.view",
			Modify: "admin.settings.modify",
		}))
		r.Get("/provisioning-status", h.getProvisioningStatus)
	})
}

func (h *Handler) RegisterPublicRoutes(r chi.Router) {
	r.Post("/server/register", h.legacyGone)
	r.Post("/server/check-availability", h.legacyCheckGone)

	r.Route("/registration", func(r chi.Router) {
		r.Post("/", h.registerWeb)
		r.Get("/verify", h.verifyEmailGet)
		r.Post("/verify", h.verifyEmailPost)
		r.Get("/setup-status", h.setupStatusGet)
		r.Post("/setup-status", h.setupStatusPost)
		r.Post("/auto-login", h.autoLogin)
		r.Post("/cli", h.registerCLI)
		r.Post("/cli/status", h.cliSetupStatus)
		r.Post("/api-key", h.getAPIKeyFromToken)
		r.Post("/check-availability", h.checkAvailability)
	})
}

func (h *Handler) BootstrapIfConfigured(ctx context.Context) (*models.Server, error) {
	if h.cfg.BootstrapAdminEmail == "" && h.cfg.BootstrapDomain == "" && h.cfg.BootstrapServerName == "" {
		// Check if any server exists; if not and SingleTenant, log hint
		srv, _ := h.GetDefault(ctx)
		return srv, nil
	}
	existing, _ := h.GetDefault(ctx)
	if existing != nil {
		return existing, nil
	}
	name := h.cfg.BootstrapServerName
	if name == "" {
		name = "My Minecraft Server"
	}
	domain := httpx.NormalizeServerDomain(h.cfg.BootstrapDomain)
	if domain == "" {
		domain = httpx.NormalizeServerDomain(h.cfg.ModlDomain)
	}
	if domain == "" {
		domain = "localhost"
	}
	adminEmail := strings.ToLower(strings.TrimSpace(h.cfg.BootstrapAdminEmail))
	if adminEmail == "" {
		adminEmail = "admin@" + domain
	}
	apiKey := h.cfg.BootstrapAPIKey
	if apiKey == "" {
		apiKey = httpx.RandomAPIKey()
	}
	srv, err := h.createServer(ctx, name, domain, adminEmail, apiKey, true)
	if err != nil {
		return nil, err
	}
	log.Printf("[BOOTSTRAP] Created default server name=%q domain=%q admin=%q api_key=%s", srv.ServerName, srv.CustomDomain, srv.AdminEmail, srv.APIKey)
	return srv, nil
}

func (h *Handler) createServer(ctx context.Context, serverName, customDomain, adminEmail, apiKey string, verified bool) (*models.Server, error) {
	now := httpx.NowMs()
	if apiKey == "" {
		apiKey = httpx.RandomAPIKey()
	}
	verifToken := ""
	if !verified {
		verifToken = httpx.RandomToken(32)
	}
	srv := &models.Server{
		ID:                     httpx.NewID(),
		ServerName:             serverName,
		CustomDomain:           httpx.NormalizeServerDomain(customDomain),
		AdminEmail:             strings.ToLower(strings.TrimSpace(adminEmail)),
		EmailVerified:          verified,
		EmailVerificationToken: verifToken,
		ProvisioningStatus:     "COMPLETED",
		CLISetupToken:          httpx.RandomToken(24),
		APIKey:                 apiKey,
		CreatedAt:              now,
		UpdatedAt:              now,
	}
	evInt := 0
	if srv.EmailVerified {
		evInt = 1
	}
	_, err := h.db.ExecContext(ctx, `
		INSERT INTO servers (
			id, server_name, custom_domain, custom_domain_override, admin_email,
			email_verified, email_verification_token, provisioning_status,
			provisioning_notes, provisioning_sign_in_token, cli_setup_token,
			api_key, created_at, updated_at
		) VALUES (?, ?, ?, '', ?, ?, ?, ?, '', '', ?, ?, ?, ?)`,
		srv.ID, srv.ServerName, srv.CustomDomain, srv.AdminEmail,
		evInt, srv.EmailVerificationToken, srv.ProvisioningStatus,
		srv.CLISetupToken, srv.APIKey, srv.CreatedAt, srv.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return srv, nil
}

const serverSelectCols = `
	id, server_name, custom_domain, custom_domain_override, admin_email,
	email_verified, email_verification_token, provisioning_status,
	provisioning_notes, provisioning_sign_in_token, COALESCE(provisioning_sign_in_token_expires_at, 0),
	cli_setup_token, api_key, online_player_count, user_count, ticket_count,
	COALESCE(last_stats_updated_at, 0), COALESCE(last_activity_at, 0),
	COALESCE(staff_permissions_updated_at, 0), COALESCE(punishment_types_updated_at, 0),
	created_at, updated_at`

func scanServer(row interface{ Scan(dest ...any) error }) (*models.Server, error) {
	var s models.Server
	var ev int
	err := row.Scan(
		&s.ID, &s.ServerName, &s.CustomDomain, &s.CustomDomainOverride, &s.AdminEmail,
		&ev, &s.EmailVerificationToken, &s.ProvisioningStatus,
		&s.ProvisioningNotes, &s.ProvisioningSignInToken, &s.ProvisioningSignInTokenExp,
		&s.CLISetupToken, &s.APIKey, &s.OnlinePlayerCount, &s.UserCount, &s.TicketCount,
		&s.LastStatsUpdatedAt, &s.LastActivityAt,
		&s.StaffPermissionsUpdatedAt, &s.PunishmentTypesUpdatedAt,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	s.EmailVerified = ev != 0
	return &s, nil
}

func (h *Handler) GetByDomain(ctx context.Context, domain string) (*models.Server, error) {
	norm := httpx.NormalizeServerDomain(domain)
	if norm == "" {
		return nil, sql.ErrNoRows
	}
	// Match exact custom_domain, custom_domain_override, or subdomain prefix before "." + ModlDomain
	row := h.db.QueryRowContext(ctx,
		`SELECT `+serverSelectCols+` FROM servers WHERE LOWER(custom_domain) = ? OR LOWER(custom_domain_override) = ? LIMIT 1`,
		norm, norm,
	)
	srv, err := scanServer(row)
	if err == nil {
		return srv, nil
	}
	if h.cfg.ModlDomain != "" && strings.HasSuffix(norm, "."+strings.ToLower(h.cfg.ModlDomain)) {
		sub := strings.TrimSuffix(norm, "."+strings.ToLower(h.cfg.ModlDomain))
		row2 := h.db.QueryRowContext(ctx,
			`SELECT `+serverSelectCols+` FROM servers WHERE LOWER(custom_domain) = ? LIMIT 1`,
			sub,
		)
		return scanServer(row2)
	}
	return nil, err
}

func (h *Handler) GetByAPIKey(ctx context.Context, apiKey string) (*models.Server, error) {
	row := h.db.QueryRowContext(ctx,
		`SELECT `+serverSelectCols+` FROM servers WHERE api_key = ? AND api_key != '' LIMIT 1`,
		apiKey,
	)
	return scanServer(row)
}

func (h *Handler) GetDefault(ctx context.Context) (*models.Server, error) {
	row := h.db.QueryRowContext(ctx, `SELECT `+serverSelectCols+` FROM servers ORDER BY created_at ASC LIMIT 1`)
	return scanServer(row)
}

func (h *Handler) getProvisioningStatus(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Server not found")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.ProvisioningStatusResponse{
		Status:        toProtoProvisioningStatus(srv.ProvisioningStatus),
		ServerName:    srv.ServerName,
		EmailVerified: srv.EmailVerified,
	})
}

func (h *Handler) legacyGone(w http.ResponseWriter, r *http.Request) {
	httpx.WriteProto(w, r, http.StatusGone, &modlv1.ServerRegisterResponse{
		Success: false,
		Message: "Use /v1/public/registration instead.",
	})
}

func (h *Handler) legacyCheckGone(w http.ResponseWriter, r *http.Request) {
	httpx.WriteProto(w, r, http.StatusGone, &modlv1.ServerAvailabilityResponse{
		Message: strPtr("Use /v1/public/registration/check-availability instead."),
	})
}

func (h *Handler) registerWeb(w http.ResponseWriter, r *http.Request) {
	var req modlv1.PublicRegistrationRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if !h.isAvailable(r.Context(), req.Email, req.ServerName, req.CustomDomain) {
		httpx.WriteError(w, r, http.StatusConflict, "Server name, domain, or email is already registered")
		return
	}
	// In self-hosted mode without SMTP, auto-verify immediately so the user can proceed cleanly
	autoVerify := !h.mailer.IsConfigured()
	srv, err := h.createServer(r.Context(), req.ServerName, req.CustomDomain, req.Email, "", autoVerify)
	if err != nil {
		httpx.WriteError(w, r, http.StatusConflict, "Failed to register server")
		return
	}
	if !autoVerify {
		verifyURL := fmt.Sprintf("https://%s/verify?token=%s", h.cfg.ModlDomain, srv.EmailVerificationToken)
		_ = h.mailer.Send(srv.AdminEmail, "Verify your modl server registration",
			fmt.Sprintf("<p>Click to verify <b>%s</b>:</p><p><a href=\"%s\">%s</a></p>", srv.ServerName, verifyURL, verifyURL))
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PublicRegistrationResponse{
		Success: true,
		Message: "Registration successful",
		Server:  &modlv1.ServerSummary{Id: srv.ID, Name: srv.ServerName},
	})
}

func (h *Handler) registerCLI(w http.ResponseWriter, r *http.Request) {
	var req modlv1.CliRegistrationRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if !h.isAvailable(r.Context(), req.Email, req.ServerName, req.CustomDomain) {
		httpx.WriteError(w, r, http.StatusConflict, "Server name, domain, or email is already registered")
		return
	}
	autoVerify := !h.mailer.IsConfigured()
	srv, err := h.createServer(r.Context(), req.ServerName, req.CustomDomain, req.Email, "", autoVerify)
	if err != nil {
		httpx.WriteError(w, r, http.StatusConflict, "Failed to register server")
		return
	}
	if !autoVerify {
		verifyURL := fmt.Sprintf("https://%s/verify?token=%s", h.cfg.ModlDomain, srv.EmailVerificationToken)
		_ = h.mailer.Send(srv.AdminEmail, "Verify your modl server registration",
			fmt.Sprintf("<p>Click to verify <b>%s</b>:</p><p><a href=\"%s\">%s</a></p>", srv.ServerName, verifyURL, verifyURL))
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.CliRegistrationResponse{
		Success:    true,
		Message:    "Registration successful",
		Server:     &modlv1.ServerSummary{Id: srv.ID, Name: srv.ServerName},
		SetupToken: &srv.CLISetupToken,
	})
}

func (h *Handler) verifyEmailGet(w http.ResponseWriter, r *http.Request) {
	h.verifyEmailInternal(w, r, r.URL.Query().Get("token"))
}

func (h *Handler) verifyEmailPost(w http.ResponseWriter, r *http.Request) {
	var req modlv1.EmailVerificationTokenRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	h.verifyEmailInternal(w, r, req.Token)
}

func (h *Handler) verifyEmailInternal(w http.ResponseWriter, r *http.Request, token string) {
	if strings.TrimSpace(token) == "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "Missing verification token")
		return
	}
	row := h.db.QueryRowContext(r.Context(), `SELECT `+serverSelectCols+` FROM servers WHERE email_verification_token = ? LIMIT 1`, token)
	srv, err := scanServer(row)
	if err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid or expired verification token")
		return
	}
	autoLogin := httpx.RandomToken(32)
	exp := time.Now().Add(10 * time.Minute).UnixMilli()
	_, err = h.db.ExecContext(r.Context(),
		`UPDATE servers SET email_verified = 1, email_verification_token = '', provisioning_status = 'COMPLETED', provisioning_sign_in_token = ?, provisioning_sign_in_token_expires_at = ?, updated_at = ? WHERE id = ?`,
		autoLogin, exp, httpx.NowMs(), srv.ID,
	)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to verify email")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.EmailVerificationResponse{
		Success:        true,
		Message:        "Email verified successfully",
		Subdomain:      &srv.CustomDomain,
		AutoLoginToken: &autoLogin,
	})
}

func (h *Handler) setupStatusGet(w http.ResponseWriter, r *http.Request) {
	h.setupStatusInternal(w, r, r.URL.Query().Get("token"))
}

func (h *Handler) setupStatusPost(w http.ResponseWriter, r *http.Request) {
	var req modlv1.SetupStatusRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	h.setupStatusInternal(w, r, req.Token)
}

func (h *Handler) setupStatusInternal(w http.ResponseWriter, r *http.Request, token string) {
	if strings.TrimSpace(token) == "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "Missing token")
		return
	}
	row := h.db.QueryRowContext(r.Context(),
		`SELECT `+serverSelectCols+` FROM servers WHERE provisioning_sign_in_token = ? OR email_verification_token = ? OR cli_setup_token = ? LIMIT 1`,
		token, token, token,
	)
	srv, err := scanServer(row)
	if err != nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Setup token not found")
		return
	}
	st := toProtoProvisioningStatus(srv.ProvisioningStatus)
	msg := "Ready"
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.SetupStatusResponse{
		Subdomain:          &srv.CustomDomain,
		ServerName:         &srv.ServerName,
		EmailVerified:      &srv.EmailVerified,
		ProvisioningStatus: &st,
		Message:            &msg,
	})
}

func (h *Handler) autoLogin(w http.ResponseWriter, r *http.Request) {
	var req modlv1.AutoLoginRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	row := h.db.QueryRowContext(r.Context(), `SELECT `+serverSelectCols+` FROM servers WHERE provisioning_sign_in_token = ? LIMIT 1`, req.Token)
	srv, err := scanServer(row)
	if err != nil || srv.ProvisioningSignInTokenExp < httpx.NowMs() {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid or expired auto-login token")
		return
	}
	_, _ = h.db.ExecContext(r.Context(), `UPDATE servers SET provisioning_sign_in_token = '', provisioning_sign_in_token_expires_at = NULL WHERE id = ?`, srv.ID)
	if h.sessions != nil {
		sess, err := h.sessions.CreateSession(r.Context(), srv.ID, srv.AdminEmail, httpx.GetClientIP(r), r.UserAgent())
		if err == nil {
			http.SetCookie(w, httpx.CreateSessionCookie(h.cfg, sess.ID))
		}
	}
	redirect := "/"
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.AutoLoginResponse{
		Success:     true,
		Message:     "Logged in",
		RedirectUrl: &redirect,
	})
}

func (h *Handler) cliSetupStatus(w http.ResponseWriter, r *http.Request) {
	var req modlv1.CliSetupStatusRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	row := h.db.QueryRowContext(r.Context(), `SELECT `+serverSelectCols+` FROM servers WHERE cli_setup_token = ? LIMIT 1`, req.Token)
	srv, err := scanServer(row)
	if err != nil {
		httpx.WriteProto(w, r, http.StatusBadRequest, &modlv1.CliSetupStatusResponse{
			Success: false,
			Message: strPtr("Invalid or expired setup token."),
		})
		return
	}
	st := toProtoProvisioningStatus(srv.ProvisioningStatus)
	var apiKeyPtr *string
	if srv.EmailVerified {
		apiKeyPtr = &srv.APIKey
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.CliSetupStatusResponse{
		Success:            true,
		EmailVerified:      &srv.EmailVerified,
		ProvisioningStatus: &st,
		ApiKey:             apiKeyPtr,
		Message:            strPtr("Ready"),
	})
}

func (h *Handler) getAPIKeyFromToken(w http.ResponseWriter, r *http.Request) {
	var req modlv1.ProvisioningTokenRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	row := h.db.QueryRowContext(r.Context(), `SELECT `+serverSelectCols+` FROM servers WHERE provisioning_sign_in_token = ? OR cli_setup_token = ? LIMIT 1`, req.Token, req.Token)
	srv, err := scanServer(row)
	if err != nil {
		httpx.WriteProto(w, r, http.StatusBadRequest, &modlv1.RegistrationApiKeyResponse{
			Success: false,
			Message: strPtr("Invalid or expired token."),
		})
		return
	}
	panelURL := "https://" + srv.CustomDomain
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.RegistrationApiKeyResponse{
		Success:  true,
		ApiKey:   &srv.APIKey,
		PanelUrl: &panelURL,
	})
}

func (h *Handler) checkAvailability(w http.ResponseWriter, r *http.Request) {
	var req modlv1.ServerAvailabilityRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	emailAvail, nameAvail, subAvail := true, true, true
	if req.Email != nil && strings.TrimSpace(*req.Email) != "" {
		var c int
		_ = h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM servers WHERE LOWER(admin_email) = ?`, strings.ToLower(strings.TrimSpace(*req.Email))).Scan(&c)
		emailAvail = c == 0
	}
	if req.ServerName != nil && strings.TrimSpace(*req.ServerName) != "" {
		var c int
		_ = h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM servers WHERE LOWER(server_name) = ?`, strings.ToLower(strings.TrimSpace(*req.ServerName))).Scan(&c)
		nameAvail = c == 0
	}
	if req.CustomDomain != nil && strings.TrimSpace(*req.CustomDomain) != "" {
		norm := httpx.NormalizeServerDomain(*req.CustomDomain)
		var c int
		_ = h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM servers WHERE LOWER(custom_domain) = ?`, norm).Scan(&c)
		subAvail = c == 0 && subdomainRe.MatchString(norm)
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.ServerAvailabilityResponse{
		EmailAvailable:     emailAvail,
		NameAvailable:      nameAvail,
		SubdomainAvailable: subAvail,
	})
}

func (h *Handler) isAvailable(ctx context.Context, email, name, domain string) bool {
	var c int
	_ = h.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM servers WHERE LOWER(custom_domain) = ? OR LOWER(server_name) = ?`,
		httpx.NormalizeServerDomain(domain), strings.ToLower(strings.TrimSpace(name)),
	).Scan(&c)
	return c == 0
}

func toProtoProvisioningStatus(s string) modlv1.ProvisioningStatus {
	switch strings.ToUpper(s) {
	case "PENDING":
		return modlv1.ProvisioningStatus_PROVISIONING_STATUS_PENDING
	case "IN_PROGRESS":
		return modlv1.ProvisioningStatus_PROVISIONING_STATUS_IN_PROGRESS
	case "COMPLETED":
		return modlv1.ProvisioningStatus_PROVISIONING_STATUS_COMPLETED
	case "FAILED":
		return modlv1.ProvisioningStatus_PROVISIONING_STATUS_FAILED
	default:
		return modlv1.ProvisioningStatus_PROVISIONING_STATUS_COMPLETED
	}
}

func strPtr(s string) *string { return &s }
