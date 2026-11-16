package appeal

import (
	"context"
	"fmt"
	"sort"
	"strings"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/httpx"
	"modl-go/internal/models"
	"modl-go/internal/modules/ticket"
	"modl-go/internal/realtime"
)

type Service struct {
	store     *Store
	ticketSvc *ticket.Service
	hub       *realtime.Hub
}

func NewService(store *Store, ticketSvc *ticket.Service, hub *realtime.Hub) *Service {
	return &Service{
		store:     store,
		ticketSvc: ticketSvc,
		hub:       hub,
	}
}

func (s *Service) GetAppealsByPunishment(ctx context.Context, serverID, punishmentID string) ([]*ticket.Ticket, error) {
	return s.store.FindAppealsByPunishmentID(ctx, serverID, punishmentID)
}

func (s *Service) GetAppealRaw(ctx context.Context, serverID, appealID string) (*ticket.Ticket, error) {
	t, err := s.store.ticketStore.FindByID(ctx, serverID, appealID)
	if err != nil {
		return nil, err
	}
	if !t.Type.IsAppeal() {
		return nil, ticket.ErrNotFound
	}
	return t, nil
}

func (s *Service) CreateAppeal(ctx context.Context, srv *models.Server, req *modlv1.CreateAppealRequest, presentedToken string, tokenRequired bool) (*ticket.Ticket, error) {
	if strings.TrimSpace(req.PunishmentId) == "" {
		return nil, ticket.BadRequestError{Msg: "punishmentId is required"}
	}
	if !ticket.IsValidUUID(req.PlayerUuid) {
		return nil, ticket.BadRequestError{Msg: "Invalid playerUuid"}
	}
	email := ticket.NormalizeEmailIfValid(req.Email)
	if email == "" {
		return nil, ticket.BadRequestError{Msg: "Invalid email address"}
	}

	playerUUID := ticket.NormalizeUUID(req.PlayerUuid)
	info, err := s.store.FindPlayerWithPunishment(ctx, srv.ID, playerUUID, req.PunishmentId)
	if err != nil {
		return nil, err
	}

	tokenValid := strings.TrimSpace(presentedToken) != "" &&
		s.ticketSvc.ValidateAppealCreateToken(ctx, srv.ID, req.PunishmentId, playerUUID, presentedToken)
	if tokenRequired && !tokenValid {
		return nil, ticket.ForbiddenError{Msg: "Appeal verification required"}
	}

	existingAppeals, err := s.store.FindAppealsByPunishmentID(ctx, srv.ID, req.PunishmentId)
	if err != nil {
		return nil, err
	}
	for _, existing := range existingAppeals {
		wf, err := ticket.ParseAppealWorkflowStatus(existing.AppealWorkflowStatus)
		if existing.AppealWorkflowStatus == "" || err != nil || !wf.IsTerminal() {
			if tokenValid {
				return existing, nil
			}
			return nil, ticket.ConflictError{Msg: "An appeal already exists for this punishment"}
		}
	}
	if info.Pardoned {
		return nil, ticket.ConflictError{Msg: "This punishment has already been pardoned"}
	}

	data := map[string]any{
		"punishmentId": req.PunishmentId,
		"playerUuid":   playerUUID,
		"contactEmail": email,
		"creatorEmail": email,
	}
	if req.AdditionalData != nil {
		for k, v := range req.AdditionalData.AsMap() {
			data[k] = v
		}
	}

	now := httpx.NowMs()
	initialContent := buildInitialAppealContent(req)
	initialReply := ticket.TicketReply{
		ID:          httpx.NewID(),
		Name:        info.PlayerUsername,
		Content:     initialContent,
		Type:        "player",
		CreatedMs:   now,
		Staff:       false,
		Attachments: ticket.ValuesToAttachmentMaps(req.Attachments),
	}

	t := &ticket.Ticket{
		ServerID:             srv.ID,
		Type:                 ticket.CategoryAppeal,
		Subject:              fmt.Sprintf("Punishment Appeal - %s", info.PlayerUsername),
		CreatorUUID:          playerUUID,
		CreatorName:          info.PlayerUsername,
		PunishmentID:         req.PunishmentId,
		Tags:                 []string{"appeal"},
		Replies:              []ticket.TicketReply{initialReply},
		Notes:                []ticket.TicketNote{},
		Data:                 data,
		Priority:             ticket.PriorityNormal,
		AssignedTo:           []string{},
		AppealWorkflowStatus: string(ticket.AppealOpen),
		CreatedAtMs:          now,
		UpdatedAtMs:          now,
	}
	t.ApplyLifecycleStatus(ticket.StatusOpen)

	saved, err := s.store.ticketStore.InsertWithUniqueID(ctx, ticket.CategoryAppeal.TicketPrefix(), t)
	if err != nil {
		return nil, err
	}
	_ = s.store.LinkAppealToPunishment(ctx, srv.ID, playerUUID, req.PunishmentId, saved.ID)
	s.invalidateAppeal(srv.ID, saved.ID)
	return saved, nil
}

