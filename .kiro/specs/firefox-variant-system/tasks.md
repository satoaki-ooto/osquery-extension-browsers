# Implementation Plan

- [x] 1. Create Firefox variant system core structure
  - Create `internal/browsers/firefox/variants.go` file with BrowserVariant struct and DetectBrowserVariants function
  - Implement platform-specific variant detection logic for Windows, macOS, and Linux
  - Add support for Firefox, Firefox ESR, Firefox Developer Edition, and Firefox Nightly variants
  - _Requirements: 1.1, 1.2, 4.1, 4.2_

- [ ] 2. Update Firefox finder to use variant system
  - Modify `internal/browsers/firefox/finder.go` to utilize the new variant system
  - Replace hardcoded paths with variant-based path detection
  - Ensure backward compatibility with existing functionality
  - _Requirements: 4.3_

- [ ] 3. Update Firefox history retrieval with variant information
  - Modify `internal/browsers/firefox/history.go` to include variant information in HistoryEntry
  - Update FindHistory function to accept variant parameter and set BrowserVariant field
  - Ensure each history entry contains the correct variant name
  - _Requirements: 2.1, 2.2, 4.3_

- [ ] 4. Update Firefox profile discovery with variant information
  - Modify `internal/browsers/firefox/profile.go` to include variant information in Profile
  - Update FindProfiles function to work with variant system and set BrowserVariant field
  - Ensure each profile contains the correct variant name
  - _Requirements: 3.1, 3.2, 4.3_

- [ ] 5. Create unit tests for variant detection
  - Write tests for DetectBrowserVariants function covering all platforms
  - Test variant detection when directories exist and don't exist
  - Test handling of multiple variants on the same system
  - _Requirements: 1.1, 1.2, 1.3_

- [ ] 6. Create unit tests for history retrieval with variants
  - Write tests for updated FindHistory function with variant information
  - Test that BrowserVariant field is correctly set in HistoryEntry
  - Test error handling when variant-specific databases are inaccessible
  - _Requirements: 2.1, 2.2, 2.3, 2.4_

- [ ] 7. Create unit tests for profile discovery with variants
  - Write tests for updated FindProfiles function with variant information
  - Test that BrowserVariant field is correctly set in Profile
  - Test handling of missing profiles.ini files for different variants
  - _Requirements: 3.1, 3.2, 3.3, 3.4_

- [ ] 8. Update common detector to support Firefox variants
  - Modify `internal/browsers/common/detector.go` to detect Firefox variants
  - Add functions to check for different Firefox variant installations
  - Ensure integration with existing browser detection system
  - _Requirements: 1.1, 4.4_

- [ ] 9. Create integration tests for complete variant system
  - Write end-to-end tests that verify variant detection, profile discovery, and history retrieval work together
  - Test scenarios with multiple Firefox variants installed
  - Test cross-platform compatibility
  - _Requirements: 1.1, 2.4, 3.4, 4.4_

- [ ] 10. Validate backward compatibility and refactor cleanup
  - Ensure all existing Firefox functionality continues to work
  - Remove any deprecated code or unused functions
  - Update documentation and comments to reflect variant system
  - _Requirements: 4.3, 4.4_