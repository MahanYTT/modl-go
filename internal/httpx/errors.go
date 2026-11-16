package httpx

import (
	"encoding/json"
	"net/http"
	"strings"

	modlv1 "modl-go/gen/modl/v1"

	"google.golang.org/protobuf/proto"
)

type ErrorResponse struct {
	Status  int    `json:"status"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

func ShouldWriteProtobuf(r *http.Request) bool {
	p := r.URL.Path
	if strings.HasPrefix(p, "/v3") {
		return true
	}
	if strings.HasPrefix(p, "/v1") || strings.HasPrefix(p, "/v2/minecraft") {
		return false
	}
	accept := strings.ToLower(r.Header.Get("Accept"))
	return strings.Contains(accept, "application/x-protobuf") || strings.Contains(accept, "application/protobuf")
}

func MachineCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "INVALID_ARGUMENT"
	case http.StatusUnauthorized:
		return "UNAUTHENTICATED"
	case http.StatusForbidden:
		return "PERMISSION_DENIED"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusTooManyRequests:
		return "RESOURCE_EXHAUSTED"
	case http.StatusUnsupportedMediaType:
		return "UNSUPPORTED_MEDIA_TYPE"
	case http.StatusMethodNotAllowed:
		return "METHOD_NOT_ALLOWED"
	case http.StatusServiceUnavailable:
		return "UNAVAILABLE"
	default:
		if status >= 500 {
			return "INTERNAL"
		}
		return "INVALID_ARGUMENT"
	}
}

func WriteError(w http.ResponseWriter, r *http.Request, status int, message string) {
	WriteErrorCode(w, r, status, MachineCode(status), message)
}

func WriteErrorCode(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	if ShouldWriteProtobuf(r) {
		msg := &modlv1.ApiError{
			StatusCode: int32(status),
			Code:       code,
			Message:    message,
		}
		b, err := proto.Marshal(msg)
		if err == nil {
			w.Header().Set("Content-Type", "application/x-protobuf")
			w.WriteHeader(status)
			_, _ = w.Write(b)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Status:  status,
		Error:   message,
		Message: message,
	})
}
