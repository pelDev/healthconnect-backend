package application_dto

import "github.com/google/uuid"

type LoginReq struct {
	Email    string `json:"email,required"`
	Password string `json:"password,required"`
}

type LoginRes struct {
	SessionID uuid.UUID `json:"session_id,required"`
}
