package ticket

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"log"
	"net/mail"
	"sort"
	"strings"
	"time"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/ai"
	"modl-go/internal/config"
	"modl-go/internal/email"
	"modl-go/internal/httpx"
	"modl-go/internal/models"
	"modl-go/internal/realtime"

	"google.golang.org/protobuf/types/known/structpb"
)

const (
	MaxVerificationAttempts       = 5
	VerificationCodeExpiryMs      = 5 * 60 * 1000
	VerificationTokenExpiryMs     = 5 * 60 * 1000
	DefaultReplyPlaceholder       = ""
	StaffNoteType                 = "staff-note"
)

var (
	ErrTicketLocked    = errors.New("ticket is locked")
	ErrAlreadyClaimed  = errors.New("ticket is already claimed")
	ErrBadRequest      = errors.New("bad request")
	ErrForbidden       = errors.New("forbidden")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrTooManyAttempts = errors.New("too many failed verification attempts. Please request a new code.")
)

type BadRequestError struct{ Msg string }

func (e BadRequestError) Error() string { return e.Msg }

type ConflictError struct{ Msg string }

func (e ConflictError) Error() string { return e.Msg }

type UnauthorizedError struct{ Msg string }

func (e UnauthorizedError) Error() string { return e.Msg }

type ForbiddenError struct{ Msg string }

func (e ForbiddenError) Error() string { return e.Msg }

type Service struct {
	store    *Store
	cfg      *config.Config
	mailer   *email.Sender
	hub      *realtime.Hub
	aiClient *ai.Client
}

func NewService(store *Store, cfg *config.Config, mailer *email.Sender, hub *realtime.Hub, aiClient *ai.Client) *Service {
	return &Service{
		store:    store,
		cfg:      cfg,
		mailer:   mailer,
		hub:      hub,
		aiClient: aiClient,
	}
}

func (s *Service) Store() *Store {
	return s.store
}

// --- Email & Verification Helpers ---

func NormalizeEmail(raw string) string {
	n := strings.ToLower(strings.TrimSpace(raw))
	if n == "" {
		return ""
	}
	return n
}

func IsValidEmail(raw string) bool {
	n := NormalizeEmail(raw)
	if n == "" || len(n) > 254 {
		return false
	}
	addr, err := mail.ParseAddress(n)
	if err != nil {
		return false
	}
	return strings.EqualFold(addr.Address, n)
}

func NormalizeEmailIfValid(raw string) string {
	if IsValidEmail(raw) {
		return NormalizeEmail(raw)
	}
	return ""
}

func MaskEmail(raw string) string {
	n := NormalizeEmail(raw)
	if n == "" {
		return "<empty>"
	}
	at := strings.IndexByte(n, '@')
	if at <= 0 {
		return "***"
	}
	return n[:1] + "***" + n[at:]
}

func ResolveContactEmail(t *Ticket) string {
	if t == nil {
		return ""
	}
	if t.Data != nil {
		if v, ok := t.Data["creatorEmail"].(string); ok {
			if e := NormalizeEmailIfValid(v); e != "" {
				return e
			}
		}
		if v, ok := t.Data["contactEmail"].(string); ok {
			if e := NormalizeEmailIfValid(v); e != "" {
				return e
			}
		}
	}
	if t.FormData != nil {
		if v, ok := t.FormData["creatorEmail"].(string); ok {
			if e := NormalizeEmailIfValid(v); e != "" {
				return e
			}
		}
		if v, ok := t.FormData["email"].(string); ok {
			if e := NormalizeEmailIfValid(v); e != "" {
				return e
			}
		}
	}
	if strings.Contains(t.CreatorUUID, "@") {
		if e := NormalizeEmailIfValid(t.CreatorUUID); e != "" {
			return e
		}
	}
	return ""
}

func (s *Service) RequestVerificationCode(ctx context.Context, srv *models.Server, ticketID string) (string, error) {
	t, err := s.store.FindByID(ctx, srv.ID, ticketID)
	if err != nil {
		return "", err
	}
	if t.Hidden {
		return "", ErrNotFound
	}
	contactEmail := ResolveContactEmail(t)
	if contactEmail == "" {
		return "", BadRequestError{Msg: "No email address associated with this ticket"}
	}

	code := httpx.GenerateNumericCode(6)
	secret := ""
	if s.cfg != nil {
		secret = s.cfg.AuthCodeHashSecret
	}
	v := &TicketVerification{
		ID:             httpx.NewID(),
		ServerID:       srv.ID,
		TicketID:       ticketID,
		Token:          "",
		CodeHash:       httpx.HashOneTimeCode(code, secret),
		Email:          contactEmail,
		FailedAttempts: 0,
		ExpiresAt:      httpx.NowMs() + VerificationCodeExpiryMs,
	}
	if err := s.store.UpsertVerification(ctx, v); err != nil {
		return "", err
	}
	if s.mailer != nil {
		_ = s.mailer.SendTicketVerification(contactEmail, srv.ServerName, code)
	}
	return MaskEmail(contactEmail), nil
}

func (s *Service) VerifyCode(ctx context.Context, srv *models.Server, ticketID, code string) (string, error) {
	t, err := s.store.FindByID(ctx, srv.ID, ticketID)
	if err != nil {
		return "", err
	}
	if t.Hidden {
		return "", ErrNotFound
	}
	v, err := s.store.FindVerificationBySubject(ctx, srv.ID, ticketID)
	if err != nil {
		return "", BadRequestError{Msg: "No verification code requested for this ticket"}
	}
	if v.ExpiresAt < httpx.NowMs() {
		_ = s.store.DeleteVerificationByID(ctx, v.ID)
		return "", UnauthorizedError{Msg: "Verification code has expired. Please request a new one."}
	}
	if v.FailedAttempts >= MaxVerificationAttempts {
		_ = s.store.DeleteVerificationByID(ctx, v.ID)
		return "", UnauthorizedError{Msg: ErrTooManyAttempts.Error()}
	}
	secret := ""
	if s.cfg != nil {
		secret = s.cfg.AuthCodeHashSecret
	}
	if !httpx.VerifyOneTimeCode(strings.TrimSpace(code), v.CodeHash, secret) {
		count, _ := s.store.IncrementVerificationFailure(ctx, v.ID)
		if count >= MaxVerificationAttempts {
			_ = s.store.DeleteVerificationByID(ctx, v.ID)
			return "", UnauthorizedError{Msg: ErrTooManyAttempts.Error()}
		}
		return "", UnauthorizedError{Msg: "Invalid verification code"}
	}
	token := httpx.RandomToken(32)
	if err := s.store.UpdateVerificationToken(ctx, v.ID, token, httpx.NowMs()+VerificationTokenExpiryMs); err != nil {
		return "", err
	}
	return token, nil
}

func (s *Service) ValidateToken(ctx context.Context, serverID, subject, token string) bool {
	if strings.TrimSpace(token) == "" || strings.TrimSpace(subject) == "" {
		return false
	}
	v, err := s.store.FindVerificationBySubject(ctx, serverID, subject)
	if err != nil || v.Token == "" {
		return false
	}
	if v.ExpiresAt <= httpx.NowMs() {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(v.Token), []byte(strings.TrimSpace(token))) == 1
}

func (s *Service) ValidateAppealCreateToken(ctx context.Context, serverID, punishmentID, playerUUID, token string) bool {
	subject := fmt.Sprintf("appeal-create:%s:%s", punishmentID, playerUUID)
	return s.ValidateToken(ctx, serverID, subject, token)
}

