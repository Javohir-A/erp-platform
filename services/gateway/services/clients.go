package services

import (
	"context"
	"fmt"

	authv1 "erp/platform/genproto/erp/auth/v1"
	finv1 "erp/platform/genproto/erp/finance/v1"
	hrv1 "erp/platform/genproto/erp/hr/v1"
	procv1 "erp/platform/genproto/erp/procurement/v1"
	whv1 "erp/platform/genproto/erp/warehouse/v1"
	"erp/platform/gateway/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Clients struct {
	Auth        authv1.AuthServiceClient
	HR          hrv1.HRServiceClient
	Procurement procv1.ProcurementServiceClient
	Warehouse   whv1.WarehouseServiceClient
	Finance     finv1.FinanceServiceClient
	conns       []*grpc.ClientConn
}

func dial(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	return grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

func NewClients(ctx context.Context, cfg config.Config) (*Clients, error) {
	type p struct {
		addr string
		name string
	}
	addrs := []p{
		{cfg.AuthAddr, "auth"},
		{cfg.HRAddr, "hr"},
		{cfg.ProcAddr, "procurement"},
		{cfg.WarehouseAddr, "warehouse"},
		{cfg.FinanceAddr, "finance"},
	}
	var conns []*grpc.ClientConn
	for _, a := range addrs {
		if a.addr == "" {
			for _, c := range conns {
				_ = c.Close()
			}
			return nil, fmt.Errorf("empty grpc address: %s", a.name)
		}
		conn, err := dial(ctx, a.addr)
		if err != nil {
			for _, c := range conns {
				_ = c.Close()
			}
			return nil, fmt.Errorf("dial %s: %w", a.name, err)
		}
		conns = append(conns, conn)
	}
	c := &Clients{
		Auth:        authv1.NewAuthServiceClient(conns[0]),
		HR:          hrv1.NewHRServiceClient(conns[1]),
		Procurement: procv1.NewProcurementServiceClient(conns[2]),
		Warehouse:   whv1.NewWarehouseServiceClient(conns[3]),
		Finance:     finv1.NewFinanceServiceClient(conns[4]),
		conns:       conns,
	}
	return c, nil
}

func (c *Clients) Close() {
	for _, conn := range c.conns {
		_ = conn.Close()
	}
}
