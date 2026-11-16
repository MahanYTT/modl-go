package dashboard

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/models"
)

const (
	dayMs                      = 24 * 60 * 60 * 1000
	recentPunishmentWindowDays = 7
	maxRecentTicketsLimit      = 20
	maxRecentPunishmentsLimit  = 20
	maxActivityLimit           = 100
	maxActivityDays            = 90
	maxQueryResults            = 200
)

type MinecraftStats struct {
	UnresolvedReports      int64 `json:"unresolvedReports"`
	UnresolvedTickets      int64 `json:"unresolvedTickets"`
	OnlineStaff            int64 `json:"onlineStaff"`
	OnlinePlayers          int64 `json:"onlinePlayers"`
	ActiveBans             int64 `json:"activeBans"`
	ActiveMutes            int64 `json:"activeMutes"`
	TotalActivePunishments int64 `json:"totalActivePunishments"`
	TotalPlayers           int64 `json:"totalPlayers"`
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func resolveRangeDays(period string) int {
	switch strings.TrimSpace(period) {
	case "7d":
		return 7
	case "90d":
		return 90
	case "365d", "1y":
		return 365
	default:
		return 30
	}
}

func clampLimit(value, maxVal int) int {
	if value < 1 {
		return 1
	}
	if value > maxVal {
		return maxVal
	}
	return value
}

func (s *Service) GetMinecraftStats(ctx context.Context, srv *models.Server) (*MinecraftStats, error) {
	var unresolvedReports, unresolvedTickets int64
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM tickets WHERE server_id = ? AND LOWER(type) IN ('player', 'chat') AND LOWER(status) IN ('open', 'unfinished')`,
		srv.ID,
	).Scan(&unresolvedReports); err != nil {
		return nil, err
	}
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM tickets WHERE server_id = ? AND LOWER(type) IN ('support', 'bug', 'appeal') AND LOWER(status) IN ('open', 'unfinished')`,
		srv.ID,
	).Scan(&unresolvedTickets); err != nil {
		return nil, err
	}

	onlineUUIDs, totalPlayers, err := s.loadOnlinePlayers(ctx, srv.ID)
	if err != nil {
		return nil, err
	}

	var onlineStaff int64
	staffRows, err := s.db.QueryContext(ctx,
		`SELECT assigned_minecraft_uuid FROM staff WHERE server_id = ? AND assigned_minecraft_uuid != ''`,
		srv.ID,
	)
	if err != nil {
		return nil, err
	}
	for staffRows.Next() {
		var uuid string
		if err := staffRows.Scan(&uuid); err != nil {
			staffRows.Close()
			return nil, err
		}
		if onlineUUIDs[strings.ToLower(strings.TrimSpace(uuid))] {
			onlineStaff++
		}
	}
	if err := staffRows.Err(); err != nil {
		staffRows.Close()
		return nil, err
	}
	staffRows.Close()

	bans, mutes, totalActive, err := s.countActivePunishments(ctx, srv.ID)
	if err != nil {
		return nil, err
	}

	return &MinecraftStats{
		UnresolvedReports:      unresolvedReports,
		UnresolvedTickets:      unresolvedTickets,
		OnlineStaff:            onlineStaff,
		OnlinePlayers:          int64(len(onlineUUIDs)),
		ActiveBans:             bans,
		ActiveMutes:            mutes,
		TotalActivePunishments: totalActive,
		TotalPlayers:           totalPlayers,
	}, nil
}

