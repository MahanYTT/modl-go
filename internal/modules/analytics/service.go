package analytics

import (
	"context"
	"database/sql"
	"encoding/json"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/models"
)

const dayMs = 24 * 60 * 60 * 1000

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
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

func (s *Service) serverLocation(ctx context.Context, serverID string) *time.Location {
	var dataJSON string
	err := s.db.QueryRowContext(ctx,
		`SELECT data_json FROM settings WHERE server_id = ? AND type = 'general' LIMIT 1`,
		serverID,
	).Scan(&dataJSON)
	if err != nil || strings.TrimSpace(dataJSON) == "" {
		return time.UTC
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(dataJSON), &m); err != nil {
		return time.UTC
	}
	for _, key := range []string{"analyticsTimeZone", "timezone", "timeZone"} {
		if tz, ok := m[key].(string); ok && strings.TrimSpace(tz) != "" {
			if loc, err := time.LoadLocation(strings.TrimSpace(tz)); err == nil {
				return loc
			}
		}
	}
	return time.UTC
}

func parseDateValue(raw any) (int64, bool) {
	if raw == nil {
		return 0, false
	}
	switch v := raw.(type) {
	case float64:
		if v <= 0 {
			return 0, false
		}
		n := int64(v)
		if n < 100_000_000_000 {
			return n * 1000, true
		}
		return n, true
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			return 0, false
		}
		if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
			return t.UnixMilli(), true
		}
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return t.UnixMilli(), true
		}
		if n, err := strconv.ParseInt(s, 10, 64); err == nil && n > 0 {
			if n < 100_000_000_000 {
				return n * 1000, true
			}
			return n, true
		}
	}
	return 0, false
}

type ipEntry struct {
	Country    string `json:"country"`
	Proxy      bool   `json:"proxy"`
	Hosting    bool   `json:"hosting"`
	FirstLogin any    `json:"firstLogin"`
	Logins     []any  `json:"logins"`
}

func earliestPlayerLoginMs(firstSeen int64, ipAddressesJSON string) (int64, bool) {
	var ips []ipEntry
	if err := json.Unmarshal([]byte(ipAddressesJSON), &ips); err == nil && len(ips) > 0 {
		var minMs int64
		found := false
		for _, ip := range ips {
			if ms, ok := parseDateValue(ip.FirstLogin); ok {
				if !found || ms < minMs {
					minMs = ms
					found = true
				}
			}
		}
		if found {
			return minMs, true
		}
	}
	if firstSeen > 0 {
		return firstSeen, true
	}
	return 0, false
}

func calcPercentChange(recent, previous int64) int32 {
	if previous == 0 {
		if recent > 0 {
			return 100
		}
		return 0
	}
	return int32(math.Round(float64(recent-previous) / float64(previous) * 100))
}

func (s *Service) GetOverview(ctx context.Context, srv *models.Server) (*modlv1.AnalyticsOverviewResponse, error) {
	nowMs := time.Now().UnixMilli()
	thirtyDaysAgo := nowMs - 30*dayMs
	sixtyDaysAgo := nowMs - 60*dayMs

	var totalTickets, activeTickets, recentTickets, previousTickets int64
	if err := s.db.QueryRowContext(ctx,
		`SELECT
		   COUNT(*),
		   COALESCE(SUM(CASE WHEN LOWER(status) = 'open' THEN 1 ELSE 0 END), 0),
		   COALESCE(SUM(CASE WHEN created_at >= ? THEN 1 ELSE 0 END), 0),
		   COALESCE(SUM(CASE WHEN created_at >= ? AND created_at < ? THEN 1 ELSE 0 END), 0)
		 FROM tickets WHERE server_id = ?`,
		thirtyDaysAgo, sixtyDaysAgo, thirtyDaysAgo, srv.ID,
	).Scan(&totalTickets, &activeTickets, &recentTickets, &previousTickets); err != nil {
		return nil, err
	}

	var totalPlayers, recentPlayers, previousPlayers int64
	pRows, err := s.db.QueryContext(ctx,
		`SELECT first_seen, ip_addresses_json FROM players WHERE server_id = ?`, srv.ID,
	)
	if err != nil {
		return nil, err
	}
	for pRows.Next() {
		var firstSeen int64
		var ipJSON string
		if err := pRows.Scan(&firstSeen, &ipJSON); err != nil {
			pRows.Close()
			return nil, err
		}
		totalPlayers++
		if earliest, ok := earliestPlayerLoginMs(firstSeen, ipJSON); ok {
			if earliest >= thirtyDaysAgo {
				recentPlayers++
			} else if earliest >= sixtyDaysAgo && earliest < thirtyDaysAgo {
				previousPlayers++
			}
		}
	}
	if err := pRows.Err(); err != nil {
		pRows.Close()
		return nil, err
	}
	pRows.Close()

	var totalStaff int64
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM staff WHERE server_id = ?`, srv.ID).Scan(&totalStaff); err != nil {
		return nil, err
	}
	if strings.TrimSpace(srv.AdminEmail) != "" {
		var cnt int
		if err := s.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM staff WHERE server_id = ? AND LOWER(email) = LOWER(?)`,
			srv.ID, strings.TrimSpace(srv.AdminEmail),
		).Scan(&cnt); err != nil {
			return nil, err
		}
		if cnt == 0 {
			totalStaff++
		}
	}

	overview := &modlv1.OverviewResponse{
		TotalTickets:  totalTickets,
		TotalPlayers:  totalPlayers,
		TotalStaff:    totalStaff,
		ActiveTickets: activeTickets,
		TicketChange:  calcPercentChange(recentTickets, previousTickets),
		PlayerChange:  calcPercentChange(recentPlayers, previousPlayers),
	}

	return &modlv1.AnalyticsOverviewResponse{
		Overview: map[string]*modlv1.OverviewResponse{
			"7d":  overview,
			"30d": overview,
			"90d": overview,
			"all": overview,
		},
	}, nil
}

