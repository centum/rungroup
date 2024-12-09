package rungroup

import (
	"context"
	"net/http"
	"time"
)

type wrapServer struct {
	shutdownTimeout time.Duration
}

var newWrapHttpServer = func(shutdownTimeout time.Duration) *wrapServer {
	return &wrapServer{shutdownTimeout: shutdownTimeout}
}

func WithShutdownTimeout(shutdownTimeout time.Duration) func(w *wrapServer) {
	return func(w *wrapServer) {
		w.shutdownTimeout = shutdownTimeout
	}
}

func HttpServerActors(httpServer *http.Server, opts ...func(server *wrapServer)) (func() error, func() error) {
	w := newWrapHttpServer(5 * time.Second)
	for _, opt := range opts {
		opt(w)
	}
	return func() error {
			return httpServer.ListenAndServe()
		}, func() error {
			ctx := context.Background()
			if w.shutdownTimeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, w.shutdownTimeout)
				defer cancel()
			}
			return httpServer.Shutdown(ctx)
		}
}
