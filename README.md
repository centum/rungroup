# rungroup

[![Apache 2 licensed](https://img.shields.io/badge/license-Apache2-blue.svg)](https://raw.githubusercontent.com/centum/rungroup/refs/heads/master/LICENSE)
[![test](https://github.com/centum/rungroup/actions/workflows/test.yml/badge.svg?branch=master&event=push)](https://github.com/centum/rungroup/actions/workflows/test.yml)

`rungroup` is inspired by the [run](https://github.com/oklog/run) package.

`rungroup` is a mechanism for launching a group of goroutines and managing their life cycle based on [errgroup](golang.org/x/sync/errgroup)

Create a RunGroup, and then add actors to it.
Actors are defined as a pair of functions: an **execute** function, which should run synchronously; and an **interrupt** function, which, when invoked, should cause the execute function to return. 
Finally, invoke Wait, which waits until the first actor exits, invokes the interrupt functions, and finally returns control to the caller only once all actors have returned.
This general-purpose API allows callers to model pretty much any runnable task, and achieve well-defined lifecycle semantics for the group.
Can be used to gracefully shutdown a service.

## Examples

### Create the rungroup

```go
g := rungroup.NewRunGroup()
```

### Add actor with interrupt function

```go
ctx, cancel := context.WithCancel(context.Background())
g.Add(func() error {
	... main actor processed...
	return nil
}, func() error {
	cancel() // terminate main process
	return nil
})
```

### Add actor with cancellation context

```go
g.AddCtx(func(ctx context.Context) error {
	for {
		select {
			case <-ctx.Done():
				return
			default:
			   ...main process...
		}
	}
})
```

### Add actor with net.Listener

```go
ln, _ := net.Listen("tcp", ":8080")
g.Add(func() error {
	return http.Serve(ln, nil)
}, func() error {
	ln.Close()
})
```

### Add actor with io.ReadCloser

```go
var conn io.ReadCloser = ...
g.Add(func() error {
	s := bufio.NewScanner(conn)
	for s.Scan() {
		println(s.Text())
	}
	return s.Err()
}, func(error) {
	conn.Close()
})
```

### Add http actor

```go
httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, World!"))
})

g.Add(rungroup.HttpServerActors(&http.Server{
	Addr:    ":8000",
	Handler: httpHandler,
}))
```


### Add grpc actor

```go
grpcServer := grpc.NewServer()
listener, err := net.Listen("tcp", addr)
g.Add(GrpcServerActors(grpcServer, listener))
```
### Add signal listener actor

```go
g.Add(rungroup.SignalHandlerActors(syscall.SIGINT, syscall.SIGTERM))
```

### Run the group of actors
```go
if err := g.Run(); err != nil {
	print("terminate with error: %s", err.Error())
}
```
