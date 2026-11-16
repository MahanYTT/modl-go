package role

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/httpx"
	"modl-go/internal/models"
	"modl-go/internal/realtime"

	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const SuperAdminRoleID = "super-admin"

var basePermissions = []models.Permission{
	{ID: "admin.settings.view", Name: "View Settings", Description: "View all system settings (includes all sub-permissions)", Category: "admin"},
	{ID: "admin.settings.view.punishments", Name: "View Punishments Config", Description: "View punishment type configuration", Category: "admin", Parent: "admin.settings.view"},
	{ID: "admin.settings.view.content", Name: "View Content", Description: "View homepage cards, knowledgebase, media", Category: "admin", Parent: "admin.settings.view"},
	{ID: "admin.settings.view.domain", Name: "View Domain", Description: "View custom domain configuration", Category: "admin", Parent: "admin.settings.view"},
	{ID: "admin.settings.view.migration", Name: "View Migration", Description: "View import/export data configuration", Category: "admin", Parent: "admin.settings.view"},
	{ID: "admin.settings.view.storage", Name: "View Storage", Description: "View storage configuration", Category: "admin", Parent: "admin.settings.view"},
	{ID: "admin.settings.modify", Name: "Modify Settings", Description: "Full control over system settings (includes all sub-permissions)", Category: "admin"},
	{ID: "admin.settings.modify.punishments", Name: "Modify Punishments Config", Description: "Create/edit/delete punishment types", Category: "admin", Parent: "admin.settings.modify"},
	{ID: "admin.settings.modify.content", Name: "Modify Content", Description: "Edit homepage cards, knowledgebase, media", Category: "admin", Parent: "admin.settings.modify"},
	{ID: "admin.settings.modify.domain", Name: "Modify Domain", Description: "Change custom domain configuration", Category: "admin", Parent: "admin.settings.modify"},
	{ID: "admin.settings.modify.migration", Name: "Modify Migration", Description: "Import/export data between platforms", Category: "admin", Parent: "admin.settings.modify"},
	{ID: "admin.settings.modify.storage", Name: "Modify Storage", Description: "Configure storage backends and limits", Category: "admin", Parent: "admin.settings.modify"},
	{ID: "admin.staff.manage", Name: "Manage Staff", Description: "Full staff management (includes all sub-permissions)", Category: "admin"},
	{ID: "admin.staff.manage.members", Name: "Manage Members", Description: "Invite, remove, and reassign staff", Category: "admin", Parent: "admin.staff.manage"},
	{ID: "admin.staff.manage.roles", Name: "Manage Roles", Description: "Create/edit/delete roles and permissions", Category: "admin", Parent: "admin.staff.manage"},
	{ID: "admin.audit.view", Name: "View Audit", Description: "Full audit access (includes all sub-permissions)", Category: "admin"},
	{ID: "admin.audit.view.dashboard", Name: "View Dashboard", Description: "View dashboard statistics", Category: "admin", Parent: "admin.audit.view"},
	{ID: "admin.audit.view.analytics", Name: "View Analytics", Description: "View player and ticket analytics", Category: "admin", Parent: "admin.audit.view"},
	{ID: "admin.audit.view.logs", Name: "View Logs", Description: "View audit trail of staff actions", Category: "admin", Parent: "admin.audit.view"},
	{ID: "admin.audit.rollback", Name: "Rollback Audit Actions", Description: "Roll back punishments and perform destructive bulk audit operations", Category: "admin", SuperAdminOnly: true},
	{ID: "punishment.view", Name: "View Punishments", Description: "View player profiles, punishments, and linked accounts", Category: "punishment"},
	{ID: "punishment.modify", Name: "Modify Punishments", Description: "Full control over existing punishments (includes all sub-permissions)", Category: "punishment"},
	{ID: "punishment.modify.pardon", Name: "Pardon Punishments", Description: "Pardon punishments and clear associated points", Category: "punishment", Parent: "punishment.modify"},
	{ID: "punishment.modify.duration", Name: "Modify Duration", Description: "Change punishment duration", Category: "punishment", Parent: "punishment.modify"},
	{ID: "punishment.modify.note", Name: "Add Notes", Description: "Add staff notes to punishments", Category: "punishment", Parent: "punishment.modify"},
	{ID: "punishment.modify.evidence", Name: "Manage Evidence", Description: "Add and view evidence on punishments", Category: "punishment", Parent: "punishment.modify"},
	{ID: "punishment.modify.options", Name: "Toggle Options", Description: "Toggle alt-blocking and stat-wipe options", Category: "punishment", Parent: "punishment.modify"},
	{ID: "staff.chat.toggle", Name: "Toggle Chat", Description: "Toggle server chat on/off", Category: "staff"},
	{ID: "staff.chat.clear", Name: "Clear Chat", Description: "Clear server chat", Category: "staff"},
	{ID: "staff.chat.slow", Name: "Slow Chat", Description: "Set slow mode on server chat", Category: "staff"},
	{ID: "staff.maintenance", Name: "Maintenance Mode", Description: "Toggle server maintenance mode", Category: "staff"},
	{ID: "staff.modactions", Name: "Moderation Actions", Description: "Staff mode, vanish, freeze, and target players", Category: "staff"},
	{ID: "staff.intercept", Name: "Intercept Chat", Description: "Intercept and view all network chat", Category: "staff"},
	{ID: "staff.chatlogs", Name: "Chat Logs", Description: "View player chat history", Category: "staff"},
	{ID: "staff.commandlogs", Name: "Command Logs", Description: "View player command history", Category: "staff"},
	{ID: "ticket.view.all", Name: "View All Tickets", Description: "View all tickets (includes all sub-permissions)", Category: "ticket"},
	{ID: "ticket.view.all.notes", Name: "View Staff Notes", Description: "View internal staff notes on tickets", Category: "ticket", Parent: "ticket.view.all"},
	{ID: "ticket.reply.all", Name: "Reply to All Tickets", Description: "Reply to all ticket types (includes all sub-permissions)", Category: "ticket"},
	{ID: "ticket.reply.all.notes", Name: "Add Staff Notes", Description: "Add staff-only internal notes", Category: "ticket", Parent: "ticket.reply.all"},
	{ID: "appeal.modify", Name: "Modify Appeals", Description: "Reply to and update appeals", Category: "ticket"},
	{ID: "ticket.close.all", Name: "Close/Reopen All Tickets", Description: "Close and reopen all ticket types (includes all sub-permissions)", Category: "ticket"},
	{ID: "ticket.close.all.lock", Name: "Lock Tickets", Description: "Lock tickets to prevent further replies", Category: "ticket", Parent: "ticket.close.all"},
	{ID: "ticket.manage", Name: "Manage Tickets", Description: "Advanced ticket management (includes all sub-permissions)", Category: "ticket"},
	{ID: "ticket.manage.tags", Name: "Manage Tags", Description: "Add and remove tags from tickets", Category: "ticket", Parent: "ticket.manage"},
	{ID: "ticket.manage.hide", Name: "Hide Tickets", Description: "Hide tickets from public view", Category: "ticket", Parent: "ticket.manage"},
	{ID: "ticket.manage.subscribe", Name: "Manage Subscriptions", Description: "Manage ticket notification subscriptions", Category: "ticket", Parent: "ticket.manage"},
	{ID: "ticket.delete.all", Name: "Delete Tickets", Description: "Delete tickets from the system", Category: "ticket"},
}

