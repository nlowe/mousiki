package audiotest

import (
	"os"
	"os/signal"

	"github.com/eiannone/keyboard"
	"github.com/nlowe/mousiki/audio"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var clientCmd = &cobra.Command{
	Use:     "client",
	Short:   "Test streaming music over HTTP",
	Long:    "Play a track over HTTP. Use with 'mousiki audiotest server'.",
	Example: "mousiki audiotest client http://localhost:5000/stream",
	Args:    cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		player, err := audio.NewBeepFFmpegPipeline()
		if err != nil {
			return err
		}
		defer func() {
			if err := player.Close(); err != nil {
				panic(err)
			}
		}()

		go func() {
			for progress := range player.ProgressChan() {
				logrus.WithField("progress", progress.String()).Info("Player made progress")
			}
		}()

		player.UpdateStream(args[0], viper.GetFloat64("gain"))

		if err := keyboard.Open(); err != nil {
			return err
		}
		defer func() {
			_ = keyboard.Close()
		}()

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		go func() {
			for {
				r, k, _ := keyboard.GetKey()
				if k == keyboard.KeySpace {
					if player.IsPlaying() {
						player.Pause()
					} else {
						player.Play()
					}
				} else if r == 'q' {
					close(c)
				}
			}
		}()

		select {
		case err := <-player.DoneChan():
			return err
		case <-c:
			return nil
		}
	},
}

func init() {
	flags := clientCmd.PersistentFlags()

	flags.Float64P("gain", "g", 0.0, "Relative File Gain (in dB) to apply")

	_ = viper.BindPFlags(flags)

	RootCmd.AddCommand(clientCmd)
}