type PublicAccessResult struct {
	Allowed   bool
	NotFound  bool
	Verified  bool
	EmailHint string
}

func (s *Service) AuthorizePublicAccess(ctx context.Context, srv *models.Server, t *Ticket, presentedToken string) PublicAccessResult {
	if t == nil || t.Hidden {
		return PublicAccessResult{NotFound: true}
	}
	if s.ValidateToken(ctx, srv.ID, t.ID, presentedToken) {
		return PublicAccessResult{Allowed: true, Verified: true}
	}
	if contactEmail := ResolveContactEmail(t); contactEmail != "" {
		return PublicAccessResult{Allowed: false, EmailHint: MaskEmail(contactEmail)}
	}
	if t.EmailAuthEnabled {
		return PublicAccessResult{Allowed: false, EmailHint: ""}
	}
	return PublicAccessResult{Allowed: true, Verified: false}
}

func (s *Service) AuthorizePublicSubmission(ctx context.Context, srv *models.Server, t *Ticket, presentedToken string) PublicAccessResult {
	if t == nil || t.Hidden {
		return PublicAccessResult{NotFound: true}
	}
	if s.ValidateToken(ctx, srv.ID, t.ID, presentedToken) {
		return PublicAccessResult{Allowed: true, Verified: true}
	}
	if contactEmail := ResolveContactEmail(t); contactEmail != "" {
		return PublicAccessResult{Allowed: false, EmailHint: MaskEmail(contactEmail)}
	}
	return PublicAccessResult{Allowed: true, Verified: false}
}

func (s *Service) BuildFormFieldAllowlist(ctx context.Context, serverID string, typ TicketCategory) map[string]bool {
	cfg := s.store.GetFormByType(ctx, serverID, string(typ))
	if cfg == nil || len(cfg.Fields) == 0 {
		return map[string]bool{}
	}
	allow := make(map[string]bool, len(cfg.Fields))
	for _, f := range cfg.Fields {
		if f == nil || strings.TrimSpace(f.Id) == "" {
			continue
		}
		if strings.EqualFold(f.Type, "email") || strings.Contains(strings.ToLower(f.Id), "email") {
			continue
		}
		allow[f.Id] = true
	}
	return allow
}

// --- Notifications ---

func (s *Service) NotifyTicketReply(ctx context.Context, srv *models.Server, t *Ticket, reply TicketReply) {
	if !reply.Staff {
		return
	}
	if contactEmail := ResolveContactEmail(t); contactEmail != "" && s.mailer != nil {
		author := html.EscapeString(nonEmpty(reply.Name, "Staff"))
		content := strings.ReplaceAll(html.EscapeString(reply.Content), "\n", "<br>")
		subject := fmt.Sprintf("[%s] New reply on ticket #%s", srv.ServerName, t.ID)
		body := fmt.Sprintf("<p><b>%s</b> replied to your ticket <b>#%s</b> (%s):</p><blockquote>%s</blockquote>",
			author, html.EscapeString(t.ID), html.EscapeString(t.Subject), content)
		_ = s.mailer.Send(contactEmail, subject, body)
	}
	if strings.TrimSpace(t.CreatorUUID) != "" && s.store.PlayerExists(ctx, srv.ID, t.CreatorUUID) {
		author := nonEmpty(reply.Name, "Staff")
		msg := fmt.Sprintf("%s replied to your ticket #%s", author, t.ID)
		notifID := httpx.NewID()
		now := httpx.NowMs()
		dataMap := map[string]any{
			"ticketId":         t.ID,
			"replyAuthor":      author,
			"targetPlayerUuid": strings.ToLower(t.CreatorUUID),
		}
		dataJSON, _ := json.Marshal(dataMap)
		_ = s.store.InsertPlayerNotification(ctx, notifID, srv.ID, t.CreatorUUID, "TICKET_REPLY", "Ticket Reply", msg, string(dataJSON), now)
		if s.hub != nil {
			uuid := strings.ToLower(t.CreatorUUID)
			s.hub.PushPlayerNotifications(srv.ID, []*modlv1.SyncPlayerNotification{{
				Id:               notifID,
				Message:          msg,
				Type:             "TICKET_REPLY",
				TargetPlayerUuid: &uuid,
				Data:             MapToStruct(dataMap),
				Timestamp:        &now,
			}})
		}
	}
}

func (s *Service) NotifyTicketClosed(ctx context.Context, srv *models.Server, t *Ticket) {
	if contactEmail := ResolveContactEmail(t); contactEmail != "" && s.mailer != nil {
		subject := fmt.Sprintf("[%s] Ticket #%s closed", srv.ServerName, t.ID)
		body := fmt.Sprintf("<p>Your ticket <b>#%s</b> (%s) has been closed.</p>",
			html.EscapeString(t.ID), html.EscapeString(t.Subject))
		_ = s.mailer.Send(contactEmail, subject, body)
	}
	if strings.TrimSpace(t.CreatorUUID) != "" && s.store.PlayerExists(ctx, srv.ID, t.CreatorUUID) {
		msg := fmt.Sprintf("Your ticket #%s has been closed", t.ID)
		notifID := httpx.NewID()
		now := httpx.NowMs()
		dataMap := map[string]any{
			"ticketId":         t.ID,
			"targetPlayerUuid": strings.ToLower(t.CreatorUUID),
		}
		dataJSON, _ := json.Marshal(dataMap)
		_ = s.store.InsertPlayerNotification(ctx, notifID, srv.ID, t.CreatorUUID, "TICKET_CLOSED", "Ticket Closed", msg, string(dataJSON), now)
		if s.hub != nil {
			uuid := strings.ToLower(t.CreatorUUID)
			s.hub.PushPlayerNotifications(srv.ID, []*modlv1.SyncPlayerNotification{{
				Id:               notifID,
				Message:          msg,
				Type:             "TICKET_CLOSED",
				TargetPlayerUuid: &uuid,
				Data:             MapToStruct(dataMap),
				Timestamp:        &now,
			}})
		}
	}
}

func (s *Service) InvalidateTickets(serverID, ticketID string) {
	if s.hub != nil {
		s.hub.InvalidatePanel(serverID, modlv1.PanelResource_PANEL_RESOURCE_TICKETS, ticketID)
	}
}

func (s *Service) InvalidateNotifications(serverID string) {
	if s.hub != nil {
		s.hub.InvalidatePanel(serverID, modlv1.PanelResource_PANEL_RESOURCE_NOTIFICATIONS, "")
	}
}

// --- Subscriptions ---

func (s *Service) MarkTicketAsRead(ctx context.Context, serverID, staffEmail, ticketID string) {
	if strings.TrimSpace(staffEmail) == "" || strings.TrimSpace(ticketID) == "" {
		return
	}
	subs, err := s.store.FindStaffSubscriptions(ctx, serverID, staffEmail)
	if err != nil {
		return
	}
	changed := false
	now := httpx.NowMs()
	for i := range subs {
		if subs[i].TicketID == ticketID && subs[i].Active {
			subs[i].LastReadAt = now
			changed = true
		}
	}
	if changed {
		_ = s.store.SaveStaffSubscriptions(ctx, serverID, staffEmail, subs)
	}
}

