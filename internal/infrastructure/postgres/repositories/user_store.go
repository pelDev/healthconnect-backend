package postgres_repos

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pelDev/health-connect/internal/domain/models"
	"github.com/pelDev/health-connect/internal/domain/repositories"
	"github.com/pelDev/health-connect/internal/infrastructure/postgres/sqlc"
)

type userStorage struct {
	q *sqlc.Queries
}

func (u *userStorage) Save(ctx context.Context, entity *models.User) error {
	params := sqlc.UpsertUserParams{
		ID:        entity.ID,
		FirstName: entity.FirstName,
		LastName:  entity.LastName,
		Email:     entity.Email,
		Password:  entity.Password,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}

	return u.q.UpsertUser(ctx, params)
}

func (u *userStorage) Delete(ctx context.Context, id uuid.UUID) error {
	return u.q.DeleteUserByID(ctx, id)
}

func (u *userStorage) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	row, err := u.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	var user models.User
	u.mapToDomain(row, &user)
	return &user, nil
}

func (u *userStorage) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	row, err := u.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	var user models.User
	u.mapToDomain(row, &user)
	return &user, nil
}

func (e *userStorage) mapToDomain(db sqlc.User, m *models.User) {
	m.ID = db.ID
	m.FirstName = db.FirstName
	m.LastName = db.LastName
	m.Email = db.Email
	m.Password = db.Password
	m.CreatedAt = db.CreatedAt
	m.Password = db.Password
}

func NewUserStorage(q *sqlc.Queries) repositories.UserStorage {
	return &userStorage{
		q: q,
	}
}
