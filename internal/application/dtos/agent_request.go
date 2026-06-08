package application_dto

type ReferDoctorData struct {
	Summary            string  `json:"summary,required"`
	Symptoms           string  `json:"symptoms,required"`
	Sex                string  `json:"sex,required"`
	Age                string  `json:"age,required"`
	Onset              string  `json:"onset,required"`
	Duration           *string `json:"duration,omitempty"`
	Severity           *string `json:"severity,omitempty"`
	MedicalHistory     *string `json:"medical_history,omitempty"`
	CurrentMedications *string `json:"current_medications,omitempty"`
	Allergies          *string `json:"allergies,omitempty"`
}

type RaiseEmergencyData struct {
	Summary string `json:"summary,required"`
}
