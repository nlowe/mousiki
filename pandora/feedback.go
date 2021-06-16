package pandora

import "time"

var FeedbackCSVHeaders = []string{
	"id", "createdOn", "album", "artist", "song", "station", "positive", "musicID", "pandoraID", "stationID",
}

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

func (f Feedback) MarshalCSV() []string {
	positive := "false"
	if f.IsPositive {
		positive = "true"
	}

	return []string{
		f.ID,
		f.CreatedOn.Format(time.RFC3339),
		f.AlbumTitle,
		f.ArtistName,
		f.SongTitle,
		f.StationName,
		positive,
		f.MusicID,
		f.PandoraID,
		f.StationID,
	}
}
