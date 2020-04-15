package cmd

import (
	"math"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/google/uuid"
	"github.com/nlowe/mousiki/audio"
	"github.com/sirupsen/logrus"
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

		progress := cview.NewProgressBar()
		progress.SetMax(int(progressData.Duration.Seconds()))
		progress.SetProgress(int(progressData.Progress.Seconds()))
		progressText := cview.NewTextView().
			SetText(progressData.String()).
			SetTextAlign(cview.AlignRight)

		transport := cview.NewGrid().
			SetColumns(0, 13).
			AddItem(progress, 0, 0, 1, 1, 0, 0, false).
			AddItem(progressText, 0, 1, 1, 1, 0, 0, false)

		nowPlaying := cview.NewGrid().
			SetRows(1, 1, 1).
			SetColumns(-2, -6, -2).
			AddItem(cview.NewTextView().SetText("[gold]Test Track[-]").SetDynamicColors(true).SetTextAlign(cview.AlignCenter), 0, 1, 1, 1, 0, 0, false).
			AddItem(cview.NewTextView().SetText("[blue]Dummy Artist[-]").SetDynamicColors(true).SetTextAlign(cview.AlignCenter), 1, 1, 1, 1, 0, 0, false).
			AddItem(cview.NewTextView().SetText("[orange]Some Album[-]").SetDynamicColors(true).SetTextAlign(cview.AlignCenter), 2, 1, 1, 1, 0, 0, false)

		nowPlayingWrapper := cview.NewGrid().
			SetRows(3, 1).
			AddItem(nowPlaying, 0, 0, 1, 1, 0, 0, false).
			AddItem(transport, 1, 0, 1, 1, 0, 0, false)

		nowPlayingWrapper.SetTitle(" Now Playing - Test Station Radio ").SetBorder(true)

		history := cview.NewTextView().
			ScrollToEnd().
			SetDynamicColors(true).
			SetWrap(false).
			SetWordWrap(false).
			SetText(`[green]Coming Home[-] - [blue]Avenged Sevenfold[-] - [orange]Avenged Sevenfold[-]
[gold]82nd All The Way[-] - [blue]Sabaton[-] - [orange]The Great War[-]
[green]Captain Morgan's Revenge[-] - [blue]Alestorm[-] - [orange]Captain Morgan's Revenge[-]
[red]Some Country Song[-] - [blue]Random Artist[-] - [orange]Popular Album[-]`)
		history.SetDrawFunc(func(_ tcell.Screen, x, y, w, h int) (rx int, ry int, rw int, rh int) {
			contentHeight := strings.Count(strings.TrimSuffix(history.GetText(true), "\n"), "\n") + 1

			rx = x + 1
			rh = int(math.Min(float64(h-2), float64(contentHeight)))
			ry = (y + h - 1) - rh
			rw = w - 2
			return
		})
		history.SetTitle(" Previously Played ").SetBorder(true)

		upNext := cview.NewTextView().
			SetDynamicColors(true).
			SetText(`[green]Full Circle[-] - [blue]Five Finger Death Punch[-] - [orange]F8[-]
[gold]Brighter Side Of Grey[-] - [blue]Five Finger Death Punch[-] - [orange]F8[-]
[gold]M.I.A[-] - [blue]Avenged Sevenfold[-] - [orange]City of Evil[-]`)
		upNext.SetTitle(" Up Next ").SetBorder(true)

		player := cview.NewGrid().
			AddItem(history, 0, 0, 1, 1, 0, 0, false).
			AddItem(nowPlayingWrapper, 1, 0, 1, 1, 0, 0, false).
			AddItem(upNext, 2, 0, 1, 1, 0, 0, false).
			SetRows(-1, 6, 6)

		logs := cview.NewTextView().
			SetDynamicColors(true).
			ScrollToEnd()

		logrus.SetOutput(cview.ANSIWriter(logs))
		logrus.WithField("foo", "bar").Info("Test Info")
		logrus.WithField("foo", "bar").Warn("Test Warning with a really long line foo bar fizz buzz a b c one two three testing testing asdf fdsa")
		logrus.WithField("foo", "bar").Error("Test Error")

		logs.SetTitle(" Logs ").SetBorder(true)

		rootLayout := cview.NewGrid().
			SetRows(-10, -2).
			SetColumns(-1).
			AddItem(player, 0, 0, 1, 1, 0, 0, false).
			AddItem(logs, 1, 0, 1, 1, 0, 0, false)

		app := cview.NewApplication().
			SetRoot(rootLayout, true)

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
