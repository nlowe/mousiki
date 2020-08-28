package ui

import (
	"github.com/gdamore/tcell"
	"github.com/nlowe/mousiki/mousiki"
	"github.com/sirupsen/logrus"
	"gitlab.com/tslocum/cview"
)

const narrativePopupPageName = "narrativePopup"

type narrativePopup struct {
	*CenteredModal
	root *cview.TextView

	cancelFunc func()

	controller *mousiki.StationController
	pager      *cview.Pages

	log logrus.FieldLogger
}

func NewNarrativePopupForPager(cancelFunc func(), pager *cview.Pages, controller *mousiki.StationController) *narrativePopup {
	result := &narrativePopup{
		root:       cview.NewTextView(),
		cancelFunc: cancelFunc,
		controller: controller,
		pager:      pager,
		log:        logrus.WithField("prefix", narrativePopupPageName),
	}

	result.root.SetWrap(true).
		SetWordWrap(true).
		SetTitle("Explanation").
		SetBorder(true).
		SetBorderPadding(1, 1, 1, 1)

	result.CenteredModal = NewCenteredModal(result.root)

	pager.AddPage(narrativePopupPageName, result, true, false)
	return result
}

func (n *narrativePopup) Open() {
	if page, _ := n.pager.GetFrontPage(); page == narrativePopupPageName {
		return
	}

	if n.controller.NowPlaying() == nil {
		n.log.Debug("No Track is playing")
		return
	}

	n.log.Info("Fetching Track Narrative")
	narrative, err := n.controller.ExplainCurrentTrack()
	if err != nil {
		n.log.WithError(err).Errorf("Failed to explain current track")
		return
	}

	n.root.SetText(narrative.Paragraph)
	n.pager.ShowPage(narrativePopupPageName)
}

func (n *narrativePopup) Close() {
	if page, _ := n.pager.GetFrontPage(); page != narrativePopupPageName {
		return
	}

	n.pager.HidePage(narrativePopupPageName)
}

func (n *narrativePopup) HandleKey(ev *tcell.EventKey) *tcell.EventKey {
	if ev.Key() == tcell.KeyEscape || (ev.Key() == tcell.KeyRune && ev.Rune() == 'e') {
		n.Close()

		return nil
	}

	return ev
}
