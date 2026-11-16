package player

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/httpx"

	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrAlreadyExist = errors.New("already exists")
)

const (
	dataMetaStarted           = "_started"
	dataMetaTypeOrdinal       = "_typeOrdinal"
	dataMetaModifications     = "_modifications"
	dataMetaAttachedTicketIds = "_attachedTicketIds"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) DB() *sql.DB {
	return s.db
}

// --- Settings Loaders ---

func (s *Store) GetPunishmentTypes(ctx context.Context, serverID string) []PunishmentType {
	var dataJSON string
	err := s.db.QueryRowContext(ctx,
		`SELECT data_json FROM settings WHERE server_id = ? AND type = 'punishmentTypes' LIMIT 1`,
		serverID,
	).Scan(&dataJSON)
	if err != nil || strings.TrimSpace(dataJSON) == "" {
		return DefaultPunishmentTypes()
	}
	var wrap modlv1.PanelPunishmentTypesResponse
	if (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal([]byte(dataJSON), &wrap) == nil && len(wrap.PunishmentTypes) > 0 {
		out := make([]PunishmentType, 0, len(wrap.PunishmentTypes))
		for _, pt := range wrap.PunishmentTypes {
			out = append(out, protoToPunishmentType(pt))
		}
		return out
	}
	// Fallback direct JSON array/object
	var raw struct {
		PunishmentTypes []PunishmentType `json:"punishmentTypes"`
	}
	if json.Unmarshal([]byte(dataJSON), &raw) == nil && len(raw.PunishmentTypes) > 0 {
		return raw.PunishmentTypes
	}
	return DefaultPunishmentTypes()
}

func protoToPunishmentType(pt *modlv1.PunishmentType) PunishmentType {
	out := PunishmentType{
		Name:              pt.Name,
		Category:          pt.Category,
		StaffDescription:  pt.StaffDescription,
		PlayerDescription: pt.PlayerDescription,
	}
	if pt.Id != nil {
		out.ID = int(*pt.Id)
	}
	if pt.Ordinal != nil {
		out.Ordinal = int(*pt.Ordinal)
	} else {
		out.Ordinal = out.ID
	}
	if pt.Customizable != nil {
		out.Customizable = *pt.Customizable
	}
	if pt.SingleSeverityPunishment != nil {
		out.SingleSeverityPunishment = *pt.SingleSeverityPunishment
	}
	if pt.SingleSeverityPoints != nil {
		v := int(*pt.SingleSeverityPoints)
		out.SingleSeverityPoints = &v
	}
	if pt.CustomPoints != nil {
		v := int(*pt.CustomPoints)
		out.CustomPoints = &v
	}
	if pt.CanBeAltBlocking != nil {
		out.CanBeAltBlocking = *pt.CanBeAltBlocking
	}
	if pt.CanBeStatWiping != nil {
		out.CanBeStatWiping = *pt.CanBeStatWiping
	}
	if pt.Appealable != nil {
		out.Appealable = *pt.Appealable
	}
	if pt.PermanentUntilSkinChange != nil {
		out.PermanentUntilSkinChange = *pt.PermanentUntilSkinChange
	}
	if pt.PermanentUntilUsernameChange != nil {
		out.PermanentUntilUsernameChange = *pt.PermanentUntilUsernameChange
	}
	if pt.Points != nil {
		out.Points = &PunishmentPoints{
			Low:     int(pt.Points.Low),
			Regular: int(pt.Points.Regular),
			Severe:  int(pt.Points.Severe),
		}
	}
	convDur := func(d *modlv1.DurationDetail) *DurationDetail {
		if d == nil {
			return nil
		}
		return &DurationDetail{Value: int64(d.Value), Unit: d.Unit, Type: d.Type}
	}
	convOff := func(o *modlv1.OffenseLevelDurations) *OffenseLevelDurations {
		if o == nil {
			return nil
		}
		return &OffenseLevelDurations{
			First:    convDur(o.First),
			Medium:   convDur(o.Medium),
			Habitual: convDur(o.Habitual),
		}
	}
	if pt.Durations != nil {
		out.Durations = &PunishmentDurations{
			Low:     convOff(pt.Durations.Low),
			Regular: convOff(pt.Durations.Regular),
			Severe:  convOff(pt.Durations.Severe),
		}
	}
	if pt.SingleSeverityDurations != nil {
		out.SingleSeverityDurations = convOff(pt.SingleSeverityDurations)
	}
	if pt.AppealForm != nil && len(pt.AppealForm.Fields) > 0 {
		b, err := (protojson.MarshalOptions{EmitUnpopulated: false}).Marshal(pt.AppealForm)
		if err == nil {
			var m map[string]any
			if json.Unmarshal(b, &m) == nil {
				out.AppealForm = m
			}
		}
	}
	return out
}

func (s *Store) GetThresholdSettings(ctx context.Context, serverID string) OffenderThresholdSettings {
	def := DefaultThresholdSettings()
	var dataJSON string
	err := s.db.QueryRowContext(ctx,
		`SELECT data_json FROM settings WHERE server_id = ? AND type = 'statusThresholds' LIMIT 1`,
		serverID,
	).Scan(&dataJSON)
	if err != nil || strings.TrimSpace(dataJSON) == "" {
		return def
	}
	var pb modlv1.OffenderThresholdSettings
	if (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal([]byte(dataJSON), &pb) == nil {
		if pb.Social != nil {
			def.Social = CategoryThresholds{
				Medium:            int(pb.Social.Medium),
				Habitual:          int(pb.Social.Habitual),
				PointExpiryMonths: int(pb.Social.PointExpiryMonths),
			}
		}
		if pb.Gameplay != nil {
			def.Gameplay = CategoryThresholds{
				Medium:            int(pb.Gameplay.Medium),
				Habitual:          int(pb.Gameplay.Habitual),
				PointExpiryMonths: int(pb.Gameplay.PointExpiryMonths),
			}
		}
		return def
	}
	var raw OffenderThresholdSettings
	if json.Unmarshal([]byte(dataJSON), &raw) == nil {
		return raw
	}
	return def
}

func (s *Store) GetWebhookSettings(ctx context.Context, serverID string) *modlv1.WebhookSettings {
	var dataJSON string
	err := s.db.QueryRowContext(ctx,
		`SELECT data_json FROM settings WHERE server_id = ? AND type = 'webhooks' LIMIT 1`,
		serverID,
	).Scan(&dataJSON)
	if err != nil || strings.TrimSpace(dataJSON) == "" {
		return nil
	}
	var ws modlv1.WebhookSettings
	if (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal([]byte(dataJSON), &ws) != nil {
		return nil
	}
	return &ws
}

// --- Player Queries & Mutations ---

func (s *Store) scanPlayerRow(scanner interface{ Scan(...any) error }) (*Player, error) {
	var (
		p                             Player
		usernamesJSON, notesJSON      string
		ipsJSON, dataJSON             string
	)
	err := scanner.Scan(
		&p.ID, &p.ServerID, &p.MinecraftUUID, &p.CurrentUsername,
		&usernamesJSON, &notesJSON, &ipsJSON, &dataJSON,
		&p.FirstSeenMs, &p.LastSeenMs,
	)
	if err != nil {
		return nil, err
	}
	p.Usernames = decodeUsernames(usernamesJSON)
	p.Notes = decodeNotes(notesJSON)
	p.IPAddresses = decodeIPs(ipsJSON)
	p.Data = decodeMap(dataJSON)
	p.Punishments = []*Punishment{}
	return &p, nil
}

const playerSelectCols = `id, server_id, minecraft_uuid, current_username, usernames_json, notes_json, ip_addresses_json, data_json, first_seen, last_seen`

func (s *Store) FindPlayerByUUID(ctx context.Context, serverID, rawUUID string) (*Player, error) {
	norm, ok := NormalizeUUID(rawUUID)
	if !ok {
		norm = strings.ToLower(strings.TrimSpace(rawUUID))
	}
	row := s.db.QueryRowContext(ctx,
		`SELECT `+playerSelectCols+` FROM players WHERE server_id = ? AND LOWER(minecraft_uuid) = ? LIMIT 1`,
		serverID, norm,
	)
	p, err := s.scanPlayerRow(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	puns, err := s.loadPunishmentsForPlayer(ctx, serverID, p.MinecraftUUID)
	if err != nil {
		return nil, err
	}
	p.Punishments = puns
	return p, nil
}

func (s *Store) FindPlayersByUUIDs(ctx context.Context, serverID string, uuids []string) ([]*Player, error) {
	if len(uuids) == 0 {
		return nil, nil
	}
	seen := make(map[string]bool, len(uuids))
	normList := make([]any, 0, len(uuids))
	ph := make([]string, 0, len(uuids))
	for _, u := range uuids {
		n, ok := NormalizeUUID(u)
		if !ok {
			n = strings.ToLower(strings.TrimSpace(u))
		}
		if n == "" || seen[n] {
			continue
		}
		seen[n] = true
		normList = append(normList, n)
		ph = append(ph, "?")
	}
	if len(normList) == 0 {
		return nil, nil
	}
	args := append([]any{serverID}, normList...)
	q := fmt.Sprintf(`SELECT %s FROM players WHERE server_id = ? AND LOWER(minecraft_uuid) IN (%s)`,
		playerSelectCols, strings.Join(ph, ","))
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	var out []*Player
	byUUID := make(map[string]*Player)
	for rows.Next() {
		p, err := s.scanPlayerRow(rows)
		if err != nil {
			_ = rows.Close()
			return nil, err
		}
		out = append(out, p)
		byUUID[strings.ToLower(p.MinecraftUUID)] = p
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return nil, err
	}
	if err := s.attachPunishmentsToPlayers(ctx, serverID, byUUID); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Store) FindBestByUsername(ctx context.Context, serverID, username string) (*Player, error) {
	u := strings.TrimSpace(username)
	if u == "" {
		return nil, nil
	}
	// Exact current_username match first (newest last_seen)
	row := s.db.QueryRowContext(ctx,
		`SELECT `+playerSelectCols+` FROM players WHERE server_id = ? AND LOWER(current_username) = LOWER(?) ORDER BY last_seen DESC LIMIT 1`,
		serverID, u,
	)
	p, err := s.scanPlayerRow(row)
	if err == nil {
		p.Punishments, err = s.loadPunishmentsForPlayer(ctx, serverID, p.MinecraftUUID)
		return p, err
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	// Fallback: search historical usernames_json
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+playerSelectCols+` FROM players WHERE server_id = ? AND LOWER(usernames_json) LIKE ? ESCAPE '\' ORDER BY last_seen DESC LIMIT 50`,
		serverID, "%"+escapeLike(u)+"%",
	)
	if err != nil {
		return nil, err
	}
	var matched *Player
	for rows.Next() {
		cand, scanErr := s.scanPlayerRow(rows)
		if scanErr != nil {
			continue
		}
		for _, entry := range cand.Usernames {
			if strings.EqualFold(entry.Username, u) {
				matched = cand
				break
			}
		}
		if matched != nil {
			break
		}
	}
	err = rows.Err()
	_ = rows.Close()
	if matched != nil {
		matched.Punishments, _ = s.loadPunishmentsForPlayer(ctx, serverID, matched.MinecraftUUID)
		return matched, nil
	}
	return nil, err
}

func (s *Store) SearchPlayers(ctx context.Context, serverID, query string, limit int) ([]*Player, error) {
	if limit <= 0 {
		limit = 25
	}
	q := strings.TrimSpace(query)
	like := "%" + escapeLike(q) + "%"
	rows, err := s.db.QueryContext(ctx, `
		SELECT `+playerSelectCols+` FROM players
		WHERE server_id = ?
		  AND (LOWER(current_username) LIKE ? ESCAPE '\'
		       OR LOWER(minecraft_uuid) LIKE ? ESCAPE '\'
		       OR LOWER(usernames_json) LIKE ? ESCAPE '\')
		ORDER BY last_seen DESC
		LIMIT ?`,
		serverID, like, like, like, limit,
	)
	if err != nil {
		return nil, err
	}
	var out []*Player
	byUUID := make(map[string]*Player)
	for rows.Next() {
		p, err := s.scanPlayerRow(rows)
		if err != nil {
			_ = rows.Close()
			return nil, err
		}
		out = append(out, p)
		byUUID[strings.ToLower(p.MinecraftUUID)] = p
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return nil, err
	}
	if err := s.attachPunishmentsToPlayers(ctx, serverID, byUUID); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Store) CreatePlayer(ctx context.Context, serverID, uuidStr, username string) (*Player, error) {
	norm, ok := NormalizeUUID(uuidStr)
	if !ok {
		return nil, fmt.Errorf("invalid uuid")
	}
	existing, err := s.FindPlayerByUUID(ctx, serverID, norm)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAlreadyExist
	}
	now := httpx.NowMs()
	p := &Player{
		ID:              httpx.NewID(),
		ServerID:        serverID,
		MinecraftUUID:   norm,
		CurrentUsername: username,
		Usernames:       []UsernameEntry{{Username: username, DateMs: now}},
		Notes:           []NoteEntry{},
		IPAddresses:     []IPEntry{},
		Punishments:     []*Punishment{},
		Data:            map[string]any{PlayerKeyFirstJoin: now},
		FirstSeenMs:     now,
		LastSeenMs:      now,
	}
	if err := s.insertPlayer(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Store) insertPlayer(ctx context.Context, p *Player) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO players (
			id, server_id, minecraft_uuid, current_username,
			usernames_json, notes_json, ip_addresses_json, data_json,
			first_seen, last_seen
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.ServerID, p.MinecraftUUID, p.CurrentUsername,
		mustJSON(p.Usernames, "[]"),
		mustJSON(p.Notes, "[]"),
		mustJSON(p.IPAddresses, "[]"),
		mustJSON(p.Data, "{}"),
		p.FirstSeenMs, p.LastSeenMs,
	)
	if err == nil {
		_, _ = s.db.ExecContext(ctx,
			`UPDATE servers SET user_count = (SELECT COUNT(*) FROM players WHERE server_id = ?), updated_at = ? WHERE id = ?`,
			p.ServerID, httpx.NowMs(), p.ServerID,
		)
	}
	return err
}

func (s *Store) savePlayerRecord(ctx context.Context, p *Player) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE players SET
			current_username = ?,
			usernames_json = ?,
			notes_json = ?,
			ip_addresses_json = ?,
			data_json = ?,
			last_seen = ?
		WHERE server_id = ? AND id = ?`,
		p.CurrentUsername,
		mustJSON(p.Usernames, "[]"),
		mustJSON(p.Notes, "[]"),
		mustJSON(p.IPAddresses, "[]"),
		mustJSON(p.Data, "{}"),
		p.LastSeenMs,
		p.ServerID, p.ID,
	)
	return err
}

func (s *Store) upsertPlayerIPIndex(ctx context.Context, serverID, playerID, uuidStr, ipAddr string, nowMs int64) {
	if strings.TrimSpace(ipAddr) == "" {
		return
	}
	_, _ = s.db.ExecContext(ctx, `
		INSERT INTO player_ips (server_id, player_id, minecraft_uuid, ip_address, first_seen, last_seen)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(server_id, player_id, ip_address) DO UPDATE SET last_seen = excluded.last_seen`,
		serverID, playerID, uuidStr, ipAddr, nowMs, nowMs,
	)
}

func (s *Store) UpsertLoginPlayer(ctx context.Context, serverID, uuidStr, username, ipAddr string, ipInfo map[string]any, skinHash, serverName string) (*Player, bool, error) {
	norm, ok := NormalizeUUID(uuidStr)
	if !ok {
		return nil, false, fmt.Errorf("invalid uuid")
	}
	now := httpx.NowMs()
	p, err := s.FindPlayerByUUID(ctx, serverID, norm)
	if err != nil {
		return nil, false, err
	}

	isNewIP := false
	if p == nil {
		data := map[string]any{
			PlayerKeyIsOnline:             true,
			PlayerKeyLastLogin:            now,
			PlayerKeyFirstJoin:            now,
			PlayerKeyTotalPlaytimeSeconds: int64(0),
		}
		if serverName != "" {
			data[PlayerKeyLastServer] = serverName
		}
		if skinHash != "" {
			data[PlayerKeyLastSkinHash] = skinHash
		}
		var ips []IPEntry
		if ipAddr != "" {
			isNewIP = true
			entry := IPEntry{
				IPAddress:    ipAddr,
				FirstLoginMs: now,
				LoginsMs:     []int64{now},
			}
			applyIPInfoToEntry(&entry, ipInfo)
			applyIPInfoToPlayerData(data, ipInfo)
			ips = append(ips, entry)
		}
		p = &Player{
			ID:              httpx.NewID(),
			ServerID:        serverID,
			MinecraftUUID:   norm,
			CurrentUsername: username,
			Usernames:       []UsernameEntry{{Username: username, DateMs: now}},
			Notes:           []NoteEntry{},
			IPAddresses:     ips,
			Punishments:     []*Punishment{},
			Data:            data,
			FirstSeenMs:     now,
			LastSeenMs:      now,
		}
		if err := s.insertPlayer(ctx, p); err != nil {
			return nil, false, err
		}
		if ipAddr != "" {
			s.upsertPlayerIPIndex(ctx, serverID, p.ID, p.MinecraftUUID, ipAddr, now)
		}
		return p, isNewIP, nil
	}

	// Update existing player
	if p.Data == nil {
		p.Data = map[string]any{}
	}
	p.LastSeenMs = now
	p.Data[PlayerKeyIsOnline] = true
	p.Data[PlayerKeyLastLogin] = now
	if serverName != "" {
		p.Data[PlayerKeyLastServer] = serverName
	}
	if skinHash != "" {
		p.Data[PlayerKeyLastSkinHash] = skinHash
	}

	if username != "" {
		if len(p.Usernames) == 0 || !strings.EqualFold(p.Usernames[len(p.Usernames)-1].Username, username) {
			p.Usernames = append(p.Usernames, UsernameEntry{Username: username, DateMs: now})
		}
		p.CurrentUsername = username
	}

	if ipAddr != "" {
		foundIdx := -1
		for i := range p.IPAddresses {
			if p.IPAddresses[i].IPAddress == ipAddr {
				foundIdx = i
				break
			}
		}
		if foundIdx >= 0 {
			p.IPAddresses[foundIdx].LoginsMs = append(p.IPAddresses[foundIdx].LoginsMs, now)
			applyIPInfoToEntry(&p.IPAddresses[foundIdx], ipInfo)
		} else {
			isNewIP = true
			entry := IPEntry{
				IPAddress:    ipAddr,
				FirstLoginMs: now,
				LoginsMs:     []int64{now},
			}
			applyIPInfoToEntry(&entry, ipInfo)
			p.IPAddresses = append(p.IPAddresses, entry)
		}
		applyIPInfoToPlayerData(p.Data, ipInfo)
		s.upsertPlayerIPIndex(ctx, serverID, p.ID, p.MinecraftUUID, ipAddr, now)
	}

	if err := s.savePlayerRecord(ctx, p); err != nil {
		return nil, false, err
	}
	return p, isNewIP, nil
}

func applyIPInfoToEntry(entry *IPEntry, info map[string]any) {
	if len(info) == 0 {
		return
	}
	if v, ok := info["country"].(string); ok && v != "" {
		entry.Country = &v
	}
	if v, ok := info["region"].(string); ok && v != "" {
		entry.Region = &v
	}
	if v, ok := info["asn"].(string); ok && v != "" {
		entry.ASN = &v
	}
	if v, ok := info["proxy"].(bool); ok {
		entry.Proxy = v
	}
	if v, ok := info["hosting"].(bool); ok {
		entry.Hosting = v
	}
}

func applyIPInfoToPlayerData(data, info map[string]any) {
	if len(info) == 0 || data == nil {
		return
	}
	for _, k := range []string{PlayerKeyCountry, PlayerKeyRegion, PlayerKeyASN, PlayerKeyProxy, PlayerKeyHosting} {
		if v, exists := info[k]; exists && v != nil {
			data[k] = v
		}
	}
}

func (s *Store) MarkDisconnected(ctx context.Context, serverID, rawUUID string, sessionDurationMs int64) error {
	p, err := s.FindPlayerByUUID(ctx, serverID, rawUUID)
	if err != nil || p == nil {
		return err
	}
	if p.Data == nil {
		p.Data = map[string]any{}
	}
	p.Data[PlayerKeyIsOnline] = false
	if sessionDurationMs > 0 {
		prev := p.TotalPlaytimeSeconds()
		p.Data[PlayerKeyTotalPlaytimeSeconds] = prev + sessionDurationMs/1000
	}
	p.LastSeenMs = httpx.NowMs()
	return s.savePlayerRecord(ctx, p)
}

func (s *Store) UpdateLastServer(ctx context.Context, serverID, rawUUID, serverName string) error {
	p, err := s.FindPlayerByUUID(ctx, serverID, rawUUID)
	if err != nil || p == nil {
		return err
	}
	if p.Data == nil {
		p.Data = map[string]any{}
	}
	p.Data[PlayerKeyLastServer] = serverName
	return s.savePlayerRecord(ctx, p)
}

func (s *Store) FindOnlinePlayers(ctx context.Context, serverID string, limit int) ([]*Player, error) {
	if limit <= 0 {
		limit = 500
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT `+playerSelectCols+` FROM players
		WHERE server_id = ?
		  AND COALESCE(json_extract(data_json, '$.isOnline'), 0) = 1
		ORDER BY last_seen DESC
		LIMIT ?`,
		serverID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Player
	for rows.Next() {
		p, err := s.scanPlayerRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) MarkStalePlayersOffline(ctx context.Context, serverID string, onlineUUIDs map[string]bool, serverName string, nowMs int64) {
	onlinePlayers, err := s.FindOnlinePlayers(ctx, serverID, 1000)
	if err != nil {
		return
	}
	for _, p := range onlinePlayers {
		if onlineUUIDs[strings.ToLower(p.MinecraftUUID)] {
			continue
		}
		if serverName != "" {
			if ls := p.LastServer(); ls != "" && !strings.EqualFold(ls, serverName) {
				continue
			}
		}
		if p.Data == nil {
			p.Data = map[string]any{}
		}
		p.Data[PlayerKeyIsOnline] = false
		_ = s.savePlayerRecord(ctx, p)
	}
	for u := range onlineUUIDs {
		p, err := s.FindPlayerByUUID(ctx, serverID, u)
		if err != nil || p == nil {
			continue
		}
		changed := false
		if !p.IsOnline() {
			if p.Data == nil {
				p.Data = map[string]any{}
			}
			p.Data[PlayerKeyIsOnline] = true
			changed = true
		}
		if serverName != "" && p.LastServer() != serverName {
			if p.Data == nil {
				p.Data = map[string]any{}
			}
			p.Data[PlayerKeyLastServer] = serverName
			changed = true
		}
		if changed {
			p.LastSeenMs = nowMs
			_ = s.savePlayerRecord(ctx, p)
		}
	}
}

func (s *Store) AddUsername(ctx context.Context, serverID, rawUUID, username string) (*Player, error) {
	p, err := s.FindPlayerByUUID(ctx, serverID, rawUUID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrNotFound
	}
	now := httpx.NowMs()
	p.Usernames = append(p.Usernames, UsernameEntry{Username: username, DateMs: now})
	p.CurrentUsername = username
	p.LastSeenMs = now
	if err := s.savePlayerRecord(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Store) AddPlayerNote(ctx context.Context, serverID, rawUUID, text, issuerName, issuerID string) (*Player, error) {
	p, err := s.FindPlayerByUUID(ctx, serverID, rawUUID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrNotFound
	}
	storedIssuerName := issuerName
	if issuerID != "" {
		storedIssuerName = ""
	}
	p.Notes = append(p.Notes, NoteEntry{
		ID:         NewShortID(),
		Text:       text,
		DateMs:     httpx.NowMs(),
		IssuerName: storedIssuerName,
		IssuerID:   issuerID,
	})
	if err := s.savePlayerRecord(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Store) AddPlayerIP(ctx context.Context, serverID, rawUUID, ipAddr string) (*Player, error) {
	p, err := s.FindPlayerByUUID(ctx, serverID, rawUUID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrNotFound
	}
	now := httpx.NowMs()
	found := false
	for i := range p.IPAddresses {
		if p.IPAddresses[i].IPAddress == ipAddr {
			p.IPAddresses[i].LoginsMs = append(p.IPAddresses[i].LoginsMs, now)
			found = true
			break
		}
	}
	if !found {
		p.IPAddresses = append(p.IPAddresses, IPEntry{
			IPAddress:    ipAddr,
			FirstLoginMs: now,
			LoginsMs:     []int64{now},
		})
	}
	s.upsertPlayerIPIndex(ctx, serverID, p.ID, p.MinecraftUUID, ipAddr, now)
	if err := s.savePlayerRecord(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Store) SubmitIPInfo(ctx context.Context, serverID, rawUUID, ipAddr, country, region, asn string, proxy, hosting bool) error {
	p, err := s.FindPlayerByUUID(ctx, serverID, rawUUID)
	if err != nil || p == nil {
		return err
	}
	updated := false
	for i := range p.IPAddresses {
		if p.IPAddresses[i].IPAddress == ipAddr {
			if country != "" {
				c := country
				p.IPAddresses[i].Country = &c
			}
			if region != "" {
				r := region
				p.IPAddresses[i].Region = &r
			}
			if asn != "" {
				a := asn
				p.IPAddresses[i].ASN = &a
			}
			p.IPAddresses[i].Proxy = proxy
			p.IPAddresses[i].Hosting = hosting
			updated = true
			break
		}
	}
	if !updated {
		now := httpx.NowMs()
		entry := IPEntry{
			IPAddress:    ipAddr,
			Proxy:        proxy,
			Hosting:      hosting,
			FirstLoginMs: now,
			LoginsMs:     []int64{now},
		}
		if country != "" {
			c := country
			entry.Country = &c
		}
		if region != "" {
			r := region
			entry.Region = &r
		}
		if asn != "" {
			a := asn
			entry.ASN = &a
		}
		p.IPAddresses = append(p.IPAddresses, entry)
	}
	if p.Data == nil {
		p.Data = map[string]any{}
	}
	if country != "" {
		p.Data[PlayerKeyCountry] = country
	}
	if region != "" {
		p.Data[PlayerKeyRegion] = region
	}
	if asn != "" {
		p.Data[PlayerKeyASN] = asn
	}
	p.Data[PlayerKeyProxy] = proxy
	p.Data[PlayerKeyHosting] = hosting
	return s.savePlayerRecord(ctx, p)
}

func (s *Store) FindPlayersByIPAddresses(ctx context.Context, serverID string, ips []string, excludeUUID string, limit int) ([]*Player, error) {
	if len(ips) == 0 {
		return nil, nil
	}
	ph := make([]string, len(ips))
	args := []any{serverID}
	for i, ip := range ips {
		ph[i] = "?"
		args = append(args, ip)
	}
	exClause := ""
	if excludeUUID != "" {
		exClause = " AND LOWER(minecraft_uuid) != LOWER(?)"
		args = append(args, excludeUUID)
	}
	if limit <= 0 {
		limit = 200
	}
	args = append(args, limit)
	q := fmt.Sprintf(`
		SELECT DISTINCT minecraft_uuid FROM player_ips
		WHERE server_id = ? AND ip_address IN (%s)%s
		LIMIT ?`,
		strings.Join(ph, ","), exClause,
	)
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	var uuids []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err == nil {
			uuids = append(uuids, u)
		}
	}
	_ = rows.Close()
	if len(uuids) == 0 {
		return nil, nil
	}
	return s.FindPlayersByUUIDs(ctx, serverID, uuids)
}

func (s *Store) AddLinkedAccounts(ctx context.Context, serverID, rawUUID string, linkedUUIDs []string, nowMs int64) error {
	if len(linkedUUIDs) == 0 {
		return nil
	}
	p, err := s.FindPlayerByUUID(ctx, serverID, rawUUID)
	if err != nil || p == nil {
		return err
	}
	if p.Data == nil {
		p.Data = map[string]any{}
	}
	existing := p.LinkedAccountUUIDs()
	seen := make(map[string]bool, len(existing)+len(linkedUUIDs))
	merged := make([]string, 0, len(existing)+len(linkedUUIDs))
	for _, u := range existing {
		low := strings.ToLower(u)
		if !seen[low] {
			seen[low] = true
			merged = append(merged, low)
		}
	}
	for _, u := range linkedUUIDs {
		low := strings.ToLower(u)
		if low != strings.ToLower(p.MinecraftUUID) && !seen[low] {
			seen[low] = true
			merged = append(merged, low)
		}
	}
	p.Data[PlayerKeyLinkedAccounts] = merged
	p.Data[PlayerKeyLastLinkedUpdate] = nowMs
	return s.savePlayerRecord(ctx, p)
}

// --- Punishment Queries & Mutations ---

const punishmentSelectCols = `
	id, server_id, player_id, player_uuid, player_username,
	type, punishment_type_id, reason, issuer_uuid, issuer_name,
	issued_at, expires_at, duration_ms, pardoned, pardoned_by,
	pardoned_at, pardon_reason, points, ip_ban, ip_address,
	silent, wipe_stats, block_alts, notes_json, evidence_json,
	data_json, updated_at`

func scanPunishmentRow(scanner interface{ Scan(...any) error }) (*Punishment, error) {
	var (
		pun                              Punishment
		typeIDStr, reasonCol             string
		notesJSON, evJSON, dataJSON      string
		expiresAt, durationMs, pardonedAt sql.NullInt64
		pardonedInt, ipBanInt, silentInt int
		wipeInt, blockAltsInt            int
		updatedAt                        int64
	)
	err := scanner.Scan(
		&pun.ID, &pun.ServerID, &pun.PlayerID, &pun.PlayerUUID, &pun.PlayerUsername,
		&pun.Type, &typeIDStr, &reasonCol, &pun.IssuerID, &pun.IssuerName,
		&pun.IssuedMs, &expiresAt, &durationMs, &pardonedInt, &pun.PardonedBy,
		&pardonedAt, &pun.PardonReason, &pun.Points, &ipBanInt, &pun.IPAddress,
		&silentInt, &wipeInt, &blockAltsInt, &notesJSON, &evJSON,
		&dataJSON, &updatedAt,
	)
	if err != nil {
		return nil, err
	}
	if expiresAt.Valid {
		v := expiresAt.Int64
		pun.ExpiresMs = &v
	}
	if durationMs.Valid {
		v := durationMs.Int64
		pun.DurationMs = &v
	}
	if pardonedAt.Valid {
		v := pardonedAt.Int64
		pun.PardonedAtMs = &v
	}
	pun.Pardoned = pardonedInt != 0
	pun.IPBan = ipBanInt != 0
	pun.Silent = silentInt != 0
	pun.WipeStats = wipeInt != 0
	pun.BlockAlts = blockAltsInt != 0

	pun.Notes = decodePunishmentNotes(notesJSON)
	pun.Evidence = decodePunishmentEvidence(evJSON)
	rawData := decodeMap(dataJSON)

	if ord, err := strconv.Atoi(strings.TrimSpace(typeIDStr)); err == nil {
		pun.TypeOrdinal = ord
	} else if v, ok := rawData[dataMetaTypeOrdinal]; ok {
		if n, ok2 := asInt64(v); ok2 {
			pun.TypeOrdinal = int(n)
		}
	}
	if v, ok := rawData[dataMetaStarted]; ok && v != nil {
		if ms, ok2 := parseFlexibleTimeMs(v); ok2 {
			pun.StartedMs = &ms
		}
	}
	if v, ok := rawData[dataMetaModifications]; ok && v != nil {
		pun.Modifications = decodeModsFromAny(v)
	} else {
		pun.Modifications = []PunishmentModification{}
	}
	if v, ok := rawData[dataMetaAttachedTicketIds]; ok && v != nil {
		pun.AttachedTicketIds = decodeStringsFromAny(v)
	} else {
		pun.AttachedTicketIds = []string{}
	}

	delete(rawData, dataMetaStarted)
	delete(rawData, dataMetaTypeOrdinal)
	delete(rawData, dataMetaModifications)
	delete(rawData, dataMetaAttachedTicketIds)

	if reasonCol != "" {
		if _, exists := rawData[PunishmentKeyReason]; !exists {
			rawData[PunishmentKeyReason] = reasonCol
		}
	}
	if pun.DurationMs != nil && *pun.DurationMs != 0 {
		if _, exists := rawData[PunishmentKeyDuration]; !exists {
			rawData[PunishmentKeyDuration] = *pun.DurationMs
		}
	}
	if pun.BlockAlts {
		if _, exists := rawData[PunishmentKeyAltBlocking]; !exists {
			rawData[PunishmentKeyAltBlocking] = true
		}
	}
	if pun.WipeStats {
		if _, exists := rawData[PunishmentKeyWipeAfterExpiry]; !exists {
			rawData[PunishmentKeyWipeAfterExpiry] = true
		}
	}
	if pun.Pardoned {
		if _, exists := rawData[PunishmentKeyStatus]; !exists {
			rawData[PunishmentKeyStatus] = PunishmentStatusPardoned
		}
	}
	pun.Data = rawData
	return &pun, nil
}

func (s *Store) loadPunishmentsForPlayer(ctx context.Context, serverID, playerUUID string) ([]*Punishment, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT `+punishmentSelectCols+` FROM punishments
		WHERE server_id = ? AND LOWER(player_uuid) = LOWER(?)
		ORDER BY issued_at ASC`,
		serverID, playerUUID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Punishment
	for rows.Next() {
		p, err := scanPunishmentRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) attachPunishmentsToPlayers(ctx context.Context, serverID string, byUUID map[string]*Player) error {
	if len(byUUID) == 0 {
		return nil
	}
	ph := make([]string, 0, len(byUUID))
	args := make([]any, 0, len(byUUID)+1)
	args = append(args, serverID)
	for u := range byUUID {
		ph = append(ph, "?")
		args = append(args, u)
	}
	q := fmt.Sprintf(`SELECT %s FROM punishments WHERE server_id = ? AND LOWER(player_uuid) IN (%s) ORDER BY issued_at ASC`,
		punishmentSelectCols, strings.Join(ph, ","))
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		pun, err := scanPunishmentRow(rows)
		if err != nil {
			return err
		}
		if pl, ok := byUUID[strings.ToLower(pun.PlayerUUID)]; ok {
			pl.Punishments = append(pl.Punishments, pun)
		}
	}
	return rows.Err()
}

func (s *Store) buildPunishmentDataEnvelope(p *Punishment) string {
	m := make(map[string]any, len(p.Data)+4)
	for k, v := range p.Data {
		m[k] = v
	}
	m[dataMetaTypeOrdinal] = p.TypeOrdinal
	if p.StartedMs != nil {
		m[dataMetaStarted] = *p.StartedMs
	} else {
		delete(m, dataMetaStarted)
	}
	if p.Modifications == nil {
		m[dataMetaModifications] = []PunishmentModification{}
	} else {
		m[dataMetaModifications] = p.Modifications
	}
	if p.AttachedTicketIds == nil {
		m[dataMetaAttachedTicketIds] = []string{}
	} else {
		m[dataMetaAttachedTicketIds] = p.AttachedTicketIds
	}
	return mustJSON(m, "{}")
}

func (s *Store) InsertPunishment(ctx context.Context, p *Punishment) error {
	now := httpx.NowMs()
	var expiresVal any
	if p.ExpiresMs != nil {
		expiresVal = *p.ExpiresMs
	}
	var durVal any
	if p.DurationMs != nil {
		durVal = *p.DurationMs
	} else if d, ok := p.DataInt64(PunishmentKeyDuration); ok {
		durVal = d
	}
	var pardonedAtVal any
	if p.PardonedAtMs != nil {
		pardonedAtVal = *p.PardonedAtMs
	}
	reason := p.DataString(PunishmentKeyReason)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO punishments (
			id, server_id, player_id, player_uuid, player_username,
			type, punishment_type_id, reason, issuer_uuid, issuer_name,
			issued_at, expires_at, duration_ms, pardoned, pardoned_by,
			pardoned_at, pardon_reason, points, ip_ban, ip_address,
			silent, wipe_stats, block_alts, notes_json, evidence_json,
			data_json, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.ServerID, p.PlayerID, p.PlayerUUID, p.PlayerUsername,
		p.Type, strconv.Itoa(p.TypeOrdinal), reason, p.IssuerID, p.IssuerName,
		p.IssuedMs, expiresVal, durVal, boolToInt(p.Pardoned), p.PardonedBy,
		pardonedAtVal, p.PardonReason, p.Points, boolToInt(p.IPBan), p.IPAddress,
		boolToInt(p.Silent), boolToInt(p.WipeStats), boolToInt(p.BlockAlts),
		mustJSON(p.Notes, "[]"), mustJSON(p.Evidence, "[]"),
		s.buildPunishmentDataEnvelope(p), now,
	)
	return err
}

func (s *Store) UpdatePunishment(ctx context.Context, p *Punishment) error {
	now := httpx.NowMs()
	var expiresVal any
	if p.ExpiresMs != nil {
		expiresVal = *p.ExpiresMs
	}
	var durVal any
	if p.DurationMs != nil {
		durVal = *p.DurationMs
	} else if d, ok := p.DataInt64(PunishmentKeyDuration); ok {
		durVal = d
	}
	var pardonedAtVal any
	if p.PardonedAtMs != nil {
		pardonedAtVal = *p.PardonedAtMs
	}
	reason := p.DataString(PunishmentKeyReason)
	_, err := s.db.ExecContext(ctx, `
		UPDATE punishments SET
			type = ?, punishment_type_id = ?, reason = ?,
			expires_at = ?, duration_ms = ?, pardoned = ?, pardoned_by = ?,
			pardoned_at = ?, pardon_reason = ?, points = ?,
			wipe_stats = ?, block_alts = ?, notes_json = ?, evidence_json = ?,
			data_json = ?, updated_at = ?
		WHERE server_id = ? AND id = ?`,
		p.Type, strconv.Itoa(p.TypeOrdinal), reason,
		expiresVal, durVal, boolToInt(p.Pardoned), p.PardonedBy,
		pardonedAtVal, p.PardonReason, p.Points,
		boolToInt(p.WipeStats), boolToInt(p.BlockAlts),
		mustJSON(p.Notes, "[]"), mustJSON(p.Evidence, "[]"),
		s.buildPunishmentDataEnvelope(p), now,
		p.ServerID, p.ID,
	)
	return err
}

func (s *Store) FindPunishmentContext(ctx context.Context, serverID, punishmentID string) (*Player, *Punishment, error) {
	var playerUUID string
	err := s.db.QueryRowContext(ctx,
		`SELECT player_uuid FROM punishments WHERE server_id = ? AND id = ? LIMIT 1`,
		serverID, punishmentID,
	).Scan(&playerUUID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	pl, err := s.FindPlayerByUUID(ctx, serverID, playerUUID)
	if err != nil || pl == nil {
		return nil, nil, err
	}
	for _, pun := range pl.Punishments {
		if pun.ID == punishmentID {
			return pl, pun, nil
		}
	}
	return nil, nil, nil
}

func (s *Store) FindPlayersWithPunishments(ctx context.Context, serverID string, limit int) ([]*Player, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT DISTINCT player_uuid FROM punishments
		WHERE server_id = ?
		ORDER BY issued_at DESC
		LIMIT ?`,
		serverID, limit,
	)
	if err != nil {
		return nil, err
	}
	var uuids []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err == nil {
			uuids = append(uuids, u)
		}
	}
	_ = rows.Close()
	return s.FindPlayersByUUIDs(ctx, serverID, uuids)
}

func (s *Store) FindPlayersWithPunishmentsIssuedAfter(ctx context.Context, serverID string, cutoffMs int64, limit int) ([]*Player, error) {
	if limit <= 0 {
		limit = 500
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT DISTINCT player_uuid FROM punishments
		WHERE server_id = ? AND issued_at > ?
		ORDER BY issued_at DESC
		LIMIT ?`,
		serverID, cutoffMs, limit,
	)
	if err != nil {
		return nil, err
	}
	var uuids []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err == nil {
			uuids = append(uuids, u)
		}
	}
	_ = rows.Close()
	return s.FindPlayersByUUIDs(ctx, serverID, uuids)
}

func (s *Store) FindPlayersByLinkedBanID(ctx context.Context, serverID, parentPunishmentID string) ([]*Player, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT DISTINCT player_uuid FROM punishments
		WHERE server_id = ? AND json_extract(data_json, '$.linkedBanId') = ?`,
		serverID, parentPunishmentID,
	)
	if err != nil {
		return nil, err
	}
	var uuids []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err == nil {
			uuids = append(uuids, u)
		}
	}
	_ = rows.Close()
	return s.FindPlayersByUUIDs(ctx, serverID, uuids)
}

// --- Chat & Command Logs ---

func (s *Store) InsertChatLogs(ctx context.Context, serverID string, entries []SyncChatLogInput) error {
	if len(entries) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO chat_logs (id, server_id, uuid, username, message, timestamp, server_name)
		VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, e := range entries {
		norm, ok := NormalizeUUID(e.UUID)
		if !ok {
			norm = strings.ToLower(strings.TrimSpace(e.UUID))
		}
		if _, err := stmt.ExecContext(ctx, httpx.NewID(), serverID, norm, e.Username, e.Message, e.Timestamp, e.Server); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) InsertCommandLogs(ctx context.Context, serverID string, entries []SyncCommandLogInput) error {
	if len(entries) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO command_logs (id, server_id, uuid, username, command, timestamp, server_name)
		VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, e := range entries {
		norm, ok := NormalizeUUID(e.UUID)
		if !ok {
			norm = strings.ToLower(strings.TrimSpace(e.UUID))
		}
		if _, err := stmt.ExecContext(ctx, httpx.NewID(), serverID, norm, e.Username, e.Command, e.Timestamp, e.Server); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) GetChatLogs(ctx context.Context, serverID, rawUUID string, limit int) ([]ChatLogRecord, error) {
	norm, ok := NormalizeUUID(rawUUID)
	if !ok {
		norm = strings.ToLower(strings.TrimSpace(rawUUID))
	}
	if limit <= 0 {
		limit = 200
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT uuid, username, message, server_name, timestamp FROM chat_logs
		WHERE server_id = ? AND LOWER(uuid) = ?
		ORDER BY timestamp DESC LIMIT ?`,
		serverID, norm, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]ChatLogRecord, 0)
	for rows.Next() {
		var rec ChatLogRecord
		if err := rows.Scan(&rec.UUID, &rec.Username, &rec.Message, &rec.Server, &rec.Timestamp); err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

func (s *Store) GetCommandLogs(ctx context.Context, serverID, rawUUID string, limit int) ([]CommandLogRecord, error) {
	norm, ok := NormalizeUUID(rawUUID)
	if !ok {
		norm = strings.ToLower(strings.TrimSpace(rawUUID))
	}
	if limit <= 0 {
		limit = 200
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT uuid, username, command, server_name, timestamp FROM command_logs
		WHERE server_id = ? AND LOWER(uuid) = ?
		ORDER BY timestamp DESC LIMIT ?`,
		serverID, norm, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]CommandLogRecord, 0)
	for rows.Next() {
		var rec CommandLogRecord
		if err := rows.Scan(&rec.UUID, &rec.Username, &rec.Command, &rec.Server, &rec.Timestamp); err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

// --- Notifications ---

func (s *Store) GetPendingNotificationsForPlayer(ctx context.Context, serverID, rawUUID string) ([]map[string]any, error) {
	norm, ok := NormalizeUUID(rawUUID)
	if !ok {
		norm = strings.ToLower(strings.TrimSpace(rawUUID))
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, type, title, message, data_json, created_at
		FROM player_notifications
		WHERE server_id = ? AND LOWER(player_uuid) = ? AND delivered = 0
		ORDER BY created_at ASC`,
		serverID, norm,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, typ, title, msg, dataJSON string
		var createdAt int64
		if err := rows.Scan(&id, &typ, &title, &msg, &dataJSON, &createdAt); err != nil {
			return nil, err
		}
		item := map[string]any{
			"id":               id,
			"type":             typ,
			"message":          msg,
			"timestamp":        createdAt,
			"targetPlayerUuid": norm,
		}
		if title != "" {
			item["title"] = title
		}
		if d := decodeMap(dataJSON); len(d) > 0 {
			item["data"] = d
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) AcknowledgeNotifications(ctx context.Context, serverID, rawUUID string, ids []string) (int, error) {
	norm, ok := NormalizeUUID(rawUUID)
	if !ok {
		norm = strings.ToLower(strings.TrimSpace(rawUUID))
	}
	now := httpx.NowMs()
	count := 0
	if len(ids) > 0 {
		ph := make([]string, len(ids))
		args := make([]any, 0, len(ids)+3)
		args = append(args, now, serverID, norm)
		for i, id := range ids {
			ph[i] = "?"
			args = append(args, id)
		}
		res, err := s.db.ExecContext(ctx,
			fmt.Sprintf(`UPDATE player_notifications SET delivered = 1, read_at = ? WHERE server_id = ? AND LOWER(player_uuid) = ? AND delivered = 0 AND id IN (%s)`, strings.Join(ph, ",")),
			args...,
		)
		if err != nil {
			return 0, err
		}
		n, _ := res.RowsAffected()
		count += int(n)
	}
	// Also clear any legacy pendingNotifications stored inside player data_json
	pl, err := s.FindPlayerByUUID(ctx, serverID, norm)
	if err == nil && pl != nil {
		legacy := pl.LegacyPendingNotifications()
		if len(legacy) > 0 {
			idSet := make(map[string]bool, len(ids))
			for _, id := range ids {
				idSet[id] = true
			}
			var remaining []map[string]any
			for _, item := range legacy {
				itemID, _ := item["id"].(string)
				if itemID != "" && idSet[itemID] {
					count++
					continue
				}
				remaining = append(remaining, item)
			}
			if remaining == nil {
				remaining = []map[string]any{}
			}
			pl.Data[PlayerKeyPendingNotifications] = remaining
			_ = s.savePlayerRecord(ctx, pl)
		}
	}
	return count, nil
}

// --- Tickets & Appeals ---

func (s *Store) FindReportedPlayerTickets(ctx context.Context, serverID, playerUUID string, limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, type, creator_name, creator_uuid, reported_player, reported_player_uuid,
		       subject, status, priority, assigned_to_json, chat_messages_json, data_json, created_at
		FROM tickets
		WHERE server_id = ? AND LOWER(reported_player_uuid) = LOWER(?) AND type IN ('player', 'chat')
		ORDER BY created_at DESC LIMIT ?`,
		serverID, playerUUID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var (
			id, typ, creatorName, creatorUUID, repPlayer, repUUID string
			subject, status, priority, assignedJSON, chatJSON, dataJSON string
			createdAt                                             int64
		)
		if err := rows.Scan(&id, &typ, &creatorName, &creatorUUID, &repPlayer, &repUUID,
			&subject, &status, &priority, &assignedJSON, &chatJSON, &dataJSON, &createdAt); err != nil {
			return nil, err
		}
		var assigned []string
		_ = json.Unmarshal([]byte(assignedJSON), &assigned)
		var chatMsgs []map[string]any
		_ = json.Unmarshal([]byte(chatJSON), &chatMsgs)
		var content string
		d := decodeMap(dataJSON)
		if c, ok := d["content"].(string); ok {
			content = c
		}
		out = append(out, map[string]any{
			"id":                 id,
			"type":               typ,
			"category":           typ,
			"reporterName":       creatorName,
			"reporterUuid":       creatorUUID,
			"reportedPlayerName": repPlayer,
			"reportedPlayerUuid": repUUID,
			"subject":            subject,
			"content":            content,
			"status":             status,
			"priority":           priority,
			"assignedTo":         assigned,
			"chatMessages":       chatMsgs,
			"createdAt":          createdAt,
		})
	}
	return out, rows.Err()
}

func (s *Store) FindAppealsByPunishmentID(ctx context.Context, serverID, punishmentID string) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, status, appeal_workflow_status, locked, created_at
		FROM tickets
		WHERE server_id = ? AND type = 'appeal' AND punishment_id = ?
		ORDER BY created_at DESC`,
		serverID, punishmentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id, status, wfStatus string
		var lockedInt int
		var createdAt int64
		if err := rows.Scan(&id, &status, &wfStatus, &lockedInt, &createdAt); err != nil {
			return nil, err
		}
		m := map[string]any{
			"id":            id,
			"submittedDate": createdAt,
			"status":        strings.ToLower(status),
			"locked":        lockedInt != 0,
		}
		if wfStatus != "" {
			m["appealWorkflowStatus"] = wfStatus
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) FindRecentTicketsAfter(ctx context.Context, serverID string, afterMs int64, limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, type, subject, creator_name, reported_player, replies_json, data_json, created_at
		FROM tickets
		WHERE server_id = ? AND status != 'unfinished' AND created_at > ?
		ORDER BY created_at DESC LIMIT ?`,
		serverID, afterMs, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id, typ, subject, creatorName, repPlayer, repliesJSON, dataJSON string
		var createdAt int64
		if err := rows.Scan(&id, &typ, &subject, &creatorName, &repPlayer, &repliesJSON, &dataJSON, &createdAt); err != nil {
			return nil, err
		}
		var replies []map[string]any
		_ = json.Unmarshal([]byte(repliesJSON), &replies)
		firstReply := ""
		if len(replies) > 0 {
			if c, ok := replies[0]["content"].(string); ok {
				firstReply = c
			}
		}
		d := decodeMap(dataJSON)
		createdServer, _ := d["createdServer"].(string)
		out = append(out, map[string]any{
			"id":                id,
			"type":              typ,
			"subject":           subject,
			"creatorName":       creatorName,
			"reportedPlayer":    repPlayer,
			"firstReplyContent": firstReply,
			"createdServer":     createdServer,
			"createdAt":         createdAt,
		})
	}
	return out, rows.Err()
}

func (s *Store) CloseTicketsForPunishment(ctx context.Context, serverID string, ticketIDs []string, issuerName string) {
	if len(ticketIDs) == 0 {
		return
	}
	now := httpx.NowMs()
	for _, tid := range ticketIDs {
		_, _ = s.db.ExecContext(ctx,
			`UPDATE tickets SET status = 'closed', locked = 1, updated_at = ? WHERE server_id = ? AND id = ?`,
			now, serverID, tid,
		)
	}
}

func (s *Store) ReopenTicketsForPunishment(ctx context.Context, serverID string, ticketIDs []string, issuerName string) {
	if len(ticketIDs) == 0 {
		return
	}
	now := httpx.NowMs()
	for _, tid := range ticketIDs {
		_, _ = s.db.ExecContext(ctx,
			`UPDATE tickets SET status = 'open', locked = 0, updated_at = ? WHERE server_id = ? AND id = ?`,
			now, serverID, tid,
		)
	}
}

func (s *Store) ApplyAppealOutcome(ctx context.Context, serverID, appealTicketID, punishmentID, outcome string) {
	if appealTicketID == "" || outcome == "" {
		return
	}
	now := httpx.NowMs()
	_, _ = s.db.ExecContext(ctx,
		`UPDATE tickets SET appeal_workflow_status = ?, status = 'closed', locked = 1, updated_at = ?
		 WHERE server_id = ? AND id = ? AND (punishment_id = ? OR punishment_id = '')`,
		outcome, now, serverID, appealTicketID, punishmentID,
	)
}

// --- Evidence Upload Tokens ---

func (s *Store) CreateEvidenceUploadToken(ctx context.Context, serverID, punishmentID, playerUUID, issuerName string) (string, error) {
	rawToken := uuid.NewString()
	sum := sha256.Sum256([]byte(rawToken))
	hash := hex.EncodeToString(sum[:])
	now := time.Now().UTC()
	exp := now.Add(30 * time.Minute)
	if issuerName == "" {
		issuerName = "Unknown"
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO evidence_upload_tokens (token_hash, server_id, punishment_id, player_uuid, issuer_name, created_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		hash, serverID, punishmentID, playerUUID, issuerName, now.UnixMilli(), exp.UnixMilli(),
	)
	if err != nil {
		return "", err
	}
	return rawToken, nil
}

// --- Staff / Issuer / Server / Sync Helpers ---

func (s *Store) ResolveIssuerNames(ctx context.Context, serverID string, issuerIDs []string) map[string]string {
	out := make(map[string]string)
	if len(issuerIDs) == 0 {
		return out
	}
	uniq := make([]any, 0, len(issuerIDs))
	ph := make([]string, 0, len(issuerIDs))
	seen := make(map[string]bool)
	for _, id := range issuerIDs {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		uniq = append(uniq, id)
		ph = append(ph, "?")
	}
	if len(uniq) == 0 {
		return out
	}
	args := append([]any{serverID}, uniq...)
	q := fmt.Sprintf(`SELECT id, username, assigned_minecraft_username, email FROM staff WHERE server_id = ? AND id IN (%s)`,
		strings.Join(ph, ","))
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id, uname, mcUname, email string
			if err := rows.Scan(&id, &uname, &mcUname, &email); err == nil {
				name := uname
				if name == "" {
					name = mcUname
				}
				if name == "" {
					name = email
				}
				if name != "" {
					out[id] = name
				}
			}
		}
	}
	return out
}

func (s *Store) ResolveIssuerName(ctx context.Context, serverID, issuerID, fallback string) string {
	if issuerID != "" {
		m := s.ResolveIssuerNames(ctx, serverID, []string{issuerID})
		if v, ok := m[issuerID]; ok && v != "" {
			return v
		}
		if fallback != "" {
			return fallback
		}
		return "Unknown Staff"
	}
	if fallback != "" {
		return fallback
	}
	return "Console"
}

func (s *Store) GetActiveStaffMembers(ctx context.Context, serverID, adminEmail string, onlinePlayerIPs map[string]string) ([]map[string]any, error) {
	rolePerms := make(map[string]struct {
		name  string
		perms []string
	})
	roleRows, err := s.db.QueryContext(ctx, `SELECT id, name, permissions_json FROM staff_roles WHERE server_id = ?`, serverID)
	if err == nil {
		for roleRows.Next() {
			var rid, rname, pjson string
			if err := roleRows.Scan(&rid, &rname, &pjson); err == nil {
				var ps []string
				_ = json.Unmarshal([]byte(pjson), &ps)
				if ps == nil {
					ps = []string{}
				}
				rolePerms[rid] = struct {
					name  string
					perms []string
				}{name: rname, perms: ps}
			}
		}
		_ = roleRows.Close()
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, email, username, role_id, assigned_minecraft_uuid, assigned_minecraft_username,
		       two_factor_session_ip, COALESCE(two_factor_session_expires_at, 0)
		FROM staff
		WHERE server_id = ? AND assigned_minecraft_uuid != ''`,
		serverID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nowMs := httpx.NowMs()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, email, uname, roleID, mcUUID, mcUname, sessIP string
		var sessExp int64
		if err := rows.Scan(&id, &email, &uname, &roleID, &mcUUID, &mcUname, &sessIP, &sessExp); err != nil {
			return nil, err
		}
		if adminEmail != "" && strings.EqualFold(email, adminEmail) {
			roleID = "super-admin"
		}
		rl := rolePerms[roleID]
		perms := rl.perms
		if perms == nil {
			perms = []string{}
		}
		roleName := rl.name
		if roleName == "" {
			roleName = roleID
		}
		normUUID, ok := NormalizeUUID(mcUUID)
		if !ok {
			normUUID = strings.ToLower(mcUUID)
		}
		curIP := onlinePlayerIPs[normUUID]
		valid2FA := sessExp > nowMs && sessIP != "" && curIP != "" && sessIP == curIP
		out = append(out, map[string]any{
			"minecraftUuid":         normUUID,
			"minecraftUsername":     mcUname,
			"staffUsername":         uname,
			"staffId":               id,
			"staffRole":             roleName,
			"permissions":           perms,
			"email":                 email,
			"twoFactorSessionValid": valid2FA,
		})
	}
	return out, rows.Err()
}

func (s *Store) CollectAndClearPending2FA(ctx context.Context, serverID string) []map[string]any {
	rows, err := s.db.QueryContext(ctx,
		`SELECT assigned_minecraft_uuid FROM staff WHERE server_id = ? AND two_factor_pending_delivery = 1 AND assigned_minecraft_uuid != ''`,
		serverID,
	)
	if err != nil {
		return nil
	}
	var out []map[string]any
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err == nil && u != "" {
			out = append(out, map[string]any{"minecraftUuid": u})
		}
	}
	_ = rows.Close()
	if len(out) > 0 {
		_, _ = s.db.ExecContext(ctx,
			`UPDATE staff SET two_factor_pending_delivery = 0, updated_at = ? WHERE server_id = ? AND two_factor_pending_delivery = 1`,
			httpx.NowMs(), serverID,
		)
	}
	return out
}

func (s *Store) GetActiveMigrationTask(ctx context.Context, serverID string) map[string]any {
	var taskID, typ string
	err := s.db.QueryRowContext(ctx,
		`SELECT task_id, type FROM migrations WHERE server_id = ? AND status IN ('PENDING', 'RUNNING', 'IN_PROGRESS') ORDER BY started_at DESC LIMIT 1`,
		serverID,
	).Scan(&taskID, &typ)
	if err != nil || taskID == "" {
		return nil
	}
	return map[string]any{"taskId": taskID, "type": typ}
}

func (s *Store) GetServerUpdatedTimestamps(ctx context.Context, serverID string) (staffPermsUpdatedAt, punishmentTypesUpdatedAt *int64) {
	var sp, pt sql.NullInt64
	err := s.db.QueryRowContext(ctx,
		`SELECT staff_permissions_updated_at, punishment_types_updated_at FROM servers WHERE id = ? LIMIT 1`,
		serverID,
	).Scan(&sp, &pt)
	if err != nil {
		return nil, nil
	}
	if sp.Valid {
		v := sp.Int64
		staffPermsUpdatedAt = &v
	}
	if pt.Valid {
		v := pt.Int64
		punishmentTypesUpdatedAt = &v
	}
	return staffPermsUpdatedAt, punishmentTypesUpdatedAt
}

func (s *Store) UpdateServerActivity(ctx context.Context, serverID string, nowMs int64, onlineCount *int) {
	if onlineCount != nil {
		_, _ = s.db.ExecContext(ctx,
			`UPDATE servers SET last_activity_at = ?, online_player_count = ?, updated_at = ? WHERE id = ?`,
			nowMs, *onlineCount, nowMs, serverID,
		)
		return
	}
	_, _ = s.db.ExecContext(ctx,
		`UPDATE servers SET last_activity_at = ?, updated_at = ? WHERE id = ?`,
		nowMs, nowMs, serverID,
	)
}

func (s *Store) UpsertServerInstanceSnapshot(ctx context.Context, serverID, serverName string, onlineCount int, platformType, serverVersion, clientIP, pluginVersion string, nowMs int64) {
	boundaryMs := (nowMs / 300000) * 300000
	entry := map[string]any{
		"serverName":        serverName,
		"onlinePlayerCount": onlineCount,
		"platformType":      platformType,
		"serverVersion":     serverVersion,
		"clientIp":          clientIP,
		"pluginVersion":     pluginVersion,
		"lastHeartbeat":     nowMs,
	}
	var existingID, existingJSON string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, servers_json FROM server_instance_snapshots WHERE server_id = ? AND date_ms = ? LIMIT 1`,
		serverID, boundaryMs,
	).Scan(&existingID, &existingJSON)
	if errors.Is(err, sql.ErrNoRows) {
		_, _ = s.db.ExecContext(ctx,
			`INSERT INTO server_instance_snapshots (id, server_id, date_ms, servers_json, created_at) VALUES (?, ?, ?, ?, ?)`,
			httpx.NewID(), serverID, boundaryMs, mustJSON([]map[string]any{entry}, "[]"), nowMs,
		)
		return
	}
	if err != nil {
		return
	}
	var list []map[string]any
	_ = json.Unmarshal([]byte(existingJSON), &list)
	replaced := false
	for i := range list {
		if sn, _ := list[i]["serverName"].(string); sn == serverName {
			list[i] = entry
			replaced = true
			break
		}
	}
	if !replaced {
		list = append(list, entry)
	}
	_, _ = s.db.ExecContext(ctx,
		`UPDATE server_instance_snapshots SET servers_json = ? WHERE id = ?`,
		mustJSON(list, "[]"), existingID,
	)
}

func (s *Store) RecordModerationLog(ctx context.Context, serverID, source, description string) {
	if serverID == "" {
		return
	}
	_, _ = s.db.ExecContext(ctx,
		`INSERT INTO server_logs (id, server_id, description, level, source, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		httpx.NewID(), serverID, description, "moderation", source, httpx.NowMs(),
	)
}

// --- JSON / Scan Helpers ---

func escapeLike(v string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(strings.ToLower(v))
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func mustJSON(v any, def string) string {
	if v == nil {
		return def
	}
	b, err := json.Marshal(v)
	if err != nil || string(b) == "null" {
		return def
	}
	return string(b)
}

func decodeMap(raw string) map[string]any {
	m := make(map[string]any)
	if strings.TrimSpace(raw) == "" || raw == "null" {
		return m
	}
	_ = json.Unmarshal([]byte(raw), &m)
	if m == nil {
		return map[string]any{}
	}
	return m
}

func decodeUsernames(raw string) []UsernameEntry {
	var arr []map[string]any
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		return []UsernameEntry{}
	}
	out := make([]UsernameEntry, 0, len(arr))
	for _, m := range arr {
		u, _ := m["username"].(string)
		ms, _ := parseFlexibleTimeMs(m["date"])
		out = append(out, UsernameEntry{Username: u, DateMs: ms})
	}
	return out
}

func decodeNotes(raw string) []NoteEntry {
	var arr []map[string]any
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		return []NoteEntry{}
	}
	out := make([]NoteEntry, 0, len(arr))
	for _, m := range arr {
		id, _ := m["id"].(string)
		text, _ := m["text"].(string)
		in, _ := m["issuerName"].(string)
		iid, _ := m["issuerId"].(string)
		ms, _ := parseFlexibleTimeMs(m["date"])
		out = append(out, NoteEntry{ID: id, Text: text, DateMs: ms, IssuerName: in, IssuerID: iid})
	}
	return out
}

func decodeIPs(raw string) []IPEntry {
	var arr []map[string]any
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		return []IPEntry{}
	}
	out := make([]IPEntry, 0, len(arr))
	for _, m := range arr {
		ipAddr, _ := m["ipAddress"].(string)
		e := IPEntry{IPAddress: ipAddr}
		if v, ok := m["country"].(string); ok && v != "" {
			e.Country = &v
		}
		if v, ok := m["region"].(string); ok && v != "" {
			e.Region = &v
		}
		if v, ok := m["asn"].(string); ok && v != "" {
			e.ASN = &v
		}
		e.Proxy, _ = m["proxy"].(bool)
		e.Hosting, _ = m["hosting"].(bool)
		e.FirstLoginMs, _ = parseFlexibleTimeMs(m["firstLogin"])
		if logins, ok := m["logins"].([]any); ok {
			for _, l := range logins {
				if ms, ok := parseFlexibleTimeMs(l); ok {
					e.LoginsMs = append(e.LoginsMs, ms)
				}
			}
		}
		if e.LoginsMs == nil {
			e.LoginsMs = []int64{}
		}
		out = append(out, e)
	}
	return out
}

func decodePunishmentNotes(raw string) []PunishmentNote {
	var arr []map[string]any
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		return []PunishmentNote{}
	}
	out := make([]PunishmentNote, 0, len(arr))
	for _, m := range arr {
		id, _ := m["id"].(string)
		text, _ := m["text"].(string)
		in, _ := m["issuerName"].(string)
		iid, _ := m["issuerId"].(string)
		ms, _ := parseFlexibleTimeMs(m["date"])
		out = append(out, PunishmentNote{ID: id, Text: text, DateMs: ms, IssuerName: in, IssuerID: iid})
	}
	return out
}

func decodePunishmentEvidence(raw string) []PunishmentEvidence {
	var arr []map[string]any
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		return []PunishmentEvidence{}
	}
	out := make([]PunishmentEvidence, 0, len(arr))
	for _, m := range arr {
		typ, _ := m["type"].(string)
		ms, _ := parseFlexibleTimeMs(m["uploadedAt"])
		ev := PunishmentEvidence{
			Type:         typ,
			UploadedAtMs: ms,
		}
		if v, ok := m["text"].(string); ok && v != "" {
			ev.Text = &v
		}
		if v, ok := m["url"].(string); ok && v != "" {
			ev.URL = &v
		}
		if v, ok := m["uploadedBy"].(string); ok && v != "" {
			ev.UploadedBy = &v
		}
		if v, ok := m["uploadedById"].(string); ok && v != "" {
			ev.UploadedByID = &v
		}
		if v, ok := m["fileName"].(string); ok && v != "" {
			ev.FileName = &v
		}
		if v, ok := m["fileType"].(string); ok && v != "" {
			ev.FileType = &v
		}
		if v, ok := m["fileSize"]; ok && v != nil {
			if sz, ok2 := asInt64(v); ok2 {
				ev.FileSize = &sz
			}
		}
		out = append(out, ev)
	}
	return out
}

func decodeModsFromAny(v any) []PunishmentModification {
	arr, ok := v.([]any)
	if !ok {
		return []PunishmentModification{}
	}
	out := make([]PunishmentModification, 0, len(arr))
	for _, item := range arr {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		mod := PunishmentModification{
			ID:             strFromMap(m, "id"),
			Type:           strFromMap(m, "type"),
			IssuerName:     strFromMap(m, "issuerName"),
			IssuerID:       strFromMap(m, "issuerId"),
			Reason:         strFromMap(m, "reason"),
			AppealTicketID: strFromMap(m, "appealTicketId"),
		}
		mod.DateMs, _ = parseFlexibleTimeMs(m["date"])
		if eff, ok := m["effectiveDuration"]; ok && eff != nil {
			if val, ok2 := asInt64(eff); ok2 {
				mod.EffectiveDuration = &val
			}
		}
		if d, ok := m["data"].(map[string]any); ok {
			mod.Data = d
		}
		out = append(out, mod)
	}
	return out
}

func decodeStringsFromAny(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		return []string{}
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	return out
}

func strFromMap(m map[string]any, key string) string {
	s, _ := m[key].(string)
	return s
}

func asInt64(v any) (int64, bool) {
	switch t := v.(type) {
	case int64:
		return t, true
	case int:
		return int64(t), true
	case float64:
		return int64(t), true
	case string:
		n, err := strconv.ParseInt(strings.TrimSpace(t), 10, 64)
		return n, err == nil
	}
	return 0, false
}

func SortPunishmentsNewestFirst(puns []*Punishment) []*Punishment {
	cp := make([]*Punishment, len(puns))
	copy(cp, puns)
	sort.SliceStable(cp, func(i, j int) bool {
		return cp[i].IssuedMs > cp[j].IssuedMs
	})
	return cp
}

