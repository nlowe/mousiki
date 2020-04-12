package api

import "github.com/nlowe/mousiki/pandora"

type FragmentRequestReason string

const (
	FragmentRequestReasonNormal FragmentRequestReason = "Normal"
	FragmentRequestReasonString FragmentRequestReason = "Skip"
)

type GetPlaylistFragmentRequest struct {
	StationID             string                `json:"stationId"`
	IsStationStart        bool                  `json:"isStationStart"`
	FragmentRequestReason FragmentRequestReason `json:"fragmentRequestReason"`
	AudioFormat           pandora.AudioFormat   `json:"audioFormat"`
	StartingAtTrackId     *string               `json:"startingAtTrackId"`

	OnDemandArtistMessageArtistUidHex *string `json:"onDemandArtistMessageArtistUidHex"`
	OnDemandArtistMessageIdHex        *string `json:"onDemandArtistMessageIdHex"`
}

type GetPlaylistFragmentResponse struct {
	Tracks          []pandora.Track `json:"tracks"`
	IsBingeSkipping bool            `json:"isBingeSkipping"`
}
