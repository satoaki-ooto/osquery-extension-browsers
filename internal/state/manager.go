package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CollectionState represents the state of browser history collection.
type CollectionState struct {
	LastCollectionTime int64            `json:"last_collection_time"`
	ProfileStates      map[string]int64 `json:"profile_states"`
}

// StateManager manages the collection state for incremental history collection.
type StateManager struct {
	stateFile string
	mu        sync.RWMutex
	state     *CollectionState
}

// NewStateManager creates a new StateManager with the specified state file path.
// If the state file doesn't exist, it will be created with default values.
func NewStateManager(stateFile string) (*StateManager, error) {
	sm := &StateManager{
		stateFile: stateFile,
		state: &CollectionState{
			ProfileStates: make(map[string]int64),
		},
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(stateFile), 0755); err != nil {
		return nil, fmt.Errorf("failed to create state directory: %w", err)
	}

	// Load existing state if file exists
	if _, err := os.Stat(stateFile); err == nil {
		if err := sm.load(); err != nil {
			// Log warning but don't fail - start with fresh state
			fmt.Printf("Warning: failed to load state file, using fresh state: %v\n", err)
		}
	}

	return sm, nil
}

// GetLastCollectionTime returns the last collection time for the entire system.
func (sm *StateManager) GetLastCollectionTime() time.Time {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sm.state.LastCollectionTime == 0 {
		return time.Time{}
	}
	return time.Unix(sm.state.LastCollectionTime, 0)
}

// GetProfileLastCollectionTime returns the last collection time for a specific profile.
func (sm *StateManager) GetProfileLastCollectionTime(profileID string) time.Time {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if timestamp, ok := sm.state.ProfileStates[profileID]; ok && timestamp > 0 {
		return time.Unix(timestamp, 0)
	}
	return time.Time{}
}

// UpdateLastCollectionTime updates the last collection time for the entire system.
func (sm *StateManager) UpdateLastCollectionTime(t time.Time) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.state.LastCollectionTime = t.Unix()
	return sm.save()
}

// UpdateProfileLastCollectionTime updates the last collection time for a specific profile.
func (sm *StateManager) UpdateProfileLastCollectionTime(profileID string, t time.Time) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.state.ProfileStates[profileID] = t.Unix()
	return sm.save()
}

// Reset clears all collection state, triggering a full collection on next run.
func (sm *StateManager) Reset() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.state = &CollectionState{
		ProfileStates: make(map[string]int64),
	}
	return sm.save()
}

// load reads the state from the state file.
func (sm *StateManager) load() error {
	data, err := os.ReadFile(sm.stateFile)
	if err != nil {
		return err
	}

	var state CollectionState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	// Initialize profile states map if nil
	if state.ProfileStates == nil {
		state.ProfileStates = make(map[string]int64)
	}

	sm.state = &state
	return nil
}

// save writes the state to the state file.
func (sm *StateManager) save() error {
	data, err := json.MarshalIndent(sm.state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to temporary file first, then rename for atomic write
	tmpFile := sm.stateFile + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	// Rename temporary file to actual file (atomic on Unix)
	if err := os.Rename(tmpFile, sm.stateFile); err != nil {
		os.Remove(tmpFile) // Clean up temporary file
		return fmt.Errorf("failed to rename state file: %w", err)
	}

	return nil
}

// GetState returns a copy of the current state.
func (sm *StateManager) GetState() CollectionState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Return a deep copy to avoid race conditions
	return CollectionState{
		LastCollectionTime: sm.state.LastCollectionTime,
		ProfileStates:      copyMap(sm.state.ProfileStates),
	}
}

// copyMap creates a deep copy of a string to int64 map.
func copyMap(m map[string]int64) map[string]int64 {
	result := make(map[string]int64, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}
