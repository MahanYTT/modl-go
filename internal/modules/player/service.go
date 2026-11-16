package player

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/config"
	"modl-go/internal/httpx"
	"modl-go/internal/models"
	"modl-go/internal/realtime"
)

type Service struct {
	store      *Store
	cfg        *config.Config
	hub        *realtime.Hub
	perms      models.PermissionChecker
	httpClient *http.Client

	avatarMu    sync.RWMutex
	avatarCache map[string]cachedAvatar
}

type cachedAvatar struct {
	data      []byte
	expiresAt time.Time
}

func NewService(store *Store, cfg *config.Config, hub *realtime.Hub, perms models.PermissionChecker) *Service {
	return &Service{
		store:       store,
		cfg:         cfg,
		hub:         hub,
		perms:       perms,
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		avatarCache: make(map[string]cachedAvatar),
	}
}

func (svc *Service) Store() *Store {
	return svc.store
}

// --- Status & Duration Calculations ---

type PlayerStatus struct {
	Social         string
	Gameplay       string
	SocialPoints   int
	GameplayPoints int
}

func (svc *Service) CalculateStatus(ctx context.Context, serverID string, punishments []*Punishment) PlayerStatus {
	types := svc.store.GetPunishmentTypes(ctx, serverID)
	typesByOrd := indexTypesByOrdinal(types)
	thresholds := svc.store.GetThresholdSettings(ctx, serverID)
	socialExpMs := thresholds.Social.PointExpiryMs()
	gameplayExpMs := thresholds.Gameplay.PointExpiryMs()
	now := httpx.NowMs()

	socialPts := 0
	gameplayPts := 0

	for _, p := range punishments {
		if !svc.isPunishmentEligible(p) {
			continue
		}
		pt, ok := typesByOrd[p.TypeOrdinal]
		if !ok {
			continue
		}
		if effExp := svc.GetEffectiveExpiry(p); effExp != nil {
			expWindow := gameplayExpMs
			if pt.IsSocial() {
				expWindow = socialExpMs
			}
			if *effExp+expWindow < now {
				continue
			}
		}
		sev := p.DataString(PunishmentKeySeverity)
		if sev == "" {
			sev = "regular"
		}
		pts := pt.PointsForSeverity(sev)
		if pt.IsSocial() {
			socialPts += pts
		} else if pt.IsGameplay() {
			gameplayPts += pts
		}
	}

	return PlayerStatus{
		Social:         thresholds.Social.OffenderLevel(socialPts),
		Gameplay:       thresholds.Gameplay.OffenderLevel(gameplayPts),
		SocialPoints:   socialPts,
		GameplayPoints: gameplayPts,
	}
}

func (svc *Service) isPunishmentEligible(p *Punishment) bool {
	if p == nil || p.TypeOrdinal == 0 {
		return false
	}
	if p.IsUnstarted() || p.IsPardoned() {
		return false
	}
	return true
}

func (svc *Service) IsPunishmentActive(p *Punishment) bool {
	if !svc.isPunishmentEligible(p) {
		return false
	}
	eff := svc.GetEffectiveExpiry(p)
	if eff == nil {
		return true
	}
	return *eff >= httpx.NowMs()
}

func (svc *Service) IsPunishmentNaturallyExpired(p *Punishment) bool {
	if p == nil || p.StartedMs == nil || p.TypeOrdinal == 0 || p.IsPardoned() {
		return false
	}
	eff := svc.GetEffectiveExpiry(p)
	if eff == nil {
		return false
	}
	return *eff < httpx.NowMs()
}

func (svc *Service) GetEffectiveExpiry(p *Punishment) *int64 {
	if p == nil || p.Data == nil {
		return nil
	}
	var latestDurMod *PunishmentModification
	for i := range p.Modifications {
		if IsDurationChangeModType(p.Modifications[i].Type) {
			latestDurMod = &p.Modifications[i]
		}
	}
	if latestDurMod != nil {
		if latestDurMod.EffectiveDuration == nil || *latestDurMod.EffectiveDuration <= 0 {
			return nil
		}
		exp := latestDurMod.DateMs + *latestDurMod.EffectiveDuration
		return &exp
	}
	dur, ok := p.DataInt64(PunishmentKeyDuration)
	if !ok || dur <= 0 {
		return nil
	}
	base := httpx.NowMs()
	if p.StartedMs != nil {
		base = *p.StartedMs
	}
	exp := base + dur
	return &exp
}

func (svc *Service) GetEffectiveCategory(p *Punishment, typesByOrd map[int]PunishmentType) string {
	if p == nil {
		return ""
	}
	if stored := p.DataString(PunishmentKeyEnforcementCategory); stored != "" {
		return stored
	}
	if p.DataString(PunishmentKeyLinkedBanID) != "" {
		return CategoryBan
	}
	pt, ok := typesByOrd[p.TypeOrdinal]
	if !ok {
		return ""
	}
	return svc.GetEffectiveCategoryForType(&pt, p.Data)
}

func (svc *Service) GetEffectiveCategoryForType(pt *PunishmentType, data map[string]any) string {
	if pt == nil {
		return ""
	}
	if data != nil {
		if stored, ok := data[PunishmentKeyEnforcementCategory].(string); ok && stored != "" {
			return stored
		}
		if lb, ok := data[PunishmentKeyLinkedBanID].(string); ok && lb != "" {
			return CategoryBan
		}
	}
	if pt.IsKick() {
		return ""
	}
	if pt.IsBan() {
		return CategoryBan
	}
	if pt.IsMute() {
		return CategoryMute
	}
	if data != nil {
		sev, _ := data[PunishmentKeySeverity].(string)
		if sev == "" {
			sev = "regular"
		}
		offLevel, _ := data[PunishmentKeyOffenseLevel].(string)
		if offLevel == "" {
			st, _ := data[PunishmentKeyStatus].(string)
			switch strings.ToLower(st) {
			case "medium":
				offLevel = "medium"
			case "habitual":
				offLevel = "habitual"
			default:
				offLevel = "first"
			}
		}
		if d := pt.GetDurationDetail(sev, offLevel); d != nil {
			if d.IsBan() {
				return CategoryBan
			}
			if d.IsMute() {
				return CategoryMute
			}
		}
	}
	return ""
}

func (svc *Service) CalculateDuration(ctx context.Context, serverID string, existing []*Punishment, typeOrdinal int, severity string) (durationMs *int64, displayStatus string, offenseLevel string) {
	types := svc.store.GetPunishmentTypes(ctx, serverID)
	typesByOrd := indexTypesByOrdinal(types)
	pt, ok := typesByOrd[typeOrdinal]
	if !ok {
		return nil, "", ""
	}
	thresholds := svc.store.GetThresholdSettings(ctx, serverID)
	st := svc.CalculateStatus(ctx, serverID, existing)
	relevantPts := st.GameplayPoints
	if pt.IsSocial() {
		relevantPts = st.SocialPoints
	}
	if pt.IsSocial() {
		offenseLevel = thresholds.Social.OffenseLevelInternal(relevantPts)
	} else {
		offenseLevel = thresholds.Gameplay.OffenseLevelInternal(relevantPts)
	}
	displayStatus = OffenseDisplayStatus(offenseLevel)
	detail := pt.GetDurationDetail(NormalizeSeverity(severity), offenseLevel)
	if detail != nil {
		ms := detail.ToMilliseconds()
		if ms != 0 {
			durationMs = &ms
		}
	}
	return durationMs, displayStatus, offenseLevel
}

// --- Permission Validation ---

func (svc *Service) ValidatePunishmentPermission(ctx context.Context, srv *models.Server, email string, typeOrdinal int) (int, string) {
	if srv == nil || strings.TrimSpace(email) == "" {
		return 403, "No authenticated user found for permission check"
	}
	if svc.perms.IsSuperAdmin(srv, email) {
		return 200, ""
	}
	types := svc.store.GetPunishmentTypes(ctx, srv.ID)
	pt, ok := indexTypesByOrdinal(types)[typeOrdinal]
	if !ok {
		return 400, "Invalid punishment type"
	}
	applyPerm := "punishment.apply." + strings.ReplaceAll(strings.ToLower(pt.Name), " ", "-")
	roleID := svc.perms.PanelRoleID(ctx, srv, email)
	if !svc.perms.HasPermission(ctx, srv, roleID, applyPerm) {
		return 403, "You do not have permission to apply this punishment type"
	}
	return 200, ""
}

// --- Punishment Lifecycle ---

var clientSettableDataKeys = map[string]bool{
	PunishmentKeyAltBlocking:       true,
	PunishmentKeyWipeAfterExpiry:   true,
	PunishmentKeyBanLinkedAccounts: true,
	PunishmentKeyKickSameIP:        true,
	PunishmentKeySilent:            true,
	PunishmentKeyLinkedBanID:       true,
	PunishmentKeySeverity:          true,
	PunishmentKeyStatus:            true,
	PunishmentKeyDuration:          true,
	PunishmentKeyReason:            true,
	PunishmentKeyAIGenerated:       true,
	PunishmentKeyPendingAck:        true,
}

func (svc *Service) CreatePunishment(ctx context.Context, srv *models.Server, rawUUID string, req CreatePunishmentInput, isMinecraft bool) (string, int, string) {
	if len(req.Reason) > MaxReasonLength {
		return "", 400, fmt.Sprintf("reason must be at most %d characters", MaxReasonLength)
	}
	if r, ok := req.Data[PunishmentKeyReason].(string); ok && len(r) > MaxReasonLength {
		return "", 400, fmt.Sprintf("reason must be at most %d characters", MaxReasonLength)
	}
	for _, ev := range req.Evidence {
		if ev.FileURL != "" && !ValidateSafeURL(ev.FileURL) {
			return "", 400, "Invalid evidence URL"
		}
	}

	pl, err := svc.store.FindPlayerByUUID(ctx, srv.ID, rawUUID)
	if err != nil {
		return "", 500, "Failed to load player"
	}
	if pl == nil {
		return "", 404, "Player not found"
	}

	now := httpx.NowMs()
	data := make(map[string]any)
	for k, v := range req.Data {
		if clientSettableDataKeys[k] {
			data[k] = v
		}
	}
	if req.Severity != "" {
		data[PunishmentKeySeverity] = req.Severity
	}
	if req.Status != "" {
		data[PunishmentKeyStatus] = req.Status
	}

	calcDuration := req.Duration
	if calcDuration == nil && req.Severity != "" {
		dur, dispStatus, offLevel := svc.CalculateDuration(ctx, srv.ID, pl.Punishments, req.TypeOrdinal, req.Severity)
		calcDuration = dur
		if dispStatus != "" {
			if existingSt, ok := data[PunishmentKeyStatus]; !ok || existingSt == nil || existingSt == "" {
				data[PunishmentKeyStatus] = dispStatus
			}
		}
		if offLevel != "" {
			if _, ok := data[PunishmentKeyOffenseLevel]; !ok {
				data[PunishmentKeyOffenseLevel] = offLevel
			}
		}
	}
	if calcDuration == nil {
		if d, ok := asInt64(data[PunishmentKeyDuration]); ok {
			calcDuration = &d
		}
	}
	if calcDuration != nil && *calcDuration != 0 {
		data[PunishmentKeyDuration] = *calcDuration
	}
	if strings.TrimSpace(req.Reason) != "" {
		data[PunishmentKeyReason] = req.Reason
	}

	types := svc.store.GetPunishmentTypes(ctx, srv.ID)
	typesByOrd := indexTypesByOrdinal(types)
	var newPT *PunishmentType
	if pt, ok := typesByOrd[req.TypeOrdinal]; ok {
		newPT = &pt
	}

	if newPT != nil {
		if newPT.PermanentUntilUsernameChange {
			if _, ok := data[PunishmentKeyBlockedName]; !ok {
				if uname := pl.LatestUsername(); uname != "" && uname != "Unknown" {
					data[PunishmentKeyBlockedName] = uname
				}
			}
		}
		if newPT.PermanentUntilSkinChange {
			if _, ok := data[PunishmentKeyBlockedSkin]; !ok {
				if sk := pl.LastSkinHash(); sk != "" {
					data[PunishmentKeyBlockedSkin] = sk
				}
			}
		}
	}

	newCategory := ""
	if newPT != nil {
		newCategory = svc.GetEffectiveCategoryForType(newPT, data)
	}
	if newCategory != "" {
		data[PunishmentKeyEnforcementCategory] = newCategory
		hasExisting := false
		for _, ex := range pl.Punishments {
			if svc.GetEffectiveCategory(ex, typesByOrd) != newCategory {
				continue
			}
			if svc.IsPunishmentActive(ex) || ex.IsUnstarted() {
				hasExisting = true
				break
			}
		}
		if hasExisting {
			data[PunishmentKeyStatus] = PunishmentStatusUnstarted
		}
	}

	if isMinecraft {
		delete(data, PunishmentKeyPendingAck)
		data[PunishmentKeyStatus] = PunishmentStatusUnstarted
	} else {
		if pendingAck, _ := data[PunishmentKeyPendingAck].(bool); pendingAck {
			delete(data, PunishmentKeyPendingAck)
			data[PunishmentKeyStatus] = PunishmentStatusUnstarted
		}
	}

	storedIssuerName := storedIssuer(req.IssuerID, req.IssuerName)
	notes := buildCreationNotes(req, now, storedIssuerName, req.IssuerID, newPT, newCategory, calcDuration)

	evidence := make([]PunishmentEvidence, 0, len(req.Evidence))
	for _, ev := range req.Evidence {
		evIssuerID := ev.IssuerID
		if evIssuerID == "" {
			evIssuerID = req.IssuerID
		}
		rawEvName := ev.IssuerName
		if rawEvName == "" {
			rawEvName = req.IssuerName
		}
		evIssuerName := storedIssuer(evIssuerID, rawEvName)
		evType := ev.Type
		if evType == "" {
			evType = "text"
		}
		var textPtr, urlPtr, byPtr, byIDPtr, fnPtr, ftPtr *string
		if ev.Text != "" {
			t := ev.Text
			textPtr = &t
		}
		if ev.FileURL != "" {
			u := ev.FileURL
			urlPtr = &u
		}
		if evIssuerName != "" {
			b := evIssuerName
			byPtr = &b
		}
		if evIssuerID != "" {
			bid := evIssuerID
			byIDPtr = &bid
		}
		if ev.FileName != "" {
			fn := ev.FileName
			fnPtr = &fn
		}
		if ev.FileType != "" {
			ft := ev.FileType
			ftPtr = &ft
		}
		evidence = append(evidence, PunishmentEvidence{
			Text:         textPtr,
			URL:          urlPtr,
			Type:         evType,
			UploadedBy:   byPtr,
			UploadedByID: byIDPtr,
			UploadedAtMs: now,
			FileName:     fnPtr,
			FileType:     ftPtr,
			FileSize:     ev.FileSize,
		})
	}

	punID := NewShortID()
	typeName := "Unknown"
	points := 0
	if newPT != nil {
		typeName = newPT.Name
		sev := req.Severity
		if sev == "" {
			sev = "regular"
		}
		points = newPT.PointsForSeverity(sev)
	}
	altBlock, _ := data[PunishmentKeyAltBlocking].(bool)
	wipeStats, _ := data[PunishmentKeyWipeAfterExpiry].(bool)
	silent, _ := data[PunishmentKeySilent].(bool)

	tickets := req.AttachedTicketIds
	if tickets == nil {
		tickets = []string{}
	}

	pun := &Punishment{
		ID:                punID,
		ServerID:          srv.ID,
		PlayerID:          pl.ID,
		PlayerUUID:        pl.MinecraftUUID,
		PlayerUsername:    pl.LatestUsername(),
		Type:              typeName,
		TypeOrdinal:       req.TypeOrdinal,
		IssuerName:        storedIssuerName,
		IssuerID:          req.IssuerID,
		IssuedMs:          now,
		DurationMs:        calcDuration,
		Points:            points,
		Silent:            silent,
		WipeStats:         wipeStats,
		BlockAlts:         altBlock,
		Modifications:     []PunishmentModification{},
		Notes:             notes,
		Evidence:          evidence,
		AttachedTicketIds: tickets,
		Data:              data,
	}

	if err := svc.store.InsertPunishment(ctx, pun); err != nil {
		return "", 500, "Failed to create punishment"
	}

	issuerDisplay := svc.store.ResolveIssuerName(ctx, srv.ID, req.IssuerID, req.IssuerName)
	if len(req.AttachedTicketIds) > 0 {
		svc.store.CloseTicketsForPunishment(ctx, srv.ID, req.AttachedTicketIds, issuerDisplay)
	}

	svc.store.RecordModerationLog(ctx, srv.ID, issuerDisplay, fmt.Sprintf("%s issued for %s", typeName, pl.LatestUsername()))

	pl.Punishments = append(pl.Punishments, pun)
	svc.publishPunishmentIssued(ctx, srv, pl, pun)

	durText := "Permanent"
	if calcDuration != nil {
		durText = FormatDuration(*calcDuration, *calcDuration < 0)
	}
	svc.sendWebhookAsync(ctx, srv, map[string]string{
		"id":         punID,
		"playerName": pl.LatestUsername(),
		"type":       typeName,
		"severity":   req.Severity,
		"reason":     req.Reason,
		"duration":   durText,
		"issuer":     issuerDisplay,
		"ticketId":   strings.Join(req.AttachedTicketIds, ", "),
	})

	return punID, 200, ""
}

