package main

import (
	"context"
	"log"
	"net"

	finv1 "erp/platform/genproto/erp/finance/v1"
	"erp/platform/finance-service/config"
	grpcsvc "erp/platform/finance-service/grpc"
	"erp/platform/finance-service/storage"

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
	finv1.RegisterFinanceServiceServer(srv, grpcsvc.New(st))
	log.Println("finance gRPC on", cfg.GRPCAddr)
	if err := srv.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