func buildInitialAppealContent(req *modlv1.CreateAppealRequest) string {
	var parts []string
	if req.Reason != nil && strings.TrimSpace(*req.Reason) != "" {
		parts = append(parts, "Reason: "+strings.TrimSpace(*req.Reason))
	}
	if req.Evidence != nil && strings.TrimSpace(*req.Evidence) != "" {
		parts = append(parts, "Evidence: "+strings.TrimSpace(*req.Evidence))
	}
	if req.AdditionalData != nil {
		m := req.AdditionalData.AsMap()
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := m[k]
			if v == nil {
				continue
			}
			str := strings.TrimSpace(fmt.Sprintf("%v", v))
			if str == "" {
				continue
			}
			parts = append(parts, fmt.Sprintf("%s: %s", ticket.ResolveFieldLabel(k, req.FieldLabels), str))
		}
	}
	if len(parts) == 0 {
		return "Appeal submitted."
	}
	return strings.Join(parts, "\n\n")
}

func (s *Service) AddPanelReply(ctx context.Context, srv *models.Server, appealID string, req *modlv1.AddAppealReplyRequest, actorEmail string) (ticket.TicketReply, error) {
	t, err := s.GetAppealRaw(ctx, srv.ID, appealID)
	if err != nil {
		return ticket.TicketReply{}, err
	}
	if t.Locked {
		return ticket.TicketReply{}, ticket.ConflictError{Msg: "Appeal is locked and cannot accept new replies"}
	}
	now := httpx.NowMs()
	name := strings.TrimSpace(req.Name)
	if req.Staff && strings.TrimSpace(actorEmail) != "" {
		if u := s.store.ticketStore.FindStaffUsernameByEmail(ctx, srv.ID, actorEmail); u != "" {
			name = u
		}
	}
	if name == "" {
		if req.Staff {
			name = "Staff"
		} else {
			name = t.CreatorName
		}
	}
	typ := strings.TrimSpace(req.Type)
	if typ == "" {
		if req.Staff {
			typ = "staff"
		} else {
			typ = "user"
		}
	}

	reply := ticket.TicketReply{
		ID:          httpx.NewID(),
		Name:        name,
		Content:     req.Content,
		Type:        typ,
		CreatedMs:   now,
		Staff:       req.Staff,
		Action:      req.GetAction(),
		Avatar:      req.GetAvatar(),
		Attachments: ticket.ValuesToAttachmentMaps(req.Attachments),
	}
	t.Replies = append(t.Replies, reply)
	t.UpdatedAtMs = now
	if err := s.store.ticketStore.SaveTicket(ctx, t); err != nil {
		return ticket.TicketReply{}, err
	}
	if req.Staff && strings.TrimSpace(actorEmail) != "" {
		s.ticketSvc.EnsureSubscription(ctx, srv.ID, actorEmail, t.ID)
	}
	s.ticketSvc.NotifyTicketReply(ctx, srv, t, reply)
	s.ticketSvc.RecordSubscriptionUpdates(ctx, srv, t, reply, actorEmail)
	s.invalidateAppeal(srv.ID, t.ID)
	return reply, nil
}

func (s *Service) AddPublicReplyFromProto(ctx context.Context, srv *models.Server, appealID string, req *modlv1.AddAppealReplyRequest) (ticket.TicketReply, error) {
	t, err := s.GetAppealRaw(ctx, srv.ID, appealID)
	if err != nil {
		return ticket.TicketReply{}, err
	}
	if t.Locked {
		return ticket.TicketReply{}, ticket.ConflictError{Msg: "Appeal is locked and cannot accept new replies"}
	}
	now := httpx.NowMs()
	reply := ticket.TicketReply{
		ID:          httpx.NewID(),
		Name:        t.CreatorName,
		Content:     req.Content,
		Type:        "user",
		CreatedMs:   now,
		Staff:       false,
		Attachments: ticket.ValuesToAttachmentMaps(req.Attachments),
	}
	t.Replies = append(t.Replies, reply)
	t.UpdatedAtMs = now
	if err := s.store.ticketStore.SaveTicket(ctx, t); err != nil {
		return ticket.TicketReply{}, err
	}
	s.ticketSvc.RecordSubscriptionUpdates(ctx, srv, t, reply, "")
	s.invalidateAppeal(srv.ID, t.ID)
	return reply, nil
}

