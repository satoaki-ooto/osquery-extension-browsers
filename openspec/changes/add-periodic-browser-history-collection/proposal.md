# Proposal: Add Incremental Browser History Collection

**Change ID:** `add-periodic-browser-history-collection`

## Summary

Add incremental collection capabilities to the browser history extension, enabling efficient periodic queries by supporting time-based filtering. This allows osquery's scheduled query feature to collect only new browser history entries since the last collection.

## Background & Motivation

Currently, the osquery browser history extension only supports retrieving all history entries on each query. When used with osquery's scheduled queries for continuous monitoring, this approach:

1. **Inefficient**: Retrieves the same history entries repeatedly
2. **Resource intensive**: Processes large amounts of duplicate data
3. **Difficult to track**: No built-in way to identify new entries since last check

By adding incremental collection support, the extension can:
- Work efficiently with osquery's scheduled query feature
- Reduce resource usage by only processing new entries
- Enable continuous monitoring scenarios
- Support compliance and security monitoring use cases

## Scope

### In Scope
- Add time-based filtering to history collection functions
- Extend virtual table schema to support incremental queries
- Add configuration for state management
- Update documentation with scheduled query examples

### Out of Scope
- Built-in scheduling (use osquery's scheduled queries)
- External storage (use osquery's logging)
- Complex data retention (use log rotation)
- Real-time alerting (can be built on top)

## Affected Capabilities

This change will affect and require updates to:

1. **History Collection** (`internal/browsers/chromium/history.go`, `internal/browsers/firefox/history.go`)
   - Add time-based filtering support
   - Add `since` parameter to collection functions

2. **Virtual Table** (`cmd/browser_extend_extension/main.go`)
   - Add `since` column for filtering
   - Add `last_visit_time` column for incremental queries

3. **Configuration** (new)
   - Add state file management for tracking last collection time

## Risks & Considerations

### Performance Impact
- **Risk**: Additional filtering may impact query performance
- **Mitigation**: Use indexed timestamp columns in browser databases

### State Management
- **Risk**: State file corruption or loss may cause duplicate collection
- **Mitigation**: Implement robust error handling and recovery

### Browser Compatibility
- **Risk**: Different timestamp formats across browsers
- **Mitigation**: Maintain existing robust parsing logic

## Success Criteria

1. Incremental queries return only entries newer than specified timestamp
2. Works seamlessly with osquery scheduled queries
3. Performance impact is minimal
4. Existing functionality remains unchanged
5. Documentation includes clear usage examples
6. All tests pass with good coverage

## Alternatives Considered

1. **Full periodic collection**: Build complete collection service
   - **Cons**: Overly complex, duplicates osquery functionality

2. **External state management**: Use external database for state
   - **Cons**: Adds unnecessary complexity and dependencies

3. **Current approach**: Add incremental filtering only
   - **Pros**: Simple, leverages osquery features, minimal changes
   - **Cons**: Requires osquery for scheduling

The proposed solution provides the best balance of functionality and simplicity.