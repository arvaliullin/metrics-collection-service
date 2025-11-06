package config

type PostgresConfig struct {
	Dsn string `envconfig:"DSN" default:"postgres://postgres_user:postgres_password@postgres:5432/postgres_db?sslmode=disable"`
}
