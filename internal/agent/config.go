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
	Key            string `envconfig:"KEY"`
	CryptoKey      string `envconfig:"CRYPTO_KEY"`
	RateLimit      int    `envconfig:"RATE_LIMIT" default:"3"`
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
	var flagKey string
	var flagCryptoKey string
	var flagRateLimit int

	flag.IntVar(&flagPollInterval, "p", cfg.PollInterval, "частота опроса метрик")
	flag.IntVar(&flagReportInterval, "r", cfg.ReportInterval, "частота отправки метрик на сервер")
	flag.StringVar(&flagAddress, "a", cfg.Address, "адрес и порт HTTP-сервера")
	flag.StringVar(&flagKey, "k", "", "ключ")
	flag.StringVar(&flagCryptoKey, "crypto-key", "", "путь к файлу с публичным ключом")
	flag.IntVar(&flagRateLimit, "l", cfg.RateLimit, "максимальное количество одновременных исходящих запросов")

	flag.Parse()

	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "p":
			cfg.PollInterval = flagPollInterval
		case "r":
			cfg.ReportInterval = flagReportInterval
		case "a":
			cfg.Address = flagAddress
		case "k":
			if cfg.Key == "" {
				cfg.Key = flagKey
			}
		case "crypto-key":
			if cfg.CryptoKey == "" {
				cfg.CryptoKey = flagCryptoKey
			}
		case "l":
			cfg.RateLimit = flagRateLimit
		}
	})

	if args := flag.Args(); len(args) > 0 {
		fmt.Fprintf(os.Stderr, "неизвестные аргументы: %v\n", args)
		os.Exit(2)
	}

	utils.NormalizeBaseURL(&cfg.Address)

	return &cfg
}
