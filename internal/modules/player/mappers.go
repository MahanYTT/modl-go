package player

import (
	"encoding/json"
	"fmt"
	"time"

	modlv1 "modl-go/gen/modl/v1"

	"google.golang.org/protobuf/types/known/structpb"
)

func epochISO(ms int64) string {
	if ms == 0 {
		return ""
	}
	return time.UnixMilli(ms).UTC().Format(time.RFC3339Nano)
}

func mapToStruct(m map[string]any) *structpb.Struct {
	if len(m) == 0 {
		return &structpb.Struct{Fields: map[string]*structpb.Value{}}
	}
	clean := sanitizeStructMap(m)
	st, err := structpb.NewStruct(clean)
	if err != nil {
		return &structpb.Struct{Fields: map[string]*structpb.Value{}}
	}
	return st
}

func sanitizeStructMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = sanitizeStructValue(v)
	}
	return out
}

func sanitizeStructValue(v any) any {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case bool, string, float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int32:
		return float64(t)
	case int64:
		return float64(t)
	case uint:
		return float64(t)
	case uint64:
		return float64(t)
	case *string:
		if t == nil {
			return nil
		}
		return *t
	case *int64:
		if t == nil {
			return nil
		}
		return float64(*t)
	case *int:
		if t == nil {
			return nil
		}
		return float64(*t)
	case *bool:
		if t == nil {
			return nil
		}
		return *t
	case map[string]any:
		return sanitizeStructMap(t)
	case []string:
		arr := make([]any, len(t))
		for i, s := range t {
			arr[i] = s
		}
		return arr
	case []int64:
		arr := make([]any, len(t))
		for i, n := range t {
			arr[i] = float64(n)
		}
		return arr
	case []map[string]any:
		arr := make([]any, len(t))
		for i, m := range t {
			arr[i] = sanitizeStructMap(m)
		}
		return arr
	case []any:
		arr := make([]any, len(t))
		for i, item := range t {
			arr[i] = sanitizeStructValue(item)
		}
		return arr
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return fmt.Sprint(t)
		}
		var generic any
		if json.Unmarshal(b, &generic) == nil {
			return sanitizeStructValue(generic)
		}
		return fmt.Sprint(t)
	}
}

func structToMap(s *structpb.Struct) map[string]any {
	if s == nil {
		return nil
	}
	return s.AsMap()
}

// --- Panel Proto Mappers ---

func toPlayerSearchResultsProto(items []PlayerSearchResult) *modlv1.PlayerSearchResultsResponse {
	out := &modlv1.PlayerSearchResultsResponse{
		Items: make([]*modlv1.PlayerSearchResult, 0, len(items)),
	}
	for _, it := range items {
		var lastOnline string
		if it.LastOnline != nil {
			lastOnline = epochISO(*it.LastOnline)
		}
		out.Items = append(out.Items, &modlv1.PlayerSearchResult{
			Uuid:       it.UUID,
			Username:   it.Username,
			Status:     it.Status,
			LastOnline: lastOnline,
			IsOnline:   it.IsOnline,
		})
	}
	return out
}

func toPlayerDetailProto(d *PlayerDetailView) *modlv1.PlayerDetailResponse {
	out := &modlv1.PlayerDetailResponse{
		Id:             d.ID,
		MinecraftUuid:  d.MinecraftUUID,
		Usernames:      make([]*modlv1.UsernameEntry, 0, len(d.Usernames)),
		Notes:          make([]*modlv1.BackendNoteEntry, 0, len(d.Notes)),
		IpAddresses:    make([]*modlv1.IPEntry, 0, len(d.IPAddresses)),
		Punishments:    make([]*modlv1.PunishmentResponse, 0, len(d.Punishments)),
		Data:           mapToStruct(d.Data),
		Social:         d.Social,
		Gameplay:       d.Gameplay,
		SocialPoints:   int32(d.SocialPoints),
		GameplayPoints: int32(d.GameplayPoints),
		LastServer:     d.LastServer,
		PlaytimeHours:  d.PlaytimeHours,
	}
	for _, u := range d.Usernames {
		out.Usernames = append(out.Usernames, &modlv1.UsernameEntry{
			Username: u.Username,
			Date:     epochISO(u.DateMs),
		})
	}
	for _, n := range d.Notes {
		out.Notes = append(out.Notes, &modlv1.BackendNoteEntry{
			Id:         n.ID,
			Text:       n.Text,
			Date:       epochISO(n.DateMs),
			IssuerName: n.IssuerName,
			IssuerId:   n.IssuerID,
		})
	}
	for _, ip := range d.IPAddresses {
		out.IpAddresses = append(out.IpAddresses, toIPEntryProto(ip))
	}
	for _, p := range d.Punishments {
		out.Punishments = append(out.Punishments, toPunishmentResponseProto(p))
	}
	if d.LatestIPData != nil {
		out.LatestIpData = toIPEntryProto(*d.LatestIPData)
	}
	return out
}

