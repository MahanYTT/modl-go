package player

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/config"
	"modl-go/internal/httpx"
	"modl-go/internal/models"
	"modl-go/internal/realtime"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc   *Service
	store *Store
	cfg   *config.Config
	perms models.PermissionChecker
}

func NewHandler(db *sql.DB, cfg *config.Config, hub *realtime.Hub, perms models.PermissionChecker) *Handler {
	store := NewStore(db)
	svc := NewService(store, cfg, hub, perms)
	return &Handler{
		svc:   svc,
		store: store,
		cfg:   cfg,
		perms: perms,
	}
}

func (h *Handler) Service() *Service {
	return h.svc
}

func (h *Handler) RegisterPanelRoutes(r chi.Router) {
	r.Route("/players", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(httpx.RequirePanelPolicy(h.perms, models.AccessPolicy{Rule: models.RulePermitAll}))
			r.Post("/{uuid}/punishments", h.panelCreatePunishment)
		})
		r.Group(func(r chi.Router) {
			r.Use(httpx.RequirePanelPolicy(h.perms, models.AccessPolicy{Rule: models.RulePlayerAccess}))
			r.Get("/", h.panelSearchPlayers)
			r.Post("/", h.panelCreatePlayer)
			r.Get("/punishments/search", h.panelSearchPunishments)
			r.Get("/punishments/{punishmentId}", h.panelGetPunishmentByID)
			r.Get("/punishments/{punishmentId}/linked-bans", h.panelGetLinkedBans)
			r.Get("/{uuid}", h.panelGetPlayer)
			r.Post("/{uuid}/usernames", h.panelAddUsername)
			r.Post("/{uuid}/notes", h.panelAddNote)
			r.Post("/{uuid}/ips", h.panelAddIP)
			r.Post("/{uuid}/punishments/{punishmentId}/modifications", h.panelAddModification)
			r.Get("/{uuid}/punishments/active", h.panelGetActivePunishments)
			r.Get("/{uuid}/punishments/preview", h.panelPreviewPunishment)
			r.Post("/{uuid}/punishments/{punishmentId}/notes", h.panelAddPunishmentNote)
			r.Post("/{uuid}/punishments/{punishmentId}/evidence", h.panelAddEvidence)
			r.Post("/{uuid}/punishments/{punishmentId}/tickets", h.panelModifyPunishmentTickets)
			r.Get("/{uuid}/linked", h.panelGetLinkedAccounts)
			r.Post("/{uuid}/find-linked", h.panelFindAndLinkAccounts)
		})
	})
}

func (h *Handler) RegisterPublicRoutes(r chi.Router) {
	r.Get("/players/avatar/{uuid}", h.publicProxyAvatar)
	r.Get("/punishment/{punishmentId}/appeal-info", h.publicGetAppealInfo)
}

func (h *Handler) RegisterMinecraftV1Routes(r chi.Router) {
	r.Route("/players", func(r chi.Router) {
		r.Post("/login", h.mcLoginV1)
		r.Post("/disconnect", h.mcDisconnectV1)
		r.Post("/update-server", h.mcUpdateServerV1)
		r.Get("/online", h.mcGetOnlinePlayersV1)
		r.Get("/", h.mcGetPlayerByQueryV1)
		r.Get("/by-name", h.mcGetPlayerByUsernameV1)
		r.Post("/lookup", h.mcLookupPlayerV1)
		r.Post("/lookup-profile", h.mcLookupProfileV1)
		r.Post("/submit-ip-info", h.mcSubmitIPInfoV1)
		r.Post("/pardon", h.mcPardonPlayerV1)
		r.Post("/chat-log", h.mcSubmitChatLogsV1)
		r.Post("/command-log", h.mcSubmitCommandLogsV1)
		r.Post("/sync", h.mcSyncV1)
		r.Get("/{uuid}", h.mcGetPlayerByUUIDV1)
		r.Post("/{uuid}/notes", h.mcCreateNoteV1)
		r.Get("/{uuid}/linked-accounts", h.mcGetLinkedAccountsV1)
		r.Get("/{uuid}/punishments", h.mcGetPlayerPunishmentsV1)
		r.Get("/{uuid}/notes", h.mcGetPlayerNotesV1)
		r.Get("/{uuid}/reports", h.mcGetPlayerReportsV1)
		r.Get("/{uuid}/chat-logs", h.mcGetChatLogsV1)
		r.Get("/{uuid}/command-logs", h.mcGetCommandLogsV1)
	})
	r.Route("/punishments", func(r chi.Router) {
		r.Post("/create", h.mcCreatePunishmentV1)
		r.Post("/dynamic", h.mcCreatePunishmentDynamicV1)
		r.Get("/recent", h.mcGetRecentPunishmentsV1)
		r.Get("/preview", h.mcPreviewPunishmentV1)
		r.Post("/acknowledge", h.mcAcknowledgePunishmentV1)
		r.Get("/{punishmentId}", h.mcGetPunishmentByIDV1)
		r.Post("/{punishmentId}/upload-token", h.mcCreateUploadTokenV1)
		r.Post("/{punishmentId}/pardon", h.mcPardonPunishmentV1)
		r.Post("/{punishmentId}/note", h.mcAddPunishmentNoteV1)
		r.Post("/{punishmentId}/evidence", h.mcAddPunishmentEvidenceV1)
		r.Post("/{punishmentId}/duration", h.mcChangePunishmentDurationV1)
		r.Post("/{punishmentId}/toggle", h.mcTogglePunishmentOptionV1)
		r.Post("/{punishmentId}/stat-wipe-acknowledge", h.mcAcknowledgeStatWipeV1)
		r.Post("/{punishmentId}/tickets", h.mcModifyPunishmentTicketsV1)
	})
	r.Post("/notifications/acknowledge", h.mcAcknowledgeNotificationsV1)
}

func (h *Handler) RegisterMinecraftV2Routes(r chi.Router) {
	r.Post("/startup", h.mcStartupV2)
	r.Post("/players/sync", h.mcSyncV2)
}

func (h *Handler) RegisterMinecraftV3Routes(r chi.Router) {
	r.Route("/players", func(r chi.Router) {
		r.Post("/login", h.mcLoginV3)
		r.Post("/disconnect", h.mcDisconnectV3)
		r.Post("/update-server", h.mcUpdateServerV3)
		r.Post("/submit-ip-info", h.mcSubmitIPInfoV3)
		r.Get("/online", h.mcGetOnlinePlayersV3)
		r.Get("/", h.mcGetPlayerByQueryV3)
		r.Get("/by-name", h.mcGetPlayerByUsernameV3)
		r.Post("/lookup", h.mcLookupPlayerV3)
		r.Post("/lookup-profile", h.mcLookupProfileV3)
		r.Post("/pardon", h.mcPardonPlayerV3)
		r.Post("/chat-log", h.mcSubmitChatLogsV3)
		r.Post("/command-log", h.mcSubmitCommandLogsV3)
		r.Post("/sync", h.mcSyncV3)
		r.Get("/{uuid}", h.mcGetPlayerByUUIDV3)
		r.Post("/{uuid}/notes", h.mcCreateNoteV3)
		r.Get("/{uuid}/linked-accounts", h.mcGetLinkedAccountsV3)
		r.Get("/{uuid}/punishments", h.mcGetPlayerPunishmentsV3)
		r.Get("/{uuid}/notes", h.mcGetPlayerNotesV3)
		r.Get("/{uuid}/reports", h.mcGetPlayerReportsV3)
	})
	r.Post("/presence", h.mcPresenceV3)
	r.Post("/logs", h.mcLogsV3)
	r.Post("/status", h.mcStatusV3)
	r.Route("/punishments", func(r chi.Router) {
		r.Get("/preview", h.mcPreviewPunishmentV3)
		r.Get("/recent", h.mcGetRecentPunishmentsV3)
		r.Post("/dynamic", h.mcCreatePunishmentDynamicV3)
		r.Post("/create", h.mcCreatePunishmentV3)
		r.Post("/acknowledge", h.mcAcknowledgePunishmentV3)
		r.Get("/{punishmentId}", h.mcGetPunishmentByIDV3)
		r.Post("/{punishmentId}/stat-wipe-acknowledge", h.mcAcknowledgeStatWipeV3)
		r.Post("/{punishmentId}/duration", h.mcChangePunishmentDurationV3)
		r.Post("/{punishmentId}/toggle", h.mcTogglePunishmentOptionV3)
		r.Post("/{punishmentId}/note", h.mcAddPunishmentNoteV3)
		r.Post("/{punishmentId}/evidence", h.mcAddPunishmentEvidenceV3)
		r.Post("/{punishmentId}/upload-token", h.mcCreateUploadTokenV3)
		r.Post("/{punishmentId}/pardon", h.mcPardonPunishmentV3)
		r.Post("/{punishmentId}/tickets", h.mcModifyPunishmentTicketsV3)
	})
	r.Post("/notifications/acknowledge", h.mcAcknowledgeNotificationsV3)
}

// ============================================================================
// PANEL HANDLERS
// ============================================================================

func (h *Handler) panelSearchPlayers(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	q := strings.TrimSpace(r.URL.Query().Get("search"))
	if len(q) < 2 {
		httpx.WriteError(w, r, http.StatusBadRequest, "search must be at least 2 characters")
		return
	}
	res := h.svc.SearchPlayers(r.Context(), srv.ID, q)
	httpx.WriteProto(w, r, http.StatusOK, toPlayerSearchResultsProto(res))
}

func (h *Handler) panelGetPlayer(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	norm, ok := NormalizeUUID(chi.URLParam(r, "uuid"))
	if !ok {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid UUID")
		return
	}
	detail, err := h.svc.GetPlayerDetails(r.Context(), srv.ID, norm)
	if errors.Is(err, ErrNotFound) {
		httpx.WriteError(w, r, http.StatusNotFound, "Player not found")
		return
	}
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load player")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, toPlayerDetailProto(detail))
}