func buildCreationNotes(req CreatePunishmentInput, now int64, storedIssuerName, issuerID string, pt *PunishmentType, category string, calcDur *int64) []PunishmentNote {
	enf := "punishment"
	if pt != nil && pt.IsKick() {
		enf = "kick"
	} else if category == CategoryBan {
		enf = "ban"
	} else if category == CategoryMute {
		enf = "mute"
	}
	var issuedNote string
	if enf == "kick" {
		issuedNote = "issued kick"
	} else if calcDur != nil && *calcDur > 0 {
		issuedNote = "issued " + FormatDuration(*calcDur, false) + " " + enf
	} else {
		issuedNote = "issued permanent " + enf
	}
	notes := []PunishmentNote{{
		ID:         NewShortID(),
		Text:       issuedNote,
		DateMs:     now,
		IssuerName: storedIssuerName,
		IssuerID:   issuerID,
	}}
	if strings.TrimSpace(req.Reason) != "" {
		notes = append(notes, PunishmentNote{
			ID:         NewShortID(),
			Text:       req.Reason,
			DateMs:     now,
			IssuerName: storedIssuerName,
			IssuerID:   issuerID,
		})
	}
	for _, nr := range req.Notes {
		nid := nr.IssuerID
		if nid == "" {
			nid = issuerID
		}
		nname := nr.IssuerName
		if nname == "" {
			nname = req.IssuerName
		}
		notes = append(notes, PunishmentNote{
			ID:         NewShortID(),
			Text:       nr.Text,
			DateMs:     now,
			IssuerName: storedIssuer(nid, nname),
			IssuerID:   nid,
		})
	}
	return notes
}

func (svc *Service) AcknowledgePunishment(ctx context.Context, srv *models.Server, rawUUID, punishmentID string) OperationResult {
	pl, err := svc.store.FindPlayerByUUID(ctx, srv.ID, rawUUID)
	if err != nil {
		return OperationResult{HTTPStatus: 500, Message: "Failed to load player"}
	}
	if pl == nil {
		return OperationResult{HTTPStatus: 404, Message: "Player not found: " + rawUUID}
	}
	var target *Punishment
	for _, p := range pl.Punishments {
		if p.ID == punishmentID {
			target = p
			break
		}
	}
	if target == nil {
		return OperationResult{HTTPStatus: 404, Message: "Punishment not found: " + punishmentID + " for player: " + rawUUID}
	}
	if target.StartedMs != nil && !target.IsUnstarted() {
		return OperationResult{HTTPStatus: 200, Success: true, Message: "Punishment already acknowledged", Count: 0}
	}
	now := httpx.NowMs()
	if target.StartedMs == nil {
		target.StartedMs = &now
	}
	if st, _ := target.Data[PunishmentKeyStatus].(string); strings.EqualFold(st, PunishmentStatusUnstarted) {
		delete(target.Data, PunishmentKeyStatus)
	}
	if dur, ok := target.DataInt64(PunishmentKeyDuration); ok && dur > 0 {
		exp := *target.StartedMs + dur
		target.ExpiresMs = &exp
	}
	if err := svc.store.UpdatePunishment(ctx, target); err != nil {
		return OperationResult{HTTPStatus: 500, Message: "Failed to acknowledge punishment"}
	}
	return OperationResult{HTTPStatus: 200, Success: true, Message: "Punishment acknowledged", Count: 1}
}

func (svc *Service) PardonPunishment(ctx context.Context, srv *models.Server, punishmentID, issuerName, issuerID, reason string) OperationResult {
	pl, pun, err := svc.store.FindPunishmentContext(ctx, srv.ID, punishmentID)
	if err != nil {
		return OperationResult{HTTPStatus: 500, Message: "Failed to load punishment"}
	}
	if pun == nil {
		return OperationResult{HTTPStatus: 404, Message: "Punishment not found"}
	}
	if pun.IsPardoned() {
		return OperationResult{HTTPStatus: 200, Success: false, Message: "Punishment has already been pardoned", Count: 0}
	}
	if err := svc.applyManualPardon(ctx, srv, pl, pun, issuerName, issuerID, reason); err != nil {
		return OperationResult{HTTPStatus: 500, Message: "Failed to pardon punishment"}
	}
	svc.publishPunishmentModified(ctx, srv, pl, pun)
	if pun.DataBool(PunishmentKeyAltBlocking) {
		svc.CascadePardonLinkedBans(ctx, srv, punishmentID)
	}
	return OperationResult{HTTPStatus: 200, Success: true, Message: "Punishment pardoned", Count: 1}
}

func (svc *Service) PardonPlayer(ctx context.Context, srv *models.Server, playerName, punishmentType, issuerName, issuerID, reason string) OperationResult {
	pl, err := svc.store.FindBestByUsername(ctx, srv.ID, playerName)
	if err != nil {
		return OperationResult{HTTPStatus: 500, Message: "Failed to look up player"}
	}
	if pl == nil {
		return OperationResult{HTTPStatus: 404, Message: "Player not found: " + playerName}
	}
	typesByOrd := indexTypesByOrdinal(svc.store.GetPunishmentTypes(ctx, srv.ID))
	var targets []*Punishment
	for _, p := range pl.Punishments {
		if !svc.IsPunishmentActive(p) {
			continue
		}
		if strings.TrimSpace(punishmentType) != "" {
			ptName := "Unknown"
			if pt, ok := typesByOrd[p.TypeOrdinal]; ok {
				ptName = pt.Name
			}
			cat := svc.GetEffectiveCategory(p, typesByOrd)
			if !strings.EqualFold(ptName, punishmentType) && !strings.EqualFold(cat, punishmentType) {
				continue
			}
		}
		targets = append(targets, p)
	}
	if len(targets) == 0 {
		return OperationResult{HTTPStatus: 200, Success: false, Message: "No active punishments to pardon", Count: 0}
	}
	altParents := make([]string, 0)
	for _, t := range targets {
		if err := svc.applyManualPardon(ctx, srv, pl, t, issuerName, issuerID, reason); err == nil {
			svc.publishPunishmentModified(ctx, srv, pl, t)
			if t.DataBool(PunishmentKeyAltBlocking) {
				altParents = append(altParents, t.ID)
			}
		}
	}
	for _, parentID := range altParents {
		svc.CascadePardonLinkedBans(ctx, srv, parentID)
	}
	return OperationResult{
		HTTPStatus: 200,
		Success:    true,
		Message:    fmt.Sprintf("Pardoned %d punishment(s)", len(targets)),
		Count:      len(targets),
	}
}

func (svc *Service) applyManualPardon(ctx context.Context, srv *models.Server, pl *Player, pun *Punishment, issuerName, issuerID, reason string) error {
	resolvedName := storedIssuer(issuerID, issuerName)
	now := httpx.NowMs()
	mod := PunishmentModification{
		ID:         NewShortID(),
		Type:       ModManualPardon,
		DateMs:     now,
		IssuerName: resolvedName,
		IssuerID:   issuerID,
		Reason:     reason,
	}
	notes := []PunishmentNote{{
		ID:         NewShortID(),
		Text:       "pardoned punishment",
		DateMs:     now,
		IssuerName: resolvedName,
		IssuerID:   issuerID,
	}}
	if strings.TrimSpace(reason) != "" {
		notes = append(notes, PunishmentNote{
			ID:         NewShortID(),
			Text:       reason,
			DateMs:     now,
			IssuerName: resolvedName,
			IssuerID:   issuerID,
		})
	}
	pun.Modifications = append(pun.Modifications, mod)
	pun.Notes = append(pun.Notes, notes...)
	pun.Pardoned = true
	pun.PardonedBy = svc.store.ResolveIssuerName(ctx, srv.ID, issuerID, issuerName)
	pun.PardonedAtMs = &now
	pun.PardonReason = reason
	pun.Data[PunishmentKeyStatus] = PunishmentStatusPardoned
	return svc.store.UpdatePunishment(ctx, pun)
}

func (svc *Service) applySystemPardon(ctx context.Context, srv *models.Server, pun *Punishment, reason string) error {
	now := httpx.NowMs()
	mod := PunishmentModification{
		ID:         NewShortID(),
		Type:       ModSystemPardon,
		DateMs:     now,
		IssuerName: SystemIssuer,
		Reason:     reason,
	}
	note := PunishmentNote{
		ID:         NewShortID(),
		Text:       reason,
		DateMs:     now,
		IssuerName: SystemIssuer,
	}
	pun.Modifications = append(pun.Modifications, mod)
	pun.Notes = append(pun.Notes, note)
	pun.Pardoned = true
	pun.PardonedBy = SystemIssuer
	pun.PardonedAtMs = &now
	pun.PardonReason = reason
	pun.Data[PunishmentKeyStatus] = PunishmentStatusPardoned
	return svc.store.UpdatePunishment(ctx, pun)
}

func (svc *Service) CascadePardonLinkedBans(ctx context.Context, srv *models.Server, parentPunishmentID string) int {
	players, err := svc.store.FindPlayersByLinkedBanID(ctx, srv.ID, parentPunishmentID)
	if err != nil {
		return 0
	}
	count := 0
	for _, pl := range players {
		for _, pun := range pl.Punishments {
			if pun.DataString(PunishmentKeyLinkedBanID) != parentPunishmentID || !svc.IsPunishmentActive(pun) {
				continue
			}
			if err := svc.applySystemPardon(ctx, srv, pun, "Auto-pardoned: parent ban was pardoned"); err == nil {
				svc.publishPunishmentModified(ctx, srv, pl, pun)
				count++
			}
		}
	}
	return count
}

func (svc *Service) CascadeDurationChangeToLinkedBans(ctx context.Context, srv *models.Server, parentPunishmentID string, newDuration *int64) int {
	players, err := svc.store.FindPlayersByLinkedBanID(ctx, srv.ID, parentPunishmentID)
	if err != nil {
		return 0
	}
	count := 0
	now := httpx.NowMs()
	eff := int64(-1)
	if newDuration != nil {
		eff = *newDuration
	}
	for _, pl := range players {
		for _, pun := range pl.Punishments {
			if pun.DataString(PunishmentKeyLinkedBanID) != parentPunishmentID || !svc.IsPunishmentActive(pun) {
				continue
			}
			mod := PunishmentModification{
				ID:                NewShortID(),
				Type:              ModManualDurationChange,
				DateMs:            now,
				IssuerName:        SystemIssuer,
				Reason:            "Cascaded from parent ban duration change",
				EffectiveDuration: &eff,
			}
			note := PunishmentNote{
				ID:         NewShortID(),
				Text:       "Duration changed (cascaded from parent ban)",
				DateMs:     now,
				IssuerName: SystemIssuer,
			}
			pun.Data[PunishmentKeyDuration] = eff
			pun.DurationMs = &eff
			if eff > 0 {
				exp := now + eff
				pun.ExpiresMs = &exp
			} else {
				pun.ExpiresMs = nil
			}
			if pun.StartedMs == nil {
				pun.StartedMs = &now
			}
			pun.Modifications = append(pun.Modifications, mod)
			pun.Notes = append(pun.Notes, note)
			if err := svc.store.UpdatePunishment(ctx, pun); err == nil {
				svc.publishPunishmentModified(ctx, srv, pl, pun)
				count++
			}
		}
	}
	return count
}

