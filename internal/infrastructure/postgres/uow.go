package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	application_ports "github.com/pelDev/health-connect/internal/application/ports"
	"github.com/pelDev/health-connect/internal/domain/repositories"
)

type PostgresUoW struct {
	tx pgx.Tx

	done bool
}

// UserRepo implements application_ports.UnitOfWork.
func (u *PostgresUoW) UserRepo() repositories.UserStorage {
	panic("unimplemented")
}

func (u *PostgresUoW) Commit(ctx context.Context) error {
	if u.done {
		return nil
	}
	u.done = true
	return u.tx.Commit(ctx)
}

func (u *PostgresUoW) Rollback(ctx context.Context) error {
	if u.done {
		return nil
	}
	u.done = true
	return u.tx.Rollback(ctx)
}

func NewPostgresUoW(ctx context.Context, db *pgxpool.Pool) (application_ports.UnitOfWork, error) {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}

	return &PostgresUoW{
		tx: tx,
	}, nil
}
