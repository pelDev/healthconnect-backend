package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        uuid.UUID
	VID       uuid.UUID // Represents the user Identifier
	Reference *string
	EndedAt   *time.Time
	CreatedAt time.Time
}
