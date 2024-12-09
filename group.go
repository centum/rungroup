package rungroup

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type actor struct {
	execute   func() error
	interrupt func() error
}

type RunGroup struct {
	*errgroup.Group
	ctx context.Context
	actors []actor
}

// NewRunGroup create actors group
func NewRunGroup() *RunGroup {
	g := &RunGroup{
		actors: make([]actor, 0, 2),
	}
	g.Group, g.ctx = errgroup.WithContext(context.Background())
	return g
}

// Add is function add a new execute and interrupt actor
func (g *RunGroup) Add(execute func() error, interrupt func() error) {
	g.actors = append(g.actors, actor{
		execute: execute,
		interrupt: interrupt,
	})
}

// AddCtx is function add a new execute with cancellation context
func (g *RunGroup) AddCtx(execute func(ctx context.Context) error) {
	g.actors = append(g.actors, actor{
		execute: func() error {
			return execute(g.ctx)
		},
	})
}

// Run is function run all added actors
func (g *RunGroup) Run() error {
	for _, a := range g.actors {
		g.Go(a.execute)
		if a.interrupt != nil {
			g.Go(func() error {
				<-g.ctx.Done()
				return a.interrupt()
			})
		}
	}
	return g.Wait()
}
