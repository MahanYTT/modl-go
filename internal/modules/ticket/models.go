package ticket

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type TicketCategory string

const (
	CategoryBug         TicketCategory = "bug"
	CategoryPlayer      TicketCategory = "player"
	CategoryChat        TicketCategory = "chat"
	CategoryAppeal      TicketCategory = "appeal"
	CategoryApplication TicketCategory = "application"
	CategorySupport     TicketCategory = "support"
)

type TicketStatus string

const (
	StatusUnfinished TicketStatus = "unfinished"
	StatusOpen       TicketStatus = "open"
	StatusClosed     TicketStatus = "closed"
)

type TicketPriority string

const (
	PriorityLow    TicketPriority = "low"
	PriorityNormal TicketPriority = "normal"
	PriorityHigh   TicketPriority = "high"
)

type AppealWorkflowStatus string

const (
	AppealOpen                    AppealWorkflowStatus = "open"
	AppealUnderReview             AppealWorkflowStatus = "under_review"
	AppealPendingPlayerResponse   AppealWorkflowStatus = "pending_player_response"
	AppealApproved                AppealWorkflowStatus = "approved"
	AppealRejected                AppealWorkflowStatus = "rejected"
)

var (
	nonAlphaNumRe   = regexp.MustCompile(`[^a-z0-9]+`)
	edgeUnderscore  = regexp.MustCompile(`^_+|_+$`)
	uuidDashedRe    = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	uuidUndashedRe  = regexp.MustCompile(`^[0-9a-fA-F]{32}$`)
	mcUsernameRe    = regexp.MustCompile(`^[a-zA-Z0-9_.]{2,16}$`)
)

func NormalizeAliasKey(v string) string {
	s := strings.ToLower(strings.TrimSpace(v))
	s = nonAlphaNumRe.ReplaceAllString(s, "_")
	return edgeUnderscore.ReplaceAllString(s, "")
}

func ParseTicketCategory(v string) (TicketCategory, error) {
	switch NormalizeAliasKey(v) {
	case "bug", "bug_report":
		return CategoryBug, nil
	case "player", "player_report", "report":
		return CategoryPlayer, nil
	case "chat", "chat_report":
		return CategoryChat, nil
	case "appeal", "ban_appeal":
		return CategoryAppeal, nil
	case "application", "staff", "staff_application", "apply":
		return CategoryApplication, nil
	case "support", "general_support":
		return CategorySupport, nil
	default:
		return "", fmt.Errorf("unknown ticket category: %s", v)
	}
}

func (c TicketCategory) DisplayName() string {
	switch c {
	case CategoryBug:
		return "Bug Report"
	case CategoryPlayer:
		return "Player Report"
	case CategoryChat:
		return "Chat Report"
	case CategoryAppeal:
		return "Ban Appeal"
	case CategoryApplication:
		return "Staff Application"
	case CategorySupport:
		return "General Support"
	default:
		return "General Support"
	}
}

func (c TicketCategory) TicketPrefix() string {
	switch c {
	case CategoryBug:
		return "BUG"
	case CategoryPlayer:
		return "PLAYER"
	case CategoryChat:
		return "CHAT"
	case CategoryAppeal:
		return "APPEAL"
	case CategoryApplication:
		return "STAFF"
	case CategorySupport:
		return "SUPPORT"
	default:
		return "TKT"
	}
}

func (c TicketCategory) IsReport() bool {
	return c == CategoryPlayer || c == CategoryChat
}

func (c TicketCategory) IsAppeal() bool {
	return c == CategoryAppeal
}

func ReportCategoryIDs() []string {
	return []string{string(CategoryPlayer), string(CategoryChat)}
}

func IsCanonicalBucket(v string) bool {
	switch v {
	case "bug", "report", "appeal", "support", "staff":
		return true
	default:
		return false
	}
}

func CategoryIDsForBucket(bucket string) []string {
	switch NormalizeAliasKey(bucket) {
	case "bug":
		return []string{string(CategoryBug)}
	case "report":
		return []string{string(CategoryPlayer), string(CategoryChat)}
	case "appeal":
		return []string{string(CategoryAppeal)}
	case "staff":
		return []string{string(CategoryApplication)}
	case "support":
		return []string{string(CategorySupport)}
	default:
		return nil
	}
}

func ParseTicketStatus(v string) (TicketStatus, error) {
	switch NormalizeAliasKey(v) {
	case "unfinished", "draft":
		return StatusUnfinished, nil
	case "open", "new", "active", "pending", "in_progress", "inprogress":
		return StatusOpen, nil
	case "closed", "resolved", "complete", "completed", "done":
		return StatusClosed, nil
	default:
		return "", fmt.Errorf("unknown ticket status: %s", v)
	}
}

