package pandora

import "time"

type Feedback struct {
	ID string `json:"feedbackId"`

	CreatedOn time.Time `json:"feedbackDateCreated"`

	AlbumTitle  string `json:"albumTitle"`
	ArtistName  string `json:"artistName"`
	SongTitle   string `json:"songTitle"`
	StationName string `json:"stationName"`

	IsPositive bool `json:"isPositive"`

	MusicID   string `json:"musicId"`
	PandoraID string `json:"pandoraId"`
	StationID string `json:"stationId"`
}
