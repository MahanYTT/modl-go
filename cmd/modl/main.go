package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"modl-go/internal/ai"
	"modl-go/internal/config"
	"modl-go/internal/db"
	"modl-go/internal/email"
	"modl-go/internal/httpx"
	"modl-go/internal/modules/analytics"
	"modl-go/internal/modules/appeal"
	"modl-go/internal/modules/audit"
	"modl-go/internal/modules/auth"
	"modl-go/internal/modules/dashboard"
	"modl-go/internal/modules/health"
	"modl-go/internal/modules/homepage"
	"modl-go/internal/modules/knowledgebase"
	modllog "modl-go/internal/modules/log"
	"modl-go/internal/modules/media"
	"modl-go/internal/modules/migration"
	"modl-go/internal/modules/player"
	"modl-go/internal/modules/role"
	"modl-go/internal/modules/server"
	"modl-go/internal/modules/settings"
	"modl-go/internal/modules/staff"
	"modl-go/internal/modules/stubs"
	"modl-go/internal/modules/ticket"
	"modl-go/internal/realtime"
	"modl-go/internal/storage"
	"modl-go/internal/webui"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func buildServerHandler(cfg *config.Config) (http.Handler, func(), error) {
	database, err := db.Open(cfg.DBPath)
	if err != nil {
		return nil, nil, err
	}

	store, err := storage.NewLocalStore(database, cfg.StorageDir)
	if err != nil {
		_ = database.Close()
		return nil, nil, err
	}

	mailer := email.NewSender(cfg)
	aiClient := ai.NewClient(cfg)

	roleH := role.NewHandler(database, nil)
	serverH := server.NewHandler(database, cfg, mailer, roleH)
	authH := auth.NewHandler(database, cfg, mailer, roleH)
	serverH.SetSessionCreator(authH)

	hub := realtime.NewHub(cfg, serverH, authH, roleH)
	roleH.SetHub(hub)

	staffH := staff.NewHandler(database, cfg, mailer, hub, roleH)
	settingsH := settings.NewHandler(database, cfg, hub, roleH, store)
	playerH := player.NewHandler(database, cfg, hub, roleH)
	ticketH := ticket.NewHandler(database, cfg, mailer, hub, roleH, aiClient)
	appealH := appeal.NewHandler(database, cfg, mailer, hub, roleH, ticketH)

	kbSvc := knowledgebase.NewService(database)
	kbH := knowledgebase.NewHandler(kbSvc, hub)
	hpSvc := homepage.NewService(database, kbSvc)
	hpH := homepage.NewHandler(hpSvc, hub)
	mediaSvc := media.NewService(database, store, cfg)
	mediaH := media.NewHandler(mediaSvc)

	dashH := dashboard.NewHandler(database, roleH)
	analyticsH := analytics.NewHandler(database, roleH)
	auditH := audit.NewHandler(database, hub, roleH)
	logH := modllog.NewHandler(database, roleH)
	migH := migration.NewHandler(database, cfg, hub, roleH)
	healthH := health.NewHandler()
	stubsH := stubs.NewHandler()

	ctx := context.Background()
	if bootSrv, err := serverH.BootstrapIfConfigured(ctx); err != nil {
		log.Printf("bootstrap warning: %v", err)
	} else if bootSrv != nil {
		_, _ = roleH.EnsureDefaultRoles(ctx, bootSrv.ID)
	}

	apiRouter := buildAPIRouter(
		cfg, serverH, authH, roleH, staffH, settingsH,
		playerH, ticketH, appealH, kbH, hpH, mediaH,
		dashH, analyticsH, auditH, logH, migH, healthH, stubsH, hub,
	)

	spa := webui.Handler(cfg.PanelDistDir)

	root := chi.NewRouter()
	root.Use(middleware.Recoverer)
	root.Use(httpx.CORSMiddleware(cfg))

	// Serve API at both /v1|/v2|/v3 (Minecraft plugin & direct API) and /api/v1|/api/v2|/api/v3 (panel SPA default)
	root.Mount("/api", http.StripPrefix("/api", apiRouter))
	root.HandleFunc("/v1", apiRouter.ServeHTTP)
	root.HandleFunc("/v1/*", apiRouter.ServeHTTP)
	root.HandleFunc("/v2/*", apiRouter.ServeHTTP)
	root.HandleFunc("/v3/*", apiRouter.ServeHTTP)
	root.HandleFunc("/actuator/*", apiRouter.ServeHTTP)

	// Embedded panel SPA fallback for everything else
	root.NotFound(spa.ServeHTTP)

	cleanup := func() {
		_ = database.Close()
	}
	return root, cleanup, nil
}

func main() {
	cfg := config.Load()

	root, cleanup, err := buildServerHandler(cfg)
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}
	defer cleanup()

	listenAddr := cfg.Port
	if !strings.Contains(listenAddr, ":") {
		listenAddr = ":" + listenAddr
	}
	httpSrv := &http.Server{
		Addr:              listenAddr,
		Handler:           root,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("modl-go listening on %s (sqlite=%s, storage=%s, single_tenant=%v)",
			listenAddr, cfg.DBPath, cfg.StorageDir, cfg.SingleTenant)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server error: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutCtx)
	log.Println("modl-go shutdown complete")
}

