package rungroup

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
)

// mockSignalChannel для имитации signal.Notify
func mockSignalChannel(sigCh chan<- os.Signal, sig os.Signal) {
	go func() {
		time.Sleep(100 * time.Millisecond)
		sigCh <- sig
	}()
}

func TestSignalHandlerActors_HandleSignal(t *testing.T) {
	// Create test signal
	testSignal := os.Interrupt

	// Redefine signal.Notify for test
	origNotify := notifyFunc
	defer func() { notifyFunc = origNotify }()
	notifyFunc = func(c chan<- os.Signal, sig ...os.Signal) {
		go mockSignalChannel(c, testSignal)
	}

	// Create handler
	handleSignal, _ := SignalHandlerActors(testSignal)

	// Checking that the correct error is returned when receiving a signal
	err := handleSignal()
	if err == nil || err.Error() != "received signal interrupt" {
		t.Fatalf("expected error 'received signal interrupt', got %v", err)
	}
}

func TestSignalHandlerActors_ContextCancel(t *testing.T) {
	// create handler
	handleSignal, cancel := SignalHandlerActors()

	// Cancel actor
	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = cancel()
	}()

	// checking that the correct error after canceling
	err := handleSignal()
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected error context.Canceled, got %v", err)
	}
}
