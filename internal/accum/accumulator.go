package accum

import (
    "sync"
)

const MaxScale = 10.0

type Sample struct {
    Counter  uint64
    Enabled  uint64
    Running  uint64
}

type Delta struct {
    Value       uint64
    ScaledValue float64
    Reset       bool
}

type Item struct {
    LastCounter  uint64
    LastEnabled  uint64
    LastRunning  uint64
    ScaledValue  float64
    Initialized  bool
}

type Accumulator struct {
    mu    sync.Mutex
    items map[string]Item
}

func NewAccumulator() *Accumulator {
    return &Accumulator{
        items: make(map[string]Item),
    }
}

func (a *Accumulator) Update(key string, s Sample) Delta {
    a.mu.Lock()
    defer a.mu.Unlock()

    item, exists := a.items[key]
    if !exists {
        a.items[key] = Item{
            LastCounter: s.Counter,
            LastEnabled: s.Enabled,
            LastRunning: s.Running,
            ScaledValue: 0,
            Initialized: true,
        }
        return Delta{Value: 0, ScaledValue: 0, Reset: false}
    }

    var valDelta uint64
    reset := false

    if s.Counter < item.LastCounter {
        reset = true
        valDelta = 0
    } else {
        valDelta = s.Counter - item.LastCounter
    }

    var scale float64 = 1.0
    enabledDelta := s.Enabled - item.LastEnabled
    runningDelta := s.Running - item.LastRunning

    if runningDelta > 0 && enabledDelta >= runningDelta {
        scale = float64(enabledDelta) / float64(runningDelta)
        if scale > MaxScale {
            scale = MaxScale
        }
    }

    scaledVal := float64(valDelta) * scale
    cumScaled := item.ScaledValue + scaledVal

    a.items[key] = Item{
        LastCounter: s.Counter,
        LastEnabled: s.Enabled,
        LastRunning: s.Running,
        ScaledValue: cumScaled,
        Initialized: true,
    }

    return Delta{
        Value:       valDelta,
        ScaledValue: cumScaled,
        Reset:       reset,
    }
}

func (a *Accumulator) DropStale(live map[string]bool) {
    a.mu.Lock()
    defer a.mu.Unlock()

    for k := range a.items {
        if !live[k] {
            delete(a.items, k)
        }
    }
}

func (a *Accumulator) Flush() map[string]float64 {
    a.mu.Lock()
    defer a.mu.Unlock()

    if len(a.items) == 0 {
        return nil
    }

    result := make(map[string]float64, len(a.items))
    for k, item := range a.items {
        result[k] = item.ScaledValue
    }
    return result
}

func (a *Accumulator) Keys() []string {
    a.mu.Lock()
    defer a.mu.Unlock()

    keys := make([]string, 0, len(a.items))
    for k := range a.items {
        keys = append(keys, k)
    }
    return keys
}

func (a *Accumulator) Scaled(key string) float64 {
    a.mu.Lock()
    defer a.mu.Unlock()

    if item, ok := a.items[key]; ok {
        return item.ScaledValue
    }
    return 0
}
