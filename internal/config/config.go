package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
)

type ServerConfig struct {
	Address         string         `envconfig:"ADDRESS" default:"localhost:8080"`
	StoreInterval   int            `envconfig:"STORE_INTERVAL" default:"300"`
	FileStoragePath string         `envconfig:"FILE_STORAGE_PATH" default:"/tmp/metrics-db.json"`
	Restore         bool           `envconfig:"RESTORE" default:"false"`
	Key             string         `envconfig:"KEY"`
	DatabaseConfig  PostgresConfig `envconfig:"DATABASE"`
}

func LoadConfig() *ServerConfig {
	var cfg ServerConfig

	err := envconfig.Process("", &cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ошибка чтения переменных окружения: %v\n", err)
		os.Exit(1)
	}

	var flagAddress string
	flag.StringVar(&flagAddress, "a", "", "адрес и порт HTTP-сервера")

	var flagStoreInterval int
	flag.IntVar(&flagStoreInterval, "i", -1, "интервал сохранения в секундах")

	var flagFileStoragePath string
	flag.StringVar(&flagFileStoragePath, "f", "", "путь к файлу хранилища")

	var flagRestore bool
	flag.BoolVar(&flagRestore, "r", false, "загружать данные из файла при старте")

	var flagKey string
	flag.StringVar(&flagKey, "k", "", "ключ")

	var flagDsn string
	flag.StringVar(&flagDsn, "d", "", "строка подключения к БД")

	flag.Parse()

	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "a":
			cfg.Address = flagAddress
		case "i":
			cfg.StoreInterval = flagStoreInterval
		case "f":
			cfg.FileStoragePath = flagFileStoragePath
		case "d":
			cfg.DatabaseConfig.Dsn = flagDsn
		case "k":
			cfg.Key = flagKey
		case "r":
			cfg.Restore = flagRestore
		}
	})

	if args := flag.Args(); len(args) > 0 {
		fmt.Fprintf(os.Stderr, "неизвестные аргументы: %v\n", args)
		os.Exit(2)
	}

	return &cfg
}
