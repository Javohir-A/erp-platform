package main

import (
	"context"
	"log"
	"net"

	whv1 "erp/platform/genproto/erp/warehouse/v1"
	"erp/platform/warehouse-service/config"
	grpcsvc "erp/platform/warehouse-service/grpc"
	"erp/platform/warehouse-service/storage"

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
	whv1.RegisterWarehouseServiceServer(srv, grpcsvc.New(st))
	log.Println("warehouse gRPC on", cfg.GRPCAddr)
	if err := srv.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
