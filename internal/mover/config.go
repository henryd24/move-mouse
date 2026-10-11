// Package mover contiene la logica de movimiento del mouse, independiente del SO.
package mover

import (
	"fmt"
	"strings"
	"time"
)

// Duration es un time.Duration que se (de)serializa como texto ("1s", "5m").
type Duration time.Duration

func (d Duration) D() time.Duration { return time.Duration(d) }

func (d Duration) MarshalText() ([]byte, error) { return []byte(time.Duration(d).String()), nil }

func (d *Duration) UnmarshalText(b []byte) error {
	v, err := time.ParseDuration(strings.TrimSpace(string(b)))
	if err != nil {
		return err
	}
	*d = Duration(v)
	return nil
}

const (
	ModeRandom = "random" // salta a posiciones aleatorias de la pantalla
	ModeJiggle = "jiggle" // micro movimientos discretos alrededor de la posicion actual
)

// Region limita el area de pantalla usada en modo random.
type Region struct {
	X int `toml:"x"`
	Y int `toml:"y"`
	W int `toml:"w"`
	H int `toml:"h"`
}

// Schedule restringe la ejecucion a una ventana horaria (formato HH:MM) y dias.
type Schedule struct {
	Start string   `toml:"start"`
	End   string   `toml:"end"`
	Days  []string `toml:"days"` // mon,tue,wed,thu,fri,sat,sun (vacio = todos)
}

// Config agrupa todas las opciones del programa.
type Config struct {
	Mode         string   `toml:"mode"`
	MinInterval  Duration `toml:"min_interval"`
	MaxInterval  Duration `toml:"max_interval"`
	MoveDuration Duration `toml:"move_duration"`
	JigglePixels int      `toml:"jiggle_pixels"`
	Region       *Region  `toml:"region"`
	RunFor       Duration `toml:"run_for"`
	Click        bool     `toml:"click"`
	Scroll       bool     `toml:"scroll"`
	StopOnMouse  bool     `toml:"stop_on_mouse"`
	StopKey      string   `toml:"stop_key"` // "any" o "none"
	KeepAwake    bool     `toml:"keep_awake"`
	Schedule     Schedule `toml:"schedule"`
}

// Default devuelve la configuracion por defecto (equivalente al comportamiento original).
func Default() Config {
	return Config{
		Mode:         ModeRandom,
		MinInterval:  Duration(1 * time.Second),
		MaxInterval:  Duration(5 * time.Second),
		MoveDuration: Duration(500 * time.Millisecond),
		JigglePixels: 3,
		StopKey:      "any",
	}
}

// Validate comprueba que la configuracion sea coherente.
func (c Config) Validate() error {
	if c.Mode != ModeRandom && c.Mode != ModeJiggle {
		return fmt.Errorf("modo invalido %q (usa %s o %s)", c.Mode, ModeRandom, ModeJiggle)
	}
	if c.MinInterval < 0 || c.MaxInterval < c.MinInterval {
		return fmt.Errorf("intervalos invalidos: min=%s max=%s", c.MinInterval.D(), c.MaxInterval.D())
	}
	if c.MoveDuration < 0 || c.RunFor < 0 {
		return fmt.Errorf("las duraciones no pueden ser negativas")
	}
	if c.JigglePixels < 1 {
		return fmt.Errorf("jiggle_pixels debe ser >= 1")
	}
	if r := c.Region; r != nil && (r.W <= 0 || r.H <= 0 || r.X < 0 || r.Y < 0) {
		return fmt.Errorf("region invalida: %+v", *r)
	}
	if k := strings.ToLower(c.StopKey); k != "" && k != "any" && k != "none" {
		return fmt.Errorf("stop_key invalido %q (usa any o none)", c.StopKey)
	}
	return c.Schedule.Validate()
}
