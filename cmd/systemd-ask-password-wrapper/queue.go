package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type workqueue interface {
	TryQueue(string) error
	Cancel(string)
}

type serialqueue struct {
	jobs      chan job
	cancelmap map[string]context.CancelFunc
	mux       sync.Mutex
}

func newSerialqueue(bufsize int) *serialqueue {
	return &serialqueue{
		jobs:      make(chan job, bufsize),
		cancelmap: make(map[string]context.CancelFunc),
	}
}

var errJobQueueFull = errors.New("job queue is full")

func (sq *serialqueue) TryQueue(name string) error {
	fmt.Println("tryqueue", name)

	ctxJob, cancel := context.WithCancel(context.Background())
	j := job{
		name:   name,
		ctx:    ctxJob,
		cancel: cancel,
	}

	select {
	case sq.jobs <- j:
		sq.mux.Lock()
		defer sq.mux.Unlock()
		sq.cancelmap[j.name] = j.cancel
		return nil
	default:
		return fmt.Errorf("%w", errJobQueueFull)
	}
}

func (sq *serialqueue) Cancel(name string) {
	fmt.Println("cancel", name)
	sq.mux.Lock()
	defer sq.mux.Unlock()

	cancel, ok := sq.cancelmap[name]
	if !ok {
		return
	}
	cancel()
	delete(sq.cancelmap, name)
}

func ignoreCtxCancelled(err error) error {
	if !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

type worker func(job job) error

func (sq *serialqueue) Work(w worker) error {
	for job := range sq.jobs {
		fmt.Println("working on job", job.name)
		if err := w(job); ignoreCtxCancelled(err) != nil {
			sq.cleanup(job.name)
			return fmt.Errorf("worker errored: %w", err)
		}
		sq.cleanup(job.name)

	}
	return nil
}

func (sq *serialqueue) cleanup(name string) {
	sq.mux.Lock()
	defer sq.mux.Unlock()
	delete(sq.cancelmap, name)
}