func (s *Service) UpdateStatus(ctx context.Context, srv *models.Server, appealID string, req *modlv1.UpdateAppealStatusRequest, actorEmail string) (*ticket.Ticket, error) {
	t, err := s.GetAppealRaw(ctx, srv.ID, appealID)
	if err != nil {
		return nil, err
	}

	staffName := ""
	if req.StaffUsername != nil && strings.TrimSpace(*req.StaffUsername) != "" {
		staffName = strings.TrimSpace(*req.StaffUsername)
	} else if actorEmail != "" {
		staffName = s.store.ticketStore.ResolvePanelActor(ctx, srv.ID, actorEmail)
	}
	if staffName == "" || staffName == "Unknown" {
		staffName = "System"
	}

	currentWF, _ := ticket.ParseAppealWorkflowStatus(t.AppealWorkflowStatus)
	var requestedWF ticket.AppealWorkflowStatus
	hasRequestedWF := false
	if strings.TrimSpace(req.Status) != "" {
		parsed, err := ticket.ParseAppealWorkflowStatus(req.Status)
		if err != nil {
			return nil, ticket.BadRequestError{Msg: "Invalid appeal status"}
		}
		requestedWF = parsed
		hasRequestedWF = true
	}

	if hasRequestedWF && currentWF != "" && currentWF.IsTerminal() && requestedWF.IsTerminal() && currentWF != requestedWF {
		return nil, ticket.ConflictError{
			Msg: fmt.Sprintf("Appeal is already %s and cannot be changed directly to %s; reopen the appeal first.",
				currentWF.DisplayName(), requestedWF.DisplayName()),
		}
	}

	now := httpx.NowMs()
	wasClosed := t.Status.IsTerminal()
	statusChanged := false
	var systemReplies []ticket.TicketReply

	if hasRequestedWF && requestedWF != currentWF {
		lifecycle := requestedWF.ToTicketStatus()
		t.AppealWorkflowStatus = string(requestedWF)
		t.ApplyLifecycleStatus(lifecycle)
		statusChanged = true

		systemReplies = append(systemReplies, ticket.TicketReply{
			ID:        httpx.NewID(),
			Name:      staffName,
			Content:   fmt.Sprintf("Appeal status changed to %s.", requestedWF.DisplayName()),
			Type:      "system",
			CreatedMs: now,
			Staff:     true,
			Action:    "APPEAL_STATUS_" + strings.ToUpper(string(requestedWF)),
		})
	}

	if req.Resolution != nil {
		if t.Data == nil {
			t.Data = make(map[string]any)
		}
		prev, _ := t.Data["resolution"].(string)
		if *req.Resolution != prev {
			t.Data["resolution"] = *req.Resolution
			act := "RESOLUTION_" + strings.ToUpper(strings.ReplaceAll(*req.Resolution, " ", "_"))
			systemReplies = append(systemReplies, ticket.TicketReply{
				ID:        httpx.NewID(),
				Name:      staffName,
				Content:   fmt.Sprintf("Resolution set to %s.", *req.Resolution),
				Type:      "system",
				CreatedMs: now,
				Staff:     true,
				Action:    act,
			})
		}
	}

	if req.Locked != nil && *req.Locked != t.Locked {
		t.Locked = *req.Locked
		verb := "unlocked"
		act := "UNLOCKED"
		if *req.Locked {
			verb = "locked"
			act = "LOCKED"
		}
		systemReplies = append(systemReplies, ticket.TicketReply{
			ID:        httpx.NewID(),
			Name:      staffName,
			Content:   fmt.Sprintf("Ticket %s.", verb),
			Type:      "system",
			CreatedMs: now,
			Staff:     true,
			Action:    act,
		})
	}

	t.Replies = append(t.Replies, systemReplies...)
	t.UpdatedAtMs = now
	if err := s.store.ticketStore.SaveTicket(ctx, t); err != nil {
		return nil, err
	}

	if statusChanged && hasRequestedWF && requestedWF.IsTerminal() {
		punishmentID := t.PunishmentID
		if punishmentID == "" && t.Data != nil {
			punishmentID, _ = t.Data["punishmentId"].(string)
		}
		playerUUID := t.CreatorUUID
		if playerUUID == "" && t.Data != nil {
			playerUUID, _ = t.Data["playerUuid"].(string)
		}
		if punishmentID != "" && playerUUID != "" {
			if requestedWF == ticket.AppealApproved {
				_ = s.store.ApplyAppealApproval(ctx, srv.ID, playerUUID, punishmentID, staffName, t.ID)
			} else if requestedWF == ticket.AppealRejected {
				_ = s.store.RecordAppealRejection(ctx, srv.ID, playerUUID, punishmentID, staffName, t.ID)
			}
		}
	}

	if !wasClosed && t.Status.IsTerminal() {
		s.ticketSvc.NotifyTicketClosed(ctx, srv, t)
	}

	s.invalidateAppeal(srv.ID, t.ID)
	if s.hub != nil {
		s.hub.InvalidatePanel(srv.ID, modlv1.PanelResource_PANEL_RESOURCE_PUNISHMENTS, "")
	}
	return t, nil
}

func (s *Service) invalidateAppeal(serverID, appealID string) {
	if s.hub != nil {
		s.hub.InvalidatePanel(serverID, modlv1.PanelResource_PANEL_RESOURCE_APPEALS, appealID)
		s.hub.InvalidatePanel(serverID, modlv1.PanelResource_PANEL_RESOURCE_TICKETS, appealID)
	}
}
