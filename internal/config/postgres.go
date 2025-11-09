package config

type PostgresConfig struct {
	Dsn string `envconfig:"DSN"`
}