func (s TicketStatus) IsTerminal() bool {
	return s == StatusClosed
}

func ParseTicketPriority(v string) (TicketPriority, error) {
	if strings.TrimSpace(v) == "" {
		return PriorityNormal, nil
	}
	switch NormalizeAliasKey(v) {
	case "low", "minor":
		return PriorityLow, nil
	case "normal", "medium", "default", "standard":
		return PriorityNormal, nil
	case "high", "urgent", "critical", "highest":
		return PriorityHigh, nil
	default:
		return "", fmt.Errorf("unknown ticket priority: %s", v)
	}
}

func ParseAppealWorkflowStatus(v string) (AppealWorkflowStatus, error) {
	switch NormalizeAliasKey(v) {
	case "open":
		return AppealOpen, nil
	case "under_review", "underreview":
		return AppealUnderReview, nil
	case "pending_player_response", "pendingplayerresponse":
		return AppealPendingPlayerResponse, nil
	case "approved", "approve", "accepted", "accept":
		return AppealApproved, nil
	case "rejected", "reject", "dismiss", "dismissed", "denied", "deny":
		return AppealRejected, nil
	default:
		return "", fmt.Errorf("unknown appeal workflow status: %s", v)
	}
}

func (a AppealWorkflowStatus) DisplayName() string {
	switch a {
	case AppealOpen:
		return "Open"
	case AppealUnderReview:
		return "Under Review"
	case AppealPendingPlayerResponse:
		return "Pending Player Response"
	case AppealApproved:
		return "Approved"
	case AppealRejected:
		return "Rejected"
	default:
		return string(a)
	}
}

func (a AppealWorkflowStatus) IsTerminal() bool {
	return a == AppealApproved || a == AppealRejected
}

func (a AppealWorkflowStatus) ToTicketStatus() TicketStatus {
	if a.IsTerminal() {
		return StatusClosed
	}
	return StatusOpen
}

func NormalizeUUID(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func DashedUUID(raw string) (string, bool) {
	c := strings.TrimSpace(raw)
	if uuidUndashedRe.MatchString(c) {
		c = c[0:8] + "-" + c[8:12] + "-" + c[12:16] + "-" + c[16:20] + "-" + c[20:32]
	}
	if !uuidDashedRe.MatchString(c) {
		return "", false
	}
	return strings.ToLower(c), true
}

func IsValidUUID(raw string) bool {
	return uuidDashedRe.MatchString(strings.TrimSpace(raw))
}

func IsValidMinecraftUsername(raw string) bool {
	return mcUsernameRe.MatchString(strings.TrimSpace(raw))
}

func NormalizeSingleAssignee(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	return strings.ToLower(trimmed)
}

func NormalizeAssignees(raw []string) []string {
	if len(raw) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(raw))
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		n := NormalizeSingleAssignee(item)
		if n == "" {
			continue
		}
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	return out
}

func ResolveAssignedTo(assignedTo []string, assignedToLegacy string) []string {
	if len(assignedTo) > 0 {
		return NormalizeAssignees(assignedTo)
	}
	if strings.TrimSpace(assignedToLegacy) != "" {
		return NormalizeAssignees([]string{assignedToLegacy})
	}
	return []string{}
}

type TicketReply struct {
	ID                string           `json:"id"`
	Name              string           `json:"name"`
	Content           string           `json:"content"`
	Type              string           `json:"type"`
	CreatedMs         int64            `json:"created"`
	Staff             bool             `json:"staff"`
	Action            string           `json:"action,omitempty"`
	Avatar            string           `json:"avatar,omitempty"`
	Attachments       []map[string]any `json:"attachments,omitempty"`
	CreatorIdentifier string           `json:"creatorIdentifier,omitempty"`
}

type TicketNote struct {
	ID         string `json:"id"`
	Text       string `json:"text"`
	IssuerName string `json:"issuerName"`
	Avatar     string `json:"avatar,omitempty"`
	DateMs     int64  `json:"date"`
}

type ChatMessage struct {
	Content     string `json:"content"`
	TimestampMs int64  `json:"timestamp"`
	Sender      string `json:"sender,omitempty"`
}

type SuggestedAction struct {
	PunishmentTypeID int32  `json:"punishmentTypeId"`
	Severity         string `json:"severity,omitempty"`
}

type AIAnalysisResult struct {
	Analysis                string           `json:"analysis,omitempty"`
	Confidence              *float64         `json:"confidence,omitempty"`
	AnalyzedAtMs            int64            `json:"analyzedAt,omitempty"`
	RawResponse             string           `json:"rawResponse,omitempty"`
	SuggestedAction         *SuggestedAction `json:"suggestedAction,omitempty"`
	WasAppliedAutomatically bool             `json:"wasAppliedAutomatically"`
	Dismissed               bool             `json:"dismissed"`
}

