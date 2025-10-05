package main

import (
	"net/http"

	"github.com/arvaliullin/metrics-collection-service/internal/handler/counter"
	"github.com/arvaliullin/metrics-collection-service/internal/handler/gauge"
	"github.com/arvaliullin/metrics-collection-service/internal/repository"
)

func main() {

	storage := repository.NewEmptyMemStorage()
	counterHandler := counter.NewCounterHandler(storage)
	gaugeHandler := gauge.NewGaugeHandler(storage)

	mux := http.NewServeMux()

	mux.Handle(`POST /update/counter/{id}/{value}`, counterHandler)
	mux.Handle(`POST /update/gauge/{id}/{value}`, gaugeHandler)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