func (svc *Service) PromoteUnstartedPunishments(ctx context.Context, srv *models.Server, pl *Player) []string {
	typesByOrd := indexTypesByOrdinal(svc.store.GetPunishmentTypes(ctx, srv.ID))
	var promotedIDs []string

	for _, cat := range []string{CategoryBan, CategoryMute} {
		hasActive := false
		for _, p := range pl.Punishments {
			if svc.GetEffectiveCategory(p, typesByOrd) == cat && svc.IsPunishmentActive(p) {
				hasActive = true
				break
			}
		}
		if hasActive {
			continue
		}
		var oldest *Punishment
		for _, p := range pl.Punishments {
			if svc.GetEffectiveCategory(p, typesByOrd) == cat && p.IsUnstarted() {
				if oldest == nil || p.IssuedMs < oldest.IssuedMs {
					oldest = p
				}
			}
		}
		if oldest != nil {
			delete(oldest.Data, PunishmentKeyStatus)
			if err := svc.store.UpdatePunishment(ctx, oldest); err == nil {
				promotedIDs = append(promotedIDs, oldest.ID)
				svc.publishPunishmentIssued(ctx, srv, pl, oldest)
			}
		}
	}
	return promotedIDs
}

func (svc *Service) EnforceAltBlockingBans(ctx context.Context, srv *models.Server, pl *Player) []string {
	linkedUUIDs := pl.LinkedAccountUUIDs()
	if len(linkedUUIDs) == 0 {
		return nil
	}
	typesByOrd := indexTypesByOrdinal(svc.store.GetPunishmentTypes(ctx, srv.ID))
	for _, p := range pl.Punishments {
		if svc.GetEffectiveCategory(p, typesByOrd) == CategoryBan && svc.IsPunishmentActive(p) {
			return nil
		}
	}
	linkedPlayers, err := svc.store.FindPlayersByUUIDs(ctx, srv.ID, linkedUUIDs)
	if err != nil {
		return nil
	}
	now := httpx.NowMs()
	for _, lp := range linkedPlayers {
		for _, p := range lp.Punishments {
			if !svc.IsPunishmentActive(p) || !p.DataBool(PunishmentKeyAltBlocking) {
				continue
			}
			if svc.GetEffectiveCategory(p, typesByOrd) != CategoryBan {
				continue
			}
			var linkedDur *int64
			if exp := svc.GetEffectiveExpiry(p); exp != nil {
				rem := *exp - now
				if rem <= 0 {
					continue
				}
				linkedDur = &rem
			}
			id, err := svc.createLinkedBanPunishment(ctx, srv, pl, p.ID, lp.MinecraftUUID, linkedDur)
			if err == nil && id != "" {
				return []string{id}
			}
			return nil
		}
	}
	return nil
}

func (svc *Service) createLinkedBanPunishment(ctx context.Context, srv *models.Server, pl *Player, parentPunishmentID, parentPlayerUUID string, dur *int64) (string, error) {
	now := httpx.NowMs()
	data := map[string]any{
		PunishmentKeyLinkedBanID:         parentPunishmentID,
		PunishmentKeyLinkedBanParentUUID: parentPlayerUUID,
		PunishmentKeyEnforcementCategory: CategoryBan,
	}
	var expPtr *int64
	if dur != nil {
		data[PunishmentKeyDuration] = *dur
		if *dur > 0 {
			exp := now + *dur
			expPtr = &exp
		}
	}
	noteDur := "issued permanent ban"
	if dur != nil && *dur > 0 {
		noteDur = "issued " + FormatDuration(*dur, false) + " ban"
	}
	pun := &Punishment{
		ID:             NewShortID(),
		ServerID:       srv.ID,
		PlayerID:       pl.ID,
		PlayerUUID:     pl.MinecraftUUID,
		PlayerUsername: pl.LatestUsername(),
		Type:           "Linked Ban",
		TypeOrdinal:    LinkedBanTypeOrdinal,
		IssuerName:     SystemIssuer,
		IssuedMs:       now,
		StartedMs:      &now,
		ExpiresMs:      expPtr,
		DurationMs:     dur,
		Notes: []PunishmentNote{
			{ID: NewShortID(), Text: noteDur, DateMs: now, IssuerName: SystemIssuer},
			{ID: NewShortID(), Text: "Automatically issued linked ban due to alt-blocking ban on linked account", DateMs: now, IssuerName: SystemIssuer},
		},
		Modifications:     []PunishmentModification{},
		Evidence:          []PunishmentEvidence{},
		AttachedTicketIds: []string{},
		Data:              data,
	}
	if err := svc.store.InsertPunishment(ctx, pun); err != nil {
		return "", err
	}
	pl.Punishments = append(pl.Punishments, pun)
	return pun.ID, nil
}

func (svc *Service) CheckRestrictionAutoPardons(ctx context.Context, srv *models.Server, pl *Player, currentUsername, currentSkinHash string) []string {
	typesByOrd := indexTypesByOrdinal(svc.store.GetPunishmentTypes(ctx, srv.ID))
	var pardoned []string
	for _, p := range pl.Punishments {
		if !svc.IsPunishmentActive(p) {
			continue
		}
		pt, ok := typesByOrd[p.TypeOrdinal]
		if !ok {
			continue
		}
		if pt.PermanentUntilUsernameChange && currentUsername != "" {
			blockedName := p.DataString(PunishmentKeyBlockedName)
			if blockedName != "" && !strings.EqualFold(blockedName, currentUsername) {
				reason := fmt.Sprintf("Auto-pardoned: username changed from '%s' to '%s'", blockedName, currentUsername)
				if err := svc.applySystemPardon(ctx, srv, p, reason); err == nil {
					svc.publishPunishmentModified(ctx, srv, pl, p)
					pardoned = append(pardoned, p.ID)
				}
			}
		}
		if pt.PermanentUntilSkinChange && currentSkinHash != "" {
			blockedSkin := p.DataString(PunishmentKeyBlockedSkin)
			if blockedSkin != "" && blockedSkin != currentSkinHash {
				if err := svc.applySystemPardon(ctx, srv, p, "Auto-pardoned: skin changed"); err == nil {
					svc.publishPunishmentModified(ctx, srv, pl, p)
					pardoned = append(pardoned, p.ID)
				}
			}
		}
	}
	return pardoned
}

// --- Punishment Mutations ---

func (svc *Service) AddModification(ctx context.Context, srv *models.Server, rawUUID, punishmentID string, req AddModificationInput) (*Player, int, string) {
	pl, err := svc.store.FindPlayerByUUID(ctx, srv.ID, rawUUID)
	if err != nil {
		return nil, 500, "Failed to load player"
	}
	if pl == nil {
		return nil, 404, "Player not found"
	}
	var target *Punishment
	for _, p := range pl.Punishments {
		if p.ID == punishmentID {
			target = p
			break
		}
	}
	if target == nil {
		return nil, 404, "Punishment not found"
	}
	now := httpx.NowMs()
	modIssuerName := storedIssuer(req.IssuerID, req.IssuerName)
	mod := PunishmentModification{
		ID:                NewShortID(),
		Type:              req.Type,
		DateMs:            now,
		IssuerName:        modIssuerName,
		IssuerID:          req.IssuerID,
		Reason:            req.Reason,
		EffectiveDuration: req.EffectiveDuration,
		AppealTicketID:    req.AppealTicketID,
	}
	target.Modifications = append(target.Modifications, mod)
	if IsPardonModType(req.Type) {
		target.Pardoned = true
		target.PardonedBy = svc.store.ResolveIssuerName(ctx, srv.ID, req.IssuerID, req.IssuerName)
		target.PardonedAtMs = &now
		target.PardonReason = req.Reason
		target.Data[PunishmentKeyStatus] = PunishmentStatusPardoned
	}
	if req.EffectiveDuration != nil {
		target.Data[PunishmentKeyDuration] = *req.EffectiveDuration
		target.DurationMs = req.EffectiveDuration
		if target.StartedMs == nil {
			target.StartedMs = &now
		}
		if *req.EffectiveDuration > 0 {
			exp := now + *req.EffectiveDuration
			target.ExpiresMs = &exp
		} else {
			target.ExpiresMs = nil
		}
	}
	if outcome := AppealOutcomeForModType(req.Type); outcome != "" && req.AppealTicketID != "" {
		svc.store.ApplyAppealOutcome(ctx, srv.ID, req.AppealTicketID, punishmentID, outcome)
	}
	if err := svc.store.UpdatePunishment(ctx, target); err != nil {
		return nil, 500, "Failed to update punishment"
	}
	svc.publishPunishmentModified(ctx, srv, pl, target)
	return pl, 200, ""
}

func (svc *Service) ChangeDuration(ctx context.Context, srv *models.Server, punishmentID string, newDuration *int64, issuerName, issuerID string) OperationResult {
	pl, pun, err := svc.store.FindPunishmentContext(ctx, srv.ID, punishmentID)
	if err != nil {
		return OperationResult{HTTPStatus: 500, Message: "Failed to load punishment"}
	}
	if pun == nil {
		return OperationResult{HTTPStatus: 404, Message: "Punishment not found"}
	}
	now := httpx.NowMs()
	resolvedName := storedIssuer(issuerID, issuerName)
	eff := int64(-1)
	if newDuration != nil {
		eff = *newDuration
	}
	mod := PunishmentModification{
		ID:                NewShortID(),
		Type:              ModManualDurationChange,
		DateMs:            now,
		IssuerName:        resolvedName,
		IssuerID:          issuerID,
		Reason:            "Duration changed",
		EffectiveDuration: &eff,
	}
	durText := "permanent"
	if newDuration != nil && *newDuration >= 0 {
		durText = FormatDuration(*newDuration, false)
	}
	note := PunishmentNote{
		ID:         NewShortID(),
		Text:       "changed duration to " + durText,
		DateMs:     now,
		IssuerName: resolvedName,
		IssuerID:   issuerID,
	}
	pun.Modifications = append(pun.Modifications, mod)
	pun.Notes = append(pun.Notes, note)
	pun.Data[PunishmentKeyDuration] = eff
	pun.DurationMs = &eff
	if pun.StartedMs == nil {
		pun.StartedMs = &now
	}
	if eff > 0 {
		exp := now + eff
		pun.ExpiresMs = &exp
	} else {
		pun.ExpiresMs = nil
	}
	if err := svc.store.UpdatePunishment(ctx, pun); err != nil {
		return OperationResult{HTTPStatus: 500, Message: "Failed to update punishment"}
	}
	svc.publishPunishmentModified(ctx, srv, pl, pun)

	if pun.DataBool(PunishmentKeyAltBlocking) {
		cascaded := svc.CascadeDurationChangeToLinkedBans(ctx, srv, punishmentID, newDuration)
		if cascaded > 0 {
			plural := ""
			if cascaded > 1 {
				plural = "s"
			}
			return OperationResult{
				HTTPStatus: 200,
				Success:    true,
				Message:    fmt.Sprintf("Duration changed (cascaded to %d linked ban%s)", cascaded, plural),
				Count:      cascaded + 1,
			}
		}
	}
	return OperationResult{HTTPStatus: 200, Success: true, Message: "Duration changed", Count: 1}
}

func (svc *Service) ToggleOption(ctx context.Context, srv *models.Server, punishmentID, option string, enabled bool, issuerName, issuerID string) OperationResult {
	var dataKey, displayName string
	switch strings.ToLower(strings.TrimSpace(option)) {
	case "altblocking", "alt_blocking", "alt-blocking":
		dataKey = PunishmentKeyAltBlocking
		displayName = "alt blocking"
	case "wipeafterexpiry", "wipe_after_expiry", "statwipe", "stat_wipe", "stat-wipe", "statwiping":
		dataKey = PunishmentKeyWipeAfterExpiry
		displayName = "stat wipe"
	default:
		return OperationResult{HTTPStatus: 400, Message: "Invalid option"}
	}
	pl, pun, err := svc.store.FindPunishmentContext(ctx, srv.ID, punishmentID)
	if err != nil {
		return OperationResult{HTTPStatus: 500, Message: "Failed to load punishment"}
	}
	if pun == nil {
		return OperationResult{HTTPStatus: 404, Message: "Punishment not found"}
	}
	now := httpx.NowMs()
	resolvedName := storedIssuer(issuerID, issuerName)
	action := "disabled "
	if enabled {
		action = "enabled "
	}
	note := PunishmentNote{
		ID:         NewShortID(),
		Text:       action + displayName,
		DateMs:     now,
		IssuerName: resolvedName,
		IssuerID:   issuerID,
	}
	pun.Data[dataKey] = enabled
	if dataKey == PunishmentKeyAltBlocking {
		pun.BlockAlts = enabled
	} else if dataKey == PunishmentKeyWipeAfterExpiry {
		pun.WipeStats = enabled
	}
	pun.Notes = append(pun.Notes, note)
	if err := svc.store.UpdatePunishment(ctx, pun); err != nil {
		return OperationResult{HTTPStatus: 500, Message: "Failed to update punishment"}
	}
	svc.publishPunishmentModified(ctx, srv, pl, pun)
	return OperationResult{HTTPStatus: 200, Success: true, Message: "Option toggled", Count: 1}
}