func (s *Service) loadOnlinePlayers(ctx context.Context, serverID string) (map[string]bool, int64, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT minecraft_uuid, data_json FROM players WHERE server_id = ?`,
		serverID,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	online := make(map[string]bool)
	var total int64
	for rows.Next() {
		var uuid, dataJSON string
		if err := rows.Scan(&uuid, &dataJSON); err != nil {
			return nil, 0, err
		}
		total++
		if isPlayerOnline(dataJSON) && uuid != "" {
			online[strings.ToLower(strings.TrimSpace(uuid))] = true
		}
	}
	return online, total, rows.Err()
}

func (s *Service) countActivePunishments(ctx context.Context, serverID string) (bans, mutes, total int64, err error) {
	byOrdinal, byName := loadPunishmentTypes(ctx, s.db, serverID)
	nowMs := time.Now().UnixMilli()

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, player_uuid, player_username, type, punishment_type_id, reason, issuer_name,
		        issued_at, expires_at, duration_ms, pardoned, data_json
		 FROM punishments
		 WHERE server_id = ?`,
		serverID,
	)
	if err != nil {
		return 0, 0, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, playerUUID, playerUsername, typeRaw, typeID, reason, issuerName, dataJSON string
		var issuedAt int64
		var expiresAt, durationMs sql.NullInt64
		var pardoned int
		if err := rows.Scan(&id, &playerUUID, &playerUsername, &typeRaw, &typeID, &reason, &issuerName,
			&issuedAt, &expiresAt, &durationMs, &pardoned, &dataJSON); err != nil {
			return 0, 0, 0, err
		}
		p := parsePunishment(id, playerUUID, playerUsername, typeRaw, typeID, reason, issuerName,
			issuedAt, expiresAt, durationMs, pardoned, dataJSON, byName)
		if !p.isActive(nowMs) {
			continue
		}
		total++
		switch p.effectiveCategory(byOrdinal) {
		case "BAN":
			bans++
		case "MUTE":
			mutes++
		}
	}
	return bans, mutes, total, rows.Err()
}

func (s *Service) GetMetrics(ctx context.Context, srv *models.Server, period string) (*modlv1.DashboardMetricsResponse, error) {
	windowDays := int64(resolveRangeDays(period))
	nowMs := time.Now().UnixMilli()
	windowStart := nowMs - windowDays*dayMs
	priorWindowStart := nowMs - windowDays*2*dayMs

	var totalTickets, openTickets, totalPunishments, staffCount int64
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tickets WHERE server_id = ?`, srv.ID).Scan(&totalTickets); err != nil {
		return nil, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tickets WHERE server_id = ? AND LOWER(status) = 'open'`, srv.ID).Scan(&openTickets); err != nil {
		return nil, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM punishments WHERE server_id = ?`, srv.ID).Scan(&totalPunishments); err != nil {
		return nil, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM staff WHERE server_id = ?`, srv.ID).Scan(&staffCount); err != nil {
		return nil, err
	}
	if strings.TrimSpace(srv.AdminEmail) != "" {
		var exists int
		if err := s.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM staff WHERE server_id = ? AND LOWER(email) = LOWER(?)`,
			srv.ID, strings.TrimSpace(srv.AdminEmail),
		).Scan(&exists); err != nil {
			return nil, err
		}
		if exists == 0 {
			staffCount++
		}
	}

	_, _, activePunishments, err := s.countActivePunishments(ctx, srv.ID)
	if err != nil {
		return nil, err
	}

	var recentTickets, prevTickets int64
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM tickets WHERE server_id = ? AND created_at >= ?`, srv.ID, windowStart,
	).Scan(&recentTickets); err != nil {
		return nil, err
	}
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM tickets WHERE server_id = ? AND created_at >= ? AND created_at < ?`,
		srv.ID, priorWindowStart, windowStart,
	).Scan(&prevTickets); err != nil {
		return nil, err
	}

	var totalPlayers, recentPlayers, prevPlayers int64
	playerRows, err := s.db.QueryContext(ctx,
		`SELECT first_seen, data_json FROM players WHERE server_id = ?`, srv.ID,
	)
	if err != nil {
		return nil, err
	}
	for playerRows.Next() {
		var firstSeen int64
		var dataJSON string
		if err := playerRows.Scan(&firstSeen, &dataJSON); err != nil {
			playerRows.Close()
			return nil, err
		}
		totalPlayers++
		joined := extractPlayerFirstJoinMs(firstSeen, dataJSON)
		if joined >= windowStart {
			recentPlayers++
		} else if joined >= priorWindowStart && joined < windowStart {
			prevPlayers++
		}
	}
	if err := playerRows.Err(); err != nil {
		playerRows.Close()
		return nil, err
	}
	playerRows.Close()

	var ticketsTrend int32
	if prevTickets > 0 {
		ticketsTrend = int32(math.Round((float64(recentTickets-prevTickets) / float64(prevTickets)) * 100))
	}
	var playersTrend int32
	if prevPlayers > 0 {
		playersTrend = int32(math.Round((float64(recentPlayers-prevPlayers) / float64(prevPlayers)) * 100))
	}

	return &modlv1.DashboardMetricsResponse{
		TotalTickets:      totalTickets,
		OpenTickets:       openTickets,
		TotalPlayers:      totalPlayers,
		TotalPunishments:  totalPunishments,
		ActivePunishments: activePunishments,
		TotalStaff:        staffCount,
		TicketsTrend:      ticketsTrend,
		PlayersTrend:      playersTrend,
	}, nil
}

type ticketReplyJSON struct {
	Name    string `json:"name"`
	Content string `json:"content"`
	Created int64  `json:"created"`
}

func (s *Service) GetRecentTickets(ctx context.Context, srv *models.Server, limit int) ([]*modlv1.RecentTicketResponse, error) {
	safeLimit := clampLimit(limit, maxRecentTicketsLimit)
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, subject, status, priority, created_at, creator_name, type, replies_json
		 FROM tickets
		 WHERE server_id = ? AND LOWER(status) != 'unfinished'
		 ORDER BY created_at DESC
		 LIMIT ?`,
		srv.ID, safeLimit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*modlv1.RecentTicketResponse, 0)
	for rows.Next() {
		var id, subject, status, priority, creatorName, ticketType, repliesJSON string
		var createdAt int64
		if err := rows.Scan(&id, &subject, &status, &priority, &createdAt, &creatorName, &ticketType, &repliesJSON); err != nil {
			return nil, err
		}
		if strings.TrimSpace(subject) == "" {
			subject = "No Subject"
		}
		if strings.TrimSpace(status) == "" {
			status = "open"
		}
		if strings.TrimSpace(priority) == "" {
			priority = "normal"
		}
		if strings.TrimSpace(creatorName) == "" {
			creatorName = "Unknown"
		}
		if strings.TrimSpace(ticketType) == "" {
			ticketType = "support"
		}
		var initialMsg string
		var replies []ticketReplyJSON
		if err := json.Unmarshal([]byte(repliesJSON), &replies); err == nil && len(replies) > 0 {
			initialMsg = replies[0].Content
		}
		out = append(out, &modlv1.RecentTicketResponse{
			Id:             id,
			Title:          subject,
			InitialMessage: initialMsg,
			Status:         status,
			Priority:       priority,
			CreatedAt:      createdAt,
			PlayerName:     creatorName,
			Type:           ticketType,
		})
	}
	return out, rows.Err()
}

