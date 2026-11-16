package staff

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/config"
	"modl-go/internal/email"
	"modl-go/internal/httpx"
	"modl-go/internal/models"
	"modl-go/internal/modules/role"
	"modl-go/internal/realtime"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	db     *sql.DB
	cfg    *config.Config
	mailer *email.Sender
	hub    *realtime.Hub
	roles  *role.Handler
}

func NewHandler(db *sql.DB, cfg *config.Config, mailer *email.Sender, hub *realtime.Hub, roles *role.Handler) *Handler {
	return &Handler{db: db, cfg: cfg, mailer: mailer, hub: hub, roles: roles}
}

func (h *Handler) RegisterPanelRoutes(r chi.Router) {
	r.Route("/staff", func(r chi.Router) {
		r.Use(httpx.RequirePanelPolicy(h.roles, models.AccessPolicy{
			Rule:   models.RuleRequirePermission,
			View:   "admin.staff.manage.members",
			Modify: "admin.staff.manage.members",
		}))
		r.Get("/", h.getAllPanelStaff)
		r.Get("/check-username/{username}", h.checkUsername)
		r.Get("/available-players", h.getAvailablePlayers)
		r.Get("/{username}", h.getStaffByUsername)
		r.Post("/", h.createStaff)
		r.Patch("/{username}", h.updateStaff)
		r.Patch("/{id}/role", h.updateStaffRole)
		r.Delete("/{id}", h.deleteStaff)
		r.Post("/invite", h.inviteStaff)
		r.Post("/invitations/{id}/resend", h.resendInvitation)
		r.Patch("/{email}/minecraft-player", h.assignMinecraftPlayer)
	})
}

func (h *Handler) RegisterPublicRoutes(r chi.Router) {
	r.Route("/staff", func(r chi.Router) {
		r.Get("/invitations/accept", h.acceptInvitationGet)
		r.Post("/invitations/accept", h.acceptInvitationPost)
		r.Post("/2fa/verify/{token}", h.verify2FAToken)
	})
}

func (h *Handler) RegisterMinecraftV1Routes(r chi.Router) {
	r.Route("/staff", func(r chi.Router) {
		r.Get("/", h.getMinecraftStaffV1)
		r.Get("/permissions", h.getMinecraftStaffPermsV1)
		r.Patch("/{id}/role", h.updateMinecraftStaffRoleV1)
		r.Post("/disconnect", h.staffDisconnectV1)
		r.Post("/2fa/generate", h.generate2FATokenV1)
	})
}

func (h *Handler) RegisterMinecraftV3Routes(r chi.Router) {
	r.Route("/staff", func(r chi.Router) {
		r.Get("/", h.getMinecraftStaffV3)
		r.Get("/permissions", h.getMinecraftStaffPermsV3)
		r.Patch("/{id}/role", h.updateMinecraftStaffRoleV3)
		r.Post("/disconnect", h.staffDisconnectV3)
	})
}

type staffRow struct {
	ID                        string
	ServerID                  string
	Email                     string
	Username                  string
	RoleID                    string
	AssignedMinecraftUUID     string
	AssignedMinecraftUsername string
	Language                  string
	DateFormat                string
	LastSeen                  int64
	CreatedAt                 int64
	UpdatedAt                 int64
}

