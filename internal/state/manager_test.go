package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewStateManager(t *testing.T) {
	// Create temporary directory for test state file
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "state.json")

	// Test creating new state manager
	sm, err := NewStateManager(stateFile)
	if err != nil {
		t.Fatalf("NewStateManager() error = %v", err)
	}

	// Check that state file was created
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		t.Error("State file was not created")
	}

	// Check initial state
	if sm.state == nil {
		t.Error("State was not initialized")
	}

	if sm.state.LastCollectionTime != 0 {
		t.Errorf("LastCollectionTime = %d, want 0", sm.state.LastCollectionTime)
	}

	if sm.state.ProfileStates == nil {
		t.Error("ProfileStates was not initialized")
	}
}

func TestGetLastCollectionTime(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "state.json")

	sm, err := NewStateManager(stateFile)
	if err != nil {
		t.Fatalf("NewStateManager() error = %v", err)
	}

	// Test initial state (zero time)
	got := sm.GetLastCollectionTime()
	if !got.IsZero() {
		t.Errorf("GetLastCollectionTime() = %v, want zero time", got)
	}

	// Update and test
	testTime := time.Unix(1234567890, 0)
	if err := sm.UpdateLastCollectionTime(testTime); err != nil {
		t.Fatalf("UpdateLastCollectionTime() error = %v", err)
	}

	got = sm.GetLastCollectionTime()
	if !got.Equal(testTime) {
		t.Errorf("GetLastCollectionTime() = %v, want %v", got, testTime)
	}
}

func TestGetProfileLastCollectionTime(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "state.json")

	sm, err := NewStateManager(stateFile)
	if err != nil {
		t.Fatalf("NewStateManager() error = %v", err)
	}

	// Test initial state (zero time)
	got := sm.GetProfileLastCollectionTime("profile1")
	if !got.IsZero() {
		t.Errorf("GetProfileLastCollectionTime() = %v, want zero time", got)
	}

	// Update and test
	testTime := time.Unix(1234567890, 0)
	if err := sm.UpdateProfileLastCollectionTime("profile1", testTime); err != nil {
		t.Fatalf("UpdateProfileLastCollectionTime() error = %v", err)
	}

	got = sm.GetProfileLastCollectionTime("profile1")
	if !got.Equal(testTime) {
		t.Errorf("GetProfileLastCollectionTime() = %v, want %v", got, testTime)
	}

	// Test different profile
	got = sm.GetProfileLastCollectionTime("profile2")
	if !got.IsZero() {
		t.Errorf("GetProfileLastCollectionTime(profile2) = %v, want zero time", got)
	}
}

func TestUpdateLastCollectionTime(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "state.json")

	sm, err := NewStateManager(stateFile)
	if err != nil {
		t.Fatalf("NewStateManager() error = %v", err)
	}

	testTime := time.Unix(1234567890, 0)
	if err := sm.UpdateLastCollectionTime(testTime); err != nil {
		t.Fatalf("UpdateLastCollectionTime() error = %v", err)
	}

	// Verify state was updated
	if sm.state.LastCollectionTime != testTime.Unix() {
		t.Errorf("LastCollectionTime = %d, want %d", sm.state.LastCollectionTime, testTime.Unix())
	}

	// Verify state was persisted
	sm2, err := NewStateManager(stateFile)
	if err != nil {
		t.Fatalf("NewStateManager() error = %v", err)
	}

	got := sm2.GetLastCollectionTime()
	if !got.Equal(testTime) {
		t.Errorf("GetLastCollectionTime() after reload = %v, want %v", got, testTime)
	}
}

func TestUpdateProfileLastCollectionTime(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "state.json")

	sm, err := NewStateManager(stateFile)
	if err != nil {
		t.Fatalf("NewStateManager() error = %v", err)
	}

	testTime := time.Unix(1234567890, 0)
	if err := sm.UpdateProfileLastCollectionTime("profile1", testTime); err != nil {
		t.Fatalf("UpdateProfileLastCollectionTime() error = %v", err)
	}

	// Verify state was updated
	if sm.state.ProfileStates["profile1"] != testTime.Unix() {
		t.Errorf("ProfileStates[profile1] = %d, want %d", sm.state.ProfileStates["profile1"], testTime.Unix())
	}

	// Verify state was persisted
	sm2, err := NewStateManager(stateFile)
	if err != nil {
		t.Fatalf("NewStateManager() error = %v", err)
	}

	got := sm2.GetProfileLastCollectionTime("profile1")
	if !got.Equal(testTime) {
		t.Errorf("GetProfileLastCollectionTime() after reload = %v, want %v", got, testTime)
	}
}

