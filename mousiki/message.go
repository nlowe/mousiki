package mousiki

import "github.com/nlowe/mousiki/pandora"

type MessageTrackChanged struct {
	Track   *pandora.Track
	Station pandora.Station
}
