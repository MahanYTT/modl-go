package player

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	LinkedBanTypeOrdinal = 4
	SystemIssuer         = "System"
	MaxReasonLength      = 1000
	MaxNoteLength        = 4000
	MaxChatBatchEntries  = 500
	MaxSearchQueryLen    = 200
)

var (
	uuidDashedRe    = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	uuidUndashedRe  = regexp.MustCompile(`^[0-9a-fA-F]{32}$`)
	mcUsernameRe    = regexp.MustCompile(`^[a-zA-Z0-9_.]{1,16}$`)
	mcLogUsernameRe = regexp.MustCompile(`^[a-zA-Z0-9_.]{2,16}$`)
)

// Player data keys
const (
	PlayerKeyIsOnline             = "isOnline"
	PlayerKeyLastLogin            = "lastLogin"
	PlayerKeyLastServer           = "lastServer"
	PlayerKeyTotalPlaytimeSeconds = "totalPlaytimeSeconds"
	PlayerKeyFirstJoin            = "firstJoin"
	PlayerKeyLastSkinHash         = "lastSkinHash"
	PlayerKeyLinkedAccounts       = "linkedAccounts"
	PlayerKeyLastLinkedUpdate     = "lastLinkedUpdate"
	PlayerKeyPendingNotifications = "pendingNotifications"
	PlayerKeyCountry              = "country"
	PlayerKeyRegion               = "region"
	PlayerKeyASN                  = "asn"
	PlayerKeyProxy                = "proxy"
	PlayerKeyHosting              = "hosting"
)

// Punishment data keys
const (
	PunishmentKeyStatus              = "status"
	PunishmentKeyDuration            = "duration"
	PunishmentKeyReason              = "reason"
	PunishmentKeySeverity            = "severity"
	PunishmentKeyOffenseLevel        = "offenseLevel"
	PunishmentKeyEnforcementCategory = "enforcementCategory"
	PunishmentKeyAltBlocking         = "altBlocking"
	PunishmentKeyWipeAfterExpiry     = "wipeAfterExpiry"
	PunishmentKeyStatWipeCompleted   = "statWipeCompleted"
	PunishmentKeyPendingAck          = "pendingAcknowledgement"
	PunishmentKeyLinkedBanID         = "linkedBanId"
	PunishmentKeyLinkedBanParentUUID = "linkedBanParentUuid"
	PunishmentKeyBlockedName         = "blockedName"
	PunishmentKeyBlockedSkin         = "blockedSkin"
	PunishmentKeyAppealOutcome       = "appealOutcome"
	PunishmentKeyAppealTicketID      = "appealTicketId"
	PunishmentKeyAppealedTicketID    = "appealedTicketId"
	PunishmentKeySilent              = "silent"
	PunishmentKeyBanLinkedAccounts   = "banLinkedAccounts"
	PunishmentKeyKickSameIP          = "kickSameIP"
	PunishmentKeyAIGenerated         = "aiGenerated"
)

// Punishment statuses
const (
	PunishmentStatusUnstarted = "unstarted"
	PunishmentStatusPardoned  = "pardoned"
)

// Enforcement categories
const (
	CategoryBan  = "BAN"
	CategoryMute = "MUTE"
	CategoryKick = "KICK"
)

// Modification types
const (
	ModManualPardon         = "MANUAL_PARDON"
	ModSystemPardon         = "SYSTEM_PARDON"
	ModAppealPardon         = "APPEAL_PARDON"
	ModManualDurationChange = "MANUAL_DURATION_CHANGE"
	ModAppealDurationChange = "APPEAL_DURATION_CHANGE"
)

func IsPardonModType(t string) bool {
	switch strings.ToUpper(strings.TrimSpace(t)) {
	case ModManualPardon, ModSystemPardon, ModAppealPardon, "PARDON":
		return true
	default:
		return false
	}
}

func IsDurationChangeModType(t string) bool {
	switch strings.ToUpper(strings.TrimSpace(t)) {
	case ModManualDurationChange, ModAppealDurationChange, "DURATION_CHANGE":
		return true
	default:
		return false
	}
}

func AppealOutcomeForModType(t string) string {
	switch strings.ToUpper(strings.TrimSpace(t)) {
	case ModAppealPardon, ModAppealDurationChange:
		return "approved"
	default:
		return ""
	}
}

