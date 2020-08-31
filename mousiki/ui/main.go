package ui

import (
	"context"
	"fmt"
	"io"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell"
	"github.com/nlowe/mousiki/audio"
	"github.com/nlowe/mousiki/mousiki"
	"github.com/nlowe/mousiki/pandora"
	"github.com/sirupsen/logrus"
	"gitlab.com/tslocum/cview"
)

const pageMain = "main"

type mainWindow struct {
	*cview.Pages

	stationPicker  *stationPicker
	narrativePopup *narrativePopup

	nowPlaying        mousiki.MessageTrackChanged
	nowPlayingSong    *cview.TextView
	nowPlayingArtist  *cview.TextView
	nowPlayingAlbum   *cview.TextView
	nowPlayingWrapper *cview.Grid

	shortcuts *cview.Grid

	progress     *cview.ProgressBar
	progressText *cview.TextView

	history *cview.TextView
	upNext  *cview.TextView

	player     audio.Player
	controller *mousiki.StationController

	quitRequested chan struct{}

	w   io.Writer
	log logrus.FieldLogger
}

func MainWindow(cancelFunc func(), player audio.Player, controller *mousiki.StationController) *mainWindow {
	logView := cview.NewTextView().
		SetDynamicColors(true).
		ScrollToEnd()

	logView.SetTitle(" Logs ").
		SetBorder(true)

	root := &mainWindow{
		Pages: cview.NewPages(),

		nowPlayingSong:   cview.NewTextView().SetDynamicColors(true),
		nowPlayingArtist: cview.NewTextView().SetDynamicColors(true),
		nowPlayingAlbum:  cview.NewTextView().SetDynamicColors(true),

		shortcuts: cview.NewGrid().SetRows(-1).
			SetColumns(-1, -1, 25, -1, -1, -1, -1),

		progress:     cview.NewProgressBar(),
		progressText: cview.NewTextView().SetTextAlign(cview.AlignRight),

		history: cview.NewTextView().SetDynamicColors(true).SetWordWrap(true),
		upNext:  cview.NewTextView().SetDynamicColors(true).SetWordWrap(true),

		player:     player,
		controller: controller,

		quitRequested: make(chan struct{}),

		w:   cview.ANSIWriter(logView),
		log: logrus.WithField("prefix", "ui"),
	}

	grid := cview.NewGrid()

	root.AddPage(pageMain, grid, true, true)
	root.stationPicker = NewStationPickerForPager(cancelFunc, root.Pages, controller)
	root.narrativePopup = NewNarrativePopupForPager(cancelFunc, root.Pages, controller)

	root.history.ScrollToEnd().
		SetDrawFunc(func(_ tcell.Screen, x, y, w, h int) (rx int, ry int, rw int, rh int) {
			contentHeight := strings.Count(strings.TrimSuffix(root.history.GetText(true), "\n"), "\n") + 1

			rx = x + 1
			rh = int(math.Min(float64(h-2), float64(contentHeight)))
			ry = (y + h - 1) - rh
			rw = w - 2
			return
		}).SetTitle(" Previously Played ").SetBorder(true)

	root.upNext.SetTitle(" Up Next ").
		SetBorder(true)

	root.nowPlayingSong.SetDynamicColors(true).SetTextAlign(cview.AlignCenter).SetText("?")
	root.nowPlayingArtist.SetDynamicColors(true).SetTextAlign(cview.AlignCenter).SetText("?")
	root.nowPlayingAlbum.SetDynamicColors(true).SetTextAlign(cview.AlignCenter).SetText("?")

	nowPlaying := cview.NewGrid().
		SetRows(1, 1, 1).
		SetColumns(-2, -6, -2).
		AddItem(root.nowPlayingSong, 0, 1, 1, 1, 0, 0, false).
		AddItem(root.nowPlayingArtist, 1, 1, 1, 1, 0, 0, false).
		AddItem(root.nowPlayingAlbum, 2, 1, 1, 1, 0, 0, false)

	transport := cview.NewGrid().
		SetColumns(0, 13).
		AddItem(root.progress, 0, 0, 1, 1, 0, 0, false).
		AddItem(root.progressText, 0, 1, 1, 1, 0, 0, false)

	root.nowPlayingWrapper = cview.NewGrid().
		SetRows(3, 1).
		AddItem(nowPlaying, 0, 0, 1, 1, 0, 0, false).
		AddItem(transport, 1, 0, 1, 1, 0, 0, false)

	root.nowPlayingWrapper.SetTitle(" No Station Selected ").
		SetBorder(true)

	stationView := cview.NewGrid().
		SetRows(-1, 6, 6).
		SetColumns(-1).
		AddItem(root.history, 0, 0, 1, 1, 0, 0, false).
		AddItem(root.nowPlayingWrapper, 1, 0, 1, 1, 0, 0, false).
		AddItem(root.upNext, 2, 0, 1, 1, 0, 0, false)

	grid.SetRows(-10, -2, 1).
		SetColumns(-1).
		AddItem(stationView, 0, 0, 1, 1, 0, 0, false).
		AddItem(logView, 1, 0, 1, 1, 0, 0, false).
		AddItem(root.shortcuts, 2, 0, 1, 1, 0, 0, false)

	root.Pages.SetChangedFunc(root.updateShortcuts)

	return root
}

