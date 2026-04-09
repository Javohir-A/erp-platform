package main

import (
	"context"
	"log"
	"net"
	"os"

	"erp/platform/auth-service/config"
	grpcsvc "erp/platform/auth-service/grpc"
	"erp/platform/auth-service/storage"
	authv1 "erp/platform/genproto/erp/auth/v1"

	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	store, err := storage.NewStore(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatal("postgres: ", err)
	}
	defer store.Close()

	adminEmail := os.Getenv("BOOTSTRAP_ADMIN_EMAIL")
	adminPass := os.Getenv("BOOTSTRAP_ADMIN_PASSWORD")
	if adminEmail != "" && adminPass != "" {
		if err := grpcsvc.BootstrapAdmin(ctx, store, adminEmail, adminPass); err != nil {
			log.Fatal("bootstrap admin: ", err)
		}
	}

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatal(err)
	}
	srv := grpc.NewServer()
	authv1.RegisterAuthServiceServer(srv, grpcsvc.NewAuthServer(cfg, store))

	log.Println("auth gRPC listening on", cfg.GRPCAddr)
	if err := srv.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