func (s *Service) GetRecentPunishments(ctx context.Context, srv *models.Server, limit int) ([]*modlv1.RecentPunishmentResponse, error) {
	safeLimit := clampLimit(limit, maxRecentPunishmentsLimit)
	nowMs := time.Now().UnixMilli()
	cutoff := nowMs - int64(recentPunishmentWindowDays)*dayMs
	byOrdinal, byName := loadPunishmentTypes(ctx, s.db, srv.ID)

	rows, err := s.db.QueryContext(ctx,
		`SELECT p.id, p.player_uuid, p.player_username, COALESCE(pl.current_username, ''), COALESCE(pl.usernames_json, '[]'),
		        p.type, p.punishment_type_id, p.reason, p.issuer_name, p.issued_at, p.expires_at, p.duration_ms, p.pardoned, p.data_json
		 FROM punishments p
		 LEFT JOIN players pl ON pl.server_id = p.server_id AND LOWER(pl.minecraft_uuid) = LOWER(p.player_uuid)
		 WHERE p.server_id = ? AND p.issued_at >= ?
		 ORDER BY p.issued_at DESC
		 LIMIT ?`,
		srv.ID, cutoff, safeLimit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*modlv1.RecentPunishmentResponse, 0)
	for rows.Next() {
		var id, playerUUID, pUsername, curUsername, usernamesJSON, typeRaw, typeID, reason, issuerName, dataJSON string
		var issuedAt int64
		var expiresAt, durationMs sql.NullInt64
		var pardoned int
		if err := rows.Scan(&id, &playerUUID, &pUsername, &curUsername, &usernamesJSON,
			&typeRaw, &typeID, &reason, &issuerName, &issuedAt, &expiresAt, &durationMs, &pardoned, &dataJSON); err != nil {
			return nil, err
		}
		parsed := parsePunishment(id, playerUUID, pUsername, typeRaw, typeID, reason, issuerName,
			issuedAt, expiresAt, durationMs, pardoned, dataJSON, byName)
		playerName := pUsername
		if strings.TrimSpace(playerName) == "" {
			playerName = extractLatestUsername(curUsername, usernamesJSON)
		}
		if issuerName == "" {
			issuerName = "Unknown"
		}
		out = append(out, &modlv1.RecentPunishmentResponse{
			Id:         parsed.ID,
			PlayerName: playerName,
			PlayerUuid: parsed.PlayerUUID,
			Type:       parsed.resolveTypeName(byOrdinal),
			Reason:     parsed.Reason,
			IssuerName: issuerName,
			Issued:     parsed.IssuedAt,
			Active:     parsed.isActive(nowMs),
		})
	}
	return out, rows.Err()
}

