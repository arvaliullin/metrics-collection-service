package server

import (
	"context"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"

	pb "github.com/arvaliullin/metrics-collection-service/api/pb/gen"
	"github.com/arvaliullin/metrics-collection-service/internal/config"
	grpcserver "github.com/arvaliullin/metrics-collection-service/internal/grpc"
	"github.com/arvaliullin/metrics-collection-service/internal/handler/get"
	"github.com/arvaliullin/metrics-collection-service/internal/handler/html"
	"github.com/arvaliullin/metrics-collection-service/internal/handler/ping"
	"github.com/arvaliullin/metrics-collection-service/internal/handler/update"
	"github.com/arvaliullin/metrics-collection-service/internal/handler/updates"
	"github.com/arvaliullin/metrics-collection-service/internal/http/middleware"
	"github.com/arvaliullin/metrics-collection-service/internal/repository"
	"github.com/arvaliullin/metrics-collection-service/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/grpc"

	_ "github.com/arvaliullin/metrics-collection-service/docs"
)

// handlers содержит все HTTP обработчики приложения
type handlers struct {
	update     *update.UpdateHandler
	updateJSON *update.UpdateJSONHandler
	updates    *updates.UpdatesHandler
	get        *get.GetHandler
	getJSON    *get.GetJSONHandler
	html       *html.HTMLHandler
	ping       *ping.PingHandler
}

// ServerApp представляет основное серверное приложение со всеми зависимостями
type ServerApp struct {
	Cfg          *config.ServerConfig
	handlers     *handlers
	server       *http.Server
	grpcServer   *grpc.Server
	storage      repository.MetricStorage
	logger       zerolog.Logger
	auditService *service.AuditService
	trustedNet   *net.IPNet
}

// New создает новый экземпляр ServerApp с инициализированными зависимостями
func New(ctx context.Context) *ServerApp {
	cfg := config.LoadConfig()

	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger().
		Level(zerolog.InfoLevel)

	logger.Info().
		Str("address", cfg.Address).
		Str("grpc_address", cfg.GRPCAddress).
		Int("store_interval", cfg.StoreInterval).
		Str("file_storage_path", cfg.FileStoragePath).
		Bool("restore", cfg.Restore).
		Str("db_dsn", cfg.DatabaseConfig.Dsn).
		Str("key", cfg.Key).
		Str("crypto_key", cfg.CryptoKey).
		Str("trusted_subnet", cfg.TrustedSubnet).
		Msg("server configuration loaded")

	var trustedNet *net.IPNet
	if cfg.TrustedSubnet != "" {
		_, parsedNet, parseErr := net.ParseCIDR(cfg.TrustedSubnet)
		if parseErr != nil {
			logger.Fatal().Err(parseErr).Str("trusted_subnet", cfg.TrustedSubnet).Msg("invalid trusted subnet CIDR")
		}
		trustedNet = parsedNet
	}

	storage, err := NewStorage(ctx, cfg, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize storage")
	}

	notifier := service.NewAuditNotifier(logger)
	auditService := service.NewAuditService(notifier, logger)
	auditService.InitializeReceivers(cfg)

	metricsService := service.NewServerMetricsService(storage, logger)

	var gs *grpc.Server
	if cfg.GRPCAddress != "" {
		gs = grpc.NewServer(
			grpc.UnaryInterceptor(grpcserver.TrustedSubnetInterceptor(trustedNet, logger)),
		)
		pb.RegisterMetricsServer(gs, grpcserver.NewMetricsServer(metricsService, logger))
	}

	app := &ServerApp{
		Cfg:          cfg,
		storage:      storage,
		logger:       logger,
		auditService: auditService,
		grpcServer:   gs,
		trustedNet:   trustedNet,
		handlers: &handlers{
			update:     update.NewUpdateHandler(metricsService, auditService),
			updateJSON: update.NewUpdateJSONHandler(metricsService, auditService),
			updates:    updates.NewUpdatesHandler(metricsService, auditService),
			get:        get.NewGetHandler(metricsService),
			getJSON:    get.NewGetJSONHandler(metricsService),
			html:       html.NewHTMLHandler(metricsService),
			ping:       ping.NewPingHandler(storage),
		},
	}

	app.setupRouter()

	return app
}

// Logger возвращает логгер приложения
func (a *ServerApp) Logger() *zerolog.Logger {
	return &a.logger
}

// setupRouter настраивает HTTP маршруты
func (a *ServerApp) setupRouter() {
	router := chi.NewRouter()
	router.Use(middleware.TrustedSubnetMiddleware(a.trustedNet, a.logger))
	router.Use(middleware.HashValidationMiddleware(a.Cfg.Key, a.logger))
	router.Use(middleware.DecryptMiddleware(a.Cfg.CryptoKey, a.logger))
	router.Use(middleware.GzipDecompressMiddleware())
	router.Use(middleware.HashResponseMiddleware(a.Cfg.Key, a.logger))
	router.Use(middleware.GzipCompressMiddleware())
	router.Use(middleware.LoggingMiddleware(a.logger))

	router.Handle(`POST /update/{type}/{id}/{value}`, a.handlers.update)
	router.Handle(`GET /value/{type}/{id}`, a.handlers.get)

	router.Handle(`POST /update`, a.handlers.updateJSON)
	router.Handle(`POST /update/`, a.handlers.updateJSON)
	router.Handle(`POST /updates`, a.handlers.updates)
	router.Handle(`POST /updates/`, a.handlers.updates)
	router.Handle(`POST /value`, a.handlers.getJSON)
	router.Handle(`POST /value/`, a.handlers.getJSON)
	router.Handle(`GET /ping`, a.handlers.ping)
	router.Handle(`GET /`, a.handlers.html)
	router.Get("/swagger/*", httpSwagger.WrapHandler)
	router.Mount("/debug/pprof/", http.DefaultServeMux)

	a.server = &http.Server{
		Addr:    a.Cfg.Address,
		Handler: router,
	}
}

// Run запускает HTTP и (при наличии конфигурации) gRPC серверы.
func (a *ServerApp) Run(ctx context.Context) error {
	go func() {
		a.logger.Info().
			Str("address", a.server.Addr).
			Msg("starting HTTP server")
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error().
				Err(err).
				Msg("HTTP server error")
		}
	}()

	if a.grpcServer != nil {
		lis, err := net.Listen("tcp", a.Cfg.GRPCAddress)
		if err != nil {
			return err
		}
		go func() {
			a.logger.Info().
				Str("address", a.Cfg.GRPCAddress).
				Msg("starting gRPC server")
			if err := a.grpcServer.Serve(lis); err != nil {
				a.logger.Error().Err(err).Msg("gRPC server error")
			}
		}()
	}

	<-ctx.Done()
	a.logger.Info().Msg("shutting down servers")

	if a.grpcServer != nil {
		a.grpcServer.GracefulStop()
	}

	if err := a.server.Shutdown(context.TODO()); err != nil {
		a.logger.Error().Err(err).Msg("error shutting down HTTP server")
	}

	if err := a.storage.Close(); err != nil {
		return err
	}

	return nil
}