var permissionCategories = map[string]string{
	"punishment": "Punishment Permissions",
	"ticket":     "Ticket Permissions",
	"admin":      "Administrative Permissions",
	"staff":      "Staff Tool Permissions",
}

type Handler struct {
	db  *sql.DB
	hub *realtime.Hub
}

func NewHandler(db *sql.DB, hub *realtime.Hub) *Handler {
	return &Handler{db: db, hub: hub}
}

func (h *Handler) SetHub(hub *realtime.Hub) {
	h.hub = hub
}

func (h *Handler) RegisterPanelRoutes(r chi.Router) {
	r.Route("/roles", func(r chi.Router) {
		r.Use(httpx.RequirePanelPolicy(h, models.AccessPolicy{
			Rule:   models.RuleRequirePermission,
			View:   "admin.staff.manage.roles",
			Modify: "admin.staff.manage.roles",
		}))
		r.Get("/", h.getAllPanelRoles)
		r.Get("/permissions", h.getPermissions)
		r.Get("/{id}", h.getPanelRoleByID)
		r.Post("/", h.createRole)
		r.Put("/{id}", h.updateRole)
		r.Delete("/{id}", h.deleteRole)
		r.Post("/reorder", h.reorderRoles)
	})
}

func (h *Handler) RegisterMinecraftV1Routes(r chi.Router) {
	r.Route("/roles", func(r chi.Router) {
		r.Get("/", h.getAllMinecraftRolesV1)
		r.Get("/{id}", h.getMinecraftRoleV1)
		r.Patch("/{id}/permissions", h.updateRolePermissionsV1)
	})
}

