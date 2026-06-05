package application_dto

type InitializeSessionRes struct {
	SessionID string                 `json:"session_id"`
	ICEConfig map[string]interface{} `json:"ice_config"`
}

type SendVoiceChatOfferReq struct {
	Sdp  string `json:"sdp,required"`
	Type string `json:"type,required"`
}

type SendVoiceChatOfferRes struct {
	Sdp string `json:"sdp,required"`
}