type Ticket struct {
	ID                   string            `json:"id"`
	ServerID             string            `json:"serverId"`
	Type                 TicketCategory    `json:"type"`
	Subject              string            `json:"subject"`
	Status               TicketStatus      `json:"status"`
	CreatorUUID          string            `json:"creatorUuid"`
	CreatorName          string            `json:"creatorName"`
	CreatorAvatar        string            `json:"creatorAvatar"`
	ReportedPlayer       string            `json:"reportedPlayer"`
	ReportedPlayerUUID   string            `json:"reportedPlayerUuid"`
	PunishmentID         string            `json:"punishmentId"`
	Tags                 []string          `json:"tags"`
	Replies              []TicketReply     `json:"replies"`
	Notes                []TicketNote      `json:"notes"`
	ChatMessages         []ChatMessage     `json:"chatMessages"`
	FormData             map[string]any    `json:"formData"`
	Data                 map[string]any    `json:"data"`
	Locked               bool              `json:"locked"`
	Priority             TicketPriority    `json:"priority"`
	AssignedTo           []string          `json:"assignedTo"`
	AppealWorkflowStatus string            `json:"appealWorkflowStatus"`
	AIAnalysis           *AIAnalysisResult `json:"aiAnalysis,omitempty"`
	EmailAuthEnabled     bool              `json:"emailAuthEnabled"`
	Hidden               bool              `json:"hidden"`
	CreatedAtMs          int64             `json:"created"`
	UpdatedAtMs          int64             `json:"updatedAt"`
}

func (t *Ticket) ApplyLifecycleStatus(status TicketStatus) {
	t.Status = status
	t.Locked = status.IsTerminal()
}

type TicketSearchFilter struct {
	Search    string
	Status    string
	Types     []string
	Author    string
	Labels    []string
	Assignees []string
}

type TicketCounts struct {
	Open   int64
	Closed int64
}

type PlayerTicketResponse struct {
	ID                 string    `json:"id"`
	Subject            string    `json:"subject"`
	Status             string    `json:"status"`
	Type               string    `json:"type"`
	Created            time.Time `json:"created"`
	CreatorName        string    `json:"creatorName"`
	CreatorUUID        string    `json:"creatorUuid"`
	ReportedPlayer     string    `json:"reportedPlayer"`
	ReportedPlayerUUID string    `json:"reportedPlayerUuid"`
	Locked             bool      `json:"locked"`
	Tags               []string  `json:"tags"`
	AssignedTo         []string  `json:"assignedTo"`
}

type StaffTicketSubscription struct {
	TicketID     string `json:"ticketId"`
	SubscribedAt int64  `json:"subscribedAt"`
	LastReadAt   int64  `json:"lastReadAt"`
	Active       bool   `json:"active"`
}

type SubscriptionUpdate struct {
	ID         string `json:"id"`
	ServerID   string `json:"serverId"`
	StaffEmail string `json:"staffEmail"`
	TicketID   string `json:"ticketId"`
	EventType  string `json:"eventType"`
	Summary    string `json:"summary"`
	Actor      string `json:"actor"`
	Read       bool   `json:"read"`
	CreatedAt  int64  `json:"createdAt"`
}

type TicketVerification struct {
	ID             string
	ServerID       string
	TicketID       string
	Token          string
	CodeHash       string
	Email          string
	FailedAttempts int
	ExpiresAt      int64
}

// Minecraft V1 JSON DTOs (1:1 with upstream MinecraftV1Response)

type MinecraftCreateTicketRequestJSON struct {
	CreatorUUID        string   `json:"creatorUuid"`
	CreatorName        string   `json:"creatorName"`
	Type               string   `json:"type"`
	Subject            string   `json:"subject"`
	Description        string   `json:"description"`
	ReportedPlayerUUID string   `json:"reportedPlayerUuid"`
	ReportedPlayerName string   `json:"reportedPlayerName"`
	ChatMessages       []string `json:"chatMessages"`
	Tags               []string `json:"tags"`
	Priority           string   `json:"priority"`
	CreatedServer      string   `json:"createdServer"`
	ReplayURL          string   `json:"replayUrl"`
}

type MinecraftClaimTicketRequestJSON struct {
	PlayerUUID string `json:"playerUuid"`
	PlayerName string `json:"playerName"`
}

type AssignReportRequestJSON struct {
	Assignee string `json:"assignee"`
}

type DismissReportRequestJSON struct {
	DismissedBy string `json:"dismissedBy"`
	Reason      string `json:"reason"`
}

type ResolveReportRequestJSON struct {
	ResolvedBy   string `json:"resolvedBy"`
	Resolution   string `json:"resolution"`
	PunishmentID string `json:"punishmentId"`
}

