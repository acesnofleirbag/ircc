package main

import (
	"ircc/src/guard"

	"github.com/gdamore/tcell/v2"
)

type UI struct {
	screen tcell.Screen
}

func (self *UI) NewScreen() UI {
	screen, err := tcell.NewScreen()
	guard.Err(err)

	return UI{
		screen: screen,
	}
}
