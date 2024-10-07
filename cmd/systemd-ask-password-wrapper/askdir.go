package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/illarion/gonotify/v2"
)

const askDirectory = "/run/systemd/ask-password"

func isAskFileAbsPath(name string) bool {
	return strings.HasPrefix(name, filepath.Join(askDirectory, "ask."))
}

func isAskFile(name string) bool {
	return strings.HasPrefix(name, "ask.")
}

func list(queue workqueue) error {
	files, err := os.ReadDir(askDirectory)
	if err != nil {
		return fmt.Errorf("failed to read dir: %w", err)
	}

	for _, file := range files {
		if !isAskFile(file.Name()) {
			continue
		}

		if err := queue.TryQueue(filepath.Join(askDirectory, file.Name())); err != nil {
			return fmt.Errorf("failed to queue file %s: %w", file.Name(), err)
		}
	}
	return nil
}

func watch(ctx context.Context, queue workqueue) error {
	w, err := gonotify.NewDirWatcher(ctx, gonotify.IN_MOVED_TO|gonotify.IN_CLOSE_WRITE|gonotify.IN_DELETE, askDirectory)
	if err != nil {
		return fmt.Errorf("failed to create dir watcher: %w", err)
	}

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case event := <-w.C:
			fmt.Println("event", event)
			if !isAskFileAbsPath(event.Name) {
				fmt.Println(">> skipping non ask file")
				continue
			}
			if event.Is(gonotify.IN_MOVED_TO) {
				if err := queue.TryQueue(event.Name); err != nil {
					return fmt.Errorf("failed to queue event: %w", err)
				}
				continue
			}
			queue.Cancel(event.Name)
		}
	}

	return nil
}
