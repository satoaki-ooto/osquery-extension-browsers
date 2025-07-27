# osquery Extension Development Planning Template

## Project Overview

### Project Name

`chrome_extend_extension` - Chrome Browser Data osquery Extension

### Description

A robust osquery extension that provides access to Chrome browser history and profile information through custom tables. This extension migrates from Python to Go, leveraging Go's static typing and concurrency features for improved performance and reliability.

### Target Tables

- `chrome_history`: Browser history with timestamps, titles, URLs, profiles, and browser types
- `chrome_profiles`: Browser profile information including email addresses

## Technical Architecture

### Core Design Principles

1. **Modular Package Structure**: Prevent monolithic `main.go` by separating functionality into focused packages
2. **Concurrent Processing**: Utilize Go's goroutines and channels for parallel data processing
3. **Cross-Platform Support**: Handle OS-specific path resolution for Linux, macOS, and Windows
4. **Error Resilience**: Implement retry mechanisms for database locks and file access issues

### Package Architecture

**Adjusted Package Architecture (Chrome-like browsers + Firefox)**

```
browser_extend_extension/
├── cmd/
│   └── browser_extend_extension/
│       └── main.go          # Entry point and table registration
├── internal/                # Private packages
│   ├── browsers/            # Browser-specific implementations
│   │   ├── chromium/        # Chrome-like browsers (Chrome, Edge, Chromium)
│   │   │   ├── finder.go    # Chromium-based file discovery
│   │   │   ├── history.go   # Chromium history parsing (WebKit timestamps)
│   │   │   ├── profile.go   # Chromium profile parsing (Local State JSON)
│   │   │   └── variants.go  # Browser variant detection (Chrome, Edge, Chromium)
│   │   ├── firefox/
│   │   │   ├── finder.go    # Firefox-specific file discovery
│   │   │   ├── history.go   # Firefox history parsing (places.sqlite)
│   │   │   └── profile.go   # Firefox profile parsing (profiles.ini)
│   │   └── common/
│   │       ├── interfaces.go # Browser interface definitions
│   │       └── detector.go   # Browser installation detection
│   ├── tables/              # osquery table implementations
│   │   ├── browser_history/ # Multi-browser history table
│   │   │   ├── table.go
│   │   │   └── aggregator.go
│   │   └── browser_profiles/ # Multi-browser profile table
│   │       ├── table.go
│   │       └── aggregator.go
│   └── common/              # Shared utilities
│       ├── process.go       # Browser process detection
│       ├── retry.go         # Exponential backoff retry logic
│       └── timestamp.go     # Timestamp conversion utilities
└── go.mod
```

**Key Changes:**

- Renamed `chrome/` to `chromium/` to reflect support for Chrome-like browsers
- Added `variants.go` to handle Chrome, Edge, and Chromium detection
- Kept Firefox separate due to fundamentally different data formats
- Safari marked as future TODO (not included in current architecture)

## Implementation Roadmap

### Phase 0: Project Setup

**Objective**: Initialize Go project with required dependencies

**Tasks**:

- [x] Initialize Go module
- [x] Add required dependencies:
  - `github.com/mattn/go-sqlite3` for SQLite operations
  - `github.com/shirou/gopsutil/v3/process` for process management
  - osquery SDK dependencies
- [x] Set up basic project structure
- [x] Configure build system and CI/CD

**Deliverables**:

- Working Go project with dependencies
- Basic project structure
- Build configuration

### Phase 1: Chromium-Based Browser Foundation

**Objective**: Create extensible Chromium-based browser support (Chrome, Edge, Chromium)

**Tasks**:

- [ ] Define common browser interfaces
- [ ] Implement Chromium browser provider (`internal/browsers/chromium/`)
- [ ] Create browser variant detection (Chrome, Edge, Chromium)
- [ ] Build unified file discovery for Chromium-based browsers
- [ ] Implement shared WebKit timestamp conversion

### Phase 2: Chromium Implementation

**Objective**: Implement Chromium-based data extraction

**Tasks**:

- [ ] Chromium file discovery with variant support (`chromium/finder.go`)
- [ ] Chromium history parsing (`chromium/history.go`)
- [ ] Chromium profile parsing (`chromium/profile.go`)
- [ ] Browser variant identification (`chromium/variants.go`)
- [ ] Process detection for all Chromium variants

### Phase 3: Firefox Implementation

**Objective**: Add Firefox browser support

**Tasks**:

- [ ] Firefox file discovery (`firefox/finder.go`)
- [ ] Firefox places.sqlite parsing (`firefox/history.go`)
- [ ] Firefox profiles.ini parsing (`firefox/profile.go`)
- [ ] Unix timestamp handling for Firefox

### Phase 4: Unified Table Implementation

**Objective**: Create multi-browser osquery tables

