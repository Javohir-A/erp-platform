package main

import (
	"context"
	"log"
	"net"

	procv1 "erp/platform/genproto/erp/procurement/v1"
	"erp/platform/procurement-service/config"
	grpcsvc "erp/platform/procurement-service/grpc"
	"erp/platform/procurement-service/storage"

	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()
	st, err := storage.NewStore(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatal(err)
	}
	srv := grpc.NewServer()
	procv1.RegisterProcurementServiceServer(srv, grpcsvc.New(st))
	log.Println("procurement gRPC on", cfg.GRPCAddr)
	if err := srv.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
