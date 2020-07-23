package os

import (
	"context"
	"os"
	"os/signal"
)

// WithSignal executes fn within a context that is canceled when one of specified signals is received from the OS.
// It's useful for wrapping some functions that should be canceled when the program receives e.g. SIGINT or SIGTERM.
func WithSignal(ctx context.Context, fn func(ctx context.Context) error, signals ...os.Signal) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		defer close(done)
		done <- fn(ctx)
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, signals...)
	defer signal.Stop(sig)

	select {
	case err := <-done:
		return err
	case <-sig:
		cancel()
	}

	// Make sure we wait for the handler to exit, if it hasn't yet.
	return <-done
}
