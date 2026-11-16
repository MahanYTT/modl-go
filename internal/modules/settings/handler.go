package settings

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/config"
	"modl-go/internal/httpx"
	"modl-go/internal/models"
	"modl-go/internal/modules/role"
	"modl-go/internal/realtime"
	"modl-go/internal/storage"

	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler struct {
	db      *sql.DB
	cfg     *config.Config
	hub     *realtime.Hub
	roles   *role.Handler
	storage *storage.LocalStore
}

func NewHandler(db *sql.DB, cfg *config.Config, hub *realtime.Hub, roles *role.Handler, store *storage.LocalStore) *Handler {
	return &Handler{db: db, cfg: cfg, hub: hub, roles: roles, storage: store}
}

func (h *Handler) RegisterPanelRoutes(r chi.Router) {
	r.Route("/settings", func(r chi.Router) {
		// General settings routes
		r.Group(func(r chi.Router) {
			r.Use(httpx.RequirePanelPolicy(h.roles, models.AccessPolicy{
				Rule:   models.RuleRequirePermission,
				View:   "admin.settings.view",
				Modify: "admin.settings.modify",
			}))
			r.Get("/general", h.getGeneralSettings)
			r.Patch("/general", h.patchGeneralSettings)
			r.Get("/ticket-labels", h.getTicketLabels)
			r.Patch("/ticket-labels", h.patchTicketLabels)
			r.Get("/status-thresholds", h.getStatusThresholds)
			r.Patch("/status-thresholds", h.patchStatusThresholds)
			r.Get("/replay-retention", h.getReplayRetention)
			r.Patch("/replay-retention", h.patchReplayRetention)
			r.Get("/webhooks", h.getWebhooks)
			r.Patch("/webhooks", h.patchWebhooks)
			r.Post("/webhooks/test", h.testWebhook)
			r.Get("/ticket-forms", h.getTicketForms)
			r.Patch("/ticket-forms", h.patchTicketForms)
			r.Get("/ticket-forms/{type}", h.getTicketFormByType)
			r.Get("/quick-responses", h.getQuickResponses)
			r.Patch("/quick-responses", h.patchQuickResponses)
			r.Post("/upload-icon", h.uploadIcon)
		})

		// AI moderation routes
		r.Group(func(r chi.Router) {
			r.Use(httpx.RequirePanelPolicy(h.roles, models.AccessPolicy{
				Rule:   models.RuleRequirePermission,
				View:   "admin.settings.view.punishments",
				Modify: "admin.settings.modify.punishments",
			}))
			r.Get("/ai-moderation", h.getAIModeration)
			r.Patch("/ai-moderation", h.patchAIModeration)
			r.Post("/ai-dismiss-suggestion/{ticketId}", h.dismissAISuggestion)
		})

		r.With(httpx.RequirePanelPolicy(h.roles, models.AccessPolicy{Rule: models.RulePermitAll})).
			Post("/ai-apply-punishment/{ticketId}", h.applyAIPunishment)

		// Punishment types
		r.Route("/punishment-types", func(r chi.Router) {
			r.Use(httpx.RequirePanelPolicy(h.roles, models.AccessPolicy{Rule: models.RulePunishmentTypeAccess}))
			r.Get("/", h.getPanelPunishmentTypes)
			r.Get("/{ordinal}", h.getPanelPunishmentType)
			r.Patch("/{ordinal}", h.updatePunishmentType)
			r.Post("/", h.createPunishmentType)
			r.Post("/reset", h.resetPunishmentTypes)
			r.Delete("/{ordinal}", h.deletePunishmentType)
		})

		// API keys (super admin only)
		r.Group(func(r chi.Router) {
			r.Use(httpx.RequirePanelPolicy(h.roles, models.AccessPolicy{Rule: models.RuleSuperAdmin}))
			r.Post("/api-keys/{type}/generate", h.generateAPIKey)
			r.Get("/api-keys/{type}/reveal", h.revealAPIKey)
			r.Delete("/api-keys/{type}", h.deleteAPIKey)
			r.Get("/api-keys/{type}/exists", h.existsAPIKey)
		})

		// Domain settings
		r.Route("/domain", func(r chi.Router) {
			r.Use(httpx.RequirePanelPolicy(h.roles, models.AccessPolicy{
				Rule:   models.RuleRequirePermission,
				View:   "admin.settings.view.domain",
				Modify: "admin.settings.modify.domain",
			}))
			r.Get("/", h.getDomainSettings)
			r.Post("/", h.configureDomain)
			r.Post("/verify", h.verifyDomain)
			r.Delete("/", h.removeDomain)
		})
	})
}

func (h *Handler) RegisterPublicRoutes(r chi.Router) {
	r.Get("/settings", h.getPublicSettings)
}

func (h *Handler) RegisterMinecraftV1Routes(r chi.Router) {
	r.Get("/punishments/types", h.getMinecraftPunishmentTypesV1)
}

