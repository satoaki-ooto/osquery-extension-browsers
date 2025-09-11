# Implementation Plan

- [x] 1. Create user enumeration infrastructure
  - Implement `UserInfo` struct and `usersFromContext()` function in `internal/browser/common/detector.go`
  - Add cross-platform user enumeration logic with proper error handling
  - Include unit tests for user enumeration functionality
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 5.3_

- [x] 2. Implement helper functions for multi-user path generation
  - Create `findFirefoxPathsForUser()` helper function in `internal/browsers/firefox/finder.go`
  - Create `findChromiumPathsForUser()` helper function in `internal/browsers/chromium/finder.go`
  - Add unit tests for per-user path generation logic
  - _Requirements: 4.1, 4.2, 2.1, 2.2_

- [x] 3. Enhance FindFirefoxPaths with multi-user support
  - Modify `FindFirefoxPaths()` function to use `usersFromContext()` and scan all users
  - Implement graceful error handling for inaccessible user directories
  - Add fallback mechanism to current user when system enumeration fails
  - Write comprehensive unit tests for enhanced Firefox path detection
  - _Requirements: 4.1, 4.3, 4.4, 5.1, 5.2, 5.4_

- [x] 4. Enhance FindChromiumPaths with multi-user support
  - Modify `FindChromiumPaths()` function to use `usersFromContext()` and scan all users
  - Implement graceful error handling for inaccessible user directories
  - Add fallback mechanism to current user when system enumeration fails
  - Write comprehensive unit tests for enhanced Chromium path detection
  - _Requirements: 4.2, 4.3, 4.4, 5.1, 5.2, 5.4_

- [ ] 5. Add integration tests and performance optimization
  - Create integration tests that verify complete multi-user browser detection flow
  - Implement concurrent user scanning with goroutine worker pool for performance
  - Add benchmarks to measure performance impact of multi-user scanning
  - Verify backward compatibility with existing code that calls the enhanced functions
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 2.3, 2.4_