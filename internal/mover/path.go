package mover

import (
	"math"
	"math/rand"
)

// Point es una posicion de pantalla.
type Point struct{ X, Y int }

func dist(a, b Point) float64 { return math.Hypot(float64(a.X-b.X), float64(a.Y-b.Y)) }

// BezierPath genera una trayectoria curva (Bezier cubica con puntos de control
// aleatorios y suavizado ease-in-out) de from a to, con n pasos. El ultimo punto es siempre to.
func BezierPath(from, to Point, n int, rng *rand.Rand) []Point {
	if n < 1 {
		n = 1
	}
	d := dist(from, to)
	spread := d * 0.25
	ctrl := func(t float64) (float64, float64) {
		return float64(from.X) + (float64(to.X-from.X))*t + (rng.Float64()*2-1)*spread,
			float64(from.Y) + (float64(to.Y-from.Y))*t + (rng.Float64()*2-1)*spread
	}
	x1, y1 := ctrl(0.33)
	x2, y2 := ctrl(0.66)

	pts := make([]Point, 0, n)
	for i := 1; i <= n; i++ {
		t := float64(i) / float64(n)
		t = t * t * (3 - 2*t) // ease-in-out
		u := 1 - t
		x := u*u*u*float64(from.X) + 3*u*u*t*x1 + 3*u*t*t*x2 + t*t*t*float64(to.X)
		y := u*u*u*float64(from.Y) + 3*u*u*t*y1 + 3*u*t*t*y2 + t*t*t*float64(to.Y)
		pts = append(pts, Point{int(math.Round(x)), int(math.Round(y))})
	}
	pts[len(pts)-1] = to
	return pts
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
