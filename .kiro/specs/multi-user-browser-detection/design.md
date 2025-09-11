# Design Document

## Overview

This design implements a multi-user browser detection system that extends the existing browser finder functionality to enumerate all system users and scan their browser data directories. The system will use Go's `os/user` package for cross-platform user enumeration and enhance the existing `FindFirefoxPaths` and `FindChromiumPaths` functions to include multi-user support.

## Architecture

The enhancement follows the existing modular architecture with these key components:

1. **User Enumeration Service** (`internal/common/detector.go`)
   - `usersFromContext()` function for cross-platform user discovery
   - Error handling for permission issues and system failures

2. **Enhanced Browser Finders** (`internal/browsers/*/finder.go`)
   - Modified `FindFirefoxPaths()` and `FindChromiumPaths()` functions
   - Multi-user path generation with fallback mechanisms

3. **Cross-Platform Compatibility**
   - Windows: Uses `os/user.LookupId()` with SID enumeration
   - macOS/Linux: Uses `os/user.LookupId()` with UID enumeration
   - Fallback to current user when system enumeration fails

## Components and Interfaces

### User Enumeration Component

```go
// usersFromContext returns all system users with their home directories
func usersFromContext() ([]UserInfo, error)

type UserInfo struct {
    Username string
    HomeDir  string
    UID      string
}
```

**Implementation Strategy:**
- Use `os/user.LookupId()` to iterate through user IDs
- On Unix systems: iterate through UIDs 1000-65533 (typical user range)
- On Windows: use WMI or registry enumeration for user profiles
- Handle permission errors gracefully and continue enumeration

### Enhanced Browser Finder Interface

The existing finder functions will be enhanced to support multi-user scanning:

```go
// Enhanced function signatures (backward compatible)
func FindFirefoxPaths() []string
func FindChromiumPaths() []string

// Internal helper functions
func findFirefoxPathsForUser(userInfo UserInfo) []string
func findChromiumPathsForUser(userInfo UserInfo) []string
```

## Data Models

### UserInfo Structure
```go
type UserInfo struct {
    Username string // System username
    HomeDir  string // User's home directory path
    UID      string // User ID (platform-specific format)
}
```

### Enhanced Path Discovery
The system will generate browser paths for each user following these patterns:

**Firefox Paths per User:**
- Windows: `{UserHome}\AppData\Roaming\Mozilla\Firefox\Profiles`
- macOS: `{UserHome}/Library/Application Support/Firefox/Profiles`
- Linux: `{UserHome}/.mozilla/firefox`

**Chromium Paths per User:**
- Windows: `{UserHome}\AppData\Local\Google\Chrome\User Data`
- macOS: `{UserHome}/Library/Application Support/Google/Chrome`
- Linux: `{UserHome}/.config/google-chrome`

## Error Handling

### Permission Management
1. **Graceful Degradation**: When user enumeration fails, fall back to current user
2. **Directory Access**: Skip inaccessible user directories without failing
3. **Logging Strategy**: Log permission errors at DEBUG level to avoid noise
4. **Timeout Protection**: Implement reasonable timeouts for user directory scanning

### Error Recovery Patterns
```go
// Pseudo-code for error handling pattern
func FindFirefoxPaths() []string {
    var allPaths []string
    
    users, err := usersFromContext()
    if err != nil {
        // Fallback to current user only
        return findFirefoxPathsCurrentUser()
    }
    
    for _, user := range users {
        userPaths := findFirefoxPathsForUser(user)
        allPaths = append(allPaths, userPaths...)
    }
    
    return allPaths
}
```

## Testing Strategy

### Unit Testing
1. **User Enumeration Tests**
   - Mock user enumeration for different OS scenarios
   - Test permission error handling
   - Verify fallback mechanisms

2. **Path Generation Tests**
   - Test path generation for different user home directories
   - Verify cross-platform path formatting
   - Test with special characters in usernames/paths

3. **Integration Tests**
   - Test complete multi-user browser detection flow
   - Verify backward compatibility with existing code
   - Test performance with large numbers of users

### Test Data Scenarios
- Systems with 1, 10, 100+ users
- Users with and without browser installations
- Mixed permission scenarios (accessible/restricted directories)
- Different operating system configurations

## Performance Considerations

### Optimization Strategies
1. **Concurrent Scanning**: Use goroutines to scan multiple users simultaneously
2. **Early Termination**: Stop scanning if maximum reasonable paths are found
3. **Caching**: Cache user enumeration results for short periods
4. **Selective Scanning**: Option to limit scanning to specific user ranges

### Resource Management
```go
// Concurrent user scanning with worker pool
func scanUsersForBrowsers(users []UserInfo) []string {
    const maxWorkers = 10
    // Implementation with worker pool pattern
}
```

## Security Considerations

### Access Control
- Never attempt to escalate privileges for directory access
- Respect system permission boundaries
- Log security-relevant events appropriately
- Avoid exposing sensitive user information in error messages

### Privacy Protection
- Only access browser-related directories
- Don't traverse or log contents of user files
- Minimize data retention from user enumeration
- Follow principle of least privilege

## Migration Strategy

### Backward Compatibility
The enhanced functions maintain full backward compatibility:
- Existing function signatures remain unchanged
- Current behavior is preserved when user enumeration fails
- No breaking changes to existing API contracts

### Rollout Plan
1. **Phase 1**: Implement `usersFromContext()` function
2. **Phase 2**: Enhance `FindFirefoxPaths()` with multi-user support
3. **Phase 3**: Enhance `FindChromiumPaths()` with multi-user support
4. **Phase 4**: Add comprehensive testing and performance optimization