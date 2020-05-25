package ui

import (
	"fmt"

	"github.com/nlowe/mousiki/pandora"
)

var ratingColors = map[pandora.TrackRating]string{
	pandora.TrackRatingLike:    "gold",
	pandora.TrackRatingNeutral: "green",
	pandora.TrackRatingBan:     "red",
}

func FormatTrackTitle(t *pandora.Track) string {
	return fmt.Sprintf("[%s]%s[-]", ratingColors[t.Rating], t.SongTitle)
}

func FormatTrackArtist(t *pandora.Track) string {
	return fmt.Sprintf("[blue]%s[-]", t.ArtistName)
}

func FormatTrackAlbum(t *pandora.Track) string {
	return fmt.Sprintf("[orange]%s[-]", t.AlbumTitle)
}

func FormatTrack(t *pandora.Track, s pandora.Station) string {
	return fmt.Sprintf(
		"%s - %s - %s on [darkcyan]%s[-]",
		FormatTrackTitle(t),
		FormatTrackArtist(t),
		FormatTrackAlbum(t),
		s.Name,
	)
}
