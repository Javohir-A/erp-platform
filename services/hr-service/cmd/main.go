package main

import (
	"context"
	"log"
	"net"

	"erp/platform/hr-service/config"
	grpcsvc "erp/platform/hr-service/grpc"
	"erp/platform/hr-service/storage"
	hrv1 "erp/platform/genproto/erp/hr/v1"

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
	hrv1.RegisterHRServiceServer(srv, grpcsvc.NewHRServer(st))
	log.Println("hr gRPC on", cfg.GRPCAddr)
	if err := srv.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