func (svc *Service) AcknowledgeStatWipe(ctx context.Context, srv *models.Server, punishmentID string) OperationResult {
	_, pun, err := svc.store.FindPunishmentContext(ctx, srv.ID, punishmentID)
	if err != nil {
		return OperationResult{HTTPStatus: 500, Message: "Failed to load punishment"}
	}
	if pun == nil {
		return OperationResult{HTTPStatus: 404, Message: "Punishment not found"}
	}
	if !pun.DataBool(PunishmentKeyWipeAfterExpiry) {
		return OperationResult{HTTPStatus: 200, Success: false, Message: "Stat wipe no longer enabled for this punishment", Count: 0}
	}
	pun.Data[PunishmentKeyStatWipeCompleted] = true
	if err := svc.store.UpdatePunishment(ctx, pun); err != nil {
		return OperationResult{HTTPStatus: 500, Message: "Failed to acknowledge stat wipe"}
	}
	return OperationResult{HTTPStatus: 200, Success: true, Message: "Stat wipe acknowledged", Count: 1}
}

func (svc *Service) ModifyPunishmentTickets(ctx context.Context, srv *models.Server, rawUUID, punishmentID string, req ModifyTicketsInput) OperationResult {
	var (
		pl  *Player
		pun *Punishment
		err error
	)
	if rawUUID != "" {
		pl, err = svc.store.FindPlayerByUUID(ctx, srv.ID, rawUUID)
		if err != nil {
			return OperationResult{HTTPStatus: 500, Message: "Failed to load player"}
		}
		if pl == nil {
			return OperationResult{HTTPStatus: 404, Message: "Player not found"}
		}
		for _, p := range pl.Punishments {
			if p.ID == punishmentID {
				pun = p
				break
			}
		}
	} else {
		pl, pun, err = svc.store.FindPunishmentContext(ctx, srv.ID, punishmentID)
		if err != nil {
			return OperationResult{HTTPStatus: 500, Message: "Failed to load punishment"}
		}
	}
	if pun == nil {
		return OperationResult{HTTPStatus: 404, Message: "Failed to modify punishment tickets"}
	}

	origSet := make(map[string]bool, len(pun.AttachedTicketIds))
	current := make([]string, 0, len(pun.AttachedTicketIds)+len(req.AddTicketIds))
	for _, id := range pun.AttachedTicketIds {
		origSet[id] = true
		current = append(current, id)
	}
	removeSet := make(map[string]bool, len(req.RemoveTicketIds))
	for _, id := range req.RemoveTicketIds {
		removeSet[id] = true
	}
	var addedIDs []string
	for _, id := range req.AddTicketIds {
		if id == "" || removeSet[id] {
			continue
		}
		if !containsString(current, id) {
			current = append(current, id)
		}
		if !origSet[id] {
			addedIDs = append(addedIDs, id)
		}
	}
	var filtered []string
	var removedIDs []string
	for _, id := range current {
		if removeSet[id] {
			if origSet[id] {
				removedIDs = append(removedIDs, id)
			}
			continue
		}
		filtered = append(filtered, id)
	}
	if filtered == nil {
		filtered = []string{}
	}
	pun.AttachedTicketIds = filtered
	if err := svc.store.UpdatePunishment(ctx, pun); err != nil {
		return OperationResult{HTTPStatus: 500, Message: "Failed to update punishment tickets"}
	}
	if req.ModifyAssociatedTickets {
		issuer := svc.store.ResolveIssuerName(ctx, srv.ID, req.IssuerID, req.IssuerName)
		if len(addedIDs) > 0 {
			svc.store.CloseTicketsForPunishment(ctx, srv.ID, addedIDs, issuer)
		}
		if len(removedIDs) > 0 {
			svc.store.ReopenTicketsForPunishment(ctx, srv.ID, removedIDs, issuer)
		}
	}
	if svc.hub != nil {
		svc.hub.InvalidatePanel(srv.ID, modlv1.PanelResource_PANEL_RESOURCE_PUNISHMENTS, pun.ID)
	}
	return OperationResult{HTTPStatus: 200, Success: true, Message: "Punishment tickets modified", Count: 1}
}

func (svc *Service) AddPunishmentNote(ctx context.Context, srv *models.Server, rawUUID, punishmentID, text, issuerName, issuerID string) OperationResult {
	var (
		pl  *Player
		pun *Punishment
		err error
	)
	if rawUUID != "" {
		pl, err = svc.store.FindPlayerByUUID(ctx, srv.ID, rawUUID)
		if err != nil {
			return OperationResult{HTTPStatus: 500, Message: "Failed to load player"}
		}
		if pl == nil {
			return OperationResult{HTTPStatus: 404, Message: "Player not found"}
		}
		for _, p := range pl.Punishments {
			if p.ID == punishmentID {
				pun = p
				break
			}
		}
	} else {
		pl, pun, err = svc.store.FindPunishmentContext(ctx, srv.ID, punishmentID)
		if err != nil {
			return OperationResult{HTTPStatus: 500, Message: "Failed to load punishment"}
		}
	}
	if pun == nil {
		return OperationResult{HTTPStatus: 404, Message: "Punishment not found"}
	}
	note := PunishmentNote{
		ID:         NewShortID(),
		Text:       text,
		DateMs:     httpx.NowMs(),
		IssuerName: storedIssuer(issuerID, issuerName),
		IssuerID:   issuerID,
	}
	pun.Notes = append(pun.Notes, note)
	if err := svc.store.UpdatePunishment(ctx, pun); err != nil {
		return OperationResult{HTTPStatus: 500, Message: "Failed to save note"}
	}
	if svc.hub != nil {
		svc.hub.InvalidatePanel(srv.ID, modlv1.PanelResource_PANEL_RESOURCE_PUNISHMENTS, pun.ID)
	}
	return OperationResult{HTTPStatus: 200, Success: true, Message: "Note added", Count: 1}
}

func (svc *Service) AddPunishmentEvidence(ctx context.Context, srv *models.Server, rawUUID, punishmentID string, req AddEvidenceInput, addSystemNote bool) OperationResult {
	if req.URL != "" && !ValidateSafeURL(req.URL) {
		return OperationResult{HTTPStatus: 400, Message: "Invalid evidence URL"}
	}
	var (
		pl  *Player
		pun *Punishment
		err error
	)
	if rawUUID != "" {
		pl, err = svc.store.FindPlayerByUUID(ctx, srv.ID, rawUUID)
		if err != nil {
			return OperationResult{HTTPStatus: 500, Message: "Failed to load player"}
		}
		if pl == nil {
			return OperationResult{HTTPStatus: 404, Message: "Player not found"}
		}
		for _, p := range pl.Punishments {
			if p.ID == punishmentID {
				pun = p
				break
			}
		}
	} else {
		pl, pun, err = svc.store.FindPunishmentContext(ctx, srv.ID, punishmentID)
		if err != nil {
			return OperationResult{HTTPStatus: 500, Message: "Failed to load punishment"}
		}
	}
	if pun == nil {
		return OperationResult{HTTPStatus: 404, Message: "Punishment not found"}
	}
	now := httpx.NowMs()
	rawIssuerName := req.IssuerName
	if rawIssuerName == "" {
		rawIssuerName = SystemIssuer
	}
	storedName := storedIssuer(req.IssuerID, rawIssuerName)
	evType := req.Type
	if evType == "" {
		evType = "url"
	}
	ev := PunishmentEvidence{
		Type:         evType,
		UploadedAtMs: now,
		FileSize:     req.FileSize,
	}
	if req.Text != "" {
		t := req.Text
		ev.Text = &t
	}
	if req.URL != "" {
		u := req.URL
		ev.URL = &u
	}
	if storedName != "" {
		b := storedName
		ev.UploadedBy = &b
	}
	if req.IssuerID != "" {
		bid := req.IssuerID
		ev.UploadedByID = &bid
	}
	if req.FileName != "" {
		fn := req.FileName
		ev.FileName = &fn
	}
	if req.FileType != "" {
		ft := req.FileType
		ev.FileType = &ft
	}
	pun.Evidence = append(pun.Evidence, ev)
	if addSystemNote {
		pun.Notes = append(pun.Notes, PunishmentNote{
			ID:         NewShortID(),
			Text:       "added evidence",
			DateMs:     now,
			IssuerName: storedName,
			IssuerID:   req.IssuerID,
		})
	}
	if err := svc.store.UpdatePunishment(ctx, pun); err != nil {
		return OperationResult{HTTPStatus: 500, Message: "Failed to add evidence"}
	}
	if svc.hub != nil {
		svc.hub.InvalidatePanel(srv.ID, modlv1.PanelResource_PANEL_RESOURCE_PUNISHMENTS, pun.ID)
	}
	return OperationResult{HTTPStatus: 200, Success: true, Message: "Evidence added", Count: 1}
}

// --- Punishment Queries & Views ---

func (svc *Service) BuildPunishmentResponseView(ctx context.Context, serverID string, pun *Punishment, pl *Player, resolvedIssuers map[string]string) PunishmentResponseView {
	types := svc.store.GetPunishmentTypes(ctx, serverID)
	typesByOrd := indexTypesByOrdinal(types)
	pt, hasPT := typesByOrd[pun.TypeOrdinal]
	typeName := "Unknown"
	isAppealable := false
	var ptPtr *PunishmentType
	if hasPT {
		typeName = pt.Name
		isAppealable = pt.Appealable
		ptPtr = &pt
	} else if pun.Type != "" {
		typeName = pun.Type
	}
	effCat := svc.GetEffectiveCategoryForType(ptPtr, pun.Data)
	active := svc.IsPunishmentActive(pun)
	exp := svc.GetEffectiveExpiry(pun)

	var pUUID, pName *string
	if pl != nil {
		u := pl.MinecraftUUID
		n := pl.LatestUsername()
		pUUID = &u
		pName = &n
	}

	var reasonPtr, sevPtr, statusPtr, effCatPtr *string
	if r := pun.DataString(PunishmentKeyReason); r != "" {
		reasonPtr = &r
	}
	if s := pun.DataString(PunishmentKeySeverity); s != "" {
		sevPtr = &s
	}
	if st := resolveOffenderStatus(pun); st != "" {
		statusPtr = &st
	}
	if effCat != "" {
		effCatPtr = &effCat
	}
	altBlock := pun.DataBool(PunishmentKeyAltBlocking)
	wipeStats := pun.DataBool(PunishmentKeyWipeAfterExpiry)

	mods := make([]PunishmentModification, len(pun.Modifications))
	for i, m := range pun.Modifications {
		mods[i] = m
		mods[i].IssuerName = resolveIssuer(m.IssuerID, m.IssuerName, resolvedIssuers)
	}
	notes := make([]PunishmentNote, len(pun.Notes))
	for i, n := range pun.Notes {
		notes[i] = n
		notes[i].IssuerName = resolveIssuer(n.IssuerID, n.IssuerName, resolvedIssuers)
	}
	evs := make([]PunishmentEvidence, len(pun.Evidence))
	for i, e := range pun.Evidence {
		evs[i] = e
		id := ""
		if e.UploadedByID != nil {
			id = *e.UploadedByID
		}
		name := ""
		if e.UploadedBy != nil {
			name = *e.UploadedBy
		}
		resolved := resolveIssuer(id, name, resolvedIssuers)
		evs[i].UploadedBy = &resolved
	}

	return PunishmentResponseView{
		ID:                pun.ID,
		Type:              typeName,
		TypeOrdinal:       pun.TypeOrdinal,
		IssuerName:        resolveIssuer(pun.IssuerID, pun.IssuerName, resolvedIssuers),
		IssuedMs:          pun.IssuedMs,
		StartedMs:         pun.StartedMs,
		IsAppealable:      isAppealable,
		Reason:            reasonPtr,
		Severity:          sevPtr,
		Status:            statusPtr,
		Active:            active,
		ExpiresMs:         exp,
		PlayerUUID:        pUUID,
		PlayerUsername:    pName,
		AltBlocking:       &altBlock,
		StatWiping:        &wipeStats,
		EffectiveCategory: effCatPtr,
		Modifications:     mods,
		Notes:             notes,
		Evidence:          evs,
		AttachedTicketIds: pun.AttachedTicketIds,
	}
}

func resolveOffenderStatus(pun *Punishment) string {
	st := pun.DataString(PunishmentKeyStatus)
	if st != "" && !strings.EqualFold(st, PunishmentStatusUnstarted) && !strings.EqualFold(st, PunishmentStatusPardoned) {
		return st
	}
	if ol := pun.DataString(PunishmentKeyOffenseLevel); ol != "" {
		return OffenseDisplayStatus(ol)
	}
	return ""
}

func (svc *Service) ToPunishmentView(pun *Punishment, typesByOrd map[int]PunishmentType, resolvedIssuers map[string]string, pl *Player) PunishmentView {
	typeName := "Unknown"
	if pt, ok := typesByOrd[pun.TypeOrdinal]; ok {
		typeName = pt.Name
	} else if pun.Type != "" {
		typeName = pun.Type
	}
	dataCopy := make(map[string]any, len(pun.Data)+1)
	for k, v := range pun.Data {
		dataCopy[k] = v
	}
	if typeName != "Unknown" {
		dataCopy["typeName"] = typeName
	}

	mods := make([]map[string]any, 0, len(pun.Modifications))
	for _, m := range pun.Modifications {
		mods = append(mods, map[string]any{
			"id":                m.ID,
			"type":              m.Type,
			"date":              m.DateMs,
			"issuerName":        resolveIssuer(m.IssuerID, m.IssuerName, resolvedIssuers),
			"effectiveDuration": m.EffectiveDuration,
			"data":              m.Data,
		})
	}
	notes := make([]map[string]any, 0, len(pun.Notes))
	for _, n := range pun.Notes {
		notes = append(notes, map[string]any{
			"id":         n.ID,
			"text":       n.Text,
			"issuerName": resolveIssuer(n.IssuerID, n.IssuerName, resolvedIssuers),
			"date":       n.DateMs,
		})
	}
	evs := make([]map[string]any, 0, len(pun.Evidence))
	for _, e := range pun.Evidence {
		uid, uname := "", ""
		if e.UploadedByID != nil {
			uid = *e.UploadedByID
		}
		if e.UploadedBy != nil {
			uname = *e.UploadedBy
		}
		evs = append(evs, map[string]any{
			"text":       e.Text,
			"url":        e.URL,
			"type":       e.Type,
			"uploadedBy": resolveIssuer(uid, uname, resolvedIssuers),
			"uploadedAt": e.UploadedAtMs,
			"fileName":   e.FileName,
			"fileType":   e.FileType,
			"fileSize":   e.FileSize,
		})
	}

	issued := pun.IssuedMs
	var pUUID, pName *string
	if pl != nil {
		u := pl.MinecraftUUID
		n := pl.LatestUsername()
		pUUID = &u
		pName = &n
	}
	tickets := pun.AttachedTicketIds
	if tickets == nil {
		tickets = []string{}
	}
	return PunishmentView{
		ID:                pun.ID,
		IssuerName:        resolveIssuer(pun.IssuerID, pun.IssuerName, resolvedIssuers),
		Issued:            &issued,
		Started:           pun.StartedMs,
		TypeOrdinal:       pun.TypeOrdinal,
		Type:              typeName,
		Modifications:     mods,
		Notes:             notes,
		Evidence:          evs,
		AttachedTicketIds: tickets,
		Data:              dataCopy,
		PlayerUUID:        pUUID,
		PlayerName:        pName,
	}
}

