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
CREATE TABLE IF NOT EXISTS products (
  id UUID PRIMARY KEY,
  sku TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  on_hand_qty BIGINT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL
);`)
	if err != nil {
		pool.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() { s.Pool.Close() }

func (s *Store) CreateProduct(ctx context.Context, sku, name string, qty int64) (id uuid.UUID, ts int64, err error) {
	id = uuid.New()
	ts = time.Now().Unix()
	_, err = s.Pool.Exec(ctx, `
		INSERT INTO products (id, sku, name, on_hand_qty, created_at) VALUES ($1,$2,$3,$4,to_timestamp($5))`,
		id, sku, name, qty, ts)
	return
}

func (s *Store) GetProduct(ctx context.Context, id uuid.UUID) (sku, name string, qty int64, ts int64, err error) {
	row := s.Pool.QueryRow(ctx, `SELECT sku, name, on_hand_qty, extract(epoch from created_at)::bigint FROM products WHERE id = $1`, id)
	err = row.Scan(&sku, &name, &qty, &ts)
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrNotFound
	}
	return
}

func (s *Store) ListProducts(ctx context.Context, limit, offset int32) ([]struct {
	ID, SKU, Name string
	Qty           int64
	Ts            int64
}, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT id::text, sku, name, on_hand_qty, extract(epoch from created_at)::bigint FROM products
		ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []struct {
		ID, SKU, Name string
		Qty           int64
		Ts            int64
	}
	for rows.Next() {
		var r struct {
			ID, SKU, Name string
			Qty           int64
			Ts            int64
		}
		if err := rows.Scan(&r.ID, &r.SKU, &r.Name, &r.Qty, &r.Ts); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) AdjustStock(ctx context.Context, id uuid.UUID, delta int64) (newQty int64, err error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	row := tx.QueryRow(ctx, `SELECT on_hand_qty FROM products WHERE id = $1 FOR UPDATE`, id)
	var q int64
	if err := row.Scan(&q); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, err
	}
	q += delta
	if q < 0 {
		return 0, errors.New("insufficient stock")
	}
	_, err = tx.Exec(ctx, `UPDATE products SET on_hand_qty = $2 WHERE id = $1`, id, q)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return q, nil
}
