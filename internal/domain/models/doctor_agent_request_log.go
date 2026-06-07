package models

import (
	"time"

	"github.com/google/uuid"
)

type DocAgentRequestLogAction string

const (
	DocAgentRequestLogActionDeclineRequest   DocAgentRequestLogAction = "decline_request"
	DocAgentRequestLogActionAcceptRequest    DocAgentRequestLogAction = "accept_request"
	DocAgentRequestLogActionSendMessage      DocAgentRequestLogAction = "send_message"
	DocAgentRequestLogActionSendPrescription DocAgentRequestLogAction = "send_prescription"
	DocAgentRequestLogActionViewRequest      DocAgentRequestLogAction = "view_request"
	DocAgentRequestLogActionEscalate         DocAgentRequestLogAction = "escalate"
)

type DocAgentRequestLog struct {
	ID        uuid.UUID
	RequestID uuid.UUID
	DocID     uuid.UUID
	Action    DocAgentRequestLogAction
	CreatedAt time.Time
}