func (svc *Service) ToSimplePunishment(pun *Punishment, typesByOrd map[int]PunishmentType, resolvedIssuers map[string]string) SimplePunishmentView {
	typeName := "Unknown"
	var playerDesc *string
	var ptPtr *PunishmentType
	if pt, ok := typesByOrd[pun.TypeOrdinal]; ok {
		typeName = pt.Name
		ptPtr = &pt
		if pt.PlayerDescription != "" {
			d := pt.PlayerDescription
			playerDesc = &d
		}
	} else if pun.Type != "" {
		typeName = pun.Type
	}
	cat := svc.GetEffectiveCategoryForType(ptPtr, pun.Data)
	if cat == "" {
		cat = "OTHER"
	}
	reason := pun.DataString(PunishmentKeyReason)
	if strings.TrimSpace(reason) == "" {
		for _, n := range pun.Notes {
			if !IsAutoGeneratedNote(n.Text) {
				reason = n.Text
				break
			}
		}
	}
	if strings.TrimSpace(reason) == "" {
		reason = "No reason specified"
	}
	mods := make([]SimplePunishmentModView, 0, len(pun.Modifications))
	for _, m := range pun.Modifications {
		var eff int64
		if m.EffectiveDuration != nil {
			eff = *m.EffectiveDuration
		}
		var ts *int64
		if m.DateMs != 0 {
			d := m.DateMs
			ts = &d
		}
		mods = append(mods, SimplePunishmentModView{
			Type:              m.Type,
			Timestamp:         ts,
			EffectiveDuration: eff,
			IssuerName:        resolveIssuer(m.IssuerID, m.IssuerName, resolvedIssuers),
		})
	}
	return SimplePunishmentView{
		ID:                pun.ID,
		Type:              typeName,
		Category:          cat,
		TypeOrdinal:       pun.TypeOrdinal,
		Ordinal:           pun.TypeOrdinal,
		Started:           pun.StartedMs != nil,
		Expiration:        svc.GetEffectiveExpiry(pun),
		Description:       reason,
		IssuerName:        resolveIssuer(pun.IssuerID, pun.IssuerName, resolvedIssuers),
		IssuedAt:          pun.IssuedMs,
		PlayerDescription: playerDesc,
		Modifications:     mods,
	}
}

func (svc *Service) PreviewPunishment(ctx context.Context, serverID, rawUUID string, typeOrdinal int) PunishmentPreview {
	pl, err := svc.store.FindPlayerByUUID(ctx, serverID, rawUUID)
	if err != nil || pl == nil {
		return PunishmentPreview{Status: 400, Success: false, Message: "Player not found"}
	}
	types := svc.store.GetPunishmentTypes(ctx, serverID)
	pt, ok := indexTypesByOrdinal(types)[typeOrdinal]
	if !ok {
		return PunishmentPreview{Status: 400, Success: false, Message: "Punishment type not found"}
	}
	thresholds := svc.store.GetThresholdSettings(ctx, serverID)
	st := svc.CalculateStatus(ctx, serverID, pl.Punishments)
	isSocial := pt.IsSocial()
	relevantPts := st.GameplayPoints
	var offLevel string
	if isSocial {
		relevantPts = st.SocialPoints
		offLevel = thresholds.Social.OffenseLevelInternal(relevantPts)
	} else {
		offLevel = thresholds.Gameplay.OffenseLevelInternal(relevantPts)
	}
	displayStatus := OffenseDisplayStatus(offLevel)

	prev := PunishmentPreview{
		Status:                       200,
		Success:                      true,
		OffenderStatus:               displayStatus,
		SocialStatus:                 st.Social,
		GameplayStatus:               st.Gameplay,
		SocialPoints:                 st.SocialPoints,
		GameplayPoints:               st.GameplayPoints,
		SingleSeverityPunishment:     pt.SingleSeverityPunishment,
		PermanentUntilUsernameChange: pt.PermanentUntilUsernameChange,
		PermanentUntilSkinChange:     pt.PermanentUntilSkinChange,
		CanBeAltBlocking:             pt.CanBeAltBlocking,
		CanBeStatWiping:              pt.CanBeStatWiping,
		Category:                     pt.Category,
	}
	if pt.SingleSeverityPunishment || pt.PermanentUntilUsernameChange || pt.PermanentUntilSkinChange {
		prev.SingleSeverity = buildSeverityPreview(&pt, "regular", offLevel, st, isSocial, thresholds)
	} else {
		prev.Lenient = buildSeverityPreview(&pt, "low", offLevel, st, isSocial, thresholds)
		prev.Regular = buildSeverityPreview(&pt, "regular", offLevel, st, isSocial, thresholds)
		prev.Aggravated = buildSeverityPreview(&pt, "severe", offLevel, st, isSocial, thresholds)
	}
	return prev
}

func buildSeverityPreview(pt *PunishmentType, sev, offLevel string, st PlayerStatus, isSocial bool, thresholds OffenderThresholdSettings) *SeverityPreview {
	pts := pt.PointsForSeverity(sev)
	detail := pt.GetDurationDetail(sev, offLevel)
	var durMs int64
	perm := false
	pType := "unknown"
	if detail != nil {
		durMs = detail.ToMilliseconds()
		perm = detail.IsPermanent()
		if detail.IsBan() || pt.IsBan() {
			pType = "ban"
		} else if detail.IsMute() || pt.IsMute() {
			pType = "mute"
		} else if detail.IsKick() || pt.IsKick() {
			pType = "kick"
		}
	} else if pt.IsBan() {
		pType = "ban"
	} else if pt.IsMute() {
		pType = "mute"
	} else if pt.IsKick() {
		pType = "kick"
	}
	newSocial := st.SocialPoints
	newGameplay := st.GameplayPoints
	if isSocial {
		newSocial += pts
	} else {
		newGameplay += pts
	}
	return &SeverityPreview{
		Severity:          sev,
		Points:            pts,
		DurationMs:        durMs,
		DurationFormatted: FormatDuration(durMs, perm),
		PunishmentType:    pType,
		Permanent:         perm,
		NewSocialStatus:   thresholds.Social.OffenderLevel(newSocial),
		NewGameplayStatus: thresholds.Gameplay.OffenderLevel(newGameplay),
		NewSocialPoints:   newSocial,
		NewGameplayPoints: newGameplay,
	}
}

func (svc *Service) SearchPunishments(ctx context.Context, serverID, query string, activeOnly bool) []PunishmentSearchResult {
	players, err := svc.store.FindPlayersWithPunishments(ctx, serverID, 50)
	if err != nil {
		return []PunishmentSearchResult{}
	}
	issuerIDs := collectIssuerIDsForPlayers(players)
	resolved := svc.store.ResolveIssuerNames(ctx, serverID, issuerIDs)
	lowQ := strings.ToLower(query)

	out := make([]PunishmentSearchResult, 0)
	for _, pl := range players {
		uname := pl.LatestUsername()
		for _, p := range pl.Punishments {
			active := svc.IsPunishmentActive(p)
			if activeOnly && !active {
				continue
			}
			issuerName := resolveIssuer(p.IssuerID, p.IssuerName, resolved)
			reason := p.DataString(PunishmentKeyReason)
			matches := strings.Contains(p.ID, query) ||
				strings.Contains(strings.ToLower(issuerName), lowQ) ||
				strings.Contains(strings.ToLower(reason), lowQ)
			if !matches {
				continue
			}
			statusLabel := "Inactive"
			if active {
				statusLabel = "Active"
			}
			out = append(out, PunishmentSearchResult{
				ID:          p.ID,
				PlayerName:  uname,
				TypeOrdinal: p.TypeOrdinal,
				Status:      statusLabel,
				Issued:      p.IssuedMs,
			})
			if len(out) >= 20 {
				return out
			}
		}
	}
	return out
}

func (svc *Service) GetRecentPunishments(ctx context.Context, serverID string, hours int) []PunishmentView {
	cutoff := httpx.NowMs() - int64(hours)*3600*1000
	players, err := svc.store.FindPlayersWithPunishmentsIssuedAfter(ctx, serverID, cutoff, 500)
	if err != nil {
		return []PunishmentView{}
	}
	typesByOrd := indexTypesByOrdinal(svc.store.GetPunishmentTypes(ctx, serverID))
	type ctxPair struct {
		pl  *Player
		pun *Punishment
	}
	var pairs []ctxPair
	var issuerIDs []string
	for _, pl := range players {
		for _, p := range pl.Punishments {
			if p.IssuedMs > cutoff {
				pairs = append(pairs, ctxPair{pl: pl, pun: p})
				issuerIDs = append(issuerIDs, collectIssuerIDs(p)...)
			}
		}
	}
	resolved := svc.store.ResolveIssuerNames(ctx, serverID, issuerIDs)
	views := make([]PunishmentView, 0, len(pairs))
	for _, pair := range pairs {
		views = append(views, svc.ToPunishmentView(pair.pun, typesByOrd, resolved, pair.pl))
	}
	sort.SliceStable(views, func(i, j int) bool {
		var li, lj int64
		if views[i].Issued != nil {
			li = *views[i].Issued
		}
		if views[j].Issued != nil {
			lj = *views[j].Issued
		}
		return li > lj
	})
	if len(views) > 100 {
		views = views[:100]
	}
	return views
}

func (svc *Service) GetLinkedBansForParent(ctx context.Context, serverID, parentPunishmentID string) []LinkedBanView {
	players, err := svc.store.FindPlayersByLinkedBanID(ctx, serverID, parentPunishmentID)
	if err != nil {
		return []LinkedBanView{}
	}
	out := make([]LinkedBanView, 0)
	for _, pl := range players {
		uname := pl.LatestUsername()
		for _, p := range pl.Punishments {
			if p.DataString(PunishmentKeyLinkedBanID) == parentPunishmentID {
				out = append(out, LinkedBanView{
					PunishmentID: p.ID,
					PlayerUUID:   pl.MinecraftUUID,
					PlayerName:   uname,
					Active:       svc.IsPunishmentActive(p),
				})
			}
		}
	}
	return out
}

func (svc *Service) GetPublicAppealInfo(ctx context.Context, serverID, punishmentID string) (view *AppealInfoView, notStartedMsg string, found bool) {
	pl, pun, err := svc.store.FindPunishmentContext(ctx, serverID, punishmentID)
	if err != nil || pun == nil {
		return nil, "", false
	}
	if pun.StartedMs == nil {
		return nil, "This punishment has not been started yet and cannot be appealed at this time.", true
	}
	resolved := svc.store.ResolveIssuerNames(ctx, serverID, collectIssuerIDs(pun))
	resp := svc.BuildPunishmentResponseView(ctx, serverID, pun, pl, resolved)

	appeals, _ := svc.store.FindAppealsByPunishmentID(ctx, serverID, punishmentID)
	var existingAppeal map[string]any
	if len(appeals) > 0 {
		existingAppeal = appeals[0]
	}
	var appealForm map[string]any
	if pt, ok := indexTypesByOrdinal(svc.store.GetPunishmentTypes(ctx, serverID))[pun.TypeOrdinal]; ok {
		appealForm = pt.AppealForm
	}
	return &AppealInfoView{
		ID:             resp.ID,
		Type:           resp.Type,
		IssuedMs:       resp.IssuedMs,
		ExpiresMs:      resp.ExpiresMs,
		Active:         resp.Active,
		Appealable:     resp.IsAppealable,
		PlayerUUID:     pl.MinecraftUUID,
		ExistingAppeal: existingAppeal,
		AppealForm:     appealForm,
	}, "", true
}

// --- Player Queries & Login ---

func (svc *Service) SearchPlayers(ctx context.Context, serverID, query string) []PlayerSearchResult {
	players, err := svc.store.SearchPlayers(ctx, serverID, query, 25)
	if err != nil {
		return []PlayerSearchResult{}
	}
	out := make([]PlayerSearchResult, 0, len(players))
	for _, pl := range players {
		st := svc.CalculateStatus(ctx, serverID, pl.Punishments)
		overall := st.Gameplay
		if statusRank(st.Social) > statusRank(overall) {
			overall = st.Social
		}
		lastOnline := pl.LastLoginMs()
		if lastOnline == nil && pl.LastSeenMs > 0 {
			ls := pl.LastSeenMs
			lastOnline = &ls
		}
		out = append(out, PlayerSearchResult{
			UUID:       pl.MinecraftUUID,
			Username:   pl.LatestUsername(),
			Status:     overall,
			LastOnline: lastOnline,
			IsOnline:   pl.IsOnline(),
		})
	}
	return out
}

func statusRank(s string) int {
	switch strings.ToLower(s) {
	case "habitual":
		return 3
	case "medium":
		return 2
	default:
		return 1
	}
}

