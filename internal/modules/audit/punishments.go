package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

const dayMs = 24 * 60 * 60 * 1000

type punishmentTypeMeta struct {
	Ordinal  int
	Name     string
	Category string
}

var defaultPunishmentTypes = []punishmentTypeMeta{
	{0, "Kick", "Administrative"},
	{1, "Manual Mute", "Administrative"},
	{2, "Manual Ban", "Administrative"},
	{3, "Security Ban", "Administrative"},
	{4, "Linked Ban", "Administrative"},
	{5, "Blacklist", "Administrative"},
	{6, "Chat Abuse", "Social"},
	{7, "Anti Social", "Social"},
	{8, "Targeting", "Social"},
	{9, "Bad Content", "Social"},
	{10, "Bad Username", "Social"},
	{11, "Bad Skin", "Social"},
	{12, "Team Abuse", "Gameplay"},
	{13, "Game Abuse", "Gameplay"},
	{14, "Cheating", "Gameplay"},
	{15, "Game Trading", "Gameplay"},
	{16, "Account Abuse", "Gameplay"},
	{17, "Systems Abuse", "Gameplay"},
}

func loadPunishmentTypes(ctx context.Context, db *sql.DB, serverID string) (map[int]punishmentTypeMeta, map[string]punishmentTypeMeta) {
	byOrdinal := make(map[int]punishmentTypeMeta, len(defaultPunishmentTypes))
	byName := make(map[string]punishmentTypeMeta, len(defaultPunishmentTypes))
	for _, pt := range defaultPunishmentTypes {
		byOrdinal[pt.Ordinal] = pt
		byName[strings.ToLower(pt.Name)] = pt
	}

	var dataJSON string
	err := db.QueryRowContext(ctx,
		`SELECT data_json FROM settings WHERE server_id = ? AND type = 'punishmentTypes' LIMIT 1`,
		serverID,
	).Scan(&dataJSON)
	if err != nil || strings.TrimSpace(dataJSON) == "" {
		return byOrdinal, byName
	}

	var parsed []struct {
		ID       *int   `json:"id"`
		Ordinal  *int   `json:"ordinal"`
		Name     string `json:"name"`
		Category string `json:"category"`
	}
	if err := json.Unmarshal([]byte(dataJSON), &parsed); err != nil || len(parsed) == 0 {
		return byOrdinal, byName
	}

	byOrdinal = make(map[int]punishmentTypeMeta, len(parsed))
	byName = make(map[string]punishmentTypeMeta, len(parsed))
	for _, pt := range parsed {
		ord := -1
		if pt.Ordinal != nil {
			ord = *pt.Ordinal
		} else if pt.ID != nil {
			ord = *pt.ID
		}
		if ord < 0 {
			continue
		}
		meta := punishmentTypeMeta{Ordinal: ord, Name: pt.Name, Category: pt.Category}
		byOrdinal[ord] = meta
		if pt.Name != "" {
			byName[strings.ToLower(pt.Name)] = meta
		}
	}
	return byOrdinal, byName
}

type evidenceEntry struct {
	Text     string `json:"text"`
	URL      string `json:"url"`
	Type     string `json:"type"`
	FileName string `json:"fileName"`
}

type noteEntry struct {
	ID         string `json:"id"`
	Text       string `json:"text"`
	Date       int64  `json:"date"`
	IssuerName string `json:"issuerName"`
	IssuerID   string `json:"issuerId,omitempty"`
}

type modificationEntry struct {
	ID                string         `json:"id"`
	Type              string         `json:"type"`
	Date              int64          `json:"date"`
	IssuerName        string         `json:"issuerName,omitempty"`
	IssuerID          string         `json:"issuerId,omitempty"`
	Reason            string         `json:"reason,omitempty"`
	EffectiveDuration *int64         `json:"effectiveDuration,omitempty"`
	AppealTicketID    string         `json:"appealTicketId,omitempty"`
	Data              map[string]any `json:"data,omitempty"`
}

