package migration

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"modl-go/internal/httpx"
	"modl-go/internal/models"
)

const (
	batchSize              = 500
	progressUpdateInterval = 1000
	maxFailureMsgLen       = 900
)

type playerRecord struct {
	MinecraftUUID string                   `json:"minecraftUuid"`
	Usernames     []map[string]any         `json:"usernames"`
	Notes         []map[string]any         `json:"notes"`
	IPAddresses   []map[string]any         `json:"ipAddresses"`
	Punishments   []map[string]any         `json:"punishments"`
	Data          map[string]any           `json:"data"`
}

func (s *Service) ProcessFile(ctx context.Context, srv *models.Server, filePath string) {
	defer func() {
		_ = os.Remove(filePath)
	}()

	_, _ = s.UpdateProgress(ctx, srv, "processing_data", "Reading and validating migration file...", int32Ptr(0), int32Ptr(0), nil)

	processed, skipped, total, err := s.streamMigrationFile(ctx, srv, filePath)
	if err != nil {
		if errors.Is(err, errMigrationCancelled) {
			return
		}
		msg := truncate("Migration failed: "+err.Error(), maxFailureMsgLen)
		p := processed
		sk := skipped
		_, _ = s.UpdateProgress(ctx, srv, "failed", msg, &p, &sk, total)
		return
	}

	p := processed
	sk := skipped
	_, _ = s.UpdateProgress(ctx, srv, "completed", "Migration completed successfully", &p, &sk, total)
}

var errMigrationCancelled = errors.New("migration cancelled")

func (s *Service) streamMigrationFile(ctx context.Context, srv *models.Server, filePath string) (int32, int32, *int32, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, 0, nil, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	dec.UseNumber()

	tok, err := dec.Token()
	if err != nil || tok != json.Delim('{') {
		return 0, 0, nil, errors.New("migration data must be a JSON object")
	}

	var hasPlayers, isArray, playersStreamed bool
	var declaredTotal *int32
	var processed, skipped int32

	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return processed, skipped, declaredTotal, err
		}
		key, _ := keyTok.(string)

		switch key {
		case "metadata":
			var meta map[string]any
			if err := dec.Decode(&meta); err != nil {
				return processed, skipped, declaredTotal, err
			}
			if rawCnt, ok := meta["playerCount"]; ok {
				if cnt, ok := asInt64(rawCnt); ok {
					v := int32(cnt)
					if cnt > int64(^uint32(0)>>1) {
						v = int32(^uint32(0) >> 1)
					}
					declaredTotal = &v
				}
			}
		case "players":
			hasPlayers = true
			arrTok, err := dec.Token()
			if err != nil {
				return processed, skipped, declaredTotal, err
			}
			if arrTok != json.Delim('[') {
				return processed, skipped, declaredTotal, errors.New("migration data must contain a 'players' array")
			}
			isArray = true
			if msg := validateHeader(hasPlayers, isArray, declaredTotal); msg != "" {
				return processed, skipped, declaredTotal, errors.New(msg)
			}

			s.announceProcessing(ctx, srv, processed, skipped, declaredTotal)

			var seen int64
			batch := make([]playerRecord, 0, batchSize)
			lastAnnounced := int32(0)

			for dec.More() {
				seen++
				if seen > MaxPlayerRecords {
					return processed, skipped, declaredTotal, errors.New("players array exceeds maximum length of 1,000,000")
				}
				var rec playerRecord
				if err := dec.Decode(&rec); err != nil {
					skipped++
					continue
				}
				batch = append(batch, rec)
				if len(batch) >= batchSize {
					p, sk := s.processBatch(ctx, srv.ID, batch)
					processed += p
					skipped += sk
					batch = batch[:0]
					if processed-lastAnnounced >= progressUpdateInterval {
						s.announceProcessing(ctx, srv, processed, skipped, declaredTotal)
						lastAnnounced = processed
					}
					if !s.IsActiveMigrationPresent(ctx, srv.ID) {
						return processed, skipped, declaredTotal, errMigrationCancelled
					}
				}
			}

			// Consume closing ']'
			if _, err := dec.Token(); err != nil && !errors.Is(err, io.EOF) {
				return processed, skipped, declaredTotal, err
			}
			if seen == 0 {
				return processed, skipped, declaredTotal, errors.New("players array cannot be empty")
			}
			if len(batch) > 0 {
				p, sk := s.processBatch(ctx, srv.ID, batch)
				processed += p
				skipped += sk
				if !s.IsActiveMigrationPresent(ctx, srv.ID) {
					return processed, skipped, declaredTotal, errMigrationCancelled
				}
			}
			playersStreamed = true
		default:
			var discard any
			if err := dec.Decode(&discard); err != nil {
				return processed, skipped, declaredTotal, err
			}
		}
	}

	if !playersStreamed {
		msg := validateHeader(hasPlayers, isArray, declaredTotal)
		if msg == "" {
			msg = "Migration data must contain a 'players' array"
		}
		return processed, skipped, declaredTotal, errors.New(msg)
	}

	return processed, skipped, declaredTotal, nil
}

