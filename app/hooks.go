package main

import (
	hook "github.com/robotn/gohook"
)

func keyLookup(name string) (uint16, bool) {
	c, ok := hook.Keycode[name]
	return c, ok
}

// watchKeyboard escucha el teclado a nivel global (requiere sesion grafica;
// en macOS pide permiso de Accesibilidad).
func (a *App) watchKeyboard() {
	for ev := range hook.Start() {
		switch ev.Kind {
		case hook.KeyDown:
			a.onKey(ev.Keycode, true)
		case hook.KeyUp:
			a.onKey(ev.Keycode, false)
		}
	}
}