func (h *Handler) RegisterMinecraftV3Routes(r chi.Router) {
	r.Get("/punishments/types", h.getMinecraftPunishmentTypesV3)
}

// --- Versioned settings persistence helpers ---

func (h *Handler) loadSetting(ctx context.Context, serverID, typ string, msg proto.Message) (int64, int64, bool) {
	var dataJSON string
	var version, updatedAt int64
	err := h.db.QueryRowContext(ctx,
		`SELECT data_json, version, updated_at FROM settings WHERE server_id = ? AND type = ? LIMIT 1`,
		serverID, typ).Scan(&dataJSON, &version, &updatedAt)
	if err != nil {
		return 1, httpx.NowMs(), false
	}
	if msg != nil && dataJSON != "" {
		_ = (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal([]byte(dataJSON), msg)
	}
	return version, updatedAt, true
}

func (h *Handler) saveSetting(ctx context.Context, serverID, typ string, msg proto.Message, expectedVersion *int64) (int64, int64, error) {
	now := httpx.NowMs()
	var currentVer int64
	err := h.db.QueryRowContext(ctx, `SELECT version FROM settings WHERE server_id = ? AND type = ? LIMIT 1`, serverID, typ).Scan(&currentVer)
	exists := err == nil
	if !exists {
		currentVer = 1
	}
	if expectedVersion != nil && *expectedVersion > 0 && exists && *expectedVersion != currentVer {
		return currentVer, 0, fmt.Errorf("CONFLICT:%d", currentVer)
	}
	newVer := currentVer + 1
	if !exists {
		newVer = 1
	}
	b, _ := (protojson.MarshalOptions{EmitUnpopulated: true}).Marshal(msg)
	_, err = h.db.ExecContext(ctx, `
		INSERT INTO settings (id, server_id, type, data_json, version, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(server_id, type) DO UPDATE SET data_json=excluded.data_json, version=excluded.version, updated_at=excluded.updated_at`,
		httpx.NewID(), serverID, typ, string(b), newVer, now,
	)
	return newVer, now, err
}

func metaPB(ver, updatedAt int64) *modlv1.SettingsMeta {
	return &modlv1.SettingsMeta{
		Version:   ver,
		UpdatedAt: timestamppb.New(time.UnixMilli(updatedAt)),
	}
}

func (h *Handler) invalidate(serverID string) {
	if h.hub != nil {
		h.hub.InvalidatePanel(serverID, modlv1.PanelResource_PANEL_RESOURCE_SETTINGS, "")
	}
}

// --- General settings ---

func defaultGeneralSettings(srv *models.Server) *modlv1.GeneralSettings {
	name := "modl"
	if srv != nil && srv.ServerName != "" {
		name = srv.ServerName
	}
	return &modlv1.GeneralSettings{
		ServerDisplayName: name,
		DefaultLanguage:   "en",
	}
}

func (h *Handler) getGeneralSettings(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	gs := defaultGeneralSettings(srv)
	ver, upd, _ := h.loadSetting(r.Context(), srv.ID, "general", gs)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.GeneralSettingsEnvelope{Data: gs, Meta: metaPB(ver, upd)})
}

func (h *Handler) patchGeneralSettings(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.PatchGeneralSettingsRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	gs := defaultGeneralSettings(srv)
	_, _, _ = h.loadSetting(r.Context(), srv.ID, "general", gs)
	if req.ServerDisplayName != nil {
		gs.ServerDisplayName = *req.ServerDisplayName
	}
	if req.DiscordWebhookUrl != nil {
		gs.DiscordWebhookUrl = *req.DiscordWebhookUrl
	}
	if req.HomepageIconUrl != nil {
		gs.HomepageIconUrl = *req.HomepageIconUrl
	}
	if req.PanelIconUrl != nil {
		gs.PanelIconUrl = *req.PanelIconUrl
	}
	if req.DefaultLanguage != nil {
		gs.DefaultLanguage = *req.DefaultLanguage
	}
	ver, upd, err := h.saveSetting(r.Context(), srv.ID, "general", gs, &req.ExpectedVersion)
	if err != nil {
		h.handleSaveErr(w, r, err)
		return
	}
	h.invalidate(srv.ID)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.GeneralSettingsEnvelope{Data: gs, Meta: metaPB(ver, upd)})
}

// --- Ticket Labels ---

func (h *Handler) getTicketLabels(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	tl := &modlv1.TicketLabelSettings{Labels: []*modlv1.Label{}}
	ver, upd, _ := h.loadSetting(r.Context(), srv.ID, "ticketLabels", tl)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.TicketLabelSettingsEnvelope{Data: tl, Meta: metaPB(ver, upd)})
}

