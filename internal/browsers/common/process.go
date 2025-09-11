package common

import (
	"github.com/shirou/gopsutil/v3/process"
)

// IsProcessRunning checks if a process with the given name is currently running
func IsProcessRunning(processName string) (bool, error) {
	processes, err := process.Processes()
	if err != nil {
		return false, err
	}

	for _, p := range processes {
		name, err := p.Name()
		if err != nil {
			// Skip processes we can't get the name for
			continue
		}

		if name == processName {
			return true, nil
		}
	}

	return false, nil
}
