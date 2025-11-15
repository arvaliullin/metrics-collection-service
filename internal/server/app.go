package server

import (
	"context"
	"net/http"
	"os"

	"github.com/arvaliullin/metrics-collection-service/internal/config"
	"github.com/arvaliullin/metrics-collection-service/internal/handler/get"
	"github.com/arvaliullin/metrics-collection-service/internal/handler/html"
	"github.com/arvaliullin/metrics-collection-service/internal/handler/ping"
	"github.com/arvaliullin/metrics-collection-service/internal/handler/update"
	"github.com/arvaliullin/metrics-collection-service/internal/handler/updates"
	"github.com/arvaliullin/metrics-collection-service/internal/repository/file"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
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
	Cfg      *config.ServerConfig
	handlers *handlers
	server   *http.Server
	storage  MetricStorage
	logger   zerolog.Logger
}

// New создает новый экземпляр ServerApp с инициализированными зависимостями
func New(ctx context.Context) *ServerApp {
	cfg := config.LoadConfig()

	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger().
		Level(zerolog.InfoLevel)

	storage, err := createStorage(ctx, cfg, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize storage")
	}

	app := &ServerApp{
		Cfg:     cfg,
		storage: storage,
		logger:  logger,
		handlers: &handlers{
			update:     update.NewUpdateHandler(storage),
			updateJSON: update.NewUpdateJSONHandler(storage),
			updates:    updates.NewUpdatesHandler(storage),
			get:        get.NewGetHandler(storage),
			getJSON:    get.NewGetJSONHandler(storage),
			html:       html.NewHTMLHandler(storage),
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
	router.Use(GzipDecompressMiddleware())
	router.Use(GzipCompressMiddleware())
	router.Use(loggingMiddleware(a.logger))

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

	a.server = &http.Server{
		Addr:    a.Cfg.Address,
		Handler: router,
	}
}

// Run запускает сервер
func (a *ServerApp) Run(ctx context.Context) error {
	go func() {
		a.logger.Info().
			Str("address", a.server.Addr).
			Msg("starting server")
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error().
				Err(err).
				Msg("server error")
		}
	}()

	<-ctx.Done()
	a.logger.Info().Msg("shutting down server")

	if err := a.server.Shutdown(context.TODO()); err != nil {
		a.logger.Error().Err(err).Msg("error shutting down server")
	}

	if fileStorage, ok := a.storage.(*file.Repository); ok {
		if err := fileStorage.Close(); err != nil {
			return err
		}
	}

	return nil
}
