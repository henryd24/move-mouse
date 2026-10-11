package mover

import "github.com/go-vgo/robotgo"

// Robot implementa Driver con robotgo.
type Robot struct{}

func (Robot) ScreenSize() (int, int) { return robotgo.GetScreenSize() }
func (Robot) Location() Point        { x, y := robotgo.Location(); return Point{x, y} }
func (Robot) MoveTo(p Point)         { robotgo.Move(p.X, p.Y) }
func (Robot) Click()                 { _ = robotgo.Click("left") }
func (Robot) Scroll(dy int)          { robotgo.Scroll(0, dy) }