func displayTicketCategory(ticketType string) string {
	switch strings.ToLower(strings.TrimSpace(ticketType)) {
	case "bug":
		return "Bug Report"
	case "player":
		return "Player Report"
	case "chat":
		return "Chat Report"
	case "appeal":
		return "Ban Appeal"
	case "application":
		return "Staff Application"
	default:
		return "General Support"
	}
}

func (s *Service) GetRecentActivity(ctx context.Context, srv *models.Server, staffEmail string, limit, days int) ([]*modlv1.ActivityItemResponse, error) {
	safeLimit := clampLimit(limit, maxActivityLimit)
	safeDays := clampLimit(days, maxActivityDays)

	var staffUsername string
	err := s.db.QueryRowContext(ctx,
		`SELECT username FROM staff WHERE server_id = ? AND LOWER(email) = LOWER(?) LIMIT 1`,
		srv.ID, strings.TrimSpace(staffEmail),
	).Scan(&staffUsername)
	if err == sql.ErrNoRows || strings.TrimSpace(staffUsername) == "" {
		return []*modlv1.ActivityItemResponse{}, nil
	}
	if err != nil {
		return nil, err
	}

	cutoff := time.Now().UnixMilli() - int64(safeDays)*dayMs
	activities := make([]*modlv1.ActivityItemResponse, 0)

	// 1. Ticket activities
	tRows, err := s.db.QueryContext(ctx,
		`SELECT id, subject, type, creator_name, created_at, updated_at, assigned_to_json, replies_json
		 FROM tickets
		 WHERE server_id = ? AND updated_at >= ?
		 ORDER BY updated_at DESC
		 LIMIT ?`,
		srv.ID, cutoff, maxQueryResults,
	)
	if err != nil {
		return nil, err
	}
	for tRows.Next() {
		var id, subject, ticketType, creatorName, assignedJSON, repliesJSON string
		var createdAt, updatedAt int64
		if err := tRows.Scan(&id, &subject, &ticketType, &creatorName, &createdAt, &updatedAt, &assignedJSON, &repliesJSON); err != nil {
			tRows.Close()
			return nil, err
		}
		var assigned []string
		_ = json.Unmarshal([]byte(assignedJSON), &assigned)
		isAssigned := false
		for _, a := range assigned {
			if strings.EqualFold(strings.TrimSpace(a), staffUsername) {
				isAssigned = true
				break
			}
		}
		var replies []ticketReplyJSON
		_ = json.Unmarshal([]byte(repliesJSON), &replies)
		hasStaffReply := false
		for _, rep := range replies {
			if strings.EqualFold(strings.TrimSpace(rep.Name), staffUsername) {
				hasStaffReply = true
				break
			}
		}
		if !strings.EqualFold(creatorName, staffUsername) && !isAssigned && !hasStaffReply {
			continue
		}

		subj := subject
		if strings.TrimSpace(subj) == "" {
			subj = "No Subject"
		}
		catDisplay := displayTicketCategory(ticketType)
		viewAction := []*modlv1.ActivityAction{{
			Label:   "View Ticket",
			Link:    "/panel/tickets/" + id,
			Primary: true,
		}}

		if creatorName == staffUsername && createdAt > cutoff {
			activities = append(activities, &modlv1.ActivityItemResponse{
				Id:          "ticket-created-" + id,
				Type:        "new_ticket",
				Color:       "blue",
				Title:       "Created ticket: " + subj,
				Time:        createdAt,
				Description: "Created " + catDisplay + " ticket",
				Actions:     viewAction,
			})
		}

		for _, rep := range replies {
			if rep.Created <= cutoff {
				continue
			}
			isStaffReply := strings.EqualFold(rep.Name, staffUsername)
			actionType := "New reply"
			color := "blue"
			replyName := rep.Name
			if replyName == "" {
				replyName = "Unknown"
			}
			desc := replyName + " replied to " + catDisplay + " ticket"
			if isStaffReply {
				actionType = "My reply"
				color = "green"
				desc = "You replied to " + catDisplay + " ticket"
			}
			activities = append(activities, &modlv1.ActivityItemResponse{
				Id:          fmt.Sprintf("ticket-reply-%s-%d", id, rep.Created),
				Type:        "mod_action",
				Color:       color,
				Title:       actionType + " on ticket: " + subj,
				Time:        rep.Created,
				Description: desc,
				Actions:     viewAction,
			})
		}
	}
	if err := tRows.Err(); err != nil {
		tRows.Close()
		return nil, err
	}
	tRows.Close()

	// 2. Punishment activities issued by staffUsername
	byOrdinal, byName := loadPunishmentTypes(ctx, s.db, srv.ID)
	pRows, err := s.db.QueryContext(ctx,
		`SELECT p.id, p.player_uuid, p.player_username, COALESCE(pl.current_username, ''), COALESCE(pl.usernames_json, '[]'),
		        p.type, p.punishment_type_id, p.reason, p.issuer_name, p.issued_at, p.expires_at, p.duration_ms, p.pardoned, p.data_json
		 FROM punishments p
		 LEFT JOIN players pl ON pl.server_id = p.server_id AND LOWER(pl.minecraft_uuid) = LOWER(p.player_uuid)
		 WHERE p.server_id = ? AND LOWER(p.issuer_name) = LOWER(?) AND p.issued_at >= ?
		 ORDER BY p.issued_at DESC
		 LIMIT ?`,
		srv.ID, staffUsername, cutoff, maxQueryResults,
	)
	if err != nil {
		return nil, err
	}
	defer pRows.Close()

	for pRows.Next() {
		var id, playerUUID, pUsername, curUsername, usernamesJSON, typeRaw, typeID, reason, issuerName, dataJSON string
		var issuedAt int64
		var expiresAt, durationMs sql.NullInt64
		var pardoned int
		if err := pRows.Scan(&id, &playerUUID, &pUsername, &curUsername, &usernamesJSON,
			&typeRaw, &typeID, &reason, &issuerName, &issuedAt, &expiresAt, &durationMs, &pardoned, &dataJSON); err != nil {
			return nil, err
		}
		parsed := parsePunishment(id, playerUUID, pUsername, typeRaw, typeID, reason, issuerName,
			issuedAt, expiresAt, durationMs, pardoned, dataJSON, byName)
		typeName := parsed.resolveTypeName(byOrdinal)
		username := pUsername
		if strings.TrimSpace(username) == "" {
			username = extractLatestUsername(curUsername, usernamesJSON)
		}
		activities = append(activities, &modlv1.ActivityItemResponse{
			Id:          "punishment-" + id,
			Type:        "new_punishment",
			Color:       "red",
			Title:       "Applied " + typeName + " to " + username,
			Time:        issuedAt,
			Description: "Applied " + typeName + " punishment",
			Actions: []*modlv1.ActivityAction{{
				Label:   "View Player",
				Link:    "/panel/players/" + playerUUID,
				Primary: true,
			}},
		})
	}
	if err := pRows.Err(); err != nil {
		return nil, err
	}

	sort.Slice(activities, func(i, j int) bool {
		return activities[i].Time > activities[j].Time
	})
	if len(activities) > safeLimit {
		activities = activities[:safeLimit]
	}
	return activities, nil
}

func (s *Service) ListVisibleAlerts(ctx context.Context, superAdmin bool) ([]*modlv1.PanelSystemAlertResponse, error) {
	nowMs := time.Now().UnixMilli()
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, message, severity, audience, expires_at
		 FROM system_alerts
		 WHERE expires_at IS NULL OR expires_at > ?
		 ORDER BY created_at DESC`,
		nowMs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*modlv1.PanelSystemAlertResponse, 0)
	for rows.Next() {
		var id, msg, sev, aud string
		var expiresAt sql.NullInt64
		if err := rows.Scan(&id, &msg, &sev, &aud, &expiresAt); err != nil {
			return nil, err
		}
		if !superAdmin {
			norm := strings.ToUpper(strings.TrimSpace(aud))
			if norm != "" && norm != "ALL_PANEL_USERS" && norm != "ALL" {
				continue
			}
		}
		var exp int64
		if expiresAt.Valid {
			exp = expiresAt.Int64
		}
		out = append(out, &modlv1.PanelSystemAlertResponse{
			Id:        id,
			Message:   msg,
			Severity:  sev,
			Audience:  aud,
			ExpiresAt: exp,
		})
	}
	return out, rows.Err()
}
