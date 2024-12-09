package rungroup

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestGroupZeroActor(t *testing.T) {
	g := NewRunGroup()
	res := make(chan error)
	go func() { res <- g.Run() }()
	select {
	case err := <-res:
		if err != nil {
			t.Errorf("%v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout")
	}
}

func TestGroupOneActor(t *testing.T) {
	myError := errors.New("foobar")
	g := NewRunGroup()
	g.Add(func() error { return myError }, func() error { return nil })
	res := make(chan error)
	go func() { res <- g.Run() }()
	select {
	case err := <-res:
		if want, have := myError, err; want != have {
			t.Errorf("want %v, have %v", want, have)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout")
	}
}

func TestGroupManyActors(t *testing.T) {
	interrupt := errors.New("interrupt")
	g := NewRunGroup()
	g.Add(func() error { return interrupt }, func() error { return nil })
	cancel := make(chan struct{})
	g.Add(func() error { <-cancel; return nil }, func() error {
		close(cancel)
		return nil
	})
	res := make(chan error)
	go func() { res <- g.Run() }()
	select {
	case err := <-res:
		if want, have := interrupt, err; want != have {
			t.Errorf("want %v, have %v", want, have)
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("timeout")
	}
}

func TestGroupOneActorWithCancelContext(t *testing.T) {
	interrupt := errors.New("interrupt")
	g := NewRunGroup()
	terminate1actor := make(chan struct{}, 1)
	defer close(terminate1actor)
	g.AddCtx(func(ctx context.Context) error {
		<-ctx.Done()
		terminate1actor <- struct{}{}
		return nil
	})
	cancel2actor := make(chan struct{})
	defer close(cancel2actor)
	g.Add(func() error { <-cancel2actor; return interrupt }, func() error { return nil })

	res := make(chan error)
	defer close(res)
	go func() { res <- g.Run() }()

	cancel2actor <- struct{}{} // terminate second actor

	select {
	case err := <-res:
		if want, have := interrupt, err; want != have {
			t.Errorf("want %v, have %v", want, have)
		}
		select {
		case <-terminate1actor:
		case <-time.After(100 * time.Millisecond):
			t.Errorf("timeout 1")
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("timeout 2")
	}
}
