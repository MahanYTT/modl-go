package ticket

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	modlv1 "modl-go/gen/modl/v1"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var SystemDataAllowlist = map[string]bool{
	"subject":  true,
	"status":   true,
	"type":     true,
	"category": true,
	"priority": true,
}

func SanitizeForStruct(v any) any {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, item := range val {
			out[k] = SanitizeForStruct(item)
		}
		return out
	case map[string]string:
		out := make(map[string]any, len(val))
		for k, item := range val {
			out[k] = item
		}
		return out
	case []any:
		out := make([]any, len(val))
		for i, item := range val {
			out[i] = SanitizeForStruct(item)
		}
		return out
	case []string:
		out := make([]any, len(val))
		for i, item := range val {
			out[i] = item
		}
		return out
	case []map[string]any:
		out := make([]any, len(val))
		for i, item := range val {
			out[i] = SanitizeForStruct(item)
		}
		return out
	case int:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case time.Time:
		return val.UTC().Format(time.RFC3339)
	case string, float64, float32, bool:
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}

func MapToStruct(m map[string]any) *structpb.Struct {
	if len(m) == 0 {
		s, _ := structpb.NewStruct(map[string]any{})
		return s
	}
	clean, _ := SanitizeForStruct(m).(map[string]any)
	s, err := structpb.NewStruct(clean)
	if err != nil {
		empty, _ := structpb.NewStruct(map[string]any{})
		return empty
	}
	return s
}

func RetainAllowed(data map[string]any, allowlist map[string]bool) map[string]any {
	if len(data) == 0 || len(allowlist) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any)
	for k, v := range data {
		if allowlist[k] {
			out[k] = v
		}
	}
	return out
}

func ToPublicStruct(data map[string]any, allowlist map[string]bool) *structpb.Struct {
	return MapToStruct(RetainAllowed(data, allowlist))
}

func StructToMap(s *structpb.Struct) map[string]any {
	if s == nil {
		return nil
	}
	return s.AsMap()
}

func ValuesToAttachmentMaps(vals []*structpb.Value) []map[string]any {
	if len(vals) == 0 {
		return nil
	}
	out := make([]map[string]any, 0, len(vals))
	for _, v := range vals {
		if v == nil {
			continue
		}
		raw := v.AsInterface()
		if m, ok := raw.(map[string]any); ok {
			out = append(out, m)
		} else {
			out = append(out, map[string]any{"value": raw})
		}
	}
	return out
}

func StructsToAttachmentMaps(structs []*structpb.Struct) []map[string]any {
	if len(structs) == 0 {
		return nil
	}
	out := make([]map[string]any, 0, len(structs))
	for _, s := range structs {
		if s == nil {
			continue
		}
		out = append(out, s.AsMap())
	}
	return out
}

func ResolveFieldLabel(key string, fieldLabels map[string]string) string {
	if fieldLabels != nil {
		if lbl, ok := fieldLabels[key]; ok && lbl != "" {
			return lbl
		}
	}
	return FormatFieldKey(key)
}

func FormatFieldKey(key string) string {
	if strings.TrimSpace(key) == "" {
		return key
	}
	spaced := strings.ReplaceAll(key, "_", " ")
	var buf strings.Builder
	runes := []rune(spaced)
	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) && !unicode.IsSpace(runes[i-1]) {
			buf.WriteRune(' ')
		}
		buf.WriteRune(r)
	}
	words := strings.Fields(buf.String())
	for i, w := range words {
		wr := []rune(w)
		if len(wr) > 0 {
			words[i] = strings.ToUpper(string(wr[0])) + strings.ToLower(string(wr[1:]))
		}
	}
	return strings.Join(words, " ")
}

func MsToTimestamp(ms int64) *timestamppb.Timestamp {
	if ms <= 0 {
		return nil
	}
	return timestamppb.New(time.UnixMilli(ms).UTC())
}

func MsToTime(ms int64) time.Time {
	if ms <= 0 {
		return time.Time{}
	}
	return time.UnixMilli(ms).UTC()
}

func ProcessRepliesWithStaffAvatars(ctx context.Context, store *Store, serverID string, t *Ticket) []TicketReply {
	if len(t.Replies) == 0 {
		return t.Replies
	}
	creatorName := t.CreatorName
	if strings.TrimSpace(creatorName) == "" {
		creatorName = "Player"
	}
	var staffUsernames []string
	seen := make(map[string]struct{})
	for _, r := range t.Replies {
		if r.Staff && strings.TrimSpace(r.Avatar) == "" && strings.TrimSpace(r.Name) != "" {
			if _, ok := seen[r.Name]; !ok {
				seen[r.Name] = struct{}{}
				staffUsernames = append(staffUsernames, r.Name)
			}
		}
	}
	var avatarMap map[string]string
	if len(staffUsernames) > 0 && store != nil {
		uuids := store.FindStaffMinecraftUUIDsByUsernames(ctx, serverID, staffUsernames)
		if len(uuids) > 0 {
			avatarMap = make(map[string]string, len(uuids))
			for u, uuid := range uuids {
				avatarMap[u] = fmt.Sprintf("https://mc-heads.net/avatar/%s/32", uuid)
			}
		}
	}

	out := make([]TicketReply, len(t.Replies))
	for i, r := range t.Replies {
		name := r.Name
		if strings.TrimSpace(name) == "" {
			if r.Staff {
				name = "Staff"
			} else {
				name = creatorName
			}
		}
		typ := r.Type
		if strings.TrimSpace(typ) == "" {
			if r.Staff {
				typ = "staff"
			} else {
				typ = "user"
			}
		}
		avatar := r.Avatar
		if r.Staff && strings.TrimSpace(avatar) == "" && avatarMap != nil {
			if a, ok := avatarMap[name]; ok {
				avatar = a
			}
		}
		r.Name = name
		r.Type = typ
		r.Avatar = avatar
		out[i] = r
	}
	return out
}