func toIPEntryProto(ip IPEntry) *modlv1.IPEntry {
	logins := make([]string, 0, len(ip.LoginsMs))
	for _, ms := range ip.LoginsMs {
		logins = append(logins, epochISO(ms))
	}
	var ipPtr *string
	if ip.IPAddress != "" {
		s := ip.IPAddress
		ipPtr = &s
	}
	return &modlv1.IPEntry{
		IpAddress:  ipPtr,
		Country:    ip.Country,
		Region:     ip.Region,
		Asn:        ip.ASN,
		Proxy:      ip.Proxy,
		Hosting:    ip.Hosting,
		FirstLogin: epochISO(ip.FirstLoginMs),
		Logins:     logins,
	}
}

func toPunishmentResponseProto(p PunishmentResponseView) *modlv1.PunishmentResponse {
	mods := make([]*modlv1.PunishmentModification, 0, len(p.Modifications))
	for _, m := range p.Modifications {
		var inPtr, iidPtr, appPtr *string
		if m.IssuerName != "" {
			s := m.IssuerName
			inPtr = &s
		}
		if m.IssuerID != "" {
			s := m.IssuerID
			iidPtr = &s
		}
		if m.AppealTicketID != "" {
			s := m.AppealTicketID
			appPtr = &s
		}
		var dStruct *structpb.Struct
		if m.Data != nil {
			dStruct = mapToStruct(m.Data)
		}
		mods = append(mods, &modlv1.PunishmentModification{
			Id:                m.ID,
			Type:              m.Type,
			Date:              m.DateMs,
			IssuerName:        inPtr,
			IssuerId:          iidPtr,
			Reason:            m.Reason,
			EffectiveDuration: m.EffectiveDuration,
			AppealTicketId:    appPtr,
			Data:              dStruct,
		})
	}
	notes := make([]*modlv1.PunishmentNote, 0, len(p.Notes))
	for _, n := range p.Notes {
		var inPtr, iidPtr *string
		if n.IssuerName != "" {
			s := n.IssuerName
			inPtr = &s
		}
		if n.IssuerID != "" {
			s := n.IssuerID
			iidPtr = &s
		}
		notes = append(notes, &modlv1.PunishmentNote{
			Id:         n.ID,
			Text:       n.Text,
			Date:       n.DateMs,
			IssuerName: inPtr,
			IssuerId:   iidPtr,
		})
	}
	evs := make([]*modlv1.PunishmentEvidence, 0, len(p.Evidence))
	for _, e := range p.Evidence {
		evs = append(evs, &modlv1.PunishmentEvidence{
			Type:         e.Type,
			Text:         e.Text,
			Url:          e.URL,
			UploadedBy:   e.UploadedBy,
			UploadedById: e.UploadedByID,
			UploadedAt:   e.UploadedAtMs,
			FileName:     e.FileName,
			FileType:     e.FileType,
			FileSize:     e.FileSize,
		})
	}
	return &modlv1.PunishmentResponse{
		Id:                p.ID,
		Type:              p.Type,
		TypeOrdinal:       int32(p.TypeOrdinal),
		IssuerName:        p.IssuerName,
		Issued:            p.IssuedMs,
		Started:           p.StartedMs,
		IsAppealable:      p.IsAppealable,
		Reason:            p.Reason,
		Severity:          p.Severity,
		Status:            p.Status,
		Active:            p.Active,
		Expires:           p.ExpiresMs,
		PlayerUuid:        p.PlayerUUID,
		PlayerUsername:    p.PlayerUsername,
		AltBlocking:       p.AltBlocking,
		StatWiping:        p.StatWiping,
		EffectiveCategory: p.EffectiveCategory,
		Modifications:     mods,
		Notes:             notes,
		Evidence:          evs,
		AttachedTicketIds: p.AttachedTicketIds,
	}
}