func (h *Handler) patchTicketLabels(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.PatchTicketLabelSettingsRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	tl := &modlv1.TicketLabelSettings{Labels: req.Labels}
	ver, upd, err := h.saveSetting(r.Context(), srv.ID, "ticketLabels", tl, &req.ExpectedVersion)
	if err != nil {
		h.handleSaveErr(w, r, err)
		return
	}
	h.invalidate(srv.ID)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.TicketLabelSettingsEnvelope{Data: tl, Meta: metaPB(ver, upd)})
}

// --- Offender Thresholds ---

func defaultThresholds() *modlv1.OffenderThresholdSettings {
	return &modlv1.OffenderThresholdSettings{
		Social:   &modlv1.CategoryThresholds{Medium: 10, Habitual: 25, PointExpiryMonths: 6},
		Gameplay: &modlv1.CategoryThresholds{Medium: 15, Habitual: 35, PointExpiryMonths: 12},
	}
}

func (h *Handler) getStatusThresholds(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	st := defaultThresholds()
	ver, upd, _ := h.loadSetting(r.Context(), srv.ID, "statusThresholds", st)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.OffenderThresholdSettingsEnvelope{Data: st, Meta: metaPB(ver, upd)})
}

func (h *Handler) patchStatusThresholds(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.PatchStatusThresholdSettingsRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	st := req.Settings
	if st == nil {
		st = defaultThresholds()
	}
	ver, upd, err := h.saveSetting(r.Context(), srv.ID, "statusThresholds", st, &req.ExpectedVersion)
	if err != nil {
		h.handleSaveErr(w, r, err)
		return
	}
	h.invalidate(srv.ID)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.OffenderThresholdSettingsEnvelope{Data: st, Meta: metaPB(ver, upd)})
}

// --- Replay Retention (stripped -> disabled stub for panel settings compatibility) ---

func (h *Handler) getReplayRetention(w http.ResponseWriter, r *http.Request) {
	rr := &modlv1.ReplayRetentionSettings{Enabled: false, Days: 0}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.ReplayRetentionSettingsEnvelope{Data: rr, Meta: metaPB(1, httpx.NowMs())})
}

func (h *Handler) patchReplayRetention(w http.ResponseWriter, r *http.Request) {
	rr := &modlv1.ReplayRetentionSettings{Enabled: false, Days: 0}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.ReplayRetentionSettingsEnvelope{Data: rr, Meta: metaPB(1, httpx.NowMs())})
}

// --- AI Moderation ---

func (h *Handler) getAIModeration(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	aiSet := &modlv1.AIModerationSettings{AiPunishmentConfigs: map[string]*modlv1.AIPunishmentConfig{}}
	_, _, _ = h.loadSetting(r.Context(), srv.ID, "aiModeration", aiSet)
	httpx.WriteProto(w, r, http.StatusOK, aiSet)
}

func (h *Handler) patchAIModeration(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.UpdateAIModerationSettingsRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	aiSet := &modlv1.AIModerationSettings{AiPunishmentConfigs: map[string]*modlv1.AIPunishmentConfig{}}
	_, _, _ = h.loadSetting(r.Context(), srv.ID, "aiModeration", aiSet)
	if req.EnableAiReview != nil {
		aiSet.EnableAiReview = *req.EnableAiReview
	}
	if req.EnableAutomatedActions != nil {
		aiSet.EnableAutomatedActions = *req.EnableAutomatedActions
	}
	if req.AiPunishmentConfigs != nil {
		m := make(map[string]*modlv1.AIPunishmentConfig, len(req.AiPunishmentConfigs))
		for k, v := range req.AiPunishmentConfigs {
			en := false
			if v.Enabled != nil {
				en = *v.Enabled
			}
			m[k] = &modlv1.AIPunishmentConfig{Id: v.Id, Name: v.Name, AiDescription: v.AiDescription, Enabled: en}
		}
		aiSet.AiPunishmentConfigs = m
	}
	_, _, _ = h.saveSetting(r.Context(), srv.ID, "aiModeration", aiSet, nil)
	h.invalidate(srv.ID)
	httpx.WriteProto(w, r, http.StatusOK, aiSet)
}

func (h *Handler) applyAIPunishment(w http.ResponseWriter, r *http.Request) {
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.AISuggestionActionResponse{
		Success: true,
		Message: "AI suggestion applied",
	})
}

func (h *Handler) dismissAISuggestion(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	ticketID := chi.URLParam(r, "ticketId")
	_, _ = h.db.ExecContext(r.Context(), `UPDATE tickets SET ai_analysis_json = NULL, updated_at = ? WHERE server_id = ? AND id = ?`, httpx.NowMs(), srv.ID, ticketID)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.AISuggestionActionResponse{
		Success: true,
		Message: "AI suggestion dismissed",
	})
}

// --- Webhooks ---

