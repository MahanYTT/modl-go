package models

import (
	"context"
	"net/http"
)

type Server struct {
	ID                         string `json:"id"`
	ServerName                 string `json:"serverName"`
	CustomDomain               string `json:"customDomain"`
	CustomDomainOverride       string `json:"customDomainOverride"`
	AdminEmail                 string `json:"adminEmail"`
	EmailVerified              bool   `json:"emailVerified"`
	EmailVerificationToken     string `json:"emailVerificationToken"`
	ProvisioningStatus         string `json:"provisioningStatus"`
	ProvisioningNotes          string `json:"provisioningNotes"`
	ProvisioningSignInToken    string `json:"provisioningSignInToken"`
	ProvisioningSignInTokenExp int64  `json:"provisioningSignInTokenExpiresAt"`
	CLISetupToken              string `json:"cliSetupToken"`
	APIKey                     string `json:"apiKey"`
	OnlinePlayerCount          int64  `json:"onlinePlayerCount"`
	UserCount                  int64  `json:"userCount"`
	TicketCount                int64  `json:"ticketCount"`
	LastStatsUpdatedAt         int64  `json:"lastStatsUpdatedAt"`
	LastActivityAt             int64  `json:"lastActivityAt"`
	StaffPermissionsUpdatedAt  int64  `json:"staffPermissionsUpdatedAt"`
	PunishmentTypesUpdatedAt   int64  `json:"punishmentTypesUpdatedAt"`
	CreatedAt                  int64  `json:"createdAt"`
	UpdatedAt                  int64  `json:"updatedAt"`
}

type AuthSession struct {
	ID        string `json:"id"`
	ServerID  string `json:"serverId"`
	Email     string `json:"email"`
	CreatedAt int64  `json:"createdAt"`
	ExpiresAt int64  `json:"expiresAt"`
	IPAddress string `json:"ipAddress"`
	UserAgent string `json:"userAgent"`
}

type Permission struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Category       string `json:"category"`
	Parent         string `json:"parent,omitempty"`
	SuperAdminOnly bool   `json:"superAdminOnly,omitempty"`
}

type StaffRole struct {
	ID          string   `json:"id"`
	ServerID    string   `json:"serverId"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	IsDefault   bool     `json:"isDefault"`
	Order       int      `json:"order"`
	CreatedAt   int64    `json:"createdAt"`
	UpdatedAt   int64    `json:"updatedAt"`
}

type ServerResolver interface {
	GetByDomain(ctx context.Context, domain string) (*Server, error)
	GetByAPIKey(ctx context.Context, apiKey string) (*Server, error)
	GetDefault(ctx context.Context) (*Server, error)
}

type SessionResolver interface {
	FindAndRefresh(ctx context.Context, serverID, token string) (*AuthSession, error)
}

type PermissionChecker interface {
	IsSuperAdmin(srv *Server, email string) bool
	IsAuthorizedEmail(ctx context.Context, srv *Server, email string) bool
	PanelRoleID(ctx context.Context, srv *Server, email string) string
	HasPermission(ctx context.Context, srv *Server, roleID, perm string) bool
	HasAnyWithPrefix(ctx context.Context, srv *Server, roleID, prefix string) bool
}

type AccessRule int

const (
	RuleRequirePermission AccessRule = iota
	RulePermitAll
	RulePlayerAccess
	RulePunishmentTypeAccess
	RuleAppealReply
	RuleSuperAdmin
)

type AccessPolicy struct {
	Rule                  AccessRule
	View                  string
	Modify                string
	SupersedesPermissions []string
}

func IsReadOnlyMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}