func TestReset(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "state.json")

	sm, err := NewStateManager(stateFile)
	if err != nil {
		t.Fatalf("NewStateManager() error = %v", err)
	}

	// Set some state
	testTime := time.Unix(1234567890, 0)
	sm.UpdateLastCollectionTime(testTime)
	sm.UpdateProfileLastCollectionTime("profile1", testTime)

	// Reset
	if err := sm.Reset(); err != nil {
		t.Fatalf("Reset() error = %v", err)
	}

	// Verify state was cleared
	if sm.state.LastCollectionTime != 0 {
		t.Errorf("LastCollectionTime after reset = %d, want 0", sm.state.LastCollectionTime)
	}

	if len(sm.state.ProfileStates) != 0 {
		t.Errorf("ProfileStates length after reset = %d, want 0", len(sm.state.ProfileStates))
	}

	// Verify reset was persisted
	sm2, err := NewStateManager(stateFile)
	if err != nil {
		t.Fatalf("NewStateManager() error = %v", err)
	}

	got := sm2.GetLastCollectionTime()
	if !got.IsZero() {
		t.Errorf("GetLastCollectionTime() after reload = %v, want zero time", got)
	}
}

func TestGetState(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "state.json")

	sm, err := NewStateManager(stateFile)
	if err != nil {
		t.Fatalf("NewStateManager() error = %v", err)
	}

	// Set some state
	testTime := time.Unix(1234567890, 0)
	sm.UpdateLastCollectionTime(testTime)
	sm.UpdateProfileLastCollectionTime("profile1", testTime)

	// Get state copy
	state := sm.GetState()

	// Verify copy
	if state.LastCollectionTime != testTime.Unix() {
		t.Errorf("State.LastCollectionTime = %d, want %d", state.LastCollectionTime, testTime.Unix())
	}

	if state.ProfileStates["profile1"] != testTime.Unix() {
		t.Errorf("State.ProfileStates[profile1] = %d, want %d", state.ProfileStates["profile1"], testTime.Unix())
	}

	// Modify copy and verify original is unchanged
	state.LastCollectionTime = 9999999999
	state.ProfileStates["profile1"] = 8888888888

	originalState := sm.GetState()
	if originalState.LastCollectionTime != testTime.Unix() {
		t.Errorf("Original state was modified: LastCollectionTime = %d, want %d", originalState.LastCollectionTime, testTime.Unix())
	}
}

func TestLoadExistingState(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "state.json")

	// Create initial state manager and set some state
	sm1, err := NewStateManager(stateFile)
	if err != nil {
		t.Fatalf("NewStateManager() error = %v", err)
	}

	testTime := time.Unix(1234567890, 0)
	sm1.UpdateLastCollectionTime(testTime)
	sm1.UpdateProfileLastCollectionTime("profile1", testTime)

	// Create new state manager with same file
	sm2, err := NewStateManager(stateFile)
	if err != nil {
		t.Fatalf("NewStateManager() error = %v", err)
	}

	// Verify state was loaded
	got := sm2.GetLastCollectionTime()
	if !got.Equal(testTime) {
		t.Errorf("GetLastCollectionTime() = %v, want %v", got, testTime)
	}

	gotProfile := sm2.GetProfileLastCollectionTime("profile1")
	if !gotProfile.Equal(testTime) {
		t.Errorf("GetProfileLastCollectionTime() = %v, want %v", gotProfile, testTime)
	}
}

func TestConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "state.json")

	sm, err := NewStateManager(stateFile)
	if err != nil {
		t.Fatalf("NewStateManager() error = %v", err)
	}

	// Run concurrent operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			testTime := time.Unix(int64(1234567890+id), 0)
			sm.UpdateLastCollectionTime(testTime)
			sm.GetLastCollectionTime()
			sm.UpdateProfileLastCollectionTime("profile1", testTime)
			sm.GetProfileLastCollectionTime("profile1")
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify state is consistent
	state := sm.GetState()
	if state.LastCollectionTime == 0 {
		t.Error("LastCollectionTime was not updated")
	}
}
