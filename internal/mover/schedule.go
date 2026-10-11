package mover

import (
	"fmt"
	"strings"
	"time"
)

var dayNames = map[string]time.Weekday{
	"sun": time.Sunday, "mon": time.Monday, "tue": time.Tuesday, "wed": time.Wednesday,
	"thu": time.Thursday, "fri": time.Friday, "sat": time.Saturday,
}

func parseClock(s string) (int, error) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return 0, fmt.Errorf("hora invalida %q (usa HH:MM)", s)
	}
	return t.Hour()*60 + t.Minute(), nil
}

// Validate comprueba el formato del horario.
func (s Schedule) Validate() error {
	if (s.Start == "") != (s.End == "") {
		return fmt.Errorf("schedule requiere start y end juntos")
	}
	if s.Start != "" {
		if _, err := parseClock(s.Start); err != nil {
			return err
		}
		if _, err := parseClock(s.End); err != nil {
			return err
		}
	}
	for _, d := range s.Days {
		if _, ok := dayNames[strings.ToLower(d)]; !ok {
			return fmt.Errorf("dia invalido %q", d)
		}
	}
	return nil
}

// Active indica si t cae dentro del horario. Un horario vacio esta siempre activo.
// Soporta ventanas que cruzan la medianoche (ej. 22:00-06:00).
func (s Schedule) Active(t time.Time) bool {
	if len(s.Days) > 0 {
		ok := false
		for _, d := range s.Days {
			if dayNames[strings.ToLower(d)] == t.Weekday() {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	if s.Start == "" {
		return true
	}
	start, _ := parseClock(s.Start)
	end, _ := parseClock(s.End)
	now := t.Hour()*60 + t.Minute()
	if start <= end {
		return now >= start && now < end
	}
	return now >= start || now < end
}
