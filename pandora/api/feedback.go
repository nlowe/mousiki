package api

type AddFeedbackRequest struct {
	TrackToken string `json:"trackToken"`
	IsPositive bool   `json:"isPositive"`
}

type AddFeedbackResponse struct {
	ID        string `json:"feedbackId"`
	StationId string `json:"stationId"`
	MusicId   string `json:"musicId"`
	PandoraId string `json:"pandoraId"`

	IsPositive bool `json:"isPositive"`
}

type AddTiredRequest struct {
	TrackToken string `json:"trackToken"`
}

type AddTiredResponse struct{}
