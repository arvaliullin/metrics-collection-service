package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/arvaliullin/metrics-collection-service/internal/handler/get"
	"github.com/arvaliullin/metrics-collection-service/internal/handler/html"
	"github.com/arvaliullin/metrics-collection-service/internal/handler/update"
	"github.com/arvaliullin/metrics-collection-service/internal/repository"
	"github.com/go-chi/chi/v5"
)

func main() {

	endpoint := flag.String("a", "localhost:8080", "http server endpoint")
	flag.Parse()

	if args := flag.Args(); len(args) > 0 {
		fmt.Fprintf(os.Stderr, "неизвестные аргументы: %v\n", args)
		os.Exit(2)
	}

	storage := repository.NewEmptyMemStorage()
	updateHandler := update.NewUpdateHandler(storage)
	getHandler := get.NewGetHandler(storage)
	htmlHandler := html.NewHTMLHandler(storage)

	router := chi.NewRouter()
	router.Handle(`POST /update/{type}/{id}/{value}`, updateHandler)
	router.Handle(`GET /value/{type}/{id}`, getHandler)
	router.Handle(`GET /`, htmlHandler)
	if err := http.ListenAndServe(*endpoint, router); err != nil {
		os.Exit(2)
	}
}
