package cgroup

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

type Manager struct {
	mu    sync.RWMutex
	paths map[string]string
	root  string
}

func NewManager(root string) *Manager { return &Manager{paths: map[string]string{}, root: root} }
func (m *Manager) ResolveID(cgroupPath string) (string, error) {
	if !strings.HasPrefix(cgroupPath, "/") {
		cgroupPath = filepath.Join(m.root, cgroupPath)
	}
	var st syscall.Stat_t
	if err := syscall.Stat(cgroupPath, &st); err != nil {
		return "", err
	}
	id := strconv.FormatUint(uint64(st.Ino), 10)
	m.mu.Lock()
	m.paths[id] = cgroupPath
	m.mu.Unlock()
	return id, nil
}
func (m *Manager) Enumerate() (map[string]string, error) {
	valid := map[string]string{}
	filepath.Walk(m.root, func(p string, info os.FileInfo, werr error) error {
		if werr != nil || info == nil || !info.IsDir() {
			return nil
		}
		if id, rerr := m.ResolveID(p); rerr == nil {
			valid[id] = p
		}
		return nil
	})
	return valid, nil
}
func (m *Manager) GetPath(idStr string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.paths[idStr]
}
