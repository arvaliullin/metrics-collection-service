package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
)

type ServerConfig struct {
	Address string `envconfig:"ADDRESS" default:"localhost:8080"`
}

func LoadConfig() *ServerConfig {
	var cfg ServerConfig

	err := envconfig.Process("", &cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ошибка чтения переменных окружения: %v\n", err)
		os.Exit(1)
	}

	var flagAddress string
	flag.StringVar(&flagAddress, "a", cfg.Address, "адрес и порт HTTP-сервера")
	flag.Parse()

	cfg.Address = flagAddress

	if args := flag.Args(); len(args) > 0 {
		fmt.Fprintf(os.Stderr, "неизвестные аргументы: %v\n", args)
		os.Exit(2)
	}

	return &cfg
}