func (s *Service) EnsureSubscription(ctx context.Context, serverID, staffEmail, ticketID string) {
	if strings.TrimSpace(staffEmail) == "" || strings.TrimSpace(ticketID) == "" {
		return
	}
	subs, err := s.store.FindStaffSubscriptions(ctx, serverID, staffEmail)
	if err != nil {
		return
	}
	now := httpx.NowMs()
	for _, sub := range subs {
		if sub.TicketID == ticketID && sub.Active {
			s.MarkTicketAsRead(ctx, serverID, staffEmail, ticketID)
			return
		}
	}
	subs = append(subs, StaffTicketSubscription{
		TicketID:     ticketID,
		SubscribedAt: now,
		LastReadAt:   now,
		Active:       true,
	})
	_ = s.store.SaveStaffSubscriptions(ctx, serverID, staffEmail, subs)
}

func (s *Service) Unsubscribe(ctx context.Context, serverID, staffEmail, ticketID string) (bool, error) {
	subs, err := s.store.FindStaffSubscriptions(ctx, serverID, staffEmail)
	if err != nil {
		return false, err
	}
	changed := false
	for i := range subs {
		if subs[i].TicketID == ticketID && subs[i].Active {
			subs[i].Active = false
			changed = true
		}
	}
	if !changed {
		return false, nil
	}
	if err := s.store.SaveStaffSubscriptions(ctx, serverID, staffEmail, subs); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Service) GetSubscriptions(ctx context.Context, serverID, staffEmail string) (*modlv1.TicketSubscriptionsResponse, error) {
	subs, err := s.store.FindStaffSubscriptions(ctx, serverID, staffEmail)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return &modlv1.TicketSubscriptionsResponse{Subscriptions: []*modlv1.TicketSubscriptionResponse{}}, nil
		}
		return nil, err
	}
	var ids []string
	var active []StaffTicketSubscription
	for _, sub := range subs {
		if sub.Active && sub.TicketID != "" {
			active = append(active, sub)
			ids = append(ids, sub.TicketID)
		}
	}
	tickets, err := s.store.FindByIDs(ctx, serverID, ids)
	if err != nil {
		return nil, err
	}
	titleMap := make(map[string]string, len(tickets))
	for _, t := range tickets {
		titleMap[t.ID] = nonEmpty(t.Subject, "Untitled Ticket")
	}
	resp := &modlv1.TicketSubscriptionsResponse{
		Subscriptions: make([]*modlv1.TicketSubscriptionResponse, 0, len(active)),
	}
	for _, sub := range active {
		title := titleMap[sub.TicketID]
		if title == "" {
			title = "Ticket #" + sub.TicketID
		}
		resp.Subscriptions = append(resp.Subscriptions, &modlv1.TicketSubscriptionResponse{
			TicketId:     sub.TicketID,
			TicketTitle:  title,
			SubscribedAt: sub.SubscribedAt,
		})
	}
	return resp, nil
}

func (s *Service) RecordSubscriptionUpdates(ctx context.Context, srv *models.Server, t *Ticket, reply TicketReply, actorEmail string) {
	allStaff, err := s.store.ListAllStaffWithSubscriptions(ctx, srv.ID)
	if err != nil {
		return
	}
	normActor := strings.ToLower(strings.TrimSpace(actorEmail))
	actorName := nonEmpty(reply.Name, "User")
	now := httpx.NowMs()
	anyInserted := false
	for _, st := range allStaff {
		if !st.NotifyEnabled || !st.TicketReplies {
			continue
		}
		if normActor != "" && strings.EqualFold(st.Email, normActor) {
			continue
		}
		subscribed := false
		for _, sub := range st.Subscriptions {
			if sub.TicketID == t.ID && sub.Active {
				subscribed = true
				break
			}
		}
		if !subscribed {
			continue
		}
		summary := reply.Content
		if len(summary) > 200 {
			summary = summary[:200]
		}
		_ = s.store.InsertSubscriptionUpdate(ctx, &SubscriptionUpdate{
			ID:         httpx.NewID(),
			ServerID:   srv.ID,
			StaffEmail: st.Email,
			TicketID:   t.ID,
			EventType:  "reply",
			Summary:    summary,
			Actor:      actorName,
			Read:       false,
			CreatedAt:  now,
		})
		anyInserted = true
	}
	if anyInserted {
		s.InvalidateNotifications(srv.ID)
	}
}

func (s *Service) GetSubscriptionUpdates(ctx context.Context, serverID, staffEmail string, limit int) (*modlv1.SubscriptionUpdatesResponse, error) {
	updates, _, err := s.store.FindSubscriptionUpdates(ctx, serverID, staffEmail, limit)
	if err != nil {
		return nil, err
	}
	idSet := make(map[string]struct{})
	var ids []string
	for _, u := range updates {
		if _, ok := idSet[u.TicketID]; !ok {
			idSet[u.TicketID] = struct{}{}
			ids = append(ids, u.TicketID)
		}
	}
	tickets, _ := s.store.FindByIDs(ctx, serverID, ids)
	ticketMap := make(map[string]*Ticket, len(tickets))
	for _, t := range tickets {
		ticketMap[t.ID] = t
	}
	out := make([]*modlv1.SubscriptionUpdateResponse, 0, len(updates))
	for _, u := range updates {
		title := "Ticket #" + u.TicketID
		isStaff := false
		if t, ok := ticketMap[u.TicketID]; ok {
			title = nonEmpty(t.Subject, title)
			for i := len(t.Replies) - 1; i >= 0; i-- {
				if t.Replies[i].CreatedMs == u.CreatedAt || t.Replies[i].Content == u.Summary {
					isStaff = t.Replies[i].Staff
					break
				}
			}
		}
		out = append(out, &modlv1.SubscriptionUpdateResponse{
			Id:           u.ID,
			TicketId:     u.TicketID,
			TicketTitle:  title,
			ReplyContent: u.Summary,
			ReplyBy:      u.Actor,
			ReplyAt:      u.CreatedAt,
			IsStaffReply: isStaff,
			IsRead:       u.Read,
		})
	}
	return &modlv1.SubscriptionUpdatesResponse{Updates: out}, nil
}

func (s *Service) GetAssignedTicketUpdates(ctx context.Context, serverID, staffEmail string, limit int) (*modlv1.SubscriptionUpdatesResponse, error) {
	username := s.store.FindStaffUsernameByEmail(ctx, serverID, staffEmail)
	if username == "" {
		return &modlv1.SubscriptionUpdatesResponse{Updates: []*modlv1.SubscriptionUpdateResponse{}}, nil
	}
	tickets, err := s.store.FindRecentAssignedWithReplies(ctx, serverID, username, limit)
	if err != nil {
		return nil, err
	}
	out := make([]*modlv1.SubscriptionUpdateResponse, 0, len(tickets))
	for _, t := range tickets {
		if len(t.Replies) == 0 {
			continue
		}
		last := t.Replies[len(t.Replies)-1]
		out = append(out, &modlv1.SubscriptionUpdateResponse{
			Id:           t.ID + ":" + last.ID,
			TicketId:     t.ID,
			TicketTitle:  nonEmpty(t.Subject, "Ticket #"+t.ID),
			ReplyContent: last.Content,
			ReplyBy:      nonEmpty(last.Name, "User"),
			ReplyAt:      last.CreatedMs,
			IsStaffReply: last.Staff,
			IsRead:       false,
		})
	}
	return &modlv1.SubscriptionUpdatesResponse{Updates: out}, nil
}