func (svc *Service) GetPlayerDetails(ctx context.Context, serverID, rawUUID string) (*PlayerDetailView, error) {
	pl, err := svc.store.FindPlayerByUUID(ctx, serverID, rawUUID)
	if err != nil {
		return nil, err
	}
	if pl == nil {
		return nil, ErrNotFound
	}
	st := svc.CalculateStatus(ctx, serverID, pl.Punishments)
	resolved := svc.store.ResolveIssuerNames(ctx, serverID, collectIssuerIDsForPlayers([]*Player{pl}))
	puns := make([]PunishmentResponseView, 0, len(pl.Punishments))
	for _, p := range SortPunishmentsNewestFirst(pl.Punishments) {
		puns = append(puns, svc.BuildPunishmentResponseView(ctx, serverID, p, pl, resolved))
	}
	notes := make([]NoteEntry, len(pl.Notes))
	for i, n := range pl.Notes {
		notes[i] = n
		notes[i].IssuerName = resolveIssuer(n.IssuerID, n.IssuerName, resolved)
	}
	var latestIP *IPEntry
	if len(pl.IPAddresses) > 0 {
		cp := pl.IPAddresses[len(pl.IPAddresses)-1]
		latestIP = &cp
	}
	var lastServerPtr *string
	if ls := pl.LastServer(); ls != "" {
		lastServerPtr = &ls
	}
	return &PlayerDetailView{
		ID:             pl.ID,
		MinecraftUUID:  pl.MinecraftUUID,
		Usernames:      pl.Usernames,
		Notes:          notes,
		IPAddresses:    pl.IPAddresses,
		Punishments:    puns,
		Data:           pl.Data,
		Social:         st.Social,
		Gameplay:       st.Gameplay,
		SocialPoints:   st.SocialPoints,
		GameplayPoints: st.GameplayPoints,
		LatestIPData:   latestIP,
		LastServer:     lastServerPtr,
		PlaytimeHours:  float64(pl.TotalPlaytimeSeconds()) / 3600.0,
	}, nil
}

func (svc *Service) GetLinkedAccountsPanel(ctx context.Context, serverID, rawUUID string) []LinkedAccountView {
	pl, err := svc.store.FindPlayerByUUID(ctx, serverID, rawUUID)
	if err != nil || pl == nil {
		return []LinkedAccountView{}
	}
	uuids := pl.LinkedAccountUUIDs()
	if len(uuids) == 0 {
		return []LinkedAccountView{}
	}
	linkedPlayers, err := svc.store.FindPlayersByUUIDs(ctx, serverID, uuids)
	if err != nil {
		return []LinkedAccountView{}
	}
	typesByOrd := indexTypesByOrdinal(svc.store.GetPunishmentTypes(ctx, serverID))
	out := make([]LinkedAccountView, 0, len(linkedPlayers))
	for _, lp := range linkedPlayers {
		bans, mutes := 0, 0
		for _, p := range lp.Punishments {
			if svc.IsPunishmentActive(p) {
				switch svc.GetEffectiveCategory(p, typesByOrd) {
				case CategoryBan:
					bans++
				case CategoryMute:
					mutes++
				}
			}
		}
		out = append(out, LinkedAccountView{
			MinecraftUUID:    lp.MinecraftUUID,
			Username:         lp.LatestUsername(),
			ActiveBans:       bans,
			ActiveMutes:      mutes,
			LastLinkedUpdate: lp.LastLinkedUpdateMs(),
		})
	}
	return out
}

func (svc *Service) FindAndLinkAccounts(ctx context.Context, serverID, rawUUID string) (bool, string, int) {
	pl, err := svc.store.FindPlayerByUUID(ctx, serverID, rawUUID)
	if err != nil || pl == nil {
		return false, "Player not found", 0
	}
	const sixHoursMs = 6 * 60 * 60 * 1000
	now := httpx.NowMs()

	ipSet := make(map[string]bool)
	for _, ip := range pl.IPAddresses {
		if !ip.Proxy && ip.IPAddress != "" {
			ipSet[ip.IPAddress] = true
		}
	}
	if len(ipSet) == 0 {
		for _, ip := range pl.IPAddresses {
			if len(ip.LoginsMs) > 0 && now-ip.LoginsMs[len(ip.LoginsMs)-1] < sixHoursMs {
				ipSet[ip.IPAddress] = true
			}
		}
	}
	if len(ipSet) == 0 {
		return true, "No valid IPs to match", 0
	}
	ips := make([]string, 0, len(ipSet))
	for k := range ipSet {
		ips = append(ips, k)
	}
	matches, err := svc.store.FindPlayersByIPAddresses(ctx, serverID, ips, pl.MinecraftUUID, 200)
	if err != nil {
		return false, "Failed to search linked accounts", 0
	}
	var linked []*Player
	for _, m := range matches {
		if strings.EqualFold(m.MinecraftUUID, pl.MinecraftUUID) {
			continue
		}
		shouldLink := false
		for _, mip := range m.IPAddresses {
			if !ipSet[mip.IPAddress] {
				continue
			}
			if !mip.Proxy || hasRecentMatchingLogin(pl, m, mip.IPAddress, sixHoursMs) {
				shouldLink = true
				break
			}
		}
		if shouldLink {
			linked = append(linked, m)
		}
	}
	if len(linked) > 0 {
		linkedUUIDs := make([]string, 0, len(linked))
		for _, lp := range linked {
			linkedUUIDs = append(linkedUUIDs, lp.MinecraftUUID)
		}
		_ = svc.store.AddLinkedAccounts(ctx, serverID, pl.MinecraftUUID, linkedUUIDs, now)
		for _, lp := range linked {
			_ = svc.store.AddLinkedAccounts(ctx, serverID, lp.MinecraftUUID, []string{pl.MinecraftUUID}, now)
		}
	}
	return true, "Linking complete", len(linked)
}

func hasRecentMatchingLogin(p1, p2 *Player, ipAddr string, windowMs int64) bool {
	var l1, l2 []int64
	for _, ip := range p1.IPAddresses {
		if ip.IPAddress == ipAddr {
			l1 = ip.LoginsMs
			break
		}
	}
	for _, ip := range p2.IPAddresses {
		if ip.IPAddress == ipAddr {
			l2 = ip.LoginsMs
			break
		}
	}
	for _, a := range l1 {
		for _, b := range l2 {
			diff := a - b
			if diff < 0 {
				diff = -diff
			}
			if diff < windowMs {
				return true
			}
		}
	}
	return false
}

type LoginResult struct {
	HTTPStatus           int
	ActivePunishments    []SimplePunishmentView
	PendingNotifications []map[string]any
	PendingIPLookups     []string
	PendingStatWipes     []map[string]any
}

func (svc *Service) Login(ctx context.Context, srv *models.Server, rawUUID, username, ipAddr string, ipInfo map[string]any, skinHash, serverName string) (*LoginResult, error) {
	pl, isNewIP, err := svc.store.UpsertLoginPlayer(ctx, srv.ID, rawUUID, username, ipAddr, ipInfo, skinHash, serverName)
	if err != nil {
		return nil, err
	}
	svc.FindAndLinkAccounts(ctx, srv.ID, pl.MinecraftUUID)
	if reloaded, _ := svc.store.FindPlayerByUUID(ctx, srv.ID, pl.MinecraftUUID); reloaded != nil {
		pl = reloaded
	}
	promoted := svc.PromoteUnstartedPunishments(ctx, srv, pl)
	autoPardoned := svc.CheckRestrictionAutoPardons(ctx, srv, pl, username, skinHash)
	linkedBans := svc.EnforceAltBlockingBans(ctx, srv, pl)
	if len(promoted) > 0 || len(autoPardoned) > 0 || len(linkedBans) > 0 {
		if reloaded, _ := svc.store.FindPlayerByUUID(ctx, srv.ID, pl.MinecraftUUID); reloaded != nil {
			pl = reloaded
		}
	}

	typesByOrd := indexTypesByOrdinal(svc.store.GetPunishmentTypes(ctx, srv.ID))
	resolved := svc.store.ResolveIssuerNames(ctx, srv.ID, collectIssuerIDsForPlayers([]*Player{pl}))

	activePuns := make([]SimplePunishmentView, 0)
	for _, p := range pl.Punishments {
		if !svc.IsPunishmentActive(p) {
			continue
		}
		if pt, ok := typesByOrd[p.TypeOrdinal]; ok && pt.IsKick() {
			continue
		}
		activePuns = append(activePuns, svc.ToSimplePunishment(p, typesByOrd, resolved))
	}
	for _, p := range pl.Punishments {
		if p.TypeOrdinal == 0 && p.StartedMs == nil && !p.IsPardoned() {
			activePuns = append(activePuns, svc.ToSimplePunishment(p, typesByOrd, resolved))
		}
	}
	activePuns = deduplicateSimplePunishments(activePuns)

	pendingNotifs, _ := svc.store.GetPendingNotificationsForPlayer(ctx, srv.ID, pl.MinecraftUUID)
	pendingNotifs = append(pendingNotifs, pl.LegacyPendingNotifications()...)
	if pendingNotifs == nil {
		pendingNotifs = []map[string]any{}
	}

	pendingLookups := []string{}
	if ipAddr != "" && isNewIP {
		for _, ip := range pl.IPAddresses {
			if ip.IPAddress == ipAddr && ip.Country == nil {
				pendingLookups = append(pendingLookups, ipAddr)
				break
			}
		}
	}

	pendingStatWipes := []map[string]any{}
	for _, p := range pl.Punishments {
		if p.DataBool(PunishmentKeyWipeAfterExpiry) && !p.DataBool(PunishmentKeyStatWipeCompleted) && svc.IsPunishmentNaturallyExpired(p) {
			pendingStatWipes = append(pendingStatWipes, map[string]any{
				"minecraftUuid": pl.MinecraftUUID,
				"username":      pl.LatestUsername(),
				"punishmentId":  p.ID,
			})
		}
	}

	statusCode := 200
	if len(pl.Usernames) <= 1 {
		statusCode = 201
	}
	return &LoginResult{
		HTTPStatus:           statusCode,
		ActivePunishments:    activePuns,
		PendingNotifications: pendingNotifs,
		PendingIPLookups:     pendingLookups,
		PendingStatWipes:     pendingStatWipes,
	}, nil
}

func deduplicateSimplePunishments(list []SimplePunishmentView) []SimplePunishmentView {
	oldestByCat := make(map[string]SimplePunishmentView)
	var catOrder []string
	out := make([]SimplePunishmentView, 0, len(list))
	for _, sp := range list {
		if sp.Category == CategoryBan || sp.Category == CategoryMute {
			if ex, ok := oldestByCat[sp.Category]; !ok {
				oldestByCat[sp.Category] = sp
				catOrder = append(catOrder, sp.Category)
			} else if sp.IssuedAt < ex.IssuedAt {
				oldestByCat[sp.Category] = sp
			}
		} else {
			out = append(out, sp)
		}
	}
	for _, c := range catOrder {
		out = append(out, oldestByCat[c])
	}
	return out
}

func (svc *Service) BuildPlayerProfileMap(ctx context.Context, srv *models.Server, pl *Player, punishmentLimit, noteLimit *int) map[string]any {
	types := svc.store.GetPunishmentTypes(ctx, srv.ID)
	typesByOrd := indexTypesByOrdinal(types)
	resolved := svc.store.ResolveIssuerNames(ctx, srv.ID, collectIssuerIDsForPlayers([]*Player{pl}))

	usernames := make([]map[string]any, 0, len(pl.Usernames))
	for _, u := range pl.Usernames {
		usernames = append(usernames, map[string]any{
			"username": u.Username,
			"date":     u.DateMs,
		})
	}

	allNotes := make([]NoteEntry, len(pl.Notes))
	copy(allNotes, pl.Notes)
	if noteLimit != nil {
		sort.SliceStable(allNotes, func(i, j int) bool { return allNotes[i].DateMs > allNotes[j].DateMs })
		if *noteLimit < 0 {
			allNotes = nil
		} else if len(allNotes) > *noteLimit {
			allNotes = allNotes[:*noteLimit]
		}
	}
	notes := make([]map[string]any, 0, len(allNotes))
	for _, n := range allNotes {
		notes = append(notes, map[string]any{
			"id":         n.ID,
			"text":       n.Text,
			"date":       n.DateMs,
			"issuerName": resolveIssuer(n.IssuerID, n.IssuerName, resolved),
			"issuerId":   n.IssuerID,
		})
	}

	ips := make([]map[string]any, 0, len(pl.IPAddresses))
	for _, ip := range pl.IPAddresses {
		ips = append(ips, map[string]any{
			"ipAddress":  ip.IPAddress,
			"country":    ip.Country,
			"region":     ip.Region,
			"asn":        ip.ASN,
			"proxy":      ip.Proxy,
			"hosting":    ip.Hosting,
			"firstLogin": ip.FirstLoginMs,
			"logins":     ip.LoginsMs,
		})
	}

	allPuns := pl.Punishments
	if punishmentLimit != nil {
		allPuns = SortPunishmentsNewestFirst(allPuns)
		if *punishmentLimit < 0 {
			allPuns = nil
		} else if len(allPuns) > *punishmentLimit {
			allPuns = allPuns[:*punishmentLimit]
		}
	}
	puns := make([]PunishmentView, 0, len(allPuns))
	for _, p := range allPuns {
		puns = append(puns, svc.ToPunishmentView(p, typesByOrd, resolved, nil))
	}

	pendingNotifs, _ := svc.store.GetPendingNotificationsForPlayer(ctx, srv.ID, pl.MinecraftUUID)
	pendingNotifs = append(pendingNotifs, pl.LegacyPendingNotifications()...)
	if pendingNotifs == nil {
		pendingNotifs = []map[string]any{}
	}

	profile := map[string]any{
		"id":                   pl.ID,
		"minecraftUuid":        pl.MinecraftUUID,
		"usernames":            usernames,
		"notes":                notes,
		"ipAddresses":          ips,
		"punishments":          puns,
		"pendingNotifications": pendingNotifs,
		"data":                 pl.Data,
	}
	if punishmentLimit != nil {
		profile["punishmentCount"] = len(pl.Punishments)
	}
	if noteLimit != nil {
		profile["noteCount"] = len(pl.Notes)
	}
	return profile
}

