package api

import "github.com/nlowe/mousiki/pandora"

type GetPlaylistFragmentRequest struct {
	StationID         string              `json:"stationId"`
	IsStationStart    bool                `json:"isStationStart"`
	AudioFormat       pandora.AudioFormat `json:"AudioFormat"`
	StartingAtTrackId *string             `json:"startingAtTrackId"`

	OnDemandArtistMessageArtistUidHex *string `json:"onDemandArtistMessageArtistUidHex"`
	OnDemandArtistMessageIdHex        *string `json:"onDemandArtistMessageIdHex"`
}

type GetPlaylistFragmentResponse struct {
	Tracks          []pandora.Track `json:"tracks"`
	IsBingeSkipping bool            `json:"isBingeSkipping"`
}