func NormalizeUUID(raw string) (string, bool) {
	s := strings.TrimSpace(raw)
	if uuidDashedRe.MatchString(s) {
		return strings.ToLower(s), true
	}
	if uuidUndashedRe.MatchString(s) {
		s = strings.ToLower(s)
		return fmt.Sprintf("%s-%s-%s-%s-%s", s[0:8], s[8:12], s[12:16], s[16:20], s[20:32]), true
	}
	return "", false
}

func UndashedUUID(raw string) string {
	n, ok := NormalizeUUID(raw)
	if !ok {
		return ""
	}
	return strings.ReplaceAll(n, "-", "")
}

func NewShortID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func ValidateSafeURL(raw string) bool {
	s := strings.TrimSpace(raw)
	if s == "" {
		return true
	}
	if strings.HasPrefix(s, "/") && !strings.HasPrefix(s, "//") {
		return true
	}
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	if host == "" {
		return false
	}
	if host == "localhost" || strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsUnspecified() || ip.IsMulticast() {
			return false
		}
	}
	return true
}

func NormalizeSeverity(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "low", "lenient", "minor":
		return "low"
	case "severe", "aggravated", "high", "major":
		return "severe"
	default:
		return "regular"
	}
}

func OffenseDisplayStatus(offenseLevel string) string {
	if strings.EqualFold(offenseLevel, "first") {
		return "low"
	}
	return offenseLevel
}

func FormatDuration(durationMs int64, isPermanent bool) string {
	if isPermanent || durationMs < 0 {
		return "Permanent"
	}
	if durationMs == 0 {
		return "Instant"
	}
	totalSeconds := durationMs / 1000
	days := totalSeconds / 86400
	hours := (totalSeconds % 86400) / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if seconds > 0 && days == 0 {
		parts = append(parts, fmt.Sprintf("%ds", seconds))
	}
	if len(parts) == 0 {
		return "0s"
	}
	return strings.Join(parts, " ")
}

func IsAutoGeneratedNote(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	return lower == "" ||
		lower == "issued punishment" ||
		strings.HasPrefix(lower, "issued ") ||
		lower == "pardoned punishment" ||
		lower == "added evidence" ||
		strings.HasPrefix(lower, "changed duration to ") ||
		strings.HasPrefix(lower, "enabled ") ||
		strings.HasPrefix(lower, "disabled ")
}

// --- Domain Models ---

type UsernameEntry struct {
	Username string `json:"username"`
	DateMs   int64  `json:"date"`
}

type IPEntry struct {
	IPAddress    string  `json:"ipAddress"`
	Country      *string `json:"country,omitempty"`
	Region       *string `json:"region,omitempty"`
	ASN          *string `json:"asn,omitempty"`
	Proxy        bool    `json:"proxy"`
	Hosting      bool    `json:"hosting"`
	FirstLoginMs int64   `json:"firstLogin"`
	LoginsMs     []int64 `json:"logins"`
}

type NoteEntry struct {
	ID         string `json:"id"`
	Text       string `json:"text"`
	DateMs     int64  `json:"date"`
	IssuerName string `json:"issuerName,omitempty"`
	IssuerID   string `json:"issuerId,omitempty"`
}

type PunishmentModification struct {
	ID                string         `json:"id"`
	Type              string         `json:"type"`
	DateMs            int64          `json:"date"`
	IssuerName        string         `json:"issuerName,omitempty"`
	IssuerID          string         `json:"issuerId,omitempty"`
	Reason            string         `json:"reason"`
	EffectiveDuration *int64         `json:"effectiveDuration,omitempty"`
	AppealTicketID    string         `json:"appealTicketId,omitempty"`
	Data              map[string]any `json:"data,omitempty"`
}

type PunishmentNote struct {
	ID         string `json:"id"`
	Text       string `json:"text"`
	DateMs     int64  `json:"date"`
	IssuerName string `json:"issuerName,omitempty"`
	IssuerID   string `json:"issuerId,omitempty"`
}

type PunishmentEvidence struct {
	Text         *string `json:"text,omitempty"`
	URL          *string `json:"url,omitempty"`
	Type         string  `json:"type"`
	UploadedBy   *string `json:"uploadedBy,omitempty"`
	UploadedByID *string `json:"uploadedById,omitempty"`
	UploadedAtMs int64   `json:"uploadedAt"`
	FileName     *string `json:"fileName,omitempty"`
	FileType     *string `json:"fileType,omitempty"`
	FileSize     *int64  `json:"fileSize,omitempty"`
}