// --- Ticket Creation & Mutations ---

func (s *Service) CreateTicket(ctx context.Context, srv *models.Server, req *modlv1.CreateTicketRequest, isPublic, forceUnfinished bool) (*Ticket, error) {
	cat, err := ParseTicketCategory(req.Type)
	if err != nil {
		return nil, BadRequestError{Msg: "Invalid ticket type"}
	}

	creatorEmail := ""
	if req.CreatorEmail != nil && strings.TrimSpace(*req.CreatorEmail) != "" {
		creatorEmail = NormalizeEmailIfValid(*req.CreatorEmail)
		if creatorEmail == "" {
			return nil, BadRequestError{Msg: "Invalid creator email"}
		}
	}

	formData := StructToMap(req.FormData)
	data := make(map[string]any)
	if creatorEmail != "" {
		data["creatorEmail"] = creatorEmail
	}
	if req.CreatorIdentifier != nil && strings.TrimSpace(*req.CreatorIdentifier) != "" {
		data["creatorIdentifier"] = strings.TrimSpace(*req.CreatorIdentifier)
	}

	emailAuthEnabled := req.GetEmailAuthEnabled()
	if isPublic && !forceUnfinished {
		if cfg := s.store.GetFormByType(ctx, srv.ID, string(cat)); cfg != nil {
			if cfg.RequireEmail && creatorEmail == "" {
				return nil, BadRequestError{Msg: "Email is required for this ticket type"}
			}
			if cfg.RequireEmailAuth {
				emailAuthEnabled = true
			}
		}
	}

	now := httpx.NowMs()
	creatorName := strings.TrimSpace(req.GetCreatorName())
	if creatorName == "" {
		creatorName = "API User"
	}
	creatorUUID := strings.TrimSpace(req.GetCreatorUuid())

	desc := ""
	if req.Description != nil {
		desc = strings.TrimSpace(*req.Description)
	}
	if desc == "" && len(formData) > 0 {
		desc = formatFormDataDescription(formData, req.FieldLabels)
	}

	var replies []TicketReply
	if desc != "" || len(req.Attachments) > 0 {
		replies = append(replies, TicketReply{
			ID:                httpx.NewID(),
			Name:              creatorName,
			Content:           desc,
			Type:              "user",
			CreatedMs:         now,
			Staff:             false,
			Attachments:       StructsToAttachmentMaps(req.Attachments),
			CreatorIdentifier: req.GetCreatorIdentifier(),
		})
	}

	var chatMsgs []ChatMessage
	for _, cm := range req.ChatMessages {
		if cm == nil {
			continue
		}
		m := cm.AsMap()
		content, _ := m["content"].(string)
		sender, _ := m["sender"].(string)
		ts := now
		if rawTS, ok := m["timestamp"].(float64); ok && rawTS > 0 {
			ts = int64(rawTS)
		}
		chatMsgs = append(chatMsgs, ChatMessage{
			Content:     content,
			TimestampMs: ts,
			Sender:      sender,
		})
	}

	tags := make([]string, 0, len(req.Tags))
	for _, tg := range req.Tags {
		if strings.TrimSpace(tg) != "" {
			tags = append(tags, strings.TrimSpace(tg))
		}
	}

	prio := PriorityNormal
	if req.Priority != nil && strings.TrimSpace(*req.Priority) != "" {
		if p, err := ParseTicketPriority(*req.Priority); err == nil {
			prio = p
		}
	}

	status := StatusOpen
	if forceUnfinished {
		status = StatusUnfinished
	} else if !cat.IsReport() && strings.TrimSpace(req.GetSubject()) == "" {
		status = StatusUnfinished
	}

	subject := strings.TrimSpace(req.GetSubject())
	if subject == "" {
		subject = cat.DisplayName()
	}

	t := &Ticket{
		ServerID:           srv.ID,
		Type:               cat,
		Subject:            subject,
		CreatorUUID:        creatorUUID,
		CreatorName:        creatorName,
		ReportedPlayer:     strings.TrimSpace(req.GetReportedPlayerName()),
		ReportedPlayerUUID: strings.TrimSpace(req.GetReportedPlayerUuid()),
		Tags:               tags,
		Replies:            replies,
		Notes:              []TicketNote{},
		ChatMessages:       chatMsgs,
		FormData:           formData,
		Data:               data,
		Priority:           prio,
		AssignedTo:         []string{},
		EmailAuthEnabled:   emailAuthEnabled,
		Hidden:             false,
		CreatedAtMs:        now,
		UpdatedAtMs:        now,
	}
	t.ApplyLifecycleStatus(status)
	if cat.IsAppeal() {
		t.AppealWorkflowStatus = string(AppealOpen)
	}

	saved, err := s.store.InsertWithUniqueID(ctx, cat.TicketPrefix(), t)
	if err != nil {
		return nil, err
	}
	if !forceUnfinished {
		s.InvalidateTickets(srv.ID, saved.ID)
	}
	if saved.EmailAuthEnabled && creatorEmail != "" {
		_, _ = s.RequestVerificationCode(ctx, srv, saved.ID)
	}
	return saved, nil
}

func (s *Service) SubmitExistingTicketForm(ctx context.Context, srv *models.Server, t *Ticket, req *modlv1.SubmitTicketFormRequest, emailVerified bool) (*Ticket, error) {
	if t.Locked || t.Status.IsTerminal() {
		return nil, ConflictError{Msg: "Ticket is locked and cannot be resubmitted"}
	}
	now := httpx.NowMs()
	if t.Status != StatusOpen {
		t.ApplyLifecycleStatus(StatusOpen)
	}
	t.UpdatedAtMs = now
	if strings.TrimSpace(req.Subject) != "" {
		t.Subject = strings.TrimSpace(req.Subject)
	}
	if t.Data == nil {
		t.Data = make(map[string]any)
	}
	if t.FormData == nil {
		t.FormData = make(map[string]any)
	}
	reqFormData := StructToMap(req.FormData)
	for k, v := range reqFormData {
		t.FormData[k] = v
		if k != "creatorEmail" && k != "creatorIdentifier" {
			t.Data[k] = v
		}
	}
	if v, ok := reqFormData["emailAuthEnabled"].(bool); ok && v {
		t.EmailAuthEnabled = true
		t.Data["emailAuthEnabled"] = true
	}
	if cfg := s.store.GetFormByType(ctx, srv.ID, string(t.Type)); cfg != nil && cfg.RequireEmailAuth {
		t.EmailAuthEnabled = true
		t.Data["emailAuthEnabled"] = true
	}

	existingEmail, _ := t.Data["creatorEmail"].(string)
	reqEmail := ""
	if req.CreatorEmail != nil {
		reqEmail = NormalizeEmailIfValid(*req.CreatorEmail)
	}
	if reqEmail != "" && (existingEmail == "" || emailVerified) {
		t.Data["creatorEmail"] = reqEmail
	}
	if req.CreatorIdentifier != nil && strings.TrimSpace(*req.CreatorIdentifier) != "" {
		t.Data["creatorIdentifier"] = strings.TrimSpace(*req.CreatorIdentifier)
	}

	desc := formatFormDataDescription(reqFormData, req.FieldLabels)
	attachments := StructsToAttachmentMaps(req.Attachments)
	if desc != "" || len(attachments) > 0 {
		t.Replies = append(t.Replies, TicketReply{
			ID:                httpx.NewID(),
			Name:              nonEmpty(t.CreatorName, "Web User"),
			Content:           desc,
			Type:              "user",
			CreatedMs:         now,
			Staff:             false,
			Attachments:       attachments,
			CreatorIdentifier: req.GetCreatorIdentifier(),
		})
	}
	if err := s.store.SaveTicket(ctx, t); err != nil {
		return nil, err
	}
	s.InvalidateTickets(srv.ID, t.ID)
	return t, nil
}

