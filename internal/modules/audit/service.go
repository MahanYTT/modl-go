package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/httpx"

	"google.golang.org/protobuf/types/known/timestamppb"
)

var errCannotRollback = errors.New("this punishment cannot be rolled back")

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

type staffRow struct {
	ID                        string
	Username                  string
	Email                     string
	RoleID                    string
	AssignedMinecraftUUID     string
	AssignedMinecraftUsername string
	UpdatedAt                 int64
}

func (s *Service) loadStaff(ctx context.Context, serverID string) ([]staffRow, map[string]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, username, email, role_id, assigned_minecraft_uuid, assigned_minecraft_username, updated_at
		 FROM staff WHERE server_id = ?`, serverID,
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var list []staffRow
	byID := make(map[string]string)
	for rows.Next() {
		var r staffRow
		if err := rows.Scan(&r.ID, &r.Username, &r.Email, &r.RoleID, &r.AssignedMinecraftUUID, &r.AssignedMinecraftUsername, &r.UpdatedAt); err != nil {
			return nil, nil, err
		}
		list = append(list, r)
		if r.ID != "" && r.Username != "" {
			byID[r.ID] = r.Username
			byID[strings.ToLower(r.ID)] = r.Username
		}
		if r.AssignedMinecraftUUID != "" && r.Username != "" {
			byID[strings.ToLower(r.AssignedMinecraftUUID)] = r.Username
		}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return list, byID, nil
}

func (s *Service) loadRoleNames(ctx context.Context, serverID string) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name FROM staff_roles WHERE server_id = ?`, serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	names := map[string]string{"super-admin": "Super Admin"}
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		names[id] = name
	}
	return names, rows.Err()
}

type ticketReply struct {
	Name    string `json:"name"`
	Content string `json:"content"`
	Created int64  `json:"created"`
	Type    string `json:"type"`
	Staff   bool   `json:"staff"`
}

func (s *Service) GetStaffPerformance(ctx context.Context, serverID, period string) ([]*modlv1.StaffPerformanceResponse, error) {
	nowMs := time.Now().UnixMilli()
	startMs, hasStart := getStartDateMs(period, nowMs)

	staffList, staffNameByID, err := s.loadStaff(ctx, serverID)
	if err != nil {
		return nil, err
	}
	roleNames, err := s.loadRoleNames(ctx, serverID)
	if err != nil {
		return nil, err
	}

	// 1. Audit log counts & lastActive by lowercase source
	type logAgg struct {
		count      int32
		lastActive int64
	}
	activityBySource := make(map[string]*logAgg)
	lQuery := `SELECT source, created_at FROM audit_logs WHERE server_id = ?`
	lArgs := []any{serverID}
	if hasStart {
		lQuery += ` AND created_at >= ?`
		lArgs = append(lArgs, startMs)
	}
	lRows, err := s.db.QueryContext(ctx, lQuery, lArgs...)
	if err != nil {
		return nil, err
	}
	for lRows.Next() {
		var src string
		var createdAt int64
		if err := lRows.Scan(&src, &createdAt); err != nil {
			lRows.Close()
			return nil, err
		}
		key := strings.ToLower(strings.TrimSpace(src))
		if key == "" {
			continue
		}
		agg := activityBySource[key]
		if agg == nil {
			agg = &logAgg{}
			activityBySource[key] = agg
		}
		agg.count++
		if createdAt > agg.lastActive {
			agg.lastActive = createdAt
		}
	}
	if err := lRows.Err(); err != nil {
		lRows.Close()
		return nil, err
	}
	lRows.Close()

	// 2. Ticket replies count & response times per staff
	ticketRepliesByStaff := make(map[string]int32)
	responseTimesByStaff := make(map[string][]int32)

	tRows, err := s.db.QueryContext(ctx,
		`SELECT created_at, replies_json FROM tickets WHERE server_id = ? AND LOWER(status) != 'unfinished'`,
		serverID,
	)
	if err != nil {
		return nil, err
	}
	for tRows.Next() {
		var createdAt int64
		var repliesJSON string
		if err := tRows.Scan(&createdAt, &repliesJSON); err != nil {
			tRows.Close()
			return nil, err
		}
		var replies []ticketReply
		if err := json.Unmarshal([]byte(repliesJSON), &replies); err != nil || len(replies) == 0 {
			continue
		}
		firstReplyByStaff := make(map[string]int64)
		for _, rep := range replies {
			name := strings.ToLower(strings.TrimSpace(rep.Name))
			if name == "" {
				continue
			}
			if !hasStart || rep.Created >= startMs {
				ticketRepliesByStaff[name]++
			}
			if rep.Created > 0 {
				if prev, ok := firstReplyByStaff[name]; !ok || rep.Created < prev {
					firstReplyByStaff[name] = rep.Created
				}
			}
		}
		if !hasStart || createdAt >= startMs {
			for name, firstMs := range firstReplyByStaff {
				if firstMs >= createdAt && createdAt > 0 {
					mins := int32(math.Round(float64(firstMs-createdAt) / 60000.0))
					responseTimesByStaff[name] = append(responseTimesByStaff[name], mins)
				}
			}
		}
	}
	if err := tRows.Err(); err != nil {
		tRows.Close()
		return nil, err
	}
	tRows.Close()

	// 3. Punishments issued by staff
	punishmentsByKey := make(map[string]int32)
	pQuery := `SELECT issuer_uuid, issuer_name FROM punishments WHERE server_id = ?`
	pArgs := []any{serverID}
	if hasStart {
		pQuery += ` AND issued_at >= ?`
		pArgs = append(pArgs, startMs)
	}
	pRows, err := s.db.QueryContext(ctx, pQuery, pArgs...)
	if err != nil {
		return nil, err
	}
	for pRows.Next() {
		var issuerUUID, issuerName string
		if err := pRows.Scan(&issuerUUID, &issuerName); err != nil {
			pRows.Close()
			return nil, err
		}
		key := ""
		if issuerUUID != "" {
			if resolved, ok := staffNameByID[strings.ToLower(issuerUUID)]; ok {
				key = strings.ToLower(resolved)
			} else {
				key = strings.ToLower(issuerUUID)
			}
		}
		if key == "" && issuerName != "" {
			key = strings.ToLower(strings.TrimSpace(issuerName))
		}
		if key != "" {
			punishmentsByKey[key]++
		}
	}
	if err := pRows.Err(); err != nil {
		pRows.Close()
		return nil, err
	}
	pRows.Close()

	out := make([]*modlv1.StaffPerformanceResponse, 0, len(staffList))
	for _, st := range staffList {
		if strings.TrimSpace(st.Username) == "" {
			continue
		}
		lowerU := strings.ToLower(st.Username)
		agg := activityBySource[lowerU]
		var totalActions int32
		lastActive := st.UpdatedAt
		if agg != nil {
			totalActions = agg.count
			if agg.lastActive > 0 {
				lastActive = agg.lastActive
			}
		}
		if lastActive <= 0 {
			lastActive = nowMs
		}

		ticketActions := ticketRepliesByStaff[lowerU]
		modActions := punishmentsByKey[lowerU]
		if st.AssignedMinecraftUsername != "" && !strings.EqualFold(st.AssignedMinecraftUsername, st.Username) {
			modActions += punishmentsByKey[strings.ToLower(st.AssignedMinecraftUsername)]
		}
		if st.ID != "" {
			modActions += punishmentsByKey[strings.ToLower(st.ID)]
		}
		if ticketActions > 0 || modActions > 0 {
			if ticketActions+modActions > totalActions {
				totalActions = ticketActions + modActions
			}
		}

		roleName := "User"
		if st.RoleID != "" {
			if rn, ok := roleNames[st.RoleID]; ok && rn != "" {
				roleName = rn
			} else {
				roleName = st.RoleID
			}
		}

		var avgResp int32
		if times := responseTimesByStaff[lowerU]; len(times) > 0 {
			var sum int64
			for _, m := range times {
				sum += int64(m)
			}
			avgResp = int32(sum / int64(len(times)))
		}

		out = append(out, &modlv1.StaffPerformanceResponse{
			Id:                st.ID,
			Username:          st.Username,
			Role:              roleName,
			TotalActions:      totalActions,
			TicketResponses:   ticketActions,
			PunishmentsIssued: modActions,
			AvgResponseTime:   avgResp,
			LastActive:        timestamppb.New(time.UnixMilli(lastActive)),
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].TotalActions != out[j].TotalActions {
			return out[i].TotalActions > out[j].TotalActions
		}
		return out[i].Username < out[j].Username
	})
	return out, nil
}