type parsedPunishment struct {
	ID                string
	PlayerID          string
	PlayerUUID        string
	PlayerUsername    string
	TypeRaw           string
	TypeID            string
	TypeOrdinal       int
	Reason            string
	IssuerUUID        string
	IssuerName        string
	IssuedAt          int64
	StartedAt         *int64
	ExpiresAt         *int64
	DurationMs        *int64
	Pardoned          bool
	NotesJSON         string
	EvidenceJSON      string
	DataJSON          string
	DataMap           map[string]any
	Modifications     []modificationEntry
	Evidence          []evidenceEntry
	AttachedTicketIDs []string
}

func parsePunishmentRow(
	id, playerID, playerUUID, playerUsername, typeRaw, typeID, reason, issuerUUID, issuerName string,
	issuedAt int64, expiresAt, durationMs sql.NullInt64, pardoned int,
	notesJSON, evidenceJSON, dataJSON string,
	byName map[string]punishmentTypeMeta,
) parsedPunishment {
	dataMap := map[string]any{}
	if strings.TrimSpace(dataJSON) != "" && dataJSON != "{}" {
		_ = json.Unmarshal([]byte(dataJSON), &dataMap)
	}

	var dataOrd *int
	if v, ok := dataMap["typeOrdinal"].(float64); ok {
		n := int(v)
		dataOrd = &n
	}
	ord := -1
	if dataOrd != nil {
		ord = *dataOrd
	} else if n, err := strconv.Atoi(strings.TrimSpace(typeID)); err == nil {
		ord = n
	} else if n, err := strconv.Atoi(strings.TrimSpace(typeRaw)); err == nil {
		ord = n
	} else if meta, ok := byName[strings.ToLower(strings.TrimSpace(typeRaw))]; ok {
		ord = meta.Ordinal
	}

	if reason == "" {
		if r, ok := dataMap["reason"].(string); ok {
			reason = r
		}
	}

	var startedPtr *int64
	if s, ok := dataMap["started"].(float64); ok && s > 0 {
		v := int64(s)
		startedPtr = &v
	}

	var expPtr *int64
	if expiresAt.Valid {
		v := expiresAt.Int64
		expPtr = &v
	}
	var durPtr *int64
	if durationMs.Valid {
		v := durationMs.Int64
		durPtr = &v
	} else if d, ok := dataMap["duration"].(float64); ok {
		v := int64(d)
		durPtr = &v
	}

	var mods []modificationEntry
	if rawMods, ok := dataMap["modifications"]; ok {
		if b, err := json.Marshal(rawMods); err == nil {
			_ = json.Unmarshal(b, &mods)
		}
	}

	var attached []string
	if rawTickets, ok := dataMap["attachedTicketIds"]; ok {
		if b, err := json.Marshal(rawTickets); err == nil {
			_ = json.Unmarshal(b, &attached)
		}
	}

	var evidence []evidenceEntry
	if strings.TrimSpace(evidenceJSON) != "" && evidenceJSON != "[]" {
		_ = json.Unmarshal([]byte(evidenceJSON), &evidence)
	}

	isPardoned := pardoned != 0
	if st, ok := dataMap["status"].(string); ok && strings.EqualFold(st, "PARDONED") {
		isPardoned = true
	}
	if !isPardoned {
		for _, m := range mods {
			switch strings.ToUpper(m.Type) {
			case "MANUAL_PARDON", "APPEAL_ACCEPT", "SYSTEM_PARDON":
				isPardoned = true
			}
		}
	}

	return parsedPunishment{
		ID:                id,
		PlayerID:          playerID,
		PlayerUUID:        playerUUID,
		PlayerUsername:    playerUsername,
		TypeRaw:           typeRaw,
		TypeID:            typeID,
		TypeOrdinal:       ord,
		Reason:            reason,
		IssuerUUID:        issuerUUID,
		IssuerName:        issuerName,
		IssuedAt:          issuedAt,
		StartedAt:         startedPtr,
		ExpiresAt:         expPtr,
		DurationMs:        durPtr,
		Pardoned:          isPardoned,
		NotesJSON:         notesJSON,
		EvidenceJSON:      evidenceJSON,
		DataJSON:          dataJSON,
		DataMap:           dataMap,
		Modifications:     mods,
		Evidence:          evidence,
		AttachedTicketIDs: attached,
	}
}

