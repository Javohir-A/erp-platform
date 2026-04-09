package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cast"
)

type Config struct {
	GRPCAddr    string
	PostgresDSN string
}

func Load() Config {
	_ = godotenv.Load(".env")
	return Config{
		GRPCAddr:    cast.ToString(env("GRPC_ADDR", ":50054")),
		PostgresDSN: cast.ToString(env("POSTGRES_DSN", "postgres://erp:erp@localhost:5432/warehouse?sslmode=disable")),
	}
}

func env(k string, d any) any {
	if v, ok := os.LookupEnv(k); ok {
		return v
	}
	return d
}