func (h *Handler) panelCreatePlayer(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.CreatePlayerRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	_, err := h.store.CreatePlayer(r.Context(), srv.ID, req.MinecraftUuid, req.Username)
	if errors.Is(err, ErrAlreadyExist) {
		httpx.WriteError(w, r, http.StatusConflict, "Player already exists")
		return
	}
	if err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	httpx.WriteProto(w, r, http.StatusCreated, &modlv1.SimpleResponse{Success: true})
}

func (h *Handler) panelAddUsername(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	norm, ok := NormalizeUUID(chi.URLParam(r, "uuid"))
	if !ok {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid UUID")
		return
	}
	var req modlv1.AddPlayerUsernameRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	_, err := h.store.AddUsername(r.Context(), srv.ID, norm, req.Username)
	if errors.Is(err, ErrNotFound) {
		httpx.WriteError(w, r, http.StatusNotFound, "Player not found")
		return
	}
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to add username")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.SimpleResponse{Success: true})
}

func (h *Handler) panelAddNote(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	norm, ok := NormalizeUUID(chi.URLParam(r, "uuid"))
	if !ok {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid UUID")
		return
	}
	var req modlv1.CreatePlayerNoteRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	_, err := h.store.AddPlayerNote(r.Context(), srv.ID, norm, req.Text, derefStr(req.IssuerName), derefStr(req.IssuerId))
	if errors.Is(err, ErrNotFound) {
		httpx.WriteError(w, r, http.StatusNotFound, "Player not found")
		return
	}
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to add note")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.SimpleResponse{Success: true})
}

func (h *Handler) panelAddIP(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	norm, ok := NormalizeUUID(chi.URLParam(r, "uuid"))
	if !ok {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid UUID")
		return
	}
	var req modlv1.AddPlayerIpRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	_, err := h.store.AddPlayerIP(r.Context(), srv.ID, norm, req.IpAddress)
	if errors.Is(err, ErrNotFound) {
		httpx.WriteError(w, r, http.StatusNotFound, "Player not found")
		return
	}
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to add IP")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.SimpleResponse{Success: true})
}

func (h *Handler) panelCreatePunishment(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	email := httpx.GetSessionEmail(r)
	norm, ok := NormalizeUUID(chi.URLParam(r, "uuid"))
	if !ok {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid UUID")
		return
	}
	var req modlv1.PanelCreatePunishmentRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if code, msg := h.svc.ValidatePunishmentPermission(r.Context(), srv, email, int(req.TypeOrdinal)); code != 200 {
		httpx.WriteError(w, r, code, msg)
		return
	}
	notes := make([]CreateNoteInput, 0, len(req.Notes))
	for _, n := range req.Notes {
		notes = append(notes, CreateNoteInput{
			Text:       n.Text,
			IssuerName: derefStr(n.IssuerName),
			IssuerID:   derefStr(n.IssuerId),
		})
	}
	evs := make([]CreateEvidenceInput, 0, len(req.Evidence))
	for _, e := range req.Evidence {
		evs = append(evs, CreateEvidenceInput{
			Text:       e.Text,
			IssuerName: derefStr(e.IssuerName),
			IssuerID:   derefStr(e.IssuerId),
			Type:       derefStr(e.Type),
			FileURL:    derefStr(e.FileUrl),
			FileName:   derefStr(e.FileName),
			FileType:   derefStr(e.FileType),
			FileSize:   e.FileSize,
		})
	}
	input := CreatePunishmentInput{
		IssuerName:        derefStr(req.IssuerName),
		IssuerID:          derefStr(req.IssuerId),
		TypeOrdinal:       int(req.TypeOrdinal),
		Notes:             notes,
		Evidence:          evs,
		AttachedTicketIds: req.AttachedTicketIds,
		Severity:          derefStr(req.Severity),
		Status:            derefStr(req.Status),
		Data:              structToMap(req.Data),
		Reason:            derefStr(req.Reason),
		Duration:          req.Duration,
	}
	_, statusCode, msg := h.svc.CreatePunishment(r.Context(), srv, norm, input, false)
	if statusCode != 200 {
		httpx.WriteError(w, r, statusCode, msg)
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.SimpleResponse{Success: true})
}

func (h *Handler) panelAddModification(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	norm, ok := NormalizeUUID(chi.URLParam(r, "uuid"))
	if !ok {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid UUID")
		return
	}
	punishmentID := chi.URLParam(r, "punishmentId")
	var req modlv1.PanelAddModificationRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	_, statusCode, msg := h.svc.AddModification(r.Context(), srv, norm, punishmentID, AddModificationInput{
		Type:              req.Type,
		IssuerName:        derefStr(req.IssuerName),
		IssuerID:          derefStr(req.IssuerId),
		EffectiveDuration: req.EffectiveDuration,
		Reason:            derefStr(req.Reason),
		AppealTicketID:    derefStr(req.AppealTicketId),
	})
	if statusCode != 200 {
		httpx.WriteError(w, r, statusCode, msg)
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.SimpleResponse{Success: true})
}

func (h *Handler) panelGetActivePunishments(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	norm, ok := NormalizeUUID(chi.URLParam(r, "uuid"))
	if !ok {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid UUID")
		return
	}
	pl, err := h.store.FindPlayerByUUID(r.Context(), srv.ID, norm)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load player")
		return
	}
	items := make([]*modlv1.PunishmentResponse, 0)
	if pl != nil {
		resolved := h.store.ResolveIssuerNames(r.Context(), srv.ID, collectIssuerIDsForPlayers([]*Player{pl}))
		for _, p := range pl.Punishments {
			if h.svc.IsPunishmentActive(p) {
				items = append(items, toPunishmentResponseProto(h.svc.BuildPunishmentResponseView(r.Context(), srv.ID, p, nil, resolved)))
			}
		}
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.ActivePunishmentsResponse{Items: items})
}

func (h *Handler) panelPreviewPunishment(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	norm, ok := NormalizeUUID(chi.URLParam(r, "uuid"))
	if !ok {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid UUID")
		return
	}
	ordStr := r.URL.Query().Get("typeOrdinal")
	ord, err := strconv.Atoi(ordStr)
	if err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid typeOrdinal")
		return
	}
	prev := h.svc.PreviewPunishment(r.Context(), srv.ID, norm, ord)
	httpx.WriteProto(w, r, http.StatusOK, toPunishmentPreviewProto(prev))
}

func (h *Handler) panelGetPunishmentByID(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	pid := chi.URLParam(r, "punishmentId")
	pl, pun, err := h.store.FindPunishmentContext(r.Context(), srv.ID, pid)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load punishment")
		return
	}
	if pun == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Punishment not found")
		return
	}
	resolved := h.store.ResolveIssuerNames(r.Context(), srv.ID, collectIssuerIDs(pun))
	view := h.svc.BuildPunishmentResponseView(r.Context(), srv.ID, pun, pl, resolved)
	httpx.WriteProto(w, r, http.StatusOK, toPunishmentResponseProto(view))
}

func (h *Handler) panelSearchPunishments(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(q) < 2 {
		httpx.WriteError(w, r, http.StatusBadRequest, "q must be at least 2 characters")
		return
	}
	activeOnly := strings.EqualFold(r.URL.Query().Get("activeOnly"), "true")
	res := h.svc.SearchPunishments(r.Context(), srv.ID, q, activeOnly)
	httpx.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) panelAddPunishmentNote(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	norm, ok := NormalizeUUID(chi.URLParam(r, "uuid"))
	if !ok {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid UUID")
		return
	}
	pid := chi.URLParam(r, "punishmentId")
	var req modlv1.PanelAddPunishmentNoteRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	res := h.svc.AddPunishmentNote(r.Context(), srv, norm, pid, req.Text, derefStr(req.IssuerName), derefStr(req.IssuerId))
	if res.HTTPStatus != 200 {
		httpx.WriteError(w, r, res.HTTPStatus, res.Message)
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.SimpleResponse{Success: true})
}

func (h *Handler) panelAddEvidence(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	norm, ok := NormalizeUUID(chi.URLParam(r, "uuid"))
	if !ok {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid UUID")
		return
	}
	pid := chi.URLParam(r, "punishmentId")
	var req modlv1.PanelAddEvidenceRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	res := h.svc.AddPunishmentEvidence(r.Context(), srv, norm, pid, AddEvidenceInput{
		Text:       derefStr(req.Text),
		Type:       req.Type,
		IssuerName: derefStr(req.IssuerName),
		IssuerID:   derefStr(req.IssuerId),
		URL:        derefStr(req.Url),
		FileName:   derefStr(req.FileName),
		FileType:   derefStr(req.FileType),
		FileSize:   req.FileSize,
	}, false)
	if res.HTTPStatus != 200 {
		httpx.WriteError(w, r, res.HTTPStatus, res.Message)
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.SimpleResponse{Success: true})
}

func (h *Handler) panelModifyPunishmentTickets(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	norm, ok := NormalizeUUID(chi.URLParam(r, "uuid"))
	if !ok {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid UUID")
		return
	}
	pid := chi.URLParam(r, "punishmentId")
	var req modlv1.ModifyPunishmentTicketsRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	res := h.svc.ModifyPunishmentTickets(r.Context(), srv, norm, pid, ModifyTicketsInput{
		AddTicketIds:            req.AddTicketIds,
		RemoveTicketIds:         req.RemoveTicketIds,
		ModifyAssociatedTickets: req.ModifyAssociatedTickets,
		IssuerName:              req.IssuerName,
		IssuerID:                req.IssuerId,
	})
	if res.HTTPStatus != 200 {
		httpx.WriteError(w, r, res.HTTPStatus, res.Message)
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.SimpleResponse{Success: true})
}

func (h *Handler) panelGetLinkedAccounts(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	norm, ok := NormalizeUUID(chi.URLParam(r, "uuid"))
	if !ok {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid UUID")
		return
	}
	items := h.svc.GetLinkedAccountsPanel(r.Context(), srv.ID, norm)
	httpx.WriteProto(w, r, http.StatusOK, toPanelLinkedAccountsProto(items))
}

func (h *Handler) panelGetLinkedBans(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	pid := chi.URLParam(r, "punishmentId")
	items := h.svc.GetLinkedBansForParent(r.Context(), srv.ID, pid)
	httpx.WriteProto(w, r, http.StatusOK, toPanelLinkedBansProto(items))
}