func toPunishmentPreviewProto(p PunishmentPreview) *modlv1.PunishmentPreviewResponse {
	convSev := func(s *SeverityPreview) *modlv1.PunishmentPreviewResponse_SeverityPreview {
		if s == nil {
			return nil
		}
		return &modlv1.PunishmentPreviewResponse_SeverityPreview{
			Severity:          s.Severity,
			DurationFormatted: s.DurationFormatted,
			PunishmentType:    s.PunishmentType,
			NewSocialStatus:   s.NewSocialStatus,
			NewGameplayStatus: s.NewGameplayStatus,
			Permanent:         s.Permanent,
			Points:            int32(s.Points),
			NewSocialPoints:   int32(s.NewSocialPoints),
			NewGameplayPoints: int32(s.NewGameplayPoints),
			DurationMs:        s.DurationMs,
		}
	}
	return &modlv1.PunishmentPreviewResponse{
		Message:                      p.Message,
		SocialStatus:                 p.SocialStatus,
		GameplayStatus:               p.GameplayStatus,
		OffenderStatus:               p.OffenderStatus,
		Category:                     p.Category,
		Lenient:                      convSev(p.Lenient),
		Regular:                      convSev(p.Regular),
		Aggravated:                   convSev(p.Aggravated),
		SingleSeverity:               convSev(p.SingleSeverity),
		Success:                      p.Success,
		SingleSeverityPunishment:     p.SingleSeverityPunishment,
		PermanentUntilUsernameChange: p.PermanentUntilUsernameChange,
		PermanentUntilSkinChange:     p.PermanentUntilSkinChange,
		CanBeAltBlocking:             p.CanBeAltBlocking,
		CanBeStatWiping:              p.CanBeStatWiping,
		Status:                       int32(p.Status),
		SocialPoints:                 int32(p.SocialPoints),
		GameplayPoints:               int32(p.GameplayPoints),
	}
}

func toPanelLinkedAccountsProto(items []LinkedAccountView) *modlv1.PanelLinkedAccountsResponse {
	out := &modlv1.PanelLinkedAccountsResponse{
		LinkedAccounts: make([]*modlv1.LinkedAccountResponse, 0, len(items)),
	}
	for _, it := range items {
		var llu *string
		if it.LastLinkedUpdate != nil {
			s := epochISO(*it.LastLinkedUpdate)
			llu = &s
		}
		out.LinkedAccounts = append(out.LinkedAccounts, &modlv1.LinkedAccountResponse{
			MinecraftUuid:    it.MinecraftUUID,
			Username:         it.Username,
			ActiveBans:       int32(it.ActiveBans),
			ActiveMutes:      int32(it.ActiveMutes),
			LastLinkedUpdate: llu,
		})
	}
	return out
}

func toPanelLinkedBansProto(items []LinkedBanView) *modlv1.PanelLinkedBansResponse {
	out := &modlv1.PanelLinkedBansResponse{
		LinkedBans: make([]*modlv1.PanelLinkedBanEntry, 0, len(items)),
	}
	for _, it := range items {
		out.LinkedBans = append(out.LinkedBans, &modlv1.PanelLinkedBanEntry{
			PunishmentId: it.PunishmentID,
			PlayerUuid:   it.PlayerUUID,
			PlayerName:   it.PlayerName,
			Active:       it.Active,
		})
	}
	return out
}

func toPublicAppealInfoProto(v *AppealInfoView) *modlv1.PublicPunishmentAppealInfoResponse {
	resp := &modlv1.PublicPunishmentAppealInfoResponse{
		Id:         v.ID,
		Type:       v.Type,
		Issued:     v.IssuedMs,
		Expires:    v.ExpiresMs,
		Active:     v.Active,
		Appealable: v.Appealable,
		PlayerUuid: v.PlayerUUID,
	}
	if v.ExistingAppeal != nil {
		resp.ExistingAppeal = mapToStruct(v.ExistingAppeal)
	}
	if v.AppealForm != nil {
		resp.AppealForm = mapToStruct(v.AppealForm)
	}
	return resp
}

// --- Minecraft V3 Proto Mappers ---