type Punishment struct {
	ID                string
	ServerID          string
	PlayerID          string
	PlayerUUID        string
	PlayerUsername    string
	Type              string
	TypeOrdinal       int
	IssuerName        string
	IssuerID          string
	IssuedMs          int64
	StartedMs         *int64
	ExpiresMs         *int64
	DurationMs        *int64
	Pardoned          bool
	PardonedBy        string
	PardonedAtMs      *int64
	PardonReason      string
	Points            int
	IPBan             bool
	IPAddress         string
	Silent            bool
	WipeStats         bool
	BlockAlts         bool
	Modifications     []PunishmentModification
	Notes             []PunishmentNote
	Evidence          []PunishmentEvidence
	AttachedTicketIds []string
	Data              map[string]any
}

func (p *Punishment) IsPardoned() bool {
	if p.Pardoned {
		return true
	}
	if s, _ := p.Data[PunishmentKeyStatus].(string); strings.EqualFold(s, PunishmentStatusPardoned) {
		return true
	}
	for _, m := range p.Modifications {
		if IsPardonModType(m.Type) {
			return true
		}
	}
	return false
}

func (p *Punishment) IsUnstarted() bool {
	s, _ := p.Data[PunishmentKeyStatus].(string)
	if !strings.EqualFold(s, PunishmentStatusUnstarted) {
		return false
	}
	return !p.IsPardoned()
}