func (h *Handler) panelFindAndLinkAccounts(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	norm, ok := NormalizeUUID(chi.URLParam(r, "uuid"))
	if !ok {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid UUID")
		return
	}
	okLink, msg, found := h.svc.FindAndLinkAccounts(r.Context(), srv.ID, norm)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PanelFindAndLinkAccountsResponse{
		Success:             okLink,
		Message:             msg,
		LinkedAccountsFound: int32(found),
	})
}

// ============================================================================
// PUBLIC HANDLERS
// ============================================================================

func (h *Handler) publicProxyAvatar(w http.ResponseWriter, r *http.Request) {
	norm, ok := NormalizeUUID(chi.URLParam(r, "uuid"))
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	size := httpx.QueryInt(r, "size", 32, 8, 512)
	overlay := true
	if raw := strings.TrimSpace(r.URL.Query().Get("overlay")); raw != "" {
		overlay = !strings.EqualFold(raw, "false")
	}
	avatar := h.svc.ProxyAvatar(r.Context(), norm, size, overlay)
	if len(avatar) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(avatar)
}

func (h *Handler) publicGetAppealInfo(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	pid := chi.URLParam(r, "punishmentId")
	info, notStartedMsg, found := h.svc.GetPublicAppealInfo(r.Context(), srv.ID, pid)
	if !found {
		httpx.WriteError(w, r, http.StatusNotFound, "Punishment not found")
		return
	}
	if notStartedMsg != "" {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": notStartedMsg})
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, toPublicAppealInfoProto(info))
}

// ============================================================================
// MINECRAFT V1 & V2 HANDLERS (JSON)
// ============================================================================

func (h *Handler) mcLoginV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req struct {
		MinecraftUUID string         `json:"minecraftUUID"`
		Username      string         `json:"username"`
		IP            string         `json:"ip"`
		IPInfo        map[string]any `json:"ipInfo"`
		SkinHash      string         `json:"skinHash"`
		ServerName    string         `json:"serverName"`
	}
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid JSON")
		return
	}
	norm, ok := NormalizeUUID(req.MinecraftUUID)
	if !ok || strings.TrimSpace(req.Username) == "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "minecraftUUID and username are required")
		return
	}
	res, err := h.svc.Login(r.Context(), srv, norm, req.Username, req.IP, req.IPInfo, req.SkinHash, req.ServerName)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Login failed")
		return
	}
	body := map[string]any{
		"status":               res.HTTPStatus,
		"activePunishments":    res.ActivePunishments,
		"pendingNotifications": res.PendingNotifications,
	}
	if len(res.PendingIPLookups) > 0 {
		body["pendingIpLookups"] = res.PendingIPLookups
	}
	if len(res.PendingStatWipes) > 0 {
		body["pendingStatWipes"] = res.PendingStatWipes
	}
	httpx.WriteJSON(w, res.HTTPStatus, body)
}

func (h *Handler) mcDisconnectV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req struct {
		MinecraftUuid     string `json:"minecraftUuid"`
		SessionDurationMs int64  `json:"sessionDurationMs"`
	}
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid JSON")
		return
	}
	_ = h.store.MarkDisconnected(r.Context(), srv.ID, req.MinecraftUuid, req.SessionDurationMs)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": 200, "success": true})
}

func (h *Handler) mcUpdateServerV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req struct {
		MinecraftUuid string `json:"minecraftUuid"`
		ServerName    string `json:"serverName"`
	}
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid JSON")
		return
	}
	_ = h.store.UpdateLastServer(r.Context(), srv.ID, req.MinecraftUuid, req.ServerName)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": 200, "success": true})
}

func (h *Handler) mcGetOnlinePlayersV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	players, _ := h.store.FindOnlinePlayers(r.Context(), srv.ID, 500)
	items := make([]map[string]any, 0, len(players))
	for _, pl := range players {
		items = append(items, map[string]any{
			"uuid":            pl.MinecraftUUID,
			"username":        pl.LatestUsername(),
			"joinedAt":        pl.LastLoginMs(),
			"totalPlaytimeMs": pl.TotalPlaytimeSeconds() * 1000,
		})
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": 200, "players": items})
}

func (h *Handler) mcGetPlayerByUUIDV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	norm, ok := NormalizeUUID(chi.URLParam(r, "uuid"))
	if !ok {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid UUID")
		return
	}
	pl, _ := h.store.FindPlayerByUUID(r.Context(), srv.ID, norm)
	if pl == nil {
		httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"status": 404, "message": "Player not found"})
		return
	}
	pLim := queryOptionalInt(r, "punishmentLimit")
	nLim := queryOptionalInt(r, "noteLimit")
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status":  200,
		"profile": h.svc.BuildPlayerProfileMap(r.Context(), srv, pl, pLim, nLim),
	})
}

func (h *Handler) mcGetPlayerByQueryV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	rawUUID := strings.TrimSpace(r.URL.Query().Get("minecraftUuid"))
	if rawUUID == "" {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": 400, "message": "minecraftUuid parameter required"})
		return
	}
	queryMojang := !strings.EqualFold(r.URL.Query().Get("queryMojang"), "false")
	pl, _ := h.store.FindPlayerByUUID(r.Context(), srv.ID, rawUUID)
	if pl == nil && queryMojang {
		if mp := h.svc.LookupMojangByUUID(r.Context(), rawUUID); mp != nil {
			httpx.WriteJSON(w, http.StatusOK, map[string]any{
				"status":  200,
				"message": "Player found via Mojang",
				"player":  mojangProfileMap(mp),
			})
			return
		}
	}
	if pl == nil {
		httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"status": 404, "message": "Player not found"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status":  200,
		"message": "Player found",
		"player":  h.svc.BuildPlayerProfileMap(r.Context(), srv, pl, nil, nil),
	})
}

func (h *Handler) mcGetPlayerByUsernameV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	uname := strings.TrimSpace(r.URL.Query().Get("username"))
	queryMojang := !strings.EqualFold(r.URL.Query().Get("queryMojang"), "false")
	pl, _ := h.store.FindBestByUsername(r.Context(), srv.ID, uname)
	if pl == nil && queryMojang {
		if mp := h.svc.LookupMojangByUsername(r.Context(), uname); mp != nil {
			httpx.WriteJSON(w, http.StatusOK, map[string]any{
				"status":  200,
				"message": "Player found via Mojang",
				"player":  mojangProfileMap(mp),
			})
			return
		}
	}
	if pl == nil {
		httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"status": 404, "message": "Player not found"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status":  200,
		"message": "Player found",
		"player":  h.svc.BuildPlayerProfileMap(r.Context(), srv, pl, nil, nil),
	})
}

func (h *Handler) mcLookupPlayerV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req struct {
		Query       string `json:"query"`
		QueryMojang *bool  `json:"queryMojang"`
	}
	if err := httpx.ReadJSON(r, &req); err != nil || strings.TrimSpace(req.Query) == "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "query is required")
		return
	}
	shouldMojang := req.QueryMojang == nil || *req.QueryMojang
	q := strings.TrimSpace(req.Query)
	norm, isUUID := NormalizeUUID(q)
	var pl *Player
	if isUUID {
		pl, _ = h.store.FindPlayerByUUID(r.Context(), srv.ID, norm)
	} else {
		pl, _ = h.store.FindBestByUsername(r.Context(), srv.ID, q)
	}
	if pl == nil && shouldMojang {
		var mp *MojangProfile
		if isUUID {
			mp = h.svc.LookupMojangByUUID(r.Context(), norm)
		} else {
			mp = h.svc.LookupMojangByUsername(r.Context(), q)
		}
		if mp != nil {
			httpx.WriteJSON(w, http.StatusOK, map[string]any{
				"status":  200,
				"message": "Player found via Mojang",
				"data": map[string]any{
					"minecraftUuid":     mp.UUID,
					"currentUsername":   mp.Name,
					"previousUsernames": []string{},
					"firstSeen":         nil,
					"lastSeen":          nil,
					"isOnline":          false,
					"punishmentStats": map[string]any{
						"totalPunishments":  0,
						"activePunishments": 0,
						"bans":              0,
						"mutes":             0,
						"kicks":             0,
						"warnings":          0,
					},
					"recentPunishments": []any{},
				},
			})
			return
		}
	}
	if pl == nil {
		httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"status": 404, "message": "Player not found"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status":  200,
		"message": "Player found",
		"data":    h.svc.BuildLookupResponseMap(r.Context(), srv, pl),
	})
}

func (h *Handler) mcLookupProfileV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req struct {
		Query       string `json:"query"`
		QueryMojang *bool  `json:"queryMojang"`
	}
	if err := httpx.ReadJSON(r, &req); err != nil || strings.TrimSpace(req.Query) == "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "query is required")
		return
	}
	shouldMojang := req.QueryMojang == nil || *req.QueryMojang
	q := strings.TrimSpace(req.Query)
	norm, isUUID := NormalizeUUID(q)
	var pl *Player
	if isUUID {
		pl, _ = h.store.FindPlayerByUUID(r.Context(), srv.ID, norm)
	} else {
		pl, _ = h.store.FindBestByUsername(r.Context(), srv.ID, q)
	}
	if pl == nil && shouldMojang {
		var mp *MojangProfile
		if isUUID {
			mp = h.svc.LookupMojangByUUID(r.Context(), norm)
		} else {
			mp = h.svc.LookupMojangByUsername(r.Context(), q)
		}
		if mp != nil {
			pl, _ = h.store.FindPlayerByUUID(r.Context(), srv.ID, mp.UUID)
		}
	}
	if pl == nil {
		httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"status": 404, "message": "Player not found"})
		return
	}
	pLim := queryOptionalInt(r, "punishmentLimit")
	nLim := queryOptionalInt(r, "noteLimit")
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status":  200,
		"profile": h.svc.BuildPlayerProfileMap(r.Context(), srv, pl, pLim, nLim),
	})
}

