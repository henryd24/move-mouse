package hotkey

import "testing"

var table = map[string]uint16{"ctrl": 29, "shift": 42, "m": 50}

func lookup(n string) (uint16, bool) { c, ok := table[n]; return c, ok }

func TestParseErrors(t *testing.T) {
	for _, s := range []string{"", "ctrl+", "ctrl+zz"} {
		if _, err := Parse(s, lookup); err == nil {
			t.Errorf("%q deberia fallar", s)
		}
	}
}

func TestMatcher(t *testing.T) {
	c, err := Parse("Ctrl+Shift+M", lookup)
	if err != nil {
		t.Fatal(err)
	}
	m := NewMatcher(c)
	if m.Feed(29, true) || m.Feed(42, true) {
		t.Fatal("no debe disparar con solo modificadores")
	}
	if !m.Feed(50, true) {
		t.Fatal("debe disparar al pulsar la ultima tecla")
	}
	if m.Feed(50, true) {
		t.Fatal("la repeticion no debe disparar")
	}
	m.Feed(50, false)
	if !m.Feed(50, true) {
		t.Fatal("debe volver a disparar tras soltar y pulsar")
	}
}

func TestMatcherRightModifierAndMissing(t *testing.T) {
	c, _ := Parse("ctrl+m", lookup)
	m := NewMatcher(c)
	if m.Feed(50, true) {
		t.Fatal("falta ctrl")
	}
	m.Feed(50, false)
	m.Feed(3613, true) // ctrl derecho
	if !m.Feed(50, true) {
		t.Fatal("ctrl derecho debe valer")
	}
}
