package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// ServerConfig содержит конфигурацию сервера.
type ServerConfig struct {
	Address         string         `envconfig:"ADDRESS" default:"localhost:8080"`
	StoreInterval   int            `envconfig:"STORE_INTERVAL" default:"300"`
	FileStoragePath string         `envconfig:"FILE_STORAGE_PATH" default:"/tmp/metrics-db.json"`
	Restore         bool           `envconfig:"RESTORE" default:"false"`
	Key             string         `envconfig:"KEY"`
	CryptoKey       string         `envconfig:"CRYPTO_KEY"`
	DatabaseConfig  PostgresConfig `envconfig:"DATABASE"`
	AuditFile       string         `envconfig:"AUDIT_FILE"`
	AuditURL        string         `envconfig:"AUDIT_URL"`
}

type serverFileConfig struct {
	Address       string `json:"address"`
	Restore       bool   `json:"restore"`
	StoreInterval string `json:"store_interval"`
	StoreFile     string `json:"store_file"`
	DatabaseDSN   string `json:"database_dsn"`
	CryptoKey     string `json:"crypto_key"`
	Key           string `json:"key"`
	AuditFile     string `json:"audit_file"`
	AuditURL      string `json:"audit_url"`
}

type serverFlags struct {
	configPath      string
	address         string
	storeInterval   int
	fileStoragePath string
	restore         bool
	key             string
	dsn             string
	auditFile       string
	auditURL        string
	cryptoKey       string
}

func defineServerFlags() (*flag.FlagSet, *serverFlags) {
	fs := flag.CommandLine
	f := &serverFlags{}

	f.configPath = os.Getenv("CONFIG")
	fs.StringVar(&f.configPath, "c", f.configPath, "путь к файлу конфигурации")
	fs.StringVar(&f.configPath, "config", f.configPath, "путь к файлу конфигурации")
	fs.StringVar(&f.address, "a", "", "адрес и порт HTTP-сервера")
	fs.IntVar(&f.storeInterval, "i", -1, "интервал сохранения в секундах")
	fs.StringVar(&f.fileStoragePath, "f", "", "путь к файлу хранилища")
	fs.BoolVar(&f.restore, "r", false, "загружать данные из файла при старте")
	fs.StringVar(&f.key, "k", "", "ключ")
	fs.StringVar(&f.dsn, "d", "", "строка подключения к БД")
	fs.StringVar(&f.auditFile, "audit-file", "", "путь к файлу для аудита")
	fs.StringVar(&f.auditURL, "audit-url", "", "URL для отправки аудита")
	fs.StringVar(&f.cryptoKey, "crypto-key", "", "путь к файлу с приватным ключом")

	return fs, f
}

func applyServerFileConfig(cfg *ServerConfig, path string) {
	var fileCfg serverFileConfig
	if err := LoadFileConfig(path, &fileCfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if fileCfg.Address != "" {
		cfg.Address = fileCfg.Address
	}
	cfg.Restore = fileCfg.Restore
	if fileCfg.StoreInterval != "" {
		d, err := time.ParseDuration(fileCfg.StoreInterval)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ошибка разбора store_interval: %v\n", err)
			os.Exit(1)
		}
		cfg.StoreInterval = int(d.Seconds())
	}
	if fileCfg.StoreFile != "" {
		cfg.FileStoragePath = fileCfg.StoreFile
	}
	if fileCfg.DatabaseDSN != "" {
		cfg.DatabaseConfig.Dsn = fileCfg.DatabaseDSN
	}
	if fileCfg.CryptoKey != "" {
		cfg.CryptoKey = fileCfg.CryptoKey
	}
	if fileCfg.Key != "" {
		cfg.Key = fileCfg.Key
	}
	if fileCfg.AuditFile != "" {
		cfg.AuditFile = fileCfg.AuditFile
	}
	if fileCfg.AuditURL != "" {
		cfg.AuditURL = fileCfg.AuditURL
	}
}

func applyServerFlags(cfg *ServerConfig, f *serverFlags, fs *flag.FlagSet) {
	fs.Visit(func(fl *flag.Flag) {
		switch fl.Name {
		case "a":
			cfg.Address = f.address
		case "i":
			cfg.StoreInterval = f.storeInterval
		case "f":
			cfg.FileStoragePath = f.fileStoragePath
		case "d":
			cfg.DatabaseConfig.Dsn = f.dsn
		case "k":
			if cfg.Key == "" {
				cfg.Key = f.key
			}
		case "crypto-key":
			if cfg.CryptoKey == "" {
				cfg.CryptoKey = f.cryptoKey
			}
		case "r":
			cfg.Restore = f.restore
		case "audit-file":
			cfg.AuditFile = f.auditFile
		case "audit-url":
			cfg.AuditURL = f.auditURL
		}
	})
}

// LoadConfig загружает конфигурацию сервера из файла, переменных окружения и флагов командной строки.
func LoadConfig() *ServerConfig {
	var cfg ServerConfig

	fs, f := defineServerFlags()
	fs.Parse(os.Args[1:])

	if f.configPath != "" {
		applyServerFileConfig(&cfg, f.configPath)
	}

	if err := envconfig.Process("", &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "ошибка чтения переменных окружения: %v\n", err)
		os.Exit(1)
	}

	applyServerFlags(&cfg, f, fs)

	if args := fs.Args(); len(args) > 0 {
		fmt.Fprintf(os.Stderr, "неизвестные аргументы: %v\n", args)
		os.Exit(2)
	}

	return &cfg
}