func (p *Punishment) DataString(key string) string {
	if p.Data == nil {
		return ""
	}
	v, ok := p.Data[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

func (p *Punishment) DataBool(key string) bool {
	if p.Data == nil {
		return false
	}
	v, ok := p.Data[key]
	if !ok || v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return strings.EqualFold(val, "true")
	case float64:
		return val != 0
	case int:
		return val != 0
	case int64:
		return val != 0
	}
	return false
}

func (p *Punishment) DataInt64(key string) (int64, bool) {
	if p.Data == nil {
		return 0, false
	}
	v, ok := p.Data[key]
	if !ok || v == nil {
		return 0, false
	}
	switch val := v.(type) {
	case int64:
		return val, true
	case int:
		return int64(val), true
	case float64:
		return int64(val), true
	}
	return 0, false
}

type Player struct {
	ID              string
	ServerID        string
	MinecraftUUID   string
	CurrentUsername string
	Usernames       []UsernameEntry
	Notes           []NoteEntry
	IPAddresses     []IPEntry
	Punishments     []*Punishment
	Data            map[string]any
	FirstSeenMs     int64
	LastSeenMs      int64
}

func (pl *Player) LatestUsername() string {
	if len(pl.Usernames) > 0 {
		latest := pl.Usernames[len(pl.Usernames)-1].Username
		if strings.TrimSpace(latest) != "" {
			return latest
		}
	}
	if strings.TrimSpace(pl.CurrentUsername) != "" {
		return pl.CurrentUsername
	}
	return "Unknown"
}

func (pl *Player) IsOnline() bool {
	if pl.Data == nil {
		return false
	}
	b, _ := pl.Data[PlayerKeyIsOnline].(bool)
	return b
}

func (pl *Player) LastLoginMs() *int64 {
	if pl.Data == nil {
		return nil
	}
	return parseFlexibleTimePtr(pl.Data[PlayerKeyLastLogin])
}

func (pl *Player) LastServer() string {
	if pl.Data == nil {
		return ""
	}
	s, _ := pl.Data[PlayerKeyLastServer].(string)
	return s
}

func (pl *Player) TotalPlaytimeSeconds() int64 {
	if pl.Data == nil {
		return 0
	}
	switch v := pl.Data[PlayerKeyTotalPlaytimeSeconds].(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	}
	return 0
}

func (pl *Player) LastSkinHash() string {
	if pl.Data == nil {
		return ""
	}
	s, _ := pl.Data[PlayerKeyLastSkinHash].(string)
	return s
}

func (pl *Player) LinkedAccountUUIDs() []string {
	if pl.Data == nil {
		return nil
	}
	raw, ok := pl.Data[PlayerKeyLinkedAccounts]
	if !ok || raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func (pl *Player) LastLinkedUpdateMs() *int64 {
	if pl.Data == nil {
		return nil
	}
	return parseFlexibleTimePtr(pl.Data[PlayerKeyLastLinkedUpdate])
}

func (pl *Player) LegacyPendingNotifications() []map[string]any {
	if pl.Data == nil {
		return nil
	}
	raw, ok := pl.Data[PlayerKeyPendingNotifications]
	if !ok || raw == nil {
		return nil
	}
	arr, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]map[string]any, 0, len(arr))
	for _, item := range arr {
		if m, ok := item.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

// --- Settings Models (compatible with internal/modules/settings & upstream Java) ---

type DurationDetail struct {
	Value int64  `json:"value"`
	Unit  string `json:"unit"`
	Type  string `json:"type"`
}

func (d *DurationDetail) IsPermanent() bool {
	if d == nil {
		return false
	}
	t := strings.ToLower(strings.TrimSpace(d.Type))
	u := strings.ToLower(strings.TrimSpace(d.Unit))
	return strings.Contains(t, "permanent") || u == "permanent"
}

func (d *DurationDetail) IsBan() bool {
	if d == nil {
		return false
	}
	return strings.Contains(strings.ToLower(d.Type), "ban")
}

func (d *DurationDetail) IsMute() bool {
	if d == nil {
		return false
	}
	return strings.Contains(strings.ToLower(d.Type), "mute")
}

func (d *DurationDetail) IsKick() bool {
	if d == nil {
		return false
	}
	return strings.Contains(strings.ToLower(d.Type), "kick")
}

func (d *DurationDetail) ToMilliseconds() int64 {
	if d == nil {
		return 0
	}
	if d.IsPermanent() {
		return -1
	}
	u := strings.ToLower(strings.TrimSpace(d.Unit))
	switch u {
	case "seconds", "second":
		return d.Value * 1000
	case "minutes", "minute":
		return d.Value * 60 * 1000
	case "hours", "hour":
		return d.Value * 60 * 60 * 1000
	case "days", "day":
		return d.Value * 24 * 60 * 60 * 1000
	case "weeks", "week":
		return d.Value * 7 * 24 * 60 * 60 * 1000
	case "months", "month":
		return d.Value * 30 * 24 * 60 * 60 * 1000
	default:
		return -1
	}
}

type OffenseLevelDurations struct {
	First    *DurationDetail `json:"first,omitempty"`
	Medium   *DurationDetail `json:"medium,omitempty"`
	Habitual *DurationDetail `json:"habitual,omitempty"`
}

func (o *OffenseLevelDurations) Get(level string) *DurationDetail {
	if o == nil {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "first", "low":
		return o.First
	case "medium":
		return o.Medium
	case "habitual", "severe":
		return o.Habitual
	default:
		return o.First
	}
}

type PunishmentDurations struct {
	Low     *OffenseLevelDurations `json:"low,omitempty"`
	Regular *OffenseLevelDurations `json:"regular,omitempty"`
	Severe  *OffenseLevelDurations `json:"severe,omitempty"`
}

func (d *PunishmentDurations) Get(severity, offenseLevel string) *DurationDetail {
	if d == nil {
		return nil
	}
	switch NormalizeSeverity(severity) {
	case "low":
		return d.Low.Get(offenseLevel)
	case "severe":
		return d.Severe.Get(offenseLevel)
	default:
		return d.Regular.Get(offenseLevel)
	}
}

type PunishmentPoints struct {
	Low     int `json:"low"`
	Regular int `json:"regular"`
	Severe  int `json:"severe"`
}

type PunishmentType struct {
	ID                           int                    `json:"id"`
	Name                         string                 `json:"name"`
	Category                     string                 `json:"category"`
	Customizable                 bool                   `json:"customizable"`
	Ordinal                      int                    `json:"ordinal"`
	Durations                    *PunishmentDurations   `json:"durations,omitempty"`
	SingleSeverityPunishment     bool                   `json:"singleSeverityPunishment"`
	SingleSeverityDurations      *OffenseLevelDurations `json:"singleSeverityDurations,omitempty"`
	SingleSeverityPoints         *int                   `json:"singleSeverityPoints,omitempty"`
	Points                       *PunishmentPoints      `json:"points,omitempty"`
	CustomPoints                 *int                   `json:"customPoints,omitempty"`
	StaffDescription             string                 `json:"staffDescription"`
	PlayerDescription            string                 `json:"playerDescription"`
	CanBeAltBlocking             bool                   `json:"canBeAltBlocking"`
	CanBeStatWiping              bool                   `json:"canBeStatWiping"`
	Appealable                   bool                   `json:"appealable"`
	AppealForm                   map[string]any         `json:"appealForm,omitempty"`
	PermanentUntilSkinChange     bool                   `json:"permanentUntilSkinChange"`
	PermanentUntilUsernameChange bool                   `json:"permanentUntilUsernameChange"`
}

func (pt *PunishmentType) IsSocial() bool {
	return strings.EqualFold(pt.Category, "social")
}

func (pt *PunishmentType) IsGameplay() bool {
	return strings.EqualFold(pt.Category, "gameplay")
}

func (pt *PunishmentType) IsKick() bool {
	n := strings.ToLower(pt.Name)
	c := strings.ToLower(pt.Category)
	if strings.Contains(n, "kick") || strings.Contains(c, "kick") {
		return true
	}
	return pt.Ordinal == 0 && !strings.Contains(n, "ban") && !strings.Contains(n, "mute")
}

func (pt *PunishmentType) IsMute() bool {
	n := strings.ToLower(pt.Name)
	c := strings.ToLower(pt.Category)
	return strings.Contains(n, "mute") || strings.Contains(c, "mute")
}

func (pt *PunishmentType) IsBan() bool {
	n := strings.ToLower(pt.Name)
	c := strings.ToLower(pt.Category)
	return strings.Contains(n, "ban") || strings.Contains(n, "blacklist") || strings.Contains(c, "ban")
}

func (pt *PunishmentType) PointsForSeverity(severity string) int {
	if pt.CustomPoints != nil {
		return *pt.CustomPoints
	}
	if pt.SingleSeverityPoints != nil {
		return *pt.SingleSeverityPoints
	}
	if pt.Points == nil {
		return 0
	}
	switch NormalizeSeverity(severity) {
	case "low":
		return pt.Points.Low
	case "severe":
		return pt.Points.Severe
	default:
		return pt.Points.Regular
	}
}

func (pt *PunishmentType) GetDurationDetail(severity, offenseLevel string) *DurationDetail {
	if pt.SingleSeverityPunishment && pt.SingleSeverityDurations != nil {
		return pt.SingleSeverityDurations.Get(offenseLevel)
	}
	if pt.Durations != nil {
		return pt.Durations.Get(severity, offenseLevel)
	}
	return nil
}

type CategoryThresholds struct {
	Medium            int `json:"medium"`
	Habitual          int `json:"habitual"`
	PointExpiryMonths int `json:"pointExpiryMonths"`
}

func (ct CategoryThresholds) PointExpiryMs() int64 {
	m := ct.PointExpiryMonths
	if m <= 0 {
		m = 24
	}
	return int64(m) * 30 * 24 * 60 * 60 * 1000
}

func (ct CategoryThresholds) OffenderLevel(points int) string {
	if points >= ct.Habitual {
		return "Habitual"
	}
	if points >= ct.Medium {
		return "Medium"
	}
	return "Low"
}

func (ct CategoryThresholds) OffenseLevelInternal(points int) string {
	if points >= ct.Habitual {
		return "habitual"
	}
	if points >= ct.Medium {
		return "medium"
	}
	return "first"
}

type OffenderThresholdSettings struct {
	Social   CategoryThresholds `json:"social"`
	Gameplay CategoryThresholds `json:"gameplay"`
}

func DefaultThresholdSettings() OffenderThresholdSettings {
	return OffenderThresholdSettings{
		Social:   CategoryThresholds{Medium: 10, Habitual: 25, PointExpiryMonths: 6},
		Gameplay: CategoryThresholds{Medium: 15, Habitual: 35, PointExpiryMonths: 12},
	}
}

func DefaultPunishmentTypes() []PunishmentType {
	hrs := func(v int64) *DurationDetail { return &DurationDetail{Value: v, Unit: "HOURS", Type: "TEMPORARY"} }
	days := func(v int64) *DurationDetail { return &DurationDetail{Value: v, Unit: "DAYS", Type: "TEMPORARY"} }
	perm := func() *DurationDetail { return &DurationDetail{Value: 0, Unit: "PERMANENT", Type: "PERMANENT"} }
	return []PunishmentType{
		{
			ID: 0, Ordinal: 0, Name: "Kick", Category: "GAMEPLAY", Customizable: false,
			StaffDescription:  "Kick a player from the server",
			PlayerDescription: "You were kicked from the server",
		},
		{
			ID: 1, Ordinal: 1, Name: "Warn", Category: "SOCIAL", Customizable: true,
			StaffDescription:  "Issue an official warning",
			PlayerDescription: "You received a warning",
			Points:            &PunishmentPoints{Low: 1, Regular: 2, Severe: 5},
		},
		{
			ID: 2, Ordinal: 2, Name: "Mute", Category: "SOCIAL", Customizable: true, Appealable: true, CanBeAltBlocking: true,
			StaffDescription:  "Mute a player in chat",
			PlayerDescription: "You are muted in chat",
			Points:            &PunishmentPoints{Low: 2, Regular: 5, Severe: 10},
			Durations: &PunishmentDurations{
				Low:     &OffenseLevelDurations{First: hrs(1), Medium: hrs(6), Habitual: days(1)},
				Regular: &OffenseLevelDurations{First: days(1), Medium: days(3), Habitual: days(7)},
				Severe:  &OffenseLevelDurations{First: days(7), Medium: days(30), Habitual: perm()},
			},
		},
		{
			ID: 3, Ordinal: 3, Name: "Ban", Category: "GAMEPLAY", Customizable: true, Appealable: true, CanBeAltBlocking: true, CanBeStatWiping: true,
			StaffDescription:  "Ban a player from the server",
			PlayerDescription: "You are banned from the server",
			Points:            &PunishmentPoints{Low: 5, Regular: 10, Severe: 20},
			Durations: &PunishmentDurations{
				Low:     &OffenseLevelDurations{First: days(1), Medium: days(3), Habitual: days(7)},
				Regular: &OffenseLevelDurations{First: days(7), Medium: days(30), Habitual: days(90)},
				Severe:  &OffenseLevelDurations{First: days(30), Medium: perm(), Habitual: perm()},
			},
		},
		{
			ID: 4, Ordinal: 4, Name: "Blacklist", Category: "GAMEPLAY", Customizable: false, CanBeAltBlocking: true, CanBeStatWiping: true,
			StaffDescription:  "Permanently blacklist a player and alts",
			PlayerDescription: "You are blacklisted from the server",
			SingleSeverityPunishment: true,
			SingleSeverityDurations:  &OffenseLevelDurations{First: perm(), Medium: perm(), Habitual: perm()},
		},
	}
}

func intPtr(v int) *int { return &v }

// --- View / DTO Types (for V1/V2 JSON & service responses) ---

type SimplePunishmentModView struct {
	Type              string `json:"type"`
	Timestamp         *int64 `json:"timestamp"`
	EffectiveDuration int64  `json:"effectiveDuration"`
	IssuerName        string `json:"issuerName"`
}

type SimplePunishmentView struct {
	ID                string                    `json:"id"`
	Type              string                    `json:"type"`
	Category          string                    `json:"category"`
	TypeOrdinal       int                       `json:"typeOrdinal"`
	Ordinal           int                       `json:"ordinal"`
	Started           bool                      `json:"started"`
	Expiration        *int64                    `json:"expiration"`
	Description       string                    `json:"description"`
	IssuerName        string                    `json:"issuerName"`
	IssuedAt          int64                     `json:"issuedAt"`
	PlayerDescription *string                   `json:"playerDescription"`
	Modifications     []SimplePunishmentModView `json:"modifications"`
}

type PunishmentView struct {
	ID                string           `json:"id"`
	IssuerName        string           `json:"issuerName"`
	Issued            *int64           `json:"issued"`
	Started           *int64           `json:"started"`
	TypeOrdinal       int              `json:"typeOrdinal"`
	Type              string           `json:"type"`
	Modifications     []map[string]any `json:"modifications"`
	Notes             []map[string]any `json:"notes"`
	Evidence          []map[string]any `json:"evidence"`
	AttachedTicketIds []string         `json:"attachedTicketIds"`
	Data              map[string]any   `json:"data"`
	PlayerUUID        *string          `json:"playerUuid,omitempty"`
	PlayerName        *string          `json:"playerName,omitempty"`
}

type PunishmentResponseView struct {
	ID                string
	Type              string
	TypeOrdinal       int
	IssuerName        string
	IssuedMs          int64
	StartedMs         *int64
	IsAppealable      bool
	Reason            *string
	Severity          *string
	Status            *string
	Active            bool
	ExpiresMs         *int64
	PlayerUUID        *string
	PlayerUsername    *string
	AltBlocking       *bool
	StatWiping        *bool
	EffectiveCategory *string
	Modifications     []PunishmentModification
	Notes             []PunishmentNote
	Evidence          []PunishmentEvidence
	AttachedTicketIds []string
}

type PunishmentSearchResult struct {
	ID          string `json:"id"`
	PlayerName  string `json:"playerName"`
	TypeOrdinal int    `json:"typeOrdinal"`
	Status      string `json:"status"`
	Issued      int64  `json:"issued"`
}

type SeverityPreview struct {
	Severity          string `json:"severity"`
	Points            int    `json:"points"`
	DurationMs        int64  `json:"durationMs"`
	DurationFormatted string `json:"durationFormatted"`
	PunishmentType    string `json:"punishmentType"`
	Permanent         bool   `json:"permanent"`
	NewSocialStatus   string `json:"newSocialStatus"`
	NewGameplayStatus string `json:"newGameplayStatus"`
	NewSocialPoints   int    `json:"newSocialPoints"`
	NewGameplayPoints int    `json:"newGameplayPoints"`
}

type PunishmentPreview struct {
	Status                       int              `json:"status"`
	Success                      bool             `json:"success"`
	Message                      string           `json:"message,omitempty"`
	SocialStatus                 string           `json:"socialStatus,omitempty"`
	GameplayStatus               string           `json:"gameplayStatus,omitempty"`
	SocialPoints                 int              `json:"socialPoints"`
	GameplayPoints               int              `json:"gameplayPoints"`
	OffenderStatus               string           `json:"offenderStatus,omitempty"`
	Lenient                      *SeverityPreview `json:"lenient,omitempty"`
	Regular                      *SeverityPreview `json:"regular,omitempty"`
	Aggravated                   *SeverityPreview `json:"aggravated,omitempty"`
	SingleSeverity               *SeverityPreview `json:"singleSeverity,omitempty"`
	SingleSeverityPunishment     bool             `json:"singleSeverityPunishment"`
	PermanentUntilUsernameChange bool             `json:"permanentUntilUsernameChange"`
	PermanentUntilSkinChange     bool             `json:"permanentUntilSkinChange"`
	CanBeAltBlocking             bool             `json:"canBeAltBlocking"`
	CanBeStatWiping              bool             `json:"canBeStatWiping"`
	Category                     string           `json:"category,omitempty"`
}

type LinkedAccountView struct {
	MinecraftUUID    string `json:"minecraftUuid"`
	Username         string `json:"username"`
	ActiveBans       int    `json:"activeBans"`
	ActiveMutes      int    `json:"activeMutes"`
	LastLinkedUpdate *int64 `json:"lastLinkedUpdate,omitempty"`
}

type LinkedBanView struct {
	PunishmentID string `json:"punishmentId"`
	PlayerUUID   string `json:"playerUuid"`
	PlayerName   string `json:"playerName"`
	Active       bool   `json:"active"`
}

type PlayerSearchResult struct {
	UUID       string
	Username   string
	Status     string
	LastOnline *int64
	IsOnline   bool
}

type PlayerDetailView struct {
	ID             string
	MinecraftUUID  string
	Usernames      []UsernameEntry
	Notes          []NoteEntry
	IPAddresses    []IPEntry
	Punishments    []PunishmentResponseView
	Data           map[string]any
	Social         string
	Gameplay       string
	SocialPoints   int
	GameplayPoints int
	LatestIPData   *IPEntry
	LastServer     *string
	PlaytimeHours  float64
}

type SyncPunishmentEntry struct {
	MinecraftUUID string               `json:"minecraftUuid"`
	Username      string               `json:"username"`
	Punishment    SimplePunishmentView `json:"punishment"`
}

type SyncDataView struct {
	PendingPunishments         []SyncPunishmentEntry `json:"pendingPunishments"`
	RecentlyStartedPunishments []SyncPunishmentEntry `json:"recentlyStartedPunishments"`
	RecentlyModifiedPunishments []SyncPunishmentEntry `json:"recentlyModifiedPunishments"`
	PlayerNotifications        []map[string]any      `json:"playerNotifications"`
	StaffNotifications         []map[string]any      `json:"staffNotifications"`
	ActiveStaffMembers         []map[string]any      `json:"activeStaffMembers"`
	PendingStatWipes           []map[string]any      `json:"pendingStatWipes"`
	StaffPermissionsUpdatedAt  *int64                `json:"staffPermissionsUpdatedAt"`
	PunishmentTypesUpdatedAt   *int64                `json:"punishmentTypesUpdatedAt"`
	Staff2faVerifications      []map[string]any      `json:"staff2faVerifications"`
	MigrationTask              map[string]any        `json:"migrationTask"`
}

type SyncResult struct {
	Timestamp string       `json:"timestamp"`
	Data      SyncDataView `json:"data"`
}

type SyncOnlinePlayerInput struct {
	UUID              string
	Username          string
	IPAddress         string
	SessionDurationMs int64
}

type SyncChatLogInput struct {
	UUID      string
	Username  string
	Message   string
	Timestamp int64
	Server    string
}

type SyncCommandLogInput struct {
	UUID      string
	Username  string
	Command   string
	Timestamp int64
	Server    string
}

type SyncServerStatusInput struct {
	OnlinePlayerCount int
	MaxPlayers        int
	ServerVersion     string
	Timestamp         string
	PlatformType      string
	PluginVersion     string
}

type CreateNoteInput struct {
	Text       string
	IssuerName string
	IssuerID   string
}

type CreateEvidenceInput struct {
	Text       string
	IssuerName string
	IssuerID   string
	Type       string
	FileURL    string
	FileName   string
	FileType   string
	FileSize   *int64
}

type CreatePunishmentInput struct {
	IssuerName        string
	IssuerID          string
	TypeOrdinal       int
	Notes             []CreateNoteInput
	Evidence          []CreateEvidenceInput
	AttachedTicketIds []string
	Severity          string
	Status            string
	Data              map[string]any
	Reason            string
	Duration          *int64
}

type AddModificationInput struct {
	Type              string
	IssuerName        string
	IssuerID          string
	EffectiveDuration *int64
	Reason            string
	AppealTicketID    string
}

type AddEvidenceInput struct {
	Text       string
	Type       string
	IssuerName string
	IssuerID   string
	URL        string
	FileName   string
	FileType   string
	FileSize   *int64
}

type ModifyTicketsInput struct {
	AddTicketIds            []string
	RemoveTicketIds         []string
	ModifyAssociatedTickets bool
	IssuerName              string
	IssuerID                string
}

type ChatLogRecord struct {
	UUID      string `json:"uuid"`
	Username  string `json:"username"`
	Message   string `json:"message"`
	Server    string `json:"server"`
	Timestamp int64  `json:"timestamp"`
}

type CommandLogRecord struct {
	UUID      string `json:"uuid"`
	Username  string `json:"username"`
	Command   string `json:"command"`
	Server    string `json:"server"`
	Timestamp int64  `json:"timestamp"`
}

type AppealInfoView struct {
	ID             string
	Type           string
	IssuedMs       int64
	ExpiresMs      *int64
	Active         bool
	Appealable     bool
	PlayerUUID     string
	ExistingAppeal map[string]any
	AppealForm     map[string]any
}

type OperationResult struct {
	HTTPStatus int
	Success    bool
	Message    string
	Count      int
}

func parseFlexibleTimeMs(v any) (int64, bool) {
	switch t := v.(type) {
	case int64:
		return t, true
	case int:
		return int64(t), true
	case float64:
		return int64(t), true
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return 0, false
		}
		if parsed, err := time.Parse(time.RFC3339Nano, s); err == nil {
			return parsed.UnixMilli(), true
		}
		if parsed, err := time.Parse(time.RFC3339, s); err == nil {
			return parsed.UnixMilli(), true
		}
	}
	return 0, false
}

func parseFlexibleTimePtr(v any) *int64 {
	ms, ok := parseFlexibleTimeMs(v)
	if !ok {
		return nil
	}
	return &ms
}
