# Core Extension Specification

## Purpose
Defines the core functionality and behavior of the osquery browser history extension.

## ADDED Requirements

### Requirement: Incremental Collection Support
The extension MUST support incremental collection by accepting time-based filters in queries.

#### Scenario: Time-based Filtering
**Given** a query with WHERE visit_time > 1234567890
**When** the extension processes the query
**Then** only history entries newer than the timestamp are returned

#### Scenario: Backward Compatibility
**Given** a query without time-based filters
**When** the extension processes the query
**Then** all history entries are returned as before

### Requirement: Enhanced Virtual Table Schema
The extension MUST provide additional columns for efficient incremental queries.

#### Scenario: Visit Time Column
**Given** the virtual table schema includes visit_time column
**When** querying with time-based filters
**Then** the column can be used for efficient filtering

#### Scenario: Additional Metadata Columns
**Given** the virtual table includes visit_count and browser_variant columns
**When** querying history data
**Then** additional metadata is available for analysis

## MODIFIED Requirements

### Requirement: History Collection Interface
**Original**: FindHistory(profile) returns all history entries
**Modified**: FindHistory(profile, options) supports optional time-based filtering

The browser history collection functions SHALL support optional parameters for incremental collection.

#### Scenario: Incremental Collection
**Given** a profile with history entries
**When** calling FindHistory with since parameter
**Then** only entries after the specified time are returned

#### Scenario: Full Collection
**Given** a profile with history entries
**When** calling FindHistory without since parameter
**Then** all entries are returned for backward compatibility

## REMOVED Requirements

*None - This change adds functionality without removing existing capabilities*