type MinecraftTicketCreatedJSON struct {
	Status   int    `json:"status"`
	Success  bool   `json:"success"`
	TicketID string `json:"ticketId"`
	Message  string `json:"message"`
}

type MinecraftTicketListItemJSON struct {
	ID               string    `json:"id"`
	Type             string    `json:"type"`
	Category         string    `json:"category"`
	Subject          string    `json:"subject"`
	Status           string    `json:"status"`
	PlayerName       string    `json:"playerName"`
	PlayerUUID       string    `json:"playerUuid"`
	Priority         string    `json:"priority"`
	AssignedTo       []string  `json:"assignedTo"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	HasStaffResponse bool      `json:"hasStaffResponse"`
	ReplyCount       int       `json:"replyCount"`
	Locked           bool      `json:"locked"`
}

type MinecraftTicketListJSON struct {
	Status  int                           `json:"status"`
	Tickets []MinecraftTicketListItemJSON `json:"tickets"`
}

type MinecraftPlayerTicketJSON struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Category  string    `json:"category"`
	Subject   string    `json:"subject"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type MinecraftPlayerTicketListJSON struct {
	Status  int                         `json:"status"`
	Tickets []MinecraftPlayerTicketJSON `json:"tickets"`
}

type MinecraftTicketLookupJSON struct {
	ID                string    `json:"id"`
	Type              string    `json:"type"`
	Category          string    `json:"category"`
	Subject           string    `json:"subject"`
	Status            string    `json:"status"`
	PlayerName        string    `json:"playerName"`
	PlayerUUID        string    `json:"playerUuid"`
	CreatedAt         time.Time `json:"createdAt"`
	FirstReplyContent *string   `json:"firstReplyContent,omitempty"`
}

type MinecraftTicketLookupListJSON struct {
	Status  int                         `json:"status"`
	Tickets []MinecraftTicketLookupJSON `json:"tickets"`
}

type MinecraftTicketDetailReplyJSON struct {
	ID         string    `json:"id"`
	Content    string    `json:"content"`
	AuthorName string    `json:"authorName"`
	AuthorID   string    `json:"authorId"`
	IsStaff    bool      `json:"isStaff"`
	CreatedAt  time.Time `json:"createdAt"`
}

type MinecraftChatMessageJSON struct {
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Sender    string    `json:"sender,omitempty"`
}

type MinecraftTicketDetailViewJSON struct {
	ID           string                           `json:"id"`
	Type         string                           `json:"type"`
	Category     string                           `json:"category"`
	Subject      string                           `json:"subject"`
	Status       string                           `json:"status"`
	PlayerName   string                           `json:"playerName"`
	PlayerUUID   string                           `json:"playerUuid"`
	Priority     string                           `json:"priority"`
	AssignedTo   []string                         `json:"assignedTo"`
	CreatedAt    time.Time                        `json:"createdAt"`
	UpdatedAt    time.Time                        `json:"updatedAt"`
	Locked       bool                             `json:"locked"`
	Replies      []MinecraftTicketDetailReplyJSON `json:"replies"`
	ChatMessages []MinecraftChatMessageJSON       `json:"chatMessages,omitempty"`
	ReplayURL    *string                          `json:"replayUrl,omitempty"`
}

type MinecraftTicketDetailJSON struct {
	Status int                           `json:"status"`
	Ticket MinecraftTicketDetailViewJSON `json:"ticket"`
}

type MinecraftTicketClaimJSON struct {
	Status   int    `json:"status"`
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	TicketID string `json:"ticketId"`
	Subject  string `json:"subject"`
}

type MinecraftReportViewJSON struct {
	ID                 string                     `json:"id"`
	Type               string                     `json:"type"`
	ReporterName       string                     `json:"reporterName"`
	ReporterUUID       string                     `json:"reporterUuid"`
	ReportedPlayerUUID string                     `json:"reportedPlayerUuid"`
	ReportedPlayerName string                     `json:"reportedPlayerName"`
	Subject            string                     `json:"subject"`
	Content            string                     `json:"content"`
	Status             string                     `json:"status"`
	Priority           string                     `json:"priority"`
	CreatedAt          time.Time                  `json:"createdAt"`
	AssignedTo         []string                   `json:"assignedTo"`
	ChatMessages       []MinecraftChatMessageJSON `json:"chatMessages,omitempty"`
	ReplayURL          *string                    `json:"replayUrl,omitempty"`
}

type MinecraftReportListJSON struct {
	Status  int                       `json:"status"`
	Reports []MinecraftReportViewJSON `json:"reports"`
}

type MinecraftReportOperationJSON struct {
	Status  int    `json:"status"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type MinecraftNotFoundJSON struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}
