package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)


func setupCgroup(name string, pid int, mem string, cpuCores float64) (string, error) {
	if err := ensureCgroupDelegation(); err != nil {
		return "", err
	}
	cgroupPath := filepath.Join("/sys/fs/cgroup", "mygodocker", name+"-"+strconv.Itoa(pid))
	if err := os.MkdirAll(cgroupPath, 0755); err != nil {
		return "", fmt.Errorf("creating cgroup dir: %w", err)
	}
	if mem != "" {
		limit, err := parseMemLimit(mem)
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(filepath.Join(cgroupPath, "memory.max"), []byte(limit), 0644); err != nil {
			return "", fmt.Errorf("memory.max: %w", err)
		}
	}
	if cpuCores > 0 {
		if err := os.WriteFile(filepath.Join(cgroupPath, "cpu.max"), []byte(parseCPULimit(cpuCores)), 0644); err != nil {
			return "", fmt.Errorf("cpu.max: %w", err)
		}
	}
	if err := os.WriteFile(filepath.Join(cgroupPath, "cgroup.procs"), []byte(strconv.Itoa(pid)), 0644); err != nil {
		return "", fmt.Errorf("cgroup.procs: %w", err)
	}
	return cgroupPath, nil
}

func ensureCgroupDelegation() error {
	if err := os.MkdirAll("/sys/fs/cgroup/mygodocker", 0755); err != nil {
		return err
	}
	for _, dir := range []string{"/sys/fs/cgroup", "/sys/fs/cgroup/mygodocker"} {
		f := filepath.Join(dir, "cgroup.subtree_control") 
		if err := os.WriteFile(f, []byte("+memory +cpu"), 0644); err != nil {
			return fmt.Errorf("enabling controllers at %s: %w", dir, err)
		} 
	}
	return nil
}

func parseCPULimit(cores float64) string {
	const period = 100000
	quota := int(cores * float64(period))
	return fmt.Sprintf("%d %d", quota, period)
}

func parseMemLimit(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "max", nil
	}
	unit := s[len(s)-1]
	numPart := s[:len(s)-1]
	switch unit {
	case 'k', 'K':
		return numPart + "K", nil
	case 'm', 'M':
		return numPart + "M", nil
	case 'g', 'G':
		return numPart + "G", nil
	default:
		if _, err := strconv.Atoi(s); err != nil {
			return "", fmt.Errorf("invalid memory value %q", s)
		}
		return s, nil
	}
}