func (h *Handler) listStaff(ctx context.Context, serverID string) ([]staffRow, error) {
	rows, err := h.db.QueryContext(ctx, `
		SELECT id, server_id, email, username, role_id, assigned_minecraft_uuid, assigned_minecraft_username,
		       language, date_format, COALESCE(last_seen, 0), created_at, updated_at
		FROM staff WHERE server_id = ? ORDER BY created_at ASC`, serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []staffRow
	for rows.Next() {
		var s staffRow
		if err := rows.Scan(&s.ID, &s.ServerID, &s.Email, &s.Username, &s.RoleID, &s.AssignedMinecraftUUID,
			&s.AssignedMinecraftUsername, &s.Language, &s.DateFormat, &s.LastSeen, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (h *Handler) roleNameMap(ctx context.Context, serverID string) map[string]string {
	roles, _ := h.roles.EnsureDefaultRoles(ctx, serverID)
	m := make(map[string]string, len(roles))
	for _, rl := range roles {
		m[rl.ID] = rl.Name
	}
	return m
}

func (h *Handler) resolveRoleID(ctx context.Context, serverID, roleInput string) string {
	roleInput = strings.TrimSpace(roleInput)
	if roleInput == "" {
		return ""
	}
	roles, _ := h.roles.EnsureDefaultRoles(ctx, serverID)
	for _, rl := range roles {
		if rl.ID == roleInput || strings.EqualFold(rl.Name, roleInput) {
			return rl.ID
		}
	}
	return roleInput
}

func (h *Handler) toProtoStaff(s staffRow, roleNames map[string]string) *modlv1.StaffResponse {
	rName := roleNames[s.RoleID]
	if rName == "" {
		rName = s.RoleID
	}
	return &modlv1.StaffResponse{
		Id:                        s.ID,
		Email:                     s.Email,
		Username:                  s.Username,
		Role:                      rName,
		Status:                    "active",
		AssignedMinecraftUuid:     s.AssignedMinecraftUUID,
		AssignedMinecraftUsername: s.AssignedMinecraftUsername,
		CreatedAt:                 s.CreatedAt,
	}
}

func (h *Handler) getAllPanelStaff(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	list, err := h.listStaff(r.Context(), srv.ID)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load staff")
		return
	}
	roleNames := h.roleNameMap(r.Context(), srv.ID)
	pb := make([]*modlv1.StaffResponse, 0, len(list)+1)

	// Synthesize super-admin entry if adminEmail isn't already in staff table
	adminFound := false
	for _, s := range list {
		if strings.EqualFold(s.Email, srv.AdminEmail) {
			adminFound = true
		}
		pb = append(pb, h.toProtoStaff(s, roleNames))
	}
	if !adminFound && srv.AdminEmail != "" {
		pb = append([]*modlv1.StaffResponse{{
			Id:        role.SuperAdminRoleID,
			Email:     srv.AdminEmail,
			Username:  strings.Split(srv.AdminEmail, "@")[0],
			Role:      "Super Admin",
			Status:    "active",
			CreatedAt: srv.CreatedAt,
		}}, pb...)
	}

	// Include pending invitations
	invRows, err := h.db.QueryContext(r.Context(), `SELECT id, email, role_id, created_at FROM invitations WHERE server_id = ? AND expires_at > ?`, srv.ID, httpx.NowMs())
	if err == nil {
		defer invRows.Close()
		for invRows.Next() {
			var id, em, rID string
			var created int64
			if invRows.Scan(&id, &em, &rID, &created) == nil {
				rName := roleNames[rID]
				if rName == "" {
					rName = rID
				}
				pb = append(pb, &modlv1.StaffResponse{
					Id:        id,
					Email:     em,
					Username:  strings.Split(em, "@")[0],
					Role:      rName,
					Status:    "pending",
					CreatedAt: created,
				})
			}
		}
	}

	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PanelStaffListResponse{Staff: pb})
}

func (h *Handler) checkUsername(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	username := strings.TrimSpace(chi.URLParam(r, "username"))
	var c int
	_ = h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM staff WHERE server_id = ? AND LOWER(username) = ?`, srv.ID, strings.ToLower(username)).Scan(&c)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.CheckUsernameResponse{Exists: c > 0})
}

func (h *Handler) getStaffByUsername(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	username := strings.TrimSpace(chi.URLParam(r, "username"))
	list, _ := h.listStaff(r.Context(), srv.ID)
	roleNames := h.roleNameMap(r.Context(), srv.ID)
	for _, s := range list {
		if strings.EqualFold(s.Username, username) || s.ID == username || strings.EqualFold(s.Email, username) {
			httpx.WriteProto(w, r, http.StatusOK, h.toProtoStaff(s, roleNames))
			return
		}
	}
	httpx.WriteError(w, r, http.StatusNotFound, "Staff member not found")
}

func (h *Handler) createStaff(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.CreateStaffRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	em := strings.ToLower(strings.TrimSpace(req.Email))
	rID := ""
	if req.Role != nil {
		rID = h.resolveRoleID(r.Context(), srv.ID, *req.Role)
	}
	now := httpx.NowMs()
	s := staffRow{
		ID:         httpx.NewID(),
		ServerID:   srv.ID,
		Email:      em,
		Username:   strings.TrimSpace(req.Username),
		RoleID:     rID,
		Language:   "en",
		DateFormat: "MM/dd/yyyy",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	_, err := h.db.ExecContext(r.Context(), `
		INSERT INTO staff (id, server_id, email, username, role_id, language, date_format, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.ID, s.ServerID, s.Email, s.Username, s.RoleID, s.Language, s.DateFormat, s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		httpx.WriteError(w, r, http.StatusConflict, "Staff member already exists")
		return
	}
	h.invalidateStaff(srv.ID, s.ID)
	httpx.WriteProto(w, r, http.StatusCreated, h.toProtoStaff(s, h.roleNameMap(r.Context(), srv.ID)))
}

func (h *Handler) updateStaff(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	target := strings.TrimSpace(chi.URLParam(r, "username"))
	var req modlv1.UpdateStaffRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	list, _ := h.listStaff(r.Context(), srv.ID)
	var found *staffRow
	for i := range list {
		if strings.EqualFold(list[i].Username, target) || list[i].ID == target || strings.EqualFold(list[i].Email, target) {
			found = &list[i]
			break
		}
	}
	if found == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Staff member not found")
		return
	}
	if req.Email != nil && strings.TrimSpace(*req.Email) != "" {
		found.Email = strings.ToLower(strings.TrimSpace(*req.Email))
	}
	if req.Role != nil && strings.TrimSpace(*req.Role) != "" {
		found.RoleID = h.resolveRoleID(r.Context(), srv.ID, *req.Role)
	}
	found.UpdatedAt = httpx.NowMs()
	_, err := h.db.ExecContext(r.Context(), `UPDATE staff SET email = ?, role_id = ?, updated_at = ? WHERE server_id = ? AND id = ?`,
		found.Email, found.RoleID, found.UpdatedAt, srv.ID, found.ID)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to update staff")
		return
	}
	h.invalidateStaff(srv.ID, found.ID)
	httpx.WriteProto(w, r, http.StatusOK, h.toProtoStaff(*found, h.roleNameMap(r.Context(), srv.ID)))
}

