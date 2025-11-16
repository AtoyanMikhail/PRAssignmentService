package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store управляет всеми репозиториями и транзакциями
type Store struct {
	pool *pgxpool.Pool
	PostgresRepository
}

// NewStore создает новый Store
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		pool:               pool,
		PostgresRepository: *NewPostgresRepository(pool),
	}
}

// ExecTx выполняет функцию внутри транзакции
func (s *Store) ExecTx(ctx context.Context, fn func(*PostgresRepository) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}

	repo := s.WithTx(tx)
	err = fn(repo)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return rbErr
		}
		return err
	}

	return tx.Commit(ctx)
}

// ExecTxWithIsolation выполняет функцию внутри транзакции с заданным уровнем изоляции
func (s *Store) ExecTxWithIsolation(ctx context.Context, isolation pgx.TxIsoLevel, fn func(*PostgresRepository) error) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: isolation,
	})
	if err != nil {
		return err
	}

	repo := s.WithTx(tx)
	err = fn(repo)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return rbErr
		}
		return err
	}

	return tx.Commit(ctx)
}