func (s *Service) GetStaffDetails(ctx context.Context, serverID, username, period string) (*modlv1.StaffDetailsResponse, error) {
	nowMs := time.Now().UnixMilli()
	startMs, hasStart := getStartDateMs(period, nowMs)
	byOrdinal, byName := loadPunishmentTypes(ctx, s.db, serverID)

	staffList, _, err := s.loadStaff(ctx, serverID)
	if err != nil {
		return nil, err
	}

	searchNames := map[string]bool{strings.ToLower(strings.TrimSpace(username)): true}
	var staffID, staffMcUUID string
	for _, st := range staffList {
		if strings.EqualFold(st.Username, username) {
			staffID = st.ID
			staffMcUUID = st.AssignedMinecraftUUID
			if st.AssignedMinecraftUsername != "" {
				searchNames[strings.ToLower(st.AssignedMinecraftUsername)] = true
			}
			break
		}
	}

	// 1. Punishments
	pRows, err := s.db.QueryContext(ctx,
		`SELECT p.id, p.player_id, p.player_uuid, p.player_username,
		        COALESCE(pl.current_username, ''), COALESCE(pl.usernames_json, '[]'),
		        p.type, p.punishment_type_id, p.reason, p.issuer_uuid, p.issuer_name,
		        p.issued_at, p.expires_at, p.duration_ms, p.pardoned, p.notes_json, p.evidence_json, p.data_json
		 FROM punishments p
		 LEFT JOIN players pl ON pl.server_id = p.server_id AND LOWER(pl.minecraft_uuid) = LOWER(p.player_uuid)
		 WHERE p.server_id = ?
		 ORDER BY p.issued_at DESC`,
		serverID,
	)
	if err != nil {
		return nil, err
	}

	punishments := make([]*modlv1.StaffDetailsResponse_PunishmentDetail, 0)
	typeBreakdownCounts := make(map[string]int32)
	dailyMap := make(map[string]*modlv1.StaffDetailsResponse_DailyActivity)

	// Initialize last 14 days
	for i := 13; i >= 0; i-- {
		d := time.UnixMilli(nowMs - int64(i)*dayMs).UTC().Format("2006-01-02")
		dailyMap[d] = &modlv1.StaffDetailsResponse_DailyActivity{Date: d}
	}

	for pRows.Next() {
		var id, playerID, playerUUID, pUsername, curUsername, usernamesJSON string
		var typeRaw, typeID, reason, issuerUUID, issuerName, notesJSON, evidenceJSON, dataJSON string
		var issuedAt int64
		var expiresAt, durationMs sql.NullInt64
		var pardoned int
		if err := pRows.Scan(&id, &playerID, &playerUUID, &pUsername, &curUsername, &usernamesJSON,
			&typeRaw, &typeID, &reason, &issuerUUID, &issuerName, &issuedAt, &expiresAt, &durationMs,
			&pardoned, &notesJSON, &evidenceJSON, &dataJSON); err != nil {
			pRows.Close()
			return nil, err
		}

		matched := searchNames[strings.ToLower(strings.TrimSpace(issuerName))] ||
			(staffID != "" && strings.EqualFold(issuerUUID, staffID)) ||
			(staffMcUUID != "" && strings.EqualFold(issuerUUID, staffMcUUID))
		if !matched {
			continue
		}
		if hasStart && issuedAt < startMs {
			continue
		}

		parsed := parsePunishmentRow(id, playerID, playerUUID, pUsername, typeRaw, typeID, reason,
			issuerUUID, issuerName, issuedAt, expiresAt, durationMs, pardoned, notesJSON, evidenceJSON, dataJSON, byName)

		playerName := pUsername
		if strings.TrimSpace(playerName) == "" {
			playerName = extractLatestUsername(curUsername, usernamesJSON)
		}
		typeName := parsed.typeName(byOrdinal)
		typeBreakdownCounts[typeName]++

		pID := playerID
		if pID == "" {
			pID = playerUUID
		}
		punishments = append(punishments, &modlv1.StaffDetailsResponse_PunishmentDetail{
			Id:         parsed.ID,
			PlayerId:   pID,
			PlayerName: playerName,
			Type:       typeName,
			Reason:     parsed.Reason,
			Duration:   formatAuditDuration(parsed.DurationMs),
			Issued:     timestamppb.New(time.UnixMilli(parsed.IssuedAt)),
			Active:     parsed.isActive(nowMs),
			RolledBack: parsed.hasModification("ROLLBACK"),
		})

		dayKey := time.UnixMilli(parsed.IssuedAt).UTC().Format("2006-01-02")
		if entry := dailyMap[dayKey]; entry != nil {
			entry.Punishments++
		}
	}
	if err := pRows.Err(); err != nil {
		pRows.Close()
		return nil, err
	}
	pRows.Close()

	// 2. Tickets responded to by username
	tRows, err := s.db.QueryContext(ctx,
		`SELECT id, subject, type, status, created_at, updated_at, replies_json
		 FROM tickets WHERE server_id = ? AND LOWER(status) != 'unfinished'
		 ORDER BY updated_at DESC`,
		serverID,
	)
	if err != nil {
		return nil, err
	}

	tickets := make([]*modlv1.StaffDetailsResponse_TicketDetail, 0)
	var responseTimes []int32
	for tRows.Next() {
		var id, subject, ticketType, status, repliesJSON string
		var createdAt, updatedAt int64
		if err := tRows.Scan(&id, &subject, &ticketType, &status, &createdAt, &updatedAt, &repliesJSON); err != nil {
			tRows.Close()
			return nil, err
		}
		var replies []ticketReply
		if err := json.Unmarshal([]byte(repliesJSON), &replies); err != nil {
			continue
		}
		var firstUserReply int64
		repliedByStaff := false
		for _, rep := range replies {
			if strings.EqualFold(strings.TrimSpace(rep.Name), username) {
				repliedByStaff = true
				if rep.Created > 0 && (firstUserReply == 0 || rep.Created < firstUserReply) {
					firstUserReply = rep.Created
				}
				if rep.Created > 0 {
					dayKey := time.UnixMilli(rep.Created).UTC().Format("2006-01-02")
					if entry := dailyMap[dayKey]; entry != nil {
						entry.Tickets++
					}
				}
			}
		}
		if !repliedByStaff {
			continue
		}
		if hasStart && updatedAt < startMs {
			continue
		}
		var respMins int32
		if firstUserReply >= createdAt && createdAt > 0 {
			respMins = int32(math.Round(float64(firstUserReply-createdAt) / 60000.0))
			responseTimes = append(responseTimes, respMins)
		}
		tickets = append(tickets, &modlv1.StaffDetailsResponse_TicketDetail{
			Id:           id,
			Subject:      subject,
			Category:     ticketType,
			Status:       status,
			LastActivity: timestamppb.New(time.UnixMilli(updatedAt)),
			ResponseTime: respMins,
		})
	}
	if err := tRows.Err(); err != nil {
		tRows.Close()
		return nil, err
	}
	tRows.Close()

	// 3. Evidence uploads from audit_logs
	eQuery := `SELECT created_at, description FROM audit_logs WHERE server_id = ? AND LOWER(source) = LOWER(?)`
	eArgs := []any{serverID, username}
	if hasStart {
		eQuery += ` AND created_at >= ?`
		eArgs = append(eArgs, startMs)
	}
	eRows, err := s.db.QueryContext(ctx, eQuery, eArgs...)
	if err != nil {
		return nil, err
	}
	var evidenceUploads int32
	for eRows.Next() {
		var createdAt int64
		var desc string
		if err := eRows.Scan(&createdAt, &desc); err != nil {
			eRows.Close()
			return nil, err
		}
		low := strings.ToLower(desc)
		if strings.Contains(low, "evidence") || strings.Contains(low, "upload") {
			evidenceUploads++
			dayKey := time.UnixMilli(createdAt).UTC().Format("2006-01-02")
			if entry := dailyMap[dayKey]; entry != nil {
				entry.Evidence++
			}
		}
	}
	if err := eRows.Err(); err != nil {
		eRows.Close()
		return nil, err
	}
	eRows.Close()

	dailyKeys := make([]string, 0, len(dailyMap))
	for k := range dailyMap {
		dailyKeys = append(dailyKeys, k)
	}
	sort.Strings(dailyKeys)
	dailyActivity := make([]*modlv1.StaffDetailsResponse_DailyActivity, 0, len(dailyKeys))
	for _, k := range dailyKeys {
		dailyActivity = append(dailyActivity, dailyMap[k])
	}

	breakdown := make([]*modlv1.StaffDetailsResponse_PunishmentTypeBreakdown, 0, len(typeBreakdownCounts))
	for t, c := range typeBreakdownCounts {
		breakdown = append(breakdown, &modlv1.StaffDetailsResponse_PunishmentTypeBreakdown{Type: t, Count: c})
	}
	sort.Slice(breakdown, func(i, j int) bool {
		if breakdown[i].Count != breakdown[j].Count {
			return breakdown[i].Count > breakdown[j].Count
		}
		return breakdown[i].Type < breakdown[j].Type
	})

	var avgResp int32
	if len(responseTimes) > 0 {
		var sum int64
		for _, m := range responseTimes {
			sum += int64(m)
		}
		avgResp = int32(sum / int64(len(responseTimes)))
	}

	return &modlv1.StaffDetailsResponse{
		Username:                username,
		Period:                  period,
		Punishments:             punishments,
		Tickets:                 tickets,
		DailyActivity:           dailyActivity,
		PunishmentTypeBreakdown: breakdown,
		EvidenceUploads:         evidenceUploads,
		Summary: &modlv1.StaffDetailsResponse_Summary{
			TotalPunishments: int32(len(punishments)),
			TotalTickets:     int32(len(tickets)),
			AvgResponseTime:  avgResp,
			EvidenceUploads:  evidenceUploads,
		},
	}, nil
}