func (s *Service) GetTicketAnalytics(ctx context.Context, srv *models.Server, period string) (*modlv1.TicketAnalyticsResponse, error) {
	nowMs := time.Now().UnixMilli()
	startMs, hasStart := getStartDateMs(period, nowMs)
	loc := s.serverLocation(ctx, srv.ID)

	rows, err := s.db.QueryContext(ctx,
		`SELECT status, type, created_at, updated_at FROM tickets WHERE server_id = ? AND LOWER(status) != 'unfinished'`,
		srv.ID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	statusCounts := make(map[string]int32)
	categoryCounts := make(map[string]int32)
	dailyCounts := make(map[string]int32)
	resSumMs := make(map[string]int64)
	resCount := make(map[string]int64)

	for rows.Next() {
		var status, ticketType string
		var createdAt, updatedAt int64
		if err := rows.Scan(&status, &ticketType, &createdAt, &updatedAt); err != nil {
			return nil, err
		}

		if !hasStart || createdAt >= startMs {
			st := status
			if st == "" {
				st = "unknown"
			}
			statusCounts[st]++

			cat := ticketType
			if cat == "" {
				cat = "unknown"
			}
			categoryCounts[cat]++

			if createdAt > 0 {
				day := time.UnixMilli(createdAt).In(loc).Format("2006-01-02")
				dailyCounts[day]++
			}
		}

		if strings.EqualFold(status, "closed") && (!hasStart || updatedAt >= startMs) && createdAt > 0 && updatedAt >= createdAt {
			cat := ticketType
			if cat == "" {
				cat = "unknown"
			}
			resSumMs[cat] += updatedAt - createdAt
			resCount[cat]++
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	byStatus := make([]*modlv1.TicketAnalyticsResponse_StatusCount, 0, len(statusCounts))
	for k, v := range statusCounts {
		byStatus = append(byStatus, &modlv1.TicketAnalyticsResponse_StatusCount{Status: k, Count: v})
	}
	sort.Slice(byStatus, func(i, j int) bool { return byStatus[i].Status < byStatus[j].Status })

	byCategory := make([]*modlv1.TicketAnalyticsResponse_CategoryCount, 0, len(categoryCounts))
	for k, v := range categoryCounts {
		byCategory = append(byCategory, &modlv1.TicketAnalyticsResponse_CategoryCount{Category: k, Count: v})
	}
	sort.Slice(byCategory, func(i, j int) bool { return byCategory[i].Category < byCategory[j].Category })

	avgRes := make([]*modlv1.TicketAnalyticsResponse_CategoryResolutionTime, 0, len(resCount))
	for cat, cnt := range resCount {
		avgHours := float64(resSumMs[cat]) / float64(cnt) / (1000.0 * 60.0 * 60.0)
		avgHours = math.Round(avgHours*10) / 10
		avgRes = append(avgRes, &modlv1.TicketAnalyticsResponse_CategoryResolutionTime{
			Category: cat,
			AvgHours: avgHours,
		})
	}
	sort.Slice(avgRes, func(i, j int) bool { return avgRes[i].Category < avgRes[j].Category })

	days := make([]string, 0, len(dailyCounts))
	for d := range dailyCounts {
		days = append(days, d)
	}
	sort.Strings(days)
	dailyTickets := make([]*modlv1.TicketAnalyticsResponse_DailyTicket, 0, len(days))
	for _, d := range days {
		dailyTickets = append(dailyTickets, &modlv1.TicketAnalyticsResponse_DailyTicket{
			Date:  d,
			Count: dailyCounts[d],
		})
	}

	return &modlv1.TicketAnalyticsResponse{
		ByStatus:                byStatus,
		ByCategory:              byCategory,
		AvgResolutionByCategory: avgRes,
		DailyTickets:            dailyTickets,
	}, nil
}

var defaultPunishmentTypeNames = map[int]string{
	0:  "Kick",
	1:  "Manual Mute",
	2:  "Manual Ban",
	3:  "Security Ban",
	4:  "Linked Ban",
	5:  "Blacklist",
	6:  "Chat Abuse",
	7:  "Anti Social",
	8:  "Targeting",
	9:  "Bad Content",
	10: "Bad Username",
	11: "Bad Skin",
	12: "Team Abuse",
	13: "Game Abuse",
	14: "Cheating",
	15: "Game Trading",
	16: "Account Abuse",
	17: "Systems Abuse",
}

func (s *Service) loadPunishmentTypeMaps(ctx context.Context, serverID string) (map[int]string, map[string]int) {
	byOrd := make(map[int]string, len(defaultPunishmentTypeNames))
	byName := make(map[string]int, len(defaultPunishmentTypeNames))
	for ord, name := range defaultPunishmentTypeNames {
		byOrd[ord] = name
		byName[strings.ToLower(name)] = ord
	}

	var dataJSON string
	err := s.db.QueryRowContext(ctx,
		`SELECT data_json FROM settings WHERE server_id = ? AND type = 'punishmentTypes' LIMIT 1`,
		serverID,
	).Scan(&dataJSON)
	if err != nil || strings.TrimSpace(dataJSON) == "" {
		return byOrd, byName
	}

	var parsed []struct {
		ID      *int   `json:"id"`
		Ordinal *int   `json:"ordinal"`
		Name    string `json:"name"`
	}
	if err := json.Unmarshal([]byte(dataJSON), &parsed); err != nil || len(parsed) == 0 {
		return byOrd, byName
	}

	byOrd = make(map[int]string, len(parsed))
	byName = make(map[string]int, len(parsed))
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
		byOrd[ord] = pt.Name
		if pt.Name != "" {
			byName[strings.ToLower(pt.Name)] = ord
		}
	}
	return byOrd, byName
}

func (s *Service) GetPunishmentAnalytics(ctx context.Context, srv *models.Server, period string) (*modlv1.PunishmentAnalyticsResponse, error) {
	nowMs := time.Now().UnixMilli()
	startMs, hasStart := getStartDateMs(period, nowMs)
	loc := s.serverLocation(ctx, srv.ID)
	typeByOrd, ordByName := s.loadPunishmentTypeMaps(ctx, srv.ID)

	staffMap := make(map[string]string)
	sRows, err := s.db.QueryContext(ctx, `SELECT id, username, assigned_minecraft_uuid FROM staff WHERE server_id = ?`, srv.ID)
	if err != nil {
		return nil, err
	}
	for sRows.Next() {
		var id, username, mcUUID string
		if err := sRows.Scan(&id, &username, &mcUUID); err != nil {
			sRows.Close()
			return nil, err
		}
		if id != "" {
			staffMap[id] = username
		}
		if mcUUID != "" {
			staffMap[strings.ToLower(mcUUID)] = username
		}
	}
	sRows.Close()

	query := `SELECT type, punishment_type_id, issuer_uuid, issuer_name, issued_at, data_json FROM punishments WHERE server_id = ?`
	args := []any{srv.ID}
	if hasStart {
		query += ` AND issued_at >= ?`
		args = append(args, startMs)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byTypeCounts := make(map[string]int32)
	byStaffCounts := make(map[string]int32)
	dailyCounts := make(map[string]int32)

	for rows.Next() {
		var typeRaw, typeID, issuerUUID, issuerName, dataJSON string
		var issuedAt int64
		if err := rows.Scan(&typeRaw, &typeID, &issuerUUID, &issuerName, &issuedAt, &dataJSON); err != nil {
			return nil, err
		}
		var data struct {
			TypeOrdinal *int `json:"typeOrdinal"`
		}
		if strings.TrimSpace(dataJSON) != "" && dataJSON != "{}" {
			_ = json.Unmarshal([]byte(dataJSON), &data)
		}

		ord := -1
		if data.TypeOrdinal != nil {
			ord = *data.TypeOrdinal
		} else if n, err := strconv.Atoi(strings.TrimSpace(typeID)); err == nil {
			ord = n
		} else if n, err := strconv.Atoi(strings.TrimSpace(typeRaw)); err == nil {
			ord = n
		} else if mapped, ok := ordByName[strings.ToLower(strings.TrimSpace(typeRaw))]; ok {
			ord = mapped
		}

		typeName := "Unknown"
		if name, ok := typeByOrd[ord]; ok && name != "" {
			typeName = name
		} else if typeRaw != "" {
			if _, err := strconv.Atoi(typeRaw); err != nil {
				typeName = typeRaw
			}
		}
		byTypeCounts[typeName]++

		issuerDisplay := strings.TrimSpace(issuerName)
		if issuerUUID != "" {
			if resolved, ok := staffMap[issuerUUID]; ok && resolved != "" {
				issuerDisplay = resolved
			} else if resolved, ok := staffMap[strings.ToLower(issuerUUID)]; ok && resolved != "" {
				issuerDisplay = resolved
			} else if issuerDisplay == "" {
				issuerDisplay = issuerUUID
			}
		}
		if issuerDisplay == "" {
			issuerDisplay = "Unknown"
		}
		byStaffCounts[issuerDisplay]++

		if issuedAt > 0 {
			day := time.UnixMilli(issuedAt).In(loc).Format("2006-01-02")
			dailyCounts[day]++
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	byType := make([]*modlv1.PunishmentAnalyticsResponse_TypeCount, 0, len(byTypeCounts))
	for k, v := range byTypeCounts {
		byType = append(byType, &modlv1.PunishmentAnalyticsResponse_TypeCount{Type: k, Count: v})
	}
	sort.Slice(byType, func(i, j int) bool {
		if byType[i].Count != byType[j].Count {
			return byType[i].Count > byType[j].Count
		}
		return byType[i].Type < byType[j].Type
	})

	byStaff := make([]*modlv1.PunishmentAnalyticsResponse_StaffPunishment, 0, len(byStaffCounts))
	for k, v := range byStaffCounts {
		byStaff = append(byStaff, &modlv1.PunishmentAnalyticsResponse_StaffPunishment{Username: k, Count: v})
	}
	sort.Slice(byStaff, func(i, j int) bool {
		if byStaff[i].Count != byStaff[j].Count {
			return byStaff[i].Count > byStaff[j].Count
		}
		return byStaff[i].Username < byStaff[j].Username
	})
	if len(byStaff) > 10 {
		byStaff = byStaff[:10]
	}

	days := make([]string, 0, len(dailyCounts))
	for d := range dailyCounts {
		days = append(days, d)
	}
	sort.Strings(days)
	daily := make([]*modlv1.PunishmentAnalyticsResponse_DailyPunishment, 0, len(days))
	for _, d := range days {
		daily = append(daily, &modlv1.PunishmentAnalyticsResponse_DailyPunishment{Date: d, Count: dailyCounts[d]})
	}

	return &modlv1.PunishmentAnalyticsResponse{
		ByType:           byType,
		BySeverity:       []*modlv1.PunishmentAnalyticsResponse_SeverityCount{},
		DailyPunishments: daily,
		ByStaff:          byStaff,
	}, nil
}

func (s *Service) GetAuditLogAnalytics(ctx context.Context, srv *models.Server, period string) (*modlv1.AuditLogsAnalyticsResponse, error) {
	nowMs := time.Now().UnixMilli()
	startMs, hasStart := getStartDateMs(period, nowMs)
	loc := s.serverLocation(ctx, srv.ID)

	levelQuery := `SELECT level, COUNT(*) FROM audit_logs WHERE server_id = ?`
	levelArgs := []any{srv.ID}
	if hasStart {
		levelQuery += ` AND created_at >= ?`
		levelArgs = append(levelArgs, startMs)
	}
	levelQuery += ` GROUP BY level`

	rows, err := s.db.QueryContext(ctx, levelQuery, levelArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byLevel := make([]*modlv1.AuditLogsAnalyticsResponse_LevelCount, 0)
	for rows.Next() {
		var lvl string
		var cnt int32
		if err := rows.Scan(&lvl, &cnt); err != nil {
			return nil, err
		}
		if lvl == "" {
			lvl = "unknown"
		}
		byLevel = append(byLevel, &modlv1.AuditLogsAnalyticsResponse_LevelCount{Level: lvl, Count: cnt})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	since24h := nowMs - dayMs
	hRows, err := s.db.QueryContext(ctx,
		`SELECT created_at FROM audit_logs WHERE server_id = ? AND created_at >= ?`,
		srv.ID, since24h,
	)
	if err != nil {
		return nil, err
	}
	defer hRows.Close()

	hourlyCounts := make(map[string]int32)
	for hRows.Next() {
		var ts int64
		if err := hRows.Scan(&ts); err != nil {
			return nil, err
		}
		hourKey := time.UnixMilli(ts).In(loc).Format("2006-01-02T15")
		hourlyCounts[hourKey]++
	}
	if err := hRows.Err(); err != nil {
		return nil, err
	}

	hours := make([]string, 0, len(hourlyCounts))
	for h := range hourlyCounts {
		hours = append(hours, h)
	}
	sort.Strings(hours)
	hourlyTrend := make([]*modlv1.AuditLogsAnalyticsResponse_HourlyCount, 0, len(hours))
	for _, h := range hours {
		hourlyTrend = append(hourlyTrend, &modlv1.AuditLogsAnalyticsResponse_HourlyCount{
			Hour:  h,
			Count: hourlyCounts[h],
		})
	}

	return &modlv1.AuditLogsAnalyticsResponse{
		ByLevel:     byLevel,
		HourlyTrend: hourlyTrend,
	}, nil
}

func (s *Service) GetPlayerActivity(ctx context.Context, srv *models.Server, period string) (*modlv1.PlayerActivityResponse, error) {
	nowMs := time.Now().UnixMilli()
	startMs, hasStart := getStartDateMs(period, nowMs)
	loc := s.serverLocation(ctx, srv.ID)

	rows, err := s.db.QueryContext(ctx,
		`SELECT first_seen, ip_addresses_json FROM players WHERE server_id = ?`, srv.ID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	newPlayersByDay := make(map[string]int32)
	countryCounts := make(map[string]int32)
	var proxyCount, hostingCount int32

	for rows.Next() {
		var firstSeen int64
		var ipJSON string
		if err := rows.Scan(&firstSeen, &ipJSON); err != nil {
			return nil, err
		}

		if earliest, ok := earliestPlayerLoginMs(firstSeen, ipJSON); ok {
			if !hasStart || earliest > startMs {
				day := time.UnixMilli(earliest).In(loc).Format("2006-01-02")
				newPlayersByDay[day]++
			}
		}

		var ips []ipEntry
		if err := json.Unmarshal([]byte(ipJSON), &ips); err != nil {
			continue
		}
		for _, ip := range ips {
			firstLoginMs, _ := parseDateValue(ip.FirstLogin)
			var loginsAfterStart int32
			hasRecentLogin := !hasStart || firstLoginMs > startMs
			for _, rawLogin := range ip.Logins {
				ms, ok := parseDateValue(rawLogin)
				if !ok {
					continue
				}
				if !hasStart || ms > startMs {
					loginsAfterStart++
					hasRecentLogin = true
				}
			}
			if !hasRecentLogin {
				continue
			}

			if ip.Proxy {
				proxyCount++
			}
			if ip.Hosting {
				hostingCount++
			}
			if country := strings.TrimSpace(ip.Country); country != "" {
				inc := loginsAfterStart
				if inc < 1 {
					inc = 1
				}
				countryCounts[country] += inc
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	days := make([]string, 0, len(newPlayersByDay))
	for d := range newPlayersByDay {
		days = append(days, d)
	}
	sort.Strings(days)
	newPlayersTrend := make([]*modlv1.PlayerActivityResponse_DailyCount, 0, len(days))
	for _, d := range days {
		newPlayersTrend = append(newPlayersTrend, &modlv1.PlayerActivityResponse_DailyCount{
			Date:  d,
			Count: newPlayersByDay[d],
		})
	}

	byCountry := make([]*modlv1.PlayerActivityResponse_CountryCount, 0, len(countryCounts))
	for country, cnt := range countryCounts {
		byCountry = append(byCountry, &modlv1.PlayerActivityResponse_CountryCount{
			Country: country,
			Count:   cnt,
		})
	}
	sort.Slice(byCountry, func(i, j int) bool {
		if byCountry[i].Count != byCountry[j].Count {
			return byCountry[i].Count > byCountry[j].Count
		}
		return byCountry[i].Country < byCountry[j].Country
	})
	if len(byCountry) > 20 {
		byCountry = byCountry[:20]
	}

	return &modlv1.PlayerActivityResponse{
		NewPlayersTrend: newPlayersTrend,
		LoginsByCountry: byCountry,
		SuspiciousActivity: &modlv1.PlayerActivityResponse_SuspiciousActivity{
			ProxyCount:   proxyCount,
			HostingCount: hostingCount,
		},
	}, nil
}