func (svc *Service) BuildLookupResponseMap(ctx context.Context, srv *models.Server, pl *Player) map[string]any {
	typesByOrd := indexTypesByOrdinal(svc.store.GetPunishmentTypes(ctx, srv.ID))
	prevUsernames := make([]string, 0)
	if len(pl.Usernames) > 1 {
		for _, u := range pl.Usernames[1:] {
			prevUsernames = append(prevUsernames, u.Username)
		}
	}
	var firstSeen, lastSeen any
	if len(pl.Usernames) > 0 {
		firstSeen = pl.Usernames[0].DateMs
		lastSeen = pl.Usernames[len(pl.Usernames)-1].DateMs
	} else if pl.FirstSeenMs > 0 {
		firstSeen = pl.FirstSeenMs
		lastSeen = pl.LastSeenMs
	}
	totalPuns := len(pl.Punishments)
	activePuns := 0
	bans, mutes, kicks, warns := 0, 0, 0, 0
	for _, p := range pl.Punishments {
		if svc.IsPunishmentActive(p) {
			activePuns++
		}
		if pt, ok := typesByOrd[p.TypeOrdinal]; ok && pt.IsKick() {
			kicks++
			continue
		}
		cat := svc.GetEffectiveCategory(p, typesByOrd)
		switch cat {
		case CategoryBan:
			bans++
		case CategoryMute:
			mutes++
		default:
			warns++
		}
	}
	sorted := SortPunishmentsNewestFirst(pl.Punishments)
	if len(sorted) > 5 {
		sorted = sorted[:5]
	}
	resolved := svc.store.ResolveIssuerNames(ctx, srv.ID, collectIssuerIDsForPlayers([]*Player{pl}))
	recentPuns := make([]map[string]any, 0, len(sorted))
	for _, p := range sorted {
		typeName := "Unknown"
		if pt, ok := typesByOrd[p.TypeOrdinal]; ok {
			typeName = pt.Name
		}
		recentPuns = append(recentPuns, map[string]any{
			"id":        p.ID,
			"type":      typeName,
			"issuer":    resolveIssuer(p.IssuerID, p.IssuerName, resolved),
			"issuedAt":  p.IssuedMs,
			"expiresAt": svc.GetEffectiveExpiry(p),
			"isActive":  svc.IsPunishmentActive(p),
		})
	}
	baseURL := "https://" + srv.CustomDomain
	return map[string]any{
		"minecraftUuid":     pl.MinecraftUUID,
		"currentUsername":   pl.LatestUsername(),
		"previousUsernames": prevUsernames,
		"firstSeen":         firstSeen,
		"lastSeen":          lastSeen,
		"isOnline":          pl.IsOnline(),
		"punishmentStats": map[string]any{
			"totalPunishments":  totalPuns,
			"activePunishments": activePuns,
			"bans":              bans,
			"mutes":             mutes,
			"kicks":             kicks,
			"warnings":          warns,
		},
		"recentPunishments": recentPuns,
		"profileUrl":        baseURL + "/player/" + pl.MinecraftUUID,
		"punishmentsUrl":    baseURL + "/player/" + pl.MinecraftUUID + "/punishments",
	}
}

// --- Sync & Startup ---

func (svc *Service) HandleStartup(ctx context.Context, srv *models.Server, serverVersion, platformType, pluginVersion string, maxPlayers int, serverName, clientIP string) map[string]any {
	now := time.Now().UTC()
	nowMs := now.UnixMilli()
	svc.store.UpdateServerActivity(ctx, srv.ID, nowMs, nil)
	svc.store.UpsertServerInstanceSnapshot(ctx, srv.ID, serverName, 0, platformType, serverVersion, clientIP, pluginVersion, nowMs)

	rtURL := ""
	if svc.cfg.RealtimeEnabled && strings.TrimSpace(svc.cfg.ModlDomain) != "" {
		scheme := "wss"
		if strings.HasPrefix(svc.cfg.ModlDomain, "localhost") || strings.HasPrefix(svc.cfg.ModlDomain, "127.0.0.1") {
			scheme = "ws"
		}
		rtURL = scheme + "://" + strings.TrimSpace(svc.cfg.ModlDomain) + "/v1/realtime/ws"
	}
	rtEnabled := svc.cfg.RealtimeEnabled && rtURL != ""
	var rtURLVal any
	var topics []string
	if rtEnabled {
		rtURLVal = rtURL
		topics = []string{
			"TOPIC_MINECRAFT_PERMISSIONS",
			"TOPIC_MINECRAFT_PUNISHMENT_TYPES",
			"TOPIC_MINECRAFT_PUNISHMENTS",
			"TOPIC_MINECRAFT_PLAYER_NOTIFICATIONS",
			"TOPIC_MINECRAFT_STAFF_2FA",
			"TOPIC_MINECRAFT_MIGRATION_TASKS",
		}
	} else {
		topics = []string{}
	}

	return map[string]any{
		"panelUrl":                "https://" + srv.CustomDomain,
		"timestamp":               now.Format(time.RFC3339Nano),
		"serverInstanceId":        NewShortID() + "-" + NewShortID(),
		"realtimeEnabled":         rtEnabled,
		"realtimeUrl":             rtURLVal,
		"realtimeProtocolVersion": realtime.ProtocolVersion,
		"realtimeTopics":          topics,
	}
}

func (svc *Service) ApplyPresence(ctx context.Context, srv *models.Server, onlinePlayers []SyncOnlinePlayerInput, serverName string) {
	nowMs := httpx.NowMs()
	cnt := len(onlinePlayers)
	svc.store.UpdateServerActivity(ctx, srv.ID, nowMs, &cnt)
	onlineMap := make(map[string]bool, len(onlinePlayers))
	for _, op := range onlinePlayers {
		if norm, ok := NormalizeUUID(op.UUID); ok {
			onlineMap[norm] = true
		}
	}
	svc.store.MarkStalePlayersOffline(ctx, srv.ID, onlineMap, serverName, nowMs)
}

func (svc *Service) SubmitLogs(ctx context.Context, srv *models.Server, chatLogs []SyncChatLogInput, commandLogs []SyncCommandLogInput) {
	if len(chatLogs) > 0 {
		_ = svc.store.InsertChatLogs(ctx, srv.ID, chatLogs)
	}
	if len(commandLogs) > 0 {
		_ = svc.store.InsertCommandLogs(ctx, srv.ID, commandLogs)
	}
}

func (svc *Service) ApplyServerStatus(ctx context.Context, srv *models.Server, status *SyncServerStatusInput, serverName, clientIP string) {
	if status == nil {
		return
	}
	svc.store.UpsertServerInstanceSnapshot(ctx, srv.ID, serverName, status.OnlinePlayerCount, status.PlatformType, status.ServerVersion, clientIP, status.PluginVersion, httpx.NowMs())
}

func (svc *Service) Sync(
	ctx context.Context,
	srv *models.Server,
	lastSyncTimestamp string,
	onlinePlayers []SyncOnlinePlayerInput,
	serverName string,
	chatLogs []SyncChatLogInput,
	commandLogs []SyncCommandLogInput,
	serverStatus *SyncServerStatusInput,
	clientIP string,
	isV2 bool,
) SyncResult {
	now := time.Now().UTC()
	nowMs := now.UnixMilli()

	svc.ApplyPresence(ctx, srv, onlinePlayers, serverName)

	lastSyncMs := nowMs - 30000
	if strings.TrimSpace(lastSyncTimestamp) != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(lastSyncTimestamp)); err == nil {
			lastSyncMs = parsed.UnixMilli()
		}
	}

	types := svc.store.GetPunishmentTypes(ctx, srv.ID)
	typesByOrd := indexTypesByOrdinal(types)
	pendingPuns := make([]SyncPunishmentEntry, 0)
	recentModPuns := make([]SyncPunishmentEntry, 0)
	playerNotifs := make([]map[string]any, 0)

	onlineUUIDs := make([]string, 0, len(onlinePlayers))
	onlineIPs := make(map[string]string, len(onlinePlayers))
	for _, op := range onlinePlayers {
		if norm, ok := NormalizeUUID(op.UUID); ok {
			onlineUUIDs = append(onlineUUIDs, norm)
			if op.IPAddress != "" {
				onlineIPs[norm] = op.IPAddress
			}
		}
	}

	if len(onlineUUIDs) > 0 {
		players, _ := svc.store.FindPlayersByUUIDs(ctx, srv.ID, onlineUUIDs)
		resolved := svc.store.ResolveIssuerNames(ctx, srv.ID, collectIssuerIDsForPlayers(players))
		for _, pl := range players {
			promoted := svc.PromoteUnstartedPunishments(ctx, srv, pl)
			if len(promoted) > 0 {
				if reloaded, _ := svc.store.FindPlayerByUUID(ctx, srv.ID, pl.MinecraftUUID); reloaded != nil {
					pl = reloaded
				}
			}
			svc.collectPlayerSyncPunishments(pl, typesByOrd, lastSyncMs, resolved, &pendingPuns, &recentModPuns)
			notifs, _ := svc.store.GetPendingNotificationsForPlayer(ctx, srv.ID, pl.MinecraftUUID)
			playerNotifs = append(playerNotifs, notifs...)
			for _, ln := range pl.LegacyPendingNotifications() {
				cp := make(map[string]any, len(ln)+1)
				for k, v := range ln {
					cp[k] = v
				}
				cp["targetPlayerUuid"] = pl.MinecraftUUID
				playerNotifs = append(playerNotifs, cp)
			}
		}
	}

	pendingPuns = deduplicateSyncPunishments(pendingPuns)
	staffNotifs := svc.collectStaffSyncEvents(ctx, srv, lastSyncMs, typesByOrd, recentModPuns)
	activeStaff, _ := svc.store.GetActiveStaffMembers(ctx, srv.ID, srv.AdminEmail, onlineIPs)
	if activeStaff == nil {
		activeStaff = []map[string]any{}
	}
	svc.SubmitLogs(ctx, srv, chatLogs, commandLogs)
	spUpdated, ptUpdated := svc.store.GetServerUpdatedTimestamps(ctx, srv.ID)
	twoFA := svc.store.CollectAndClearPending2FA(ctx, srv.ID)
	migTask := svc.store.GetActiveMigrationTask(ctx, srv.ID)
	svc.ApplyServerStatus(ctx, srv, serverStatus, serverName, clientIP)

	if isV2 {
		convertTicketURLsToRelative(staffNotifs)
		convertTicketURLsToRelative(playerNotifs)
	}

	return SyncResult{
		Timestamp: now.Format(time.RFC3339Nano),
		Data: SyncDataView{
			PendingPunishments:          pendingPuns,
			RecentlyStartedPunishments:  []SyncPunishmentEntry{},
			RecentlyModifiedPunishments: recentModPuns,
			PlayerNotifications:         playerNotifs,
			StaffNotifications:          staffNotifs,
			ActiveStaffMembers:          activeStaff,
			PendingStatWipes:            []map[string]any{},
			StaffPermissionsUpdatedAt:   spUpdated,
			PunishmentTypesUpdatedAt:    ptUpdated,
			Staff2faVerifications:       twoFA,
			MigrationTask:               migTask,
		},
	}
}

func (svc *Service) collectPlayerSyncPunishments(
	pl *Player,
	typesByOrd map[int]PunishmentType,
	lastSyncMs int64,
	resolved map[string]string,
	pending *[]SyncPunishmentEntry,
	recentlyMod *[]SyncPunishmentEntry,
) {
	uuidStr := pl.MinecraftUUID
	uname := pl.LatestUsername()
	catsActiveStarted := make(map[string]bool)
	oldestUnstarted := make(map[string]*Punishment)
	pardonedCats := make(map[string]bool)

	for _, p := range pl.Punishments {
		active := svc.IsPunishmentActive(p)
		cat := svc.GetEffectiveCategory(p, typesByOrd)
		isRecentMod := false
		isRecentPardon := false
		for _, m := range p.Modifications {
			if m.DateMs > lastSyncMs {
				isRecentMod = true
				if IsPardonModType(m.Type) {
					isRecentPardon = true
				}
			}
		}
		if isRecentMod {
			*recentlyMod = append(*recentlyMod, SyncPunishmentEntry{
				MinecraftUUID: uuidStr,
				Username:      uname,
				Punishment:    svc.ToSimplePunishment(p, typesByOrd, resolved),
			})
		}
		if isRecentPardon && cat != "" {
			pardonedCats[cat] = true
		}
		if p.TypeOrdinal == 0 && p.StartedMs == nil && !p.IsPardoned() {
			*pending = append(*pending, SyncPunishmentEntry{
				MinecraftUUID: uuidStr,
				Username:      uname,
				Punishment:    svc.ToSimplePunishment(p, typesByOrd, resolved),
			})
			continue
		}
		if !active {
			continue
		}
		if p.StartedMs != nil && p.IssuedMs > lastSyncMs && cat != "" {
			*pending = append(*pending, SyncPunishmentEntry{
				MinecraftUUID: uuidStr,
				Username:      uname,
				Punishment:    svc.ToSimplePunishment(p, typesByOrd, resolved),
			})
		}
		if cat != "" && p.StartedMs != nil {
			catsActiveStarted[cat] = true
		} else if cat != "" && p.StartedMs == nil {
			if ex, ok := oldestUnstarted[cat]; !ok || p.IssuedMs < ex.IssuedMs {
				oldestUnstarted[cat] = p
			}
		}
	}
	for cat, p := range oldestUnstarted {
		if !catsActiveStarted[cat] {
			*pending = append(*pending, SyncPunishmentEntry{
				MinecraftUUID: uuidStr,
				Username:      uname,
				Punishment:    svc.ToSimplePunishment(p, typesByOrd, resolved),
			})
		}
	}
	if len(pardonedCats) > 0 {
		for _, p := range pl.Punishments {
			if !svc.IsPunishmentActive(p) || p.StartedMs == nil {
				continue
			}
			cat := svc.GetEffectiveCategory(p, typesByOrd)
			if cat != "" && pardonedCats[cat] {
				*pending = append(*pending, SyncPunishmentEntry{
					MinecraftUUID: uuidStr,
					Username:      uname,
					Punishment:    svc.ToSimplePunishment(p, typesByOrd, resolved),
				})
				delete(pardonedCats, cat)
			}
		}
	}
}

