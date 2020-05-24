package ui

import "gitlab.com/tslocum/cview"

type CenteredModal struct {
	*cview.Grid

	root cview.Primitive
}

func NewCenteredModal(root cview.Primitive) *CenteredModal {
	grid := cview.NewGrid().
		SetRows(0, 1, 0).
		SetColumns(0, 1, 0).
		AddItem(root, 1, 1, 1, 1, 0, 0, true)

	grid.SetBackgroundTransparent(true)

	return &CenteredModal{
		Grid: grid,
		root: root,
	}
}

func (c *CenteredModal) Resize(width, height int) {
	c.Grid.SetRows(0, height, 0).
		SetColumns(0, width, 0)
}
