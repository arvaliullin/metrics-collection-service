package agent

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/arvaliullin/metrics-collection-service/internal/utils"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	PollInterval   int    `envconfig:"POLL_INTERVAL" default:"2"`
	ReportInterval int    `envconfig:"REPORT_INTERVAL" default:"10"`
	Address        string `envconfig:"ADDRESS" default:"localhost:8080"`
}

func (c *Config) GetPollInterval() time.Duration {
	return time.Duration(c.PollInterval) * time.Second
}

func (c *Config) GetReportInterval() time.Duration {
	return time.Duration(c.ReportInterval) * time.Second
}

func (c *Config) GetAddress() string {
	return c.Address
}

func loadConfig() *Config {
	var cfg Config

	err := envconfig.Process("", &cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ошибка чтения переменных окружения: %v\n", err)
		os.Exit(1)
	}

	var flagPollInterval int
	var flagReportInterval int
	var flagAddress string

	flag.IntVar(&flagPollInterval, "p", cfg.PollInterval, "частота опроса метрик")
	flag.IntVar(&flagReportInterval, "r", cfg.ReportInterval, "частота отправки метрик на сервер")
	flag.StringVar(&flagAddress, "a", cfg.Address, "адрес и порт HTTP-сервера")
	flag.Parse()

	cfg.PollInterval = flagPollInterval
	cfg.ReportInterval = flagReportInterval
	cfg.Address = flagAddress

	if args := flag.Args(); len(args) > 0 {
		fmt.Fprintf(os.Stderr, "неизвестные аргументы: %v\n", args)
		os.Exit(2)
	}

	utils.NormalizeBaseURL(&cfg.Address)

	return &cfg
}