func (h *Handler) mcCreateNoteV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	norm, ok := NormalizeUUID(chi.URLParam(r, "uuid"))
	if !ok {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid UUID")
		return
	}
	var req struct {
		Text       string `json:"text"`
		IssuerName string `json:"issuerName"`
		IssuerID   string `json:"issuerId"`
	}
	if err := httpx.ReadJSON(r, &req); err != nil || strings.TrimSpace(req.Text) == "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "text is required")
		return
	}
	_, err := h.store.AddPlayerNote(r.Context(), srv.ID, norm, req.Text, req.IssuerName, req.IssuerID)
	if errors.Is(err, ErrNotFound) {
		httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"status": 404, "message": "Player not found"})
		return
	}
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to create note")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": 200, "success": true, "message": "Note added"})
}

func (h *Handler) mcGetLinkedAccountsV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	norm, ok := NormalizeUUID(chi.URLParam(r, "uuid"))
	if !ok {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid UUID")
		return
	}
	pl, _ := h.store.FindPlayerByUUID(r.Context(), srv.ID, norm)
	if pl == nil {
		httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"status": 404, "message": "Player not found"})
		return
	}
	profiles := h.getLinkedAccountProfiles(r, srv, pl)
	pagePtr := queryOptionalInt(r, "page")
	limitPtr := queryOptionalInt(r, "limit")
	if pagePtr != nil && limitPtr != nil {
		page, limit := clampPagination(*pagePtr, *limitPtr)
		paged, total, hasMore := paginateSlice(profiles, page, limit)
		httpx.WriteJSON(w, http.StatusOK, map[string]any{
			"status":         200,
			"linkedAccounts": paged,
			"totalCount":     total,
			"page":           page,
			"hasMore":        hasMore,
		})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status":         200,
		"linkedAccounts": profiles,
	})
}

func (h *Handler) mcGetPlayerPunishmentsV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	norm, ok := NormalizeUUID(chi.URLParam(r, "uuid"))
	if !ok {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid UUID")
		return
	}
	pl, _ := h.store.FindPlayerByUUID(r.Context(), srv.ID, norm)
	if pl == nil {
		httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"status": 404, "message": "Player not found"})
		return
	}
	page := httpx.QueryInt(r, "page", 1, 1, 10000)
	limit := httpx.QueryInt(r, "limit", 7, 1, 100)
	typesByOrd := indexTypesByOrdinal(h.store.GetPunishmentTypes(r.Context(), srv.ID))
	resolved := h.store.ResolveIssuerNames(r.Context(), srv.ID, collectIssuerIDsForPlayers([]*Player{pl}))
	sorted := SortPunishmentsNewestFirst(pl.Punishments)
	views := make([]PunishmentView, 0, len(sorted))
	for _, p := range sorted {
		views = append(views, h.svc.ToPunishmentView(p, typesByOrd, resolved, nil))
	}
	paged, total, hasMore := paginateSlice(views, page, limit)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status":      200,
		"punishments": paged,
		"totalCount":  total,
		"page":        page,
		"hasMore":     hasMore,
	})
}

func (h *Handler) mcGetPlayerNotesV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	norm, ok := NormalizeUUID(chi.URLParam(r, "uuid"))
	if !ok {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid UUID")
		return
	}
	pl, _ := h.store.FindPlayerByUUID(r.Context(), srv.ID, norm)
	if pl == nil {
		httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"status": 404, "message": "Player not found"})
		return
	}
	page := httpx.QueryInt(r, "page", 1, 1, 10000)
	limit := httpx.QueryInt(r, "limit", 7, 1, 100)
	resolved := h.store.ResolveIssuerNames(r.Context(), srv.ID, collectIssuerIDsForPlayers([]*Player{pl}))
	notes := make([]map[string]any, 0, len(pl.Notes))
	for i := len(pl.Notes) - 1; i >= 0; i-- {
		n := pl.Notes[i]
		notes = append(notes, map[string]any{
			"id":         n.ID,
			"text":       n.Text,
			"date":       n.DateMs,
			"issuerName": resolveIssuer(n.IssuerID, n.IssuerName, resolved),
			"issuerId":   n.IssuerID,
		})
	}
	paged, total, hasMore := paginateSlice(notes, page, limit)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status":     200,
		"notes":      paged,
		"totalCount": total,
		"page":       page,
		"hasMore":    hasMore,
	})
}

func (h *Handler) mcGetPlayerReportsV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	norm, ok := NormalizeUUID(chi.URLParam(r, "uuid"))
	if !ok {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid UUID")
		return
	}
	reports, _ := h.store.FindReportedPlayerTickets(r.Context(), srv.ID, norm, 50)
	if reports == nil {
		reports = []map[string]any{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": 200, "reports": reports})
}

func (h *Handler) mcSubmitIPInfoV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req struct {
		MinecraftUUID string `json:"minecraftUUID"`
		IP            string `json:"ip"`
		Country       string `json:"country"`
		Region        string `json:"region"`
		ASN           string `json:"asn"`
		Proxy         bool   `json:"proxy"`
		Hosting       bool   `json:"hosting"`
	}
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid JSON")
		return
	}
	_ = h.store.SubmitIPInfo(r.Context(), srv.ID, req.MinecraftUUID, req.IP, req.Country, req.Region, req.ASN, req.Proxy, req.Hosting)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": 200, "success": true})
}

func (h *Handler) mcPardonPlayerV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req struct {
		PlayerName     string `json:"playerName"`
		IssuerName     string `json:"issuerName"`
		IssuerID       string `json:"issuerId"`
		PunishmentType string `json:"punishmentType"`
		Reason         string `json:"reason"`
	}
	if err := httpx.ReadJSON(r, &req); err != nil || strings.TrimSpace(req.PlayerName) == "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "playerName is required")
		return
	}
	res := h.svc.PardonPlayer(r.Context(), srv, req.PlayerName, req.PunishmentType, req.IssuerName, req.IssuerID, req.Reason)
	if res.HTTPStatus == 404 {
		httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"status": 404, "message": res.Message})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status":        200,
		"success":       res.Success,
		"pardonedCount": res.Count,
		"message":       res.Message,
	})
}

func (h *Handler) mcSubmitChatLogsV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req struct {
		Entries []SyncChatLogInput `json:"entries"`
	}
	if err := httpx.ReadJSON(r, &req); err != nil || len(req.Entries) == 0 || len(req.Entries) > MaxChatBatchEntries {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid chat log batch")
		return
	}
	_ = h.store.InsertChatLogs(r.Context(), srv.ID, req.Entries)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) mcSubmitCommandLogsV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req struct {
		Entries []SyncCommandLogInput `json:"entries"`
	}
	if err := httpx.ReadJSON(r, &req); err != nil || len(req.Entries) == 0 || len(req.Entries) > MaxChatBatchEntries {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid command log batch")
		return
	}
	_ = h.store.InsertCommandLogs(r.Context(), srv.ID, req.Entries)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) mcGetChatLogsV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	uuidParam := chi.URLParam(r, "uuid")
	limit := httpx.QueryInt(r, "limit", 200, 1, MaxChatBatchEntries)
	entries, _ := h.store.GetChatLogs(r.Context(), srv.ID, uuidParam, limit)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"entries": entries})
}

func (h *Handler) mcGetCommandLogsV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	uuidParam := chi.URLParam(r, "uuid")
	limit := httpx.QueryInt(r, "limit", 200, 1, MaxChatBatchEntries)
	entries, _ := h.store.GetCommandLogs(r.Context(), srv.ID, uuidParam, limit)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"entries": entries})
}

func (h *Handler) mcSyncV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req struct {
		LastSyncTimestamp string                 `json:"lastSyncTimestamp"`
		OnlinePlayers     []SyncOnlinePlayerInput `json:"onlinePlayers"`
		ServerStatus      *SyncServerStatusInput `json:"serverStatus"`
		ServerName        string                 `json:"serverName"`
		ChatLogs          []SyncChatLogInput     `json:"chatLogs"`
		CommandLogs       []SyncCommandLogInput  `json:"commandLogs"`
	}
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid JSON")
		return
	}
	res := h.svc.Sync(r.Context(), srv, req.LastSyncTimestamp, req.OnlinePlayers, req.ServerName, req.ChatLogs, req.CommandLogs, req.ServerStatus, httpx.GetClientIP(r), false)
	httpx.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) mcSyncV2(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req struct {
		LastSyncTimestamp string                  `json:"lastSyncTimestamp"`
		OnlinePlayers     []SyncOnlinePlayerInput `json:"onlinePlayers"`
		ServerName        string                  `json:"serverName"`
		ChatLogs          []SyncChatLogInput      `json:"chatLogs"`
		CommandLogs       []SyncCommandLogInput   `json:"commandLogs"`
	}
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid JSON")
		return
	}
	res := h.svc.Sync(r.Context(), srv, req.LastSyncTimestamp, req.OnlinePlayers, req.ServerName, req.ChatLogs, req.CommandLogs, nil, httpx.GetClientIP(r), true)
	httpx.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) mcStartupV2(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req struct {
		ServerVersion string `json:"serverVersion"`
		PlatformType  string `json:"platformType"`
		PluginVersion string `json:"pluginVersion"`
		MaxPlayers    int    `json:"maxPlayers"`
		ServerName    string `json:"serverName"`
	}
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid JSON")
		return
	}
	res := h.svc.HandleStartup(r.Context(), srv, req.ServerVersion, req.PlatformType, req.PluginVersion, req.MaxPlayers, req.ServerName, httpx.GetClientIP(r))
	httpx.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) mcCreatePunishmentV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	req, err := parseMinecraftCreatePunishmentJSON(r)
	if err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	_, code, msg := h.svc.CreatePunishment(r.Context(), srv, req.TargetUUID, req.Input, true)
	if code != 200 {
		httpx.WriteError(w, r, code, msg)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) mcCreatePunishmentDynamicV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	req, err := parseMinecraftCreatePunishmentJSON(r)
	if err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	pid, code, msg := h.svc.CreatePunishment(r.Context(), srv, req.TargetUUID, req.Input, true)
	if code != 200 {
		httpx.WriteError(w, r, code, msg)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status":       200,
		"message":      "Punishment created",
		"punishmentId": pid,
	})
}