func toSimplePunishmentProto(sp SimplePunishmentView) *modlv1.SimplePunishment {
	var catPtr, issPtr, pdPtr *string
	if sp.Category != "" {
		c := sp.Category
		catPtr = &c
	}
	if sp.IssuerName != "" {
		i := sp.IssuerName
		issPtr = &i
	}
	if sp.PlayerDescription != nil {
		pd := *sp.PlayerDescription
		pdPtr = &pd
	}
	ia := sp.IssuedAt
	return &modlv1.SimplePunishment{
		Type:              sp.Type,
		Category:          catPtr,
		Expiration:        sp.Expiration,
		Description:       sp.Description,
		Id:                sp.ID,
		IssuerName:        issPtr,
		IssuedAt:          &ia,
		PlayerDescription: pdPtr,
		Started:           sp.Started,
		Ordinal:           int32(sp.Ordinal),
	}
}

func toPlayerLoginResponseProto(res *LoginResult) *modlv1.PlayerLoginResponse {
	pb := &modlv1.PlayerLoginResponse{
		Status:               int32(res.HTTPStatus),
		ActivePunishments:    make([]*modlv1.SimplePunishment, 0, len(res.ActivePunishments)),
		PendingNotifications: make([]*structpb.Struct, 0, len(res.PendingNotifications)),
		PendingIpLookups:     res.PendingIPLookups,
		PendingStatWipes:     make([]*modlv1.PendingStatWipe, 0, len(res.PendingStatWipes)),
	}
	for _, sp := range res.ActivePunishments {
		pb.ActivePunishments = append(pb.ActivePunishments, toSimplePunishmentProto(sp))
	}
	for _, pn := range res.PendingNotifications {
		pb.PendingNotifications = append(pb.PendingNotifications, mapToStruct(pn))
	}
	for _, sw := range res.PendingStatWipes {
		u, _ := sw["minecraftUuid"].(string)
		un, _ := sw["username"].(string)
		pid, _ := sw["punishmentId"].(string)
		pb.PendingStatWipes = append(pb.PendingStatWipes, &modlv1.PendingStatWipe{
			MinecraftUuid: u,
			Username:      un,
			PunishmentId:  pid,
		})
	}
	return pb
}

func toAccountProto(profile map[string]any) *modlv1.Account {
	acc := &modlv1.Account{
		Id:            strFromMap(profile, "id"),
		MinecraftUuid: strFromMap(profile, "minecraftUuid"),
	}
	if usernames, ok := profile["usernames"].([]map[string]any); ok {
		for _, u := range usernames {
			ms, _ := parseFlexibleTimeMs(u["date"])
			acc.Usernames = append(acc.Usernames, &modlv1.UsernameEntry{
				Username: strFromMap(u, "username"),
				Date:     epochISO(ms),
			})
		}
	}
	if notes, ok := profile["notes"].([]map[string]any); ok {
		for _, n := range notes {
			ms, _ := parseFlexibleTimeMs(n["date"])
			acc.Notes = append(acc.Notes, &modlv1.NoteEntry{
				Id:         strFromMap(n, "id"),
				Text:       strFromMap(n, "text"),
				Date:       epochISO(ms),
				IssuerName: strFromMap(n, "issuerName"),
				IssuerId:   strFromMap(n, "issuerId"),
			})
		}
	}
	if ips, ok := profile["ipAddresses"].([]map[string]any); ok {
		for _, ip := range ips {
			entry := &modlv1.IPEntry{}
			if v, ok := ip["ipAddress"].(string); ok && v != "" {
				entry.IpAddress = &v
			}
			if v, ok := ip["country"].(*string); ok && v != nil {
				entry.Country = v
			} else if v, ok := ip["country"].(string); ok && v != "" {
				entry.Country = &v
			}
			if v, ok := ip["region"].(*string); ok && v != nil {
				entry.Region = v
			} else if v, ok := ip["region"].(string); ok && v != "" {
				entry.Region = &v
			}
			if v, ok := ip["asn"].(*string); ok && v != nil {
				entry.Asn = v
			} else if v, ok := ip["asn"].(string); ok && v != "" {
				entry.Asn = &v
			}
			entry.Proxy, _ = ip["proxy"].(bool)
			entry.Hosting, _ = ip["hosting"].(bool)
			if ms, ok := parseFlexibleTimeMs(ip["firstLogin"]); ok {
				entry.FirstLogin = epochISO(ms)
			}
			if logins, ok := ip["logins"].([]int64); ok {
				for _, ms := range logins {
					entry.Logins = append(entry.Logins, epochISO(ms))
				}
			}
			acc.IpAddresses = append(acc.IpAddresses, entry)
		}
	}
	if puns, ok := profile["punishments"].([]PunishmentView); ok {
		for _, p := range puns {
			var issued int64
			if p.Issued != nil {
				issued = *p.Issued
			}
			pr := &modlv1.PunishmentResponse{
				Id:                p.ID,
				Type:              p.Type,
				TypeOrdinal:       int32(p.TypeOrdinal),
				IssuerName:        p.IssuerName,
				Issued:            issued,
				Started:           p.Started,
				PlayerUuid:        p.PlayerUUID,
				AttachedTicketIds: p.AttachedTicketIds,
			}
			acc.Punishments = append(acc.Punishments, pr)
		}
	}
	if notifs, ok := profile["pendingNotifications"].([]map[string]any); ok {
		for _, n := range notifs {
			acc.PendingNotifications = append(acc.PendingNotifications, mapToStruct(n))
		}
	}
	if data, ok := profile["data"].(map[string]any); ok {
		acc.Data = mapToStruct(data)
	}
	return acc
}