func (s *Service) GetPunishmentsList(ctx context.Context, serverID, statusFilter string) ([]*modlv1.ActivePunishmentResponse, error) {
	nowMs := time.Now().UnixMilli()
	byOrdinal, byName := loadPunishmentTypes(ctx, s.db, serverID)
	_, staffNameByID, err := s.loadStaff(ctx, serverID)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT p.id, p.player_id, p.player_uuid, p.player_username,
		        COALESCE(pl.current_username, ''), COALESCE(pl.usernames_json, '[]'),
		        p.type, p.punishment_type_id, p.reason, p.issuer_uuid, p.issuer_name,
		        p.issued_at, p.expires_at, p.duration_ms, p.pardoned, p.notes_json, p.evidence_json, p.data_json
		 FROM punishments p
		 LEFT JOIN players pl ON pl.server_id = p.server_id AND LOWER(pl.minecraft_uuid) = LOWER(p.player_uuid)
		 WHERE p.server_id = ?
		 ORDER BY p.issued_at DESC`,
		serverID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	filterActive := strings.EqualFold(statusFilter, "active")
	filterInactive := strings.EqualFold(statusFilter, "inactive")

	out := make([]*modlv1.ActivePunishmentResponse, 0)
	for rows.Next() {
		var id, playerID, playerUUID, pUsername, curUsername, usernamesJSON string
		var typeRaw, typeID, reason, issuerUUID, issuerName, notesJSON, evidenceJSON, dataJSON string
		var issuedAt int64
		var expiresAt, durationMs sql.NullInt64
		var pardoned int
		if err := rows.Scan(&id, &playerID, &playerUUID, &pUsername, &curUsername, &usernamesJSON,
			&typeRaw, &typeID, &reason, &issuerUUID, &issuerName, &issuedAt, &expiresAt, &durationMs,
			&pardoned, &notesJSON, &evidenceJSON, &dataJSON); err != nil {
			return nil, err
		}

		parsed := parsePunishmentRow(id, playerID, playerUUID, pUsername, typeRaw, typeID, reason,
			issuerUUID, issuerName, issuedAt, expiresAt, durationMs, pardoned, notesJSON, evidenceJSON, dataJSON, byName)
		active := parsed.isActive(nowMs)
		if filterActive && !active {
			continue
		}
		if filterInactive && active {
			continue
		}

		playerName := pUsername
		if strings.TrimSpace(playerName) == "" {
			playerName = extractLatestUsername(curUsername, usernamesJSON)
		}
		pID := playerID
		if pID == "" {
			pID = playerUUID
		}
		staffName := issuerName
		if issuerUUID != "" {
			if resolved, ok := staffNameByID[strings.ToLower(issuerUUID)]; ok && resolved != "" {
				staffName = resolved
			}
		}
		if staffName == "" {
			if issuerUUID != "" {
				staffName = "Unknown Staff"
			} else {
				staffName = "Console"
			}
		}

		evidenceItems := make([]*modlv1.ActivePunishmentResponse_EvidenceItem, 0, len(parsed.Evidence))
		for _, ev := range parsed.Evidence {
			evidenceItems = append(evidenceItems, &modlv1.ActivePunishmentResponse_EvidenceItem{
				Text:     ev.Text,
				Url:      ev.URL,
				Type:     ev.Type,
				FileName: ev.FileName,
			})
		}

		var startedPb, expiresPb *timestamppb.Timestamp
		if parsed.StartedAt != nil && *parsed.StartedAt > 0 {
			startedPb = timestamppb.New(time.UnixMilli(*parsed.StartedAt))
		}
		if exp, ok := parsed.effectiveExpiryMs(); ok && exp > 0 {
			expiresPb = timestamppb.New(time.UnixMilli(exp))
		}

		var durPtr *int64
		if parsed.DurationMs != nil {
			v := *parsed.DurationMs
			durPtr = &v
		}

		out = append(out, &modlv1.ActivePunishmentResponse{
			Id:                parsed.ID,
			PlayerId:          pID,
			PlayerName:        playerName,
			Type:              parsed.typeName(byOrdinal),
			TypeOrdinal:       int32(parsed.TypeOrdinal),
			Category:          parsed.categoryName(byOrdinal),
			StaffName:         staffName,
			Reason:            parsed.Reason,
			Duration:          durPtr,
			Issued:            timestamppb.New(time.UnixMilli(parsed.IssuedAt)),
			Started:           startedPb,
			Expires:           expiresPb,
			Active:            active,
			HasEvidence:       len(evidenceItems) > 0,
			EvidenceCount:     int32(len(evidenceItems)),
			Evidence:          evidenceItems,
			AttachedTicketIds: parsed.AttachedTicketIDs,
		})
	}
	return out, rows.Err()
}

func extractPunishmentTypeFromDesc(desc string) string {
	low := strings.ToLower(desc)
	switch {
	case strings.Contains(low, "ban"):
		return "Ban"
	case strings.Contains(low, "mute"):
		return "Mute"
	case strings.Contains(low, "kick"):
		return "Kick"
	case strings.Contains(low, "warn"):
		return "Warn"
	default:
		return "Unknown"
	}
}

func (s *Service) GetPunishmentAuditLogs(ctx context.Context, serverID string, limit int, canRollbackOnly bool) ([]*modlv1.PunishmentAuditResponse, error) {
	if limit < 1 {
		limit = 1
	} else if limit > 100 {
		limit = 100
	}
	thirtyDaysAgo := time.Now().UnixMilli() - 30*dayMs

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, created_at, source, description, metadata_json
		 FROM audit_logs
		 WHERE server_id = ? AND level = 'moderation' AND created_at >= ?
		 ORDER BY created_at DESC`,
		serverID, thirtyDaysAgo,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*modlv1.PunishmentAuditResponse, 0)
	for rows.Next() {
		var id, source, desc, metaJSON string
		var createdAt int64
		if err := rows.Scan(&id, &createdAt, &source, &desc, &metaJSON); err != nil {
			return nil, err
		}
		meta := map[string]any{}
		if strings.TrimSpace(metaJSON) != "" && metaJSON != "{}" {
			_ = json.Unmarshal([]byte(metaJSON), &meta)
		}
		canRollback := true
		if v, ok := meta["canRollback"].(bool); ok && !v {
			canRollback = false
		}
		if canRollbackOnly && !canRollback {
			continue
		}

		playerID, _ := meta["playerId"].(string)
		if playerID == "" {
			playerID = "unknown"
		}
		playerName, _ := meta["playerName"].(string)
		if playerName == "" {
			playerName = "Unknown"
		}
		staffID, _ := meta["staffId"].(string)
		if staffID == "" {
			staffID = source
		}
		reason, _ := meta["reason"].(string)
		if reason == "" {
			reason = desc
		}
		duration, _ := meta["duration"].(string)
		typeName, _ := meta["punishmentType"].(string)
		if typeName == "" {
			typeName = extractPunishmentTypeFromDesc(desc)
		}

		out = append(out, &modlv1.PunishmentAuditResponse{
			Id:          id,
			Type:        typeName,
			PlayerId:    playerID,
			PlayerName:  playerName,
			StaffId:     staffID,
			StaffName:   source,
			Reason:      reason,
			Duration:    duration,
			Timestamp:   timestamppb.New(time.UnixMilli(createdAt)),
			CanRollback: canRollback,
		})
		if len(out) >= limit {
			break
		}
	}
	return out, rows.Err()
}

