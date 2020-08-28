package ui

import (
	"github.com/gdamore/tcell"
	"github.com/nlowe/mousiki/mousiki"
	"github.com/nlowe/mousiki/pandora"
	"github.com/sirupsen/logrus"
	"gitlab.com/tslocum/cview"
)

const stationPickerPageName = "stationPicker"

const (
	EscapeActionExit = iota
	EscapeActionHide
)

type stationPicker struct {
	*CenteredModal
	list *cview.List

	cancelFunc func()

	controller *mousiki.StationController
	pager      *cview.Pages

	EscapeAction int

	log logrus.FieldLogger
}

func NewStationPickerForPager(cancelFunc func(), pager *cview.Pages, controller *mousiki.StationController) *stationPicker {
	root := &stationPicker{
		list: cview.NewList(),

		cancelFunc: cancelFunc,

		controller: controller,
		pager:      pager,

		log: logrus.WithField("prefix", stationPickerPageName),
	}

	root.list.ShowSecondaryText(false).
		SetTitle(" Select Station ").
		SetBorder(true)

	root.CenteredModal = NewCenteredModal(root.list)

	pager.AddPage(stationPickerPageName, root, true, false)
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

	currentStation := s.controller.CurrentStation()

	s.list.Clear()
	for _, station := range stations {
		shortcut := ' '
		if currentStation.ID == station.ID {
			shortcut = '*'
		}

		s.log.WithField("name", station.Name).Debug("Found Station")
		s.list.AddItem(station.Name, station.ID, shortcut, s.makeSwitchFunction(station))
	}

	s.pager.ShowPage(stationPickerPageName)
}

func (s *stationPicker) Close() {
	if page, _ := s.pager.GetFrontPage(); page != stationPickerPageName {
		return
	}

	s.pager.HidePage(stationPickerPageName)
}

func (s *stationPicker) HandleKey(ev *tcell.EventKey) *tcell.EventKey {
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
