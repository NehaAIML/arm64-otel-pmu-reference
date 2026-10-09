package shared

import "testing"

func TestSnapshotStruct(t *testing.T) {
	s := Snapshot{CPU: 1, CgroupID: 100}
	if s.CPU != 1 {
		t.Errorf("Expected CPU 1, got %d", s.CPU)
	}
}