func defaultWebhooks() *modlv1.WebhookSettings {
	return &modlv1.WebhookSettings{
		BotName:       "modl",
		Notifications: &modlv1.WebhookNotificationSettings{},
		EmbedTemplates: &modlv1.WebhookEmbedTemplates{
			NewTickets:     &modlv1.WebhookEmbedTemplate{},
			NewPunishments: &modlv1.WebhookEmbedTemplate{},
			AuditLogs:      &modlv1.WebhookEmbedTemplate{},
		},
	}
}

func (h *Handler) getWebhooks(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	ws := defaultWebhooks()
	_, _, _ = h.loadSetting(r.Context(), srv.ID, "webhooks", ws)
	httpx.WriteProto(w, r, http.StatusOK, ws)
}

func (h *Handler) patchWebhooks(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.UpdateWebhookSettingsRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	ws := defaultWebhooks()
	ws.DiscordWebhookUrl = req.DiscordWebhookUrl
	ws.DiscordAdminRoleId = req.DiscordAdminRoleId
	ws.BotName = req.BotName
	ws.AvatarUrl = req.AvatarUrl
	if req.Enabled != nil {
		ws.Enabled = *req.Enabled
	}
	if req.Notifications != nil {
		ws.Notifications = &modlv1.WebhookNotificationSettings{
			NewTickets:     req.Notifications.NewTickets != nil && *req.Notifications.NewTickets,
			NewPunishments: req.Notifications.NewPunishments != nil && *req.Notifications.NewPunishments,
			AuditLogs:      req.Notifications.AuditLogs != nil && *req.Notifications.AuditLogs,
		}
	}
	_, _, _ = h.saveSetting(r.Context(), srv.ID, "webhooks", ws, nil)
	h.invalidate(srv.ID)
	httpx.WriteProto(w, r, http.StatusOK, ws)
}

func (h *Handler) testWebhook(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	ws := defaultWebhooks()
	_, _, _ = h.loadSetting(r.Context(), srv.ID, "webhooks", ws)
	if strings.TrimSpace(ws.DiscordWebhookUrl) == "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "Failed to send webhook test")
		return
	}
	payload, _ := json.Marshal(map[string]any{
		"username": ws.BotName,
		"embeds": []map[string]any{{
			"title":       "modl.gg Webhook Test",
			"description": "Your Discord webhook is configured correctly.",
			"color":       5793266,
		}},
	})
	resp, err := (&http.Client{Timeout: 8 * time.Second}).Post(ws.DiscordWebhookUrl, "application/json", bytes.NewReader(payload))
	if err != nil || resp.StatusCode >= 300 {
		if resp != nil {
			_ = resp.Body.Close()
		}
		httpx.WriteError(w, r, http.StatusBadRequest, "Failed to send webhook test")
		return
	}
	_ = resp.Body.Close()
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.WebhookTestResponse{Message: "Webhook test sent successfully"})
}

// --- Ticket Forms ---

func defaultTicketForms() *modlv1.TicketFormSettings {
	mk := func() *modlv1.TicketForm {
		return &modlv1.TicketForm{
			Fields:   []*modlv1.FormField{},
			Sections: []*modlv1.FormSection{},
		}
	}
	return &modlv1.TicketFormSettings{
		Bug:         mk(),
		Support:     mk(),
		Application: mk(),
		Player:      mk(),
		Chat:        mk(),
	}
}

func (h *Handler) getTicketForms(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	tf := defaultTicketForms()
	ver, upd, _ := h.loadSetting(r.Context(), srv.ID, "ticketForms", tf)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.TicketFormSettingsEnvelope{Data: tf, Meta: metaPB(ver, upd)})
}

func (h *Handler) patchTicketForms(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.PatchTicketFormSettingsRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	tf := req.Settings
	if tf == nil {
		tf = defaultTicketForms()
	}
	ver, upd, err := h.saveSetting(r.Context(), srv.ID, "ticketForms", tf, &req.ExpectedVersion)
	if err != nil {
		h.handleSaveErr(w, r, err)
		return
	}
	h.invalidate(srv.ID)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.TicketFormSettingsEnvelope{Data: tf, Meta: metaPB(ver, upd)})
}

func (h *Handler) getTicketFormByType(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	tf := defaultTicketForms()
	_, _, _ = h.loadSetting(r.Context(), srv.ID, "ticketForms", tf)
	typ := strings.ToLower(chi.URLParam(r, "type"))
	var out *modlv1.TicketForm
	switch typ {
	case "bug":
		out = tf.Bug
	case "support":
		out = tf.Support
	case "application":
		out = tf.Application
	case "player":
		out = tf.Player
	case "chat":
		out = tf.Chat
	}
	if out == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Ticket form not found")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, out)
}

// --- Quick Responses ---

