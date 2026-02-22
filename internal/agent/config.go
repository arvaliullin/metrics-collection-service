package agent

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/arvaliullin/metrics-collection-service/internal/config"
	"github.com/arvaliullin/metrics-collection-service/internal/utils"
	"github.com/kelseyhightower/envconfig"
)

// Config содержит конфигурацию агента.
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

type agentFlags struct {
	configPath     string
	pollInterval   int
	reportInterval int
	address        string
	key            string
	cryptoKey      string
	rateLimit      int
}

// GetPollInterval возвращает интервал опроса метрик.
func (c *Config) GetPollInterval() time.Duration {
	return time.Duration(c.PollInterval) * time.Second
}

// GetReportInterval возвращает интервал отправки метрик.
func (c *Config) GetReportInterval() time.Duration {
	return time.Duration(c.ReportInterval) * time.Second
}

// GetAddress возвращает адрес сервера.
func (c *Config) GetAddress() string {
	return c.Address
}

func defineAgentFlags() (*flag.FlagSet, *agentFlags) {
	fs := flag.CommandLine
	f := &agentFlags{}

	f.configPath = os.Getenv("CONFIG")
	fs.StringVar(&f.configPath, "c", f.configPath, "путь к файлу конфигурации")
	fs.StringVar(&f.configPath, "config", f.configPath, "путь к файлу конфигурации")
	fs.IntVar(&f.pollInterval, "p", 0, "частота опроса метрик")
	fs.IntVar(&f.reportInterval, "r", 0, "частота отправки метрик на сервер")
	fs.StringVar(&f.address, "a", "", "адрес и порт HTTP-сервера")
	fs.StringVar(&f.key, "k", "", "ключ")
	fs.StringVar(&f.cryptoKey, "crypto-key", "", "путь к файлу с публичным ключом")
	fs.IntVar(&f.rateLimit, "l", 0, "максимальное количество одновременных исходящих запросов")

	return fs, f
}

func applyAgentFileConfig(cfg *Config, path string) {
	var fileCfg agentFileConfig
	if err := config.LoadFileConfig(path, &fileCfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
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

func applyAgentFlags(cfg *Config, f *agentFlags, fs *flag.FlagSet) {
	fs.Visit(func(fl *flag.Flag) {
		switch fl.Name {
		case "p":
			cfg.PollInterval = f.pollInterval
		case "r":
			cfg.ReportInterval = f.reportInterval
		case "a":
			cfg.Address = f.address
		case "k":
			if cfg.Key == "" {
				cfg.Key = f.key
			}
		case "crypto-key":
			if cfg.CryptoKey == "" {
				cfg.CryptoKey = f.cryptoKey
			}
		case "l":
			cfg.RateLimit = f.rateLimit
		}
	})
}

func loadConfig() *Config {
	var cfg Config

	fs, f := defineAgentFlags()
	fs.Parse(os.Args[1:])

	if f.configPath != "" {
		applyAgentFileConfig(&cfg, f.configPath)
	}

	if err := envconfig.Process("", &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "ошибка чтения переменных окружения: %v\n", err)
		os.Exit(1)
	}

	applyAgentFlags(&cfg, f, fs)

	if args := fs.Args(); len(args) > 0 {
		fmt.Fprintf(os.Stderr, "неизвестные аргументы: %v\n", args)
		os.Exit(2)
	}

	utils.NormalizeBaseURL(&cfg.Address)

	return &cfg
}