func (w *mainWindow) updateShortcuts() {
	w.shortcuts.Clear()

	page, _ := w.Pages.GetFrontPage()
	if page == "main" {
		w.shortcuts.AddItem(cview.NewTextView().SetTextAlign(cview.AlignCenter).SetWrap(false).SetText("[Q] Quit"), 0, 0, 1, 1, 0, 0, false).
			AddItem(cview.NewTextView().SetTextAlign(cview.AlignCenter).SetWrap(false).SetText("[ESC] Stations"), 0, 1, 1, 1, 0, 0, false).
			AddItem(cview.NewTextView().SetTextAlign(cview.AlignCenter).SetWrap(false).SetText("[Space] Play / Pause"), 0, 2, 1, 1, 0, 0, false).
			AddItem(cview.NewTextView().SetTextAlign(cview.AlignCenter).SetWrap(false).SetText("[E] Explain"), 0, 3, 1, 1, 0, 0, false).
			AddItem(cview.NewTextView().SetTextAlign(cview.AlignCenter).SetWrap(false).SetText("[N] Next"), 0, 4, 1, 1, 0, 0, false).
			AddItem(cview.NewTextView().SetTextAlign(cview.AlignCenter).SetWrap(false).SetText("[-] Ban Song"), 0, 5, 1, 1, 0, 0, false).
			AddItem(cview.NewTextView().SetTextAlign(cview.AlignCenter).SetWrap(false).SetText("[T] Tired Of Song"), 0, 6, 1, 1, 0, 0, false).
			AddItem(cview.NewTextView().SetTextAlign(cview.AlignCenter).SetWrap(false).SetText("[+] Love Song"), 0, 7, 1, 1, 0, 0, false)
	} else if page == stationPickerPageName {
		w.shortcuts.AddItem(cview.NewTextView().SetTextAlign(cview.AlignCenter).SetWrap(false).SetText("[Q/ESC] Quit"), 0, 0, 1, 2, 0, 0, false).
			AddItem(cview.NewTextView().SetTextAlign(cview.AlignCenter).SetWrap(false).SetText("[Space/Enter] Change Station"), 0, 2, 1, 2, 0, 0, false)
	} else if page == narrativePopupPageName {
		w.shortcuts.AddItem(cview.NewTextView().SetTextAlign(cview.AlignCenter).SetWrap(false).SetText("[ESC/E] Close"), 0, 2, 1, 1, 0, 0, false)
	}
}

func (w *mainWindow) HandleKey(app *cview.Application) func(ev *tcell.EventKey) *tcell.EventKey {
	return func(ev *tcell.EventKey) *tcell.EventKey {
		if page, _ := w.GetFrontPage(); page == stationPickerPageName {
			return w.stationPicker.HandleKey(ev)
		} else if page == narrativePopupPageName {
			return w.narrativePopup.HandleKey(ev)
		}

		if ev.Key() == tcell.KeyRune && ev.Rune() == ' ' {
			if w.player.IsPlaying() {
				w.player.Pause()
				app.QueueUpdateDraw(func() {
					w.nowPlayingWrapper.SetBorderColor(tcell.ColorDarkRed)
					w.progress.SetFilledColor(tcell.ColorDimGray)
				})
			} else {
				w.player.Play()
				app.QueueUpdateDraw(func() {
					w.nowPlayingWrapper.SetBorderColor(tcell.ColorWhite)
					w.progress.SetFilledColor(tcell.ColorWhite)
				})
			}
		} else if ev.Key() == tcell.KeyF5 {
			w.log.Warn("Forcing re-draw")
			app.ForceDraw()
		} else if ev.Key() == tcell.KeyRune && ev.Rune() == 'n' {
			w.controller.Skip()
		} else if ev.Key() == tcell.KeyRune && ev.Rune() == 'q' {
			close(w.quitRequested)
		} else if ev.Key() == tcell.KeyEscape {
			w.ShowStationPicker()
		} else if ev.Key() == tcell.KeyRune && ev.Rune() == '+' {
			if err := w.controller.ProvideFeedback(pandora.TrackRatingLike); err != nil {
				w.log.WithError(err).Error("Failed to add feedback")
			}

			// Update NowPlaying with the same message to pick up the feedback
			w.updateNowPlaying(app, w.nowPlaying)
		} else if ev.Key() == tcell.KeyRune && ev.Rune() == 't' {
			if err := w.controller.ProvideFeedback(pandora.TrackRatingTired); err != nil {
				w.log.WithError(err).Error("Failed to add feedback")
			}
		} else if ev.Key() == tcell.KeyRune && ev.Rune() == '-' {
			if err := w.controller.ProvideFeedback(pandora.TrackRatingBan); err != nil {
				w.log.WithError(err).Error("Failed to add feedback")
			}
		} else if ev.Key() == tcell.KeyRune && ev.Rune() == 'e' {
			w.ShowNarrativePopup()
		} else {
			return ev
		}

		return nil
	}
}

