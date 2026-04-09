package grpcsvc

import (
	"context"

	whv1 "erp/platform/genproto/erp/warehouse/v1"
	"erp/platform/warehouse-service/storage"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	whv1.UnimplementedWarehouseServiceServer
	st *storage.Store
}

func New(st *storage.Store) *Server {
	return &Server{st: st}
}

func (s *Server) CreateProduct(ctx context.Context, in *whv1.CreateProductRequest) (*whv1.Product, error) {
	id, ts, err := s.st.CreateProduct(ctx, in.GetSku(), in.GetName(), in.GetInitialQty())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &whv1.Product{Id: id.String(), Sku: in.GetSku(), Name: in.GetName(), OnHandQty: in.GetInitialQty(), CreatedAtUnix: ts}, nil
}

func (s *Server) GetProduct(ctx context.Context, in *whv1.GetProductRequest) (*whv1.Product, error) {
	id, err := uuid.Parse(in.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "id")
	}
	sku, name, qty, ts, err := s.st.GetProduct(ctx, id)
	if err == storage.ErrNotFound {
		return nil, status.Error(codes.NotFound, "product")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &whv1.Product{Id: in.GetId(), Sku: sku, Name: name, OnHandQty: qty, CreatedAtUnix: ts}, nil
}

func (s *Server) ListProducts(ctx context.Context, in *whv1.ListProductsRequest) (*whv1.ListProductsResponse, error) {
	limit := in.GetLimit()
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.st.ListProducts(ctx, limit, in.GetOffset())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var list []*whv1.Product
	for _, r := range rows {
		list = append(list, &whv1.Product{Id: r.ID, Sku: r.SKU, Name: r.Name, OnHandQty: r.Qty, CreatedAtUnix: r.Ts})
	}
	return &whv1.ListProductsResponse{Products: list}, nil
}

func (s *Server) AdjustStock(ctx context.Context, in *whv1.AdjustStockRequest) (*whv1.AdjustStockResponse, error) {
	id, err := uuid.Parse(in.GetProductId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "product_id")
	}
	q, err := s.st.AdjustStock(ctx, id, in.GetDelta())
	if err == storage.ErrNotFound {
		return nil, status.Error(codes.NotFound, "product")
	}
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	return &whv1.AdjustStockResponse{OnHandQty: q}, nil
}