func (h *Handler) updateStaffRole(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	var req modlv1.UpdateStaffRoleRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	rID := h.resolveRoleID(r.Context(), srv.ID, req.Role)
	now := httpx.NowMs()
	res, err := h.db.ExecContext(r.Context(), `UPDATE staff SET role_id = ?, updated_at = ? WHERE server_id = ? AND (id = ? OR LOWER(email) = ?)`,
		rID, now, srv.ID, id, strings.ToLower(id))
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to update staff role")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		httpx.WriteError(w, r, http.StatusNotFound, "Staff member not found")
		return
	}
	h.invalidateStaff(srv.ID, id)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.StaffMutationResponse{Message: "Staff role updated"})
}

func (h *Handler) deleteStaff(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	res, err := h.db.ExecContext(r.Context(), `DELETE FROM staff WHERE server_id = ? AND (id = ? OR LOWER(email) = ?)`, srv.ID, id, strings.ToLower(id))
	if err == nil {
		if n, _ := res.RowsAffected(); n > 0 {
			h.invalidateStaff(srv.ID, id)
			httpx.WriteProto(w, r, http.StatusOK, &modlv1.StaffMutationResponse{Message: "Staff member removed"})
			return
		}
	}
	res2, _ := h.db.ExecContext(r.Context(), `DELETE FROM invitations WHERE server_id = ? AND id = ?`, srv.ID, id)
	if n, _ := res2.RowsAffected(); n > 0 {
		h.invalidateStaff(srv.ID, id)
		httpx.WriteProto(w, r, http.StatusOK, &modlv1.StaffMutationResponse{Message: "Invitation revoked"})
		return
	}
	httpx.WriteError(w, r, http.StatusNotFound, "Staff member not found")
}

