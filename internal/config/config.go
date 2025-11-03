package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
)

type ServerConfig struct {
	Address         string `envconfig:"ADDRESS" default:"localhost:8080"`
	StoreInterval   int    `envconfig:"STORE_INTERVAL" default:"300"`
	FileStoragePath string `envconfig:"FILE_STORAGE_PATH" default:"/tmp/metrics-db.json"`
	Restore         bool   `envconfig:"RESTORE" default:"false"`
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

	flag.Parse()

	flag.Visit(func(f *flag.Flag) {
		if f.Name == "r" {
			cfg.Restore = flagRestore
		}
	})

	if flagAddress != "" {
		cfg.Address = flagAddress
	}

	if flagStoreInterval >= 0 {
		cfg.StoreInterval = flagStoreInterval
	}

	if flagFileStoragePath != "" {
		cfg.FileStoragePath = flagFileStoragePath
	}

	if args := flag.Args(); len(args) > 0 {
		fmt.Fprintf(os.Stderr, "неизвестные аргументы: %v\n", args)
		os.Exit(2)
	}

	return &cfg
}