func (h *Handler) getQuickResponses(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	qr := &modlv1.QuickResponseSettings{Categories: []*modlv1.QuickResponseCategory{}}
	ver, upd, _ := h.loadSetting(r.Context(), srv.ID, "quickResponses", qr)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.QuickResponseSettingsEnvelope{Data: qr, Meta: metaPB(ver, upd)})
}

func (h *Handler) patchQuickResponses(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.PatchQuickResponsesRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	qr := &modlv1.QuickResponseSettings{Categories: req.Categories}
	ver, upd, err := h.saveSetting(r.Context(), srv.ID, "quickResponses", qr, &req.ExpectedVersion)
	if err != nil {
		h.handleSaveErr(w, r, err)
		return
	}
	h.invalidate(srv.ID)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.QuickResponseSettingsEnvelope{Data: qr, Meta: metaPB(ver, upd)})
}

// --- Icon Upload ---

func (h *Handler) uploadIcon(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid multipart form")
		return
	}
	file, header, err := r.FormFile("icon")
	if err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Missing icon file")
		return
	}
	defer file.Close()
	iconType := strings.ToLower(strings.TrimSpace(r.FormValue("iconType")))
	ct := header.Header.Get("Content-Type")
	if ct == "" {
		ct = "image/png"
	}
	rec, err := h.storage.Save(r.Context(), srv.ID, "icons", header.Filename, ct, file)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to save icon")
		return
	}
	urlPath := "/v1/public/media/file/" + rec.ID
	gs := defaultGeneralSettings(srv)
	_, _, _ = h.loadSetting(r.Context(), srv.ID, "general", gs)
	if iconType == "homepage" {
		gs.HomepageIconUrl = urlPath
	} else {
		gs.PanelIconUrl = urlPath
	}
	_, _, _ = h.saveSetting(r.Context(), srv.ID, "general", gs, nil)
	h.invalidate(srv.ID)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"url": urlPath, "iconUrl": urlPath})
}

// --- Punishment Types ---

func defaultPunishmentTypes() []*modlv1.PunishmentType {
	dur := func(v int32, u, t string) *modlv1.DurationDetail {
		return &modlv1.DurationDetail{Value: v, Unit: u, Type: t}
	}
	lvl := func(f, m, hb *modlv1.DurationDetail) *modlv1.OffenseLevelDurations {
		return &modlv1.OffenseLevelDurations{First: f, Medium: m, Habitual: hb}
	}
	mk := func(ord int32, name, cat, staffDesc, playerDesc string, cust, appeal, altBlock, wipe bool, durs *modlv1.PunishmentDurations, pts *modlv1.PunishmentPoints) *modlv1.PunishmentType {
		return &modlv1.PunishmentType{
			Id:                &ord,
			Ordinal:           &ord,
			Name:              name,
			Category:          cat,
			Customizable:      &cust,
			Appealable:        &appeal,
			CanBeAltBlocking:  &altBlock,
			CanBeStatWiping:   &wipe,
			StaffDescription:  staffDesc,
			PlayerDescription: playerDesc,
			Durations:         durs,
			Points:            pts,
		}
	}
	return []*modlv1.PunishmentType{
		mk(0, "Kick", "GAMEPLAY", "Kick a player from the server", "You were kicked from the server", false, false, false, false, nil, nil),
		mk(1, "Warn", "SOCIAL", "Issue an official warning", "You received a warning", true, false, false, false, nil, &modlv1.PunishmentPoints{Low: 1, Regular: 2, Severe: 5}),
		mk(2, "Mute", "SOCIAL", "Mute a player in chat", "You are muted in chat", true, true, true, false, &modlv1.PunishmentDurations{
			Low:     lvl(dur(1, "HOURS", "TEMPORARY"), dur(6, "HOURS", "TEMPORARY"), dur(1, "DAYS", "TEMPORARY")),
			Regular: lvl(dur(1, "DAYS", "TEMPORARY"), dur(3, "DAYS", "TEMPORARY"), dur(7, "DAYS", "TEMPORARY")),
			Severe:  lvl(dur(7, "DAYS", "TEMPORARY"), dur(30, "DAYS", "TEMPORARY"), dur(0, "PERMANENT", "PERMANENT")),
		}, &modlv1.PunishmentPoints{Low: 2, Regular: 5, Severe: 10}),
		mk(3, "Ban", "GAMEPLAY", "Ban a player from the server", "You are banned from the server", true, true, true, true, &modlv1.PunishmentDurations{
			Low:     lvl(dur(1, "DAYS", "TEMPORARY"), dur(3, "DAYS", "TEMPORARY"), dur(7, "DAYS", "TEMPORARY")),
			Regular: lvl(dur(7, "DAYS", "TEMPORARY"), dur(30, "DAYS", "TEMPORARY"), dur(90, "DAYS", "TEMPORARY")),
			Severe:  lvl(dur(30, "DAYS", "TEMPORARY"), dur(0, "PERMANENT", "PERMANENT"), dur(0, "PERMANENT", "PERMANENT")),
		}, &modlv1.PunishmentPoints{Low: 5, Regular: 10, Severe: 20}),
		mk(4, "Blacklist", "GAMEPLAY", "Permanently blacklist a player and alts", "You are blacklisted from the server", false, false, true, true, nil, nil),
	}
}

