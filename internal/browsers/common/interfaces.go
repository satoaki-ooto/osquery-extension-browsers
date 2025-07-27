package common

import (
	"time"
)

// Browser represents a web browser with its associated data
type Browser interface {
	// Name returns the name of the browser
	Name() string

	// Variant returns the specific variant of the browser (e.g., Chrome, Edge, Chromium)
	Variant() string

	// FindProfiles discovers all profiles for this browser
	FindProfiles() ([]Profile, error)

	// FindHistory discovers history entries for a specific profile
	FindHistory(profile Profile) ([]HistoryEntry, error)
}

// Profile represents a browser profile with its associated data
type Profile struct {
	// ID is the unique identifier for the profile
	ID string

	// Name is the display name of the profile
	Name string

	// Path is the file system path to the profile directory
	Path string

	// Email is the email address associated with the profile (if available)
	Email string

	// BrowserType is the type of browser this profile belongs to
	BrowserType string

	// BrowserVariant is the specific variant of the browser
	BrowserVariant string
}

// HistoryEntry represents a single entry in the browser history
type HistoryEntry struct {
	// ID is the unique identifier for the history entry
	ID int64

	// URL is the URL of the visited page
	URL string

	// Title is the title of the visited page
	Title string

	// VisitTime is the time when the page was visited
	VisitTime time.Time

	// VisitCount is the number of times the page was visited
	VisitCount int

	// ProfileID is the ID of the profile this entry belongs to
	ProfileID string

	// BrowserType is the type of browser this entry belongs to
	BrowserType string

	// BrowserVariant is the specific variant of the browser
	BrowserVariant string
}
