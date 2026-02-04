package scheduler

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

type Runtime struct {
	wg sync.WaitGroup
	cancel context.CancelFunc
	log *zap.Logger
}

func NewRuntime(log *zap.Logger) *Runtime {
	return &Runtime{log: log}
}

func (r *Runtime) Start(parent context.Context,components ...func(context.Context)) context.Context {
	ctx,cancel := context.WithCancel(parent)
	r.cancel = cancel

	for _,c := range components {
		r.wg.Add(1)
		go func(fn func(context.Context)){
			defer r.wg.Done()
			fn(ctx)
		}(c)
	}

	return ctx 
}

func (r *Runtime) Stop() {
	r.log.Info("runtime shutting down")
	r.cancel()
	r.wg.Wait()
	r.log.Info("runtime stopped cleanly")
}