func (s *Service) saveAuditLog(ctx context.Context, serverID, level, source, description string, metadata map[string]any, nowMs int64) error {
	metaBytes, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO audit_logs (id, server_id, created_at, level, source, description, metadata_json)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		httpx.NewID(), serverID, nowMs, level, source, description, string(metaBytes),
	)
	return err
}

func (s *Service) applyRollbackToRow(ctx context.Context, serverID string, p parsedPunishment, playerName, reason, performerUsername string, bulk bool, nowMs int64, byOrdinal map[int]punishmentTypeMeta) error {
	mod := modificationEntry{
		ID:         httpx.NewID()[:8],
		Type:       "ROLLBACK",
		Date:       nowMs,
		IssuerName: performerUsername,
		Reason:     reason,
	}
	p.Modifications = append(p.Modifications, mod)
	if p.DataMap == nil {
		p.DataMap = map[string]any{}
	}
	p.DataMap["modifications"] = p.Modifications
	p.DataMap["status"] = "PARDONED"

	dataBytes, err := json.Marshal(p.DataMap)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx,
		`UPDATE punishments
		 SET pardoned = 1, pardoned_by = ?, pardoned_at = ?, pardon_reason = ?, data_json = ?, updated_at = ?
		 WHERE id = ? AND server_id = ?`,
		performerUsername, nowMs, reason, string(dataBytes), nowMs, p.ID, serverID,
	)
	if err != nil {
		return err
	}

	typeName := p.typeName(byOrdinal)
	dispPlayer := playerName
	if strings.TrimSpace(dispPlayer) == "" {
		dispPlayer = "unknown player"
	}
	desc := fmt.Sprintf("Rolled back %s for %s", typeName, dispPlayer)
	if bulk {
		desc = fmt.Sprintf("Bulk rollback: %s for %s (issued by %s)", typeName, dispPlayer, p.IssuerName)
	}
	pID := p.PlayerID
	if pID == "" {
		pID = p.PlayerUUID
	}
	return s.saveAuditLog(ctx, serverID, "moderation", performerUsername, desc, map[string]any{
		"punishmentId":   p.ID,
		"playerId":       pID,
		"playerName":     playerName,
		"staffUsername":  p.IssuerName,
		"rollbackReason": reason,
		"punishmentType": typeName,
		"bulkRollback":   bulk,
		"canRollback":    false,
	}, nowMs)
}