func buildAPIRouter(
	cfg *config.Config,
	serverH *server.Handler,
	authH *auth.Handler,
	roleH *role.Handler,
	staffH *staff.Handler,
	settingsH *settings.Handler,
	playerH *player.Handler,
	ticketH *ticket.Handler,
	appealH *appeal.Handler,
	kbH *knowledgebase.Handler,
	hpH *homepage.Handler,
	mediaH *media.Handler,
	dashH *dashboard.Handler,
	analyticsH *analytics.Handler,
	auditH *audit.Handler,
	logH *modllog.Handler,
	migH *migration.Handler,
	healthH *health.Handler,
	stubsH *stubs.Handler,
	hub *realtime.Hub,
) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(httpx.OriginCSRFMiddleware(cfg))

	// Health & Realtime WebSocket
	healthH.RegisterRoutes(r)
	r.Get("/v1/realtime/ws", hub.ServeWS)

	// Panel-facing routes (/v1/panel/*)
	r.Route("/v1/panel", func(r chi.Router) {
		r.Use(httpx.ServerHeaderMiddleware(cfg, serverH))
		r.Use(httpx.SessionAuthMiddleware(cfg, authH))

		r.Route("/auth", authH.RegisterPanelAuthRoutes)

		serverH.RegisterPanelRoutes(r)
		roleH.RegisterPanelRoutes(r)
		staffH.RegisterPanelRoutes(r)
		settingsH.RegisterPanelRoutes(r)
		playerH.RegisterPanelRoutes(r)
		ticketH.RegisterPanelRoutes(r)
		appealH.RegisterPanelRoutes(r)
		r.Route("/knowledgebase", func(r chi.Router) { kbH.RegisterPanel(r, roleH) })
		r.Route("/homepage-cards", func(r chi.Router) { hpH.RegisterPanel(r, roleH) })
		r.Route("/media", func(r chi.Router) { mediaH.RegisterPanelMedia(r, roleH) })
		r.Route("/storage", func(r chi.Router) { mediaH.RegisterPanelStorage(r, roleH) })
		dashH.RegisterPanelRoutes(r)
		analyticsH.RegisterPanelRoutes(r)
		auditH.RegisterPanelRoutes(r)
		logH.RegisterPanelRoutes(r)
		migH.RegisterPanelRoutes(r)
		stubsH.RegisterPanelRoutes(r)
	})

	// Public-facing routes (/v1/public/*)
	r.Route("/v1/public", func(r chi.Router) {
		r.Use(httpx.ServerHeaderMiddleware(cfg, serverH))
		r.Use(httpx.SessionAuthMiddleware(cfg, authH))

		serverH.RegisterPublicRoutes(r)
		staffH.RegisterPublicRoutes(r)
		settingsH.RegisterPublicRoutes(r)
		playerH.RegisterPublicRoutes(r)
		ticketH.RegisterPublicRoutes(r)
		appealH.RegisterPublicRoutes(r)
		r.Route("/knowledgebase", kbH.RegisterPublic)
		r.Route("/homepage-cards", hpH.RegisterPublic)
		r.Route("/media", mediaH.RegisterPublicMedia)
		r.Route("/evidence-upload", mediaH.RegisterEvidenceUpload)
	})

	// Compatibility alias for /v1/media/* -> public media
	r.Route("/v1/media", func(r chi.Router) {
		r.Use(httpx.ServerHeaderMiddleware(cfg, serverH))
		mediaH.RegisterPublicMedia(r)
	})

	// Minecraft V1 routes (/v1/minecraft/*)
	r.Route("/v1/minecraft", func(r chi.Router) {
		r.Use(httpx.ApiKeyMiddleware(serverH))

		roleH.RegisterMinecraftV1Routes(r)
		staffH.RegisterMinecraftV1Routes(r)
		settingsH.RegisterMinecraftV1Routes(r)
		playerH.RegisterMinecraftV1Routes(r)
		ticketH.RegisterMinecraftV1Routes(r)
		dashH.RegisterMinecraftV1Routes(r)
		migH.RegisterMinecraftV1Routes(r)
	})

	// Minecraft V2 routes (/v2/minecraft/*)
	r.Route("/v2/minecraft", func(r chi.Router) {
		r.Use(httpx.ApiKeyMiddleware(serverH))
		playerH.RegisterMinecraftV2Routes(r)
	})

	// Minecraft V3 Protobuf routes (/v3/minecraft/*)
	r.Route("/v3/minecraft", func(r chi.Router) {
		r.Use(httpx.ApiKeyMiddleware(serverH))

		roleH.RegisterMinecraftV3Routes(r)
		staffH.RegisterMinecraftV3Routes(r)
		settingsH.RegisterMinecraftV3Routes(r)
		playerH.RegisterMinecraftV3Routes(r)
		ticketH.RegisterMinecraftV3Routes(r)
		dashH.RegisterMinecraftV3Routes(r)
	})

	_ = strings.TrimSpace
	return r
}
