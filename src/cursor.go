package main

type Cursor struct {
	x int
	y int
}

func NewCursor() Cursor {
	return Cursor{
		x: 0,
		y: 0,
	}
}
