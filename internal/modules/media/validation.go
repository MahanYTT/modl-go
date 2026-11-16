package media

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

const (
	DefaultMaxFileSize          int64 = 10 * 1024 * 1024
	StorageBulkDeleteMaxKeys          = 1000
	EvidenceMaxItems                  = 10
	DefaultFreeStorageBytes     int64 = 1024 * 1024 * 1024
	DefaultTempUploadCapBytes   int64 = 512 * 1024 * 1024
)

var dangerousExtensions = []string{
	".exe", ".bat", ".cmd", ".com", ".msi", ".scr", ".pif",
	".js", ".vbs", ".wsf", ".jar", ".sh", ".ps1", ".php",
	".asp", ".aspx", ".jsp", ".cgi", ".pl", ".py", ".rb",
}

var allowedTypes = map[string]map[string]bool{
	"evidence": {
		"image/png": true, "image/jpeg": true, "image/gif": true, "image/apng": true, "image/webp": true,
		"video/mp4": true, "video/webm": true, "video/quicktime": true, "video/x-matroska": true,
		"application/pdf": true, "text/plain": true, "text/markdown": true,
	},
	"ticket": {
		"image/png": true, "image/jpeg": true, "image/gif": true, "image/apng": true, "image/webp": true,
		"video/mp4": true, "video/webm": true, "video/quicktime": true, "video/x-matroska": true,
		"application/pdf": true, "text/plain": true, "text/markdown": true,
	},
	"appeal": {
		"image/png": true, "image/jpeg": true, "image/gif": true, "image/apng": true, "image/webp": true,
		"video/mp4": true, "video/webm": true, "video/quicktime": true, "video/x-matroska": true,
		"application/pdf": true, "text/plain": true, "text/markdown": true,
	},
	"article": {
		"image/png": true, "image/jpeg": true, "image/gif": true, "image/webp": true, "image/apng": true,
		"text/plain": true, "text/markdown": true,
	},
	"server-icon": {
		"image/png": true, "image/jpeg": true, "image/webp": true, "image/gif": true, "image/apng": true,
	},
	"replay": {
		"application/octet-stream": true,
	},
}

var supportedTypesOrder = map[string][]string{
	"evidence":     {"image/png", "image/jpeg", "image/gif", "image/apng", "image/webp", "video/mp4", "video/webm", "video/quicktime", "video/x-matroska", "application/pdf", "text/plain", "text/markdown"},
	"tickets":      {"image/png", "image/jpeg", "image/gif", "image/apng", "image/webp", "video/mp4", "video/webm", "video/quicktime", "video/x-matroska", "application/pdf", "text/plain", "text/markdown"},
	"appeals":      {"image/png", "image/jpeg", "image/gif", "image/apng", "image/webp", "video/mp4", "video/webm", "video/quicktime", "video/x-matroska", "application/pdf", "text/plain", "text/markdown"},
	"articles":     {"image/png", "image/jpeg", "image/gif", "image/webp", "image/apng", "text/plain", "text/markdown"},
	"server-icons": {"image/png", "image/jpeg", "image/webp", "image/gif", "image/apng"},
	"replays":      {"application/octet-stream"},
}

var maxSizes = map[string]int64{
	"evidence":    100 * 1024 * 1024,
	"ticket":      100 * 1024 * 1024,
	"appeal":      100 * 1024 * 1024,
	"article":     50 * 1024 * 1024,
	"server-icon": 10 * 1024 * 1024,
	"replay":      10 * 1024 * 1024,
}

var premiumMaxSizes = map[string]int64{
	"evidence":    1024 * 1024 * 1024,
	"ticket":      1024 * 1024 * 1024,
	"appeal":      1024 * 1024 * 1024,
	"article":     50 * 1024 * 1024,
	"server-icon": 10 * 1024 * 1024,
	"replay":      10 * 1024 * 1024,
}

var sanitizeRegex = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

func ValidateMetadata(fileName, contentType string, fileSize int64, uploadType string, isPremium bool) (bool, string) {
	if strings.TrimSpace(fileName) == "" {
		return false, "File name is required"
	}
	if strings.TrimSpace(contentType) == "" {
		return false, "Content type is required"
	}

	lowerName := strings.ToLower(fileName)
	for _, ext := range dangerousExtensions {
		if strings.HasSuffix(lowerName, ext) {
			return false, "File type not allowed"
		}
	}

	normalizedUploadType := NormalizeUploadType(uploadType)
	allowed, ok := allowedTypes[normalizedUploadType]
	if !ok || len(allowed) == 0 {
		return false, "Invalid upload type"
	}

	baseCT := normalizeContentType(contentType)
	if !allowed[baseCT] {
		return false, fmt.Sprintf("File type not allowed for %s", uploadType)
	}

	if fileSize <= 0 {
		return false, "Invalid file size"
	}
	maxSize := MaxSizeForType(normalizedUploadType, isPremium)
	if fileSize > maxSize {
		return false, "File exceeds maximum size of " + FormatBytesCompact(maxSize)
	}
	return true, ""
}

