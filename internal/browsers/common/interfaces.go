package common

// BrowserRoot is a browser data root owned by one operating-system user.
type BrowserRoot struct {
	Path           string
	OSUserName     string
	BrowserFamily  string
	BrowserVariant string
}

// Profile is the current inventory and metadata for one machine-local browser profile.
type Profile struct {
	ProfileID          string
	BrowserFamily      string
	BrowserVariant     string
	OSUserName         string
	Directory          string
	Path               string
	DisplayName        string
	DisplayNamePresent bool
	Account            string
	AccountPresent     bool
}

// VisitObservation is the stable portion of one native browser visit row.
type VisitObservation struct {
	ObservationID  string
	ProfileID      string
	BrowserFamily  string
	BrowserVariant string
	NativeVisitID  int64
	NativeURLID    int64
	VisitTime      int64
	VisitTimeUS    int64
}

// HistoryPage is the browser's current aggregate state for one URL/page row.
type HistoryPage struct {
	ProfileID            string
	NativeURLID          int64
	BrowserFamily        string
	BrowserVariant       string
	URL                  string
	Title                string
	TitlePresent         bool
	VisitCount           int64
	LastVisitTime        int64
	LastVisitTimePresent bool
	Hidden               int64
	HiddenPresent        bool
}
