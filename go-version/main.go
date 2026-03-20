package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/go-vgo/robotgo"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	stop := make(chan struct{})
	go moveMouse(stop)

	if err := keyboard.Open(); err != nil {
		fmt.Printf("No se pudo inicializar teclado: %v\n", err)
		close(stop)
		return
	}
	defer keyboard.Close()

	fmt.Println("Presiona cualquier tecla para detener el movimiento del mouse...")
	if _, _, err := keyboard.GetKey(); err != nil {
		fmt.Printf("No se pudo leer una tecla: %v\n", err)
	} else {
		fmt.Println("Tecla presionada. Saliendo del programa.")
	}

	close(stop)
	// Pequena espera para que la goroutine termine sin cortar salida por cierre inmediato.
	time.Sleep(100 * time.Millisecond)
}

func moveMouse(stop <-chan struct{}) {
	for {
		select {
		case <-stop:
			return
		default:
		}

		width, height := robotgo.GetScreenSize()
		x := rand.Intn(width)
		y := rand.Intn(height)

		robotgo.MoveSmooth(x, y, 0.5, 0.5)

		wait := time.Duration(1+rand.Intn(5)) * time.Second
		select {
		case <-stop:
			return
		case <-time.After(wait):
		}
	}
}