func toPlayerProfileResponseProto(profile map[string]any) *modlv1.PlayerProfileResponse {
	resp := &modlv1.PlayerProfileResponse{
		Status:  200,
		Profile: toAccountProto(profile),
	}
	if pc, ok := profile["punishmentCount"].(int); ok {
		v := int32(pc)
		resp.PunishmentCount = &v
	}
	if nc, ok := profile["noteCount"].(int); ok {
		v := int32(nc)
		resp.NoteCount = &v
	}
	return resp
}

func toPlayerLookupResponseProto(msg string, data map[string]any) *modlv1.PlayerLookupResponse {
	ld := &modlv1.PlayerLookupResponse_PlayerLookupData{
		MinecraftUuid:   strFromMap(data, "minecraftUuid"),
		CurrentUsername: strFromMap(data, "currentUsername"),
		CurrentServer:   strFromMap(data, "currentServer"),
		IpAddress:       strFromMap(data, "ipAddress"),
		Country:         strFromMap(data, "country"),
		ProfileUrl:      strFromMap(data, "profileUrl"),
		PunishmentsUrl:  strFromMap(data, "punishmentsUrl"),
		TicketsUrl:      strFromMap(data, "ticketsUrl"),
	}
	ld.IsOnline, _ = data["isOnline"].(bool)
	if ms, ok := parseFlexibleTimeMs(data["firstSeen"]); ok {
		ld.FirstSeen = epochISO(ms)
	}
	if ms, ok := parseFlexibleTimeMs(data["lastSeen"]); ok {
		ld.LastSeen = epochISO(ms)
	}
	if prev, ok := data["previousUsernames"].([]string); ok {
		ld.PreviousUsernames = prev
	}
	if stats, ok := data["punishmentStats"].(map[string]any); ok {
		ps := &modlv1.PlayerLookupResponse_PlayerLookupPunishmentStats{}
		if v, ok := stats["status"].(string); ok {
			ps.Status = v
		}
		if v, ok := asInt64(stats["totalPunishments"]); ok {
			ps.TotalPunishments = int32(v)
		}
		if v, ok := asInt64(stats["activePunishments"]); ok {
			ps.ActivePunishments = int32(v)
		}
		if v, ok := asInt64(stats["bans"]); ok {
			ps.Bans = int32(v)
		}
		if v, ok := asInt64(stats["mutes"]); ok {
			ps.Mutes = int32(v)
		}
		if v, ok := asInt64(stats["kicks"]); ok {
			ps.Kicks = int32(v)
		}
		if v, ok := asInt64(stats["warnings"]); ok {
			ps.Warnings = int32(v)
		}
		if v, ok := asInt64(stats["points"]); ok {
			ps.Points = int32(v)
		}
		ld.PunishmentStats = ps
	}
	if recPuns, ok := data["recentPunishments"].([]map[string]any); ok {
		for _, rp := range recPuns {
			item := &modlv1.PlayerLookupResponse_PlayerLookupRecentPunishment{
				Id:     strFromMap(rp, "id"),
				Type:   strFromMap(rp, "type"),
				Issuer: strFromMap(rp, "issuer"),
			}
			item.IsActive, _ = rp["isActive"].(bool)
			if ms, ok := parseFlexibleTimeMs(rp["issuedAt"]); ok {
				item.IssuedAt = epochISO(ms)
			}
			if ms, ok := parseFlexibleTimeMs(rp["expiresAt"]); ok {
				item.ExpiresAt = epochISO(ms)
			}
			ld.RecentPunishments = append(ld.RecentPunishments, item)
		}
	}
	return &modlv1.PlayerLookupResponse{
		Status:  200,
		Message: msg,
		Data:    ld,
	}
}

