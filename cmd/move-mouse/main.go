package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/henryd24/move-mouse/internal/awake"
	"github.com/henryd24/move-mouse/internal/config"
	"github.com/henryd24/move-mouse/internal/mover"
	"github.com/henryd24/move-mouse/internal/stopkey"
)

// version se inyecta en la compilacion con -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := mover.Default()
	var (
		showVersion = flag.Bool("version", false, "muestra la version y sale")
		cfgPath     = flag.String("config", "", "ruta del archivo TOML (por defecto: "+config.DefaultPath()+")")
		printCfg    = flag.Bool("print-config", false, "imprime la configuracion efectiva en TOML y sale")
		logFile     = flag.String("log", "", "escribe el log en este archivo (ademas de la salida estandar)")
		verbose     = flag.Bool("verbose", false, "registra cada movimiento")
		region      = flag.String("region", "", "limita el area: x,y,ancho,alto")
		days        = flag.String("days", "", "dias activos separados por coma: mon,tue,...")
	)
	flag.StringVar(&cfg.Mode, "mode", cfg.Mode, "modo: random | jiggle (micro movimientos discretos)")
	flag.Func("min", "intervalo minimo entre movimientos (ej. 1s)", durFlag(&cfg.MinInterval))
	flag.Func("max", "intervalo maximo entre movimientos (ej. 5s)", durFlag(&cfg.MaxInterval))
	flag.Func("speed", "duracion de cada movimiento (ej. 500ms)", durFlag(&cfg.MoveDuration))
	flag.Func("duration", "detenerse tras este tiempo (ej. 90m); 0 = sin limite", durFlag(&cfg.RunFor))
	flag.IntVar(&cfg.JigglePixels, "jiggle-px", cfg.JigglePixels, "radio en pixeles del modo jiggle")
	flag.BoolVar(&cfg.Click, "click", cfg.Click, "hace click izquierdo tras cada movimiento")
	flag.BoolVar(&cfg.Scroll, "scroll", cfg.Scroll, "hace un scroll minimo tras cada movimiento")
	flag.BoolVar(&cfg.StopOnMouse, "stop-on-mouse", cfg.StopOnMouse, "se detiene si el usuario mueve el mouse")
	flag.StringVar(&cfg.StopKey, "stop-key", cfg.StopKey, `como detener: "any" (cualquier tecla en la terminal) o "none" (solo Ctrl+C/SIGTERM)`)
	flag.BoolVar(&cfg.KeepAwake, "keep-awake", cfg.KeepAwake, "evita suspension de pantalla/equipo")
	flag.StringVar(&cfg.Schedule.Start, "from", cfg.Schedule.Start, "hora de inicio HH:MM")
	flag.StringVar(&cfg.Schedule.End, "to", cfg.Schedule.End, "hora de fin HH:MM")
	flag.Parse()

	if *showVersion {
		fmt.Println("move-mouse", version)
		return nil
	}

	// Precedencia: defaults < archivo < flags. Los flags ya escribieron sobre cfg,
	// asi que se recuerdan y se reaplican tras cargar el archivo.
	flagCfg := cfg
	set := map[string]bool{}
	flag.Visit(func(f *flag.Flag) { set[f.Name] = true })

	var err error
	if *cfgPath != "" {
		cfg, err = config.Load(*cfgPath, mover.Default())
	} else {
		cfg, err = config.LoadDefault(mover.Default())
	}
	if err != nil {
		return err
	}
	applyFlags(&cfg, flagCfg, set)
	if *region != "" {
		if cfg.Region, err = parseRegion(*region); err != nil {
			return err
		}
	}
	if *days != "" {
		cfg.Schedule.Days = strings.Split(*days, ",")
	}
	if err := cfg.Validate(); err != nil {
		return err
	}

	if *printCfg {
		b, err := config.Marshal(cfg)
		if err != nil {
			return err
		}
		fmt.Print(string(b))
		return nil
	}

	var out io.Writer = os.Stdout
	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return err
		}
		defer f.Close()
		out = io.MultiWriter(os.Stdout, f)
	}
	logger := log.New(out, "", log.LstdFlags)

	if runtime.GOOS == "linux" && os.Getenv("XDG_SESSION_TYPE") == "wayland" {
		logger.Println("aviso: Wayland detectado; el control del mouse puede no funcionar. Usa una sesion X11/Xorg.")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if cfg.KeepAwake {
		if err := awake.Start(ctx); err != nil {
			logger.Println("aviso:", err)
		}
	}
	release, err := stopkey.Watch(ctx, cfg.StopKey, cancel)
	if err != nil {
		logger.Println("aviso:", err)
	}
	defer release()

	switch strings.ToLower(cfg.StopKey) {
	case "any":
		logger.Println("Presiona cualquier tecla (o Ctrl+C) para detener.")
	case "none", "":
		logger.Println("Ejecutando; detén con Ctrl+C / SIGTERM.")
	}

	eng := mover.NewEngine(cfg, mover.Robot{})
	if *verbose {
		eng.Logf = logger.Printf
	}
	err = eng.Run(ctx)
	switch {
	case errors.Is(err, mover.ErrUserActivity), errors.Is(err, mover.ErrTimeLimit):
		logger.Println(err)
		return nil
	case err == nil:
		logger.Println("Detenido.")
	}
	return err
}

func durFlag(dst *mover.Duration) func(string) error {
	return func(s string) error {
		d, err := time.ParseDuration(s)
		if err != nil {
			return err
		}
		*dst = mover.Duration(d)
		return nil
	}
}

func parseRegion(s string) (*mover.Region, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 4 {
		return nil, fmt.Errorf("region invalida %q: usa x,y,ancho,alto", s)
	}
	v := make([]int, 4)
	for i, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return nil, fmt.Errorf("region invalida %q: %w", s, err)
		}
		v[i] = n
	}
	return &mover.Region{X: v[0], Y: v[1], W: v[2], H: v[3]}, nil
}

// applyFlags copia a dst solo los valores de src cuyos flags se pasaron explicitamente.
func applyFlags(dst *mover.Config, src mover.Config, set map[string]bool) {
	if set["mode"] {
		dst.Mode = src.Mode
	}
	if set["min"] {
		dst.MinInterval = src.MinInterval
	}
	if set["max"] {
		dst.MaxInterval = src.MaxInterval
	}
	if set["speed"] {
		dst.MoveDuration = src.MoveDuration
	}
	if set["duration"] {
		dst.RunFor = src.RunFor
	}
	if set["jiggle-px"] {
		dst.JigglePixels = src.JigglePixels
	}
	if set["click"] {
		dst.Click = src.Click
	}
	if set["scroll"] {
		dst.Scroll = src.Scroll
	}
	if set["stop-on-mouse"] {
		dst.StopOnMouse = src.StopOnMouse
	}
	if set["stop-key"] {
		dst.StopKey = src.StopKey
	}
	if set["keep-awake"] {
		dst.KeepAwake = src.KeepAwake
	}
	if set["from"] {
		dst.Schedule.Start = src.Schedule.Start
	}
	if set["to"] {
		dst.Schedule.End = src.Schedule.End
	}
}