func (h *Handler) inviteStaff(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.InviteStaffRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	emails := make([]string, 0, len(req.Emails)+1)
	if req.Email != nil && strings.TrimSpace(*req.Email) != "" {
		emails = append(emails, *req.Email)
	}
	emails = append(emails, req.Emails...)

	roleID := h.resolveRoleID(r.Context(), srv.ID, req.Role)
	var succeeded []string
	var failed []*modlv1.InviteResultResponse_FailedInvite
	now := httpx.NowMs()
	exp := now + 7*24*3600*1000

	for _, raw := range emails {
		em := strings.ToLower(strings.TrimSpace(raw))
		if em == "" || !strings.Contains(em, "@") {
			failed = append(failed, &modlv1.InviteResultResponse_FailedInvite{Email: raw, Reason: "Invalid email"})
			continue
		}
		if h.roles.IsAuthorizedEmail(r.Context(), srv, em) {
			failed = append(failed, &modlv1.InviteResultResponse_FailedInvite{Email: em, Reason: "Already a staff member"})
			continue
		}
		token := httpx.RandomToken(24)
		_, err := h.db.ExecContext(r.Context(), `
			INSERT INTO invitations (id, server_id, email, role_id, token, expires_at, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			httpx.NewID(), srv.ID, em, roleID, token, exp, now, now,
		)
		if err != nil {
			failed = append(failed, &modlv1.InviteResultResponse_FailedInvite{Email: em, Reason: "Database error"})
			continue
		}
		inviteURL := fmt.Sprintf("https://%s/accept-invitation?token=%s", srv.CustomDomain, token)
		_ = h.mailer.SendInvitation(em, srv.ServerName, inviteURL)
		succeeded = append(succeeded, em)
	}
	h.invalidateStaff(srv.ID, "")
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.InviteResultResponse{
		Message: "Invitations processed",
		Success: succeeded,
		Failed:  failed,
	})
}

func (h *Handler) resendInvitation(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	var em, token string
	err := h.db.QueryRowContext(r.Context(), `SELECT email, token FROM invitations WHERE server_id = ? AND id = ? LIMIT 1`, srv.ID, id).Scan(&em, &token)
	if err != nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Invitation not found")
		return
	}
	inviteURL := fmt.Sprintf("https://%s/accept-invitation?token=%s", srv.CustomDomain, token)
	_ = h.mailer.SendInvitation(em, srv.ServerName, inviteURL)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.StaffMutationResponse{Message: "Invitation resent"})
}

func (h *Handler) acceptInvitationGet(w http.ResponseWriter, r *http.Request) {
	h.acceptInvitationInternal(w, r, r.URL.Query().Get("token"))
}

func (h *Handler) acceptInvitationPost(w http.ResponseWriter, r *http.Request) {
	tok := r.URL.Query().Get("token")
	if tok == "" {
		var req modlv1.AcceptInvitationRequest
		if err := httpx.ReadProto(r, &req); err == nil {
			tok = req.Token
		}
	}
	h.acceptInvitationInternal(w, r, tok)
}

func (h *Handler) acceptInvitationInternal(w http.ResponseWriter, r *http.Request, token string) {
	if strings.TrimSpace(token) == "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid invitation link.")
		return
	}
	srv := httpx.GetServer(r)
	var id, em, roleID string
	var exp int64
	err := h.db.QueryRowContext(r.Context(), `SELECT id, email, role_id, expires_at FROM invitations WHERE server_id = ? AND token = ? LIMIT 1`, srv.ID, token).Scan(&id, &em, &roleID, &exp)
	if err != nil || exp < httpx.NowMs() {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid or expired invitation.")
		return
	}
	now := httpx.NowMs()
	username := strings.Split(em, "@")[0]
	_, err = h.db.ExecContext(r.Context(), `
		INSERT INTO staff (id, server_id, email, username, role_id, language, date_format, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 'en', 'MM/dd/yyyy', ?, ?)
		ON CONFLICT(server_id, email) DO UPDATE SET role_id=excluded.role_id, updated_at=excluded.updated_at`,
		httpx.NewID(), srv.ID, em, username, roleID, now, now,
	)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to accept invitation")
		return
	}
	_, _ = h.db.ExecContext(r.Context(), `DELETE FROM invitations WHERE id = ?`, id)
	h.invalidateStaff(srv.ID, "")
	roleName := h.roleNameMap(r.Context(), srv.ID)[roleID]
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.AcceptInvitationResponse{
		Message: "Invitation accepted successfully.",
		StaffMember: &modlv1.AcceptInvitationResponse_StaffInviteAcceptedMember{
			Email:    em,
			Username: username,
			Role:     roleName,
		},
	})
}

func (h *Handler) assignMinecraftPlayer(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	emailParam := strings.ToLower(strings.TrimSpace(chi.URLParam(r, "email")))
	var req modlv1.AssignMinecraftPlayerRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	uuidStr := ""
	userStr := ""
	if req.MinecraftUuid != nil {
		uuidStr = strings.TrimSpace(*req.MinecraftUuid)
	}
	if req.MinecraftUsername != nil {
		userStr = strings.TrimSpace(*req.MinecraftUsername)
	}
	res, err := h.db.ExecContext(r.Context(),
		`UPDATE staff SET assigned_minecraft_uuid = ?, assigned_minecraft_username = ?, updated_at = ? WHERE server_id = ? AND (LOWER(email) = ? OR id = ?)`,
		uuidStr, userStr, httpx.NowMs(), srv.ID, emailParam, emailParam)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to assign Minecraft player")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		httpx.WriteError(w, r, http.StatusNotFound, "Staff member not found")
		return
	}
	h.invalidateStaff(srv.ID, "")
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.StaffMutationResponse{Message: "Minecraft player assigned"})
}

func (h *Handler) getAvailablePlayers(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	rows, err := h.db.QueryContext(r.Context(), `
		SELECT minecraft_uuid, current_username FROM players
		WHERE server_id = ? AND current_username != ''
		ORDER BY last_seen DESC LIMIT 200`, srv.ID)
	if err != nil {
		httpx.WriteProto(w, r, http.StatusOK, &modlv1.AvailablePlayersResponse{})
		return
	}
	defer rows.Close()
	var out []*modlv1.AvailablePlayerResponse
	for rows.Next() {
		var u, n string
		if rows.Scan(&u, &n) == nil {
			out = append(out, &modlv1.AvailablePlayerResponse{Uuid: u, Username: n})
		}
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.AvailablePlayersResponse{Players: out})
}

func (h *Handler) buildMinecraftStaffSummary(ctx context.Context, srv *models.Server) []*modlv1.MinecraftStaffSummaryResponse {
	list, _ := h.listStaff(ctx, srv.ID)
	roleNames := h.roleNameMap(ctx, srv.ID)
	out := make([]*modlv1.MinecraftStaffSummaryResponse, 0, len(list))
	for _, s := range list {
		rName := roleNames[s.RoleID]
		if rName == "" {
			rName = s.RoleID
		}
		out = append(out, &modlv1.MinecraftStaffSummaryResponse{
			Id:                s.ID,
			Username:          s.Username,
			Email:             s.Email,
			Role:              rName,
			MinecraftUuid:     s.AssignedMinecraftUUID,
			MinecraftUsername: s.AssignedMinecraftUsername,
			Permissions:       h.roles.GetEffectivePermissions(ctx, srv, s.Email),
			LastSeen:          s.LastSeen,
			CreatedAt:         s.CreatedAt,
			UpdatedAt:         s.UpdatedAt,
		})
	}
	return out
}

func (h *Handler) buildMinecraftStaffPerms(ctx context.Context, srv *models.Server) []*modlv1.MinecraftStaffPermissionsResponse {
	list, _ := h.listStaff(ctx, srv.ID)
	roleNames := h.roleNameMap(ctx, srv.ID)
	out := make([]*modlv1.MinecraftStaffPermissionsResponse, 0, len(list))
	for _, s := range list {
		if s.AssignedMinecraftUUID == "" {
			continue
		}
		rName := roleNames[s.RoleID]
		if rName == "" {
			rName = s.RoleID
		}
		out = append(out, &modlv1.MinecraftStaffPermissionsResponse{
			MinecraftUuid:     s.AssignedMinecraftUUID,
			MinecraftUsername: s.AssignedMinecraftUsername,
			StaffUsername:     s.Username,
			StaffId:           s.ID,
			StaffRole:         rName,
			Permissions:       h.roles.GetEffectivePermissions(ctx, srv, s.Email),
			Email:             s.Email,
		})
	}
	return out
}

func (h *Handler) getMinecraftStaffV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	staffList := h.buildMinecraftStaffSummary(r.Context(), srv)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": 200, "staff": staffList})
}

func (h *Handler) getMinecraftStaffPermsV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	staffList := h.buildMinecraftStaffPerms(r.Context(), srv)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": 200, "data": map[string]any{"staff": staffList}})
}

func (h *Handler) updateMinecraftStaffRoleV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	var body struct {
		Role string `json:"role"`
	}
	if err := httpx.ReadJSON(r, &body); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid JSON")
		return
	}
	rID := h.resolveRoleID(r.Context(), srv.ID, body.Role)
	res, _ := h.db.ExecContext(r.Context(), `UPDATE staff SET role_id = ?, updated_at = ? WHERE server_id = ? AND (id = ? OR assigned_minecraft_uuid = ?)`,
		rID, httpx.NowMs(), srv.ID, id, id)
	if n, _ := res.RowsAffected(); n == 0 {
		httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"status": 404, "message": "Staff member not found"})
		return
	}
	h.invalidateStaff(srv.ID, id)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": 200, "success": true})
}

func (h *Handler) staffDisconnectV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var body struct {
		MinecraftUUID string `json:"minecraftUuid"`
	}
	if err := httpx.ReadJSON(r, &body); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid JSON")
		return
	}
	res, _ := h.db.ExecContext(r.Context(), `UPDATE staff SET last_seen = ?, updated_at = ? WHERE server_id = ? AND assigned_minecraft_uuid = ?`,
		httpx.NowMs(), httpx.NowMs(), srv.ID, body.MinecraftUUID)
	if n, _ := res.RowsAffected(); n == 0 {
		httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"status": 404, "message": "Staff member not found"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": 200, "success": true})
}

func (h *Handler) getMinecraftStaffV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.StaffListResponse{
		Status: 200,
		Staff:  h.buildMinecraftStaffSummary(r.Context(), srv),
	})
}

func (h *Handler) getMinecraftStaffPermsV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.StaffPermissionsListResponse{
		Status: 200,
		Data: &modlv1.StaffPermissionsListResponse_StaffPermissionsData{
			Staff: h.buildMinecraftStaffPerms(r.Context(), srv),
		},
	})
}

func (h *Handler) updateMinecraftStaffRoleV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	var req modlv1.UpdateStaffRoleRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	rID := h.resolveRoleID(r.Context(), srv.ID, req.Role)
	res, _ := h.db.ExecContext(r.Context(), `UPDATE staff SET role_id = ?, updated_at = ? WHERE server_id = ? AND (id = ? OR assigned_minecraft_uuid = ?)`,
		rID, httpx.NowMs(), srv.ID, id, id)
	if n, _ := res.RowsAffected(); n == 0 {
		httpx.WriteProto(w, r, http.StatusNotFound, &modlv1.MinecraftStaffOperationResponse{Status: 404, Success: false})
		return
	}
	h.invalidateStaff(srv.ID, id)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.MinecraftStaffOperationResponse{Status: 200, Success: true})
}

func (h *Handler) staffDisconnectV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.StaffDisconnectRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	res, _ := h.db.ExecContext(r.Context(), `UPDATE staff SET last_seen = ?, updated_at = ? WHERE server_id = ? AND assigned_minecraft_uuid = ?`,
		httpx.NowMs(), httpx.NowMs(), srv.ID, req.MinecraftUuid)
	if n, _ := res.RowsAffected(); n == 0 {
		httpx.WriteProto(w, r, http.StatusNotFound, &modlv1.MinecraftStaffOperationResponse{Status: 404, Success: false})
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.MinecraftStaffOperationResponse{Status: 200, Success: true})
}

func (h *Handler) generate2FATokenV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var body struct {
		MinecraftUUID string `json:"minecraftUuid"`
		IP            string `json:"ip"`
	}
	if err := httpx.ReadJSON(r, &body); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid JSON")
		return
	}
	tok := httpx.RandomToken(24)
	now := httpx.NowMs()
	res, _ := h.db.ExecContext(r.Context(),
		`UPDATE staff SET two_factor_token = ?, two_factor_token_ip = ?, two_factor_token_created_at = ?, two_factor_pending_delivery = 1, updated_at = ? WHERE server_id = ? AND assigned_minecraft_uuid = ?`,
		tok, body.IP, now, now, srv.ID, body.MinecraftUUID)
	if n, _ := res.RowsAffected(); n == 0 {
		httpx.WriteError(w, r, http.StatusNotFound, "Staff member not found")
		return
	}
	verifyURL := fmt.Sprintf("https://%s/staff-2fa/%s", srv.CustomDomain, tok)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"token":     tok,
		"verifyUrl": verifyURL,
	})
}

func (h *Handler) verify2FAToken(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	token := chi.URLParam(r, "token")
	var staffID, mcUUID, tokenIP string
	var createdAt int64
	err := h.db.QueryRowContext(r.Context(),
		`SELECT id, assigned_minecraft_uuid, two_factor_token_ip, COALESCE(two_factor_token_created_at, 0) FROM staff WHERE server_id = ? AND two_factor_token = ? LIMIT 1`,
		srv.ID, token).Scan(&staffID, &mcUUID, &tokenIP, &createdAt)
	if err != nil || httpx.NowMs()-createdAt > 15*60*1000 {
		httpx.WriteError(w, r, http.StatusNotFound, "Invalid or expired 2FA token")
		return
	}
	exp := time.Now().Add(7 * 24 * time.Hour).UnixMilli()
	_, _ = h.db.ExecContext(r.Context(),
		`UPDATE staff SET two_factor_token = '', two_factor_pending_delivery = 0, two_factor_session_ip = ?, two_factor_session_expires_at = ?, updated_at = ? WHERE id = ?`,
		tokenIP, exp, httpx.NowMs(), staffID)

	if h.hub != nil && mcUUID != "" {
		h.hub.PushStaff2FA(srv.ID, []*modlv1.SyncStaff2FaVerification{{MinecraftUuid: mcUUID}})
	}
	_ = json.NewEncoder // keep import clean if needed
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.Staff2FaVerifyResponse{Status: "verified"})
}

func (h *Handler) invalidateStaff(serverID, staffID string) {
	if h.hub != nil {
		h.hub.InvalidatePanel(serverID, modlv1.PanelResource_PANEL_RESOURCE_STAFF, staffID)
	}
}