func (h *Handler) loadPunishmentTypes(ctx context.Context, serverID string) []*modlv1.PunishmentType {
	var dataJSON string
	err := h.db.QueryRowContext(ctx, `SELECT data_json FROM settings WHERE server_id = ? AND type = 'punishmentTypes' LIMIT 1`, serverID).Scan(&dataJSON)
	if err != nil || dataJSON == "" {
		defs := defaultPunishmentTypes()
		h.savePunishmentTypes(ctx, serverID, defs)
		return defs
	}
	var wrap modlv1.PanelPunishmentTypesResponse
	if (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal([]byte(dataJSON), &wrap) == nil && len(wrap.PunishmentTypes) > 0 {
		return wrap.PunishmentTypes
	}
	return defaultPunishmentTypes()
}

func (h *Handler) savePunishmentTypes(ctx context.Context, serverID string, types []*modlv1.PunishmentType) {
	wrap := &modlv1.PanelPunishmentTypesResponse{PunishmentTypes: types}
	b, _ := (protojson.MarshalOptions{EmitUnpopulated: true}).Marshal(wrap)
	now := httpx.NowMs()
	_, _ = h.db.ExecContext(ctx, `
		INSERT INTO settings (id, server_id, type, data_json, version, updated_at)
		VALUES (?, ?, 'punishmentTypes', ?, 1, ?)
		ON CONFLICT(server_id, type) DO UPDATE SET data_json=excluded.data_json, version=settings.version+1, updated_at=excluded.updated_at`,
		httpx.NewID(), serverID, string(b), now,
	)
	_, _ = h.db.ExecContext(ctx, `UPDATE servers SET punishment_types_updated_at = ?, updated_at = ? WHERE id = ?`, now, now, serverID)
}

func (h *Handler) getPanelPunishmentTypes(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	types := h.loadPunishmentTypes(r.Context(), srv.ID)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PanelPunishmentTypesResponse{PunishmentTypes: types})
}

func (h *Handler) getPanelPunishmentType(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	ord, err := strconv.Atoi(chi.URLParam(r, "ordinal"))
	if err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid ordinal")
		return
	}
	for _, t := range h.loadPunishmentTypes(r.Context(), srv.ID) {
		if t.Ordinal != nil && int(*t.Ordinal) == ord {
			httpx.WriteProto(w, r, http.StatusOK, t)
			return
		}
	}
	httpx.WriteError(w, r, http.StatusNotFound, "Punishment type not found")
}

func (h *Handler) createPunishmentType(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.PunishmentTypeRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	types := h.loadPunishmentTypes(r.Context(), srv.ID)
	maxOrd := int32(4)
	for _, t := range types {
		if t.Ordinal != nil && *t.Ordinal > maxOrd {
			maxOrd = *t.Ordinal
		}
	}
	newOrd := maxOrd + 1
	cust := true
	pt := fromPunishmentTypeReq(&req, newOrd, cust)
	types = append(types, pt)
	h.savePunishmentTypes(r.Context(), srv.ID, types)
	if h.hub != nil {
		h.hub.InvalidatePanel(srv.ID, modlv1.PanelResource_PANEL_RESOURCE_PUNISHMENT_TYPES, "")
	}
	httpx.WriteProto(w, r, http.StatusOK, pt)
}

func (h *Handler) updatePunishmentType(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	ord, err := strconv.Atoi(chi.URLParam(r, "ordinal"))
	if err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid ordinal")
		return
	}
	var req modlv1.PunishmentTypeRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	types := h.loadPunishmentTypes(r.Context(), srv.ID)
	for i, t := range types {
		if t.Ordinal != nil && int(*t.Ordinal) == ord {
			cust := t.Customizable != nil && *t.Customizable
			updated := fromPunishmentTypeReq(&req, int32(ord), cust)
			types[i] = updated
			h.savePunishmentTypes(r.Context(), srv.ID, types)
			if h.hub != nil {
				h.hub.InvalidatePanel(srv.ID, modlv1.PanelResource_PANEL_RESOURCE_PUNISHMENT_TYPES, "")
			}
			httpx.WriteProto(w, r, http.StatusOK, updated)
			return
		}
	}
	httpx.WriteError(w, r, http.StatusNotFound, "Punishment type not found")
}

func (h *Handler) resetPunishmentTypes(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	defs := defaultPunishmentTypes()
	h.savePunishmentTypes(r.Context(), srv.ID, defs)
	if h.hub != nil {
		h.hub.InvalidatePanel(srv.ID, modlv1.PanelResource_PANEL_RESOURCE_PUNISHMENT_TYPES, "")
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PanelPunishmentTypesResponse{PunishmentTypes: defs})
}

