package api

import "github.com/nlowe/mousiki/pandora"

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

type GetFeedbackRequest struct {
	PageSize   int `json:"pageSize"`
	StartIndex int `json:"startIndex"`
}

type GetFeedbackResponse struct {
	Total    int                `json:"total"`
	Feedback []pandora.Feedback `json:"feedback"`
}
