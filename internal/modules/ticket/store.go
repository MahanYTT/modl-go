package ticket

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/httpx"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) DB() *sql.DB {
	return s.db
}

func randomSixDigits() int {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return 100000
	}
	return 100000 + int(n.Int64())
}

const ticketColumns = `id, server_id, type, subject, status, creator_uuid, creator_name, creator_avatar,
	reported_player, reported_player_uuid, punishment_id, tags_json, replies_json, notes_json,
	chat_messages_json, form_data_json, data_json, locked, priority, assigned_to_json,
	appeal_workflow_status, ai_analysis_json, email_auth_enabled, hidden, created_at, updated_at`

func scanTicket(scanner interface{ Scan(...any) error }) (*Ticket, error) {
	var (
		t                                                                  Ticket
		typeStr, statusStr, priorityStr                                    string
		tagsJSON, repliesJSON, notesJSON, chatJSON, formJSON, dataJSON     string
		assignedJSON                                                       string
		aiJSON                                                             sql.NullString
		lockedInt, emailAuthInt, hiddenInt                                 int
	)
	err := scanner.Scan(
		&t.ID, &t.ServerID, &typeStr, &t.Subject, &statusStr,
		&t.CreatorUUID, &t.CreatorName, &t.CreatorAvatar,
		&t.ReportedPlayer, &t.ReportedPlayerUUID, &t.PunishmentID,
		&tagsJSON, &repliesJSON, &notesJSON, &chatJSON, &formJSON, &dataJSON,
		&lockedInt, &priorityStr, &assignedJSON,
		&t.AppealWorkflowStatus, &aiJSON,
		&emailAuthInt, &hiddenInt, &t.CreatedAtMs, &t.UpdatedAtMs,
	)
	if err != nil {
		return nil, err
	}
	if cat, err := ParseTicketCategory(typeStr); err == nil {
		t.Type = cat
	} else {
		t.Type = TicketCategory(typeStr)
	}
	if st, err := ParseTicketStatus(statusStr); err == nil {
		t.Status = st
	} else {
		t.Status = TicketStatus(statusStr)
	}
	if pr, err := ParseTicketPriority(priorityStr); err == nil {
		t.Priority = pr
	} else {
		t.Priority = PriorityNormal
	}
	t.Locked = lockedInt != 0
	t.EmailAuthEnabled = emailAuthInt != 0
	t.Hidden = hiddenInt != 0

	_ = json.Unmarshal([]byte(tagsJSON), &t.Tags)
	if t.Tags == nil {
		t.Tags = []string{}
	}
	_ = json.Unmarshal([]byte(repliesJSON), &t.Replies)
	if t.Replies == nil {
		t.Replies = []TicketReply{}
	}
	_ = json.Unmarshal([]byte(notesJSON), &t.Notes)
	if t.Notes == nil {
		t.Notes = []TicketNote{}
	}
	_ = json.Unmarshal([]byte(chatJSON), &t.ChatMessages)
	_ = json.Unmarshal([]byte(formJSON), &t.FormData)
	_ = json.Unmarshal([]byte(dataJSON), &t.Data)
	_ = json.Unmarshal([]byte(assignedJSON), &t.AssignedTo)
	if t.AssignedTo == nil {
		t.AssignedTo = []string{}
	}
	if aiJSON.Valid && strings.TrimSpace(aiJSON.String) != "" && aiJSON.String != "null" {
		var aiRes AIAnalysisResult
		if err := json.Unmarshal([]byte(aiJSON.String), &aiRes); err == nil {
			t.AIAnalysis = &aiRes
		}
	}
	return &t, nil
}

