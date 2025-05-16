package pg

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TokenStorage struct {
	pool *pgxpool.Pool
	sb   squirrel.StatementBuilderType
}

func NewTokenStorage(pool *pgxpool.Pool) *TokenStorage {
	return &TokenStorage{
		pool: pool,
		sb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (s *TokenStorage) GetCapacity(ctx context.Context, token string) (int, error) {
	query, args, err := s.sb.
		Select("capacity").
		From("token_buckets").
		Where(squirrel.Eq{"token": token}).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build query: %w", err)
	}

	var capacity int
	err = s.pool.QueryRow(ctx, query, args...).Scan(&capacity)
	if err != nil {
		return 0, fmt.Errorf("failed to get capacity: %w", err)
	}

	return capacity, nil
}

func (s *TokenStorage) SetCapacity(ctx context.Context, token string, capacity int) error {
	query, args, err := s.sb.
		Insert("token_buckets").
		Columns("token", "capacity").
		Values(token, capacity).
		Suffix("ON CONFLICT (token) DO UPDATE SET capacity = EXCLUDED.capacity").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = s.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to set capacity: %w", err)
	}

	return nil
}
