package main

import (
	"net/http"

	"github.com/arvaliullin/metrics-collection-service/internal/handler/get"
	"github.com/arvaliullin/metrics-collection-service/internal/handler/update"
	"github.com/arvaliullin/metrics-collection-service/internal/repository"
	"github.com/go-chi/chi/v5"
)

func main() {

	storage := repository.NewEmptyMemStorage()
	updateHandler := update.NewUpdateHandler(storage)
	getHandler := get.NewGetHandler(storage)

	router := chi.NewRouter()
	router.Handle(`POST /update/{type}/{id}/{value}`, updateHandler)
	router.Handle(`GET /value/{type}/{id}`, getHandler)
	if err := http.ListenAndServe(":8080", router); err != nil {
		panic(err)
	}
}