func (h *Handler) mcGetPunishmentByIDV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	pid := chi.URLParam(r, "punishmentId")
	pl, pun, err := h.store.FindPunishmentContext(r.Context(), srv.ID, pid)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load punishment")
		return
	}
	if pun == nil {
		httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"status": 404, "message": "Punishment not found"})
		return
	}
	typesByOrd := indexTypesByOrdinal(h.store.GetPunishmentTypes(r.Context(), srv.ID))
	resolved := h.store.ResolveIssuerNames(r.Context(), srv.ID, collectIssuerIDs(pun))
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status":     200,
		"punishment": h.svc.ToPunishmentView(pun, typesByOrd, resolved, pl),
	})
}

func (h *Handler) mcCreateUploadTokenV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	pid := chi.URLParam(r, "punishmentId")
	var body struct {
		IssuerName string `json:"issuerName"`
	}
	_ = httpx.ReadJSON(r, &body)
	pl, pun, _ := h.store.FindPunishmentContext(r.Context(), srv.ID, pid)
	if pun == nil {
		httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"status": 404, "message": "Punishment not found"})
		return
	}
	token, err := h.store.CreateEvidenceUploadToken(r.Context(), srv.ID, pid, pl.MinecraftUUID, body.IssuerName)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to create upload token")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": 200, "token": token})
}

func (h *Handler) mcGetRecentPunishmentsV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	hours := httpx.QueryInt(r, "hours", 48, 1, 8760)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status":      200,
		"punishments": h.svc.GetRecentPunishments(r.Context(), srv.ID, hours),
	})
}

func (h *Handler) mcPreviewPunishmentV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	playerUUID := r.URL.Query().Get("playerUuid")
	ord, err := strconv.Atoi(r.URL.Query().Get("typeOrdinal"))
	if err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid typeOrdinal")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, h.svc.PreviewPunishment(r.Context(), srv.ID, playerUUID, ord))
}

func (h *Handler) mcAcknowledgePunishmentV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req struct {
		PunishmentID string `json:"punishmentId"`
		PlayerUUID   string `json:"playerUuid"`
	}
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid JSON")
		return
	}
	norm, ok := NormalizeUUID(req.PlayerUUID)
	if !ok || req.PunishmentID == "" {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"status": 400, "message": "punishmentId and valid playerUuid are required"})
		return
	}
	res := h.svc.AcknowledgePunishment(r.Context(), srv, norm, req.PunishmentID)
	httpx.WriteJSON(w, res.HTTPStatus, map[string]any{"status": res.HTTPStatus, "message": res.Message})
}

func (h *Handler) mcPardonPunishmentV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	pid := chi.URLParam(r, "punishmentId")
	var req struct {
		IssuerName string `json:"issuerName"`
		IssuerID   string `json:"issuerId"`
		Reason     string `json:"reason"`
	}
	_ = httpx.ReadJSON(r, &req)
	res := h.svc.PardonPunishment(r.Context(), srv, pid, req.IssuerName, req.IssuerID, req.Reason)
	if res.HTTPStatus != 200 {
		httpx.WriteJSON(w, res.HTTPStatus, map[string]any{"status": res.HTTPStatus, "message": res.Message})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status":        200,
		"success":       res.Success,
		"pardonedCount": res.Count,
		"message":       res.Message,
	})
}

func (h *Handler) mcAddPunishmentNoteV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	pid := chi.URLParam(r, "punishmentId")
	var req struct {
		IssuerName string `json:"issuerName"`
		IssuerID   string `json:"issuerId"`
		Note       string `json:"note"`
	}
	if err := httpx.ReadJSON(r, &req); err != nil || strings.TrimSpace(req.Note) == "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "note is required")
		return
	}
	res := h.svc.AddPunishmentNote(r.Context(), srv, "", pid, req.Note, req.IssuerName, req.IssuerID)
	if res.HTTPStatus != 200 {
		httpx.WriteJSON(w, res.HTTPStatus, map[string]any{"status": res.HTTPStatus, "message": res.Message})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": 200, "success": true, "message": res.Message})
}

func (h *Handler) mcAddPunishmentEvidenceV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	pid := chi.URLParam(r, "punishmentId")
	var req struct {
		IssuerName  string `json:"issuerName"`
		IssuerID    string `json:"issuerId"`
		EvidenceURL string `json:"evidenceUrl"`
	}
	if err := httpx.ReadJSON(r, &req); err != nil || strings.TrimSpace(req.EvidenceURL) == "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "evidenceUrl is required")
		return
	}
	res := h.svc.AddPunishmentEvidence(r.Context(), srv, "", pid, AddEvidenceInput{
		Type:       "url",
		IssuerName: req.IssuerName,
		IssuerID:   req.IssuerID,
		URL:        req.EvidenceURL,
	}, true)
	if res.HTTPStatus != 200 {
		httpx.WriteJSON(w, res.HTTPStatus, map[string]any{"status": res.HTTPStatus, "message": res.Message})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": 200, "success": true, "message": res.Message})
}

func (h *Handler) mcChangePunishmentDurationV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	pid := chi.URLParam(r, "punishmentId")
	var req struct {
		IssuerName  string `json:"issuerName"`
		IssuerID    string `json:"issuerId"`
		NewDuration *int64 `json:"newDuration"`
	}
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid JSON")
		return
	}
	res := h.svc.ChangeDuration(r.Context(), srv, pid, req.NewDuration, req.IssuerName, req.IssuerID)
	if res.HTTPStatus != 200 {
		httpx.WriteJSON(w, res.HTTPStatus, map[string]any{"status": res.HTTPStatus, "message": res.Message})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": 200, "success": true, "message": res.Message})
}

func (h *Handler) mcTogglePunishmentOptionV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	pid := chi.URLParam(r, "punishmentId")
	var req struct {
		IssuerName string `json:"issuerName"`
		IssuerID   string `json:"issuerId"`
		Option     string `json:"option"`
		Enabled    bool   `json:"enabled"`
	}
	if err := httpx.ReadJSON(r, &req); err != nil || strings.TrimSpace(req.Option) == "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "option is required")
		return
	}
	res := h.svc.ToggleOption(r.Context(), srv, pid, req.Option, req.Enabled, req.IssuerName, req.IssuerID)
	if res.HTTPStatus != 200 {
		httpx.WriteJSON(w, res.HTTPStatus, map[string]any{"status": res.HTTPStatus, "message": res.Message})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": 200, "success": res.Success, "message": res.Message})
}

func (h *Handler) mcAcknowledgeStatWipeV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	pid := chi.URLParam(r, "punishmentId")
	res := h.svc.AcknowledgeStatWipe(r.Context(), srv, pid)
	if res.HTTPStatus != 200 {
		httpx.WriteJSON(w, res.HTTPStatus, map[string]any{"status": res.HTTPStatus, "message": res.Message})
		return
	}
	if !res.Success {
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": 200, "message": res.Message})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": 200, "success": true, "message": res.Message})
}

func (h *Handler) mcModifyPunishmentTicketsV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	pid := chi.URLParam(r, "punishmentId")
	var req struct {
		AddTicketIds            []string `json:"addTicketIds"`
		RemoveTicketIds         []string `json:"removeTicketIds"`
		ModifyAssociatedTickets bool     `json:"modifyAssociatedTickets"`
		IssuerName              string   `json:"issuerName"`
		IssuerID                string   `json:"issuerId"`
	}
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid JSON")
		return
	}
	res := h.svc.ModifyPunishmentTickets(r.Context(), srv, "", pid, ModifyTicketsInput{
		AddTicketIds:            req.AddTicketIds,
		RemoveTicketIds:         req.RemoveTicketIds,
		ModifyAssociatedTickets: req.ModifyAssociatedTickets,
		IssuerName:              req.IssuerName,
		IssuerID:                req.IssuerID,
	})
	if res.HTTPStatus != 200 {
		httpx.WriteJSON(w, res.HTTPStatus, map[string]any{"status": res.HTTPStatus, "message": res.Message})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": 200, "success": true, "message": res.Message})
}

func (h *Handler) mcAcknowledgeNotificationsV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req struct {
		PlayerUUID      string   `json:"playerUuid"`
		NotificationIDs []string `json:"notificationIds"`
	}
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid JSON")
		return
	}
	n, _ := h.store.AcknowledgeNotifications(r.Context(), srv.ID, req.PlayerUUID, req.NotificationIDs)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status":  200,
		"success": true,
		"message": "Acknowledged " + strconv.Itoa(n) + " notification(s)",
	})
}

// ============================================================================
// MINECRAFT V3 HANDLERS (Protobuf)
// ============================================================================

func (h *Handler) mcLoginV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.PlayerLoginRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	res, err := h.svc.Login(r.Context(), srv, req.MinecraftUuid, req.Username, derefStr(req.IpAddress), structToMap(req.IpInfo), derefStr(req.SkinHash), derefStr(req.ServerName))
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Login failed")
		return
	}
	httpx.WriteProto(w, r, res.HTTPStatus, toPlayerLoginResponseProto(res))
}

func (h *Handler) mcDisconnectV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.PlayerDisconnectRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	_ = h.store.MarkDisconnected(r.Context(), srv.ID, req.MinecraftUuid, req.SessionDurationMs)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.SimpleResponse{Success: true})
}

func (h *Handler) mcUpdateServerV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.UpdatePlayerServerRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	_ = h.store.UpdateLastServer(r.Context(), srv.ID, req.MinecraftUuid, req.ServerName)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.SimpleResponse{Success: true})
}

func (h *Handler) mcSubmitIPInfoV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.SubmitPlayerIpInfoRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	_ = h.store.SubmitIPInfo(r.Context(), srv.ID, req.MinecraftUuid, req.Ip, req.Country, req.Region, req.Asn, req.Proxy, req.Hosting)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.SimpleResponse{Success: true})
}

