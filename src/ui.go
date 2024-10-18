package main

import (
	"fmt"
	"ircc/src/guard"
	"strings"

	"github.com/gdamore/tcell/v2"
)

var Events chan tcell.Event

type Cursor struct {
	offset int
	x      int
	y      int
}

type UI struct {
	screen tcell.Screen
	prompt []rune
	exit   bool
	cursor Cursor
	hold   bool
}

func NewScreen() UI {
	screen, err := tcell.NewScreen()
	guard.Err(err)

	err = screen.Init()
	guard.Err(err)

	screen.SetStyle(tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset))
	screen.SetCursorStyle(tcell.CursorStyleBlinkingBlock)
	screen.EnableMouse()
	screen.Clear()
	screen.Sync()

	Events = make(chan tcell.Event)

	_, maxy := screen.Size()

	return UI{
		screen: screen,
		exit:   false,
		cursor: Cursor{x: 0, y: maxy - 1},
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

func (self *UI) Rehydrate() {
	chat := CLIENT.GetChat()
	S := self.screen

	S.Clear()

	if len(*chat) > self.cursor.y && !self.hold {
		self.cursor.offset = len(*chat) - self.cursor.y
	}

	for line, msg := range *chat {
		var data string

		if line < self.cursor.offset {
			continue
		}

		if line > self.cursor.y+self.cursor.offset {
			break
		}

		if strings.Compare(msg.Username, "INFO") == 0 {
			data = fmt.Sprintf("[%v] %v: %v\n", msg.Timestamp.Format("15:04"), msg.Username, msg.Data)
		} else {
			data = fmt.Sprintf("[%v] @%v: %v\n", msg.Timestamp.Format("15:04"), msg.Username, msg.Data)
		}

		for col, ch := range data {
			S.SetContent(col, line-self.cursor.offset, ch, nil, tcell.StyleDefault)
		}
	}

	self.clearPrompt()
	self.displayPrompt()

	S.Show()
}

func (self *UI) process() {
	S := self.screen

	for !self.exit {
		event := <-Events

		switch event := event.(type) {
		case *tcell.EventResize:
			S.Sync()
		case *tcell.EventMouse:
			switch event.Buttons() {
			case tcell.WheelUp:
				self.offsetUp()
				break
			case tcell.WheelDown:
				self.offsetDown()
				break
			}
		case *tcell.EventKey:
			if CLIENT.mode == Mode__Normal {
				switch event.Key() {
				case tcell.KeyRune:
					ch := event.Rune()

					switch ch {
					case ':':
						CLIENT.mode = Mode__Insert
						self.prompt = append(self.prompt, ':')
						self.cursor.x += 1
						break
					case 'i':
						CLIENT.mode = Mode__Insert
						break
					case 'h':
						if self.cursor.x > 0 {
							self.cursor.x -= 1
						}
						break
					case 'l':
						if self.cursor.x < len(self.prompt) {
							self.cursor.x += 1
						}
						break
					case 'j':
						self.offsetDown()
					case 'k':
						self.offsetUp()
						break
					}
					break
				}
			} else if CLIENT.mode == Mode__Insert {
				switch event.Key() {
				case tcell.KeyEscape:
					if CLIENT.mode == Mode__Insert {
						CLIENT.mode = Mode__Normal
						self.prompt = self.prompt[:0]
						self.cursor.x = 0
					}
					break
				case tcell.KeyEnter:
					if strings.HasPrefix(string(self.prompt), ":") {
						self.Cmd(string(self.prompt[1:]))
					} else {
						CLIENT.SendMessage(string(self.prompt))
						self.prompt = self.prompt[:0]
						self.cursor.x = 0
					}
					break
				case tcell.KeyRune:
					ch := event.Rune()

					self.prompt = append(self.prompt, ch)
					self.cursor.x += 1
					break
				case tcell.KeyBackspace, tcell.KeyBackspace2:
					if self.cursor.x > 0 {
						self.cursor.x -= 1
						self.prompt = append(self.prompt[:self.cursor.x], self.prompt[self.cursor.x+1:]...)
					}
					break
				}
			}
		}

		self.Rehydrate()
	}
}

func (self *UI) clearPrompt() {
	S := self.screen

	maxx, _ := S.Size()

	for i := 0; i < maxx; i++ {
		S.SetContent(i, self.cursor.y, ' ', nil, tcell.StyleDefault)
	}
}

func (self *UI) displayPrompt() {
	S := self.screen

	maxx, _ := S.Size()
	srvlen := len(CLIENT.server.name) + 1
	mode := CLIENT.GetMode()
	promptlen := len(self.prompt)
	gap := maxx - srvlen - promptlen - len(mode)

	// render current server
	for i, ch := range CLIENT.server.name {
		S.SetContent(i, self.cursor.y, ch, nil, tcell.StyleDefault)
	}

	// render prompt
	for i, ch := range self.prompt {
		S.SetContent(i+srvlen, self.cursor.y, ch, nil, tcell.StyleDefault)
	}

	// render gap
	for i := 0; i < gap; i++ {
		S.SetContent(srvlen+promptlen+i, self.cursor.y, ' ', nil, tcell.StyleDefault)
	}

	// render mode
	for i, ch := range mode {
		S.SetContent(srvlen+promptlen+gap+i, self.cursor.y, ch, nil, tcell.StyleDefault)
	}

	S.ShowCursor(self.cursor.x+srvlen, self.cursor.y)
}

func (self *UI) offsetUp() {
	if self.cursor.offset > 0 {
		self.cursor.offset -= 1
		self.hold = true
	}
}

func (self *UI) offsetDown() {
	chat := CLIENT.GetChat()

	if len(*chat) == self.cursor.offset+self.cursor.y {
		self.hold = false
	} else if len(*chat) > self.cursor.offset+self.cursor.y {
		self.cursor.offset += 1
	}
}

func (self *UI) Cmd(cmd string) {
	if strings.Compare(cmd, "q") == 0 || strings.Compare(cmd, "Q") == 0 {
		self.exit = true
	} else if strings.HasPrefix(cmd, "!") {
		args := strings.Split(cmd, " ")

		CLIENT.ExeBin(args[0][1:], strings.Join(args[1:], ""))
	} else if strings.HasPrefix(cmd, "/") {
		CLIENT.ExeCmd(cmd[1:])
	}

	CLIENT.mode = Mode__Normal
	self.prompt = self.prompt[:0]
	self.cursor.x = 0
}