func (p parsedPunishment) hasModification(types ...string) bool {
	for _, m := range p.Modifications {
		for _, t := range types {
			if strings.EqualFold(m.Type, t) {
				return true
			}
		}
	}
	return false
}

func (p parsedPunishment) isUnstarted() bool {
	if st, ok := p.DataMap["status"].(string); ok && strings.EqualFold(st, "UNSTARTED") {
		return true
	}
	return false
}

func (p parsedPunishment) effectiveExpiryMs() (int64, bool) {
	if p.ExpiresAt != nil {
		return *p.ExpiresAt, true
	}
	var latestDur *modificationEntry
	for i := range p.Modifications {
		t := strings.ToUpper(p.Modifications[i].Type)
		if t == "MANUAL_DURATION_CHANGE" || t == "APPEAL_DURATION_CHANGE" {
			latestDur = &p.Modifications[i]
		}
	}
	if latestDur != nil {
		if latestDur.EffectiveDuration == nil || *latestDur.EffectiveDuration <= 0 {
			return 0, false
		}
		return latestDur.Date + *latestDur.EffectiveDuration, true
	}
	if p.DurationMs == nil || *p.DurationMs <= 0 {
		return 0, false
	}
	base := p.IssuedAt
	if p.StartedAt != nil && *p.StartedAt > 0 {
		base = *p.StartedAt
	}
	return base + *p.DurationMs, true
}

func (p parsedPunishment) isActive(nowMs int64) bool {
	if p.TypeOrdinal == 0 || p.isUnstarted() || p.Pardoned {
		return false
	}
	if exp, ok := p.effectiveExpiryMs(); ok {
		return exp > nowMs
	}
	return true
}

func (p parsedPunishment) typeName(byOrdinal map[int]punishmentTypeMeta) string {
	if meta, ok := byOrdinal[p.TypeOrdinal]; ok && meta.Name != "" {
		return meta.Name
	}
	if p.TypeRaw != "" {
		if _, err := strconv.Atoi(p.TypeRaw); err != nil {
			return p.TypeRaw
		}
	}
	return "Unknown"
}

func (p parsedPunishment) categoryName(byOrdinal map[int]punishmentTypeMeta) string {
	if meta, ok := byOrdinal[p.TypeOrdinal]; ok && meta.Category != "" {
		return meta.Category
	}
	return "Administrative"
}

func formatAuditDuration(durMs *int64) string {
	if durMs == nil || *durMs <= 0 {
		return "Permanent"
	}
	ms := *durMs
	days := ms / (24 * 60 * 60 * 1000)
	if days > 0 {
		return fmt.Sprintf("%dd", days)
	}
	hours := ms / (60 * 60 * 1000)
	if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}
	mins := ms / (60 * 1000)
	if mins > 0 {
		return fmt.Sprintf("%dm", mins)
	}
	return fmt.Sprintf("%ds", ms/1000)
}

func extractLatestUsername(currentUsername, usernamesJSON string) string {
	if strings.TrimSpace(currentUsername) != "" {
		return currentUsername
	}
	var entries []struct {
		Username string `json:"username"`
	}
	if err := json.Unmarshal([]byte(usernamesJSON), &entries); err == nil && len(entries) > 0 {
		if last := strings.TrimSpace(entries[len(entries)-1].Username); last != "" {
			return last
		}
	}
	return "Unknown"
}

func getStartDateMs(period string, nowMs int64) (int64, bool) {
	p := strings.TrimSpace(period)
	if p == "" {
		p = "30d"
	}
	if p == "all" {
		return 0, false
	}
	switch p {
	case "7d":
		return nowMs - 7*dayMs, true
	case "90d":
		return nowMs - 90*dayMs, true
	case "1y":
		return nowMs - 365*dayMs, true
	default:
		return nowMs - 30*dayMs, true
	}
}
