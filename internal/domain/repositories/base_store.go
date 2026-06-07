package repositories

import (
	"context"

	"github.com/google/uuid"
)

type Storage[T any] interface {
	FindByID(ctx context.Context, id uuid.UUID) (*T, error)
	Create(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id uuid.UUID) error
}
