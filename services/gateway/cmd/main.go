package main

import (
	"context"
	"log"

	"erp/platform/gateway/api"
	"erp/platform/gateway/config"
	"erp/platform/gateway/services"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	cl, err := services.NewClients(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer cl.Close()

	r := api.SetupRouter(cl)
	log.Println("API gateway HTTP on", cfg.HTTPPort)
	if err := r.Run(cfg.HTTPPort); err != nil {
		log.Fatal(err)
	}
}