func marshalOrDefault(v any, def string) string {
	if v == nil {
		return def
	}
	b, err := json.Marshal(v)
	if err != nil || string(b) == "null" {
		return def
	}
	return string(b)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (s *Store) InsertWithUniqueID(ctx context.Context, prefix string, t *Ticket) (*Ticket, error) {
	if prefix == "" {
		prefix = "TKT"
	}
	for attempt := 0; attempt < 10; attempt++ {
		t.ID = fmt.Sprintf("%s-%d", prefix, randomSixDigits())
		err := s.insertTicket(ctx, t)
		if err == nil {
			return t, nil
		}
		if strings.Contains(strings.ToLower(err.Error()), "unique") || strings.Contains(strings.ToLower(err.Error()), "constraint") {
			continue
		}
		return nil, err
	}
	return nil, fmt.Errorf("unable to allocate a unique ticket id after 10 attempts")
}

func (s *Store) insertTicket(ctx context.Context, t *Ticket) error {
	var aiJSON any
	if t.AIAnalysis != nil {
		aiJSON = marshalOrDefault(t.AIAnalysis, "")
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO tickets (
			id, server_id, type, subject, status, creator_uuid, creator_name, creator_avatar,
			reported_player, reported_player_uuid, punishment_id, tags_json, replies_json, notes_json,
			chat_messages_json, form_data_json, data_json, locked, priority, assigned_to_json,
			appeal_workflow_status, ai_analysis_json, email_auth_enabled, hidden, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.ServerID, string(t.Type), t.Subject, string(t.Status),
		t.CreatorUUID, t.CreatorName, t.CreatorAvatar,
		t.ReportedPlayer, t.ReportedPlayerUUID, t.PunishmentID,
		marshalOrDefault(t.Tags, "[]"),
		marshalOrDefault(t.Replies, "[]"),
		marshalOrDefault(t.Notes, "[]"),
		marshalOrDefault(t.ChatMessages, "[]"),
		marshalOrDefault(t.FormData, "{}"),
		marshalOrDefault(t.Data, "{}"),
		boolToInt(t.Locked), string(t.Priority),
		marshalOrDefault(t.AssignedTo, "[]"),
		t.AppealWorkflowStatus, aiJSON,
		boolToInt(t.EmailAuthEnabled), boolToInt(t.Hidden),
		t.CreatedAtMs, t.UpdatedAtMs,
	)
	return err
}

func (s *Store) SaveTicket(ctx context.Context, t *Ticket) error {
	var aiJSON any
	if t.AIAnalysis != nil {
		aiJSON = marshalOrDefault(t.AIAnalysis, "")
	}
	res, err := s.db.ExecContext(ctx, `
		UPDATE tickets SET
			type = ?, subject = ?, status = ?, creator_uuid = ?, creator_name = ?, creator_avatar = ?,
			reported_player = ?, reported_player_uuid = ?, punishment_id = ?,
			tags_json = ?, replies_json = ?, notes_json = ?, chat_messages_json = ?,
			form_data_json = ?, data_json = ?, locked = ?, priority = ?, assigned_to_json = ?,
			appeal_workflow_status = ?, ai_analysis_json = ?, email_auth_enabled = ?, hidden = ?,
			updated_at = ?
		WHERE server_id = ? AND id = ?`,
		string(t.Type), t.Subject, string(t.Status),
		t.CreatorUUID, t.CreatorName, t.CreatorAvatar,
		t.ReportedPlayer, t.ReportedPlayerUUID, t.PunishmentID,
		marshalOrDefault(t.Tags, "[]"),
		marshalOrDefault(t.Replies, "[]"),
		marshalOrDefault(t.Notes, "[]"),
		marshalOrDefault(t.ChatMessages, "[]"),
		marshalOrDefault(t.FormData, "{}"),
		marshalOrDefault(t.Data, "{}"),
		boolToInt(t.Locked), string(t.Priority),
		marshalOrDefault(t.AssignedTo, "[]"),
		t.AppealWorkflowStatus, aiJSON,
		boolToInt(t.EmailAuthEnabled), boolToInt(t.Hidden),
		t.UpdatedAtMs,
		t.ServerID, t.ID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) FindByID(ctx context.Context, serverID, ticketID string) (*Ticket, error) {
	row := s.db.QueryRowContext(ctx, `SELECT `+ticketColumns+` FROM tickets WHERE server_id = ? AND id = ? LIMIT 1`, serverID, ticketID)
	t, err := scanTicket(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return t, err
}

func (s *Store) FindByIDs(ctx context.Context, serverID string, ids []string) ([]*Ticket, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]any, 0, len(ids)+1)
	args = append(args, serverID)
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, id)
	}
	q := fmt.Sprintf(`SELECT %s FROM tickets WHERE server_id = ? AND id IN (%s)`, ticketColumns, strings.Join(placeholders, ","))
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Ticket
	for rows.Next() {
		t, err := scanTicket(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) FindByPlayer(ctx context.Context, serverID, playerUUID string) ([]*Ticket, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT `+ticketColumns+` FROM tickets
		WHERE server_id = ?
		  AND (LOWER(creator_uuid) = LOWER(?) OR LOWER(reported_player_uuid) = LOWER(?))
		  AND status != 'unfinished'
		ORDER BY created_at DESC`, serverID, playerUUID, playerUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Ticket
	for rows.Next() {
		t, err := scanTicket(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) FindRecentAssignedWithReplies(ctx context.Context, serverID, assignee string, limit int) ([]*Ticket, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	norm := NormalizeSingleAssignee(assignee)
	if norm == "" {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT `+ticketColumns+` FROM tickets
		WHERE server_id = ?
		  AND status != 'unfinished'
		  AND EXISTS (SELECT 1 FROM json_each(assigned_to_json) WHERE value = ?)
		  AND json_array_length(replies_json) > 0
		ORDER BY updated_at DESC LIMIT ?`, serverID, norm, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Ticket
	for rows.Next() {
		t, err := scanTicket(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) FindByTag(ctx context.Context, serverID, tag string) ([]*Ticket, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT `+ticketColumns+` FROM tickets
		WHERE server_id = ?
		  AND EXISTS (SELECT 1 FROM json_each(tags_json) WHERE value = ?)
		ORDER BY created_at DESC`, serverID, tag)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Ticket
	for rows.Next() {
		t, err := scanTicket(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) FindRecentByCreator(ctx context.Context, serverID, creatorUUID string, limit int) ([]*Ticket, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT `+ticketColumns+` FROM tickets
		WHERE server_id = ? AND LOWER(creator_uuid) = LOWER(?)
		ORDER BY created_at DESC LIMIT ?`, serverID, creatorUUID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Ticket
	for rows.Next() {
		t, err := scanTicket(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) FindMinecraftTickets(ctx context.Context, serverID, status, typ string, limit int) ([]*Ticket, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	clauses := []string{"server_id = ?"}
	args := []any{serverID}

	if strings.TrimSpace(status) != "" && !strings.EqualFold(strings.TrimSpace(status), "all") {
		if st, err := ParseTicketStatus(status); err == nil {
			clauses = append(clauses, "status = ?")
			args = append(args, string(st))
		}
	}
	if strings.TrimSpace(typ) != "" {
		c, a := buildTypeClause(typ)
		clauses = append(clauses, c)
		args = append(args, a...)
	} else {
		clauses = append(clauses, "type IN ('support', 'bug', 'appeal')")
	}

	args = append(args, limit)
	q := fmt.Sprintf(`SELECT %s FROM tickets WHERE %s ORDER BY created_at DESC LIMIT ?`, ticketColumns, strings.Join(clauses, " AND "))
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Ticket
	for rows.Next() {
		t, err := scanTicket(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) FindReports(ctx context.Context, serverID, status, playerUUID string, limit int, sortDesc bool) ([]*Ticket, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	clauses := []string{"server_id = ?", "type IN ('player', 'chat')"}
	args := []any{serverID}

	if strings.TrimSpace(playerUUID) != "" {
		clauses = append(clauses, "LOWER(reported_player_uuid) = LOWER(?)")
		args = append(args, strings.TrimSpace(playerUUID))
	}
	if strings.TrimSpace(status) != "" && !strings.EqualFold(strings.TrimSpace(status), "all") {
		if st, err := ParseTicketStatus(status); err == nil {
			clauses = append(clauses, "status = ?")
			args = append(args, string(st))
		}
	}
	orderBy := ""
	if sortDesc {
		orderBy = " ORDER BY created_at DESC"
	}
	args = append(args, limit)
	q := fmt.Sprintf(`SELECT %s FROM tickets WHERE %s%s LIMIT ?`, ticketColumns, strings.Join(clauses, " AND "), orderBy)
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Ticket
	for rows.Next() {
		t, err := scanTicket(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func buildTypeClause(rawType string) (string, []any) {
	norm := NormalizeAliasKey(rawType)
	cat, err := ParseTicketCategory(rawType)
	if err == nil && !IsCanonicalBucket(norm) {
		return "type = ?", []any{string(cat)}
	}
	if bucketIDs := CategoryIDsForBucket(norm); len(bucketIDs) > 0 {
		ph := make([]string, len(bucketIDs))
		args := make([]any, len(bucketIDs))
		for i, id := range bucketIDs {
			ph[i] = "?"
			args[i] = id
		}
		return fmt.Sprintf("type IN (%s)", strings.Join(ph, ",")), args
	}
	return "LOWER(type) = LOWER(?)", []any{norm}
}

func escapeLike(v string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(strings.ToLower(v))
}

func buildSearchClauses(serverID string, filter TicketSearchFilter, includeReplySearch bool) (string, []any) {
	clauses := []string{"server_id = ?"}
	args := []any{serverID}

	reqStatus, statusErr := ParseTicketStatus(filter.Status)
	if strings.TrimSpace(filter.Status) == "" || strings.EqualFold(strings.TrimSpace(filter.Status), "all") {
		statusErr = errors.New("not specified")
	}

	if statusErr != nil || reqStatus != StatusUnfinished {
		clauses = append(clauses, "status != 'unfinished'")
	}

	if strings.TrimSpace(filter.Search) != "" {
		like := "%" + escapeLike(strings.TrimSpace(filter.Search)) + "%"
		if includeReplySearch {
			clauses = append(clauses, `(
				LOWER(id) LIKE ? ESCAPE '\'
				OR LOWER(subject) LIKE ? ESCAPE '\'
				OR LOWER(creator_name) LIKE ? ESCAPE '\'
				OR EXISTS (
					SELECT 1 FROM json_each(replies_json)
					WHERE LOWER(COALESCE(json_extract(value, '$.name'), '')) LIKE ? ESCAPE '\'
					   OR LOWER(COALESCE(json_extract(value, '$.content'), '')) LIKE ? ESCAPE '\'
				)
			)`)
			args = append(args, like, like, like, like, like)
		} else {
			clauses = append(clauses, `(LOWER(id) LIKE ? ESCAPE '\' OR LOWER(subject) LIKE ? ESCAPE '\' OR LOWER(creator_name) LIKE ? ESCAPE '\')`)
			args = append(args, like, like, like)
		}
	}

	stTrim := strings.TrimSpace(filter.Status)
	if stTrim != "" && !strings.EqualFold(stTrim, "all") {
		if statusErr == nil {
			if reqStatus == StatusUnfinished {
				clauses = append(clauses, "status = 'unfinished'")
			} else if reqStatus.IsTerminal() {
				clauses = append(clauses, "locked = 1")
			} else {
				clauses = append(clauses, "locked = 0")
			}
		} else if strings.EqualFold(stTrim, "open") {
			clauses = append(clauses, "locked = 0")
		} else if strings.EqualFold(stTrim, "closed") {
			clauses = append(clauses, "locked = 1")
		}
	}

	var typeOr []string
	var typeArgs []any
	for _, t := range filter.Types {
		tTrim := strings.TrimSpace(t)
		if tTrim == "" || strings.EqualFold(tTrim, "all") {
			continue
		}
		c, a := buildTypeClause(tTrim)
		typeOr = append(typeOr, c)
		typeArgs = append(typeArgs, a...)
	}
	if len(typeOr) > 0 {
		clauses = append(clauses, "("+strings.Join(typeOr, " OR ")+")")
		args = append(args, typeArgs...)
	}

	if strings.TrimSpace(filter.Author) != "" {
		like := "%" + escapeLike(strings.TrimSpace(filter.Author)) + "%"
		clauses = append(clauses, `LOWER(creator_name) LIKE ? ESCAPE '\'`)
		args = append(args, like)
	}

	for _, lbl := range filter.Labels {
		if strings.TrimSpace(lbl) == "" {
			continue
		}
		clauses = append(clauses, `EXISTS (SELECT 1 FROM json_each(tags_json) WHERE value = ?)`)
		args = append(args, lbl)
	}

	var assOr []string
	var assArgs []any
	for _, a := range filter.Assignees {
		aTrim := strings.TrimSpace(a)
		if aTrim == "" {
			continue
		}
		if strings.EqualFold(aTrim, "none") {
			assOr = append(assOr, "(assigned_to_json IS NULL OR assigned_to_json = '' OR assigned_to_json = '[]')")
			continue
		}
		norm := NormalizeSingleAssignee(aTrim)
		if norm != "" {
			assOr = append(assOr, `EXISTS (SELECT 1 FROM json_each(assigned_to_json) WHERE value = ?)`)
			assArgs = append(assArgs, norm)
		}
	}
	if len(assOr) > 0 {
		clauses = append(clauses, "("+strings.Join(assOr, " OR ")+")")
		args = append(args, assArgs...)
	}

	return strings.Join(clauses, " AND "), args
}

func (s *Store) SearchTickets(ctx context.Context, serverID string, filter TicketSearchFilter, sortOpt string, page, limit int) ([]*Ticket, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	whereSQL, args := buildSearchClauses(serverID, filter, true)

	var total int64
	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM tickets WHERE %s`, whereSQL)
	if err := s.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	orderSQL := "created_at DESC"
	switch sortOpt {
	case "oldest":
		orderSQL = "created_at ASC"
	case "recently-updated":
		orderSQL = "updated_at DESC"
	case "least-recently-updated":
		orderSQL = "updated_at ASC"
	}

	offset := (page - 1) * limit
	pagedArgs := append(append([]any{}, args...), limit, offset)
	pagedQ := fmt.Sprintf(`SELECT %s FROM tickets WHERE %s ORDER BY %s LIMIT ? OFFSET ?`, ticketColumns, whereSQL, orderSQL)
	rows, err := s.db.QueryContext(ctx, pagedQ, pagedArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []*Ticket
	for rows.Next() {
		t, err := scanTicket(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, t)
	}
	return out, total, rows.Err()
}

func (s *Store) CountTickets(ctx context.Context, serverID string, filter TicketSearchFilter) (TicketCounts, error) {
	whereSQL, args := buildSearchClauses(serverID, filter, false)
	var counts TicketCounts
	openQ := fmt.Sprintf(`SELECT COUNT(*) FROM tickets WHERE %s AND locked = 0`, whereSQL)
	if err := s.db.QueryRowContext(ctx, openQ, args...).Scan(&counts.Open); err != nil {
		return counts, err
	}
	closedQ := fmt.Sprintf(`SELECT COUNT(*) FROM tickets WHERE %s AND locked = 1`, whereSQL)
	if err := s.db.QueryRowContext(ctx, closedQ, args...).Scan(&counts.Closed); err != nil {
		return counts, err
	}
	return counts, nil
}

// --- Settings helpers ---

func (s *Store) LoadSettingProto(ctx context.Context, serverID, typ string, msg proto.Message) bool {
	var dataJSON string
	err := s.db.QueryRowContext(ctx,
		`SELECT data_json FROM settings WHERE server_id = ? AND type = ? LIMIT 1`,
		serverID, typ).Scan(&dataJSON)
	if err != nil || strings.TrimSpace(dataJSON) == "" {
		return false
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal([]byte(dataJSON), msg); err != nil {
		return false
	}
	return true
}

// --- Staff & Subscription helpers ---

func (s *Store) FindStaffUsernameByEmail(ctx context.Context, serverID, email string) string {
	if strings.TrimSpace(email) == "" {
		return ""
	}
	var username string
	err := s.db.QueryRowContext(ctx,
		`SELECT username FROM staff WHERE server_id = ? AND LOWER(email) = LOWER(?) LIMIT 1`,
		serverID, strings.TrimSpace(email)).Scan(&username)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(username)
}

func (s *Store) FindStaffMinecraftUUIDsByUsernames(ctx context.Context, serverID string, usernames []string) map[string]string {
	out := make(map[string]string)
	if len(usernames) == 0 {
		return out
	}
	ph := make([]string, len(usernames))
	args := make([]any, 0, len(usernames)+1)
	args = append(args, serverID)
	for i, u := range usernames {
		ph[i] = "?"
		args = append(args, u)
	}
	q := fmt.Sprintf(`SELECT username, assigned_minecraft_uuid FROM staff WHERE server_id = ? AND username IN (%s)`, strings.Join(ph, ","))
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var u, uuid string
		if err := rows.Scan(&u, &uuid); err == nil && strings.TrimSpace(uuid) != "" {
			out[u] = strings.TrimSpace(uuid)
		}
	}
	return out
}

func (s *Store) FindStaffSubscriptions(ctx context.Context, serverID, email string) ([]StaffTicketSubscription, error) {
	var raw string
	err := s.db.QueryRowContext(ctx,
		`SELECT subscribed_tickets_json FROM staff WHERE server_id = ? AND LOWER(email) = LOWER(?) LIMIT 1`,
		serverID, strings.TrimSpace(email)).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var subs []StaffTicketSubscription
	if strings.TrimSpace(raw) != "" {
		_ = json.Unmarshal([]byte(raw), &subs)
	}
	if subs == nil {
		subs = []StaffTicketSubscription{}
	}
	return subs, nil
}

func (s *Store) SaveStaffSubscriptions(ctx context.Context, serverID, email string, subs []StaffTicketSubscription) error {
	b, err := json.Marshal(subs)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx,
		`UPDATE staff SET subscribed_tickets_json = ?, updated_at = ? WHERE server_id = ? AND LOWER(email) = LOWER(?)`,
		string(b), httpx.NowMs(), serverID, strings.TrimSpace(email))
	return err
}

type StaffSubInfo struct {
	Email         string
	Username      string
	Subscriptions []StaffTicketSubscription
	NotifyEnabled bool
	TicketReplies bool
}

func (s *Store) ListAllStaffWithSubscriptions(ctx context.Context, serverID string) ([]StaffSubInfo, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT email, username, subscribed_tickets_json, ticket_subscription_settings_json FROM staff WHERE server_id = ?`,
		serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []StaffSubInfo
	for rows.Next() {
		var email, username, subsJSON, settingsJSON string
		if err := rows.Scan(&email, &username, &subsJSON, &settingsJSON); err != nil {
			return nil, err
		}
		info := StaffSubInfo{
			Email:         email,
			Username:      username,
			NotifyEnabled: true,
			TicketReplies: true,
		}
		_ = json.Unmarshal([]byte(subsJSON), &info.Subscriptions)
		if strings.TrimSpace(settingsJSON) != "" && settingsJSON != "{}" {
			var cfg struct {
				Enabled            *bool `json:"enabled"`
				EmailNotifications *struct {
					Enabled       *bool `json:"enabled"`
					TicketReplies *bool `json:"ticketReplies"`
				} `json:"emailNotifications"`
			}
			if err := json.Unmarshal([]byte(settingsJSON), &cfg); err == nil {
				if cfg.Enabled != nil {
					info.NotifyEnabled = *cfg.Enabled
				}
				if cfg.EmailNotifications != nil {
					if cfg.EmailNotifications.TicketReplies != nil {
						info.TicketReplies = *cfg.EmailNotifications.TicketReplies
					}
				}
			}
		}
		out = append(out, info)
	}
	return out, rows.Err()
}

func (s *Store) InsertSubscriptionUpdate(ctx context.Context, u *SubscriptionUpdate) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO ticket_subscription_updates (id, server_id, staff_email, ticket_id, event_type, summary, actor, is_read, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		u.ID, u.ServerID, strings.ToLower(strings.TrimSpace(u.StaffEmail)), u.TicketID, u.EventType, u.Summary, u.Actor, boolToInt(u.Read), u.CreatedAt)
	return err
}

func (s *Store) FindSubscriptionUpdates(ctx context.Context, serverID, email string, limit int) ([]SubscriptionUpdate, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	normEmail := strings.ToLower(strings.TrimSpace(email))
	var unread int64
	_ = s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM ticket_subscription_updates WHERE server_id = ? AND LOWER(staff_email) = ? AND is_read = 0`,
		serverID, normEmail).Scan(&unread)

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, server_id, staff_email, ticket_id, event_type, summary, actor, is_read, created_at
		FROM ticket_subscription_updates
		WHERE server_id = ? AND LOWER(staff_email) = ?
		ORDER BY created_at DESC LIMIT ?`, serverID, normEmail, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []SubscriptionUpdate
	for rows.Next() {
		var u SubscriptionUpdate
		var readInt int
		if err := rows.Scan(&u.ID, &u.ServerID, &u.StaffEmail, &u.TicketID, &u.EventType, &u.Summary, &u.Actor, &readInt, &u.CreatedAt); err != nil {
			return nil, 0, err
		}
		u.Read = readInt != 0
		out = append(out, u)
	}
	return out, unread, rows.Err()
}

func (s *Store) MarkSubscriptionUpdateRead(ctx context.Context, serverID, email, updateID string) (bool, error) {
	res, err := s.db.ExecContext(ctx, `
		UPDATE ticket_subscription_updates SET is_read = 1
		WHERE server_id = ? AND LOWER(staff_email) = LOWER(?) AND id = ?`,
		serverID, strings.TrimSpace(email), updateID)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (s *Store) MarkAllSubscriptionUpdatesRead(ctx context.Context, serverID, email string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE ticket_subscription_updates SET is_read = 1
		WHERE server_id = ? AND LOWER(staff_email) = LOWER(?) AND is_read = 0`,
		serverID, strings.TrimSpace(email))
	return err
}

// --- Ticket verifications ---

func (s *Store) UpsertVerification(ctx context.Context, v *TicketVerification) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM ticket_verifications WHERE server_id = ? AND ticket_id = ?`, v.ServerID, v.TicketID)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO ticket_verifications (id, server_id, ticket_id, token, code_hash, email, failed_attempts, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		v.ID, v.ServerID, v.TicketID, v.Token, v.CodeHash, v.Email, v.FailedAttempts, v.ExpiresAt)
	return err
}

func (s *Store) FindVerificationBySubject(ctx context.Context, serverID, subject string) (*TicketVerification, error) {
	var v TicketVerification
	err := s.db.QueryRowContext(ctx, `
		SELECT id, server_id, ticket_id, token, code_hash, email, failed_attempts, expires_at
		FROM ticket_verifications WHERE server_id = ? AND ticket_id = ? LIMIT 1`,
		serverID, subject).Scan(&v.ID, &v.ServerID, &v.TicketID, &v.Token, &v.CodeHash, &v.Email, &v.FailedAttempts, &v.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &v, err
}

func (s *Store) IncrementVerificationFailure(ctx context.Context, id string) (int, error) {
	_, err := s.db.ExecContext(ctx, `UPDATE ticket_verifications SET failed_attempts = failed_attempts + 1 WHERE id = ?`, id)
	if err != nil {
		return 0, err
	}
	var count int
	err = s.db.QueryRowContext(ctx, `SELECT failed_attempts FROM ticket_verifications WHERE id = ?`, id).Scan(&count)
	return count, err
}

func (s *Store) DeleteVerificationByID(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM ticket_verifications WHERE id = ?`, id)
	return err
}

func (s *Store) DeleteVerificationBySubject(ctx context.Context, serverID, subject string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM ticket_verifications WHERE server_id = ? AND ticket_id = ?`, serverID, subject)
	return err
}

func (s *Store) UpdateVerificationToken(ctx context.Context, id, token string, expiresAt int64) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE ticket_verifications SET token = ?, code_hash = '', failed_attempts = 0, expires_at = ? WHERE id = ?`,
		token, expiresAt, id)
	return err
}

// --- Player & Audit helpers ---

func (s *Store) PlayerExists(ctx context.Context, serverID, playerUUID string) bool {
	var id string
	err := s.db.QueryRowContext(ctx,
		`SELECT id FROM players WHERE server_id = ? AND LOWER(minecraft_uuid) = LOWER(?) LIMIT 1`,
		serverID, strings.TrimSpace(playerUUID)).Scan(&id)
	return err == nil
}

func (s *Store) InsertPlayerNotification(ctx context.Context, id, serverID, playerUUID, notifType, title, message, dataJSON string, createdAt int64) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO player_notifications (id, server_id, player_uuid, type, title, message, data_json, delivered, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, 0, ?)`,
		id, serverID, strings.ToLower(strings.TrimSpace(playerUUID)), notifType, title, message, dataJSON, createdAt)
	return err
}

func (s *Store) RecordAuditLog(ctx context.Context, serverID, level, source, description string) {
	if strings.TrimSpace(source) == "" {
		source = "system"
	}
	_, _ = s.db.ExecContext(ctx, `
		INSERT INTO audit_logs (id, server_id, created_at, level, source, description, metadata_json)
		VALUES (?, ?, ?, ?, ?, ?, '{}')`,
		httpx.NewID(), serverID, httpx.NowMs(), level, source, description)
}

func (s *Store) ResolvePanelActor(ctx context.Context, serverID, email string) string {
	if strings.TrimSpace(email) == "" {
		return "Unknown"
	}
	if u := s.FindStaffUsernameByEmail(ctx, serverID, email); u != "" {
		return u
	}
	return email
}

func (s *Store) GetQuickResponses(ctx context.Context, serverID string) *modlv1.QuickResponseSettings {
	qr := &modlv1.QuickResponseSettings{Categories: []*modlv1.QuickResponseCategory{}}
	_ = s.LoadSettingProto(ctx, serverID, "quickResponses", qr)
	return qr
}

func DefaultTicketForms() *modlv1.TicketFormSettings {
	empty := func() *modlv1.TicketForm {
		return &modlv1.TicketForm{
			AllowEmailNotifications: true,
			Fields:                  []*modlv1.FormField{},
			Sections:                []*modlv1.FormSection{},
		}
	}
	return &modlv1.TicketFormSettings{
		Bug:         empty(),
		Support:     empty(),
		Application: empty(),
		Player:      empty(),
		Chat:        empty(),
	}
}

func (s *Store) GetTicketForms(ctx context.Context, serverID string) *modlv1.TicketFormSettings {
	tf := DefaultTicketForms()
	_ = s.LoadSettingProto(ctx, serverID, "ticketForms", tf)
	if tf.Bug == nil {
		tf.Bug = DefaultTicketForms().Bug
	}
	if tf.Support == nil {
		tf.Support = DefaultTicketForms().Support
	}
	if tf.Application == nil {
		tf.Application = DefaultTicketForms().Application
	}
	if tf.Player == nil {
		tf.Player = DefaultTicketForms().Player
	}
	if tf.Chat == nil {
		tf.Chat = DefaultTicketForms().Chat
	}
	return tf
}

func (s *Store) GetFormByType(ctx context.Context, serverID, formType string) *modlv1.TicketForm {
	tf := s.GetTicketForms(ctx, serverID)
	switch strings.ToLower(strings.TrimSpace(formType)) {
	case "bug":
		return tf.Bug
	case "support":
		return tf.Support
	case "application", "staff":
		return tf.Application
	case "player":
		return tf.Player
	case "chat":
		return tf.Chat
	default:
		return nil
	}
}
