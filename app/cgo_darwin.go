//go:build darwin

package main

// Wails (v2.10+) usa UTType en macOS 11+; `go build` plano no enlaza este framework
// (la CLI de wails lo anade por su cuenta).

/*
#cgo darwin LDFLAGS: -framework UniformTypeIdentifiers
*/
import "C"
