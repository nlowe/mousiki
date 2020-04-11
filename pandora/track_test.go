package pandora

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTrack_String(t *testing.T) {
	sut := &Track{
		MusicId:   "DummyID",
		TrackType: "Track",

		ArtistName: "DummyArtist",
		AlbumTitle: "DummyAlbum",
		SongTitle:  "DummySong",
	}

	require.Equal(t, "[Track:DummyID] DummySong - DummyArtist - DummyAlbum", sut.String())
}
