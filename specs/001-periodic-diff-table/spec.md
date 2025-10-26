# Feature Specification: Periodic Diff Table for Browser History

**Feature Branch**: `001-periodic-diff-table`
**Created**: 2025-10-22
**Status**: Draft
**Input**: User description: "osquerydの実行をすると差分が出ないためにログが出力されません。なので、定期的な出力をするようにqueryの実行時間を保持して、前回の実行時間からの差分だけを持つ専用のtableを用意したいです。"

## Problem Statement

Currently, when osqueryd executes queries against browser history, no logs are generated when there are no differences between consecutive queries. This creates visibility gaps in monitoring, as administrators cannot distinguish between "no changes occurred" and "monitoring is not functioning." Additionally, there is no mechanism to ensure regular output for compliance or auditing purposes that require periodic reporting regardless of data changes.

## Clarifications

### Session 2025-10-22

- Q: Should the system use system clock time (wall clock) or monotonic time (elapsed time since boot) for execution timestamps? → A: Wall clock time (system time)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Query Execution Time Tracking (Priority: P1)

Security administrators and system operators need to track when browser history queries were last executed, even when no data changes occurred, to verify that monitoring is functioning correctly and to meet compliance requirements for periodic reporting.

**Why this priority**: This is the foundational capability. Without tracking execution times, periodic output and time-based filtering cannot function. This represents the minimum viable feature that solves the core problem.

**Independent Test**: Can be fully tested by executing a query multiple times and verifying that execution timestamps are persisted and retrievable, delivering the value of execution time visibility.

**Acceptance Scenarios**:

1. **Given** osqueryd has never queried browser history, **When** the first query executes, **Then** the current timestamp is recorded as the last execution time
2. **Given** a previous query execution timestamp exists, **When** a new query executes, **Then** the current timestamp replaces the previous execution time
3. **Given** multiple browser history sources (Chrome, Firefox, Safari), **When** queries execute, **Then** each source maintains its own independent execution timestamp
4. **Given** osqueryd restarts, **When** the extension reloads, **Then** previously stored execution timestamps persist and are accessible

---

### User Story 2 - Time-Based Diff Filtering (Priority: P2)

Security administrators need to query only browser history entries that were added since the last query execution, reducing data volume and focusing on new activities for security monitoring.

**Why this priority**: This builds on P1 by adding filtering logic. It delivers significant value by reducing noise and improving query efficiency, but depends on execution time tracking being functional.

**Independent Test**: Can be tested by adding browser history entries at known times, executing queries, and verifying that only entries newer than the last execution timestamp are returned.

**Acceptance Scenarios**:

1. **Given** last execution time is T1 and browser history has entries with timestamps T0, T1.5, T2, **When** current query executes at T3, **Then** only entries with timestamps > T1 are returned (T1.5 and T2)
2. **Given** no previous execution time exists (first run), **When** query executes, **Then** all available browser history entries are returned
3. **Given** last execution time is T1, **When** no new browser history entries exist after T1, **Then** query returns empty result set
4. **Given** browser history entries have identical timestamps, **When** query filters by last execution time, **Then** entries with timestamp equal to last execution time are excluded (strict greater-than comparison)

---

### User Story 3 - Dedicated Periodic Diff Table (Priority: P3)

System operators need a dedicated osquery table that automatically returns only new browser history entries since the last query, simplifying query syntax and reducing manual timestamp management.

**Why this priority**: This is a convenience layer that wraps P1 and P2 into a user-friendly table interface. It provides the best user experience but is not essential for core functionality.

**Independent Test**: Can be tested by querying the dedicated table multiple times and verifying automatic time-based filtering without explicit timestamp parameters.

**Acceptance Scenarios**:

1. **Given** osqueryd is running, **When** operator queries the periodic diff table for browser history, **Then** only entries added since the previous query to this table are returned
2. **Given** multiple queries to the periodic diff table occur, **When** each query completes, **Then** subsequent queries use the completion time of the previous query as their baseline
3. **Given** the periodic diff table and the standard browser history table both exist, **When** queries execute against each, **Then** they maintain independent execution timestamps (querying standard table does not affect periodic diff table baseline)
4. **Given** osqueryd configuration specifies a query interval (e.g., every 5 minutes), **When** scheduled queries execute against the periodic diff table, **Then** each execution returns only entries added since the prior scheduled execution

---

### Edge Cases

