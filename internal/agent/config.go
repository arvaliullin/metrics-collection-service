package agent

import (
	"encoding/json"
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

type agentFileConfig struct {
	Address        string `json:"address"`
	ReportInterval string `json:"report_interval"`
	PollInterval   string `json:"poll_interval"`
	CryptoKey      string `json:"crypto_key"`
	Key            string `json:"key"`
	RateLimit      int    `json:"rate_limit"`
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

	configPath := os.Getenv("CONFIG")
	flag.StringVar(&configPath, "c", configPath, "путь к файлу конфигурации")
	flag.StringVar(&configPath, "config", configPath, "путь к файлу конфигурации")

	var flagPollInterval int
	var flagReportInterval int
	var flagAddress string
	var flagKey string
	var flagCryptoKey string
	var flagRateLimit int

	flag.IntVar(&flagPollInterval, "p", 0, "частота опроса метрик")
	flag.IntVar(&flagReportInterval, "r", 0, "частота отправки метрик на сервер")
	flag.StringVar(&flagAddress, "a", "", "адрес и порт HTTP-сервера")
	flag.StringVar(&flagKey, "k", "", "ключ")
	flag.StringVar(&flagCryptoKey, "crypto-key", "", "путь к файлу с публичным ключом")
	flag.IntVar(&flagRateLimit, "l", 0, "максимальное количество одновременных исходящих запросов")

	flag.Parse()

	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ошибка чтения файла конфигурации: %v\n", err)
			os.Exit(1)
		}
		var fileCfg agentFileConfig
		if err := json.Unmarshal(data, &fileCfg); err != nil {
			fmt.Fprintf(os.Stderr, "ошибка разбора файла конфигурации: %v\n", err)
			os.Exit(1)
		}
		if fileCfg.Address != "" {
			cfg.Address = fileCfg.Address
		}
		if fileCfg.ReportInterval != "" {
			d, err := time.ParseDuration(fileCfg.ReportInterval)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ошибка разбора report_interval: %v\n", err)
				os.Exit(1)
			}
			cfg.ReportInterval = int(d.Seconds())
		}
		if fileCfg.PollInterval != "" {
			d, err := time.ParseDuration(fileCfg.PollInterval)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ошибка разбора poll_interval: %v\n", err)
				os.Exit(1)
			}
			cfg.PollInterval = int(d.Seconds())
		}
		if fileCfg.CryptoKey != "" {
			cfg.CryptoKey = fileCfg.CryptoKey
		}
		if fileCfg.Key != "" {
			cfg.Key = fileCfg.Key
		}
		if fileCfg.RateLimit > 0 {
			cfg.RateLimit = fileCfg.RateLimit
		}
	}

	err := envconfig.Process("", &cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ошибка чтения переменных окружения: %v\n", err)
		os.Exit(1)
	}

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
