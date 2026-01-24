package common

import "time"

// HistoryOptions defines optional parameters for history collection queries.
type HistoryOptions struct {
	// Since filters history entries to only include entries after this time.
	// If nil, all history entries are returned.
	Since *time.Time

	// Limit restricts the number of history entries returned.
	// If 0 or negative, no limit is applied.
	Limit int
}

// HistoryOption is a functional option for configuring history collection.
type HistoryOption func(*HistoryOptions)

// WithSince creates an option to filter history entries after the specified time.
func WithSince(t time.Time) HistoryOption {
	return func(o *HistoryOptions) {
		o.Since = &t
	}
}

// WithLimit creates an option to limit the number of history entries returned.
func WithLimit(n int) HistoryOption {
	return func(o *HistoryOptions) {
		o.Limit = n
	}
}

// ApplyHistoryOptions applies the given options to a HistoryOptions struct.
// If no options are provided, default options are returned.
func ApplyHistoryOptions(opts []HistoryOption) *HistoryOptions {
	options := &HistoryOptions{}
	for _, opt := range opts {
		opt(options)
	}
	return options
}
