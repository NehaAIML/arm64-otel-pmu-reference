package cgroup

import (
    "fmt"
    "os"
    "path/filepath"
    "strconv"
    "syscall"
)

type Manager struct {
    root string
}

func NewManager(root string) *Manager {
    return &Manager{root: root}
}

func (m *Manager) Enumerate() (map[string]string, error) {
    results := make(map[string]string)
    
    if _, err := os.Stat(m.root); err != nil {
        return nil, fmt.Errorf("cgroup root inaccessible: %w", err)
    }

    err := filepath.Walk(m.root, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return nil
        }
        if info.IsDir() {
            var stat syscall.Stat_t
            if err := syscall.Stat(path, &stat); err == nil {
                idStr := strconv.FormatUint(stat.Ino, 10)
                results[idStr] = path
            }
        }
        return nil
    })

    if err != nil {
        return nil, fmt.Errorf("failed to enumerate cgroups: %w", err)
    }

    return results, nil
}

func (m *Manager) ResolveCgroupID(path string) (uint64, error) {
    var stat syscall.Stat_t
    if err := syscall.Stat(path, &stat); err != nil {
        return 0, fmt.Errorf("failed to stat cgroup path %s: %w", path, err)
    }
    return stat.Ino, nil
}
