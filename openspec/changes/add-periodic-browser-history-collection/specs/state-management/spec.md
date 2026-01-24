# State Management Specification

## Purpose
Defines the requirements for managing collection state to support incremental queries.

## ADDED Requirements

### Requirement: State Persistence
The system MUST persist the last collection timestamp to enable incremental collection across restarts.

#### Scenario: State File Creation
**Given** the extension runs for the first time
**When** state management is initialized
**Then** a new state file is created with default values

#### Scenario: State File Update
**Given** a successful collection completes at time T
**When** the state is updated
**Then** the state file contains timestamp T as the last collection time

#### Scenario: State File Read
**Given** an existing state file with timestamp T
**When** the extension starts
**Then** it reads timestamp T as the last collection time

### Requirement: Profile-specific State Tracking
The system MUST support tracking last collection time per browser profile.

#### Scenario: Multiple Profiles
**Given** Chrome with 2 profiles and Firefox with 1 profile
**When** collection runs for each profile
**Then** the state file tracks separate timestamps for each profile

#### Scenario: Profile-specific Incremental Collection
**Given** state shows Chrome profile A was collected at T1 and profile B at T2
**When** incremental collection runs
**Then** each profile collects entries since its own last collection time

### Requirement: Concurrent Access Safety
The system MUST handle concurrent access to the state file safely.

#### Scenario: Concurrent Reads
**Given** multiple queries reading state simultaneously
**When** state is accessed concurrently
**Then** all reads succeed without corruption

#### Scenario: Concurrent Read/Write
**Given** a query reading state while collection updates it
**When** concurrent access occurs
**Then** the state remains consistent

### Requirement: Error Recovery
The system MUST handle state file errors gracefully.

#### Scenario: Corrupted State File
**Given** a state file with invalid JSON
**When** the extension tries to read it
**Then** it resets to default state and logs a warning

#### Scenario: Missing State File
**Given** no state file exists
**When** the extension starts
**Then** it creates a new state file with default values

#### Scenario: Permission Errors
**Given** insufficient permissions to write state file
**When** collection tries to update state
**Then** it logs an error and continues without state persistence

## MODIFIED Requirements

*None - State management is a new capability*

## REMOVED Requirements

*None - State management is a new capability*