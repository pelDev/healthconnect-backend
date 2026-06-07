package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	application_ports "github.com/pelDev/health-connect/internal/application/ports"
	"github.com/pelDev/health-connect/internal/domain/repositories"
	postgres_repos "github.com/pelDev/health-connect/internal/infrastructure/postgres/repositories"
	"github.com/pelDev/health-connect/internal/infrastructure/postgres/sqlc"
)

type PostgresUoW struct {
	tx          pgx.Tx
	userStorage repositories.UserStorage

	done bool
}

func (u *PostgresUoW) UserRepo() repositories.UserStorage {
	return u.userStorage
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
		tx:          tx,
		userStorage: postgres_repos.NewUserStorage(sqlc.New(tx)),
	}, nil
}
