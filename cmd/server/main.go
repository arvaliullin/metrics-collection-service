package main

import (
	"net/http"

	"github.com/arvaliullin/metrics-collection-service/internal/handler/update"
	"github.com/arvaliullin/metrics-collection-service/internal/repository"
)

func main() {

	storage := repository.NewEmptyMemStorage()
	updateHandler := update.NewUpdateHandler(storage)
	mux := http.NewServeMux()
	mux.Handle(`POST /update/{type}/{id}/{value}`, updateHandler)
	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