func (h *Handler) mcGetOnlinePlayersV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	players, _ := h.store.FindOnlinePlayers(r.Context(), srv.ID, 500)
	pb := &modlv1.OnlinePlayersResponse{
		Status:  200,
		Players: make([]*modlv1.OnlinePlayersResponse_OnlinePlayer, 0, len(players)),
	}
	for _, pl := range players {
		var joinedAt string
		if ll := pl.LastLoginMs(); ll != nil {
			joinedAt = epochISO(*ll)
		}
		ptMs := pl.TotalPlaytimeSeconds() * 1000
		pb.Players = append(pb.Players, &modlv1.OnlinePlayersResponse_OnlinePlayer{
			Uuid:            pl.MinecraftUUID,
			Username:        pl.LatestUsername(),
			JoinedAt:        joinedAt,
			TotalPlaytimeMs: &ptMs,
		})
	}
	httpx.WriteProto(w, r, http.StatusOK, pb)
}

func (h *Handler) mcGetPlayerByUUIDV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	norm, ok := NormalizeUUID(chi.URLParam(r, "uuid"))
	if !ok {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid UUID")
		return
	}
	pl, _ := h.store.FindPlayerByUUID(r.Context(), srv.ID, norm)
	if pl == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Player not found")
		return
	}
	pLim := queryOptionalInt(r, "punishmentLimit")
	nLim := queryOptionalInt(r, "noteLimit")
	profile := h.svc.BuildPlayerProfileMap(r.Context(), srv, pl, pLim, nLim)
	httpx.WriteProto(w, r, http.StatusOK, toPlayerProfileResponseProto(profile))
}

func (h *Handler) mcGetPlayerByQueryV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	rawUUID := strings.TrimSpace(r.URL.Query().Get("minecraftUuid"))
	if rawUUID == "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "minecraftUuid parameter required")
		return
	}
	queryMojang := !strings.EqualFold(r.URL.Query().Get("queryMojang"), "false")
	pl, _ := h.store.FindPlayerByUUID(r.Context(), srv.ID, rawUUID)
	if pl == nil && queryMojang {
		if mp := h.svc.LookupMojangByUUID(r.Context(), rawUUID); mp != nil {
			httpx.WriteProto(w, r, http.StatusOK, &modlv1.PlayerGetResponse{
				Status:  200,
				Message: "Player found via Mojang",
				Player:  toAccountProto(mojangProfileMap(mp)),
			})
			return
		}
	}
	if pl == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Player not found")
		return
	}
	profile := h.svc.BuildPlayerProfileMap(r.Context(), srv, pl, nil, nil)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PlayerGetResponse{
		Status:  200,
		Message: "Player found",
		Player:  toAccountProto(profile),
	})
}

func (h *Handler) mcGetPlayerByUsernameV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	uname := strings.TrimSpace(r.URL.Query().Get("username"))
	queryMojang := !strings.EqualFold(r.URL.Query().Get("queryMojang"), "false")
	pl, _ := h.store.FindBestByUsername(r.Context(), srv.ID, uname)
	if pl == nil && queryMojang {
		if mp := h.svc.LookupMojangByUsername(r.Context(), uname); mp != nil {
			httpx.WriteProto(w, r, http.StatusOK, &modlv1.PlayerNameResponse{
				Status:  200,
				Message: "Player found via Mojang",
				Player:  toAccountProto(mojangProfileMap(mp)),
			})
			return
		}
	}
	if pl == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Player not found")
		return
	}
	profile := h.svc.BuildPlayerProfileMap(r.Context(), srv, pl, nil, nil)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PlayerNameResponse{
		Status:  200,
		Message: "Player found",
		Player:  toAccountProto(profile),
	})
}

func (h *Handler) mcLookupPlayerV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.PlayerLookupRequest
	if err := httpx.ReadProto(r, &req); err != nil || strings.TrimSpace(req.Query) == "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "query is required")
		return
	}
	shouldMojang := req.QueryMojang == nil || *req.QueryMojang
	q := strings.TrimSpace(req.Query)
	norm, isUUID := NormalizeUUID(q)
	var pl *Player
	if isUUID {
		pl, _ = h.store.FindPlayerByUUID(r.Context(), srv.ID, norm)
	} else {
		pl, _ = h.store.FindBestByUsername(r.Context(), srv.ID, q)
	}
	if pl == nil && shouldMojang {
		var mp *MojangProfile
		if isUUID {
			mp = h.svc.LookupMojangByUUID(r.Context(), norm)
		} else {
			mp = h.svc.LookupMojangByUsername(r.Context(), q)
		}
		if mp != nil {
			httpx.WriteProto(w, r, http.StatusOK, toPlayerLookupResponseProto("Player found via Mojang", map[string]any{
				"minecraftUuid":   mp.UUID,
				"currentUsername": mp.Name,
				"isOnline":        false,
				"punishmentStats": map[string]any{},
			}))
			return
		}
	}
	if pl == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Player not found")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, toPlayerLookupResponseProto("Player found", h.svc.BuildLookupResponseMap(r.Context(), srv, pl)))
}

func (h *Handler) mcLookupProfileV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.PlayerLookupRequest
	if err := httpx.ReadProto(r, &req); err != nil || strings.TrimSpace(req.Query) == "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "query is required")
		return
	}
	shouldMojang := req.QueryMojang == nil || *req.QueryMojang
	q := strings.TrimSpace(req.Query)
	norm, isUUID := NormalizeUUID(q)
	var pl *Player
	if isUUID {
		pl, _ = h.store.FindPlayerByUUID(r.Context(), srv.ID, norm)
	} else {
		pl, _ = h.store.FindBestByUsername(r.Context(), srv.ID, q)
	}
	if pl == nil && shouldMojang {
		var mp *MojangProfile
		if isUUID {
			mp = h.svc.LookupMojangByUUID(r.Context(), norm)
		} else {
			mp = h.svc.LookupMojangByUsername(r.Context(), q)
		}
		if mp != nil {
			pl, _ = h.store.FindPlayerByUUID(r.Context(), srv.ID, mp.UUID)
		}
	}
	if pl == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Player not found")
		return
	}
	pLim := queryOptionalInt(r, "punishmentLimit")
	nLim := queryOptionalInt(r, "noteLimit")
	httpx.WriteProto(w, r, http.StatusOK, toPlayerProfileResponseProto(h.svc.BuildPlayerProfileMap(r.Context(), srv, pl, pLim, nLim)))
}

func (h *Handler) mcCreateNoteV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	norm, ok := NormalizeUUID(chi.URLParam(r, "uuid"))
	if !ok {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid UUID")
		return
	}
	var req modlv1.CreatePlayerNoteRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	_, err := h.store.AddPlayerNote(r.Context(), srv.ID, norm, req.Text, derefStr(req.IssuerName), derefStr(req.IssuerId))
	if errors.Is(err, ErrNotFound) {
		httpx.WriteProto(w, r, http.StatusNotFound, &modlv1.PlayerNoteCreateResponse{Status: 404, Success: false, Message: "Player not found"})
		return
	}
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to create note")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PlayerNoteCreateResponse{Status: 200, Success: true, Message: "Note added"})
}

func (h *Handler) mcGetLinkedAccountsV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	norm, ok := NormalizeUUID(chi.URLParam(r, "uuid"))
	if !ok {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid UUID")
		return
	}
	pl, _ := h.store.FindPlayerByUUID(r.Context(), srv.ID, norm)
	if pl == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Player not found")
		return
	}
	profiles := h.getLinkedAccountProfiles(r, srv, pl)
	pagePtr := queryOptionalInt(r, "page")
	limitPtr := queryOptionalInt(r, "limit")
	pb := &modlv1.LinkedAccountsResponse{Status: 200}
	if pagePtr != nil && limitPtr != nil {
		page, limit := clampPagination(*pagePtr, *limitPtr)
		paged, total, hasMore := paginateSlice(profiles, page, limit)
		tot := int32(total)
		pb.TotalCount = &tot
		pb.Page = int32(page)
		pb.HasMore = hasMore
		for _, prof := range paged {
			pb.LinkedAccounts = append(pb.LinkedAccounts, toAccountProto(prof))
		}
	} else {
		for _, prof := range profiles {
			pb.LinkedAccounts = append(pb.LinkedAccounts, toAccountProto(prof))
		}
	}
	httpx.WriteProto(w, r, http.StatusOK, pb)
}

func (h *Handler) mcGetPlayerPunishmentsV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	norm, ok := NormalizeUUID(chi.URLParam(r, "uuid"))
	if !ok {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid UUID")
		return
	}
	pl, _ := h.store.FindPlayerByUUID(r.Context(), srv.ID, norm)
	if pl == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Player not found")
		return
	}
	page := httpx.QueryInt(r, "page", 1, 1, 10000)
	limit := httpx.QueryInt(r, "limit", 7, 1, 100)
	typesByOrd := indexTypesByOrdinal(h.store.GetPunishmentTypes(r.Context(), srv.ID))
	resolved := h.store.ResolveIssuerNames(r.Context(), srv.ID, collectIssuerIDsForPlayers([]*Player{pl}))
	sorted := SortPunishmentsNewestFirst(pl.Punishments)
	views := make([]PunishmentView, 0, len(sorted))
	for _, p := range sorted {
		views = append(views, h.svc.ToPunishmentView(p, typesByOrd, resolved, nil))
	}
	paged, total, hasMore := paginateSlice(views, page, limit)
	pb := &modlv1.PaginatedPunishmentsResponse{
		Status:     200,
		TotalCount: int32(total),
		Page:       int32(page),
		HasMore:    hasMore,
	}
	for _, v := range paged {
		pb.Punishments = append(pb.Punishments, toPunishmentListEntryProto(v))
	}
	httpx.WriteProto(w, r, http.StatusOK, pb)
}