func (s *Service) SubmitTicketForm(ctx context.Context, srv *models.Server, formType string, req *modlv1.SubmitTicketFormRequest) (*Ticket, error) {
	cat, err := ParseTicketCategory(formType)
	if err != nil {
		return nil, BadRequestError{Msg: "Invalid ticket type"}
	}
	cfg := s.store.GetFormByType(ctx, srv.ID, string(cat))

	creatorEmail := ""
	if req.CreatorEmail != nil && strings.TrimSpace(*req.CreatorEmail) != "" {
		creatorEmail = NormalizeEmailIfValid(*req.CreatorEmail)
		if creatorEmail == "" {
			return nil, BadRequestError{Msg: "Invalid email address"}
		}
	}
	if cfg != nil && cfg.RequireEmail && creatorEmail == "" {
		return nil, BadRequestError{Msg: "Email is required"}
	}

	formData := StructToMap(req.FormData)
	fieldLabels := req.FieldLabels
	if len(fieldLabels) == 0 && cfg != nil && len(cfg.Fields) > 0 {
		fieldLabels = make(map[string]string, len(cfg.Fields))
		for _, f := range cfg.Fields {
			if f != nil && f.Id != "" && f.Label != "" {
				fieldLabels[f.Id] = f.Label
			}
		}
	}

	desc := formatFormDataDescription(formData, fieldLabels)
	creatorName := "Web User"
	if formData != nil {
		for _, k := range []string{"minecraftUsername", "username", "name", "playerName"} {
			if v, ok := formData[k].(string); ok && strings.TrimSpace(v) != "" {
				creatorName = strings.TrimSpace(v)
				break
			}
		}
	}

	data := make(map[string]any)
	if creatorEmail != "" {
		data["creatorEmail"] = creatorEmail
	}
	if req.CreatorIdentifier != nil && strings.TrimSpace(*req.CreatorIdentifier) != "" {
		data["creatorIdentifier"] = strings.TrimSpace(*req.CreatorIdentifier)
	}

	now := httpx.NowMs()
	var replies []TicketReply
	if desc != "" {
		replies = append(replies, TicketReply{
			ID:                httpx.NewID(),
			Name:              creatorName,
			Content:           desc,
			Type:              "user",
			CreatedMs:         now,
			Staff:             false,
			Attachments:       StructsToAttachmentMaps(req.Attachments),
			CreatorIdentifier: req.GetCreatorIdentifier(),
		})
	}

	emailAuthEnabled := cfg != nil && cfg.RequireEmailAuth
	t := &Ticket{
		ServerID:         srv.ID,
		Type:             cat,
		Subject:          strings.TrimSpace(req.Subject),
		CreatorName:      creatorName,
		Tags:             []string{},
		Replies:          replies,
		Notes:            []TicketNote{},
		FormData:         formData,
		Data:             data,
		Priority:         PriorityNormal,
		AssignedTo:       []string{},
		EmailAuthEnabled: emailAuthEnabled,
		CreatedAtMs:      now,
		UpdatedAtMs:      now,
	}
	t.ApplyLifecycleStatus(StatusOpen)

	saved, err := s.store.InsertWithUniqueID(ctx, cat.TicketPrefix(), t)
	if err != nil {
		return nil, err
	}
	s.InvalidateTickets(srv.ID, saved.ID)
	if saved.EmailAuthEnabled && creatorEmail != "" {
		_, _ = s.RequestVerificationCode(ctx, srv, saved.ID)
	}
	return saved, nil
}

