package storage

import (
	"context"
	"errors"
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
	_, err = s.Pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS invoices (
  id UUID PRIMARY KEY,
  purchase_order_id UUID,
  status TEXT NOT NULL DEFAULT 'open',
  amount DOUBLE PRECISION NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);`)
	if err != nil {
		pool.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() { s.Pool.Close() }

func (s *Store) CreateInvoice(ctx context.Context, poID *uuid.UUID, amount float64) (id uuid.UUID, ts int64, err error) {
	id = uuid.New()
	ts = time.Now().Unix()
	var po any
	if poID != nil {
		po = *poID
	}
	_, err = s.Pool.Exec(ctx, `INSERT INTO invoices (id, purchase_order_id, status, amount, created_at) VALUES ($1,$2,'open',$3,to_timestamp($4))`,
		id, po, amount, ts)
	return
}

func (s *Store) GetInvoice(ctx context.Context, id uuid.UUID) (poID string, st string, amount float64, ts int64, err error) {
	row := s.Pool.QueryRow(ctx, `SELECT coalesce(purchase_order_id::text, ''), status, amount, extract(epoch from created_at)::bigint FROM invoices WHERE id = $1`, id)
	err = row.Scan(&poID, &st, &amount, &ts)
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrNotFound
	}
	return
}

func (s *Store) ListInvoices(ctx context.Context, limit, offset int32) ([]struct {
	ID, POID, Status string
	Amount            float64
	Ts                int64
}, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT id::text, coalesce(purchase_order_id::text, ''), status, amount, extract(epoch from created_at)::bigint
		FROM invoices ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []struct {
		ID, POID, Status string
		Amount            float64
		Ts                int64
	}
	for rows.Next() {
		var r struct {
			ID, POID, Status string
			Amount            float64
			Ts                int64
		}
		if err := rows.Scan(&r.ID, &r.POID, &r.Status, &r.Amount, &r.Ts); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
