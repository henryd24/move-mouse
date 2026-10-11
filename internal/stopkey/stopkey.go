// Package stopkey implementa las formas de detener el programa por teclado.
package stopkey

import (
	"context"
	"fmt"
	"strings"

	"github.com/eiannone/keyboard"
)

// Watch llama a cancel cuando se cumple la condicion indicada:
//   - "any":  cualquier tecla en la terminal.
//   - "none": nunca (solo senales / timeout).
//
// Un hotkey global no se soporta: la libreria de hooks falla al cargar sin servidor grafico.
//
// Devuelve una funcion que libera los recursos.
func Watch(ctx context.Context, spec string, cancel context.CancelFunc) (func(), error) {
	switch strings.ToLower(strings.TrimSpace(spec)) {
	case "", "none":
		return func() {}, nil
	case "any":
		return watchAny(cancel)
	default:
		return func() {}, fmt.Errorf("stop-key %q no soportado (usa any o none)", spec)
	}
}

func watchAny(cancel context.CancelFunc) (func(), error) {
	if err := keyboard.Open(); err != nil {
		return func() {}, fmt.Errorf("no se pudo leer el teclado de la terminal: %w", err)
	}
	go func() {
		if _, _, err := keyboard.GetKey(); err == nil {
			cancel()
		}
	}()
	return func() { _ = keyboard.Close() }, nil
}
