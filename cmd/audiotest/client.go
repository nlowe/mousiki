package audiotest

import (
	"os"
	"os/signal"

	"github.com/nlowe/mousiki/audio"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var clientCmd = &cobra.Command{
	Use:     "client",
	Short:   "Test streaming music over HTTP",
	Long:    "Play a track over HTTP. Use with 'mousiki audiotest server'.",
	Example: "mousiki audiotest client http://localhost:5000/stream",
	Args:    cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		player, err := audio.NewGstreamerPipeline()
		if err != nil {
			return err
		}

		go func() {
			for progress := range player.ProgressChan() {
				logrus.WithField("progress", progress.String()).Info("Player made progress")
			}
		}()

		player.UpdateStream(args[0], 1.0)

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		select {
		case err := <-player.DoneChan():
			return err
		case <-c:
			return nil
		}
	},
}

func init() {
	RootCmd.AddCommand(clientCmd)
}