func (h *Handler) RegisterMinecraftV3Routes(r chi.Router) {
	r.Route("/roles", func(r chi.Router) {
		r.Get("/", h.getAllMinecraftRolesV3)
		r.Get("/{id}", h.getMinecraftRoleV3)
		r.Patch("/{id}/permissions", h.updateRolePermissionsV3)
	})
}

// --- PermissionChecker implementation ---

func (h *Handler) IsSuperAdmin(srv *models.Server, email string) bool {
	if srv == nil || strings.TrimSpace(email) == "" {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(srv.AdminEmail), strings.TrimSpace(email))
}

func (h *Handler) IsAuthorizedEmail(ctx context.Context, srv *models.Server, email string) bool {
	if h.IsSuperAdmin(srv, email) {
		return true
	}
	var c int
	err := h.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM staff WHERE server_id = ? AND LOWER(email) = ?`, srv.ID, strings.ToLower(strings.TrimSpace(email))).Scan(&c)
	return err == nil && c > 0
}

func (h *Handler) PanelRoleID(ctx context.Context, srv *models.Server, email string) string {
	if srv == nil || strings.TrimSpace(email) == "" {
		return ""
	}
	var roleID string
	_ = h.db.QueryRowContext(ctx, `SELECT role_id FROM staff WHERE server_id = ? AND LOWER(email) = ? LIMIT 1`, srv.ID, strings.ToLower(strings.TrimSpace(email))).Scan(&roleID)
	return roleID
}

func (h *Handler) HasPermission(ctx context.Context, srv *models.Server, roleID, perm string) bool {
	if roleID == "" {
		return false
	}
	role, err := h.findRoleByID(ctx, srv.ID, roleID)
	if err != nil || role == nil {
		return false
	}
	if isSuperAdminOnlyPerm(perm) && role.ID != SuperAdminRoleID {
		return false
	}
	return roleGrants(role, perm)
}

func (h *Handler) HasAnyWithPrefix(ctx context.Context, srv *models.Server, roleID, prefix string) bool {
	if roleID == "" {
		return false
	}
	role, err := h.findRoleByID(ctx, srv.ID, roleID)
	if err != nil || role == nil {
		return false
	}
	if role.ID == SuperAdminRoleID {
		return true
	}
	for _, p := range role.Permissions {
		if strings.HasPrefix(p, prefix) {
			return true
		}
	}
	return false
}

func (h *Handler) GetEffectivePermissions(ctx context.Context, srv *models.Server, email string) []string {
	allPerms := h.allPermissionIDs(ctx, srv.ID)
	if h.IsSuperAdmin(srv, email) {
		return allPerms
	}
	roleID := h.PanelRoleID(ctx, srv, email)
	if roleID == "" {
		return []string{}
	}
	role, err := h.findRoleByID(ctx, srv.ID, roleID)
	if err != nil || role == nil {
		return []string{}
	}
	var out []string
	for _, p := range allPerms {
		if roleGrants(role, p) && (!isSuperAdminOnlyPerm(p) || role.ID == SuperAdminRoleID) {
			out = append(out, p)
		}
	}
	return out
}

func roleGrants(role *models.StaffRole, perm string) bool {
	if role.ID == SuperAdminRoleID {
		return true
	}
	for _, p := range role.Permissions {
		if p == perm || p == "*" {
			return true
		}
		// Parent permission grants sub-permissions (e.g. "admin.settings.view" grants "admin.settings.view.*")
		if strings.HasPrefix(perm, p+".") {
			return true
		}
	}
	return false
}

func isSuperAdminOnlyPerm(id string) bool {
	for _, p := range basePermissions {
		if p.ID == id {
			return p.SuperAdminOnly
		}
	}
	return false
}

func (h *Handler) allPermissionIDs(ctx context.Context, serverID string) []string {
	out := make([]string, 0, len(basePermissions))
	for _, p := range basePermissions {
		out = append(out, p.ID)
	}
	for _, p := range h.punishmentPermissions(ctx, serverID) {
		out = append(out, p.ID)
	}
	return out
}

func (h *Handler) punishmentPermissions(ctx context.Context, serverID string) []models.Permission {
	var dataJSON string
	err := h.db.QueryRowContext(ctx, `SELECT data_json FROM settings WHERE server_id = ? AND type = 'punishment_types' LIMIT 1`, serverID).Scan(&dataJSON)
	if err != nil || dataJSON == "" {
		return nil
	}
	var raw struct {
		Types []struct {
			Name string `json:"name"`
		} `json:"types"`
	}
	if json.Unmarshal([]byte(dataJSON), &raw) != nil {
		return nil
	}
	out := make([]models.Permission, 0, len(raw.Types))
	for _, t := range raw.Types {
		if t.Name == "" {
			continue
		}
		id := "punishment.apply." + strings.ReplaceAll(strings.ToLower(t.Name), " ", "-")
		out = append(out, models.Permission{
			ID:          id,
			Name:        "Apply " + t.Name,
			Description: "Permission to apply " + t.Name + " punishments",
			Category:    "punishment",
		})
	}
	return out
}

func (h *Handler) EnsureDefaultRoles(ctx context.Context, serverID string) ([]models.StaffRole, error) {
	roles, err := h.listRoles(ctx, serverID)
	if err == nil && len(roles) > 0 {
		return roles, nil
	}
	now := httpx.NowMs()
	defaults := []models.StaffRole{
		{
			ID:          httpx.NewID(),
			ServerID:    serverID,
			Name:        "Administrator",
			Description: "Full access to moderation, tickets, and settings",
			Permissions: []string{"admin.settings.view", "admin.settings.modify", "admin.staff.manage", "admin.audit.view", "punishment.view", "punishment.modify", "ticket.view.all", "ticket.reply.all", "ticket.close.all", "ticket.manage", "appeal.modify", "staff.modactions", "staff.chatlogs", "staff.commandlogs"},
			IsDefault:   true,
			Order:       0,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          httpx.NewID(),
			ServerID:    serverID,
			Name:        "Moderator",
			Description: "Handle player punishments, reports, and tickets",
			Permissions: []string{"punishment.view", "punishment.modify", "ticket.view.all", "ticket.reply.all", "ticket.close.all", "appeal.modify", "staff.modactions", "staff.chatlogs", "staff.commandlogs"},
			IsDefault:   true,
			Order:       1,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          httpx.NewID(),
			ServerID:    serverID,
			Name:        "Helper",
			Description: "View player info and respond to support tickets",
			Permissions: []string{"punishment.view", "ticket.view.all", "ticket.reply.all"},
			IsDefault:   true,
			Order:       2,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
	for _, rl := range defaults {
		_ = h.upsertRole(ctx, &rl)
	}
	return defaults, nil
}

func (h *Handler) listRoles(ctx context.Context, serverID string) ([]models.StaffRole, error) {
	rows, err := h.db.QueryContext(ctx,
		`SELECT id, server_id, name, description, permissions_json, is_default, ord, created_at, updated_at FROM staff_roles WHERE server_id = ? ORDER BY ord ASC, created_at ASC`,
		serverID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.StaffRole
	for rows.Next() {
		var r models.StaffRole
		var permsJSON string
		var isDef int
		if err := rows.Scan(&r.ID, &r.ServerID, &r.Name, &r.Description, &permsJSON, &isDef, &r.Order, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(permsJSON), &r.Permissions)
		if r.Permissions == nil {
			r.Permissions = []string{}
		}
		r.IsDefault = isDef != 0
		out = append(out, r)
	}
	return out, rows.Err()
}

func (h *Handler) findRoleByID(ctx context.Context, serverID, roleID string) (*models.StaffRole, error) {
	row := h.db.QueryRowContext(ctx,
		`SELECT id, server_id, name, description, permissions_json, is_default, ord, created_at, updated_at FROM staff_roles WHERE server_id = ? AND id = ? LIMIT 1`,
		serverID, roleID,
	)
	var r models.StaffRole
	var permsJSON string
	var isDef int
	if err := row.Scan(&r.ID, &r.ServerID, &r.Name, &r.Description, &permsJSON, &isDef, &r.Order, &r.CreatedAt, &r.UpdatedAt); err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(permsJSON), &r.Permissions)
	if r.Permissions == nil {
		r.Permissions = []string{}
	}
	r.IsDefault = isDef != 0
	return &r, nil
}

func (h *Handler) upsertRole(ctx context.Context, r *models.StaffRole) error {
	b, _ := json.Marshal(r.Permissions)
	isDef := 0
	if r.IsDefault {
		isDef = 1
	}
	_, err := h.db.ExecContext(ctx, `
		INSERT INTO staff_roles (id, server_id, name, description, permissions_json, is_default, ord, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name=excluded.name, description=excluded.description, permissions_json=excluded.permissions_json, is_default=excluded.is_default, ord=excluded.ord, updated_at=excluded.updated_at`,
		r.ID, r.ServerID, r.Name, r.Description, string(b), isDef, r.Order, r.CreatedAt, r.UpdatedAt,
	)
	return err
}

func (h *Handler) getAllPanelRoles(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	roles, err := h.EnsureDefaultRoles(r.Context(), srv.ID)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to load roles")
		return
	}
	pb := make([]*modlv1.RoleResponse, 0, len(roles))
	for _, rl := range roles {
		pb = append(pb, toProtoRoleResponse(&rl))
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PanelRoleListResponse{Roles: pb})
}

func (h *Handler) getPermissions(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var pbPerms []*modlv1.PermissionsResponse_Permission
	for _, p := range basePermissions {
		if p.SuperAdminOnly {
			continue
		}
		item := &modlv1.PermissionsResponse_Permission{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Category:    p.Category,
		}
		if p.Parent != "" {
			par := p.Parent
			item.ParentId = &par
		}
		pbPerms = append(pbPerms, item)
	}
	for _, p := range h.punishmentPermissions(r.Context(), srv.ID) {
		pbPerms = append(pbPerms, &modlv1.PermissionsResponse_Permission{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Category:    p.Category,
		})
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.PermissionsResponse{
		Permissions: pbPerms,
		Categories:  permissionCategories,
	})
}

func (h *Handler) getPanelRoleByID(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	rl, err := h.findRoleByID(r.Context(), srv.ID, id)
	if err != nil || rl == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Role not found")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.RoleDetailResponse{Role: toProtoRoleResponse(rl)})
}

func (h *Handler) createRole(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.RoleRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	roles, _ := h.listRoles(r.Context(), srv.ID)
	now := httpx.NowMs()
	rl := &models.StaffRole{
		ID:          httpx.NewID(),
		ServerID:    srv.ID,
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Permissions: filterGrantable(req.Permissions),
		Order:       len(roles),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := h.upsertRole(r.Context(), rl); err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to create role")
		return
	}
	h.invalidateRoles(srv.ID, rl.ID)
	httpx.WriteProto(w, r, http.StatusCreated, &modlv1.RoleMutationResponse{
		Message: "Role created successfully",
		Role:    toProtoRoleResponse(rl),
	})
}

func (h *Handler) updateRole(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	rl, err := h.findRoleByID(r.Context(), srv.ID, id)
	if err != nil || rl == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Role not found")
		return
	}
	var req modlv1.RoleRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	rl.Name = strings.TrimSpace(req.Name)
	rl.Description = strings.TrimSpace(req.Description)
	rl.Permissions = filterGrantable(req.Permissions)
	rl.UpdatedAt = httpx.NowMs()
	if err := h.upsertRole(r.Context(), rl); err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to update role")
		return
	}
	h.invalidateRoles(srv.ID, rl.ID)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.RoleMutationResponse{
		Message: "Role updated successfully",
		Role:    toProtoRoleResponse(rl),
	})
}

func (h *Handler) deleteRole(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	res, err := h.db.ExecContext(r.Context(), `DELETE FROM staff_roles WHERE server_id = ? AND id = ?`, srv.ID, id)
	if err != nil {
		httpx.WriteError(w, r, http.StatusInternalServerError, "Failed to delete role")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		httpx.WriteError(w, r, http.StatusNotFound, "Role not found")
		return
	}
	h.invalidateRoles(srv.ID, id)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.RoleMutationResponse{Message: "Role deleted successfully"})
}

func (h *Handler) reorderRoles(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	var req modlv1.ReorderRolesRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	now := httpx.NowMs()
	for _, item := range req.RoleOrder {
		_, _ = h.db.ExecContext(r.Context(), `UPDATE staff_roles SET ord = ?, updated_at = ? WHERE server_id = ? AND id = ?`, item.Order, now, srv.ID, item.Id)
	}
	h.invalidateRoles(srv.ID, "")
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.RoleMutationResponse{Message: "Roles reordered successfully"})
}

func (h *Handler) getAllMinecraftRolesV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	roles, _ := h.EnsureDefaultRoles(r.Context(), srv.ID)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": 200, "roles": roles})
}

func (h *Handler) getMinecraftRoleV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	rl, err := h.findRoleByID(r.Context(), srv.ID, chi.URLParam(r, "id"))
	if err != nil || rl == nil {
		httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"status": 404, "message": "Role not found"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": 200, "role": rl})
}

func (h *Handler) updateRolePermissionsV1(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	var body struct {
		Permissions []string `json:"permissions"`
	}
	if err := httpx.ReadJSON(r, &body); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "Invalid JSON")
		return
	}
	rl, err := h.findRoleByID(r.Context(), srv.ID, id)
	if err != nil || rl == nil {
		httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"status": 404, "message": "Role not found"})
		return
	}
	rl.Permissions = filterGrantable(body.Permissions)
	rl.UpdatedAt = httpx.NowMs()
	_ = h.upsertRole(r.Context(), rl)
	h.invalidateRoles(srv.ID, rl.ID)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": 200, "message": "Permissions updated successfully", "role": rl})
}

func (h *Handler) getAllMinecraftRolesV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	roles, _ := h.EnsureDefaultRoles(r.Context(), srv.ID)
	pb := make([]*modlv1.MinecraftRole, 0, len(roles))
	for _, rl := range roles {
		pb = append(pb, toProtoMinecraftRole(&rl))
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.MinecraftRoleListResponse{Status: 200, Roles: pb})
}

func (h *Handler) getMinecraftRoleV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	rl, err := h.findRoleByID(r.Context(), srv.ID, chi.URLParam(r, "id"))
	if err != nil || rl == nil {
		httpx.WriteError(w, r, http.StatusNotFound, "Role not found")
		return
	}
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.MinecraftRoleDetailResponse{Status: 200, Role: toProtoMinecraftRole(rl)})
}

func (h *Handler) updateRolePermissionsV3(w http.ResponseWriter, r *http.Request) {
	srv := httpx.GetServer(r)
	id := chi.URLParam(r, "id")
	var req modlv1.UpdateRolePermissionsRequest
	if err := httpx.ReadProto(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	rl, err := h.findRoleByID(r.Context(), srv.ID, id)
	if err != nil || rl == nil {
		httpx.WriteProto(w, r, http.StatusNotFound, &modlv1.MinecraftRoleMutationResponse{Status: 404, Success: false, Message: "Role not found"})
		return
	}
	rl.Permissions = filterGrantable(req.Permissions)
	rl.UpdatedAt = httpx.NowMs()
	_ = h.upsertRole(r.Context(), rl)
	h.invalidateRoles(srv.ID, rl.ID)
	httpx.WriteProto(w, r, http.StatusOK, &modlv1.MinecraftRoleMutationResponse{
		Status:  200,
		Success: true,
		Message: "Role permissions updated",
	})
}

func (h *Handler) invalidateRoles(serverID, roleID string) {
	if h.hub != nil {
		h.hub.InvalidatePanel(serverID, modlv1.PanelResource_PANEL_RESOURCE_ROLES, roleID)
	}
}

func filterGrantable(perms []string) []string {
	out := make([]string, 0, len(perms))
	seen := make(map[string]bool)
	for _, p := range perms {
		p = strings.TrimSpace(p)
		if p == "" || isSuperAdminOnlyPerm(p) || seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	return out
}

func toProtoRoleResponse(r *models.StaffRole) *modlv1.RoleResponse {
	return &modlv1.RoleResponse{
		Id:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		Permissions: r.Permissions,
		IsDefault:   r.IsDefault,
		Order:       int32(r.Order),
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func toProtoMinecraftRole(r *models.StaffRole) *modlv1.MinecraftRole {
	return &modlv1.MinecraftRole{
		Id:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		Permissions: r.Permissions,
		IsDefault:   r.IsDefault,
		Order:       int32(r.Order),
		CreatedAt:   timestamppb.New(time.UnixMilli(r.CreatedAt)),
		UpdatedAt:   timestamppb.New(time.UnixMilli(r.UpdatedAt)),
	}
}
