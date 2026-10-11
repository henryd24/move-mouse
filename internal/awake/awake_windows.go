//go:build windows

package awake

import (
	"context"
	"runtime"
	"syscall"
)

const (
	esContinuous      = 0x80000000
	esSystemRequired  = 0x00000001
	esDisplayRequired = 0x00000002
)

// Start impide la suspension hasta que ctx termine (SetThreadExecutionState).
func Start(ctx context.Context) error {
	proc := syscall.NewLazyDLL("kernel32.dll").NewProc("SetThreadExecutionState")
	ready := make(chan struct{})
	go func() {
		// El estado de ejecucion es por hilo: se fija el hilo del SO.
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		proc.Call(uintptr(esContinuous | esSystemRequired | esDisplayRequired))
		close(ready)
		<-ctx.Done()
		proc.Call(uintptr(esContinuous))
	}()
	<-ready
	return nil
}
