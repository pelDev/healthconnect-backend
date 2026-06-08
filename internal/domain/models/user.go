package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/pelDev/health-connect/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
}

func NewUser(firstName, lastName, email string, rawPassword string) (User, error) {
	now := time.Now().UTC()
	hashed, err := HashPassword(rawPassword)
	if err != nil {
		return User{}, err
	}

	return User{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		FirstName: firstName,
		LastName:  lastName,
		Email:     utils.NormalizeEmail(email),
		Password:  hashed,
	}, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (u User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
