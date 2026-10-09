package mapreader

type Snapshot struct {
    Counter uint64
    Enabled uint64
    Running uint64
}

type MapReader struct{}

func NewMapReader() *MapReader {
    return &MapReader{}
}

func (m *MapReader) ReadSnapshots() map[string]Snapshot {
    return map[string]Snapshot{
        "0:1234:arm64.pmu.cycles":       {Counter: 2000, Enabled: 6000, Running: 6000},
        "0:1234:arm64.pmu.instructions": {Counter: 5000, Enabled: 6000, Running: 6000},
        "0:1234:arm64.pmu.cache_misses": {Counter: 120, Enabled: 6000, Running: 6000},
    }
}
