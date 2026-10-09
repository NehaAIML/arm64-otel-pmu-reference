package accum

import (
	"math"
	"sync"
)

const MaxScale = 10.0

type Sample struct{ Counter, Enabled, Running uint64 }
type state struct {
	counter, enabled, running uint64
	scaled                    float64
}
type Delta struct {
	Scaled float64
	Reset  bool
}
type Accumulator struct {
	mu     sync.RWMutex
	states map[string]*state
}

func NewAccumulator() *Accumulator { return &Accumulator{states: map[string]*state{}} }
func (a *Accumulator) Update(key string, s Sample) Delta {
	a.mu.Lock()
	defer a.mu.Unlock()
	st, ok := a.states[key]
	if !ok {
		a.states[key] = &state{counter: s.Counter, enabled: s.Enabled, running: s.Running}
		return Delta{}
	}
	if s.Counter < st.counter {
		st.counter, st.enabled, st.running = s.Counter, s.Enabled, s.Running
		return Delta{Reset: true}
	}
	dCount := float64(s.Counter - st.counter)
	dEn := s.Enabled - st.enabled
	dRun := s.Running - st.running
	scale := 1.0
	if dRun > 0 && dEn >= dRun {
		scale = math.Min(float64(dEn)/float64(dRun), MaxScale)
	}
	st.counter, st.enabled, st.running = s.Counter, s.Enabled, s.Running
	d := dCount * scale
	st.scaled += d
	return Delta{Scaled: d}
}
func (a *Accumulator) Scaled(key string) float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if st, ok := a.states[key]; ok {
		return st.scaled
	}
	return 0
}
func (a *Accumulator) Keys() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]string, 0, len(a.states))
	for k := range a.states {
		out = append(out, k)
	}
	return out
}
func (a *Accumulator) DropStale(valid map[string]bool) []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	var dropped []string
	for k := range a.states {
		if !valid[k] {
			delete(a.states, k)
			dropped = append(dropped, k)
		}
	}
	return dropped
}
