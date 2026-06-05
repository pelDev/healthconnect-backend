package models

import "github.com/google/uuid"

type MessageRole string

const (
	MessageRoleUser       MessageRole = "user"
	MessageRoleAssistance MessageRole = "assistant"
	MessageRoleDoctor     MessageRole = "doctor"
)

type Message struct {
	ID        uuid.UUID
	SessionID uuid.UUID
	Role      MessageRole
	Content   string
	DoctorID  *uuid.UUID
	VID       *uuid.UUID // Acts as user identifier
}
