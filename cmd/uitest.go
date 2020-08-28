package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nlowe/mousiki/audio"
	"github.com/nlowe/mousiki/mocks"
	"github.com/nlowe/mousiki/mousiki"
	"github.com/nlowe/mousiki/mousiki/ui"
	"github.com/nlowe/mousiki/pandora"
	"github.com/nlowe/mousiki/pandora/api"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

var uiTestCmd = &cobra.Command{
	Use:   "uitest",
	Short: "Test UI for mousiki",
	Long:  "A debug command for testing the UI of mousiki",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		progressData := audio.PlaybackProgress{
			Progress: 42 * time.Second,
			Duration: 187 * time.Second,
		}

		testProgress := make(chan audio.PlaybackProgress, 1)
		testProgress <- progressData
		var testProgressW <-chan audio.PlaybackProgress = testProgress

		doneChan := make(chan error, 1)
		var doneChanW <-chan error = doneChan

		playing := true

		player := &mocks.Player{}
		player.On("IsPlaying").Return(func() bool {
			return playing
		})
		player.On("ProgressChan").Return(testProgressW)
		player.On("DoneChan").Return(doneChanW)
		player.On("Pause").Run(func(_ mock.Arguments) {
			playing = false
		}).Return()
		player.On("Play").Run(func(_ mock.Arguments) {
			playing = true
		}).Return()
		player.On("UpdateStream", mock.Anything, mock.Anything).Return()

		ctx, cancel := context.WithCancel(context.TODO())

		app := ui.New(ctx, cancel, player, mousiki.NewStationController(testDataAPI(), player))
		return app.Run()
	},
}

func testDataAPI() api.Client {
	client := &mocks.Client{}

	var testStations []pandora.Station
	for i := 0; i < 25; i++ {
		id := uuid.Must(uuid.NewRandom())

		testStations = append(testStations, pandora.Station{
			ID:   id.String(),
			Name: fmt.Sprintf("Test Station %s", id.String()),
		})
	}

	client.On("GetStations").Return(testStations, nil)
	client.On("GetMoreTracks", mock.Anything).Return(func(_ string) []pandora.Track {
		return []pandora.Track{
			{
				StationId:  uuid.Must(uuid.NewRandom()).String(),
				MusicId:    uuid.Must(uuid.NewRandom()).String(),
				TrackToken: uuid.Must(uuid.NewRandom()).String(),
				ArtistName: "Test Artist",
				AlbumTitle: "Test Album",
				SongTitle:  fmt.Sprintf("Test Track %s", uuid.Must(uuid.NewRandom())),
			},
		}
	}, nil)
	client.On("AddFeedback", mock.Anything, mock.Anything).Return(nil)
	client.On("AddTired", mock.Anything).Return(nil)
	client.On("GetNarrative", mock.Anything, mock.Anything).Return(pandora.Narrative{
		Intro: "Based on what you've told us so far, we're playing this track because it features:",
		FocusTraits: []string{
			"vocal harmonies",
			"mixed acoustic and electric instrumentation",
			"minor key tonality",
			"mild rhythmic syncopation",
			"heavy electric rhythm guitars",
		},
		Paragraph: "Based on what you've told us so far, we're playing this track because it features vocal harmonies, mixed acoustic and electric instrumentation, minor key tonality, mild rhythmic syncopation and heavy electric rhythm guitars.",
	}, nil)

	return client
}

func init() {
	RootCmd.AddCommand(uiTestCmd)
}
