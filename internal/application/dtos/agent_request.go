package application_dto

type ReferDoctorData struct {
	Summary  string `json:"summary,required"`
	Symptoms string `json:"symptoms,required"`
}

type RaiseEmergencyData struct {
	Summary string `json:"summary,required"`
}
