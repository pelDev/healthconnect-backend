package models

import (
	"time"

	"github.com/google/uuid"
)

type SymptomsSeverity string

const (
	SymptomsSeverityMild     SymptomsSeverity = "mild"
	SymptomsSeverityModerate SymptomsSeverity = "moderate"
	SymptomsSeveritySevere   SymptomsSeverity = "severe"
)

type SymptomList struct {
	SessionID   uuid.UUID
	Symptoms    []string          `json:"symptoms"`
	Duration    string            `json:"duration"`
	Severity    SymptomsSeverity  `json:"severity"`          // mild, moderate, severe
	Context     map[string]string `json:"context,omitempty"` // e.g., "onset": "2 hours ago"
	CollectedAt time.Time         `json:"collected_at"`      // timestamp
}
