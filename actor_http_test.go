package rungroup

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestHttpServerActors(t *testing.T) {
	// Create mock handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))  //nolint:errcheck
	})

	// Create server with use address :11081
	addr := ":11081"
	startServer, stopServer := HttpServerActors(&http.Server{
		Addr:    addr,
		Handler: handler,
	})

	// Testing start server
	go func() {  //nolint:staticcheck
		if err := startServer(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf("failed to start server: %v", err)  //nolint
		}
	}()
	time.Sleep(100 * time.Millisecond) // We are waiting for the server to start exactly

	// Send a request to the server
	resp, err := http.Get("http://localhost" + addr)
	if err != nil {
		t.Fatalf("failed to send GET request: %v", err)
	}
	defer resp.Body.Close()

	// Checking the response from the server
	if resp.StatusCode != http.StatusOK {
		t.Errorf("unexpected status code: got %v, want %v", resp.StatusCode, http.StatusOK)
	}

	// Testing the stopping server
	if err := stopServer(); err != nil {
		t.Fatalf("failed to stop server: %v", err)
	}
}

func TestHttpServerActors_WithOptions(t *testing.T) {
	// Prepare a mocked constructor
	whs := &wrapServer{}

	origNewWrapHttpServer := newWrapHttpServer
	defer func() {
		newWrapHttpServer = origNewWrapHttpServer
	}()
	newWrapHttpServer = func(shutdownTimeout time.Duration) *wrapServer {
		whs.shutdownTimeout = shutdownTimeout
		return whs
	}

	// Prepare a mocked handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World with Options!"))  //nolint:errcheck
	})

	// Create server with use address :11081
	addr := ":11081"

	// The custom timeout for stopping
	shutdownTimeout := 2 * time.Minute

	// Create the server with options
	startServer, stopServer := HttpServerActors(&http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	},
		WithShutdownTimeout(shutdownTimeout),
	)

	// Testing start server
	go func() {  //nolint:staticcheck
		if err := startServer(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf("failed to start server: %v", err)  //nolint
		}
	}()
	time.Sleep(100 * time.Millisecond) // Ждем, чтобы сервер точно стартовал

	// Send a request to the server
	resp, err := http.Get("http://localhost" + addr)
	if err != nil {
		t.Fatalf("failed to send GET request: %v", err)
	}
	defer resp.Body.Close()

	// Checking the response from the server
	if resp.StatusCode != http.StatusOK {
		t.Errorf("unexpected status code: got %v, want %v", resp.StatusCode, http.StatusOK)
	}

	// Testing stopping server
	if err := stopServer(); err != nil {
		t.Fatalf("failed to stop server: %v", err)
	}

	// Testing shutdownTimeout options
	if whs.shutdownTimeout != shutdownTimeout {
		t.Fatalf("unexpected shutdown timeout: got %v, want %v", whs.shutdownTimeout, shutdownTimeout)
	}

}
