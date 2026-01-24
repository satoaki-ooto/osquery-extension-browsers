# History Collection Specification

## Purpose
Defines the requirements for browser history collection with incremental support.

## ADDED Requirements

### Requirement: Time-based Filtering
The history collection functions MUST support filtering by visit time to enable incremental collection.

#### Scenario: Chromium Incremental Collection
**Given** a Chrome profile with history entries
**When** FindHistory is called with since parameter set to 1 hour ago
**Then** only entries visited within the last hour are returned

#### Scenario: Firefox Incremental Collection
**Given** a Firefox profile with history entries
**When** FindHistory is called with since parameter set to 1 hour ago
**Then** only entries visited within the last hour are returned

#### Scenario: No Matching Entries
**Given** a profile with no entries newer than the since timestamp
**When** FindHistory is called with that since parameter
**Then** an empty slice is returned

### Requirement: Optional Parameters
The history collection functions MUST support optional parameters for flexible querying.

#### Scenario: Functional Options Pattern
**Given** the FindHistory function accepts optional parameters
**When** calling with WithSince(t) option
**Then** the function filters by the specified time

#### Scenario: Multiple Options
**Given** the FindHistory function accepts multiple options
**When** calling with WithSince(t) and WithLimit(n) options
**Then** the function applies both filters

### Requirement: Efficient Querying
The history collection functions MUST use database indexes for efficient time-based filtering.

#### Scenario: Indexed Timestamp Query
**Given** a browser database with indexed timestamp column
**When** querying with since parameter
**Then** the database uses the index for efficient retrieval

#### Scenario: Large Database Performance
**Given** a browser database with 100,000+ history entries
**When** querying with since parameter for recent entries
**Then** query performance remains acceptable (< 1 second)

## MODIFIED Requirements

### Requirement: History Collection Signature
**Original**: FindHistory(profile) returns all history entries
**Modified**: FindHistory(profile, options) supports optional filtering

The history collection functions SHALL maintain backward compatibility while adding optional parameters.

#### Scenario: Backward Compatibility
**Given** existing code calling FindHistory(profile) without options
**When** the function is executed
**Then** it returns all history entries as before

#### Scenario: New Optional Parameters
**Given** updated FindHistory function with options support
**When** calling without options (nil or empty)
**Then** it behaves identically to the original function

## REMOVED Requirements

*None - This change adds functionality without removing existing capabilities*