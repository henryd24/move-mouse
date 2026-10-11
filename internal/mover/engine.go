package mover

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

// Driver abstrae el control del mouse para poder probar el motor sin SO.
type Driver interface {
	ScreenSize() (w, h int)
	Location() Point
	MoveTo(p Point)
	Click()
	Scroll(dy int)
}

// Errores que terminan Run de forma esperada.
var (
	ErrUserActivity = errors.New("movimiento del mouse detectado: el usuario tomo el control")
	ErrTimeLimit    = errors.New("tiempo limite alcanzado")
)

const (
	userMoveTolerance = 5.0                   // pixeles de diferencia para considerar actividad del usuario
	stepInterval      = 10 * time.Millisecond // tiempo entre pasos de la trayectoria
	scheduleCheck     = 15 * time.Second
)

// Engine ejecuta el bucle de movimiento.
type Engine struct {
	Cfg    Config
	Driver Driver
	Rand   *rand.Rand
	Now    func() time.Time
	Logf   func(format string, args ...any)

	// Sleep es inyectable para pruebas; debe respetar ctx.
	Sleep func(ctx context.Context, d time.Duration) error
}

// NewEngine crea un motor con valores por defecto razonables.
func NewEngine(cfg Config, d Driver) *Engine {
	return &Engine{
		Cfg:    cfg,
		Driver: d,
		Rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
		Now:    time.Now,
		Logf:   func(string, ...any) {},
		Sleep:  sleepCtx,
	}
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// Run bloquea hasta que ctx se cancela o se cumple una condicion de parada.
// Devuelve nil si se cancelo el contexto.
func (e *Engine) Run(ctx context.Context) error {
	if err := e.Cfg.Validate(); err != nil {
		return err
	}
	if e.Cfg.RunFor > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeoutCause(ctx, e.Cfg.RunFor.D(), ErrTimeLimit)
		defer cancel()
	}
	err := e.loop(ctx)
	if cause := context.Cause(ctx); errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		if errors.Is(cause, ErrTimeLimit) {
			return ErrTimeLimit
		}
		return nil
	}
	return err
}

func (e *Engine) loop(ctx context.Context) error {
	last := e.Driver.Location()
	for {
		if !e.Cfg.Schedule.Active(e.Now()) {
			if err := e.Sleep(ctx, scheduleCheck); err != nil {
				return err
			}
			last = e.Driver.Location()
			continue
		}

		if e.Cfg.StopOnMouse && dist(e.Driver.Location(), last) > userMoveTolerance {
			return ErrUserActivity
		}

		if err := e.step(ctx); err != nil {
			return err
		}
		last = e.Driver.Location()

		span := e.Cfg.MaxInterval.D() - e.Cfg.MinInterval.D()
		wait := e.Cfg.MinInterval.D()
		if span > 0 {
			wait += time.Duration(e.Rand.Int63n(int64(span) + 1))
		}
		if err := e.Sleep(ctx, wait); err != nil {
			return err
		}
	}
}

// step realiza una accion completa (mover y, opcionalmente, click/scroll).
func (e *Engine) step(ctx context.Context) error {
	w, h := e.Driver.ScreenSize()
	from := e.Driver.Location()
	to := e.target(from, w, h)

	steps := int(e.Cfg.MoveDuration.D() / stepInterval)
	if e.Cfg.Mode == ModeJiggle {
		steps = 1
	}
	e.Logf("mover %v -> %v", from, to)
	for _, p := range BezierPath(from, to, steps, e.Rand) {
		e.Driver.MoveTo(p)
		if steps > 1 {
			if err := e.Sleep(ctx, stepInterval); err != nil {
				return err
			}
		}
	}
	if e.Cfg.Click {
		e.Driver.Click()
	}
	if e.Cfg.Scroll {
		e.Driver.Scroll(e.Rand.Intn(3) - 1)
	}
	return ctx.Err()
}

func (e *Engine) target(from Point, w, h int) Point {
	if e.Cfg.Mode == ModeJiggle {
		n := e.Cfg.JigglePixels
		dx, dy := e.Rand.Intn(2*n+1)-n, e.Rand.Intn(2*n+1)-n
		if dx == 0 && dy == 0 {
			dx = n
		}
		return Point{clamp(from.X+dx, 0, w-1), clamp(from.Y+dy, 0, h-1)}
	}
	x0, y0, rw, rh := 0, 0, w, h
	if r := e.Cfg.Region; r != nil {
		x0, y0 = clamp(r.X, 0, w-1), clamp(r.Y, 0, h-1)
		rw, rh = clamp(r.W, 1, w-x0), clamp(r.H, 1, h-y0)
	}
	return Point{x0 + e.Rand.Intn(rw), y0 + e.Rand.Intn(rh)}
}
