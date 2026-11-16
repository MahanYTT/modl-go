package realtime

import (
	"context"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	modlv1 "modl-go/gen/modl/v1"
	"modl-go/internal/config"
	"modl-go/internal/httpx"
	"modl-go/internal/models"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const ProtocolVersion = 1

var upgrader = websocket.Upgrader{
	ReadBufferSize:  65536,
	WriteBufferSize: 65536,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type connection struct {
	id         string
	ws         *websocket.Conn
	mu         sync.Mutex
	serverID   string
	clientKind modlv1.ClientKind
	email      string
	topics     map[modlv1.Topic]bool
	closed     bool
}

func (c *connection) sendEnvelope(env *modlv1.RealtimeEnvelope) error {
	if env.ProtocolVersion == 0 {
		env.ProtocolVersion = ProtocolVersion
	}
	if env.Timestamp == nil {
		env.Timestamp = timestamppb.Now()
	}
	if env.EventId == "" {
		env.EventId = httpx.NewID()
	}
	b, err := proto.Marshal(env)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return websocket.ErrCloseSent
	}
	_ = c.ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return c.ws.WriteMessage(websocket.BinaryMessage, b)
}

func (c *connection) close(code int, reason string) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.closed = true
	_ = c.ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, reason), time.Now().Add(2*time.Second))
	_ = c.ws.Close()
	c.mu.Unlock()
}

type Hub struct {
	cfg      *config.Config
	servers  models.ServerResolver
	sessions models.SessionResolver
	perms    models.PermissionChecker

	mu    sync.RWMutex
	conns map[string]*connection
}

func NewHub(cfg *config.Config, servers models.ServerResolver, sessions models.SessionResolver, perms models.PermissionChecker) *Hub {
	return &Hub{
		cfg:      cfg,
		servers:  servers,
		sessions: sessions,
		perms:    perms,
		conns:    make(map[string]*connection),
	}
}

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.RealtimeEnabled {
		http.Error(w, "Realtime disabled", http.StatusServiceUnavailable)
		return
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	ws.SetReadLimit(65536)

	conn := &connection{
		id:     httpx.NewID(),
		ws:     ws,
		topics: make(map[modlv1.Topic]bool),
	}
	defer h.removeConn(conn)

	// Handshake timeout: expect ClientHello as first binary message
	_ = ws.SetReadDeadline(time.Now().Add(10 * time.Second))
	msgType, payload, err := ws.ReadMessage()
	if err != nil || msgType != websocket.BinaryMessage {
		conn.close(websocket.ClosePolicyViolation, "Expected ClientHello")
		return
	}
	var env modlv1.RealtimeEnvelope
	if err := proto.Unmarshal(payload, &env); err != nil {
		conn.close(websocket.CloseUnsupportedData, "Invalid protobuf")
		return
	}
	hello := env.GetClientHello()
	if hello == nil {
		conn.close(websocket.ClosePolicyViolation, "First message must be ClientHello")
		return
	}

	srv, email, authErr := h.authenticate(r, hello)
	if authErr != nil {
		_ = conn.sendEnvelope(&modlv1.RealtimeEnvelope{
			Payload: &modlv1.RealtimeEnvelope_Error{
				Error: &modlv1.Error{
					Code:    modlv1.ErrorCode_ERROR_CODE_UNAUTHORIZED,
					Message: authErr.Error(),
				},
			},
		})
		conn.close(websocket.ClosePolicyViolation, authErr.Error())
		return
	}

	conn.serverID = srv.ID
	conn.clientKind = hello.ClientKind
	conn.email = email

	var accepted []modlv1.Topic
	for _, t := range hello.SupportedTopics {
		if h.canSubscribe(r.Context(), srv, conn, t) {
			conn.topics[t] = true
			accepted = append(accepted, t)
		}
	}

	h.addConn(conn)

	nowStr := time.Now().UTC().Format(time.RFC3339)
	_ = conn.sendEnvelope(&modlv1.RealtimeEnvelope{
		Payload: &modlv1.RealtimeEnvelope_ServerHello{
			ServerHello: &modlv1.ServerHello{
				ProtocolVersion: ProtocolVersion,
				ConnectionId:    conn.id,
				AcceptedTopics:  accepted,
				ServerTime:      &nowStr,
			},
		},
	})

	stopHeartbeat := make(chan struct{})
	defer close(stopHeartbeat)
	go h.heartbeatLoop(conn, stopHeartbeat)

	for {
		_ = ws.SetReadDeadline(time.Now().Add(75 * time.Second))
		mt, data, err := ws.ReadMessage()
		if err != nil {
			return
		}
		if mt != websocket.BinaryMessage {
			continue
		}
		var in modlv1.RealtimeEnvelope
		if err := proto.Unmarshal(data, &in); err != nil {
			continue
		}
		switch p := in.Payload.(type) {
		case *modlv1.RealtimeEnvelope_Heartbeat:
			_ = conn.sendEnvelope(&modlv1.RealtimeEnvelope{
				Payload: &modlv1.RealtimeEnvelope_Heartbeat{
					Heartbeat: &modlv1.Heartbeat{Sequence: p.Heartbeat.Sequence},
				},
			})
		case *modlv1.RealtimeEnvelope_Subscribe:
			for _, t := range p.Subscribe.Topics {
				if h.canSubscribe(r.Context(), srv, conn, t) {
					conn.mu.Lock()
					conn.topics[t] = true
					conn.mu.Unlock()
				}
			}
		case *modlv1.RealtimeEnvelope_Unsubscribe:
			for _, t := range p.Unsubscribe.Topics {
				conn.mu.Lock()
				delete(conn.topics, t)
				conn.mu.Unlock()
			}
		}
	}
}

