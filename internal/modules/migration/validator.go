package migration

import (
	"regexp"
	"strings"
)

const (
	MaxPlayerRecords   = 1_000_000
	MaxNotesPerPlayer  = 500
	MaxIPsPerPlayer    = 500
	MaxPunishments     = 1_000
	MaxUsernames       = 500
	MaxTextLength      = 10_000
	MaxIPLength        = 45
	MaxTypeLength      = 64
	MaxUsernameLength  = 64
)

var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func normalizeUUID(raw string) string {
	trimmed := strings.ToLower(strings.TrimSpace(raw))
	if len(trimmed) == 32 && !strings.Contains(trimmed, "-") {
		return trimmed[0:8] + "-" + trimmed[8:12] + "-" + trimmed[12:16] + "-" + trimmed[16:20] + "-" + trimmed[20:32]
	}
	return trimmed
}

func isValidUUID(u string) bool {
	return u != "" && uuidRegex.MatchString(u)
}

func truncate(s string, maxLen int) string {
	if maxLen <= 0 || len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func validateHeader(hasPlayers, isPlayersArray bool, declaredTotal *int32) string {
	if !hasPlayers || !isPlayersArray {
		return "Migration data must contain a 'players' array"
	}
	if declaredTotal != nil && *declaredTotal < 0 {
		return "Invalid metadata playerCount"
	}
	return ""
}