func (s *Service) RollbackPunishment(ctx context.Context, serverID, punishmentOrAuditID, reason, performerUsername string) (bool, error) {
	byOrdinal, byName := loadPunishmentTypes(ctx, s.db, serverID)
	targetID := strings.TrimSpace(punishmentOrAuditID)

	// Allow lookup either directly by punishment.id or via audit_logs.metadata_json.punishmentId
	row := s.db.QueryRowContext(ctx,
		`SELECT p.id, p.player_id, p.player_uuid, p.player_username,
		        COALESCE(pl.current_username, ''), COALESCE(pl.usernames_json, '[]'),
		        p.type, p.punishment_type_id, p.reason, p.issuer_uuid, p.issuer_name,
		        p.issued_at, p.expires_at, p.duration_ms, p.pardoned, p.notes_json, p.evidence_json, p.data_json
		 FROM punishments p
		 LEFT JOIN players pl ON pl.server_id = p.server_id AND LOWER(pl.minecraft_uuid) = LOWER(p.player_uuid)
		 WHERE p.server_id = ? AND p.id = ?
		 LIMIT 1`,
		serverID, targetID,
	)

	var id, playerID, playerUUID, pUsername, curUsername, usernamesJSON string
	var typeRaw, typeID, pReason, issuerUUID, issuerName, notesJSON, evidenceJSON, dataJSON string
	var issuedAt int64
	var expiresAt, durationMs sql.NullInt64
	var pardoned int
	err := row.Scan(&id, &playerID, &playerUUID, &pUsername, &curUsername, &usernamesJSON,
		&typeRaw, &typeID, &pReason, &issuerUUID, &issuerName, &issuedAt, &expiresAt, &durationMs,
		&pardoned, &notesJSON, &evidenceJSON, &dataJSON)
	if err == sql.ErrNoRows {
		// Check if targetID is an audit_log id referencing a punishmentId
		var metaJSON string
		if aErr := s.db.QueryRowContext(ctx,
			`SELECT metadata_json FROM audit_logs WHERE server_id = ? AND id = ? LIMIT 1`,
			serverID, targetID,
		).Scan(&metaJSON); aErr == nil {
			var meta map[string]any
			if json.Unmarshal([]byte(metaJSON), &meta) == nil {
				if pid, ok := meta["punishmentId"].(string); ok && pid != "" {
					err = s.db.QueryRowContext(ctx,
						`SELECT p.id, p.player_id, p.player_uuid, p.player_username,
						        COALESCE(pl.current_username, ''), COALESCE(pl.usernames_json, '[]'),
						        p.type, p.punishment_type_id, p.reason, p.issuer_uuid, p.issuer_name,
						        p.issued_at, p.expires_at, p.duration_ms, p.pardoned, p.notes_json, p.evidence_json, p.data_json
						 FROM punishments p
						 LEFT JOIN players pl ON pl.server_id = p.server_id AND LOWER(pl.minecraft_uuid) = LOWER(p.player_uuid)
						 WHERE p.server_id = ? AND p.id = ?
						 LIMIT 1`,
						serverID, pid,
					).Scan(&id, &playerID, &playerUUID, &pUsername, &curUsername, &usernamesJSON,
						&typeRaw, &typeID, &pReason, &issuerUUID, &issuerName, &issuedAt, &expiresAt, &durationMs,
						&pardoned, &notesJSON, &evidenceJSON, &dataJSON)
				}
			}
		}
	}
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	parsed := parsePunishmentRow(id, playerID, playerUUID, pUsername, typeRaw, typeID, pReason,
		issuerUUID, issuerName, issuedAt, expiresAt, durationMs, pardoned, notesJSON, evidenceJSON, dataJSON, byName)
	if parsed.hasModification("ROLLBACK") {
		return false, errCannotRollback
	}
	playerName := pUsername
	if strings.TrimSpace(playerName) == "" {
		playerName = extractLatestUsername(curUsername, usernamesJSON)
	}
	if strings.TrimSpace(reason) == "" {
		reason = "Admin rollback"
	}
	if err := s.applyRollbackToRow(ctx, serverID, parsed, playerName, reason, performerUsername, false, time.Now().UnixMilli(), byOrdinal); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Service) RollbackByStaff(ctx context.Context, serverID, staffUsername string, startMs, endMs *int64, reason, performerUsername string) (int32, error) {
	byOrdinal, byName := loadPunishmentTypes(ctx, s.db, serverID)
	staffList, _, err := s.loadStaff(ctx, serverID)
	if err != nil {
		return 0, err
	}

	matchNames := map[string]bool{strings.ToLower(strings.TrimSpace(staffUsername)): true}
	var staffID, staffMcUUID string
	for _, st := range staffList {
		if strings.EqualFold(st.Username, staffUsername) {
			staffID = st.ID
			staffMcUUID = st.AssignedMinecraftUUID
			if st.AssignedMinecraftUsername != "" {
				matchNames[strings.ToLower(st.AssignedMinecraftUsername)] = true
			}
			break
		}
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT p.id, p.player_id, p.player_uuid, p.player_username,
		        COALESCE(pl.current_username, ''), COALESCE(pl.usernames_json, '[]'),
		        p.type, p.punishment_type_id, p.reason, p.issuer_uuid, p.issuer_name,
		        p.issued_at, p.expires_at, p.duration_ms, p.pardoned, p.notes_json, p.evidence_json, p.data_json
		 FROM punishments p
		 LEFT JOIN players pl ON pl.server_id = p.server_id AND LOWER(pl.minecraft_uuid) = LOWER(p.player_uuid)
		 WHERE p.server_id = ?`,
		serverID,
	)
	if err != nil {
		return 0, err
	}

	type targetRow struct {
		parsed     parsedPunishment
		playerName string
	}
	var targets []targetRow
	for rows.Next() {
		var id, playerID, playerUUID, pUsername, curUsername, usernamesJSON string
		var typeRaw, typeID, pReason, issuerUUID, issuerName, notesJSON, evidenceJSON, dataJSON string
		var issuedAt int64
		var expiresAt, durationMs sql.NullInt64
		var pardoned int
		if err := rows.Scan(&id, &playerID, &playerUUID, &pUsername, &curUsername, &usernamesJSON,
			&typeRaw, &typeID, &pReason, &issuerUUID, &issuerName, &issuedAt, &expiresAt, &durationMs,
			&pardoned, &notesJSON, &evidenceJSON, &dataJSON); err != nil {
			rows.Close()
			return 0, err
		}

		matched := matchNames[strings.ToLower(strings.TrimSpace(issuerName))] ||
			(staffID != "" && strings.EqualFold(issuerUUID, staffID)) ||
			(staffMcUUID != "" && strings.EqualFold(issuerUUID, staffMcUUID))
		if !matched {
			continue
		}
		if startMs != nil && endMs != nil {
			if issuedAt < *startMs || issuedAt > *endMs {
				continue
			}
		}

		parsed := parsePunishmentRow(id, playerID, playerUUID, pUsername, typeRaw, typeID, pReason,
			issuerUUID, issuerName, issuedAt, expiresAt, durationMs, pardoned, notesJSON, evidenceJSON, dataJSON, byName)
		if parsed.hasModification("ROLLBACK") {
			continue
		}

		playerName := pUsername
		if strings.TrimSpace(playerName) == "" {
			playerName = extractLatestUsername(curUsername, usernamesJSON)
		}
		targets = append(targets, targetRow{parsed: parsed, playerName: playerName})
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, err
	}
	rows.Close()

	if strings.TrimSpace(reason) == "" {
		reason = "Bulk rollback by admin"
	}
	nowMs := time.Now().UnixMilli()
	var count int32
	for _, t := range targets {
		if err := s.applyRollbackToRow(ctx, serverID, t.parsed, t.playerName, reason, performerUsername, true, nowMs, byOrdinal); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (s *Service) BulkPardonByType(ctx context.Context, serverID string, typeOrdinals []int32, reason, performerUsername string) (int32, error) {
	if len(typeOrdinals) == 0 {
		return 0, nil
	}
	ordSet := make(map[int]bool, len(typeOrdinals))
	for _, o := range typeOrdinals {
		ordSet[int(o)] = true
	}

	nowMs := time.Now().UnixMilli()
	byOrdinal, byName := loadPunishmentTypes(ctx, s.db, serverID)

	rows, err := s.db.QueryContext(ctx,
		`SELECT p.id, p.player_id, p.player_uuid, p.player_username,
		        COALESCE(pl.current_username, ''), COALESCE(pl.usernames_json, '[]'),
		        p.type, p.punishment_type_id, p.reason, p.issuer_uuid, p.issuer_name,
		        p.issued_at, p.expires_at, p.duration_ms, p.pardoned, p.notes_json, p.evidence_json, p.data_json
		 FROM punishments p
		 LEFT JOIN players pl ON pl.server_id = p.server_id AND LOWER(pl.minecraft_uuid) = LOWER(p.player_uuid)
		 WHERE p.server_id = ?`,
		serverID,
	)
	if err != nil {
		return 0, err
	}

	type targetRow struct {
		parsed     parsedPunishment
		playerName string
	}
	var targets []targetRow
	for rows.Next() {
		var id, playerID, playerUUID, pUsername, curUsername, usernamesJSON string
		var typeRaw, typeID, pReason, issuerUUID, issuerName, notesJSON, evidenceJSON, dataJSON string
		var issuedAt int64
		var expiresAt, durationMs sql.NullInt64
		var pardoned int
		if err := rows.Scan(&id, &playerID, &playerUUID, &pUsername, &curUsername, &usernamesJSON,
			&typeRaw, &typeID, &pReason, &issuerUUID, &issuerName, &issuedAt, &expiresAt, &durationMs,
			&pardoned, &notesJSON, &evidenceJSON, &dataJSON); err != nil {
			rows.Close()
			return 0, err
		}

		parsed := parsePunishmentRow(id, playerID, playerUUID, pUsername, typeRaw, typeID, pReason,
			issuerUUID, issuerName, issuedAt, expiresAt, durationMs, pardoned, notesJSON, evidenceJSON, dataJSON, byName)
		if !ordSet[parsed.TypeOrdinal] || !parsed.isActive(nowMs) {
			continue
		}
		if parsed.hasModification("MANUAL_PARDON", "APPEAL_ACCEPT", "SYSTEM_PARDON") {
			continue
		}

		playerName := pUsername
		if strings.TrimSpace(playerName) == "" {
			playerName = extractLatestUsername(curUsername, usernamesJSON)
		}
		targets = append(targets, targetRow{parsed: parsed, playerName: playerName})
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, err
	}
	rows.Close()

	var count int32
	for _, t := range targets {
		p := t.parsed
		p.Modifications = append(p.Modifications, modificationEntry{
			ID:         httpx.NewID()[:8],
			Type:       "MANUAL_PARDON",
			Date:       nowMs,
			IssuerName: performerUsername,
			Reason:     reason,
		})
		if p.DataMap == nil {
			p.DataMap = map[string]any{}
		}
		p.DataMap["modifications"] = p.Modifications
		p.DataMap["status"] = "PARDONED"

		dataBytes, err := json.Marshal(p.DataMap)
		if err != nil {
			return count, err
		}

		if _, err := s.db.ExecContext(ctx,
			`UPDATE punishments
			 SET pardoned = 1, pardoned_by = ?, pardoned_at = ?, pardon_reason = ?, data_json = ?, updated_at = ?
			 WHERE id = ? AND server_id = ?`,
			performerUsername, nowMs, reason, string(dataBytes), nowMs, p.ID, serverID,
		); err != nil {
			return count, err
		}

		typeName := p.typeName(byOrdinal)
		pID := p.PlayerID
		if pID == "" {
			pID = p.PlayerUUID
		}
		_ = s.saveAuditLog(ctx, serverID, "moderation", performerUsername,
			fmt.Sprintf("Bulk pardon: %s for %s", typeName, t.playerName),
			map[string]any{
				"punishmentId":   p.ID,
				"playerId":       pID,
				"playerName":     t.playerName,
				"punishmentType": typeName,
				"typeOrdinal":    p.TypeOrdinal,
				"pardonReason":   reason,
				"bulkPardon":     true,
				"canRollback":    false,
			}, nowMs)
		count++
	}
	return count, nil
}

func (s *Service) BulkSetExpirationByType(ctx context.Context, serverID string, typeOrdinals []int32, newDurationMs int64, reason, performerUsername string) (int32, error) {
	if len(typeOrdinals) == 0 {
		return 0, nil
	}
	ordSet := make(map[int]bool, len(typeOrdinals))
	for _, o := range typeOrdinals {
		ordSet[int(o)] = true
	}

	effectiveDuration := int64(-1)
	if newDurationMs > 0 {
		effectiveDuration = newDurationMs
	}
	nowMs := time.Now().UnixMilli()
	byOrdinal, byName := loadPunishmentTypes(ctx, s.db, serverID)

	rows, err := s.db.QueryContext(ctx,
		`SELECT p.id, p.player_id, p.player_uuid, p.player_username,
		        COALESCE(pl.current_username, ''), COALESCE(pl.usernames_json, '[]'),
		        p.type, p.punishment_type_id, p.reason, p.issuer_uuid, p.issuer_name,
		        p.issued_at, p.expires_at, p.duration_ms, p.pardoned, p.notes_json, p.evidence_json, p.data_json
		 FROM punishments p
		 LEFT JOIN players pl ON pl.server_id = p.server_id AND LOWER(pl.minecraft_uuid) = LOWER(p.player_uuid)
		 WHERE p.server_id = ?`,
		serverID,
	)
	if err != nil {
		return 0, err
	}

	type targetRow struct {
		parsed     parsedPunishment
		playerName string
	}
	var targets []targetRow
	for rows.Next() {
		var id, playerID, playerUUID, pUsername, curUsername, usernamesJSON string
		var typeRaw, typeID, pReason, issuerUUID, issuerName, notesJSON, evidenceJSON, dataJSON string
		var issuedAt int64
		var expiresAt, durationMs sql.NullInt64
		var pardoned int
		if err := rows.Scan(&id, &playerID, &playerUUID, &pUsername, &curUsername, &usernamesJSON,
			&typeRaw, &typeID, &pReason, &issuerUUID, &issuerName, &issuedAt, &expiresAt, &durationMs,
			&pardoned, &notesJSON, &evidenceJSON, &dataJSON); err != nil {
			rows.Close()
			return 0, err
		}

		parsed := parsePunishmentRow(id, playerID, playerUUID, pUsername, typeRaw, typeID, pReason,
			issuerUUID, issuerName, issuedAt, expiresAt, durationMs, pardoned, notesJSON, evidenceJSON, dataJSON, byName)
		if !ordSet[parsed.TypeOrdinal] || !parsed.isActive(nowMs) {
			continue
		}

		playerName := pUsername
		if strings.TrimSpace(playerName) == "" {
			playerName = extractLatestUsername(curUsername, usernamesJSON)
		}
		targets = append(targets, targetRow{parsed: parsed, playerName: playerName})
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, err
	}
	rows.Close()

	var count int32
	for _, t := range targets {
		p := t.parsed
		durVal := effectiveDuration
		p.Modifications = append(p.Modifications, modificationEntry{
			ID:                httpx.NewID()[:8],
			Type:              "MANUAL_DURATION_CHANGE",
			Date:              nowMs,
			IssuerName:        performerUsername,
			Reason:            reason,
			EffectiveDuration: &durVal,
		})
		if p.DataMap == nil {
			p.DataMap = map[string]any{}
		}
		p.DataMap["modifications"] = p.Modifications
		p.DataMap["duration"] = effectiveDuration

		dataBytes, err := json.Marshal(p.DataMap)
		if err != nil {
			return count, err
		}

		var newExpires sql.NullInt64
		if effectiveDuration > 0 {
			newExpires = sql.NullInt64{Int64: nowMs + effectiveDuration, Valid: true}
		}

		if _, err := s.db.ExecContext(ctx,
			`UPDATE punishments
			 SET duration_ms = ?, expires_at = ?, data_json = ?, updated_at = ?
			 WHERE id = ? AND server_id = ?`,
			effectiveDuration, newExpires, string(dataBytes), nowMs, p.ID, serverID,
		); err != nil {
			return count, err
		}

		typeName := p.typeName(byOrdinal)
		pID := p.PlayerID
		if pID == "" {
			pID = p.PlayerUUID
		}
		_ = s.saveAuditLog(ctx, serverID, "moderation", performerUsername,
			fmt.Sprintf("Bulk duration change: %s for %s", typeName, t.playerName),
			map[string]any{
				"punishmentId":       p.ID,
				"playerId":           pID,
				"playerName":         t.playerName,
				"punishmentType":     typeName,
				"typeOrdinal":        p.TypeOrdinal,
				"reason":             reason,
				"newDurationMs":      newDurationMs,
				"bulkDurationChange": true,
				"canRollback":        false,
			}, nowMs)
		count++
	}
	return count, nil
}
