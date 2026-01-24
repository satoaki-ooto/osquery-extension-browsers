# Storage Specification

## Purpose
Defines the storage requirements for collected browser history data.

## ADDED Requirements

### Requirement: Data Persistence
The system MUST persist collected browser history data with timestamps and metadata.

#### Scenario: Basic Storage
**Given** a browser history entry with URL "https://example.com", title "Example", and visit time
**When** the collection service stores the entry
**Then** the entry is persisted with URL, title, visit time, browser type, and profile
**And** a collection timestamp is added

#### Scenario: Deduplication
**Given** a history entry that already exists in storage
**When** the same entry is collected again
**Then** the duplicate is not stored
**And** the existing entry's last_seen timestamp is updated

### Requirement: Incremental Collection Tracking
The system MUST track the last collection time per browser profile to enable efficient incremental collection.

#### Scenario: Last Collection Tracking
**Given** a Chrome profile at path "/home/user/.config/google-chrome/Default"
**When** collection completes for this profile
**Then** the system stores the last collection timestamp
**And** subsequent collections start from this timestamp

#### Scenario: First Collection
**Given** a profile that has never been collected
**When** collection runs
**Then** all available history is collected
**And** the last collection timestamp is set to the current time

### Requirement: Data Retention
The system MUST automatically remove data older than the configured retention period.

#### Scenario: Retention Policy Execution
**Given** a retention policy of 30 days
**And** data exists that is 31 days old
**When** the cleanup process runs
**Then** the 31-day-old data is removed
**And** data 30 days old or newer is retained

#### Scenario: No Retention Limit
**Given** retention_days=0 (no limit)
**When** cleanup runs
**Then** no data is removed

### Requirement: Query Interface
The system MUST provide efficient query interfaces for accessing collected data.

#### Scenario: Time Range Query
**Given** stored history data across multiple days
**When** querying for data between specific timestamps
**Then** only data within the range is returned
**And** results are ordered by visit time

#### Scenario: Browser Filtering
**Given** stored data from Chrome and Firefox
**When** querying for only Chrome data
**Then** only Chrome history entries are returned

#### Scenario: Profile Filtering
**Given** stored data from multiple profiles
**When** querying for a specific profile
**Then** only data from that profile is returned

## MODIFIED Requirements

*None - Storage is a new capability*

## REMOVED Requirements

*None - Storage is a new capability*