- **System Clock Changes**: What happens when the system clock is adjusted backward (e.g., NTP correction)? Should entries with timestamps "in the future" relative to last execution time be included or excluded?
- **Browser History Deletion**: How does the system handle cases where users delete browser history entries that fall within the time window since last execution?
- **Concurrent Queries**: What happens when multiple osqueryd queries execute simultaneously against the same browser history source? Should they share the same last execution timestamp or maintain separate timestamps?
- **Extension Initialization Failure**: How does the system behave if the extension cannot initialize (e.g., database unavailable, permissions error) and cannot read/write execution timestamps?
- **Very Large Time Gaps**: What happens when a very long time passes between queries (e.g., weeks or months), resulting in a massive volume of new browser history entries?
- **Browser-Specific Timestamp Precision**: How does the system handle differences in timestamp precision across browsers (Chrome uses microseconds, Firefox uses seconds)?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST persist the last query execution timestamp for each browser history data source (Chrome, Firefox, Safari, Edge)
- **FR-002**: System MUST update the stored execution timestamp atomically after each successful query execution
- **FR-003**: System MUST filter browser history entries to return only those with timestamps strictly greater than the last execution timestamp
- **FR-004**: System MUST handle the first execution case (no previous timestamp) by returning all available browser history entries
- **FR-005**: System MUST maintain separate execution timestamps for different browser history sources to prevent cross-contamination
- **FR-006**: System MUST preserve execution timestamps across osqueryd restarts and extension reloads
- **FR-007**: System MUST expose a dedicated osquery table that automatically applies time-based filtering without requiring explicit timestamp parameters
- **FR-008**: System MUST ensure that querying the standard browser history table does not affect the execution timestamps used by the periodic diff table
- **FR-009**: System MUST handle cases where no new entries exist since the last execution by returning an empty result set (not an error)
- **FR-010**: System MUST use wall clock time (system time) for execution timestamps to ensure human-readable timestamps that correlate with other system logs and external events, accepting the minimal risk of clock adjustments

### Key Entities

- **Execution Timestamp Record**: Represents the last time a query was executed for a specific browser history source. Attributes include: browser type (Chrome/Firefox/Safari/Edge), timestamp value, table identifier (to distinguish periodic diff table from standard table).
- **Browser History Entry**: Represents a single browsing event with timestamp, URL, title, visit count. Relationship: filtered by Execution Timestamp Record to determine inclusion in query results.
- **Periodic Diff Table**: Virtual table that combines browser history entries with execution timestamp filtering logic, exposing a simplified interface for time-based queries.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Operators can verify query execution status by inspecting execution timestamps, even when no browser history changes occurred (100% visibility into monitoring health)
- **SC-002**: Query result sets contain only browser history entries added since the last query execution (0% duplicate entries across consecutive queries under normal operation)
- **SC-003**: System correctly handles first-run scenarios by returning all available history entries (testable by resetting execution timestamps and verifying full dataset retrieval)
- **SC-004**: Execution timestamps persist across osqueryd restarts (testable by stopping/starting osqueryd and verifying timestamp continuity)
- **SC-005**: Querying the periodic diff table requires no manual timestamp parameter management (operators use simple SELECT statements without WHERE clauses for time filtering)
- **SC-006**: Scheduled queries against the periodic diff table produce regular output at configured intervals (e.g., every 5 minutes), even when no new browser history entries exist (empty result sets still generate log entries confirming query execution)

## Assumptions

- Browser history databases are accessible via existing osquery browser history table mechanisms
- The osquery extension has persistent storage capability for maintaining execution timestamps (e.g., SQLite database or file-based storage)
- System clock is generally reliable (major clock adjustments are rare edge cases)
- Browser history timestamp formats can be normalized to a common representation for comparison
- Osquery extension execution model allows maintaining state between query invocations

## Out of Scope

- Historical playback of browser history changes (this feature only tracks "new since last query", not a complete change history)
- Browser history data transformation or enrichment beyond timestamp-based filtering
- Real-time monitoring or push-based notifications (remains pull-based query model)
- Retention policies or archival of old browser history entries
- Browser history data validation or integrity checking
- Multi-host coordination (each osqueryd instance maintains independent execution timestamps)

## Dependencies

- Existing osquery browser history table functionality must be operational
- Persistent storage mechanism (likely SQLite or file-based) for execution timestamps
- Go osquery extension framework (osquery-go or similar) for table implementation

## Risks

- **Clock Skew**: If system clock is adjusted backward, entries may be incorrectly filtered. Mitigation: Consider monotonic time or detect large clock jumps.
- **Storage Failure**: If execution timestamp storage fails, system may fall back to returning all entries (safe but potentially high volume). Mitigation: Implement robust error handling and logging.
- **Performance Impact**: Very large time gaps between queries could result in massive result sets. Mitigation: Consider optional result set limits or pagination in future iterations.
