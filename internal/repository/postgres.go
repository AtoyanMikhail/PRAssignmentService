package repository

import (
	"context"
	"fmt"

	"github.com/AtoyanMikhail/PRAssignmentService/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository implements all repository interfaces
type PostgresRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// NewPostgresRepository creates a new PostgresRepository
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		pool:    pool,
		queries: db.New(pool),
	}
}

// GetQueries returns the underlying sqlc queries
func (r *PostgresRepository) GetQueries() *db.Queries {
	return r.queries
}

// WithTx creates a new repository instance with a transaction
func (r *PostgresRepository) WithTx(tx pgx.Tx) *PostgresRepository {
	return &PostgresRepository{
		pool:    r.pool,
		queries: r.queries.WithTx(tx),
	}
}

// Verify that PostgresRepository implements all interfaces
var (
	_ TeamRepository        = (*PostgresRepository)(nil)
	_ UserRepository        = (*PostgresRepository)(nil)
	_ PullRequestRepository = (*PostgresRepository)(nil)
	_ PRReviewerRepository  = (*PostgresRepository)(nil)
	_ StatisticsRepository  = (*PostgresRepository)(nil)
)

// ExecTx executes a function within a database transaction
func (r *PostgresRepository) ExecTx(ctx context.Context, fn func(*PostgresRepository) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	txRepo := r.WithTx(tx)
	err = fn(txRepo)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}
