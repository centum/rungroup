package rungroup

import (
	"context"
	"fmt"
	"os"
	"os/signal"
)

var notifyFunc = signal.Notify

func SignalHandlerActors(sig ...os.Signal) (func() error, func() error) {
	ctx, cancel := context.WithCancel(context.Background())
	return func() error {
			c := make(chan os.Signal, 1)
			notifyFunc(c, sig...)
			select {
			case sigTerm := <-c:
				return fmt.Errorf("received signal %s", sigTerm)
			case <-ctx.Done():
				return ctx.Err()
			}
		}, func() error {
			cancel()
			return nil
		}
}
