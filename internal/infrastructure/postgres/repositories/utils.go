package postgres_repos

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func toPGInt32(v *int32) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Valid: true, Int32: *v}
}

func fromPGInt32(v pgtype.Int4) *int32 {
	if !v.Valid {
		return nil
	}

	return &v.Int32
}

func toPGText(s *string) pgtype.Text {
	if s == nil || *s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func fromPGText(s pgtype.Text) *string {
	if s.Valid {
		return &s.String
	}
	return nil
}

func toPGTimestamp(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func fromPGTimestamp(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}

	// Check for zero time
	if t.Time.IsZero() {
		return nil
	}

	// Log the format being returned
	result := &t.Time
	return result
}

func fromPGUUID(t pgtype.UUID) *uuid.UUID {
	if t.Valid {
		u, err := uuid.FromBytes(t.Bytes[:])
		if err != nil {
			return nil
		}
		return &u
	}
	return nil
}

func toPGUUID(in *uuid.UUID) pgtype.UUID {
	if in == nil {
		return pgtype.UUID{
			Valid: false,
		}
	}

	byteVal, err := in.MarshalBinary()
	if err != nil {
		log.Panicln(err)
	}

	var bytes [16]byte
	copy(bytes[:], byteVal)

	return pgtype.UUID{
		Valid: true,
		Bytes: bytes,
	}
}
