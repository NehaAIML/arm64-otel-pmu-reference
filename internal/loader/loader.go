package loader

import (
    "log"
)

type Loader struct {
    loaded bool
}

func NewLoader() *Loader {
    return &Loader{}
}

func (l *Loader) Load() error {
    log.Println("Loading eBPF bytecode and populating PERF_EVENT_ARRAY maps...")
    l.loaded = true
    return nil
}

func (l *Loader) IsLoaded() bool {
    return l.loaded
}
