package httpx

import (
	"context"
	"net/http"

	"modl-go/internal/models"
)

type ctxKey int

const (
	keyServer ctxKey = iota
	keyServerDomain
	keySession
	keyClientIP
)

func WithServer(ctx context.Context, srv *models.Server) context.Context {
	return context.WithValue(ctx, keyServer, srv)
}

func GetServer(r *http.Request) *models.Server {
	if v, ok := r.Context().Value(keyServer).(*models.Server); ok {
		return v
	}
	return nil
}

func WithServerDomain(ctx context.Context, domain string) context.Context {
	return context.WithValue(ctx, keyServerDomain, domain)
}

func GetServerDomain(r *http.Request) string {
	if v, ok := r.Context().Value(keyServerDomain).(string); ok {
		return v
	}
	return ""
}

func WithSession(ctx context.Context, sess *models.AuthSession) context.Context {
	return context.WithValue(ctx, keySession, sess)
}

func GetSession(r *http.Request) *models.AuthSession {
	if v, ok := r.Context().Value(keySession).(*models.AuthSession); ok {
		return v
	}
	return nil
}

func GetSessionEmail(r *http.Request) string {
	if s := GetSession(r); s != nil {
		return s.Email
	}
	return ""
}

func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, keyClientIP, ip)
}

func GetClientIP(r *http.Request) string {
	if v, ok := r.Context().Value(keyClientIP).(string); ok {
		return v
	}
	return r.RemoteAddr
}
