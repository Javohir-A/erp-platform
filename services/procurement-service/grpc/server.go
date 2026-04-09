package grpcsvc

import (
	"context"
	"strconv"

	procv1 "erp/platform/genproto/erp/procurement/v1"
	"erp/platform/procurement-service/storage"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	procv1.UnimplementedProcurementServiceServer
	st *storage.Store
}

func New(st *storage.Store) *Server {
	return &Server{st: st}
}

func (s *Server) CreateSupplier(ctx context.Context, in *procv1.CreateSupplierRequest) (*procv1.Supplier, error) {
	id, ts, err := s.st.CreateSupplier(ctx, in.GetName(), in.GetContactEmail())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &procv1.Supplier{Id: id.String(), Name: in.GetName(), ContactEmail: in.GetContactEmail(), CreatedAtUnix: ts}, nil
}

func (s *Server) ListSuppliers(ctx context.Context, in *procv1.ListSuppliersRequest) (*procv1.ListSuppliersResponse, error) {
	limit := in.GetLimit()
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.st.ListSuppliers(ctx, limit, in.GetOffset())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var list []*procv1.Supplier
	for _, r := range rows {
		list = append(list, &procv1.Supplier{Id: r.ID, Name: r.Name, ContactEmail: r.Email, CreatedAtUnix: r.Ts})
	}
	return &procv1.ListSuppliersResponse{Suppliers: list}, nil
}

func (s *Server) CreatePurchaseOrder(ctx context.Context, in *procv1.CreatePurchaseOrderRequest) (*procv1.PurchaseOrder, error) {
	sup, err := uuid.Parse(in.GetSupplierId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "supplier_id")
	}
	var lines []struct {
		ProductID uuid.UUID
		Qty       int32
		UnitPrice float64
	}
	for _, ln := range in.GetLines() {
		pid, err := uuid.Parse(ln.GetProductId())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "product_id")
		}
		up, err := strconv.ParseFloat(ln.GetUnitPrice(), 64)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "unit_price")
		}
		lines = append(lines, struct {
			ProductID uuid.UUID
			Qty       int32
			UnitPrice float64
		}{pid, ln.GetQuantity(), up})
	}
	poID, total, ts, err := s.st.CreatePurchaseOrder(ctx, sup, lines)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var outLines []*procv1.PurchaseOrderLine
	for _, ln := range in.GetLines() {
		outLines = append(outLines, ln)
	}
	return &procv1.PurchaseOrder{
		Id: poID.String(), SupplierId: in.GetSupplierId(), Status: "open",
		TotalAmount: strconv.FormatFloat(total, 'f', -1, 64), CreatedAtUnix: ts, Lines: outLines,
	}, nil
}

func (s *Server) GetPurchaseOrder(ctx context.Context, in *procv1.GetPurchaseOrderRequest) (*procv1.PurchaseOrder, error) {
	id, err := uuid.Parse(in.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "id")
	}
	sup, st, total, ts, lines, err := s.st.GetPurchaseOrder(ctx, id)
	if err == storage.ErrNotFound {
		return nil, status.Error(codes.NotFound, "po")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var pl []*procv1.PurchaseOrderLine
	for _, ln := range lines {
		pl = append(pl, &procv1.PurchaseOrderLine{ProductId: ln.ProductID, Quantity: ln.Qty, UnitPrice: ln.UnitPrice})
	}
	return &procv1.PurchaseOrder{
		Id: in.GetId(), SupplierId: sup, Status: st,
		TotalAmount: strconv.FormatFloat(total, 'f', -1, 64), CreatedAtUnix: ts, Lines: pl,
	}, nil
}

func (s *Server) ListPurchaseOrders(ctx context.Context, in *procv1.ListPurchaseOrdersRequest) (*procv1.ListPurchaseOrdersResponse, error) {
	limit := in.GetLimit()
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.st.ListPurchaseOrders(ctx, limit, in.GetOffset())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var list []*procv1.PurchaseOrder
	for _, r := range rows {
		list = append(list, &procv1.PurchaseOrder{
			Id: r.ID, SupplierId: r.SupplierID, Status: r.Status,
			TotalAmount: strconv.FormatFloat(r.Total, 'f', -1, 64), CreatedAtUnix: r.Ts,
		})
	}
	return &procv1.ListPurchaseOrdersResponse{Orders: list}, nil
}