func ToProtoTicketReply(r TicketReply) *modlv1.TicketReply {
	pb := &modlv1.TicketReply{
		Id:                r.ID,
		Name:              r.Name,
		Avatar:            r.Avatar,
		Content:           r.Content,
		Type:              r.Type,
		Created:           r.CreatedMs,
		Staff:             r.Staff,
		Action:            r.Action,
		CreatorIdentifier: r.CreatorIdentifier,
	}
	for _, att := range r.Attachments {
		pb.Attachments = append(pb.Attachments, MapToStruct(att))
	}
	return pb
}

func ToProtoPublicTicketReply(r TicketReply) *modlv1.PublicTicketReply {
	pb := &modlv1.PublicTicketReply{
		Id:                r.ID,
		Name:              r.Name,
		Avatar:            r.Avatar,
		Content:           r.Content,
		Type:              r.Type,
		Created:           MsToTimestamp(r.CreatedMs),
		Staff:             r.Staff,
		Action:            r.Action,
		CreatorIdentifier: r.CreatorIdentifier,
	}
	for _, att := range r.Attachments {
		pb.Attachments = append(pb.Attachments, MapToStruct(att))
	}
	return pb
}

func ToProtoTicketNote(n TicketNote) *modlv1.TicketNote {
	return &modlv1.TicketNote{
		Text:         n.Text,
		IssuerName:   n.IssuerName,
		IssuerAvatar: n.Avatar,
		Date:         n.DateMs,
	}
}

func ToProtoTicketResponse(ctx context.Context, store *Store, serverID string, t *Ticket) *modlv1.TicketResponse {
	replies := ProcessRepliesWithStaffAvatars(ctx, store, serverID, t)
	creatorName := t.CreatorName
	if strings.TrimSpace(creatorName) == "" {
		creatorName = "Unknown"
	}
	subject := t.Subject
	if strings.TrimSpace(subject) == "" {
		subject = "No Subject"
	}
	catType := t.Type
	if catType == "" {
		catType = CategorySupport
	}
	status := t.Status
	if status == "" {
		status = StatusOpen
	}

	pb := &modlv1.TicketResponse{
		Id:                   t.ID,
		Type:                 string(catType),
		Category:             catType.DisplayName(),
		Subject:              subject,
		Status:               string(status),
		AppealWorkflowStatus: t.AppealWorkflowStatus,
		CreatorName:          creatorName,
		CreatorUuid:          t.CreatorUUID,
		ReportedBy:           creatorName,
		ReportedPlayer:       t.ReportedPlayer,
		ReportedPlayerUuid:   t.ReportedPlayerUUID,
		Date:                 t.CreatedAtMs,
		Locked:               t.Locked,
		Tags:                 t.Tags,
		AssignedTo:           t.AssignedTo,
		EmailAuthEnabled:     t.EmailAuthEnabled,
		Hidden:               t.Hidden,
	}

	for _, r := range replies {
		pb.Messages = append(pb.Messages, ToProtoTicketReply(r))
	}
	for _, n := range t.Notes {
		pb.Notes = append(pb.Notes, ToProtoTicketNote(n))
	}
	for _, cm := range t.ChatMessages {
		pb.ChatMessages = append(pb.ChatMessages, &modlv1.TicketChatMessage{
			Content:   cm.Content,
			Timestamp: cm.TimestampMs,
			Sender:    cm.Sender,
		})
	}
	if t.FormData != nil {
		pb.FormData = MapToStruct(t.FormData)
	}
	if t.Data != nil {
		pb.Data = MapToStruct(t.Data)
	}
	if t.AIAnalysis != nil {
		aiPB := &modlv1.AIAnalysisResult{
			Analysis:                t.AIAnalysis.Analysis,
			CreatedAt:               t.AIAnalysis.AnalyzedAtMs,
			RawResponse:             t.AIAnalysis.RawResponse,
			WasAppliedAutomatically: t.AIAnalysis.WasAppliedAutomatically,
			Dismissed:               t.AIAnalysis.Dismissed,
		}
		if t.AIAnalysis.SuggestedAction != nil {
			aiPB.SuggestedAction = &modlv1.AIAnalysisResult_SuggestedAction{
				PunishmentTypeId: t.AIAnalysis.SuggestedAction.PunishmentTypeID,
				Severity:         t.AIAnalysis.SuggestedAction.Severity,
			}
		}
		pb.AiAnalysis = aiPB
	}
	return pb
}

