package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/henryd24/move-mouse/internal/mover"
)

func TestLoadOverridesAndKeepsDefaults(t *testing.T) {
	p := filepath.Join(t.TempDir(), "c.toml")
	err := os.WriteFile(p, []byte(`
mode = "jiggle"
min_interval = "10s"
max_interval = "30s"
[schedule]
start = "09:00"
end = "18:00"
days = ["mon", "tue"]
[region]
x = 1
y = 2
w = 3
h = 4
`), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(p, mover.Default())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Mode != "jiggle" || cfg.MinInterval.D() != 10*time.Second || cfg.MoveDuration.D() != 500*time.Millisecond {
		t.Fatalf("config inesperada: %+v", cfg)
	}
	if cfg.Region == nil || cfg.Region.H != 4 || cfg.Schedule.Days[1] != "tue" {
		t.Fatalf("config inesperada: %+v", cfg)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestRoundTripAndUnknownField(t *testing.T) {
	p := filepath.Join(t.TempDir(), "c.toml")
	if err := Save(p, mover.Default()); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(p, mover.Config{})
	if err != nil || cfg.MaxInterval.D() != 5*time.Second {
		t.Fatalf("round trip: %v %+v", err, cfg)
	}
	_ = os.WriteFile(p, []byte("bogus = 1\n"), 0o644)
	if _, err := Load(p, mover.Default()); err == nil {
		t.Fatal("deberia rechazar campos desconocidos")
	}
}