func (h *Handler) mcGetPlayerNotesV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	norm, ok := NormalizeUUID(chi.URLParam(r, "uuid"))
	if !ok {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid UUID")
		return
	}
	pl, _ := h.store.FindPlayerByUUID(r.Context(), srv.ID, norm)
	if pl == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Player not found")
		return
	}
	page := httpx.QueryInt(r, "page", 1, 1, 10000)
	limit := httpx.QueryInt(r, "limit", 7, 1, 100)
	resolved := h.store.ResolveIssuerNames(r.Context(), srv.ID, collectIssuerIDsForPlayers([]*Player{pl}))
	notes := make([]*modlv1.NoteEntry, 0, len(pl.Notes))
	for i := len(pl.Notes) - 1; i >= 0; i-- {
		n := pl.Notes[i]
		notes = append(notes, &modlv1.NoteEntry{
			Id:         n.ID,
			Text:       n.Text,
			Date:       epochISO(n.DateMs),
			IssuerName: resolveIssuer(n.IssuerID, n.IssuerName, resolved),
			IssuerId:   n.IssuerID,
		})
	}
	paged, total, hasMore := paginateSlice(notes, page, limit)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PaginatedNotesResponse{
		Status:     200,
		Notes:      paged,
		TotalCount: int32(total),
		Page:       int32(page),
		HasMore:    hasMore,
	})
}

func (h *Handler) mcGetPlayerReportsV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	norm, ok := NormalizeUUID(chi.URLParam(r, "uuid"))
	if !ok {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid UUID")
		return
	}
	reports, _ := h.store.FindReportedPlayerTickets(r.Context(), srv.ID, norm, 50)
	httpx.WriteProto(w, r, http.StatusOK, toReportsResponseProto(reports))
}

func (h *Handler) mcPardonPlayerV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.PardonPlayerRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	res := h.svc.PardonPlayer(r.Context(), srv, req.PlayerName, derefStr(req.PunishmentType), derefStr(req.IssuerName), derefStr(req.IssuerId), derefStr(req.Reason))
	httpx.WriteProto(w, r, res.HTTPStatus, &modlv1.PardonResponse{
		Status:        int32(res.HTTPStatus),
		Success:       res.Success,
		PardonedCount: int32(res.Count),
		Message:       res.Message,
	})
}

func (h *Handler) mcSubmitChatLogsV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.ChatLogBatchRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	entries := make([]SyncChatLogInput, 0, len(req.Entries))
	for _, e := range req.Entries {
		entries = append(entries, SyncChatLogInput{
			UUID:      e.Uuid,
			Username:  e.Username,
			Message:   e.Message,
			Timestamp: e.Timestamp,
			Server:    e.Server,
		})
	}
	_ = h.store.InsertChatLogs(r.Context(), srv.ID, entries)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.SimpleResponse{Success: true})
}

func (h *Handler) mcSubmitCommandLogsV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.CommandLogBatchRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	entries := make([]SyncCommandLogInput, 0, len(req.Entries))
	for _, e := range req.Entries {
		entries = append(entries, SyncCommandLogInput{
			UUID:      e.Uuid,
			Username:  e.Username,
			Command:   e.Command,
			Timestamp: e.Timestamp,
			Server:    e.Server,
		})
	}
	_ = h.store.InsertCommandLogs(r.Context(), srv.ID, entries)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.SimpleResponse{Success: true})
}

func (h *Handler) mcSyncV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.SyncRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.ChatLogs) > MaxChatBatchEntries || len(req.CommandLogs) > MaxChatBatchEntries {
		httpx.WriteError(w, r, http.StatusBadRequest, "Log batch exceeds maximum size")
		return
	}
	res := h.svc.Sync(
		r.Context(), srv, req.LastSyncTimestamp,
		syncPlayersFromProto(req.OnlinePlayers),
		derefStr(req.ServerName),
		syncChatLogsFromProto(req.ChatLogs),
		syncCmdLogsFromProto(req.CommandLogs),
		syncServerStatusFromProto(req.ServerStatus),
		httpx.GetClientIP(r), false,
	)
	httpx.WriteProto(w, r, http.StatusOK, toSyncResponseProto(res))
}

func (h *Handler) mcPresenceV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.SyncRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	h.svc.ApplyPresence(r.Context(), srv, syncPlayersFromProto(req.OnlinePlayers), derefStr(req.ServerName))
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.SimpleResponse{Success: true})
}

func (h *Handler) mcLogsV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.SyncRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.ChatLogs) > MaxChatBatchEntries || len(req.CommandLogs) > MaxChatBatchEntries {
		httpx.WriteError(w, r, http.StatusBadRequest, "Log batch exceeds maximum size")
		return
	}
	h.svc.SubmitLogs(r.Context(), srv, syncChatLogsFromProto(req.ChatLogs), syncCmdLogsFromProto(req.CommandLogs))
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.SimpleResponse{Success: true})
}

func (h *Handler) mcStatusV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.SyncRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	h.svc.ApplyServerStatus(r.Context(), srv, syncServerStatusFromProto(req.ServerStatus), derefStr(req.ServerName), httpx.GetClientIP(r))
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.SimpleResponse{Success: true})
}

func (h *Handler) mcPreviewPunishmentV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	playerUUID := r.URL.Query().Get("playerUuid")
	ord, err := strconv.Atoi(r.URL.Query().Get("typeOrdinal"))
	if err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid typeOrdinal")
		return
	}
	prev := h.svc.PreviewPunishment(r.Context(), srv.ID, playerUUID, ord)
	httpx.WriteProto(w, r, http.StatusOK, toPunishmentPreviewProto(prev))
}

func (h *Handler) mcGetRecentPunishmentsV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	hours := httpx.QueryInt(r, "hours", 48, 1, 8760)
	recent := h.svc.GetRecentPunishments(r.Context(), srv.ID, hours)
	pb := &modlv1.RecentPunishmentsResponse{
		Status:      200,
		Punishments: make([]*modlv1.RecentPunishmentsResponse_RecentPunishment, 0, len(recent)),
	}
	for _, v := range recent {
		pb.Punishments = append(pb.Punishments, toRecentPunishmentProto(v))
	}
	httpx.WriteProto(w, r, http.StatusOK, pb)
}

func (h *Handler) mcGetPunishmentByIDV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	pid := chi.URLParam(r, "punishmentId")
	pl, pun, err := h.store.FindPunishmentContext(r.Context(), srv.ID, pid)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load punishment")
		return
	}
	if pun == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Punishment not found")
		return
	}
	typesByOrd := indexTypesByOrdinal(h.store.GetPunishmentTypes(r.Context(), srv.ID))
	resolved := h.store.ResolveIssuerNames(r.Context(), srv.ID, collectIssuerIDs(pun))
	view := h.svc.ToPunishmentView(pun, typesByOrd, resolved, pl)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PunishmentDetailResponse{
		Status:     200,
		Punishment: toPunishmentDetailEntryProto(view),
	})
}

func (h *Handler) mcCreatePunishmentDynamicV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.CreatePunishmentRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	pid, code, msg := h.svc.CreatePunishment(r.Context(), srv, req.TargetUuid, fromCreatePunishmentProto(&req), true)
	if code != 200 {
		httpx.WriteError(w, r, code, msg)
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PunishmentCreateResponse{
		Status:       200,
		Message:      "Punishment created",
		PunishmentId: pid,
	})
}

func (h *Handler) mcCreatePunishmentV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.CreatePunishmentRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	_, code, msg := h.svc.CreatePunishment(r.Context(), srv, req.TargetUuid, fromCreatePunishmentProto(&req), true)
	if code != 200 {
		httpx.WriteError(w, r, code, msg)
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.SimpleResponse{Success: true})
}

func (h *Handler) mcAcknowledgePunishmentV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.PunishmentAcknowledgeRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	res := h.svc.AcknowledgePunishment(r.Context(), srv, req.PlayerUuid, req.PunishmentId)
	httpx.WriteProto(w, r, res.HTTPStatus, &modlv1.PunishmentAcknowledgeResponse{
		Status:  int32(res.HTTPStatus),
		Message: res.Message,
	})
}

func (h *Handler) mcAcknowledgeStatWipeV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	pid := chi.URLParam(r, "punishmentId")
	var req modlv1.StatWipeAcknowledgeRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	res := h.svc.AcknowledgeStatWipe(r.Context(), srv, pid)
	httpx.WriteProto(w, r, res.HTTPStatus, &modlv1.StatWipeAcknowledgeResponse{
		Status:  int32(res.HTTPStatus),
		Success: res.Success,
		Message: res.Message,
	})
}

func (h *Handler) mcChangePunishmentDurationV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	pid := chi.URLParam(r, "punishmentId")
	var req modlv1.ChangePunishmentDurationRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	res := h.svc.ChangeDuration(r.Context(), srv, pid, req.NewDuration, req.IssuerName, req.IssuerId)
	s := res.Success
	httpx.WriteProto(w, r, res.HTTPStatus, &modlv1.ChangePunishmentDurationResponse{
		Status:  int32(res.HTTPStatus),
		Success: &s,
		Message: res.Message,
	})
}

func (h *Handler) mcTogglePunishmentOptionV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	pid := chi.URLParam(r, "punishmentId")
	var req modlv1.TogglePunishmentOptionRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	res := h.svc.ToggleOption(r.Context(), srv, pid, req.Option, req.Enabled, req.IssuerName, req.IssuerId)
	s := res.Success
	httpx.WriteProto(w, r, res.HTTPStatus, &modlv1.TogglePunishmentOptionResponse{
		Status:  int32(res.HTTPStatus),
		Success: &s,
		Message: res.Message,
	})
}

func (h *Handler) mcAddPunishmentNoteV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	pid := chi.URLParam(r, "punishmentId")
	var req modlv1.AddPunishmentNoteRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	res := h.svc.AddPunishmentNote(r.Context(), srv, "", pid, req.Note, req.IssuerName, req.IssuerId)
	s := res.Success
	httpx.WriteProto(w, r, res.HTTPStatus, &modlv1.AddPunishmentNoteResponse{
		Status:  int32(res.HTTPStatus),
		Success: &s,
		Message: res.Message,
	})
}