func (h *Hub) heartbeatLoop(conn *connection, stop <-chan struct{}) {
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()
	var seq uint64
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			seq++
			if err := conn.sendEnvelope(&modlv1.RealtimeEnvelope{
				Payload: &modlv1.RealtimeEnvelope_Heartbeat{
					Heartbeat: &modlv1.Heartbeat{Sequence: seq},
				},
			}); err != nil {
				return
			}
		}
	}
}

func (h *Hub) authenticate(r *http.Request, hello *modlv1.ClientHello) (*models.Server, string, error) {
	if hello.ClientKind == modlv1.ClientKind_CLIENT_KIND_MINECRAFT_PLUGIN {
		apiKey := strings.TrimSpace(r.Header.Get("X-API-Key"))
		if apiKey == "" {
			return nil, "", httpError("Missing X-API-Key")
		}
		srv, err := h.servers.GetByAPIKey(r.Context(), apiKey)
		if err != nil || srv == nil {
			return nil, "", httpError("Invalid API key")
		}
		return srv, "", nil
	}
	if hello.ClientKind == modlv1.ClientKind_CLIENT_KIND_PANEL {
		domain := ""
		if hello.TenantId != nil {
			domain = httpx.NormalizeServerDomain(*hello.TenantId)
		}
		if domain == "" {
			domain = httpx.NormalizeServerDomain(r.Header.Get("X-Server-Domain"))
		}
		if domain == "" {
			domain = httpx.NormalizeServerDomain(r.Host)
		}
		var srv *models.Server
		var err error
		if domain != "" {
			srv, err = h.servers.GetByDomain(r.Context(), domain)
		}
		if (srv == nil || err != nil) && h.cfg.SingleTenant {
			srv, err = h.servers.GetDefault(r.Context())
		}
		if err != nil || srv == nil {
			return nil, "", httpError("Invalid tenant host")
		}
		for _, c := range r.Cookies() {
			if c.Name == h.cfg.SessionCookieName && strings.TrimSpace(c.Value) != "" {
				sess, err := h.sessions.FindAndRefresh(r.Context(), srv.ID, c.Value)
				if err == nil && sess != nil {
					return srv, sess.Email, nil
				}
			}
		}
		return nil, "", httpError("Missing panel session")
	}
	return nil, "", httpError("Unsupported client kind")
}

type httpError string

func (e httpError) Error() string { return string(e) }

