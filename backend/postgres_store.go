package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresStore implements Store backed by PostgreSQL
type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(databaseURL string) (*PostgresStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	// Auto-create table on startup (simple migration)
	if err := migrate(ctx, pool); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return &PostgresStore{pool: pool}, nil
}

func migrate(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS tasks (
			id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			title       TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			status      TEXT NOT NULL DEFAULT 'todo',
			created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

func (s *PostgresStore) All() ([]*Task, error) {
	rows, err := s.pool.Query(context.Background(),
		`SELECT id, title, description, status, created_at, updated_at
		 FROM tasks ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		t := &Task{}
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	if tasks == nil {
		tasks = make([]*Task, 0)
	}
	return tasks, rows.Err()
}

func (s *PostgresStore) Get(id string) (*Task, error) {
	t := &Task{}
	err := s.pool.QueryRow(context.Background(),
		`SELECT id, title, description, status, created_at, updated_at
		 FROM tasks WHERE id = $1`, id).
		Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get task: %w", err)
	}
	return t, nil
}

func (s *PostgresStore) Create(title, description, status string) (*Task, error) {
	t := &Task{}
	err := s.pool.QueryRow(context.Background(),
		`INSERT INTO tasks (title, description, status)
		 VALUES ($1, $2, $3)
		 RETURNING id, title, description, status, created_at, updated_at`,
		title, description, status).
		Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	return t, nil
}

func (s *PostgresStore) Update(id, title, description, status string) (*Task, error) {
	t := &Task{}
	err := s.pool.QueryRow(context.Background(),
		`UPDATE tasks SET title = $2, description = $3, status = $4, updated_at = NOW()
		 WHERE id = $1
		 RETURNING id, title, description, status, created_at, updated_at`,
		id, title, description, status).
		Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update task: %w", err)
	}
	return t, nil
}

func (s *PostgresStore) Delete(id string) error {
	result, err := s.pool.Exec(context.Background(),
		`DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *PostgresStore) Close() error {
	s.pool.Close()
	return nil
}

func (s *PostgresStore) Ping() error {
	return s.pool.Ping(context.Background())
}