func intClamp(n, low, high int) int {
	if n < low {
		return low
	}

	if n > high {
		return high
	}

	return n
}

func (w *mainWindow) OnResize(width, height int) {
	w.stationPicker.Resize(width/2, height/2)

	// TODO: Can we grow this automatically based on explanation length?
	w.narrativePopup.Resize(intClamp(width/2, 40, 120), intClamp(height/4, 10, 16))
}

func (w *mainWindow) ShowStationPicker() {
	w.stationPicker.Open()
}

func (w *mainWindow) ShowNarrativePopup() {
	w.narrativePopup.Open()
}

func (w *mainWindow) SyncData(ctx context.Context, app *cview.Application) {
	progress := w.player.ProgressChan()
	next := w.controller.NotificationChan()
	kickstart := sync.Once{}
	stationChanged := w.controller.StationChanged()
	pauseTicker := time.Tick(time.Second / 2)

	pauseColor := tcell.ColorDarkRed
	for {
		select {
		case <-w.quitRequested:
			app.Stop()
			return
		case <-ctx.Done():
			app.Stop()
			return
		case p := <-progress:
			w.updateProgress(app, p)
		case t := <-next:
			w.updateNowPlaying(app, t)
			w.updateUpNext(app)
		case station := <-stationChanged:
			w.nowPlayingWrapper.SetTitle(fmt.Sprintf(" Now Playing - %s ", station.Name))
			kickstart.Do(func() {
				w.stationPicker.EscapeAction = EscapeActionHide
				// Set the controller to playing after the first station is selected
				go w.controller.Play(ctx)
			})
		case <-pauseTicker:
			if !w.player.IsPlaying() {
				app.QueueUpdateDraw(func() {
					w.nowPlayingWrapper.SetBorderColor(pauseColor)
				})

				if pauseColor == tcell.ColorDarkRed {
					pauseColor = tcell.ColorWhite
				} else {
					pauseColor = tcell.ColorDarkRed
				}
			} else {
				app.QueueUpdateDraw(func() {
					w.nowPlayingWrapper.SetBorderColor(tcell.ColorWhite)
					w.progress.SetFilledColor(tcell.ColorWhite)
				})
			}
		}
	}
}

func (w *mainWindow) updateProgress(app *cview.Application, p audio.PlaybackProgress) {
	app.QueueUpdateDraw(func() {
		w.progress.SetMax(int(p.Duration.Seconds()))
		w.progress.SetProgress(int(p.Progress.Seconds()))
		w.progressText.SetText(p.String())

		if w.player.IsPlaying() {
			w.nowPlayingWrapper.SetBorderColor(tcell.ColorWhite)
			w.progress.SetFilledColor(tcell.ColorWhite)
		}
	})
}

func (w *mainWindow) updateNowPlaying(app *cview.Application, m mousiki.MessageTrackChanged) {
	app.QueueUpdateDraw(func() {
		w.nowPlayingSong.SetText(FormatTrackTitle(m.Track))
		w.nowPlayingArtist.SetText(FormatTrackArtist(m.Track))
		w.nowPlayingAlbum.SetText(FormatTrackAlbum(m.Track))

		if w.nowPlaying.Track != nil && w.nowPlaying.Track != m.Track {
			_, _ = w.history.Write([]byte("\n" + FormatTrack(w.nowPlaying.Track, w.nowPlaying.Station)))
		}

		w.nowPlaying = m
	})
}

func (w *mainWindow) updateUpNext(app *cview.Application) {
	app.QueueUpdateDraw(func() {
		station := w.controller.CurrentStation()

		buff := strings.Builder{}
		for _, t := range w.controller.UpNext() {
			_, _ = fmt.Fprintln(&buff, FormatTrack(&t, station))
		}

		w.upNext.SetText(strings.TrimSpace(buff.String()))
	})
}

func (w *mainWindow) Write(p []byte) (n int, err error) {
	return w.w.Write(p)
}
