// Package hotkey implementa la deteccion de combinaciones de teclas a partir de
// eventos de bajo nivel (down/up), sin depender de ninguna libreria nativa.
package hotkey

import (
	"fmt"
	"strings"
)

// Lookup resuelve el nombre de una tecla ("ctrl", "m", "f1") a su codigo.
type Lookup func(name string) (uint16, bool)

// Variantes derecha de los modificadores (la tabla base solo trae la izquierda).
var altCodes = map[uint16]uint16{
	29:   3613, // ctrl
	42:   54,   // shift
	56:   3640, // alt
	3675: 3676, // cmd/meta
}

// Combo es una lista de teclas requeridas; cada una admite varios codigos equivalentes.
type Combo [][]uint16

// Parse convierte "ctrl+shift+m" en un Combo.
func Parse(spec string, lookup Lookup) (Combo, error) {
	var combo Combo
	for _, name := range strings.Split(strings.ToLower(spec), "+") {
		name = strings.TrimSpace(name)
		if name == "" {
			return nil, fmt.Errorf("hotkey invalido %q", spec)
		}
		code, ok := lookup(name)
		if !ok {
			return nil, fmt.Errorf("tecla desconocida %q en %q", name, spec)
		}
		group := []uint16{code}
		if alt, ok := altCodes[code]; ok {
			group = append(group, alt)
		}
		combo = append(combo, group)
	}
	return combo, nil
}

// Matcher detecta cuando se completa un Combo.
type Matcher struct {
	combo   Combo
	pressed map[uint16]bool
}

func NewMatcher(c Combo) *Matcher {
	return &Matcher{combo: c, pressed: map[uint16]bool{}}
}

// Feed procesa un evento y devuelve true cuando la combinacion acaba de completarse
// (se pulsa la ultima tecla con las demas ya presionadas). Las repeticiones no disparan.
func (m *Matcher) Feed(code uint16, down bool) bool {
	if len(m.combo) == 0 {
		return false
	}
	wasDown := m.pressed[code]
	m.pressed[code] = down
	if !down || wasDown {
		return false
	}
	involved := false
	for _, group := range m.combo {
		held := false
		for _, c := range group {
			if c == code {
				involved = true
			}
			if m.pressed[c] {
				held = true
			}
		}
		if !held {
			return false
		}
	}
	return involved
}
