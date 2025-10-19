# Task Completion Checklist

## Before Submitting Changes

### Required Steps (Execute in Order):
1. **Format Code**: 
   - Run `go fmt ./...` to ensure consistent formatting
   - Verify no formatting changes remain

2. **Run Tests**: 
   - Execute `go test ./...` or `make test`
   - Verify all tests pass
   - Check for any test warnings or race conditions

3. **Run Linter**: 
   - Execute `golangci-lint run` or `make lint`
   - Fix all linting errors and warnings
   - Use `make lint-fix` for auto-fixable issues

4. **Build Verification**: 
   - Run `make build` to ensure code compiles
   - Verify binary runs without crashes
   - Test basic functionality if possible

### Optional Quality Checks:
- **Test Coverage**: Run `make test-coverage` and review coverage.html
- **Cross-platform Build**: Run `make build-current` or `make build-all` for multi-platform verification
- **Clean Build**: Run `make clean && make build` for fresh build verification
- **Race Detection**: Run `go test -race ./...` to check for race conditions
- **Vet Check**: Run `go vet ./...` for additional static analysis
- **Benchmark**: Run benchmarks if performance-critical code changed

## Code Review Guidelines

### Before Committing:
- **Error Handling**: Verify all errors are handled explicitly and descriptively
- **Interfaces**: Check that interfaces are properly implemented
- **Tests**: Ensure tests are added for new functionality
- **Documentation**: Confirm exported functions have doc comments
- **Naming**: Validate that naming follows Go conventions
- **Imports**: Check that imports are organized and unused ones removed
- **Comments**: Verify comments are meaningful (avoid stating the obvious)

### Security Review:
- Review browser data access patterns
- Ensure proper file permission handling
- Validate input sanitization for SQL queries
- Check for potential information disclosure
- Verify no sensitive data is logged

### Architecture Review:
- Confirm changes align with existing patterns
- Verify interfaces are used appropriately
- Check that common utilities are reused, not duplicated
- Ensure platform-specific code is properly separated

## Git Workflow
1. Create feature branch from main: `git checkout -b feature/description`
2. Make changes following code style conventions
3. Run complete checklist above
4. Stage relevant changes: `git add <files>`
5. Commit with descriptive message: `git commit -m "feat: add description"`
6. Push to remote: `git push origin feature/description`
7. Create pull request with clear description

## Commit Message Format
```
<type>: <short summary>

<optional body with more details>

<optional footer with references>
```

Types: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`, `perf`

## Platform-Specific Testing

### Test on Target Operating Systems:
- **macOS**: Test profile detection, History database access
- **Linux**: Test Firefox Zen variant, profile paths
- **Windows**: Test Chrome/Edge detection, path handling

### Browser Testing Matrix:
- Chromium: Chrome, Edge, Chromium, Brave, Vivaldi, Comet
- Firefox: Firefox, ESR, Developer Edition, Nightly, Zen (Linux)
- Verify profile enumeration works for multi-profile setups
- Test with both running and closed browsers

## Integration Testing

### With Osquery:
1. Build extension: `make build`
2. Start osquery with socket: `osqueryi --extension /path/to/browser-extend-extension`
3. Test virtual table: `SELECT * FROM browser_history LIMIT 10;`
4. Verify columns: browser, profile, url, title, visit_count, last_visit_time
5. Test filtering: `WHERE browser = 'chrome'`, `WHERE url LIKE '%github%'`
6. Check error handling for locked databases

### Performance Testing:
- Test with large browser history databases (10,000+ entries)
- Verify worker pool handles concurrent requests efficiently
- Check memory usage remains reasonable
- Benchmark query response times

## Documentation Updates
- Update README.md if functionality changes
- Update FIREFOX_HISTORY_CHANGES.md for Firefox schema changes
- Update AGENTS.md / CLAUDE.md for new conventions
- Add entries to troubleshooting guides if bugs fixed