package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/arvaliullin/metrics-collection-service/internal/agent"
)

func main() {

	baseURL := flag.String("a", "localhost:8080", "адрес HTTP-сервера")
	reportIntervalSec := flag.Int("r", 10, "частота отправки метрик на сервер")
	pollIntervalSec := flag.Int("p", 2, "частота опроса метрик из пакета")

	flag.Parse()

	if args := flag.Args(); len(args) > 0 {
		fmt.Fprintf(os.Stderr, "неизвестные аргументы: %v\n", args)
		os.Exit(2)
	}

	normalizeBaseURL(baseURL)

	app := agent.New(*baseURL, *pollIntervalSec, *reportIntervalSec)
	app.Run()
}

func normalizeBaseURL(u *string) {
	if strings.HasPrefix(*u, "http://") || strings.HasPrefix(*u, "https://") {
		return
	}

	*u = "http://" + *u
}