func (h *Hub) canSubscribe(ctx context.Context, srv *models.Server, conn *connection, t modlv1.Topic) bool {
	if t == modlv1.Topic_TOPIC_UNSPECIFIED {
		return false
	}
	if conn.clientKind == modlv1.ClientKind_CLIENT_KIND_MINECRAFT_PLUGIN {
		switch t {
		case modlv1.Topic_TOPIC_MINECRAFT_PERMISSIONS,
			modlv1.Topic_TOPIC_MINECRAFT_PUNISHMENT_TYPES,
			modlv1.Topic_TOPIC_MINECRAFT_STAFF_NOTIFICATIONS,
			modlv1.Topic_TOPIC_MINECRAFT_PRESENCE,
			modlv1.Topic_TOPIC_MINECRAFT_PUNISHMENTS,
			modlv1.Topic_TOPIC_MINECRAFT_PLAYER_NOTIFICATIONS,
			modlv1.Topic_TOPIC_MINECRAFT_STAFF_2FA,
			modlv1.Topic_TOPIC_MINECRAFT_MIGRATION_TASKS,
			modlv1.Topic_TOPIC_MINECRAFT_STAT_WIPES:
			return true
		}
		return false
	}
	if conn.clientKind == modlv1.ClientKind_CLIENT_KIND_PANEL {
		if h.perms.IsSuperAdmin(srv, conn.email) {
			return true
		}
		roleID := h.perms.PanelRoleID(ctx, srv, conn.email)
		switch t {
		case modlv1.Topic_TOPIC_PANEL_TICKETS, modlv1.Topic_TOPIC_PANEL_ASSIGNED_TICKETS,
			modlv1.Topic_TOPIC_PANEL_NOTIFICATIONS, modlv1.Topic_TOPIC_PANEL_APPEALS:
			return h.perms.HasPermission(ctx, srv, roleID, "ticket.view.all")
		case modlv1.Topic_TOPIC_PANEL_MIGRATIONS:
			return h.perms.HasPermission(ctx, srv, roleID, "admin.settings.view.migration")
		case modlv1.Topic_TOPIC_PANEL_PLAYERS, modlv1.Topic_TOPIC_PANEL_PUNISHMENTS:
			return h.perms.HasPermission(ctx, srv, roleID, "punishment.modify")
		case modlv1.Topic_TOPIC_PANEL_STAFF:
			return h.perms.HasPermission(ctx, srv, roleID, "admin.staff.manage.members")
		case modlv1.Topic_TOPIC_PANEL_ROLES:
			return h.perms.HasPermission(ctx, srv, roleID, "admin.staff.manage.roles")
		case modlv1.Topic_TOPIC_PANEL_PUNISHMENT_TYPES:
			return h.perms.HasPermission(ctx, srv, roleID, "admin.settings.view.punishments")
		case modlv1.Topic_TOPIC_PANEL_KNOWLEDGEBASE, modlv1.Topic_TOPIC_PANEL_HOMEPAGE:
			return h.perms.HasPermission(ctx, srv, roleID, "admin.settings.view.content")
		case modlv1.Topic_TOPIC_PANEL_SETTINGS:
			return h.perms.HasPermission(ctx, srv, roleID, "admin.settings.view")
		case modlv1.Topic_TOPIC_PANEL_AUDIT:
			return h.perms.HasPermission(ctx, srv, roleID, "admin.audit.view.logs")
		case modlv1.Topic_TOPIC_PANEL_DASHBOARD:
			return h.perms.HasPermission(ctx, srv, roleID, "admin.audit.view.dashboard")
		}
	}
	return false
}

func (h *Hub) addConn(c *connection) {
	h.mu.Lock()
	h.conns[c.id] = c
	h.mu.Unlock()
}

func (h *Hub) removeConn(c *connection) {
	h.mu.Lock()
	delete(h.conns, c.id)
	h.mu.Unlock()
	c.close(websocket.CloseNormalClosure, "")
}

func (h *Hub) Publish(serverID string, topic modlv1.Topic, env *modlv1.RealtimeEnvelope) {
	h.mu.RLock()
	var targets []*connection
	for _, c := range h.conns {
		if c.serverID != serverID {
			continue
		}
		c.mu.Lock()
		sub := c.topics[topic]
		c.mu.Unlock()
		if sub {
			targets = append(targets, c)
		}
	}
	h.mu.RUnlock()

	for _, c := range targets {
		if err := c.sendEnvelope(env); err != nil {
			log.Printf("realtime send error conn=%s: %v", c.id, err)
		}
	}
}

func (h *Hub) InvalidatePanel(serverID string, resource modlv1.PanelResource, resourceID string) {
	topic := panelResourceTopic(resource)
	if topic == modlv1.Topic_TOPIC_UNSPECIFIED {
		return
	}
	ev := &modlv1.PanelInvalidatedEvent{
		Resource: resource,
		TenantId: &serverID,
	}
	if resourceID != "" {
		ev.ResourceId = &resourceID
	}
	h.Publish(serverID, topic, &modlv1.RealtimeEnvelope{
		Payload: &modlv1.RealtimeEnvelope_PanelInvalidated{PanelInvalidated: ev},
	})
}

