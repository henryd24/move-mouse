package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/henryd24/move-mouse/internal/awake"
	"github.com/henryd24/move-mouse/internal/config"
	"github.com/henryd24/move-mouse/internal/hotkey"
	"github.com/henryd24/move-mouse/internal/mover"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// anyKeyGrace evita que las teclas del propio atajo de inicio detengan la sesion.
const anyKeyGrace = time.Second

// App expone el motor de movimiento al frontend.
type App struct {
	ctx context.Context

	mu        sync.Mutex
	cfg       mover.Config
	cancel    context.CancelFunc
	running   bool
	reason    string
	startedAt time.Time
	matcher   *hotkey.Matcher
}

func NewApp() *App { return &App{cfg: mover.Default()} }

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.cfg = a.DefaultConfig()
	_ = a.apply(a.cfg)
	go a.watchKeyboard()
}

func (a *App) shutdown(context.Context) { a.Stop() }

// DefaultConfig devuelve la configuracion inicial (la guardada, o la por defecto).
func (a *App) DefaultConfig() mover.Config {
	cfg, err := config.LoadDefault(mover.Default())
	if err != nil {
		return mover.Default()
	}
	return cfg
}

// apply valida cfg y actualiza el atajo global.
func (a *App) apply(cfg mover.Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	var m *hotkey.Matcher
	if strings.TrimSpace(cfg.Hotkey) != "" {
		combo, err := hotkey.Parse(cfg.Hotkey, keyLookup)
		if err != nil {
			return err
		}
		m = hotkey.NewMatcher(combo)
	}
	a.mu.Lock()
	a.cfg, a.matcher = cfg, m
	a.mu.Unlock()
	return nil
}

// Save valida, aplica (incluido el atajo global) y persiste la configuracion.
func (a *App) Save(cfg mover.Config) error {
	if err := a.apply(cfg); err != nil {
		return err
	}
	return config.Save(config.DefaultPath(), cfg)
}

// Start inicia el movimiento con cfg.
func (a *App) Start(cfg mover.Config) error {
	if err := a.apply(cfg); err != nil {
		return err
	}
	return a.start()
}

func (a *App) start() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.running {
		return errors.New("ya esta en ejecucion")
	}
	cfg := a.cfg
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel, a.running, a.reason, a.startedAt = cancel, true, "", time.Now()
	if cfg.KeepAwake {
		_ = awake.Start(ctx)
	}
	eng := mover.NewEngine(cfg, mover.Robot{})
	eng.Logf = func(f string, args ...any) {
		runtime.EventsEmit(a.ctx, "log", fmt.Sprintf(f, args...))
	}
	go func() {
		err := eng.Run(ctx)
		a.mu.Lock()
		a.running = false
		msg := a.reason
		a.mu.Unlock()
		switch {
		case err != nil:
			msg = err.Error()
		case msg == "":
			msg = "Detenido"
		}
		runtime.EventsEmit(a.ctx, "state", map[string]any{"running": false, "msg": msg})
	}()
	runtime.EventsEmit(a.ctx, "state", map[string]any{"running": true, "msg": "En ejecución"})
	return nil
}

// Stop detiene el movimiento.
func (a *App) Stop() { a.stopWith("Detenido") }

func (a *App) stopWith(reason string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancel != nil && a.running {
		a.reason = reason
		a.cancel()
	}
}

// Running indica si el motor esta activo.
func (a *App) Running() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.running
}

// toggle lo dispara el atajo global.
func (a *App) toggle() {
	if a.Running() {
		a.stopWith("Detenido con el atajo global")
		return
	}
	if err := a.start(); err != nil {
		runtime.EventsEmit(a.ctx, "log", "error: "+err.Error())
	}
}

// onKey procesa un evento global de teclado.
func (a *App) onKey(code uint16, down bool) {
	a.mu.Lock()
	m := a.matcher
	running := a.running
	anyKey := strings.EqualFold(a.cfg.StopKey, "any")
	inGrace := time.Since(a.startedAt) < anyKeyGrace
	a.mu.Unlock()

	if m != nil && m.Feed(code, down) {
		a.toggle()
		return
	}
	if down && running && anyKey && !inGrace {
		a.stopWith("Detenido: se presionó una tecla")
	}
}
