package appeal

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"math/big"
	"strings"

	"modl-go/internal/httpx"
	"modl-go/internal/modules/ticket"
)

type Store struct {
	db          *sql.DB
	ticketStore *ticket.Store
}

func NewStore(db *sql.DB, ticketStore *ticket.Store) *Store {
	return &Store{db: db, ticketStore: ticketStore}
}

func shortID() string {
	const alpha = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 8)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alpha))))
		if err != nil {
			b[i] = 'A'
			continue
		}
		b[i] = alpha[n.Int64()]
	}
	return string(b)
}

type PlayerPunishmentInfo struct {
	PlayerUUID     string
	PlayerUsername string
	PunishmentID   string
	Pardoned       bool
	BlockAlts      bool
}

func (s *Store) FindAppealsByPunishmentID(ctx context.Context, serverID, punishmentID string) ([]*ticket.Ticket, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id FROM tickets
		WHERE server_id = ?
		  AND type = 'appeal'
		  AND (punishment_id = ? OR json_extract(data_json, '$.punishmentId') = ?)
		ORDER BY created_at DESC`, serverID, punishmentID, punishmentID)
	if err != nil {
		return nil, err
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}
	return s.ticketStore.FindByIDs(ctx, serverID, ids)
}

func (s *Store) FindPlayerWithPunishment(ctx context.Context, serverID, playerUUID, punishmentID string) (*PlayerPunishmentInfo, error) {
	normUUID := ticket.NormalizeUUID(playerUUID)
	var playerUsername string
	var usernamesJSON string
	err := s.db.QueryRowContext(ctx, `
		SELECT current_username, usernames_json FROM players
		WHERE server_id = ? AND LOWER(minecraft_uuid) = LOWER(?) LIMIT 1`,
		serverID, normUUID).Scan(&playerUsername, &usernamesJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ticket.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	var pID, pUsername string
	var pardonedInt, blockAltsInt int
	var dataJSON string
	err = s.db.QueryRowContext(ctx, `
		SELECT id, player_username, pardoned, block_alts, data_json FROM punishments
		WHERE server_id = ? AND id = ? AND LOWER(player_uuid) = LOWER(?) LIMIT 1`,
		serverID, punishmentID, normUUID).Scan(&pID, &pUsername, &pardonedInt, &blockAltsInt, &dataJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ticket.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	username := strings.TrimSpace(playerUsername)
	if username == "" && strings.TrimSpace(usernamesJSON) != "" {
		var entries []struct {
			Username string `json:"username"`
		}
		if json.Unmarshal([]byte(usernamesJSON), &entries) == nil && len(entries) > 0 {
			username = entries[len(entries)-1].Username
		}
		if username == "" {
			var strList []string
			if json.Unmarshal([]byte(usernamesJSON), &strList) == nil && len(strList) > 0 {
				username = strList[len(strList)-1]
			}
		}
	}
	if username == "" {
		username = strings.TrimSpace(pUsername)
	}
	if username == "" {
		username = "Player"
	}

	blockAlts := blockAltsInt != 0
	if !blockAlts && strings.TrimSpace(dataJSON) != "" {
		var d map[string]any
		if json.Unmarshal([]byte(dataJSON), &d) == nil {
			if ab, ok := d["altBlocking"].(bool); ok && ab {
				blockAlts = true
			}
		}
	}

	return &PlayerPunishmentInfo{
		PlayerUUID:     normUUID,
		PlayerUsername: username,
		PunishmentID:   pID,
		Pardoned:       pardonedInt != 0,
		BlockAlts:      blockAlts,
	}, nil
}

func (s *Store) LinkAppealToPunishment(ctx context.Context, serverID, playerUUID, punishmentID, appealID string) error {
	var notesJSON, dataJSON string
	err := s.db.QueryRowContext(ctx, `
		SELECT notes_json, data_json FROM punishments
		WHERE server_id = ? AND id = ? AND LOWER(player_uuid) = LOWER(?) LIMIT 1`,
		serverID, punishmentID, ticket.NormalizeUUID(playerUUID)).Scan(&notesJSON, &dataJSON)
	if err != nil {
		return err
	}

	now := httpx.NowMs()
	var notes []map[string]any
	_ = json.Unmarshal([]byte(notesJSON), &notes)
	notes = append(notes, map[string]any{
		"id":         shortID(),
		"text":       "opened appeal (#" + appealID + ")",
		"date":       now,
		"issuerName": "System",
	})

	data := make(map[string]any)
	_ = json.Unmarshal([]byte(dataJSON), &data)
	var attached []any
	if existing, ok := data["attachedTicketIds"].([]any); ok {
		attached = existing
	}
	found := false
	for _, item := range attached {
		if str, ok := item.(string); ok && str == appealID {
			found = true
			break
		}
	}
	if !found {
		attached = append(attached, appealID)
	}
	data["attachedTicketIds"] = attached

	nb, _ := json.Marshal(notes)
	dbBytes, _ := json.Marshal(data)
	_, err = s.db.ExecContext(ctx, `
		UPDATE punishments SET notes_json = ?, data_json = ?, updated_at = ?
		WHERE server_id = ? AND id = ? AND LOWER(player_uuid) = LOWER(?)`,
		string(nb), string(dbBytes), now, serverID, punishmentID, ticket.NormalizeUUID(playerUUID))
	return err
}

func (s *Store) RecordAppealRejection(ctx context.Context, serverID, playerUUID, punishmentID, staffName, appealID string) error {
	var notesJSON, dataJSON string
	err := s.db.QueryRowContext(ctx, `
		SELECT notes_json, data_json FROM punishments
		WHERE server_id = ? AND id = ? AND LOWER(player_uuid) = LOWER(?) LIMIT 1`,
		serverID, punishmentID, ticket.NormalizeUUID(playerUUID)).Scan(&notesJSON, &dataJSON)
	if err != nil {
		return err
	}

	now := httpx.NowMs()
	if strings.TrimSpace(staffName) == "" {
		staffName = "System"
	}
	var notes []map[string]any
	_ = json.Unmarshal([]byte(notesJSON), &notes)
	notes = append(notes, map[string]any{
		"id":         shortID(),
		"text":       "rejected appeal (#" + appealID + ")",
		"date":       now,
		"issuerName": staffName,
	})

	data := make(map[string]any)
	_ = json.Unmarshal([]byte(dataJSON), &data)
	data["appealOutcome"] = "Rejected"
	data["appealTicketId"] = appealID

	nb, _ := json.Marshal(notes)
	dbBytes, _ := json.Marshal(data)
	_, err = s.db.ExecContext(ctx, `
		UPDATE punishments SET notes_json = ?, data_json = ?, updated_at = ?
		WHERE server_id = ? AND id = ? AND LOWER(player_uuid) = LOWER(?)`,
		string(nb), string(dbBytes), now, serverID, punishmentID, ticket.NormalizeUUID(playerUUID))
	return err
}

func (s *Store) ApplyAppealApproval(ctx context.Context, serverID, playerUUID, punishmentID, staffName, appealID string) error {
	var notesJSON, dataJSON string
	var blockAltsInt int
	err := s.db.QueryRowContext(ctx, `
		SELECT notes_json, data_json, block_alts FROM punishments
		WHERE server_id = ? AND id = ? AND LOWER(player_uuid) = LOWER(?) LIMIT 1`,
		serverID, punishmentID, ticket.NormalizeUUID(playerUUID)).Scan(&notesJSON, &dataJSON, &blockAltsInt)
	if err != nil {
		return err
	}

	now := httpx.NowMs()
	if strings.TrimSpace(staffName) == "" {
		staffName = "System"
	}
	var notes []map[string]any
	_ = json.Unmarshal([]byte(notesJSON), &notes)
	notes = append(notes, map[string]any{
		"id":         shortID(),
		"text":       "accepted appeal (#" + appealID + ")",
		"date":       now,
		"issuerName": staffName,
	})

	data := make(map[string]any)
	_ = json.Unmarshal([]byte(dataJSON), &data)
	var mods []any
	if existing, ok := data["modifications"].([]any); ok {
		mods = existing
	}
	mods = append(mods, map[string]any{
		"id":             shortID(),
		"type":           "APPEAL_ACCEPT",
		"date":           now,
		"issuerName":     staffName,
		"reason":         "Appeal approved",
		"appealTicketId": appealID,
	})
	data["modifications"] = mods
	data["appealOutcome"] = "Approved"
	data["appealTicketId"] = appealID
	data["status"] = "Pardoned"

	altBlocking := blockAltsInt != 0
	if !altBlocking {
		if ab, ok := data["altBlocking"].(bool); ok && ab {
			altBlocking = true
		}
	}

	nb, _ := json.Marshal(notes)
	dbBytes, _ := json.Marshal(data)
	_, err = s.db.ExecContext(ctx, `
		UPDATE punishments SET
			pardoned = 1, pardoned_by = ?, pardoned_at = ?, pardon_reason = ?,
			notes_json = ?, data_json = ?, updated_at = ?
		WHERE server_id = ? AND id = ? AND LOWER(player_uuid) = LOWER(?)`,
		staffName, now, "Appeal approved", string(nb), string(dbBytes), now,
		serverID, punishmentID, ticket.NormalizeUUID(playerUUID))
	if err != nil {
		return err
	}

	if altBlocking {
		_, _ = s.db.ExecContext(ctx, `
			UPDATE punishments SET
				pardoned = 1, pardoned_by = ?, pardoned_at = ?, pardon_reason = ?, updated_at = ?
			WHERE server_id = ? AND pardoned = 0 AND json_extract(data_json, '$.linkedBanId') = ?`,
			staffName, now, "Parent punishment appeal approved", now, serverID, punishmentID)
	}
	return nil
}
