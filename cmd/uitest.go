package cmd

import (
	"time"

	"github.com/gdamore/tcell"
	"github.com/google/uuid"
	"github.com/nlowe/mousiki/audio"
	"github.com/nlowe/mousiki/ui"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var uiTestCmd = &cobra.Command{
	Use:   "uitest",
	Short: "Test UI for mousiki",
	Long:  "A debug command for testing the UI of mousiki",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		playlist := tview.NewBox().SetTitle("Playlist").SetBorder(true)

		progressData := audio.PlaybackProgress{
			Progress: 42 * time.Second,
			Duration: 187 * time.Second,
		}
		progress := ui.NewProgressBar()
		progress.SetMax(int(progressData.Duration.Seconds()))
		progress.SetProgress(int(progressData.Progress.Seconds()))
		progressText := tview.NewTextView().
			SetText(progressData.String()).
			SetTextAlign(tview.AlignRight)

		transport := tview.NewGrid().
			SetColumns(0, 13).
			AddItem(progress, 0, 0, 1, 1, 0, 0, false).
			AddItem(progressText, 0, 1, 1, 1, 0, 0, false)
		transport.SetTitle(" Now Playing: 🎼 Testing - 🎭 Mousiki - 💿 A New Era ").SetBorder(true)

		logs := tview.NewTextView().
			SetDynamicColors(true).
			ScrollToEnd()

		logrus.SetOutput(tview.ANSIWriter(logs))
		logrus.WithField("foo", "bar").Info("Test Info")
		logrus.WithField("foo", "bar").Warn("Test Warning with a really long line foo bar fizz buzz a b c one two three testing testing asdf fdsa")
		logrus.WithField("foo", "bar").Error("Test Error")

		logs.SetTitle("Logs").SetBorder(true)

		rootLayout := tview.NewGrid().
			SetRows(-10, 3, -2).
			AddItem(playlist, 0, 0, 1, 1, 0, 0, false).
			AddItem(transport, 1, 0, 1, 1, 0, 0, false).
			AddItem(logs, 2, 0, 1, 1, 0, 0, false)

		rootLayout.
			SetBorder(true).
			SetTitle("mousiki")

		app := tview.NewApplication().
			SetRoot(rootLayout, true).
			EnableMouse(true)

		app.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
			if ev.Key() == tcell.KeyRune && ev.Rune() == 'q' {
				app.Stop()
				return nil
			} else if ev.Key() == tcell.KeyRune && ev.Rune() == 'n' {
				logrus.WithField("next", uuid.Must(uuid.NewRandom()).String()).Info("Skipping to next track")
			} else if ev.Key() == tcell.KeyRight {
				progressData.Progress += 1 * time.Second
				if progressData.Progress > progressData.Duration {
					progressData.Progress = progressData.Duration
				}
			} else if ev.Key() == tcell.KeyLeft {
				progressData.Progress -= 1 * time.Second
				if progressData.Progress < 0 {
					progressData.Progress = 0
				}
			}

			progress.SetProgress(int(progressData.Progress.Seconds()))
			progressText.SetText(progressData.String())

			return ev
		})

		return app.Run()
	},
}

func init() {
	RootCmd.AddCommand(uiTestCmd)
}