func formatFormDataDescription(formData map[string]any, fieldLabels map[string]string) string {
	if len(formData) == 0 {
		return ""
	}
	keys := make([]string, 0, len(formData))
	for k := range formData {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var lines []string
	for _, k := range keys {
		val := formData[k]
		if val == nil {
			continue
		}
		str := strings.TrimSpace(fmt.Sprintf("%v", val))
		if str == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s: %s", ResolveFieldLabel(k, fieldLabels), str))
	}
	return strings.Join(lines, "\n")
}

func (s *Service) UpdateTicket(ctx context.Context, srv *models.Server, ticketID string, req *modlv1.UpdateTicketRequest, actorEmail string) (*Ticket, error) {
	t, err := s.store.FindByID(ctx, srv.ID, ticketID)
	if err != nil {
		return nil, err
	}
	wasClosed := t.Status.IsTerminal()
	now := httpx.NowMs()

	if req.Status != nil && strings.TrimSpace(*req.Status) != "" {
		st, err := ParseTicketStatus(*req.Status)
		if err != nil {
			return nil, BadRequestError{Msg: "Invalid ticket status"}
		}
		t.ApplyLifecycleStatus(st)
	}
	if req.Locked != nil {
		t.Locked = *req.Locked
		if *req.Locked && !t.Status.IsTerminal() {
			t.Status = StatusClosed
		} else if !*req.Locked && t.Status == StatusClosed {
			t.Status = StatusOpen
		}
	}
	if req.Hidden != nil {
		t.Hidden = *req.Hidden
	}
	if len(req.Tags) > 0 {
		t.Tags = append([]string{}, req.Tags...)
	}
	if req.Data != nil {
		if t.Data == nil {
			t.Data = make(map[string]any)
		}
		for k, v := range req.Data.AsMap() {
			t.Data[k] = v
		}
	}
	if len(req.AssignedTo) > 0 {
		t.AssignedTo = NormalizeAssignees(req.AssignedTo)
	}
	if req.NewReply != nil {
		reply := buildReplyFromProto(req.NewReply, now)
		if reply.Staff && strings.TrimSpace(actorEmail) != "" {
			if u := s.store.FindStaffUsernameByEmail(ctx, srv.ID, actorEmail); u != "" {
				reply.Name = u
			}
		}
		t.Replies = append(t.Replies, reply)
		s.NotifyTicketReply(ctx, srv, t, reply)
		s.RecordSubscriptionUpdates(ctx, srv, t, reply, actorEmail)
	}
	if req.NewNote != nil {
		issuer := strings.TrimSpace(req.NewNote.IssuerName)
		if actor := s.store.FindStaffUsernameByEmail(ctx, srv.ID, actorEmail); actor != "" {
			issuer = actor
		}
		if issuer == "" {
			issuer = "Staff"
		}
		t.Notes = append(t.Notes, TicketNote{
			ID:         httpx.NewID(),
			Text:       req.NewNote.Text,
			IssuerName: issuer,
			Avatar:     req.NewNote.GetIssuerAvatar(),
			DateMs:     now,
		})
	}
	t.UpdatedAtMs = now

	if err := s.store.SaveTicket(ctx, t); err != nil {
		return nil, err
	}
	if !wasClosed && t.Status.IsTerminal() {
		s.NotifyTicketClosed(ctx, srv, t)
	}
	s.InvalidateTickets(srv.ID, t.ID)
	s.store.RecordAuditLog(ctx, srv.ID, "info", s.store.ResolvePanelActor(ctx, srv.ID, actorEmail), "Updated ticket "+t.ID)
	return t, nil
}

func (s *Service) BulkUpdate(ctx context.Context, srv *models.Server, req *modlv1.BulkTicketUpdateRequest, actorEmail string) (int, error) {
	if len(req.TicketIds) == 0 {
		return 0, BadRequestError{Msg: "No ticket IDs provided"}
	}
	tickets, err := s.store.FindByIDs(ctx, srv.ID, req.TicketIds)
	if err != nil {
		return 0, err
	}
	now := httpx.NowMs()
	updated := 0
	for _, t := range tickets {
		wasClosed := t.Status.IsTerminal()
		if req.Locked != nil {
			if *req.Locked {
				t.ApplyLifecycleStatus(StatusClosed)
			} else {
				t.ApplyLifecycleStatus(StatusOpen)
			}
		}
		if req.Hidden != nil {
			t.Hidden = *req.Hidden
		}
		if req.AssignTo != nil {
			norm := NormalizeSingleAssignee(*req.AssignTo)
			if norm == "" {
				t.AssignedTo = []string{}
			} else {
				t.AssignedTo = []string{norm}
			}
		}
		if len(req.AddLabels) > 0 {
			existing := make(map[string]struct{}, len(t.Tags))
			for _, tg := range t.Tags {
				existing[tg] = struct{}{}
			}
			for _, lbl := range req.AddLabels {
				if _, ok := existing[lbl]; !ok && strings.TrimSpace(lbl) != "" {
					t.Tags = append(t.Tags, lbl)
					existing[lbl] = struct{}{}
				}
			}
		}
		if len(req.RemoveLabels) > 0 {
			rem := make(map[string]struct{}, len(req.RemoveLabels))
			for _, lbl := range req.RemoveLabels {
				rem[lbl] = struct{}{}
			}
			kept := make([]string, 0, len(t.Tags))
			for _, tg := range t.Tags {
				if _, skip := rem[tg]; !skip {
					kept = append(kept, tg)
				}
			}
			t.Tags = kept
		}
		t.UpdatedAtMs = now
		if err := s.store.SaveTicket(ctx, t); err == nil {
			updated++
			if !wasClosed && t.Status.IsTerminal() {
				s.NotifyTicketClosed(ctx, srv, t)
			}
		}
	}
	s.InvalidateTickets(srv.ID, "")
	s.store.RecordAuditLog(ctx, srv.ID, "info", s.store.ResolvePanelActor(ctx, srv.ID, actorEmail), fmt.Sprintf("Bulk updated %d ticket(s)", updated))
	return updated, nil
}

func buildReplyFromProto(req *modlv1.AddReplyRequest, now int64) TicketReply {
	typ := strings.TrimSpace(req.GetType())
	if typ == "" {
		if req.Staff {
			typ = "staff"
		} else {
			typ = "user"
		}
	}
	return TicketReply{
		ID:                httpx.NewID(),
		Name:              strings.TrimSpace(req.Name),
		Content:           req.Content,
		Type:              typ,
		CreatedMs:         now,
		Staff:             req.Staff,
		Action:            req.GetAction(),
		Avatar:            req.GetAvatar(),
		Attachments:       StructsToAttachmentMaps(req.Attachments),
		CreatorIdentifier: req.GetCreatorIdentifier(),
	}
}

func (s *Service) AddReply(ctx context.Context, srv *models.Server, ticketID string, req *modlv1.AddReplyRequest, actorEmail string, isPanel bool) (TicketReply, error) {
	t, err := s.store.FindByID(ctx, srv.ID, ticketID)
	if err != nil {
		return TicketReply{}, err
	}
	if !isPanel && t.Hidden {
		return TicketReply{}, ErrNotFound
	}
	if t.Locked {
		return TicketReply{}, ConflictError{Msg: "Ticket is locked and cannot accept new replies"}
	}
	now := httpx.NowMs()
	reply := buildReplyFromProto(req, now)
	if !isPanel {
		reply.Staff = false
		if reply.Type == "" || strings.EqualFold(reply.Type, StaffNoteType) {
			reply.Type = "user"
		}
	} else if reply.Staff && strings.TrimSpace(actorEmail) != "" {
		if u := s.store.FindStaffUsernameByEmail(ctx, srv.ID, actorEmail); u != "" {
			reply.Name = u
		}
	}
	if strings.TrimSpace(reply.Name) == "" {
		if reply.Staff {
			reply.Name = "Staff"
		} else {
			reply.Name = nonEmpty(t.CreatorName, "Player")
		}
	}

	t.Replies = append(t.Replies, reply)
	t.UpdatedAtMs = now
	if err := s.store.SaveTicket(ctx, t); err != nil {
		return TicketReply{}, err
	}
	if isPanel && reply.Staff && strings.TrimSpace(actorEmail) != "" {
		s.EnsureSubscription(ctx, srv.ID, actorEmail, t.ID)
	}
	s.NotifyTicketReply(ctx, srv, t, reply)
	s.RecordSubscriptionUpdates(ctx, srv, t, reply, actorEmail)
	s.InvalidateTickets(srv.ID, t.ID)
	if isPanel {
		s.store.RecordAuditLog(ctx, srv.ID, "info", s.store.ResolvePanelActor(ctx, srv.ID, actorEmail), "Replied to ticket "+t.ID)
	}
	return reply, nil
}

func (s *Service) AddNote(ctx context.Context, srv *models.Server, ticketID string, req *modlv1.AddNoteRequest, actorEmail string) (TicketNote, error) {
	t, err := s.store.FindByID(ctx, srv.ID, ticketID)
	if err != nil {
		return TicketNote{}, err
	}
	now := httpx.NowMs()
	issuer := strings.TrimSpace(req.IssuerName)
	if u := s.store.FindStaffUsernameByEmail(ctx, srv.ID, actorEmail); u != "" {
		issuer = u
	}
	if issuer == "" {
		issuer = "Staff"
	}
	note := TicketNote{
		ID:         httpx.NewID(),
		Text:       req.Text,
		IssuerName: issuer,
		Avatar:     req.GetIssuerAvatar(),
		DateMs:     now,
	}
	t.Notes = append(t.Notes, note)
	t.UpdatedAtMs = now
	if err := s.store.SaveTicket(ctx, t); err != nil {
		return TicketNote{}, err
	}
	s.InvalidateTickets(srv.ID, t.ID)
	return note, nil
}

func (s *Service) AddTag(ctx context.Context, srv *models.Server, ticketID, tag string) ([]string, error) {
	t, err := s.store.FindByID(ctx, srv.ID, ticketID)
	if err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(tag)
	if trimmed == "" {
		return nil, BadRequestError{Msg: "Tag cannot be empty"}
	}
	for _, existing := range t.Tags {
		if existing == trimmed {
			return t.Tags, nil
		}
	}
	t.Tags = append(t.Tags, trimmed)
	t.UpdatedAtMs = httpx.NowMs()
	if err := s.store.SaveTicket(ctx, t); err != nil {
		return nil, err
	}
	s.InvalidateTickets(srv.ID, t.ID)
	return t.Tags, nil
}

func (s *Service) RemoveTag(ctx context.Context, srv *models.Server, ticketID, tag string) ([]string, error) {
	t, err := s.store.FindByID(ctx, srv.ID, ticketID)
	if err != nil {
		return nil, err
	}
	kept := make([]string, 0, len(t.Tags))
	for _, existing := range t.Tags {
		if existing != tag {
			kept = append(kept, existing)
		}
	}
	t.Tags = kept
	t.UpdatedAtMs = httpx.NowMs()
	if err := s.store.SaveTicket(ctx, t); err != nil {
		return nil, err
	}
	s.InvalidateTickets(srv.ID, t.ID)
	return t.Tags, nil
}

func (s *Service) ProcessQuickResponse(ctx context.Context, srv *models.Server, ticketID string, req *modlv1.QuickResponseRequest, actorEmail string) (*modlv1.QuickResponseResult, error) {
	t, err := s.store.FindByID(ctx, srv.ID, ticketID)
	if err != nil {
		return nil, err
	}
	if t.Locked {
		return &modlv1.QuickResponseResult{
			Success:  false,
			Message:  "Ticket is locked",
			TicketId: t.ID,
		}, nil
	}

	qr := s.store.GetQuickResponses(ctx, srv.ID)
	var foundAction *modlv1.QuickResponseAction
	for _, cat := range qr.Categories {
		if cat != nil && cat.Id == req.CategoryId {
			for _, act := range cat.Actions {
				if act != nil && act.Id == req.ActionId {
					foundAction = act
					break
				}
			}
		}
	}
	if foundAction == nil {
		return &modlv1.QuickResponseResult{
			Success:  false,
			Message:  "Quick response action not found",
			TicketId: t.ID,
		}, nil
	}

	staffName := s.store.ResolvePanelActor(ctx, srv.ID, actorEmail)
	if staffName == "Unknown" {
		staffName = "Staff"
	}
	now := httpx.NowMs()
	reply := TicketReply{
		ID:        httpx.NewID(),
		Name:      staffName,
		Content:   foundAction.Message,
		Type:      "staff",
		CreatedMs: now,
		Staff:     true,
		Action:    foundAction.Name,
	}
	t.Replies = append(t.Replies, reply)

	closed := false
	if foundAction.CloseTicket != nil && *foundAction.CloseTicket {
		t.ApplyLifecycleStatus(StatusClosed)
		closed = true
	}
	t.UpdatedAtMs = now
	if err := s.store.SaveTicket(ctx, t); err != nil {
		return nil, err
	}

	s.NotifyTicketReply(ctx, srv, t, reply)
	if closed {
		s.NotifyTicketClosed(ctx, srv, t)
	}
	s.RecordSubscriptionUpdates(ctx, srv, t, reply, actorEmail)
	s.InvalidateTickets(srv.ID, t.ID)

	return &modlv1.QuickResponseResult{
		Success:          true,
		Message:          "Quick response applied",
		TicketId:         t.ID,
		ActionName:       foundAction.Name,
		TicketClosed:     closed,
		PunishmentIssued: false,
		AppealOutcome:    foundAction.AppealAction,
	}, nil
}

// --- Minecraft Ticket & Report Service Methods ---

func (s *Service) CreateMinecraftTicket(ctx context.Context, srv *models.Server, req MinecraftCreateTicketRequestJSON, forceUnfinished bool) (*Ticket, error) {
	if !IsValidUUID(req.CreatorUUID) {
		return nil, BadRequestError{Msg: "Invalid creatorUuid"}
	}
	if strings.TrimSpace(req.ReportedPlayerUUID) != "" && !IsValidUUID(req.ReportedPlayerUUID) {
		return nil, BadRequestError{Msg: "Invalid reportedPlayerUuid"}
	}
	cat, err := ParseTicketCategory(req.Type)
	if err != nil {
		return nil, BadRequestError{Msg: "Invalid ticket type"}
	}
	prio, err := ParseTicketPriority(req.Priority)
	if err != nil {
		return nil, BadRequestError{Msg: "Invalid priority"}
	}

	now := httpx.NowMs()
	creatorName := strings.TrimSpace(req.CreatorName)
	if creatorName == "" {
		creatorName = "Player"
	}
	subject := strings.TrimSpace(req.Subject)
	if subject == "" {
		subject = cat.DisplayName()
	}

	var replies []TicketReply
	if strings.TrimSpace(req.Description) != "" {
		replies = append(replies, TicketReply{
			ID:        httpx.NewID(),
			Name:      creatorName,
			Content:   strings.TrimSpace(req.Description),
			Type:      "user",
			CreatedMs: now,
			Staff:     false,
		})
	}

	var chatMsgs []ChatMessage
	for _, raw := range req.ChatMessages {
		chatMsgs = append(chatMsgs, parseChatLogLine(raw, now))
	}

	data := make(map[string]any)
	if strings.TrimSpace(req.CreatedServer) != "" {
		data["createdServer"] = strings.TrimSpace(req.CreatedServer)
	}

	tags := make([]string, 0, len(req.Tags))
	for _, tg := range req.Tags {
		if strings.TrimSpace(tg) != "" {
			tags = append(tags, strings.TrimSpace(tg))
		}
	}

	t := &Ticket{
		ServerID:           srv.ID,
		Type:               cat,
		Subject:            subject,
		CreatorUUID:        NormalizeUUID(req.CreatorUUID),
		CreatorName:        creatorName,
		ReportedPlayer:     strings.TrimSpace(req.ReportedPlayerName),
		ReportedPlayerUUID: NormalizeUUID(req.ReportedPlayerUUID),
		Tags:               tags,
		Replies:            replies,
		Notes:              []TicketNote{},
		ChatMessages:       chatMsgs,
		Data:               data,
		Priority:           prio,
		AssignedTo:         []string{},
		CreatedAtMs:        now,
		UpdatedAtMs:        now,
	}
	if forceUnfinished {
		t.ApplyLifecycleStatus(StatusUnfinished)
	} else {
		t.ApplyLifecycleStatus(StatusOpen)
	}

	saved, err := s.store.InsertWithUniqueID(ctx, cat.TicketPrefix(), t)
	if err != nil {
		return nil, err
	}
	if !forceUnfinished {
		s.InvalidateTickets(srv.ID, saved.ID)
	}

	if !forceUnfinished && cat == CategoryChat && len(chatMsgs) > 0 && s.aiClient != nil && s.aiClient.Enabled() {
		go s.analyzeChatReportAsync(srv.ID, saved.ID)
	}
	return saved, nil
}

func parseChatLogLine(line string, fallbackMs int64) ChatMessage {
	// Support optional "[HH:MM:SS] <sender>: message" or plain string
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "[") {
		if idx := strings.Index(trimmed, "] "); idx > 1 {
			rest := strings.TrimSpace(trimmed[idx+2:])
			if colon := strings.Index(rest, ": "); colon > 0 {
				return ChatMessage{
					Sender:      strings.TrimSpace(rest[:colon]),
					Content:     strings.TrimSpace(rest[colon+2:]),
					TimestampMs: fallbackMs,
				}
			}
			return ChatMessage{Content: rest, TimestampMs: fallbackMs}
		}
	}
	return ChatMessage{Content: trimmed, TimestampMs: fallbackMs}
}

