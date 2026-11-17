package httpx

import (
	"net"
	"net/http"
	"strings"

	"modl-go/internal/config"
	"modl-go/internal/models"
)

func SecurityHeaders(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			if cfg.CookieSecure {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}
			next.ServeHTTP(w, r)
		})
	}
}

func ClientIPMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := resolveClientIP(cfg, r)
			next.ServeHTTP(w, r.WithContext(WithClientIP(r.Context(), ip)))
		})
	}
}

func resolveClientIP(cfg *config.Config, r *http.Request) string {
	if cfg.TrustProxyHeaders {
		if cfg.ClientIPHeader != "" {
			if h := strings.TrimSpace(r.Header.Get(cfg.ClientIPHeader)); h != "" {
				return h
			}
		}
		if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
			parts := strings.Split(xff, ",")
			if len(parts) > 0 {
				return strings.TrimSpace(parts[0])
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}

func CORSMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && isOriginAllowed(cfg, r, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key, X-Server-Domain, X-Turnstile-Token")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func resolveRequestHost(cfg *config.Config, r *http.Request) string {
	if r == nil {
		return ""
	}
	if cfg.TrustProxyHeaders {
		if fh := r.Header.Get("X-Forwarded-Host"); fh != "" {
			first, _, _ := strings.Cut(fh, ",")
			if h := NormalizeServerDomain(first); h != "" {
				return h
			}
		}
	}
	return NormalizeServerDomain(r.Host)
}

func isOriginAllowed(cfg *config.Config, r *http.Request, origin string) bool {
	host := ExtractOriginHost(origin)
	if host == "" {
		return false
	}
	requestHost := resolveRequestHost(cfg, r)
	if requestHost != "" && host == requestHost {
		return true
	}
	if IsDevLocalhost(host) && (cfg.DevelopmentMode || IsDevLocalhost(requestHost)) {
		return true
	}
	for _, sys := range cfg.CORSSystemOrigins {
		if strings.EqualFold(strings.TrimRight(sys, "/"), strings.TrimRight(origin, "/")) {
			return true
		}
	}
	for _, appDomain := range cfg.CORSAppDomains {
		d := NormalizeServerDomain(appDomain)
		if d != "" && (host == d || strings.HasSuffix(host, "."+d)) {
			return true
		}
	}
	for _, domain := range []string{cfg.ModlDomain, cfg.BootstrapDomain} {
		d := NormalizeServerDomain(domain)
		if d != "" && (host == d || strings.HasSuffix(host, "."+d)) {
			return true
		}
	}
	return false
}

func ServerHeaderMiddleware(cfg *config.Config, servers models.ServerResolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}
			path := r.URL.Path
			if strings.HasPrefix(path, "/v1/public/registration") || strings.HasPrefix(path, "/v1/public/evidence-upload") {
				next.ServeHTTP(w, r)
				return
			}

			var requestHost string
			if cfg.TrustProxyHeaders {
				if fh := r.Header.Get("X-Forwarded-Host"); fh != "" {
					first, _, _ := strings.Cut(fh, ",")
					requestHost = NormalizeServerDomain(first)
				}
			}
			if requestHost == "" {
				requestHost = NormalizeServerDomain(r.Host)
			}
			legacyDomain := NormalizeServerDomain(r.Header.Get("X-Server-Domain"))
			serverDomain := requestHost
			if serverDomain == "" {
				serverDomain = legacyDomain
			}
			if legacyDomain != "" && requestHost != "" && !IsDevLocalhost(requestHost) {
				// If frontend explicitly passes X-Server-Domain while hitting api.domain, prefer X-Server-Domain when requestHost is api.<domain>
				if strings.HasPrefix(requestHost, "api.") {
					serverDomain = legacyDomain
				}
			}

			if (cfg.DevelopmentMode || cfg.DevServerDomain != "") && IsDevLocalhost(serverDomain) && cfg.DevServerDomain != "" {
				serverDomain = NormalizeServerDomain(cfg.DevServerDomain)
			}

			var srv *models.Server
			var err error
			if serverDomain != "" {
				srv, err = servers.GetByDomain(r.Context(), serverDomain)
			}
			if (srv == nil || err != nil) && cfg.SingleTenant {
				srv, err = servers.GetDefault(r.Context())
				if srv != nil && serverDomain == "" {
					serverDomain = srv.CustomDomain
				}
			}
			if err != nil || srv == nil {
				WriteError(w, r, http.StatusNotFound, "Invalid server domain")
				return
			}

			ctx := WithServer(r.Context(), srv)
			ctx = WithServerDomain(ctx, serverDomain)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ApiKeyMiddleware(servers models.ServerResolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := strings.TrimSpace(r.Header.Get("X-API-Key"))
			if apiKey == "" {
				WriteErrorCode(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Unauthorized")
				return
			}
			srv, err := servers.GetByAPIKey(r.Context(), apiKey)
			if err != nil || srv == nil {
				WriteErrorCode(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Unauthorized")
				return
			}
			next.ServeHTTP(w, r.WithContext(WithServer(r.Context(), srv)))
		})
	}
}

func SessionAuthMiddleware(cfg *config.Config, sessions models.SessionResolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}
			srv := GetServer(r)
			if srv == nil {
				next.ServeHTTP(w, r)
				return
			}
			for _, c := range r.Cookies() {
				if c.Name != cfg.SessionCookieName || strings.TrimSpace(c.Value) == "" {
					continue
				}
				sess, err := sessions.FindAndRefresh(r.Context(), srv.ID, c.Value)
				if err == nil && sess != nil {
					http.SetCookie(w, CreateSessionCookie(cfg, c.Value))
					r = r.WithContext(WithSession(r.Context(), sess))
					break
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequirePanelAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if GetSession(r) == nil {
			WriteErrorCode(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func OriginCSRFMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if models.IsReadOnlyMethod(r.Method) {
				next.ServeHTTP(w, r)
				return
			}
			hasCookie := false
			for _, c := range r.Cookies() {
				if c.Name == cfg.SessionCookieName && c.Value != "" {
					hasCookie = true
					break
				}
			}
			if !hasCookie {
				next.ServeHTTP(w, r)
				return
			}
			origin := r.Header.Get("Origin")
			if origin != "" {
				if isOriginAllowed(cfg, r, origin) {
					next.ServeHTTP(w, r)
					return
				}
				WriteError(w, r, http.StatusForbidden, "Cross-site request blocked")
				return
			}
			if referer := r.Header.Get("Referer"); referer != "" {
				if isOriginAllowed(cfg, r, referer) {
					next.ServeHTTP(w, r)
					return
				}
				WriteError(w, r, http.StatusForbidden, "Cross-site request blocked")
				return
			}
			if strings.EqualFold(r.Header.Get("Sec-Fetch-Site"), "same-origin") {
				next.ServeHTTP(w, r)
				return
			}
			WriteError(w, r, http.StatusForbidden, "Cross-site request blocked")
		})
	}
}

func RequirePanelPolicy(perms models.PermissionChecker, policy models.AccessPolicy) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srv := GetServer(r)
			if srv == nil {
				WriteError(w, r, http.StatusNotFound, "Server not found")
				return
			}
			email := GetSessionEmail(r)
			if email == "" {
				WriteErrorCode(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
				return
			}
			if perms.IsSuperAdmin(srv, email) {
				next.ServeHTTP(w, r)
				return
			}
			if policy.Rule == models.RulePermitAll {
				next.ServeHTTP(w, r)
				return
			}
			if policy.Rule == models.RuleSuperAdmin {
				roleID := perms.PanelRoleID(r.Context(), srv, email)
				allowed := false
				for _, p := range policy.SupersedesPermissions {
					if perms.HasPermission(r.Context(), srv, roleID, p) {
						allowed = true
						break
					}
				}
				if !allowed {
					WriteError(w, r, http.StatusForbidden, "Insufficient permissions")
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			roleID := perms.PanelRoleID(r.Context(), srv, email)
			readOnly := models.IsReadOnlyMethod(r.Method)
			var ok bool
			switch policy.Rule {
			case models.RulePlayerAccess:
				if readOnly {
					ok = perms.HasPermission(r.Context(), srv, roleID, "punishment.view") ||
						perms.HasPermission(r.Context(), srv, roleID, "punishment.modify") ||
						perms.HasAnyWithPrefix(r.Context(), srv, roleID, "punishment.apply.")
				} else {
					ok = perms.HasPermission(r.Context(), srv, roleID, "punishment.modify")
				}
			case models.RulePunishmentTypeAccess:
				if readOnly {
					ok = perms.HasPermission(r.Context(), srv, roleID, "admin.settings.view.punishments") ||
						perms.HasPermission(r.Context(), srv, roleID, "punishment.view") ||
						perms.HasPermission(r.Context(), srv, roleID, "punishment.modify") ||
						perms.HasAnyWithPrefix(r.Context(), srv, roleID, "punishment.apply.")
				} else {
					ok = perms.HasPermission(r.Context(), srv, roleID, "admin.settings.modify.punishments")
				}
			case models.RuleAppealReply:
				if readOnly {
					ok = perms.HasPermission(r.Context(), srv, roleID, "ticket.view.all")
				} else {
					ok = perms.HasPermission(r.Context(), srv, roleID, "appeal.modify") ||
						perms.HasPermission(r.Context(), srv, roleID, "ticket.reply.all")
				}
			default:
				needed := policy.Modify
				if readOnly {
					needed = policy.View
				}
				if needed != "" {
					ok = perms.HasPermission(r.Context(), srv, roleID, needed)
				}
			}
			if !ok {
				WriteError(w, r, http.StatusForbidden, "Insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
