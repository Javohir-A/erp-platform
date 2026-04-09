package config

import (
	"github.com/joho/godotenv"
	"github.com/spf13/cast"
	"os"
)

type Config struct {
	GRPCAddr    string
	PostgresDSN string
}

func Load() Config {
	_ = godotenv.Load(".env")
	return Config{
		GRPCAddr:    cast.ToString(env("GRPC_ADDR", ":50052")),
		PostgresDSN: cast.ToString(env("POSTGRES_DSN", "postgres://erp:erp@localhost:5432/hr?sslmode=disable")),
	}
}

func env(k string, d any) any {
	if v, ok := os.LookupEnv(k); ok {
		return v
	}
	return d
}