func (h *Hub) PushPunishments(serverID string, pending []*modlv1.SyncPendingPunishment, modified []*modlv1.SyncModifiedPunishment) {
	h.Publish(serverID, modlv1.Topic_TOPIC_MINECRAFT_PUNISHMENTS, &modlv1.RealtimeEnvelope{
		Payload: &modlv1.RealtimeEnvelope_PunishmentPush{
			PunishmentPush: &modlv1.PunishmentPushEvent{
				Pending:  pending,
				Modified: modified,
			},
		},
	})
}

func (h *Hub) PushPlayerNotifications(serverID string, notifications []*modlv1.SyncPlayerNotification) {
	h.Publish(serverID, modlv1.Topic_TOPIC_MINECRAFT_PLAYER_NOTIFICATIONS, &modlv1.RealtimeEnvelope{
		Payload: &modlv1.RealtimeEnvelope_PlayerNotificationPush{
			PlayerNotificationPush: &modlv1.PlayerNotificationPushEvent{
				Notifications: notifications,
			},
		},
	})
}

func (h *Hub) PushStaffNotifications(serverID string, notifications []*modlv1.SyncStaffNotification) {
	h.Publish(serverID, modlv1.Topic_TOPIC_MINECRAFT_STAFF_NOTIFICATIONS, &modlv1.RealtimeEnvelope{
		Payload: &modlv1.RealtimeEnvelope_StaffNotificationPush{
			StaffNotificationPush: &modlv1.StaffNotificationPushEvent{
				Notifications: notifications,
			},
		},
	})
}

func (h *Hub) PushStaff2FA(serverID string, verifications []*modlv1.SyncStaff2FaVerification) {
	h.Publish(serverID, modlv1.Topic_TOPIC_MINECRAFT_STAFF_2FA, &modlv1.RealtimeEnvelope{
		Payload: &modlv1.RealtimeEnvelope_Staff_2FaPush{
			Staff_2FaPush: &modlv1.Staff2FaPushEvent{
				Verifications: verifications,
			},
		},
	})
}

func (h *Hub) PushMigrationTask(serverID string, task *modlv1.SyncMigrationTask) {
	h.Publish(serverID, modlv1.Topic_TOPIC_MINECRAFT_MIGRATION_TASKS, &modlv1.RealtimeEnvelope{
		Payload: &modlv1.RealtimeEnvelope_MigrationTaskPush{
			MigrationTaskPush: &modlv1.MigrationTaskPushEvent{Task: task},
		},
	})
}

func panelResourceTopic(r modlv1.PanelResource) modlv1.Topic {
	switch r {
	case modlv1.PanelResource_PANEL_RESOURCE_PLAYERS:
		return modlv1.Topic_TOPIC_PANEL_PLAYERS
	case modlv1.PanelResource_PANEL_RESOURCE_PUNISHMENTS:
		return modlv1.Topic_TOPIC_PANEL_PUNISHMENTS
	case modlv1.PanelResource_PANEL_RESOURCE_STAFF:
		return modlv1.Topic_TOPIC_PANEL_STAFF
	case modlv1.PanelResource_PANEL_RESOURCE_ROLES:
		return modlv1.Topic_TOPIC_PANEL_ROLES
	case modlv1.PanelResource_PANEL_RESOURCE_SETTINGS:
		return modlv1.Topic_TOPIC_PANEL_SETTINGS
	case modlv1.PanelResource_PANEL_RESOURCE_PUNISHMENT_TYPES:
		return modlv1.Topic_TOPIC_PANEL_PUNISHMENT_TYPES
	case modlv1.PanelResource_PANEL_RESOURCE_KNOWLEDGEBASE:
		return modlv1.Topic_TOPIC_PANEL_KNOWLEDGEBASE
	case modlv1.PanelResource_PANEL_RESOURCE_AUDIT:
		return modlv1.Topic_TOPIC_PANEL_AUDIT
	case modlv1.PanelResource_PANEL_RESOURCE_HOMEPAGE:
		return modlv1.Topic_TOPIC_PANEL_HOMEPAGE
	case modlv1.PanelResource_PANEL_RESOURCE_APPEALS:
		return modlv1.Topic_TOPIC_PANEL_APPEALS
	case modlv1.PanelResource_PANEL_RESOURCE_TICKETS:
		return modlv1.Topic_TOPIC_PANEL_TICKETS
	case modlv1.PanelResource_PANEL_RESOURCE_DASHBOARD:
		return modlv1.Topic_TOPIC_PANEL_DASHBOARD
	case modlv1.PanelResource_PANEL_RESOURCE_NOTIFICATIONS:
		return modlv1.Topic_TOPIC_PANEL_NOTIFICATIONS
	default:
		return modlv1.Topic_TOPIC_UNSPECIFIED
	}
}
