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
	if err := s.migrate(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) migrate(ctx context.Context) error {
	if _, err := s.Pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS departments (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL
);`); err != nil {
		return err
	}
	_, err := s.Pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS employees (
  id UUID PRIMARY KEY,
  department_id UUID NOT NULL REFERENCES departments(id),
  user_id UUID,
  full_name TEXT NOT NULL,
  job_title TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL
);`)
	return err
}

func (s *Store) Close() { s.Pool.Close() }

func (s *Store) CreateDepartment(ctx context.Context, name string) (id uuid.UUID, created int64, err error) {
	id = uuid.New()
	t := time.Now().Unix()
	_, err = s.Pool.Exec(ctx, `INSERT INTO departments (id, name, created_at) VALUES ($1, $2, to_timestamp($3))`, id, name, t)
	return id, t, err
}

func (s *Store) ListDepartments(ctx context.Context, limit, offset int32) ([]struct {
	ID, Name string
	Created  int64
}, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT id::text, name, extract(epoch from created_at)::bigint FROM departments
		ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []struct {
		ID, Name string
		Created  int64
	}
	for rows.Next() {
		var r struct {
			ID, Name string
			Created  int64
		}
		if err := rows.Scan(&r.ID, &r.Name, &r.Created); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) CreateEmployee(ctx context.Context, deptID uuid.UUID, userID *uuid.UUID, fullName, jobTitle string) (id uuid.UUID, created int64, err error) {
	id = uuid.New()
	t := time.Now().Unix()
	var uid any
	if userID != nil {
		uid = *userID
	}
	_, err = s.Pool.Exec(ctx, `
		INSERT INTO employees (id, department_id, user_id, full_name, job_title, created_at)
		VALUES ($1, $2, $3, $4, $5, to_timestamp($6))`,
		id, deptID, uid, fullName, jobTitle, t)
	return id, t, err
}

func (s *Store) GetEmployee(ctx context.Context, id uuid.UUID) (deptID, userID, fullName, jobTitle string, created int64, err error) {
	row := s.Pool.QueryRow(ctx, `
		SELECT department_id::text, coalesce(user_id::text, ''), full_name, job_title, extract(epoch from created_at)::bigint
		FROM employees WHERE id = $1`, id)
	var uid string
	err = row.Scan(&deptID, &uid, &fullName, &jobTitle, &created)
	userID = uid
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrNotFound
	}
	return
}

func (s *Store) ListEmployees(ctx context.Context, limit, offset int32) ([]struct {
	ID, DeptID, UserID, FullName, JobTitle string
	Created                                int64
}, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT id::text, department_id::text, coalesce(user_id::text, ''), full_name, job_title, extract(epoch from created_at)::bigint
		FROM employees ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []struct {
		ID, DeptID, UserID, FullName, JobTitle string
		Created                                int64
	}
	for rows.Next() {
		var r struct {
			ID, DeptID, UserID, FullName, JobTitle string
			Created                                int64
		}
		if err := rows.Scan(&r.ID, &r.DeptID, &r.UserID, &r.FullName, &r.JobTitle, &r.Created); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