func (s *Service) analyzeChatReportAsync(serverID, ticketID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	t, err := s.store.FindByID(ctx, serverID, ticketID)
	if err != nil || t.Type != CategoryChat || len(t.ChatMessages) == 0 {
		return
	}
	aiSettings := &modlv1.AIModerationSettings{}
	if !s.store.LoadSettingProto(ctx, serverID, "aiModeration", aiSettings) || !aiSettings.EnableAiReview {
		return
	}
	var lines []string
	for _, cm := range t.ChatMessages {
		if cm.Sender != "" {
			lines = append(lines, fmt.Sprintf("%s: %s", cm.Sender, cm.Content))
		} else {
			lines = append(lines, cm.Content)
		}
	}
	systemPrompt := "You are a Minecraft chat moderation assistant. Analyze the reported chat log and summarize whether a rule violation occurred in JSON: {\"analysis\":\"...\",\"confidence\":0.0}."
	userPrompt := fmt.Sprintf("Reported player: %s\nChat logs:\n%s", nonEmpty(t.ReportedPlayer, t.ReportedPlayerUUID), strings.Join(lines, "\n"))
	raw, err := s.aiClient.Generate(ctx, systemPrompt, userPrompt)
	if err != nil {
		log.Printf("ai analysis failed for ticket %s: %v", ticketID, err)
		return
	}
	res := &AIAnalysisResult{
		Analysis:     raw,
		AnalyzedAtMs: httpx.NowMs(),
		RawResponse:  raw,
	}
	var parsed struct {
		Analysis   string   `json:"analysis"`
		Confidence *float64 `json:"confidence"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &parsed); err == nil && parsed.Analysis != "" {
		res.Analysis = parsed.Analysis
		res.Confidence = parsed.Confidence
	}
	t.AIAnalysis = res
	t.UpdatedAtMs = httpx.NowMs()
	_ = s.store.SaveTicket(ctx, t)
	s.InvalidateTickets(serverID, ticketID)
}

func (s *Service) ClaimTicket(ctx context.Context, srv *models.Server, ticketID, playerUUID, playerName string) (*Ticket, error) {
	t, err := s.store.FindByID(ctx, srv.ID, ticketID)
	if err != nil {
		return nil, err
	}
	if !IsValidUUID(playerUUID) {
		return nil, BadRequestError{Msg: "Invalid playerUuid"}
	}
	if !IsValidMinecraftUsername(playerName) {
		return nil, BadRequestError{Msg: "Invalid playerName"}
	}
	claimerUUID := NormalizeUUID(playerUUID)
	if strings.TrimSpace(t.CreatorUUID) != "" {
		if NormalizeUUID(t.CreatorUUID) == claimerUUID {
			return t, nil
		}
		return nil, ConflictError{Msg: "Ticket is already linked to a Minecraft account"}
	}

	oldCreatorName := t.CreatorName
	origIdent := ""
	if t.Data != nil {
		if v, ok := t.Data["creatorIdentifier"].(string); ok {
			origIdent = strings.TrimSpace(v)
		}
	}
	if origIdent == "" {
		for _, r := range t.Replies {
			if !r.Staff && strings.TrimSpace(r.CreatorIdentifier) != "" {
				origIdent = strings.TrimSpace(r.CreatorIdentifier)
				break
			}
		}
	}

	t.CreatorUUID = claimerUUID
	t.CreatorName = strings.TrimSpace(playerName)
	t.UpdatedAtMs = httpx.NowMs()

	for i := range t.Replies {
		if t.Replies[i].Staff {
			continue
		}
		owned := false
		if strings.TrimSpace(t.Replies[i].CreatorIdentifier) != "" {
			owned = origIdent != "" && t.Replies[i].CreatorIdentifier == origIdent
		} else {
			owned = oldCreatorName != "" && oldCreatorName == t.Replies[i].Name
		}
		if owned {
			t.Replies[i].Name = t.CreatorName
			t.Replies[i].CreatorIdentifier = claimerUUID
		}
	}

	if err := s.store.SaveTicket(ctx, t); err != nil {
		return nil, err
	}
	s.InvalidateTickets(srv.ID, t.ID)
	return t, nil
}

func (s *Service) AssignReport(ctx context.Context, srv *models.Server, ticketID, assignee string) (*Ticket, error) {
	t, err := s.store.FindByID(ctx, srv.ID, ticketID)
	if err != nil {
		return nil, err
	}
	norm := NormalizeSingleAssignee(assignee)
	if norm == "" {
		return nil, BadRequestError{Msg: "Assignee cannot be empty"}
	}
	t.AssignedTo = []string{norm}
	t.UpdatedAtMs = httpx.NowMs()
	if err := s.store.SaveTicket(ctx, t); err != nil {
		return nil, err
	}
	s.InvalidateTickets(srv.ID, t.ID)
	return t, nil
}

func (s *Service) DismissReport(ctx context.Context, srv *models.Server, ticketID, dismissedBy, reason string) (*Ticket, error) {
	return s.closeReport(ctx, srv, ticketID, dismissedBy,
		"Thank you for submitting this report. After careful review, we have found insufficient evidence to take action at this time.",
		func(t *Ticket, now int64) {
			if t.Data == nil {
				t.Data = make(map[string]any)
			}
			if strings.TrimSpace(reason) != "" {
				t.Data["dismissReason"] = strings.TrimSpace(reason)
			}
			if strings.TrimSpace(dismissedBy) != "" {
				t.Data["dismissedBy"] = strings.TrimSpace(dismissedBy)
				t.Data["dismissedAt"] = now
			}
		})
}

func (s *Service) ResolveReport(ctx context.Context, srv *models.Server, ticketID, resolvedBy, resolution, punishmentID string) (*Ticket, error) {
	return s.closeReport(ctx, srv, ticketID, resolvedBy,
		"Thank you for creating this report. After careful review, we have accepted this and the reported player has received a punishment.",
		func(t *Ticket, now int64) {
			if t.Data == nil {
				t.Data = make(map[string]any)
			}
			if strings.TrimSpace(resolution) != "" {
				t.Data["resolution"] = strings.TrimSpace(resolution)
			}
			if strings.TrimSpace(resolvedBy) != "" {
				t.Data["resolvedBy"] = strings.TrimSpace(resolvedBy)
				t.Data["resolvedAt"] = now
			}
			if strings.TrimSpace(punishmentID) != "" {
				t.Data["linkedPunishmentId"] = strings.TrimSpace(punishmentID)
			}
		})
}

func (s *Service) closeReport(ctx context.Context, srv *models.Server, ticketID, staffName, closeMsg string, applyMeta func(*Ticket, int64)) (*Ticket, error) {
	t, err := s.store.FindByID(ctx, srv.ID, ticketID)
	if err != nil {
		return nil, err
	}
	if t.Locked || t.Status.IsTerminal() {
		return t, nil
	}
	now := httpx.NowMs()
	reply := TicketReply{
		ID:        httpx.NewID(),
		Name:      nonEmpty(staffName, "Staff"),
		Content:   closeMsg,
		Type:      "reply",
		CreatedMs: now,
		Staff:     true,
		Action:    "close",
	}
	t.Replies = append(t.Replies, reply)
	t.ApplyLifecycleStatus(StatusClosed)
	t.UpdatedAtMs = now
	if applyMeta != nil {
		applyMeta(t, now)
	}
	if err := s.store.SaveTicket(ctx, t); err != nil {
		return nil, err
	}
	s.NotifyTicketReply(ctx, srv, t, reply)
	s.NotifyTicketClosed(ctx, srv, t)
	s.InvalidateTickets(srv.ID, t.ID)
	return t, nil
}

func nonEmpty(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return strings.TrimSpace(v)
}

func ToStructPB(m map[string]any) *structpb.Struct {
	return MapToStruct(m)
}