**Tasks**:

- [ ] `browser_history` table with Chromium variants and Firefox
- [ ] `browser_profiles` table with Chromium variants and Firefox
- [ ] Browser type and variant identification in results
- [ ] Performance optimization for multi-browser queries

### Future TODO Items

- [ ] Safari support (requires extensive privacy handling and different data formats)
- [ ] Opera support (evaluate if Chromium-based or requires separate implementation)
- [ ] Additional Chromium variants (Brave, Vivaldi, etc.)

### Phase 4: Integration and Testing

**Objective**: Integrate components and ensure system reliability

**Tasks**:

- [ ] Integrate all packages in `main.go`
- [ ] Register both tables with osquery extension server
- [ ] Update extension name from `test_extension` to `chrome_extend_extension`
- [ ] Implement comprehensive logging using Go's `log` package
- [ ] Create integration tests
- [ ] Performance testing with large datasets
- [ ] Cross-platform compatibility testing
- [ ] Documentation and usage examples

**Deliverables**:

- Complete osquery extension
- Test suite with >90% coverage
- Performance benchmarks
- User documentation

## Technical Specifications

### Dependencies

```go
// Core dependencies
github.com/mattn/go-sqlite3           // SQLite driver
github.com/shirou/gopsutil/v3/process // Process management

// Standard library
database/sql                          // Database interface
encoding/json                         // JSON parsing
path/filepath                         // File path utilities
runtime                              // OS detection
sync                                 // Concurrency primitives
time                                 // Time utilities
```

### Error Handling Strategy

1. **Database Locks**: Exponential backoff retry (max 5 attempts)
2. **File Access**: Graceful degradation with detailed logging
3. **JSON Parsing**: Schema validation with fallback defaults
4. **Cross-Platform**: OS-specific error codes and messages

### Performance Considerations

- **Concurrency**: Process multiple profiles simultaneously using goroutines
- **Memory Management**: Stream large datasets to avoid memory exhaustion
- **Database Access**: Use connection pooling and prepared statements
- **Caching**: Cache file discovery results to reduce filesystem I/O

### Security Considerations

- **Read-Only Access**: All database connections use immutable, read-only mode
- **Path Validation**: Prevent directory traversal attacks
- **Error Information**: Sanitize error messages to prevent information disclosure
- **Process Detection**: Non-intrusive browser process monitoring

## Quality Assurance

### Testing Strategy

- **Unit Tests**: Individual package functionality
- **Integration Tests**: End-to-end table queries
- **Performance Tests**: Large dataset handling
- **Compatibility Tests**: Cross-platform validation

### Code Quality Standards

- **Go Standards**: Follow Go idioms and best practices
- **Documentation**: Comprehensive godoc comments
- **Error Handling**: Explicit error handling for all operations
- **Logging**: Structured logging with appropriate levels

### Monitoring and Observability

- **Performance Metrics**: Query execution time tracking
- **Error Rates**: Database access failure monitoring
- **Resource Usage**: Memory and CPU utilization tracking

## Deployment Considerations

### Build Configuration

- **Static Linking**: Single binary deployment
- **Cross-Compilation**: Support for Linux, macOS, Windows
- **Version Management**: Semantic versioning and release tags

### Installation Process

- **Package Distribution**: OS-specific package formats
- **Configuration**: Environment-based configuration options
- **Upgrades**: Backward-compatible schema evolution

## Risk Assessment

### Technical Risks

1. **Database Locks**: Browsers holding exclusive locks on History databases
   - _Mitigation_: Retry logic with exponential backoff, process detection
2. **File Format Changes**: Chrome updating database/JSON schemas
   - _Mitigation_: Version detection, schema validation, graceful degradation
3. **Performance Impact**: Large history databases affecting query performance
   - _Mitigation_: Streaming, pagination, configurable limits

### Operational Risks

1. **Cross-Platform Compatibility**: Path differences across operating systems
   - _Mitigation_: Comprehensive testing, OS-specific implementations
2. **Security Vulnerabilities**: Potential information disclosure
   - _Mitigation_: Read-only access, input validation, security audits

## Success Criteria

### Functional Requirements

- [ ] Successfully extract browser history from all supported browsers
- [ ] Accurately parse profile information including email addresses
- [ ] Handle multiple browser installations and profiles
- [ ] Provide consistent cross-platform behavior

### Performance Requirements

- [ ] Query response time <2 seconds for datasets up to 100K records
- [ ] Memory usage <100MB during normal operations
- [ ] Support concurrent access from multiple osquery instances

### Reliability Requirements

- [ ] Handle database locks gracefully without data loss
- [ ] Recover from corrupted or inaccessible files
- [ ] Maintain functionality across Chrome browser updates

This planning document serves as a comprehensive guide for implementing the osquery extension across different development tools and environments.