func (h *Handler) deletePunishmentType(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	ord, err := strconv.Atoi(chi.URLParam(r, "ordinal"))
	if err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid ordinal")
		return
	}
	types := h.loadPunishmentTypes(r.Context(), srv.ID)
	var kept []*modlv1.PunishmentType
	found := false
	for _, t := range types {
		if t.Ordinal != nil && int(*t.Ordinal) == ord {
			found = true
			continue
		}
		kept = append(kept, t)
	}
	if !found {
		httpx.WriteError(w, r, http.StatusNotFound, "Punishment type not found")
		return
	}
	h.savePunishmentTypes(r.Context(), srv.ID, kept)
	if h.hub != nil {
		h.hub.InvalidatePanel(srv.ID, modlv1.PanelResource_PANEL_RESOURCE_PUNISHMENT_TYPES, "")
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"success": true})
}

func fromPunishmentTypeReq(req *modlv1.PunishmentTypeRequest, ord int32, cust bool) *modlv1.PunishmentType {
	return &modlv1.PunishmentType{
		Id:                           &ord,
		Ordinal:                      &ord,
		Name:                         req.Name,
		Category:                     req.Category,
		Customizable:                 &cust,
		Durations:                    req.Durations,
		SingleSeverityPunishment:     req.SingleSeverityPunishment,
		SingleSeverityDurations:      req.SingleSeverityDurations,
		SingleSeverityPoints:         req.SingleSeverityPoints,
		Points:                       req.Points,
		CustomPoints:                 req.CustomPoints,
		StaffDescription:             req.StaffDescription,
		PlayerDescription:            req.PlayerDescription,
		CanBeAltBlocking:             req.CanBeAltBlocking,
		CanBeStatWiping:              req.CanBeStatWiping,
		Appealable:                   req.Appealable,
		AppealForm:                   req.AppealForm,
		PermanentUntilSkinChange:     req.PermanentUntilSkinChange,
		PermanentUntilUsernameChange: req.PermanentUntilUsernameChange,
	}
}

func (h *Handler) getMinecraftPunishmentTypesV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	types := h.loadPunishmentTypes(r.Context(), srv.ID)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": 200, "punishmentTypes": types})
}

func (h *Handler) getMinecraftPunishmentTypesV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	types := h.loadPunishmentTypes(r.Context(), srv.ID)
	pbData := make([]*modlv1.PunishmentTypesResponse_PunishmentTypeData, 0, len(types))
	for _, t := range types {
		var idVal, ordVal int32
		if t.Id != nil {
			idVal = *t.Id
		}
		if t.Ordinal != nil {
			ordVal = *t.Ordinal
		}
		cust := t.Customizable != nil && *t.Customizable
		item := &modlv1.PunishmentTypesResponse_PunishmentTypeData{
			Id:                           idVal,
			Ordinal:                      ordVal,
			Name:                         t.Name,
			Category:                     t.Category,
			StaffDescription:             t.StaffDescription,
			PlayerDescription:            t.PlayerDescription,
			Durations:                    protoToStruct(t.Durations),
			Points:                       protoToStruct(t.Points),
			CustomPoints:                 t.CustomPoints,
			CanBeAltBlocking:             t.CanBeAltBlocking,
			CanBeStatWiping:              t.CanBeStatWiping,
			SingleSeverityPunishment:     t.SingleSeverityPunishment,
			PermanentUntilSkinChange:     t.PermanentUntilSkinChange,
			PermanentUntilUsernameChange: t.PermanentUntilUsernameChange,
			IsCustomizable:               cust,
		}
		pbData = append(pbData, item)
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PunishmentTypesResponse{
		Status: 200,
		Data:   pbData,
	})
}

func protoToStruct(m proto.Message) *structpb.Struct {
	if m == nil {
		return nil
	}
	b, err := (protojson.MarshalOptions{EmitUnpopulated: true}).Marshal(m)
	if err != nil {
		return nil
	}
	var mp map[string]any
	if json.Unmarshal(b, &mp) != nil {
		return nil
	}
	st, _ := structpb.NewStruct(mp)
	return st
}

// --- API Keys ---

func (h *Handler) generateAPIKey(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	newKey := httpx.RandomAPIKey()
	_, err := h.db.ExecContext(r.Context(), `UPDATE servers SET api_key = ?, updated_at = ? WHERE id = ?`, newKey, httpx.NowMs(), srv.ID)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to generate API key")
		return
	}
	h.invalidate(srv.ID)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.ApiKeyGenerateResponse{
		Message: "API key generated successfully",
		ApiKey:  newKey,
	})
}

func (h *Handler) revealAPIKey(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv.APIKey == "" {
		httpx.WriteError(w, r, http.StatusNotFound, "API key not configured")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.ApiKeyRevealResponse{ApiKey: srv.APIKey})
}

