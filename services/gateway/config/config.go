package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cast"
)

type Config struct {
	HTTPPort    string
	AuthAddr    string
	HRAddr      string
	ProcAddr    string
	WarehouseAddr string
	FinanceAddr string
}

func Load() Config {
	_ = godotenv.Load(".env")
	return Config{
		HTTPPort:      cast.ToString(env("HTTP_PORT", ":8080")),
		AuthAddr:      cast.ToString(env("AUTH_GRPC_ADDR", "localhost:50051")),
		HRAddr:        cast.ToString(env("HR_GRPC_ADDR", "localhost:50052")),
		ProcAddr:      cast.ToString(env("PROCUREMENT_GRPC_ADDR", "localhost:50053")),
		WarehouseAddr: cast.ToString(env("WAREHOUSE_GRPC_ADDR", "localhost:50054")),
		FinanceAddr:   cast.ToString(env("FINANCE_GRPC_ADDR", "localhost:50055")),
	}
}

func env(k string, d any) any {
	if v, ok := os.LookupEnv(k); ok {
		return v
	}
	return d
}
