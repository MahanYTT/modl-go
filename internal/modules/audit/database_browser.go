package audit

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	modlv1 "modl-go/gen/modl/v1"

	"google.golang.org/protobuf/types/known/structpb"
)

var errInvalidTable = errors.New("invalid table name")

var allowedTableMapping = map[string]string{
	"players":                 "players",
	"settings":                "settings",
	"staffs":                  "staff",
	"staff":                   "staff",
	"staffroles":              "staff_roles",
	"staff_roles":             "staff_roles",
	"tickets":                 "tickets",
	"ticket_verifications":    "ticket_verifications",
	"logs":                    "audit_logs",
	"audit_logs":              "audit_logs",
	"server_logs":             "server_logs",
	"punishments":             "punishments",
	"knowledgebasecategories": "knowledgebase_categories",
	"knowledgebase_categories": "knowledgebase_categories",
	"knowledgebasearticles":   "knowledgebase_articles",
	"knowledgebase_articles":  "knowledgebase_articles",
	"homepagecards":           "homepage_cards",
	"homepage_cards":          "homepage_cards",
	"migrations":              "migrations",
	"chat_logs":               "chat_logs",
	"command_logs":            "command_logs",
}

var safeSettingsTypes = map[string]bool{
	"general":          true,
	"punishmentTypes":  true,
	"quickResponses":   true,
	"replayRetention":  true,
	"statusThresholds": true,
	"ticketForms":      true,
	"ticketLabels":     true,
}

var secretFieldPatterns = []string{
	"api_key", "apikey", "webhookurl", "token", "secret", "password",
}

const redactedValue = "[REDACTED]"

func isAllowedTable(name string) (string, bool) {
	sqlTable, ok := allowedTableMapping[strings.TrimSpace(name)]
	return sqlTable, ok
}

func isSecretKey(key string) bool {
	lower := strings.ToLower(key)
	for _, p := range secretFieldPatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

func redactValue(val any) any {
	switch v := val.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, item := range v {
			if isSecretKey(k) {
				out[k] = redactedValue
			} else {
				out[k] = redactValue(item)
			}
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = redactValue(item)
		}
		return out
	default:
		return val
	}
}

func (s *Service) GetDatabaseTable(ctx context.Context, serverID, requestedTable string, limit, skip int) (*modlv1.AuditDatabaseTableResponse, error) {
	sqlTable, ok := isAllowedTable(requestedTable)
	if !ok {
		return nil, errInvalidTable
	}
	if limit < 1 {
		limit = 1
	} else if limit > 100 {
		limit = 100
	}
	if skip < 0 {
		skip = 0
	}

	var total int64
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM `+sqlTable+` WHERE server_id = ?`, serverID,
	).Scan(&total); err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT * FROM `+sqlTable+` WHERE server_id = ? ORDER BY rowid DESC LIMIT ? OFFSET ?`,
		serverID, limit, skip,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	records := make([]*structpb.Struct, 0)
	for rows.Next() {
		rawVals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range rawVals {
			ptrs[i] = &rawVals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}

		rowMap := make(map[string]any, len(cols))
		var settingsType string
		for i, col := range cols {
			val := rawVals[i]
			if b, ok := val.([]byte); ok {
				val = string(b)
			}
			if col == "type" {
				if str, ok := val.(string); ok {
					settingsType = str
				}
			}
			if isSecretKey(col) {
				rowMap[col] = redactedValue
				continue
			}
			if strings.HasSuffix(col, "_json") {
				if str, ok := val.(string); ok && strings.TrimSpace(str) != "" {
					var parsed any
					if err := json.Unmarshal([]byte(str), &parsed); err == nil {
						rowMap[col] = redactValue(parsed)
						continue
					}
				}
			}
			rowMap[col] = val
		}

		if sqlTable == "settings" && !safeSettingsTypes[settingsType] {
			rowMap["data_json"] = redactedValue
		}

		st, err := structpb.NewStruct(rowMap)
		if err != nil {
			return nil, err
		}
		records = append(records, st)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &modlv1.AuditDatabaseTableResponse{
		Data:  records,
		Total: total,
		Limit: int32(limit),
		Skip:  int32(skip),
	}, nil
}
