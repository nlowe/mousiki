package testutil

import (
	"github.com/google/uuid"
	"github.com/nlowe/mousiki/pandora"
)

func MakeTrack() pandora.Track {
	trackID := uuid.Must(uuid.NewRandom()).String()

	return pandora.Track{
		MusicId:                    trackID,
		PandoraId:                  uuid.Must(uuid.NewRandom()).String(),
		StationId:                  uuid.Must(uuid.NewRandom()).String(),
		AudioTokenId:               uuid.Must(uuid.NewRandom()).String(),
		ArtistMusicId:              uuid.Must(uuid.NewRandom()).String(),
		TrackToken:                 uuid.Must(uuid.NewRandom()).String(),
		Identity:                   uuid.Must(uuid.NewRandom()).String(),
		UserSeed:                   uuid.Must(uuid.NewRandom()).String(),
		TrackType:                  pandora.TrackTypeTrack,
		FileGain:                   1.23,
		AudioUrl:                   "",
		AudioEncoding:              pandora.AudioFormatAACPlus,
		TrackLengthSeconds:         123,
		TrackRating:                0,
		AllowStartStationFromTrack: false,
		AllowShareTrack:            false,
		AllowTiredOfTrack:          false,
		AllowSkipTrackWithoutLimit: false,
		AllowSkip:                  false,
		AllowFeedback:              false,
		Rights:                     nil,
		AudioReceiptURL:            "",
		AudioSkipURL:               "",
		ArtistName:                 "Mousiki",
		AlbumTitle:                 "A New Era",
		SongTitle:                  "Testing",
		ComposerName:               "nlowe",
		ArtistArt:                  nil,
		AlbumArt:                   nil,
		Genre:                      nil,
		IsCompilation:              false,
		IsFeatured:                 false,
		IsBookmarked:               false,
		TrackKey: struct {
			TrackID   string            `json:"trackId"`
			TrackType pandora.TrackType `json:"trackType"`
			SpinId    string            `json:"spinId"`
		}{
			TrackID:   trackID,
			TrackType: pandora.TrackTypeTrack,
			SpinId:    uuid.Must(uuid.NewRandom()).String(),
		},
	}
}
