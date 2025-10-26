package server

import (
	"context"
	"net/http"
	"os"

	"github.com/arvaliullin/metrics-collection-service/internal/config"
	"github.com/arvaliullin/metrics-collection-service/internal/handler/get"
	"github.com/arvaliullin/metrics-collection-service/internal/handler/html"
	"github.com/arvaliullin/metrics-collection-service/internal/handler/update"
	"github.com/arvaliullin/metrics-collection-service/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

// handlers содержит все HTTP обработчики приложения
type handlers struct {
	update *update.UpdateHandler
	get    *get.GetHandler
	html   *html.HTMLHandler
}

// ServerApp представляет основное серверное приложение со всеми зависимостями
type ServerApp struct {
	Cfg      *config.ServerConfig
	handlers *handlers
	server   *http.Server
	storage  repository.MemStorage
	logger   zerolog.Logger
}

// New создает новый экземпляр ServerApp с инициализированными зависимостями
func New(ctx context.Context) *ServerApp {
	cfg := config.LoadConfig()
	storage := repository.NewEmptyMemStorage()

	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger().
		Level(zerolog.InfoLevel)

	app := &ServerApp{
		Cfg:     cfg,
		storage: storage,
		logger:  logger,
		handlers: &handlers{
			update: update.NewUpdateHandler(storage),
			get:    get.NewGetHandler(storage),
			html:   html.NewHTMLHandler(storage),
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
	router.Use(loggingMiddleware(a.logger))

	router.Handle(`POST /update/{type}/{id}/{value}`, a.handlers.update)
	router.Handle(`GET /value/{type}/{id}`, a.handlers.get)
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
	return nil
}
