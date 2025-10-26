package main

import (
	"net/http"
	"os"

	"github.com/arvaliullin/metrics-collection-service/internal/config"
	"github.com/arvaliullin/metrics-collection-service/internal/handler/get"
	"github.com/arvaliullin/metrics-collection-service/internal/handler/html"
	"github.com/arvaliullin/metrics-collection-service/internal/handler/update"
	"github.com/arvaliullin/metrics-collection-service/internal/repository"
	"github.com/go-chi/chi/v5"
)

func main() {

	cfg := config.LoadConfig()
	storage := repository.NewEmptyMemStorage()
	updateHandler := update.NewUpdateHandler(storage)
	getHandler := get.NewGetHandler(storage)
	htmlHandler := html.NewHTMLHandler(storage)

	router := chi.NewRouter()
	router.Handle(`POST /update/{type}/{id}/{value}`, updateHandler)
	router.Handle(`GET /value/{type}/{id}`, getHandler)
	router.Handle(`GET /`, htmlHandler)
	if err := http.ListenAndServe(cfg.Address, router); err != nil {
		os.Exit(2)
	}
}
