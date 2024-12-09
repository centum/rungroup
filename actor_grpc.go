package rungroup

import (
	"context"
	"net"
	"time"
)

type GrpcServer interface {
	Serve(lis net.Listener) error
	GracefulStop()
	Stop()
}

func GrpcServerActors(
	grpcServer GrpcServer,
	grpcListener net.Listener,
	opts ...func(server *wrapServer),
) (func() error, func() error) {
	w := newWrapHttpServer(5 * time.Second)
	for _, opt := range opts {
		opt(w)
	}
	return func() error {
			return grpcServer.Serve(grpcListener)
		}, func() error {
			ctx, cancelStop := context.WithCancel(context.Background())
			defer cancelStop()
			go func() {
				grpcServer.GracefulStop()
				cancelStop()
			}()
			t := time.NewTimer(w.shutdownTimeout)
			defer t.Stop()
			select {
			case <-t.C:
				grpcServer.Stop()
			case <-ctx.Done():
			}
			return nil
		}
}
