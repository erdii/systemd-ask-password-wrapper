package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	q := newSerialqueue(50)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		if err := list(q); err != nil {
			panic(err)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		if err := watch(ctx, q); err != nil {
			panic(err)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		if err := q.Work(run); err != nil {
			panic(err)
		}
		wg.Done()
	}()

	wg.Wait()
}

type job struct {
	name   string
	cancel context.CancelFunc
	ctx    context.Context
}

func (j *job) isDone() bool {
	select {
	case <-j.ctx.Done():
		return true
	default:
		return false
	}
}

func run(job job) error {
	fmt.Println("run", job.name)
	defer fmt.Println("run completed", job.name)

	// Return early if job was cancelled.
	if job.isDone() {
		return nil
	}

	ask, err := parseAskFile(job.name)
	if err != nil {
		return err
	}

	// Return early if job was cancelled.
	if job.isDone() {
		return nil
	}

	// Return early if signal returns ESRCH.
	if err := syscall.Kill(ask.pid, 0); errors.Is(err, syscall.ESRCH) {
		return nil
	} else if err != nil {
		return err
	}

	// Return early if job was cancelled.
	if job.isDone() {
		return nil
	}

	// Run wrapped command and retrieve passphrase.
	wrapped := exec.CommandContext(job.ctx, os.Args[1], os.Args[2:]...)
	passphrase, err := wrapped.Output()
	if err != nil {
		fmt.Println(fmt.Errorf("calling wrapped command: %w", err))
		// Signal that running the command errored.
		return exec.CommandContext(job.ctx, "/lib/systemd/systemd-reply-password", "0").Run()
	}

	// Return early if job was cancelled.
	if job.isDone() {
		return nil
	}

	// Send passphrase to ask socket.
	replyCtx, replyCancel := context.WithTimeout(context.Background(), time.Second)
	defer replyCancel()
	pkreply := exec.CommandContext(replyCtx, "/lib/systemd/systemd-reply-password", "1", ask.socket)
	pkreply.Stdin = bytes.NewReader(passphrase)
	return pkreply.Run()
}