func toPunishmentListEntryProto(p PunishmentView) *modlv1.PunishmentListEntry {
	var issued int64
	if p.Issued != nil {
		issued = *p.Issued
	}
	ord := int32(p.TypeOrdinal)
	entry := &modlv1.PunishmentListEntry{
		Id:                p.ID,
		IssuerName:        p.IssuerName,
		Issued:            issued,
		Started:           p.Started,
		Type:              p.Type,
		TypeOrdinal:       &ord,
		AttachedTicketIds: p.AttachedTicketIds,
		Data:              mapToStruct(p.Data),
	}
	for _, m := range p.Modifications {
		entry.Modifications = append(entry.Modifications, mapToModProto(m))
	}
	for _, n := range p.Notes {
		entry.Notes = append(entry.Notes, mapToNoteProto(n))
	}
	for _, e := range p.Evidence {
		entry.Evidence = append(entry.Evidence, mapToEvidenceProto(e))
	}
	return entry
}

func toRecentPunishmentProto(p PunishmentView) *modlv1.RecentPunishmentsResponse_RecentPunishment {
	var issued int64
	if p.Issued != nil {
		issued = *p.Issued
	}
	ord := int32(p.TypeOrdinal)
	pUUID, pName := "", ""
	if p.PlayerUUID != nil {
		pUUID = *p.PlayerUUID
	}
	if p.PlayerName != nil {
		pName = *p.PlayerName
	}
	entry := &modlv1.RecentPunishmentsResponse_RecentPunishment{
		PlayerName:        pName,
		PlayerUuid:        pUUID,
		Id:                p.ID,
		IssuerName:        p.IssuerName,
		Issued:            issued,
		Started:           p.Started,
		Type:              p.Type,
		TypeOrdinal:       &ord,
		AttachedTicketIds: p.AttachedTicketIds,
		Data:              mapToStruct(p.Data),
	}
	for _, m := range p.Modifications {
		entry.Modifications = append(entry.Modifications, mapToModProto(m))
	}
	for _, n := range p.Notes {
		entry.Notes = append(entry.Notes, mapToNoteProto(n))
	}
	for _, e := range p.Evidence {
		entry.Evidence = append(entry.Evidence, mapToEvidenceProto(e))
	}
	return entry
}

func toPunishmentDetailEntryProto(p PunishmentView) *modlv1.PunishmentDetailResponse_PunishmentDetailEntry {
	pUUID, pName := "", ""
	if p.PlayerUUID != nil {
		pUUID = *p.PlayerUUID
	}
	if p.PlayerName != nil {
		pName = *p.PlayerName
	}
	var issuedStr, startedStr string
	if p.Issued != nil {
		issuedStr = epochISO(*p.Issued)
	}
	if p.Started != nil {
		startedStr = epochISO(*p.Started)
	}
	entry := &modlv1.PunishmentDetailResponse_PunishmentDetailEntry{
		Id:                p.ID,
		PlayerUuid:        pUUID,
		PlayerName:        pName,
		IssuerName:        p.IssuerName,
		Issued:            issuedStr,
		Started:           startedStr,
		Type:              p.Type,
		TypeOrdinal:       int32(p.TypeOrdinal),
		AttachedTicketIds: p.AttachedTicketIds,
		Data:              mapToStruct(p.Data),
	}
	for _, m := range p.Modifications {
		entry.Modifications = append(entry.Modifications, mapToStruct(m))
	}
	for _, n := range p.Notes {
		entry.Notes = append(entry.Notes, mapToStruct(n))
	}
	for _, e := range p.Evidence {
		entry.Evidence = append(entry.Evidence, mapToStruct(e))
	}
	return entry
}

