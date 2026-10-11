package main

import (
	"sync"

	hook "github.com/robotn/gohook"
)

var (
	namesOnce sync.Once
	names     map[uint16]string
)

// keyName devuelve un nombre legible (el mas corto) para un codigo de tecla.
func keyName(code uint16) string {
	namesOnce.Do(func() {
		names = map[uint16]string{}
		for n, c := range hook.Keycode {
			if cur, ok := names[c]; !ok || len(n) < len(cur) {
				names[c] = n
			}
		}
	})
	return names[code]
}

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
