package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cast"
)

type Config struct {
	GRPCAddr   string
	PostgresDSN string
	JWTSecret  string
}

func Load() Config {
	_ = godotenv.Load(".env")
	c := Config{}
	c.GRPCAddr = cast.ToString(env("GRPC_ADDR", ":50051"))
	c.PostgresDSN = cast.ToString(env("POSTGRES_DSN", "postgres://erp:erp@localhost:5432/auth?sslmode=disable"))
	c.JWTSecret = cast.ToString(env("JWT_SECRET", "dev-secret-change-in-production"))
	if c.JWTSecret == "" {
		log.Fatal("JWT_SECRET required")
	}
	return c
}

func env(k string, d any) any {
	if v, ok := os.LookupEnv(k); ok {
		return v
	}
	return d
}
