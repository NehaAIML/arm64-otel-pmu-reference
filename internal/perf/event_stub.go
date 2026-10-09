//go:build !linux

package perf

type Event struct{}

func OpenPerCPU(config, freq uint64) ([]*Event, error) { return nil, nil }
func (e *Event) Close()                                {}
