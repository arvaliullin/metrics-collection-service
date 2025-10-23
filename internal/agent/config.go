package agent

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	pollInterval   time.Duration `envconfig:"POLL_INTERVAL"`
	reportInterval time.Duration `envconfig:"REPORT_INTERVAL"`
	address        string        `envconfig:"ADDRESS"`
}

func normalizeBaseURL(u *string) {
	if strings.HasPrefix(*u, "http://") || strings.HasPrefix(*u, "https://") {
		return
	}

	*u = "http://" + *u
}

func loadConfig() *Config {
	var cfg Config

	flag.StringVar(&cfg.address, "a", "localhost:8080", "адрес и порт HTTP-сервера")
	flag.DurationVar(&cfg.reportInterval, "r", 10*time.Second, "частота отправки метрик на сервер")
	flag.DurationVar(&cfg.pollInterval, "p", 2*time.Second, "частота опроса метрик")

	flag.Parse()

	err := envconfig.Process("", &cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ошибка чтения переменных окружения: %v\n", err)
		os.Exit(1)
	}

	if args := flag.Args(); len(args) > 0 {
		fmt.Fprintf(os.Stderr, "неизвестные аргументы: %v\n", args)
		os.Exit(2)
	}

	normalizeBaseURL(&cfg.address)

	return &cfg
}
