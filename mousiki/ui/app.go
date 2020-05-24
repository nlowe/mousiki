package ui

import (
	"context"

	"github.com/nlowe/mousiki/audio"
	"github.com/nlowe/mousiki/mousiki"
	"github.com/sirupsen/logrus"
	"gitlab.com/tslocum/cview"
)

func New(ctx context.Context, cancelFunc context.CancelFunc, player audio.Player, controller *mousiki.StationController) *cview.Application {
	root := MainWindow(cancelFunc, player, controller)
	app := cview.NewApplication().SetRoot(root, true)
	app.SetInputCapture(root.HandleKey(app))
	logrus.SetOutput(root)

	app.SetAfterResizeFunc(root.OnResize)
	app.QueueUpdateDraw(root.ShowStationPicker)

	go root.SyncData(ctx, app)
	return app
}
