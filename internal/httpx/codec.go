package httpx

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"buf.build/go/protovalidate"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var (
	protoJSONMarshal = protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   false,
	}
	protoJSONUnmarshal = protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
)

func ReadProto(r *http.Request, msg proto.Message) error {
	body, err := io.ReadAll(io.LimitReader(r.Body, 16<<20))
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	ct := strings.ToLower(r.Header.Get("Content-Type"))
	if strings.Contains(ct, "application/x-protobuf") || strings.Contains(ct, "application/protobuf") || strings.HasPrefix(r.URL.Path, "/v3") && !strings.Contains(ct, "json") {
		if err := proto.Unmarshal(body, msg); err != nil {
			return fmt.Errorf("unmarshal proto: %w", err)
		}
	} else {
		if len(strings.TrimSpace(string(body))) == 0 {
			body = []byte("{}")
		}
		if err := protoJSONUnmarshal.Unmarshal(body, msg); err != nil {
			return fmt.Errorf("unmarshal proto json: %w", err)
		}
	}
	if err := protovalidate.Validate(msg); err != nil {
		return fmt.Errorf("validate proto: %w", err)
	}
	return nil
}

func WriteProto(w http.ResponseWriter, r *http.Request, status int, msg proto.Message) {
	if ShouldWriteProtobuf(r) {
		b, err := proto.Marshal(msg)
		if err != nil {
			WriteError(w, r, http.StatusInternalServerError, "Failed to encode protobuf response")
			return
		}
		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(status)
		_, _ = w.Write(b)
		return
	}
	b, err := protoJSONMarshal.Marshal(msg)
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "Failed to encode JSON response")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(b)
}

func ReadJSON(r *http.Request, dst any) error {
	body, err := io.ReadAll(io.LimitReader(r.Body, 16<<20))
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		body = []byte("{}")
	}
	return json.Unmarshal(body, dst)
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
