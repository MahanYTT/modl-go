package httpx

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"modl-go/internal/config"
)

func NewID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func RandomToken(numBytes int) string {
	b := make([]byte, numBytes)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func RandomAPIKey() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return "modl_" + hex.EncodeToString(b)
}

func GenerateNumericCode(length int) string {
	var sb strings.Builder
	sb.Grow(length)
	max := big.NewInt(10)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			sb.WriteByte('0')
			continue
		}
		sb.WriteByte(byte('0' + n.Int64()))
	}
	return sb.String()
}

func HashOneTimeCode(code, secret string) string {
	if secret != "" {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(code))
		return base64.StdEncoding.EncodeToString(mac.Sum(nil))
	}
	sum := sha256.Sum256([]byte(code))
	return base64.StdEncoding.EncodeToString(sum[:])
}

func VerifyOneTimeCode(code, expectedHash, secret string) bool {
	got := HashOneTimeCode(code, secret)
	return hmac.Equal([]byte(got), []byte(expectedHash))
}

func NormalizeServerDomain(raw string) string {
	s := strings.TrimSpace(strings.ToLower(raw))
	if s == "" {
		return ""
	}
	if strings.Contains(s, "://") {
		if u, err := url.Parse(s); err == nil {
			s = u.Hostname()
		}
	}
	if host, _, err := net.SplitHostPort(s); err == nil {
		s = host
	} else if idx := strings.LastIndex(s, ":"); idx > 0 && !strings.Contains(s[:idx], ":") {
		s = s[:idx]
	}
	s = strings.TrimSuffix(s, ".")
	return s
}

func ExtractOriginHost(origin string) string {
	u, err := url.Parse(strings.TrimSpace(origin))
	if err != nil || u.Host == "" {
		return ""
	}
	return strings.ToLower(u.Hostname())
}

func IsDevLocalhost(host string) bool {
	h := strings.ToLower(strings.TrimSpace(host))
	return h == "localhost" || h == "127.0.0.1" || h == "::1" || strings.HasSuffix(h, ".localhost")
}

func QueryInt(r *http.Request, key string, def, minVal, maxVal int) int {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	if n < minVal {
		return minVal
	}
	if maxVal > 0 && n > maxVal {
		return maxVal
	}
	return n
}

func NowMs() int64 {
	return time.Now().UnixMilli()
}

func CreateSessionCookie(cfg *config.Config, token string) *http.Cookie {
	sameSite := http.SameSiteStrictMode
	if cfg.DevelopmentMode || !cfg.CookieSecure {
		sameSite = http.SameSiteLaxMode
	}
	c := &http.Cookie{
		Name:     cfg.SessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(cfg.SessionDurationSec),
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: sameSite,
	}
	if cfg.CookieDomain != "" {
		c.Domain = cfg.CookieDomain
	}
	return c
}

func WriteExpiredSessionCookies(w http.ResponseWriter, cfg *config.Config) {
	sameSite := http.SameSiteStrictMode
	if cfg.DevelopmentMode || !cfg.CookieSecure {
		sameSite = http.SameSiteLaxMode
	}
	addExpired := func(domain string) {
		c := &http.Cookie{
			Name:     cfg.SessionCookieName,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			Secure:   cfg.CookieSecure,
			SameSite: sameSite,
			Domain:   domain,
		}
		http.SetCookie(w, c)
	}
	addExpired("")
	if cfg.CookieDomain != "" {
		addExpired(cfg.CookieDomain)
		if !strings.HasPrefix(cfg.CookieDomain, ".") {
			addExpired("." + cfg.CookieDomain)
		}
	}
}

func FormatDurationMs(ms int64) string {
	if ms <= 0 {
		return "permanent"
	}
	d := time.Duration(ms) * time.Millisecond
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	if days > 0 {
		return fmt.Sprintf("%dd", days)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dm", mins)
}
