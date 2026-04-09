package storage

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	Pool *pgxpool.Pool
}

func NewStore(ctx context.Context, dsn string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	s := &Store{Pool: pool}
	if err := s.migrate(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) migrate(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS suppliers (
			id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			contact_email TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS purchase_orders (
			id UUID PRIMARY KEY,
			supplier_id UUID NOT NULL REFERENCES suppliers(id),
			status TEXT NOT NULL DEFAULT 'draft',
			total_amount DOUBLE PRECISION NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS purchase_order_lines (
			id UUID PRIMARY KEY,
			purchase_order_id UUID NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
			product_id UUID NOT NULL,
			quantity INT NOT NULL,
			unit_price DOUBLE PRECISION NOT NULL
		);`,
	}
	for _, q := range stmts {
		if _, err := s.Pool.Exec(ctx, q); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) Close() { s.Pool.Close() }

func (s *Store) CreateSupplier(ctx context.Context, name, email string) (id uuid.UUID, ts int64, err error) {
	id = uuid.New()
	ts = time.Now().Unix()
	_, err = s.Pool.Exec(ctx, `INSERT INTO suppliers (id, name, contact_email, created_at) VALUES ($1,$2,$3,to_timestamp($4))`, id, name, email, ts)
	return
}

func (s *Store) ListSuppliers(ctx context.Context, limit, offset int32) ([]struct{ ID, Name, Email string; Ts int64 }, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT id::text, name, contact_email, extract(epoch from created_at)::bigint FROM suppliers
		ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []struct{ ID, Name, Email string; Ts int64 }
	for rows.Next() {
		var r struct{ ID, Name, Email string; Ts int64 }
		if err := rows.Scan(&r.ID, &r.Name, &r.Email, &r.Ts); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) CreatePurchaseOrder(ctx context.Context, supplier uuid.UUID, lines []struct {
	ProductID uuid.UUID
	Qty       int32
	UnitPrice float64
}) (poID uuid.UUID, total float64, ts int64, err error) {
	poID = uuid.New()
	ts = time.Now().Unix()
	for _, ln := range lines {
		total += float64(ln.Qty) * ln.UnitPrice
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, 0, 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	_, err = tx.Exec(ctx, `INSERT INTO purchase_orders (id, supplier_id, status, total_amount, created_at) VALUES ($1,$2,'open',$3,to_timestamp($4))`,
		poID, supplier, total, ts)
	if err != nil {
		return uuid.Nil, 0, 0, err
	}
	for _, ln := range lines {
		lid := uuid.New()
		_, err = tx.Exec(ctx, `INSERT INTO purchase_order_lines (id, purchase_order_id, product_id, quantity, unit_price) VALUES ($1,$2,$3,$4,$5)`,
			lid, poID, ln.ProductID, ln.Qty, ln.UnitPrice)
		if err != nil {
			return uuid.Nil, 0, 0, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, 0, 0, err
	}
	return poID, total, ts, nil
}

func (s *Store) GetPurchaseOrder(ctx context.Context, id uuid.UUID) (supplier string, status string, total float64, ts int64, lines []struct {
	ProductID string
	Qty       int32
	UnitPrice string
}, err error) {
	row := s.Pool.QueryRow(ctx, `SELECT supplier_id::text, status, total_amount, extract(epoch from created_at)::bigint FROM purchase_orders WHERE id = $1`, id)
	err = row.Scan(&supplier, &status, &total, &ts)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", 0, 0, nil, ErrNotFound
	}
	if err != nil {
		return
	}
	rows, e2 := s.Pool.Query(ctx, `SELECT product_id::text, quantity, unit_price FROM purchase_order_lines WHERE purchase_order_id = $1`, id)
	if e2 != nil {
		err = e2
		return
	}
	defer rows.Close()
	for rows.Next() {
		var pid string
		var q int32
		var up float64
		if e := rows.Scan(&pid, &q, &up); e != nil {
			err = e
			return
		}
		lines = append(lines, struct {
			ProductID string
			Qty       int32
			UnitPrice string
		}{pid, q, strconv.FormatFloat(up, 'f', -1, 64)})
	}
	err = rows.Err()
	return
}

func (s *Store) ListPurchaseOrders(ctx context.Context, limit, offset int32) ([]struct {
	ID, SupplierID, Status string
	Total                  float64
	Ts                     int64
}, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT id::text, supplier_id::text, status, total_amount, extract(epoch from created_at)::bigint
		FROM purchase_orders ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []struct {
		ID, SupplierID, Status string
		Total                  float64
		Ts                     int64
	}
	for rows.Next() {
		var r struct {
			ID, SupplierID, Status string
			Total                  float64
			Ts                     int64
		}
		if err := rows.Scan(&r.ID, &r.SupplierID, &r.Status, &r.Total, &r.Ts); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
