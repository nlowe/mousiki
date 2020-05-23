package ui

import (
	"github.com/gdamore/tcell"
	"github.com/nlowe/mousiki/mousiki"
	"github.com/nlowe/mousiki/pandora"
	"github.com/sirupsen/logrus"
	"gitlab.com/tslocum/cview"
)

const (
	stationPickerPageName = "stationPicker"

	pagerWidth  = 40
	pagerHeight = 10
)

const (
	EscapeActionExit = iota
	EscapeActionHide
)

type stationPicker struct {
	*cview.List

	cancelFunc func()

	controller *mousiki.StationController
	pager      *cview.Pages

	EscapeAction int

	log logrus.FieldLogger
}

func NewStationPickerForPager(cancelFunc func(), pager *cview.Pages, controller *mousiki.StationController) *stationPicker {
	root := &stationPicker{
		List: cview.NewList(),

		cancelFunc: cancelFunc,

		controller: controller,
		pager:      pager,

		log: logrus.WithField("prefix", "stationPicker"),
	}

	root.SetTitle(" Select Station ").SetBorder(true)
	root.SetBorder(true)
	root.ShowSecondaryText(false)

	// TODO: Dynamic width / height?
	// TODO: Why does this have no transparency?
	pager.AddPage(stationPickerPageName, cview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(cview.NewFlex().SetDirection(cview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(root, pagerHeight, 1, false).
			AddItem(nil, 0, 1, false), pagerWidth, 1, false).
		AddItem(nil, 0, 1, false),
		true, false)
	return root
}

func (s *stationPicker) Open() {
	if page, _ := s.pager.GetFrontPage(); page == stationPickerPageName {
		return
	}

	s.log.Info("Fetching Stations...")
	stations, err := s.controller.ListStations()
	if err != nil {
		s.log.WithError(err).Error("Failed to fetch station list")
		return
	}

	s.Clear()
	for _, station := range stations {
		s.log.WithField("name", station.Name).Debug("Found Station")
		s.AddItem(station.Name, station.ID, ' ', s.makeSwitchFunction(station))
	}

	s.pager.ShowPage(stationPickerPageName)
}

func (s *stationPicker) Close() {
	if page, _ := s.pager.GetFrontPage(); page != stationPickerPageName {
		return
	}

	s.pager.HidePage(stationPickerPageName)
}

func (s *stationPicker) HandleKey(ev *cview.EventKey) *cview.EventKey {
	if ev.Key() == tcell.KeyEscape {
		if s.EscapeAction == EscapeActionExit {
			s.cancelFunc()
		} else {
			s.Close()
		}

		return nil
	}

	return ev
}

func (s *stationPicker) makeSwitchFunction(station pandora.Station) func() {
	return func() {
		s.log.WithFields(logrus.Fields{
			"name": station.Name,
			"id":   station.ID,
		}).Debug("Switching Stations")

		s.controller.SwitchStations(station)
		s.Close()
	}
}