func MaxSizeForType(uploadType string, isPremium bool) int64 {
	sizes := maxSizes
	if isPremium {
		sizes = premiumMaxSizes
	}
	if v, ok := sizes[NormalizeUploadType(uploadType)]; ok {
		return v
	}
	return DefaultMaxFileSize
}

func AllSupportedTypes() map[string][]string {
	out := make(map[string][]string, len(supportedTypesOrder))
	for k, v := range supportedTypesOrder {
		cp := make([]string, len(v))
		copy(cp, v)
		out[k] = cp
	}
	return out
}

func AllSizeLimits(isPremium bool) map[string]int64 {
	sizes := maxSizes
	if isPremium {
		sizes = premiumMaxSizes
	}
	return map[string]int64{
		"evidence":     sizes["evidence"],
		"tickets":      sizes["ticket"],
		"appeals":      sizes["appeal"],
		"articles":     sizes["article"],
		"server-icons": sizes["server-icon"],
		"replays":      sizes["replay"],
	}
}

func IsKeyOwnedByServer(key, serverID string) bool {
	if key == "" || serverID == "" || !strings.HasPrefix(key, serverID+"/") {
		return false
	}
	for _, seg := range strings.Split(key, "/") {
		if seg == ".." {
			return false
		}
	}
	return true
}

func ExtractUploadType(key string) string {
	parts := strings.Split(key, "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

func ExtractEntityID(key string) string {
	parts := strings.Split(key, "/")
	if len(parts) >= 4 {
		return parts[2]
	}
	return ""
}

func BuildKey(serverID, uploadType, fileName, entityID string) string {
	safeUploadType := sanitizeSegment(uploadType, "other")
	safeFileName := sanitizeFileName(fileName)

	if strings.TrimSpace(entityID) != "" {
		safeEntityID := sanitizeSegment(entityID, uuid.NewString())
		folder := safeUploadType
		if safeUploadType == "ticket" {
			folder = "tickets"
		}
		uniqueName := uuid.NewString() + "-" + safeFileName
		return fmt.Sprintf("%s/%s/%s/%s", serverID, folder, safeEntityID, uniqueName)
	}

	id := uuid.NewString()
	ext := filepath.Ext(safeFileName)
	return fmt.Sprintf("%s/%s/%s%s", serverID, safeUploadType, id, ext)
}

func ExtractFileName(key string) string {
	idx := strings.LastIndex(key, "/")
	if idx >= 0 && idx+1 < len(key) {
		return key[idx+1:]
	}
	return key
}

func CategorizeFile(key string) string {
	switch {
	case strings.Contains(key, "/evidence/"):
		return "evidence"
	case strings.Contains(key, "/tickets/"), strings.Contains(key, "/ticket/"):
		return "ticket"
	case strings.Contains(key, "/logs/"):
		return "logs"
	case strings.Contains(key, "/backup/"):
		return "backup"
	case strings.Contains(key, "/replays/"):
		return "replay"
	default:
		return "other"
	}
}

func NormalizeUploadType(uploadType string) string {
	if uploadType == "tickets" {
		return "ticket"
	}
	return uploadType
}

func StripLeadingSlash(key string) string {
	return strings.TrimPrefix(key, "/")
}

func IsAnonymousTempEntity(entityID string) bool {
	return strings.EqualFold(strings.TrimSpace(entityID), "new")
}

func TempUploadPrefixes(serverID string) []string {
	return []string{
		serverID + "/tickets/new/",
		serverID + "/appeal/new/",
	}
}

func FormatBytes(bytes int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case bytes < kb:
		return fmt.Sprintf("%d B", bytes)
	case bytes < mb:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(kb))
	case bytes < gb:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(mb))
	default:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(gb))
	}
}

func FormatBytesCompact(bytes int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case bytes < mb:
		return fmt.Sprintf("%.0f KB", float64(bytes)/float64(kb))
	case bytes < gb:
		return fmt.Sprintf("%.0f MB", float64(bytes)/float64(mb))
	default:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(gb))
	}
}

func normalizeContentType(ct string) string {
	base, _, _ := strings.Cut(ct, ";")
	return strings.ToLower(strings.TrimSpace(base))
}

func sanitizeFileName(fileName string) string {
	if strings.TrimSpace(fileName) == "" {
		return "upload.bin"
	}
	basename := strings.ReplaceAll(fileName, "\\", "/")
	if idx := strings.LastIndex(basename, "/"); idx >= 0 {
		basename = basename[idx+1:]
	}
	return sanitizeAllowed(basename, "upload.bin")
}

func sanitizeSegment(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return sanitizeAllowed(value, fallback)
}

func sanitizeAllowed(value, fallback string) string {
	sanitized := sanitizeRegex.ReplaceAllString(value, "_")
	if strings.TrimSpace(sanitized) == "" {
		return fallback
	}
	if len(sanitized) > 128 {
		return sanitized[:128]
	}
	return sanitized
}