func (h *Handler) mcAddPunishmentEvidenceV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	pid := chi.URLParam(r, "punishmentId")
	var req modlv1.AddPunishmentEvidenceRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	res := h.svc.AddPunishmentEvidence(r.Context(), srv, "", pid, AddEvidenceInput{
		Type:       "url",
		IssuerName: req.IssuerName,
		IssuerID:   req.IssuerId,
		URL:        req.EvidenceUrl,
	}, true)
	s := res.Success
	httpx.WriteProto(w, r, res.HTTPStatus, &modlv1.AddPunishmentEvidenceResponse{
		Status:  int32(res.HTTPStatus),
		Success: &s,
		Message: res.Message,
	})
}

func (h *Handler) mcCreateUploadTokenV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	pid := chi.URLParam(r, "punishmentId")
	var req modlv1.CreateEvidenceUploadTokenRequest
	_ = httpx.ReadProto(r, &req)
	pl, pun, _ := h.store.FindPunishmentContext(r.Context(), srv.ID, pid)
	if pun == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Punishment not found")
		return
	}
	token, err := h.store.CreateEvidenceUploadToken(r.Context(), srv.ID, pid, pl.MinecraftUUID, derefStr(req.IssuerName))
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to create upload token")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.EvidenceUploadTokenResponse{
		Status: 200,
		Token:  token,
	})
}

func (h *Handler) mcPardonPunishmentV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	pid := chi.URLParam(r, "punishmentId")
	var req modlv1.PardonPunishmentRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	res := h.svc.PardonPunishment(r.Context(), srv, pid, derefStr(req.IssuerName), derefStr(req.IssuerId), derefStr(req.Reason))
	httpx.WriteProto(w, r, res.HTTPStatus, &modlv1.PardonResponse{
		Status:        int32(res.HTTPStatus),
		Success:       res.Success,
		PardonedCount: int32(res.Count),
		Message:       res.Message,
	})
}

func (h *Handler) mcModifyPunishmentTicketsV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	pid := chi.URLParam(r, "punishmentId")
	var req modlv1.ModifyPunishmentTicketsRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	res := h.svc.ModifyPunishmentTickets(r.Context(), srv, "", pid, ModifyTicketsInput{
		AddTicketIds:            req.AddTicketIds,
		RemoveTicketIds:         req.RemoveTicketIds,
		ModifyAssociatedTickets: req.ModifyAssociatedTickets,
		IssuerName:              req.IssuerName,
		IssuerID:                req.IssuerId,
	})
	s := res.Success
	httpx.WriteProto(w, r, res.HTTPStatus, &modlv1.ModifyPunishmentTicketsResponse{
		Status:  int32(res.HTTPStatus),
		Success: &s,
		Message: res.Message,
	})
}

func (h *Handler) mcAcknowledgeNotificationsV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.AcknowledgeNotificationsRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	_, _ = h.store.AcknowledgeNotifications(r.Context(), srv.ID, req.PlayerUuid, req.NotificationIds)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.SimpleResponse{Success: true})
}

// ============================================================================
// HANDLER HELPERS
// ============================================================================

type parsedMCCreatePunishment struct {
	TargetUUID string
	Input      CreatePunishmentInput
}

func parseMinecraftCreatePunishmentJSON(r *http.Request) (*parsedMCCreatePunishment, error) {
	var raw struct {
		TargetUUID        string         `json:"targetUuid"`
		IssuerName        string         `json:"issuerName"`
		IssuerID          string         `json:"issuerId"`
		TypeOrdinal       *int           `json:"typeOrdinal"`
		TypeOrdinalSnake  *int           `json:"type_ordinal"`
		Reason            string         `json:"reason"`
		Duration          *int64         `json:"duration"`
		Data              map[string]any `json:"data"`
		Notes             []string       `json:"notes"`
		AttachedTicketIds []string       `json:"attachedTicketIds"`
		Severity          string         `json:"severity"`
		Status            string         `json:"status"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 4<<20)).Decode(&raw); err != nil {
		return nil, errors.New("invalid JSON body")
	}
	norm, ok := NormalizeUUID(raw.TargetUUID)
	if !ok {
		return nil, errors.New("valid targetUuid is required")
	}
	ord := 0
	if raw.TypeOrdinal != nil {
		ord = *raw.TypeOrdinal
	} else if raw.TypeOrdinalSnake != nil {
		ord = *raw.TypeOrdinalSnake
	} else {
		return nil, errors.New("typeOrdinal is required")
	}
	notes := make([]CreateNoteInput, 0, len(raw.Notes))
	for _, n := range raw.Notes {
		if strings.TrimSpace(n) != "" {
			notes = append(notes, CreateNoteInput{Text: n, IssuerName: raw.IssuerName, IssuerID: raw.IssuerID})
		}
	}
	return &parsedMCCreatePunishment{
		TargetUUID: norm,
		Input: CreatePunishmentInput{
			IssuerName:        raw.IssuerName,
			IssuerID:          raw.IssuerID,
			TypeOrdinal:       ord,
			Notes:             notes,
			AttachedTicketIds: raw.AttachedTicketIds,
			Severity:          raw.Severity,
			Status:            raw.Status,
			Data:              raw.Data,
			Reason:            raw.Reason,
			Duration:          raw.Duration,
		},
	}, nil
}

func fromCreatePunishmentProto(req *modlv1.CreatePunishmentRequest) CreatePunishmentInput {
	notes := make([]CreateNoteInput, 0, len(req.Notes))
	for _, n := range req.Notes {
		if strings.TrimSpace(n) != "" {
			notes = append(notes, CreateNoteInput{
				Text:       n,
				IssuerName: derefStr(req.IssuerName),
				IssuerID:   derefStr(req.IssuerId),
			})
		}
	}
	return CreatePunishmentInput{
		IssuerName:        derefStr(req.IssuerName),
		IssuerID:          derefStr(req.IssuerId),
		TypeOrdinal:       int(req.TypeOrdinal),
		Notes:             notes,
		AttachedTicketIds: req.AttachedTicketIds,
		Severity:          derefStr(req.Severity),
		Status:            derefStr(req.Status),
		Data:              structToMap(req.Data),
		Reason:            derefStr(req.Reason),
		Duration:          req.Duration,
	}
}

func (h *Handler) getLinkedAccountProfiles(r *http.Request, srv *models.Server, pl *Player) []map[string]any {
	ips := make([]string, 0, len(pl.IPAddresses))
	for _, ip := range pl.IPAddresses {
		if ip.IPAddress != "" {
			ips = append(ips, ip.IPAddress)
		}
	}
	var linked []*Player
	seen := map[string]bool{strings.ToLower(pl.MinecraftUUID): true}
	if len(ips) > 0 {
		rel, _ := h.store.FindPlayersByIPAddresses(r.Context(), srv.ID, ips, pl.MinecraftUUID, 20)
		for _, p := range rel {
			low := strings.ToLower(p.MinecraftUUID)
			if !seen[low] {
				seen[low] = true
				linked = append(linked, p)
			}
		}
	}
	var missing []string
	for _, u := range pl.LinkedAccountUUIDs() {
		low := strings.ToLower(u)
		if !seen[low] {
			seen[low] = true
			missing = append(missing, low)
		}
	}
	if len(missing) > 0 {
		more, _ := h.store.FindPlayersByUUIDs(r.Context(), srv.ID, missing)
		linked = append(linked, more...)
	}
	profiles := make([]map[string]any, 0, len(linked))
	for _, lp := range linked {
		profiles = append(profiles, h.svc.BuildPlayerProfileMap(r.Context(), srv, lp, nil, nil))
	}
	return profiles
}

func mojangProfileMap(mp *MojangProfile) map[string]any {
	return map[string]any{
		"minecraftUuid":        mp.UUID,
		"usernames":            []map[string]any{{"username": mp.Name}},
		"notes":                []map[string]any{},
		"ipAddresses":          []map[string]any{},
		"punishments":          []PunishmentView{},
		"pendingNotifications": []map[string]any{},
		"data":                 map[string]any{},
	}
}

func syncPlayersFromProto(items []*modlv1.SyncOnlinePlayer) []SyncOnlinePlayerInput {
	out := make([]SyncOnlinePlayerInput, 0, len(items))
	for _, it := range items {
		out = append(out, SyncOnlinePlayerInput{
			UUID:              it.Uuid,
			Username:          it.Username,
			IPAddress:         it.IpAddress,
			SessionDurationMs: it.SessionDurationMs,
		})
	}
	return out
}

func syncChatLogsFromProto(items []*modlv1.SyncChatLogEntry) []SyncChatLogInput {
	out := make([]SyncChatLogInput, 0, len(items))
	for _, it := range items {
		out = append(out, SyncChatLogInput{
			UUID:      it.Uuid,
			Username:  it.Username,
			Message:   it.Message,
			Server:    it.Server,
			Timestamp: it.Timestamp,
		})
	}
	return out
}

func syncCmdLogsFromProto(items []*modlv1.SyncCommandLogEntry) []SyncCommandLogInput {
	out := make([]SyncCommandLogInput, 0, len(items))
	for _, it := range items {
		out = append(out, SyncCommandLogInput{
			UUID:      it.Uuid,
			Username:  it.Username,
			Command:   it.Command,
			Server:    it.Server,
			Timestamp: it.Timestamp,
		})
	}
	return out
}

func syncServerStatusFromProto(s *modlv1.SyncServerStatus) *SyncServerStatusInput {
	if s == nil {
		return nil
	}
	return &SyncServerStatusInput{
		OnlinePlayerCount: int(s.OnlinePlayerCount),
		MaxPlayers:        int(s.MaxPlayers),
		ServerVersion:     s.ServerVersion,
		Timestamp:         s.Timestamp,
		PlatformType:      s.PlatformType,
		PluginVersion:     s.PluginVersion,
	}
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func queryOptionalInt(r *http.Request, name string) *int {
	raw := strings.TrimSpace(r.URL.Query().Get(name))
	if raw == "" {
		return nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return nil
	}
	return &v
}

func clampPagination(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 1
	} else if limit > 100 {
		limit = 100
	}
	return page, limit
}

func paginateSlice[T any](items []T, page, limit int) ([]T, int, bool) {
	total := len(items)
	start := (page - 1) * limit
	if start >= total {
		return []T{}, total, false
	}
	end := start + limit
	if end > total {
		end = total
	}
	return items[start:end], total, end < total
}
