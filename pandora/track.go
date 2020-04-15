package pandora

import "fmt"

type TrackType string

const (
	TrackTypeTrack TrackType = "Track"
)

type TrackRating int

// TODO: Verify these are correct
const (
	TrackRatingNeutral = 0
	TrackRatingLike    = 1
	TrackRatingBan     = -1
)

type Track struct {
	MusicId       string `json:"musicId"`
	PandoraId     string `json:"pandoraId"`
	StationId     string `json:"stationId"`
	AudioTokenId  string `json:"audioTokenId"`
	ArtistMusicId string `json:"artistMusicId"`
	TrackToken    string `json:"trackToken"`
	Identity      string `json:"identity"`

	UserSeed  string    `json:"userSeed"`
	TrackType TrackType `json:"trackType"`

	FileGain           float64     `json:"fileGain,string"`
	AudioUrl           string      `json:"audioURL"`
	AudioEncoding      AudioFormat `json:"audioEncoding"`
	TrackLengthSeconds int         `json:"trackLength"`
	Rating             TrackRating `json:"rating"`

	AllowStartStationFromTrack bool     `json:"allowStartStationFromTrack"`
	AllowShareTrack            bool     `json:"allowShareTrack"`
	AllowTiredOfTrack          bool     `json:"allowTiredOfTrack"`
	AllowSkipTrackWithoutLimit bool     `json:"allowSkipTrackWithoutLimit"`
	AllowSkip                  bool     `json:"allowSkip"`
	AllowFeedback              bool     `json:"allowFeedback"`
	Rights                     []string `json:"rights"`

	// TODO: Do we need to call these when we download/skip the track?
	AudioReceiptURL string `json:"audioReceiptURL"`
	AudioSkipURL    string `json:"audioSkipUrl"`

	ArtistName   string   `json:"ArtistName"`
	AlbumTitle   string   `json:"albumTitle"`
	SongTitle    string   `json:"songTitle"`
	ComposerName string   `json:"composerName"`
	ArtistArt    []Art    `json:"artistArt"`
	AlbumArt     []Art    `json:"albumArt"`
	Genre        []string `json:"genre"`

	IsCompilation bool `json:"isCompilation"`
	IsFeatured    bool `json:"isFeatured"`
	IsBookmarked  bool `json:"isBookmarked"`

	TrackKey struct {
		TrackID   string    `json:"trackId"`
		TrackType TrackType `json:"trackType"`
		SpinId    string    `json:"spinId"`
	} `json:"trackKey"`
}

func (t Track) String() string {
	return fmt.Sprintf("[%s:%s] %s - %s - %s", t.TrackType, t.MusicId, t.SongTitle, t.ArtistName, t.AlbumTitle)
}