func (s *Service) announceProcessing(ctx context.Context, srv *models.Server, processed, skipped int32, total *int32) {
	msg := fmt.Sprintf("Processing player records... (%d)", processed)
	if total != nil {
		msg = fmt.Sprintf("Processing player records... (%d/%d)", processed, *total)
	}
	p := processed
	sk := skipped
	_, _ = s.UpdateProgress(ctx, srv, "processing_data", msg, &p, &sk, total)
}

func (s *Service) processBatch(ctx context.Context, serverID string, batch []playerRecord) (processed, skipped int32) {
	nowMs := time.Now().UnixMilli()
	for _, rec := range batch {
		uuid := normalizeUUID(rec.MinecraftUUID)
		if !isValidUUID(uuid) {
			skipped++
			continue
		}
		if err := s.upsertPlayerRecord(ctx, serverID, uuid, rec, nowMs); err != nil {
			skipped++
			continue
		}
		processed++
	}
	return processed, skipped
}

func (s *Service) upsertPlayerRecord(ctx context.Context, serverID, uuid string, rec playerRecord, nowMs int64) error {
	usernames := capSlice(rec.Usernames, MaxUsernames)
	notes := capSlice(rec.Notes, MaxNotesPerPlayer)
	ips := capSlice(rec.IPAddresses, MaxIPsPerPlayer)
	punishments := capSlice(rec.Punishments, MaxPunishments)

	var existingID, existingUsernamesJSON, existingNotesJSON, existingIPsJSON, existingDataJSON string
	var existingFirstSeen, existingLastSeen int64
	err := s.db.QueryRowContext(ctx,
		`SELECT id, usernames_json, notes_json, ip_addresses_json, data_json, first_seen, last_seen
		 FROM players WHERE server_id = ? AND LOWER(minecraft_uuid) = LOWER(?) LIMIT 1`,
		serverID, uuid,
	).Scan(&existingID, &existingUsernamesJSON, &existingNotesJSON, &existingIPsJSON, &existingDataJSON, &existingFirstSeen, &existingLastSeen)

	playerID := existingID
	if err == sql.ErrNoRows {
		playerID = httpx.NewID()
		latestUsername := extractUsername(usernames)
		uBytes, _ := json.Marshal(nonNilSlice(usernames))
		nBytes, _ := json.Marshal(nonNilSlice(notes))
		ipBytes, _ := json.Marshal(nonNilSlice(ips))
		dMap := rec.Data
		if dMap == nil {
			dMap = map[string]any{}
		}
		dBytes, _ := json.Marshal(dMap)

		if _, insErr := s.db.ExecContext(ctx,
			`INSERT INTO players (id, server_id, minecraft_uuid, current_username, usernames_json, notes_json, ip_addresses_json, data_json, first_seen, last_seen)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			playerID, serverID, uuid, latestUsername, string(uBytes), string(nBytes), string(ipBytes), string(dBytes), nowMs, nowMs,
		); insErr != nil {
			return insErr
		}
	} else if err != nil {
		return err
	} else {
		mergedUsernames := mergeByKey(existingUsernamesJSON, usernames, func(m map[string]any) string {
			u, _ := m["username"].(string)
			d, _ := m["date"].(string)
			return strings.ToLower(u) + "|" + d
		}, MaxUsernames)
		mergedNotes := mergeByKey(existingNotesJSON, notes, func(m map[string]any) string {
			t, _ := m["text"].(string)
			d, _ := m["date"].(string)
			iss, _ := m["issuerName"].(string)
			return t + "|" + d + "|" + iss
		}, MaxNotesPerPlayer)
		mergedIPs := mergeByKey(existingIPsJSON, ips, func(m map[string]any) string {
			ip, _ := m["ipAddress"].(string)
			return ip
		}, MaxIPsPerPlayer)

		mergedData := map[string]any{}
		if strings.TrimSpace(existingDataJSON) != "" && existingDataJSON != "{}" {
			_ = json.Unmarshal([]byte(existingDataJSON), &mergedData)
		}
		for k, v := range rec.Data {
			if _, exists := mergedData[k]; !exists {
				mergedData[k] = v
			}
		}

		latestUsername := extractUsername(mergedUsernames)
		uBytes, _ := json.Marshal(mergedUsernames)
		nBytes, _ := json.Marshal(mergedNotes)
		ipBytes, _ := json.Marshal(mergedIPs)
		dBytes, _ := json.Marshal(mergedData)

		if _, updErr := s.db.ExecContext(ctx,
			`UPDATE players
			 SET current_username = COALESCE(NULLIF(?, ''), current_username),
			     usernames_json = ?, notes_json = ?, ip_addresses_json = ?, data_json = ?, last_seen = ?
			 WHERE id = ? AND server_id = ?`,
			latestUsername, string(uBytes), string(nBytes), string(ipBytes), string(dBytes), nowMs, playerID, serverID,
		); updErr != nil {
			return updErr
		}
	}

	// Upsert punishments into punishments table
	playerName := extractUsername(usernames)
	for _, p := range punishments {
		pID, _ := p["id"].(string)
		if strings.TrimSpace(pID) == "" {
			pID = httpx.NewID()
		}
		var exists int
		if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM punishments WHERE server_id = ? AND id = ?`, serverID, pID).Scan(&exists); err == nil && exists > 0 {
			continue
		}

		typeStr, _ := p["type"].(string)
		typeOrd := 0
		if v, ok := asInt64(p["typeOrdinal"]); ok {
			typeOrd = int(v)
		}
		if typeStr == "" {
			typeStr = strconv.Itoa(typeOrd)
		}
		reason, _ := p["reason"].(string)
		issuerName, _ := p["issuerName"].(string)
		issuedAt := parseTimestampMs(p["issued"], nowMs)
		startedAt := parseTimestampMs(p["started"], issuedAt)

		var durPtr sql.NullInt64
		var expPtr sql.NullInt64
		if dur, ok := asInt64(p["duration"]); ok {
			durPtr = sql.NullInt64{Int64: dur, Valid: true}
			if dur > 0 {
				expPtr = sql.NullInt64{Int64: startedAt + dur, Valid: true}
			}
		}

		notesB, _ := json.Marshal(p["notes"])
		if string(notesB) == "null" {
			notesB = []byte("[]")
		}
		evB, _ := json.Marshal(p["evidence"])
		if string(evB) == "null" {
			evB = []byte("[]")
		}

		dataMap := map[string]any{}
		if d, ok := p["data"].(map[string]any); ok {
			for k, v := range d {
				dataMap[k] = v
			}
		}
		dataMap["typeOrdinal"] = typeOrd
		if startedAt > 0 {
			dataMap["started"] = startedAt
		}
		if mods, ok := p["modifications"]; ok {
			dataMap["modifications"] = mods
		}
		if tickets, ok := p["attachedTicketIds"]; ok {
			dataMap["attachedTicketIds"] = tickets
		}
		dataB, _ := json.Marshal(dataMap)

		pardoned := 0
		if st, ok := dataMap["status"].(string); ok && strings.EqualFold(st, "PARDONED") {
			pardoned = 1
		}

		_, _ = s.db.ExecContext(ctx,
			`INSERT INTO punishments (id, server_id, player_id, player_uuid, player_username, type, punishment_type_id, reason, issuer_name, issued_at, expires_at, duration_ms, pardoned, notes_json, evidence_json, data_json, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			pID, serverID, playerID, uuid, playerName, truncate(typeStr, MaxTypeLength), strconv.Itoa(typeOrd),
			truncate(reason, MaxTextLength), truncate(issuerName, MaxUsernameLength),
			issuedAt, expPtr, durPtr, pardoned, string(notesB), string(evB), string(dataB), nowMs,
		)
	}
	return nil
}

func capSlice[T any](in []T, maxLen int) []T {
	if len(in) > maxLen {
		return in[:maxLen]
	}
	return in
}

func nonNilSlice(in []map[string]any) []map[string]any {
	if in == nil {
		return []map[string]any{}
	}
	return in
}

func mergeByKey(existingJSON string, incoming []map[string]any, keyFn func(map[string]any) string, maxLen int) []map[string]any {
	var existing []map[string]any
	if strings.TrimSpace(existingJSON) != "" && existingJSON != "[]" {
		_ = json.Unmarshal([]byte(existingJSON), &existing)
	}
	seen := make(map[string]bool, len(existing)+len(incoming))
	out := make([]map[string]any, 0, len(existing)+len(incoming))
	for _, item := range existing {
		k := keyFn(item)
		if k != "" && !seen[k] {
			seen[k] = true
			out = append(out, item)
		}
	}
	for _, item := range incoming {
		if len(out) >= maxLen {
			break
		}
		k := keyFn(item)
		if k != "" && !seen[k] {
			seen[k] = true
			out = append(out, item)
		}
	}
	return out
}

func extractUsername(usernames []map[string]any) string {
	if len(usernames) == 0 {
		return ""
	}
	last := usernames[len(usernames)-1]
	u, _ := last["username"].(string)
	return truncate(strings.TrimSpace(u), MaxUsernameLength)
}

func asInt64(raw any) (int64, bool) {
	switch v := raw.(type) {
	case json.Number:
		if n, err := v.Int64(); err == nil {
			return n, true
		}
		if f, err := v.Float64(); err == nil {
			return int64(f), true
		}
	case float64:
		return int64(v), true
	case int64:
		return v, true
	case int:
		return int64(v), true
	case string:
		if n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64); err == nil {
			return n, true
		}
	}
	return 0, false
}

func parseTimestampMs(raw any, fallback int64) int64 {
	if n, ok := asInt64(raw); ok && n > 0 {
		if n < 100_000_000_000 {
			return n * 1000
		}
		return n
	}
	if s, ok := raw.(string); ok && strings.TrimSpace(s) != "" {
		if t, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(s)); err == nil {
			return t.UnixMilli()
		}
		if t, err := time.Parse(time.RFC3339, strings.TrimSpace(s)); err == nil {
			return t.UnixMilli()
		}
	}
	return fallback
}

func int32Ptr(v int32) *int32 {
	return &v
}
