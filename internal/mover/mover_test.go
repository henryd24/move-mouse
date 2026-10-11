package mover

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"
)

type fakeDriver struct {
	w, h   int
	pos    Point
	moves  []Point
	clicks int
	// onMove permite simular al usuario moviendo el mouse.
	onMove func(*fakeDriver)
}

func (f *fakeDriver) ScreenSize() (int, int) { return f.w, f.h }
func (f *fakeDriver) Location() Point        { return f.pos }
func (f *fakeDriver) MoveTo(p Point)         { f.pos = p; f.moves = append(f.moves, p) }
func (f *fakeDriver) Click()                 { f.clicks++ }
func (f *fakeDriver) Scroll(int)             {}

func fastEngine(cfg Config, d Driver) *Engine {
	e := NewEngine(cfg, d)
	e.Rand = rand.New(rand.NewSource(1))
	e.Sleep = func(ctx context.Context, _ time.Duration) error { return ctx.Err() }
	return e
}

func TestBezierPathEndsAtTarget(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	pts := BezierPath(Point{0, 0}, Point{500, 300}, 50, rng)
	if len(pts) != 50 || pts[len(pts)-1] != (Point{500, 300}) {
		t.Fatalf("trayectoria invalida: len=%d last=%v", len(pts), pts[len(pts)-1])
	}
}

func TestValidate(t *testing.T) {
	c := Default()
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	bad := []func(*Config){
		func(c *Config) { c.Mode = "x" },
		func(c *Config) { c.MaxInterval = 0 },
		func(c *Config) { c.JigglePixels = 0 },
		func(c *Config) { c.Region = &Region{W: 0, H: 10} },
		func(c *Config) { c.Schedule = Schedule{Start: "09:00"} },
		func(c *Config) { c.Schedule = Schedule{Days: []string{"xyz"}} },
	}
	for i, mod := range bad {
		c := Default()
		mod(&c)
		if c.Validate() == nil {
			t.Errorf("caso %d deberia fallar", i)
		}
	}
}

func TestScheduleActive(t *testing.T) {
	mon := time.Date(2026, 10, 12, 10, 0, 0, 0, time.UTC) // lunes
	sat := time.Date(2026, 10, 10, 10, 0, 0, 0, time.UTC)
	s := Schedule{Start: "09:00", End: "18:00", Days: []string{"mon", "fri"}}
	if !s.Active(mon) || s.Active(sat) || s.Active(mon.Add(9*time.Hour)) {
		t.Fatal("horario laboral incorrecto")
	}
	night := Schedule{Start: "22:00", End: "06:00"}
	if !night.Active(time.Date(2026, 1, 1, 23, 0, 0, 0, time.UTC)) ||
		!night.Active(time.Date(2026, 1, 1, 5, 0, 0, 0, time.UTC)) ||
		night.Active(time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)) {
		t.Fatal("ventana nocturna incorrecta")
	}
	if !(Schedule{}).Active(sat) {
		t.Fatal("horario vacio debe estar siempre activo")
	}
}

func TestRandomStaysInRegion(t *testing.T) {
	cfg := Default()
	cfg.Region = &Region{X: 100, Y: 100, W: 50, H: 50}
	cfg.MoveDuration = 0
	d := &fakeDriver{w: 1920, h: 1080}
	e := fastEngine(cfg, d)
	for i := 0; i < 200; i++ {
		p := e.target(d.pos, d.w, d.h)
		if p.X < 100 || p.X >= 150 || p.Y < 100 || p.Y >= 150 {
			t.Fatalf("fuera de region: %v", p)
		}
	}
}

func TestJiggleIsSmallAndMoves(t *testing.T) {
	cfg := Default()
	cfg.Mode = ModeJiggle
	d := &fakeDriver{w: 100, h: 100, pos: Point{50, 50}}
	e := fastEngine(cfg, d)
	for i := 0; i < 200; i++ {
		p := e.target(Point{50, 50}, 100, 100)
		if p == (Point{50, 50}) || dist(p, Point{50, 50}) > 5 {
			t.Fatalf("jiggle invalido: %v", p)
		}
	}
}

func TestRunStopsOnUserActivity(t *testing.T) {
	cfg := Default()
	cfg.MoveDuration = 0
	cfg.StopOnMouse = true
	d := &fakeDriver{w: 1000, h: 1000}
	e := fastEngine(cfg, d)
	n := 0
	e.Sleep = func(ctx context.Context, _ time.Duration) error {
		n++
		if n == 2 { // el usuario mueve el mouse
			d.pos = Point{d.pos.X + 100, d.pos.Y}
		}
		return ctx.Err()
	}
	if err := e.Run(context.Background()); !errors.Is(err, ErrUserActivity) {
		t.Fatalf("esperaba ErrUserActivity, obtuve %v", err)
	}
}

func TestRunCancelAndClick(t *testing.T) {
	cfg := Default()
	cfg.MoveDuration = 0
	cfg.Click = true
	d := &fakeDriver{w: 1000, h: 1000}
	e := fastEngine(cfg, d)
	ctx, cancel := context.WithCancel(context.Background())
	e.Sleep = func(ctx context.Context, _ time.Duration) error {
		if d.clicks >= 3 {
			cancel()
		}
		return ctx.Err()
	}
	if err := e.Run(ctx); err != nil {
		t.Fatalf("cancelacion debe devolver nil, obtuve %v", err)
	}
	if d.clicks < 3 {
		t.Fatalf("clicks=%d", d.clicks)
	}
}

func TestRunTimeLimit(t *testing.T) {
	cfg := Default()
	cfg.MoveDuration = 0
	cfg.RunFor = Duration(50 * time.Millisecond)
	d := &fakeDriver{w: 1000, h: 1000}
	e := fastEngine(cfg, d)
	e.Sleep = sleepCtx
	cfg.MinInterval, cfg.MaxInterval = Duration(10*time.Millisecond), Duration(10*time.Millisecond)
	e.Cfg = cfg
	if err := e.Run(context.Background()); !errors.Is(err, ErrTimeLimit) {
		t.Fatalf("esperaba ErrTimeLimit, obtuve %v", err)
	}
}