func ToProtoTicketListItem(t *Ticket) *modlv1.TicketListItemResponse {
	creatorName := t.CreatorName
	if strings.TrimSpace(creatorName) == "" {
		creatorName = "Unknown"
	}
	subject := t.Subject
	if strings.TrimSpace(subject) == "" {
		subject = "No Subject"
	}
	catType := t.Type
	if catType == "" {
		catType = CategorySupport
	}
	status := t.Status
	if status == "" {
		status = StatusOpen
	}

	item := &modlv1.TicketListItemResponse{
		Id:             t.ID,
		Subject:        subject,
		Status:         string(status),
		ReportedBy:     t.CreatorUUID,
		ReportedByName: creatorName,
		Date:           t.CreatedAtMs,
		Category:       catType.DisplayName(),
		Locked:         t.Locked,
		Type:           string(catType),
		ReplyCount:     int32(len(t.Replies)),
		Tags:           t.Tags,
		AssignedTo:     t.AssignedTo,
		Hidden:         t.Hidden,
	}
	if len(t.Replies) > 0 {
		item.LastReply = ToProtoTicketReply(t.Replies[len(t.Replies)-1])
	}
	return item
}

func ToPlayerTicketResponse(t *Ticket) PlayerTicketResponse {
	creatorName := t.CreatorName
	if strings.TrimSpace(creatorName) == "" {
		creatorName = "Unknown"
	}
	subject := t.Subject
	if strings.TrimSpace(subject) == "" {
		subject = "No Subject"
	}
	catType := t.Type
	if catType == "" {
		catType = CategorySupport
	}
	status := t.Status
	if status == "" {
		status = StatusOpen
	}
	tags := t.Tags
	if tags == nil {
		tags = []string{}
	}
	assigned := t.AssignedTo
	if assigned == nil {
		assigned = []string{}
	}
	return PlayerTicketResponse{
		ID:                 t.ID,
		Subject:            subject,
		Status:             string(status),
		Type:               string(catType),
		Created:            MsToTime(t.CreatedAtMs),
		CreatorName:        creatorName,
		CreatorUUID:        t.CreatorUUID,
		ReportedPlayer:     t.ReportedPlayer,
		ReportedPlayerUUID: t.ReportedPlayerUUID,
		Locked:             t.Locked,
		Tags:               tags,
		AssignedTo:         assigned,
	}
}

func ToPublicTicketResponse(ctx context.Context, store *Store, serverID string, t *Ticket, formFieldAllowlist map[string]bool) *modlv1.PublicTicketResponse {
	replies := ProcessRepliesWithStaffAvatars(ctx, store, serverID, t)
	creatorName := t.CreatorName
	if strings.TrimSpace(creatorName) == "" {
		creatorName = "Unknown"
	}
	subject := t.Subject
	if strings.TrimSpace(subject) == "" {
		subject = "No Subject"
	}
	catType := t.Type
	if catType == "" {
		catType = CategorySupport
	}
	status := t.Status
	if status == "" {
		status = StatusOpen
	}

	pb := &modlv1.PublicTicketResponse{
		Id:               t.ID,
		IdAlias:          t.ID,
		Type:             string(catType),
		Subject:          subject,
		Status:           string(status),
		CreatorName:      creatorName,
		Creator:          creatorName,
		CreatorUuid:      "",
		ReportedBy:       creatorName,
		Category:         catType.DisplayName(),
		Locked:           t.Locked,
		ReportedPlayer:   t.ReportedPlayer,
		EmailAuthEnabled: t.EmailAuthEnabled,
		Tags:             t.Tags,
		Data:             ToPublicStruct(t.Data, SystemDataAllowlist),
		FormData:         ToPublicStruct(t.FormData, formFieldAllowlist),
	}
	if t.CreatedAtMs > 0 {
		ts := MsToTimestamp(t.CreatedAtMs)
		pb.Created = ts
		pb.Date = ts
	}
	for _, r := range replies {
		if strings.EqualFold(r.Type, "staff-note") {
			continue
		}
		r.CreatorIdentifier = ""
		pub := ToProtoPublicTicketReply(r)
		pb.Replies = append(pb.Replies, pub)
		pb.Messages = append(pb.Messages, pub)
	}
	if t.EmailAuthEnabled {
		for _, cm := range t.ChatMessages {
			pb.ChatMessages = append(pb.ChatMessages, &modlv1.PublicTicketResponse_PublicTicketChatMessage{
				Content:   cm.Content,
				Timestamp: MsToTimestamp(cm.TimestampMs),
				Sender:    cm.Sender,
			})
		}
	}
	return pb
}
