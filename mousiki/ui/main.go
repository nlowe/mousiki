package ui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math"
	"strings"
	"sync"

	"github.com/gdamore/tcell"
	"github.com/nlowe/mousiki/audio"
	"github.com/nlowe/mousiki/mousiki"
	"github.com/nlowe/mousiki/pandora"
	"github.com/sirupsen/logrus"
	"gitlab.com/tslocum/cview"
)

type mainWindow struct {
	*cview.Pages

	stationPicker *stationPicker

	nowPlaying        pandora.Track
	nowPlayingSong    *cview.TextView
	nowPlayingArtist  *cview.TextView
	nowPlayingAlbum   *cview.TextView
	nowPlayingWrapper *cview.Grid

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

		progress:     cview.NewProgressBar(),
		progressText: cview.NewTextView().SetTextAlign(cview.AlignRight),

		history: cview.NewTextView().SetDynamicColors(true),
		upNext:  cview.NewTextView().SetDynamicColors(true),

		player:     player,
		controller: controller,

		quitRequested: make(chan struct{}),

		w:   cview.ANSIWriter(logView),
		log: logrus.WithField("prefix", "ui"),
	}

	grid := cview.NewGrid()

	root.AddPage("main", grid, true, true)
	root.stationPicker = NewStationPickerForPager(cancelFunc, root.Pages, controller)

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

	grid.SetRows(-10, -2).
		SetColumns(-1).
		AddItem(stationView, 0, 0, 1, 1, 0, 0, false).
		AddItem(logView, 1, 0, 1, 1, 0, 0, false)

	return root
}

func (w *mainWindow) HandleKey(app *cview.Application) func(ev *tcell.EventKey) *tcell.EventKey {
	return func(ev *tcell.EventKey) *tcell.EventKey {
		if page, _ := w.GetFrontPage(); page == stationPickerPageName {
			return w.stationPicker.HandleKey(ev)
		}

		if ev.Key() == tcell.KeyRune && ev.Rune() == ' ' {
			if w.player.IsPlaying() {
				w.player.Pause()
			} else {
				w.player.Play()
			}
		} else if ev.Key() == tcell.KeyF5 {
			w.log.Warn("Forcing re-draw")
			app.ForceDraw()
		} else if ev.Key() == tcell.KeyRune && ev.Rune() == 'n' {
			w.controller.Skip()
		} else if ev.Key() == tcell.KeyRune && ev.Rune() == 'q' {
			close(w.quitRequested)
		} else if ev.Key() == tcell.KeyEscape {
			w.ShowStationPicker(app)
		} else {
			return ev
		}

		return nil
	}
}

func (w *mainWindow) ShowStationPicker(app *cview.Application) {
	w.stationPicker.Open()
	app.SetFocus(w.stationPicker)
}

func (w *mainWindow) SyncData(ctx context.Context, app *cview.Application) {
	progress := w.player.ProgressChan()
	next := w.controller.NotificationChan()

	go func() {
		kickstart := sync.Once{}

		for {
			station := <-w.controller.StationChanged()
			w.nowPlayingWrapper.SetTitle(fmt.Sprintf(" Now Playing - %s ", station.Name))
			kickstart.Do(func() {
				w.stationPicker.EscapeAction = EscapeActionHide
				// Set the controller to playing after the first station is selected
				go w.controller.Play(ctx)
			})
		}
	}()

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
		}
	}
}

func (w *mainWindow) updateProgress(app *cview.Application, p audio.PlaybackProgress) {
	app.QueueUpdateDraw(func() {
		w.progress.SetMax(int(p.Duration.Seconds()))
		w.progress.SetProgress(int(p.Progress.Seconds()))
		w.progressText.SetText(p.String())
	})
}

func (w *mainWindow) updateNowPlaying(app *cview.Application, t pandora.Track) {
	app.QueueUpdateDraw(func() {
		w.nowPlayingSong.SetText(FormatTrackTitle(t))
		w.nowPlayingArtist.SetText(FormatTrackArtist(t))
		w.nowPlayingAlbum.SetText(FormatTrackAlbum(t))

		if w.nowPlaying.SongTitle != "" {
			_, _ = w.history.Write([]byte("\n" + FormatTrack(w.nowPlaying)))
		}

		w.nowPlaying = t

		w.upNext.Clear()
		buff := bufio.NewWriter(w.upNext)
		for _, t := range w.controller.UpNext() {
			_, _ = buff.WriteString(FormatTrack(t) + "\n")
		}
		_ = buff.Flush()
	})
}

func (w *mainWindow) Write(p []byte) (n int, err error) {
	return w.w.Write(p)
}