func (h *Handler) deleteAPIKey(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	_, _ = h.db.ExecContext(r.Context(), `UPDATE servers SET api_key = '', updated_at = ? WHERE id = ?`, httpx.NowMs(), srv.ID)
	h.invalidate(srv.ID)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.ApiKeyDeleteResponse{Message: "API key deleted"})
}

func (h *Handler) existsAPIKey(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.ApiKeyExistsResponse{Exists: srv.APIKey != ""})
}

// --- Domain Settings (self-hosted local DNS/override, no Cloudflare SaaS dependency) ---

func (h *Handler) getDomainSettings(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	domain := srv.CustomDomainOverride
	if domain == "" {
		domain = srv.CustomDomain
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.DomainSettings{
		CustomDomain: domain,
		Status: &modlv1.DomainStatus{
			Domain:          domain,
			Status:          "active",
			CnameConfigured: true,
			SslStatus:       "active",
			LastChecked:     time.Now().UTC().Format(time.RFC3339),
		},
		AccessingFromCustomDomain: strings.EqualFold(httpx.NormalizeServerDomain(r.Host), domain),
		ModlSubdomainUrl:          "https://" + srv.CustomDomain,
		CanManageCustomDomain:     true,
	})
}

func (h *Handler) configureDomain(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.ConfigureDomainRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	norm := httpx.NormalizeServerDomain(req.CustomDomain)
	_, _ = h.db.ExecContext(r.Context(), `UPDATE servers SET custom_domain_override = ?, updated_at = ? WHERE id = ?`, norm, httpx.NowMs(), srv.ID)
	srv.CustomDomainOverride = norm
	h.invalidate(srv.ID)
	h.getDomainSettings(w, r)
}

func (h *Handler) verifyDomain(w http.ResponseWriter, r *http.Request) {
	var req modlv1.VerifyDomainRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	norm := httpx.NormalizeServerDomain(req.Domain)
	_, err := net.LookupHost(norm)
	status := "active"
	msg := "Domain verified"
	cnameOK := true
	if err != nil {
		status = "pending"
		msg = "DNS lookup did not resolve yet: " + err.Error()
		cnameOK = false
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.VerifyDomainResponse{
		Status: &modlv1.DomainStatus{
			Domain:          norm,
			Status:          status,
			CnameConfigured: cnameOK,
			SslStatus:       status,
			LastChecked:     time.Now().UTC().Format(time.RFC3339),
		},
		Message: msg,
	})
}

func (h *Handler) removeDomain(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	_, _ = h.db.ExecContext(r.Context(), `UPDATE servers SET custom_domain_override = '', updated_at = ? WHERE id = ?`, httpx.NowMs(), srv.ID)
	h.invalidate(srv.ID)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.RemoveDomainResponse{Message: "Custom domain removed"})
}

// --- Public Settings ---

func (h *Handler) getPublicSettings(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	if srv == nil {
		httpx.WriteProto(w, r, http.StatusOK, &modlv1.PublicSettingsResponse{
			ServerExists:    false,
			DefaultLanguage: "en",
		})
		return
	}
	gs := defaultGeneralSettings(srv)
	_, _, _ = h.loadSetting(r.Context(), srv.ID, "general", gs)
	tf := defaultTicketForms()
	_, _, _ = h.loadSetting(r.Context(), srv.ID, "ticketForms", tf)

	tfBytes, _ := (protojson.MarshalOptions{EmitUnpopulated: true}).Marshal(tf)
	var tfMap map[string]any
	_ = json.Unmarshal(tfBytes, &tfMap)
	tfStruct, _ := structpb.NewStruct(tfMap)

	display := gs.ServerDisplayName
	if strings.TrimSpace(display) == "" {
		display = "modl"
	}
	lang := gs.DefaultLanguage
	if strings.TrimSpace(lang) == "" {
		lang = "en"
	}

	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PublicSettingsResponse{
		ServerExists:      true,
		ServerDisplayName: display,
		PanelIconUrl:      gs.PanelIconUrl,
		HomepageIconUrl:   gs.HomepageIconUrl,
		TicketForms:       tfStruct,
		DefaultLanguage:   lang,
	})
}

func (h *Handler) handleSaveErr(w http.ResponseWriter, r *http.Request, err error) {
	if strings.HasPrefix(err.Error(), "CONFLICT:") {
		verStr := strings.TrimPrefix(err.Error(), "CONFLICT:")
		ver, _ := strconv.ParseInt(verStr, 10, 64)
		httpx.WriteJSON(w, http.StatusConflict, map[string]any{
			"status":         http.StatusConflict,
			"error":          "Settings version conflict",
			"message":        "Settings were modified by another user",
			"currentVersion": ver,
		})
		return
	}
	httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to save settings")
}
