package grpcsvc

import (
	"context"
	"strconv"

	finv1 "erp/platform/genproto/erp/finance/v1"
	"erp/platform/finance-service/storage"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	finv1.UnimplementedFinanceServiceServer
	st *storage.Store
}

func New(st *storage.Store) *Server {
	return &Server{st: st}
}

func (s *Server) CreateInvoice(ctx context.Context, in *finv1.CreateInvoiceRequest) (*finv1.Invoice, error) {
	amt, err := strconv.ParseFloat(in.GetAmount(), 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "amount")
	}
	var po *uuid.UUID
	if in.GetPurchaseOrderId() != "" {
		p, err := uuid.Parse(in.GetPurchaseOrderId())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "purchase_order_id")
		}
		po = &p
	}
	id, ts, err := s.st.CreateInvoice(ctx, po, amt)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	inv := &finv1.Invoice{Id: id.String(), Status: "open", Amount: in.GetAmount(), CreatedAtUnix: ts}
	if po != nil {
		inv.PurchaseOrderId = po.String()
	}
	return inv, nil
}

func (s *Server) GetInvoice(ctx context.Context, in *finv1.GetInvoiceRequest) (*finv1.Invoice, error) {
	id, err := uuid.Parse(in.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "id")
	}
	po, st, amt, ts, err := s.st.GetInvoice(ctx, id)
	if err == storage.ErrNotFound {
		return nil, status.Error(codes.NotFound, "invoice")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &finv1.Invoice{
		Id: in.GetId(), PurchaseOrderId: po, Status: st,
		Amount: strconv.FormatFloat(amt, 'f', -1, 64), CreatedAtUnix: ts,
	}, nil
}

func (s *Server) ListInvoices(ctx context.Context, in *finv1.ListInvoicesRequest) (*finv1.ListInvoicesResponse, error) {
	limit := in.GetLimit()
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.st.ListInvoices(ctx, limit, in.GetOffset())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var list []*finv1.Invoice
	for _, r := range rows {
		list = append(list, &finv1.Invoice{
			Id: r.ID, PurchaseOrderId: r.POID, Status: r.Status,
			Amount: strconv.FormatFloat(r.Amount, 'f', -1, 64), CreatedAtUnix: r.Ts,
		})
	}
	return &finv1.ListInvoicesResponse{Invoices: list}, nil
}
