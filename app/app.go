package main

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/henryd24/move-mouse/internal/awake"
	"github.com/henryd24/move-mouse/internal/config"
	"github.com/henryd24/move-mouse/internal/mover"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App expone el motor de movimiento al frontend.
type App struct {
	ctx context.Context

	mu      sync.Mutex
	cancel  context.CancelFunc
	running bool
}

func NewApp() *App { return &App{} }

func (a *App) startup(ctx context.Context) { a.ctx = ctx }

func (a *App) shutdown(context.Context) { a.Stop() }

// DefaultConfig devuelve la configuracion inicial (la guardada, o la por defecto).
func (a *App) DefaultConfig() mover.Config {
	cfg, err := config.LoadDefault(mover.Default())
	if err != nil {
		return mover.Default()
	}
	return cfg
}

// Save persiste la configuracion en el archivo del usuario.
func (a *App) Save(cfg mover.Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	return config.Save(config.DefaultPath(), cfg)
}

// Start inicia el movimiento con cfg.
func (a *App) Start(cfg mover.Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.running {
		return errors.New("ya esta en ejecucion")
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel, a.running = cancel, true
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
		a.mu.Unlock()
		msg := "detenido"
		if err != nil {
			msg = err.Error()
		}
		runtime.EventsEmit(a.ctx, "stopped", msg)
	}()
	return nil
}

// Stop detiene el movimiento.
func (a *App) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancel != nil {
		a.cancel()
	}
}

// Running indica si el motor esta activo.
func (a *App) Running() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.running
}
