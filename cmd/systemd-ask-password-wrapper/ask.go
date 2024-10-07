package main

import (
	_ "unsafe"

	"gopkg.in/ini.v1"
)

type ask struct {
	pid    int
	socket string
}

func parseAskFile(name string) (ask, error) {
	cfg, err := ini.Load(name)
	if err != nil {
		return ask{}, err
	}

	v, err := cfg.Section("Ask").GetKey("PID")
	if err != nil {
		return ask{}, err
	}
	pid, err := v.Int()
	if err != nil {
		return ask{}, err
	}

	v, err = cfg.Section("Ask").GetKey("Socket")
	if err != nil {
		return ask{}, err
	}
	socket := v.String()

	return ask{
		pid:    pid,
		socket: socket,
	}, nil
}
