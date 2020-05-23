package mousiki

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/magiconair/properties/assert"
	"github.com/nlowe/mousiki/mocks"
	"github.com/nlowe/mousiki/pandora"
	"github.com/nlowe/mousiki/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestStationController_Play(t *testing.T) {
	s := pandora.Station{
		ID:   uuid.Must(uuid.NewRandom()).String(),
		Name: "Dummy Station Radio",
	}
	c := &mocks.Client{}
	p := &mocks.Player{}
	sut := NewStationController(c, p)
	sut.log = testutil.NopLogger()

	next := 0
	playlist := []string{"1", "2", "3", "4"}

	ctx, cancel := context.WithCancel(context.Background())
	c.On("GetMoreTracks", mock.Anything).Run(func(args mock.Arguments) {
		require.Equal(t, s.ID, args.String(0))
	}).Return(func(u string) []pandora.Track {
		a := testutil.MakeTrack()
		a.AudioUrl = playlist[next]
		next++

		b := testutil.MakeTrack()
		b.AudioUrl = playlist[next]
		next++

		return []pandora.Track{
			a,
			b,
		}
	}, nil)

	doneCh := make(chan error, 1)
	var played []string

	var doneChRet <-chan error = doneCh
	p.On("UpdateStream", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		url := args.String(0)
		played = append(played, url)
		assert.Equal(t, url, (<-sut.NotificationChan()).AudioUrl)

		if len(played) == 3 {
			cancel()
		} else {
			doneCh <- nil
		}
	})
	p.On("DoneChan").Return(doneChRet)

	sut.SwitchStations(s)
	go sut.Play(ctx)
	<-ctx.Done()

	require.Len(t, played, 3)
	require.Equal(t, []string{"1", "2", "3"}, played)
}