func mapToModProto(m map[string]any) *modlv1.PunishmentModification {
	dateMs, _ := asInt64(m["date"])
	pm := &modlv1.PunishmentModification{
		Id:     strFromMap(m, "id"),
		Type:   strFromMap(m, "type"),
		Date:   dateMs,
		Reason: strFromMap(m, "reason"),
	}
	if v, ok := m["issuerName"].(string); ok && v != "" {
		pm.IssuerName = &v
	}
	if v, ok := m["issuerId"].(string); ok && v != "" {
		pm.IssuerId = &v
	}
	if v, ok := m["effectiveDuration"].(*int64); ok && v != nil {
		pm.EffectiveDuration = v
	} else if v, ok := asInt64(m["effectiveDuration"]); ok {
		pm.EffectiveDuration = &v
	}
	if v, ok := m["appealTicketId"].(string); ok && v != "" {
		pm.AppealTicketId = &v
	}
	if d, ok := m["data"].(map[string]any); ok && d != nil {
		pm.Data = mapToStruct(d)
	}
	return pm
}

func mapToNoteProto(n map[string]any) *modlv1.PunishmentNote {
	dateMs, _ := asInt64(n["date"])
	pn := &modlv1.PunishmentNote{
		Id:   strFromMap(n, "id"),
		Text: strFromMap(n, "text"),
		Date: dateMs,
	}
	if v, ok := n["issuerName"].(string); ok && v != "" {
		pn.IssuerName = &v
	}
	if v, ok := n["issuerId"].(string); ok && v != "" {
		pn.IssuerId = &v
	}
	return pn
}

func mapToEvidenceProto(e map[string]any) *modlv1.PunishmentEvidence {
	upMs, _ := asInt64(e["uploadedAt"])
	pe := &modlv1.PunishmentEvidence{
		Type:       strFromMap(e, "type"),
		UploadedAt: upMs,
	}
	setStr := func(dst **string, key string) {
		if ptr, ok := e[key].(*string); ok && ptr != nil {
			*dst = ptr
		} else if s, ok := e[key].(string); ok && s != "" {
			*dst = &s
		}
	}
	setStr(&pe.Text, "text")
	setStr(&pe.Url, "url")
	setStr(&pe.UploadedBy, "uploadedBy")
	setStr(&pe.UploadedById, "uploadedById")
	setStr(&pe.FileName, "fileName")
	setStr(&pe.FileType, "fileType")
	if ptr, ok := e["fileSize"].(*int64); ok && ptr != nil {
		pe.FileSize = ptr
	} else if v, ok := asInt64(e["fileSize"]); ok {
		pe.FileSize = &v
	}
	return pe
}

func toReportsResponseProto(reports []map[string]any) *modlv1.ReportsResponse {
	out := &modlv1.ReportsResponse{
		Status:  200,
		Reports: make([]*modlv1.ReportEntry, 0, len(reports)),
	}
	for _, r := range reports {
		createdAt, _ := asInt64(r["createdAt"])
		re := &modlv1.ReportEntry{
			Id:                 strFromMap(r, "id"),
			Type:               strFromMap(r, "type"),
			Category:           strFromMap(r, "category"),
			ReporterName:       strFromMap(r, "reporterName"),
			ReporterUuid:       strFromMap(r, "reporterUuid"),
			ReportedPlayerUuid: strFromMap(r, "reportedPlayerUuid"),
			ReportedPlayerName: strFromMap(r, "reportedPlayerName"),
			Subject:            strFromMap(r, "subject"),
			Content:            strFromMap(r, "content"),
			Status:             strFromMap(r, "status"),
			Priority:           strFromMap(r, "priority"),
			CreatedAt:          createdAt,
		}
		if assigned, ok := r["assignedTo"].([]string); ok {
			re.AssignedTo = assigned
		}
		if msgs, ok := r["chatMessages"].([]map[string]any); ok {
			for _, m := range msgs {
				re.ChatMessages = append(re.ChatMessages, mapToStruct(m))
			}
		}
		out.Reports = append(out.Reports, re)
	}
	return out
}

