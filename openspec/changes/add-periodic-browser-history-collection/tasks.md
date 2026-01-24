# Implementation Tasks: Add Incremental Browser History Collection

**Change ID:** `add-periodic-browser-history-collection`

## Phase 1: Core Implementation

### 1.1 History Collection Enhancement
- [ ] Add `HistoryOptions` struct with `Since` and `Limit` fields
- [ ] Add `HistoryOption` functional options pattern
- [ ] Update `FindHistory()` in chromium package to support `since` parameter
- [ ] Update `FindHistory()` in firefox package to support `since` parameter
- [ ] Add unit tests for incremental collection
- [ ] Add benchmarks for performance testing

### 1.2 Virtual Table Enhancement
- [ ] Add `visit_time` column to virtual table schema
- [ ] Add `visit_count` column to virtual table schema
- [ ] Add `browser_variant` column to virtual table schema
- [ ] Update query handler to support time-based filtering
- [ ] Add integration tests for virtual table queries

### 1.3 State Management
- [ ] Create `StateManager` struct
- [ ] Implement state file read/write operations
- [ ] Add file locking for concurrent access safety
- [ ] Implement error recovery for corrupted state files
- [ ] Add unit tests for state management

## Phase 2: Testing & Quality

### 2.1 Unit Tests
- [ ] Test history collection with various `since` values
- [ ] Test backward compatibility (nil `since`)
- [ ] Test state file operations
- [ ] Test error handling for edge cases
- [ ] Achieve >90% code coverage

### 2.2 Integration Tests
- [ ] Test full incremental collection workflow
- [ ] Test state persistence across restarts
- [ ] Test with multiple browser profiles
- [ ] Test concurrent query scenarios
- [ ] Test with real browser databases

### 2.3 Performance Tests
- [ ] Benchmark collection with large history databases
- [ ] Test query performance with incremental filters
- [ ] Measure memory usage during collection
- [ ] Test state file I/O performance

## Phase 3: Documentation & Deployment

### 3.1 Documentation
- [ ] Update README with incremental collection feature
- [ ] Add usage examples with osquery scheduled queries
- [ ] Document state file format and location
- [ ] Add troubleshooting guide
- [ ] Document performance considerations

### 3.2 Examples
- [ ] Create example osquery configuration
- [ ] Add example scheduled query configurations
- [ ] Create example scripts for state management

## Phase 4: Validation & Review

### 4.1 Functional Validation
- [ ] Verify incremental queries return only new entries
- [ ] Verify state file updates correctly
- [ ] Verify backward compatibility
- [ ] Verify error recovery mechanisms
- [ ] Verify all tests pass

### 4.2 Performance Validation
- [ ] Verify minimal performance impact
- [ ] Verify efficient resource usage
- [ ] Verify state file I/O doesn't bottleneck

### 4.3 Code Review
- [ ] Review history collection changes
- [ ] Review state management implementation
- [ ] Review test coverage
- [ ] Review documentation

## Success Criteria Checklist

- [ ] Incremental queries work correctly with `since` parameter
- [ ] State management is robust and handles errors gracefully
- [ ] Virtual table supports time-based filtering
- [ ] All tests pass with good coverage
- [ ] Documentation is complete and accurate
- [ ] Backward compatibility is maintained
- [ ] Performance impact is minimal
- [ ] Code follows project style guidelines

## Dependencies

- No external dependencies beyond standard Go libraries
- Uses existing browser detection and history extraction logic
- Leverages osquery's scheduled query feature

## Estimated Effort

- **Phase 1**: 3-4 days
- **Phase 2**: 2-3 days
- **Phase 3**: 1-2 days
- **Phase 4**: 1-2 days

**Total Estimated Effort**: 7-11 days