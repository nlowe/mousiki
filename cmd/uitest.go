package cmd

import (
	"context"
	"time"

	"github.com/nlowe/mousiki/audio"
	"github.com/nlowe/mousiki/mocks"
	"github.com/nlowe/mousiki/mousiki"
	"github.com/nlowe/mousiki/mousiki/ui"
	"github.com/nlowe/mousiki/pandora"
	"github.com/spf13/cobra"
	"gitlab.com/tslocum/cview"
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

		testStation := pandora.Station{Name: "Test Station Radio"}
		player := &mocks.Player{}
		player.On("ProgressChan").Return(testProgressW)

		root := ui.MainWindow(testStation, player, mousiki.NewStationController(testStation, &mocks.Client{}, player))
		app := cview.NewApplication().SetRoot(root, true)
		app.SetInputCapture(root.HandleKey(app))

		go root.SyncData(context.TODO(), app)
		return app.Run()
	},
}

func init() {
	RootCmd.AddCommand(uiTestCmd)
}