func toSyncResponseProto(res SyncResult) *modlv1.SyncResponse {
	d := res.Data
	sd := &modlv1.SyncData{
		StaffPermissionsUpdatedAt: d.StaffPermissionsUpdatedAt,
		PunishmentTypesUpdatedAt:  d.PunishmentTypesUpdatedAt,
	}
	for _, p := range d.PendingPunishments {
		sd.PendingPunishments = append(sd.PendingPunishments, &modlv1.SyncPendingPunishment{
			MinecraftUuid: p.MinecraftUUID,
			Username:      p.Username,
			Punishment:    toSimplePunishmentProto(p.Punishment),
		})
	}
	for _, p := range d.RecentlyStartedPunishments {
		sd.RecentlyStartedPunishments = append(sd.RecentlyStartedPunishments, &modlv1.SyncPendingPunishment{
			MinecraftUuid: p.MinecraftUUID,
			Username:      p.Username,
			Punishment:    toSimplePunishmentProto(p.Punishment),
		})
	}
	for _, p := range d.RecentlyModifiedPunishments {
		mods := make([]*modlv1.SyncPunishmentModification, 0, len(p.Punishment.Modifications))
		for _, m := range p.Punishment.Modifications {
			eff := m.EffectiveDuration
			mods = append(mods, &modlv1.SyncPunishmentModification{
				Type:              m.Type,
				Timestamp:         m.Timestamp,
				EffectiveDuration: &eff,
			})
		}
		sd.RecentlyModifiedPunishments = append(sd.RecentlyModifiedPunishments, &modlv1.SyncModifiedPunishment{
			MinecraftUuid: p.MinecraftUUID,
			Username:      p.Username,
			Punishment: &modlv1.SyncPunishmentWithModifications{
				Id:            p.Punishment.ID,
				Modifications: mods,
			},
		})
	}
	for _, pn := range d.PlayerNotifications {
		spn := &modlv1.SyncPlayerNotification{
			Id:      strFromMap(pn, "id"),
			Message: strFromMap(pn, "message"),
			Type:    strFromMap(pn, "type"),
		}
		if v, ok := pn["targetPlayerUuid"].(string); ok && v != "" {
			spn.TargetPlayerUuid = &v
		}
		if ts, ok := asInt64(pn["timestamp"]); ok {
			spn.Timestamp = &ts
		}
		if dm, ok := pn["data"].(map[string]any); ok {
			spn.Data = mapToStruct(dm)
		}
		sd.PlayerNotifications = append(sd.PlayerNotifications, spn)
	}
	for _, sn := range d.StaffNotifications {
		ssn := &modlv1.SyncStaffNotification{
			Id:      strFromMap(sn, "id"),
			Type:    strFromMap(sn, "type"),
			Message: strFromMap(sn, "message"),
		}
		if ts, ok := asInt64(sn["timestamp"]); ok {
			ssn.Timestamp = &ts
		}
		if dm, ok := sn["data"].(map[string]any); ok {
			ssn.Data = mapToStruct(dm)
		}
		sd.StaffNotifications = append(sd.StaffNotifications, ssn)
	}
	for _, st := range d.ActiveStaffMembers {
		asm := &modlv1.SyncActiveStaffMember{
			MinecraftUuid:     strFromMap(st, "minecraftUuid"),
			MinecraftUsername: strFromMap(st, "minecraftUsername"),
			StaffUsername:     strFromMap(st, "staffUsername"),
			StaffRole:         strFromMap(st, "staffRole"),
			Email:             strFromMap(st, "email"),
			StaffId:           strFromMap(st, "staffId"),
		}
		if perms, ok := st["permissions"].([]string); ok {
			asm.Permissions = perms
		}
		if b, ok := st["twoFactorSessionValid"].(bool); ok {
			asm.TwoFactorSessionValid = &b
		}
		sd.ActiveStaffMembers = append(sd.ActiveStaffMembers, asm)
	}
	for _, sw := range d.PendingStatWipes {
		sd.PendingStatWipes = append(sd.PendingStatWipes, &modlv1.SyncPendingStatWipe{
			MinecraftUuid: strFromMap(sw, "minecraftUuid"),
			Username:      strFromMap(sw, "username"),
			PunishmentId:  strFromMap(sw, "punishmentId"),
		})
	}
	for _, v := range d.Staff2faVerifications {
		sd.Staff_2FaVerifications = append(sd.Staff_2FaVerifications, &modlv1.SyncStaff2FaVerification{
			MinecraftUuid: strFromMap(v, "minecraftUuid"),
		})
	}
	if d.MigrationTask != nil {
		sd.MigrationTask = &modlv1.SyncMigrationTask{
			TaskId: strFromMap(d.MigrationTask, "taskId"),
			Type:   strFromMap(d.MigrationTask, "type"),
		}
	}
	return &modlv1.SyncResponse{
		Timestamp: res.Timestamp,
		Data:      sd,
	}
}
