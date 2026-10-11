// Package config carga la configuracion desde un archivo TOML.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"

	"github.com/henryd24/move-mouse/internal/mover"
)

// DefaultPath devuelve la ruta estandar del archivo de configuracion del usuario.
func DefaultPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "move-mouse", "config.toml")
}

// Load lee path sobre la base de cfg (los valores ausentes en el archivo se conservan).
func Load(path string, cfg mover.Config) (mover.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	dec := toml.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("%s: %w", path, err)
	}
	return cfg, nil
}

// LoadDefault carga DefaultPath si existe; si no existe devuelve cfg sin cambios.
func LoadDefault(cfg mover.Config) (mover.Config, error) {
	p := DefaultPath()
	if p == "" {
		return cfg, nil
	}
	out, err := Load(p, cfg)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	return out, err
}

// Marshal serializa cfg a TOML.
func Marshal(cfg mover.Config) ([]byte, error) { return toml.Marshal(cfg) }

// Save escribe cfg en path creando los directorios necesarios.
func Save(path string, cfg mover.Config) error {
	b, err := Marshal(cfg)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
