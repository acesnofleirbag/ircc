package main

import (
	"ircc/src/guard"

	"github.com/gdamore/tcell/v2"
)

var Events chan tcell.Event

type UI struct {
	screen tcell.Screen
	prompt []rune
	exit   bool
	cursor Cursor
}

func NewScreen() UI {
	screen, err := tcell.NewScreen()
	guard.Err(err)

	err = screen.Init()
	guard.Err(err)

	screen.SetStyle(tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset))
	screen.SetCursorStyle(tcell.CursorStyleBlinkingBlock)
	screen.EnablePaste()
	screen.Clear()
	screen.Sync()

	Events = make(chan tcell.Event)

	return UI{
		screen: screen,
		exit:   false,
		cursor: NewCursor(),
	}
}

func (self *UI) Run() {
	S := self.screen
	defer S.Fini()

	go func() {
		for {
			Events <- S.PollEvent()
		}
	}()

	self.process()
}

func (self *UI) AddLine(msg string) {
	S := self.screen

	for i, ch := range msg {
		S.SetContent(i, self.cursor.y, ch, nil, tcell.StyleDefault)
	}

	self.cursor.y += 1

	S.ShowCursor(self.cursor.x, self.cursor.y)
	S.Show()
}

func (self *UI) process() {
	S := self.screen

	for !self.exit {
		event := <-Events

		S.Clear()

		switch event := event.(type) {
		case *tcell.EventResize:
			S.Sync()
		case *tcell.EventKey:
			switch event.Key() {
			case tcell.KeyEscape:
				self.exit = true
				break
			case tcell.KeyLeft:
				self.cursor.x -= 1
				break
			case tcell.KeyRune:
				{
					ch := event.Rune()

					self.prompt = append(self.prompt, ch)
					self.cursor.x += 1
				}
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				if self.cursor.x > 0 {
					self.cursor.x -= 1
					self.prompt = append(self.prompt[:self.cursor.x], self.prompt[self.cursor.x+1:]...)
				}
				break
			}
		}

		for i, ch := range self.prompt {
			S.SetContent(i, self.cursor.y, ch, nil, tcell.StyleDefault)
		}

		S.ShowCursor(self.cursor.x, self.cursor.y)
		S.Show()
	}
}
