package dashboard

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

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

type settingsPunishmentType struct {
	ID       *int   `json:"id"`
	Ordinal  *int   `json:"ordinal"`
	Name     string `json:"name"`
	Category string `json:"category"`
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

	var parsed []settingsPunishmentType
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
		meta := punishmentTypeMeta{
			Ordinal:  ord,
			Name:     pt.Name,
			Category: pt.Category,
		}
		byOrdinal[ord] = meta
		if pt.Name != "" {
			byName[strings.ToLower(pt.Name)] = meta
		}
	}
	return byOrdinal, byName
}

type punishmentModification struct {
	Type              string `json:"type"`
	Date              int64  `json:"date"`
	EffectiveDuration *int64 `json:"effectiveDuration"`
}

type punishmentDataJSON struct {
	TypeOrdinal         *int                     `json:"typeOrdinal"`
	Started             *int64                   `json:"started"`
	Status              string                   `json:"status"`
	EnforcementCategory string                   `json:"enforcementCategory"`
	Duration            *int64                   `json:"duration"`
	Reason              string                   `json:"reason"`
	Modifications       []punishmentModification `json:"modifications"`
}

type parsedPunishment struct {
	ID             string
	PlayerUUID     string
	PlayerUsername string
	TypeRaw        string
	TypeID         string
	TypeOrdinal    int
	Reason         string
	IssuerName     string
	IssuedAt       int64
	ExpiresAt      *int64
	DurationMs     *int64
	Pardoned       bool
	Data           punishmentDataJSON
}

func parsePunishment(
	id, playerUUID, playerUsername, typeRaw, typeID, reason, issuerName string,
	issuedAt int64, expiresAt, durationMs sql.NullInt64, pardoned int, dataJSON string,
	byName map[string]punishmentTypeMeta,
) parsedPunishment {
	var data punishmentDataJSON
	if strings.TrimSpace(dataJSON) != "" && dataJSON != "{}" {
		_ = json.Unmarshal([]byte(dataJSON), &data)
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
	}

	ord := resolveTypeOrdinal(typeRaw, typeID, data.TypeOrdinal, byName)
	if reason == "" && data.Reason != "" {
		reason = data.Reason
	}

	isPardoned := pardoned != 0 || strings.EqualFold(data.Status, "PARDONED")
	if !isPardoned {
		for _, m := range data.Modifications {
			switch strings.ToUpper(m.Type) {
			case "MANUAL_PARDON", "APPEAL_ACCEPT", "SYSTEM_PARDON":
				isPardoned = true
			}
		}
	}

	return parsedPunishment{
		ID:             id,
		PlayerUUID:     playerUUID,
		PlayerUsername: playerUsername,
		TypeRaw:        typeRaw,
		TypeID:         typeID,
		TypeOrdinal:    ord,
		Reason:         reason,
		IssuerName:     issuerName,
		IssuedAt:       issuedAt,
		ExpiresAt:      expPtr,
		DurationMs:     durPtr,
		Pardoned:       isPardoned,
		Data:           data,
	}
}

func resolveTypeOrdinal(typeRaw, typeID string, dataOrd *int, byName map[string]punishmentTypeMeta) int {
	if dataOrd != nil {
		return *dataOrd
	}
	if n, err := strconv.Atoi(strings.TrimSpace(typeID)); err == nil {
		return n
	}
	if n, err := strconv.Atoi(strings.TrimSpace(typeRaw)); err == nil {
		return n
	}
	if meta, ok := byName[strings.ToLower(strings.TrimSpace(typeRaw))]; ok {
		return meta.Ordinal
	}
	return -1
}

func (p parsedPunishment) isActive(nowMs int64) bool {
	if p.TypeOrdinal == 0 {
		return false
	}
	if strings.EqualFold(p.Data.Status, "UNSTARTED") {
		return false
	}
	if p.Pardoned {
		return false
	}
	if p.ExpiresAt != nil {
		return *p.ExpiresAt > nowMs
	}

	// Check latest duration-change modification
	var latestDurMod *punishmentModification
	for i := range p.Data.Modifications {
		m := &p.Data.Modifications[i]
		t := strings.ToUpper(m.Type)
		if t == "MANUAL_DURATION_CHANGE" || t == "APPEAL_DURATION_CHANGE" {
			latestDurMod = m
		}
	}
	if latestDurMod != nil {
		if latestDurMod.EffectiveDuration == nil || *latestDurMod.EffectiveDuration <= 0 {
			return true
		}
		return latestDurMod.Date+*latestDurMod.EffectiveDuration > nowMs
	}

	var dur int64
	if p.DurationMs != nil {
		dur = *p.DurationMs
	} else if p.Data.Duration != nil {
		dur = *p.Data.Duration
	}
	if dur <= 0 {
		return true
	}
	base := p.IssuedAt
	if p.Data.Started != nil && *p.Data.Started > 0 {
		base = *p.Data.Started
	}
	return base+dur > nowMs
}

func (p parsedPunishment) resolveTypeName(byOrdinal map[int]punishmentTypeMeta) string {
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

func (p parsedPunishment) effectiveCategory(byOrdinal map[int]punishmentTypeMeta) string {
	if p.Data.EnforcementCategory != "" {
		return strings.ToUpper(p.Data.EnforcementCategory)
	}
	switch p.TypeOrdinal {
	case 0:
		return ""
	case 1, 6, 7:
		return "MUTE"
	case 2, 3, 4, 5, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17:
		return "BAN"
	}
	if meta, ok := byOrdinal[p.TypeOrdinal]; ok {
		low := strings.ToLower(meta.Category)
		if strings.Contains(low, "mute") {
			return "MUTE"
		}
		if strings.Contains(low, "ban") {
			return "BAN"
		}
		lowName := strings.ToLower(meta.Name)
		if strings.Contains(lowName, "mute") {
			return "MUTE"
		}
		if strings.Contains(lowName, "ban") || strings.Contains(lowName, "blacklist") {
			return "BAN"
		}
	}
	lowRaw := strings.ToLower(p.TypeRaw)
	if strings.Contains(lowRaw, "mute") {
		return "MUTE"
	}
	if strings.Contains(lowRaw, "ban") || strings.Contains(lowRaw, "blacklist") {
		return "BAN"
	}
	return ""
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

func isPlayerOnline(dataJSON string) bool {
	var m map[string]any
	if err := json.Unmarshal([]byte(dataJSON), &m); err != nil {
		return false
	}
	v, ok := m["isOnline"]
	if !ok {
		return false
	}
	switch b := v.(type) {
	case bool:
		return b
	case float64:
		return b != 0
	case string:
		return strings.EqualFold(b, "true") || b == "1"
	}
	return false
}

func extractPlayerFirstJoinMs(firstSeen int64, dataJSON string) int64 {
	var m map[string]any
	if err := json.Unmarshal([]byte(dataJSON), &m); err == nil {
		if raw, ok := m["firstJoin"]; ok && raw != nil {
			switch v := raw.(type) {
			case float64:
				if v > 0 {
					return int64(v)
				}
			case string:
				if t, err := time.Parse(time.RFC3339, v); err == nil {
					return t.UnixMilli()
				}
				if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
					return n
				}
			}
		}
	}
	return firstSeen
}
