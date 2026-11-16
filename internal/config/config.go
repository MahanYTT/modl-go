package config

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Port               string
	DBPath             string
	StorageDir         string
	PanelDistDir       string
	ModlDomain         string
	CookieDomain       string
	CookieSecure       bool
	DevelopmentMode    bool
	DevServerDomain    string
	SingleTenant       bool
	TrustProxyHeaders  bool
	ClientIPHeader     string
	TrustedProxyCount  int
	AuthCodeHashSecret string
	SessionCookieName  string
	SessionDurationSec int64
	CORSAppDomains     []string
	CORSSystemOrigins  []string

	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPAuth     bool
	SMTPTLS      bool
	EmailDomain  string

	GeminiAPIKey  string
	GeminiModelID string

	BootstrapServerName string
	BootstrapDomain     string
	BootstrapAdminEmail string
	BootstrapAPIKey     string

	RealtimeEnabled bool
}

func Load() *Config {
	loadDotEnv(".env")

	devMode := getBool("MODL_DEVELOPMENT_MODE", false)
	modlDomain := getEnv("MODL_DOMAIN", "localhost")
	cookieSecureDefault := !devMode && modlDomain != "localhost" && !strings.HasPrefix(modlDomain, "localhost:")

	secret := getEnv("MODL_AUTH_CODE_HASH_SECRET", "")
	if secret == "" {
		b := make([]byte, 32)
		_, _ = rand.Read(b)
		secret = hex.EncodeToString(b)
	}

	return &Config{
		Port:                getEnv("PORT", "8080"),
		DBPath:              getEnv("MODL_DB_PATH", "modl.db"),
		StorageDir:          getEnv("MODL_STORAGE_DIR", "storage"),
		PanelDistDir:        getEnv("MODL_PANEL_DIST_DIR", ""),
		ModlDomain:          modlDomain,
		CookieDomain:        getEnv("MODL_COOKIE_DOMAIN", ""),
		CookieSecure:        getBool("MODL_COOKIE_SECURE", cookieSecureDefault),
		DevelopmentMode:     devMode,
		DevServerDomain:     getEnv("MODL_DEV_SERVER_DOMAIN", ""),
		SingleTenant:        getBool("MODL_SINGLE_TENANT", true),
		TrustProxyHeaders:   getBool("MODL_TRUST_PROXY_HEADERS", false),
		ClientIPHeader:      getEnv("MODL_CLIENT_IP_HEADER", ""),
		TrustedProxyCount:   getInt("MODL_TRUSTED_PROXY_COUNT", 1),
		AuthCodeHashSecret:  secret,
		SessionCookieName:   getEnv("MODL_SESSION_COOKIE_NAME", "MODL_SESSION"),
		SessionDurationSec:  getInt64("MODL_AUTH_SESSION_DURATION_SECONDS", 14*24*60*60),
		CORSAppDomains:      splitCSV(getEnv("MODL_CORS_APP_DOMAINS", modlDomain)),
		CORSSystemOrigins:   splitCSV(getEnv("MODL_CORS_SYSTEM_ORIGINS", "http://localhost:5173,http://localhost:8080")),
		SMTPHost:            getEnv("SMTP_HOST", ""),
		SMTPPort:            getInt("SMTP_PORT", 587),
		SMTPUsername:        getEnv("SMTP_USERNAME", ""),
		SMTPPassword:        getEnv("SMTP_PASSWORD", ""),
		SMTPAuth:            getBool("SMTP_AUTH", true),
		SMTPTLS:             getBool("SMTP_TLS", true),
		EmailDomain:         getEnv("MODL_EMAIL_DOMAIN", modlDomain),
		GeminiAPIKey:        getEnv("GEMINI_API_KEY", ""),
		GeminiModelID:       getEnv("MODL_GEMINI_MODEL_ID", "gemini-2.5-flash-lite"),
		BootstrapServerName: getEnv("MODL_BOOTSTRAP_SERVER_NAME", ""),
		BootstrapDomain:     getEnv("MODL_BOOTSTRAP_DOMAIN", ""),
		BootstrapAdminEmail: getEnv("MODL_BOOTSTRAP_ADMIN_EMAIL", ""),
		BootstrapAPIKey:     getEnv("MODL_BOOTSTRAP_API_KEY", ""),
		RealtimeEnabled:     getBool("MODL_REALTIME_WS_ENABLED", true),
	}
}

func loadDotEnv(path string) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.Trim(strings.TrimSpace(v), `"'`)
		if k != "" && os.Getenv(k) == "" {
			_ = os.Setenv(k, v)
		}
	}
}

func getEnv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func getBool(key string, def bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if v == "" {
		return def
	}
	return v == "true" || v == "1" || v == "yes"
}

func getInt(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func getInt64(key string, def int64) int64 {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}
	return n
}

func splitCSV(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if t := strings.TrimSpace(part); t != "" {
			out = append(out, t)
		}
	}
	return out
}