func deduplicateSyncPunishments(entries []SyncPunishmentEntry) []SyncPunishmentEntry {
	oldestByKey := make(map[string]SyncPunishmentEntry)
	var keys []string
	out := make([]SyncPunishmentEntry, 0, len(entries))
	for _, e := range entries {
		cat := e.Punishment.Category
		if cat == CategoryBan || cat == CategoryMute {
			k := e.MinecraftUUID + "|" + cat
			if ex, ok := oldestByKey[k]; !ok {
				oldestByKey[k] = e
				keys = append(keys, k)
			} else if e.Punishment.IssuedAt < ex.Punishment.IssuedAt {
				oldestByKey[k] = e
			}
		} else {
			out = append(out, e)
		}
	}
	for _, k := range keys {
		out = append(out, oldestByKey[k])
	}
	return out
}

func (svc *Service) collectStaffSyncEvents(ctx context.Context, srv *models.Server, lastSyncMs int64, typesByOrd map[int]PunishmentType, recentMods []SyncPunishmentEntry) []map[string]any {
	out := make([]map[string]any, 0)

	// 1. Recent tickets
	if tickets, err := svc.store.FindRecentTicketsAfter(ctx, srv.ID, lastSyncMs, 20); err == nil {
		for _, t := range tickets {
			typ, _ := t["type"].(string)
			if strings.EqualFold(typ, "application") {
				continue
			}
			id, _ := t["id"].(string)
			creatorName, _ := t["creatorName"].(string)
			if creatorName == "" {
				creatorName = "Unknown"
			}
			repPlayer, _ := t["reportedPlayer"].(string)
			createdServer, _ := t["createdServer"].(string)
			subject, _ := t["subject"].(string)
			firstReply, _ := t["firstReplyContent"].(string)
			firstReply = strings.ReplaceAll(strings.ReplaceAll(firstReply, "**", ""), "```", "")
			createdAt, _ := t["createdAt"].(int64)

			isReport := strings.EqualFold(typ, "player") || strings.EqualFold(typ, "chat")
			var msg string
			if isReport {
				if repPlayer == "" {
					repPlayer = "Unknown"
				}
				catLabel := "Gameplay"
				if strings.EqualFold(typ, "chat") {
					catLabel = "Chat"
				}
				msg = creatorName + ": reported " + repPlayer
				if createdServer != "" {
					msg += " on " + createdServer
				}
				msg += " (" + catLabel + ")"
			} else {
				msg = creatorName + ": created " + id
				if createdServer != "" {
					msg += " on " + createdServer
				}
			}
			tData := map[string]any{
				"ticketId":          id,
				"creatorName":       creatorName,
				"subject":           subject,
				"firstReplyContent": firstReply,
				"ticketUrl":         "https://" + srv.CustomDomain + "/ticket/" + id,
				"ticketType":        typ,
			}
			if isReport {
				tData["reportedPlayer"] = repPlayer
				tData["category"] = typ
			}
			out = append(out, map[string]any{
				"id":        "ticket_" + id,
				"type":      "TICKET_CREATED",
				"message":   msg,
				"timestamp": createdAt,
				"data":      tData,
			})
		}
	}

	// 2. Recent punishments
	if players, err := svc.store.FindPlayersWithPunishmentsIssuedAfter(ctx, srv.ID, lastSyncMs, 50); err == nil {
		resolved := svc.store.ResolveIssuerNames(ctx, srv.ID, collectIssuerIDsForPlayers(players))
		for _, pl := range players {
			uname := pl.LatestUsername()
			for _, p := range pl.Punishments {
				if p.IssuedMs <= lastSyncMs {
					continue
				}
				typeName := "Unknown"
				isKick := false
				if pt, ok := typesByOrd[p.TypeOrdinal]; ok {
					typeName = pt.Name
					isKick = pt.IsKick()
				}
				cat := svc.GetEffectiveCategory(p, typesByOrd)
				action := "punished"
				if isKick {
					action = "kicked"
				} else if cat == CategoryBan {
					action = "banned"
				} else if cat == CategoryMute {
					action = "muted"
				}
				issuerName := resolveIssuer(p.IssuerID, p.IssuerName, resolved)
				out = append(out, map[string]any{
					"id":        "punishment_" + p.ID,
					"type":      "PUNISHMENT_ISSUED",
					"message":   fmt.Sprintf("%s: %s %s (%s)", issuerName, action, uname, typeName),
					"timestamp": p.IssuedMs,
				})
			}
		}
	}

	// 3. Recent pardons
	for _, entry := range recentMods {
		uname := entry.Username
		if uname == "" {
			uname = "Unknown"
		}
		pType := entry.Punishment.Type
		if pType == "" {
			pType = "punishment"
		}
		for _, m := range entry.Punishment.Modifications {
			if !IsPardonModType(m.Type) {
				continue
			}
			pardoner := m.IssuerName
			if pardoner == "" {
				pardoner = SystemIssuer
			}
			ts := httpx.NowMs()
			if m.Timestamp != nil {
				ts = *m.Timestamp
			}
			out = append(out, map[string]any{
				"id":        "pardon_" + entry.Punishment.ID,
				"type":      "PUNISHMENT_PARDONED",
				"message":   fmt.Sprintf("%s: pardoned %s's %s", pardoner, uname, pType),
				"timestamp": ts,
			})
		}
	}
	return out
}

func convertTicketURLsToRelative(notifs []map[string]any) {
	for _, n := range notifs {
		d, ok := n["data"].(map[string]any)
		if !ok {
			continue
		}
		if rawURL, ok := d["ticketUrl"].(string); ok {
			if idx := strings.Index(rawURL, "/ticket/"); idx >= 0 {
				d["ticketUrl"] = rawURL[idx:]
			}
		}
	}
}

// --- External Services (Crafatar & Mojang) ---

func (svc *Service) ProxyAvatar(ctx context.Context, dashedUUID string, size int, overlay bool) []byte {
	if size < 8 {
		size = 8
	} else if size > 512 {
		size = 512
	}
	key := fmt.Sprintf("%d:%t:%s", size, overlay, dashedUUID)
	now := time.Now()
	svc.avatarMu.RLock()
	if cached, ok := svc.avatarCache[key]; ok && now.Before(cached.expiresAt) {
		svc.avatarMu.RUnlock()
		return cached.data
	}
	svc.avatarMu.RUnlock()

	urls := []string{
		fmt.Sprintf("https://crafatar.com/avatars/%s?size=%d&overlay=%t", dashedUUID, size, overlay),
		fmt.Sprintf("https://minotar.net/avatar/%s/%d", dashedUUID, size),
	}
	var body []byte
	for _, u := range urls {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			continue
		}
		resp, err := svc.httpClient.Do(req)
		if err != nil {
			continue
		}
		ct := strings.ToLower(resp.Header.Get("Content-Type"))
		if resp.StatusCode == http.StatusOK && (ct == "" || strings.HasPrefix(ct, "image/")) {
			data, readErr := io.ReadAll(io.LimitReader(resp.Body, 1024*1024+1))
			_ = resp.Body.Close()
			if readErr == nil && len(data) > 0 && len(data) <= 1024*1024 {
				body = data
				break
			}
		} else {
			_ = resp.Body.Close()
		}
	}
	if len(body) > 0 {
		svc.avatarMu.Lock()
		svc.avatarCache[key] = cachedAvatar{data: body, expiresAt: now.Add(10 * time.Minute)}
		svc.avatarMu.Unlock()
	}
	return body
}

type MojangProfile struct {
	UUID string
	Name string
}

func (svc *Service) LookupMojangByUUID(ctx context.Context, rawUUID string) *MojangProfile {
	undashed := UndashedUUID(rawUUID)
	if undashed == "" {
		return nil
	}
	u := "https://sessionserver.mojang.com/session/minecraft/profile/" + undashed
	return svc.fetchMojangProfile(ctx, u)
}

func (svc *Service) LookupMojangByUsername(ctx context.Context, username string) *MojangProfile {
	if !mcUsernameRe.MatchString(strings.TrimSpace(username)) {
		return nil
	}
	u := "https://api.mojang.com/users/profiles/minecraft/" + strings.TrimSpace(username)
	return svc.fetchMojangProfile(ctx, u)
}

func (svc *Service) fetchMojangProfile(ctx context.Context, targetURL string) *MojangProfile {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil
	}
	resp, err := svc.httpClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	var payload struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64*1024)).Decode(&payload); err != nil {
		return nil
	}
	norm, ok := NormalizeUUID(payload.ID)
	if !ok || payload.Name == "" {
		return nil
	}
	return &MojangProfile{UUID: norm, Name: payload.Name}
}

// --- Realtime & Webhook Helpers ---

func (svc *Service) publishPunishmentIssued(ctx context.Context, srv *models.Server, pl *Player, pun *Punishment) {
	if svc.hub == nil {
		return
	}
	svc.hub.InvalidatePanel(srv.ID, modlv1.PanelResource_PANEL_RESOURCE_PLAYERS, pl.MinecraftUUID)
	svc.hub.InvalidatePanel(srv.ID, modlv1.PanelResource_PANEL_RESOURCE_PUNISHMENTS, pun.ID)
	typesByOrd := indexTypesByOrdinal(svc.store.GetPunishmentTypes(ctx, srv.ID))
	resolved := svc.store.ResolveIssuerNames(ctx, srv.ID, collectIssuerIDs(pun))
	sp := svc.ToSimplePunishment(pun, typesByOrd, resolved)
	svc.hub.PushPunishments(srv.ID, []*modlv1.SyncPendingPunishment{{
		MinecraftUuid: pl.MinecraftUUID,
		Username:      pl.LatestUsername(),
		Punishment:    toSimplePunishmentProto(sp),
	}}, nil)
}

func (svc *Service) publishPunishmentModified(ctx context.Context, srv *models.Server, pl *Player, pun *Punishment) {
	if svc.hub == nil {
		return
	}
	svc.hub.InvalidatePanel(srv.ID, modlv1.PanelResource_PANEL_RESOURCE_PLAYERS, pl.MinecraftUUID)
	svc.hub.InvalidatePanel(srv.ID, modlv1.PanelResource_PANEL_RESOURCE_PUNISHMENTS, pun.ID)
	mods := make([]*modlv1.SyncPunishmentModification, 0, len(pun.Modifications))
	for _, m := range pun.Modifications {
		var effPtr *int64
		if m.EffectiveDuration != nil {
			v := *m.EffectiveDuration
			effPtr = &v
		}
		var tsPtr *int64
		if m.DateMs > 0 {
			ts := m.DateMs
			tsPtr = &ts
		}
		mods = append(mods, &modlv1.SyncPunishmentModification{
			Type:              m.Type,
			Timestamp:         tsPtr,
			EffectiveDuration: effPtr,
		})
	}
	svc.hub.PushPunishments(srv.ID, nil, []*modlv1.SyncModifiedPunishment{{
		MinecraftUuid: pl.MinecraftUUID,
		Username:      pl.LatestUsername(),
		Punishment: &modlv1.SyncPunishmentWithModifications{
			Id:            pun.ID,
			Modifications: mods,
		},
	}})
}

func (svc *Service) sendWebhookAsync(ctx context.Context, srv *models.Server, vars map[string]string) {
	ws := svc.store.GetWebhookSettings(ctx, srv.ID)
	if ws == nil || !ws.Enabled || ws.Notifications == nil || !ws.Notifications.NewPunishments {
		return
	}
	targetURL := strings.TrimSpace(ws.DiscordWebhookUrl)
	if targetURL == "" || !ValidateSafeURL(targetURL) {
		return
	}
	desc := fmt.Sprintf("**Player:** %s\n**Type:** %s\n**Duration:** %s\n**Reason:** %s\n**Issuer:** %s",
		vars["playerName"], vars["type"], vars["duration"], vars["reason"], vars["issuer"])
	payload := map[string]any{
		"embeds": []map[string]any{{
			"title":       "Punishment Issued (" + vars["id"] + ")",
			"description": desc,
			"color":       0xE74C3C,
		}},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}
	go func() {
		reqCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, targetURL, bytes.NewReader(b))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := svc.httpClient.Do(req)
		if err == nil {
			_ = resp.Body.Close()
		}
	}()
}

// --- Pure Helpers ---

func indexTypesByOrdinal(types []PunishmentType) map[int]PunishmentType {
	m := make(map[int]PunishmentType, len(types))
	for _, t := range types {
		m[t.Ordinal] = t
	}
	return m
}

func storedIssuer(issuerID, issuerName string) string {
	if strings.TrimSpace(issuerID) != "" {
		return ""
	}
	return issuerName
}

func resolveIssuer(issuerID, issuerName string, resolved map[string]string) string {
	if issuerID != "" {
		if v, ok := resolved[issuerID]; ok && v != "" {
			return v
		}
	}
	if issuerName != "" {
		return issuerName
	}
	if issuerID != "" {
		return "Unknown Staff"
	}
	return "Console"
}

func collectIssuerIDs(p *Punishment) []string {
	var ids []string
	if p.IssuerID != "" {
		ids = append(ids, p.IssuerID)
	}
	for _, m := range p.Modifications {
		if m.IssuerID != "" {
			ids = append(ids, m.IssuerID)
		}
	}
	for _, n := range p.Notes {
		if n.IssuerID != "" {
			ids = append(ids, n.IssuerID)
		}
	}
	for _, e := range p.Evidence {
		if e.UploadedByID != nil && *e.UploadedByID != "" {
			ids = append(ids, *e.UploadedByID)
		}
	}
	return ids
}

func collectIssuerIDsForPlayers(players []*Player) []string {
	var ids []string
	for _, pl := range players {
		for _, n := range pl.Notes {
			if n.IssuerID != "" {
				ids = append(ids, n.IssuerID)
			}
		}
		for _, p := range pl.Punishments {
			ids = append(ids, collectIssuerIDs(p)...)
		}
	}
	return ids
}

func containsString(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}
