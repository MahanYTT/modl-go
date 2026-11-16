package appeal

import (
	"context"
	"strings"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/modules/ticket"
)

func ToProtoAppealTicketResponse(ctx context.Context, ticketStore *ticket.Store, serverID string, t *ticket.Ticket) *modlv1.TicketResponse {
	pb := ticket.ToProtoTicketResponse(ctx, ticketStore, serverID, t)
	pb.Type = string(ticket.CategoryAppeal)
	pb.Category = ticket.CategoryAppeal.DisplayName()
	if pb.AppealWorkflowStatus == "" {
		pb.AppealWorkflowStatus = string(ticket.AppealOpen)
	}
	return pb
}

func ToPublicAppealResponse(t *ticket.Ticket) *modlv1.PublicAppealResponse {
	wf := t.AppealWorkflowStatus
	if wf == "" {
		wf = string(ticket.AppealOpen)
	}
	status := string(t.Status)
	if status == "" {
		status = string(ticket.StatusOpen)
	}
	subject := t.Subject
	if strings.TrimSpace(subject) == "" {
		subject = "No Subject"
	}

	pb := &modlv1.PublicAppealResponse{
		Id:                   t.ID,
		Type:                 string(ticket.CategoryAppeal),
		Subject:              subject,
		Status:               status,
		AppealWorkflowStatus: wf,
		CreatorName:          t.CreatorName,
		CreatorUuid:          "",
		Created:              ticket.MsToTimestamp(t.CreatedAtMs),
		Locked:               t.Locked,
		Tags:                 t.Tags,
		Data:                 ticket.ToPublicStruct(t.Data, ticket.SystemDataAllowlist),
	}
	for _, r := range t.Replies {
		if strings.EqualFold(r.Type, ticket.StaffNoteType) {
			continue
		}
		r.CreatorIdentifier = ""
		pb.Messages = append(pb.Messages, ticket.ToProtoTicketReply(r))
	}
	return pb
}

func ToCreatePublicAppealResponse(t *ticket.Ticket) *modlv1.CreatePublicAppealResponse {
	wf := t.AppealWorkflowStatus
	if wf == "" {
		wf = string(t.Status)
	}
	if wf == "" {
		wf = string(ticket.AppealOpen)
	}
	created := ""
	if t.CreatedAtMs > 0 {
		created = ticket.MsToTime(t.CreatedAtMs).Format("2006-01-02T15:04:05.000Z")
	}
	return &modlv1.CreatePublicAppealResponse{
		Success:  true,
		AppealId: t.ID,
		Message:  "Appeal created successfully",
		Appeal: &modlv1.CreatePublicAppealResponse_CreatePublicAppealInfo{
			Id:                   t.ID,
			IdAlias:              t.ID,
			Type:                 string(ticket.CategoryAppeal),
			Subject:              t.Subject,
			Status:               wf,
			AppealWorkflowStatus: wf,
			Created:              created,
		},
	}
}

func ToAddPublicAppealReplyResponse(r ticket.TicketReply) *modlv1.AddPublicAppealReplyResponse {
	r.CreatorIdentifier = ""
	return &modlv1.AddPublicAppealReplyResponse{
		Success: true,
		Message: "Reply added successfully",
		Reply:   ticket.ToProtoPublicTicketReply(r),
	}
}
