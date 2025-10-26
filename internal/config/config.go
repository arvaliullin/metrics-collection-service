package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
)

type ServerConfig struct {
	Address string `envconfig:"ADDRESS"`
}

func LoadConfig() *ServerConfig {

	var cfg ServerConfig

	flag.StringVar(&cfg.Address, "a", "localhost:8080", "адрес и порт HTTP-сервера")
	flag.Parse()

	if err := envconfig.Process("", &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "ошибка чтения переменных окружения: %v\n", err)
		os.Exit(1)
	}

	if args := flag.Args(); len(args) > 0 {
		fmt.Fprintf(os.Stderr, "неизвестные аргументы: %v\n", args)
		os.Exit(2)
	}

	return &cfg
}
