package common

import (
	"testing"
	"time"
)

func TestApplyHistoryOptions(t *testing.T) {
	tests := []struct {
		name     string
		opts     []HistoryOption
		wantSince *time.Time
		wantLimit int
	}{
		{
			name:     "no options",
			opts:     nil,
			wantSince: nil,
			wantLimit: 0,
		},
		{
			name:     "empty options",
			opts:     []HistoryOption{},
			wantSince: nil,
			wantLimit: 0,
		},
		{
			name:     "with since",
			opts:     []HistoryOption{WithSince(time.Unix(1234567890, 0))},
			wantSince: timePtr(time.Unix(1234567890, 0)),
			wantLimit: 0,
		},
		{
			name:     "with limit",
			opts:     []HistoryOption{WithLimit(100)},
			wantSince: nil,
			wantLimit: 100,
		},
		{
			name: "with both since and limit",
			opts: []HistoryOption{
				WithSince(time.Unix(1234567890, 0)),
				WithLimit(50),
			},
			wantSince: timePtr(time.Unix(1234567890, 0)),
			wantLimit: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyHistoryOptions(tt.opts)

			if got.Since == nil && tt.wantSince != nil {
				t.Errorf("Since = nil, want %v", tt.wantSince)
			} else if got.Since != nil && tt.wantSince == nil {
				t.Errorf("Since = %v, want nil", got.Since)
			} else if got.Since != nil && tt.wantSince != nil && !got.Since.Equal(*tt.wantSince) {
				t.Errorf("Since = %v, want %v", got.Since, tt.wantSince)
			}

			if got.Limit != tt.wantLimit {
				t.Errorf("Limit = %d, want %d", got.Limit, tt.wantLimit)
			}
		})
	}
}

func TestWithSince(t *testing.T) {
	testTime := time.Unix(1234567890, 0)
	opt := WithSince(testTime)

	options := &HistoryOptions{}
	opt(options)

	if options.Since == nil {
		t.Fatal("Since was not set")
	}

	if !options.Since.Equal(testTime) {
		t.Errorf("Since = %v, want %v", options.Since, testTime)
	}
}

func TestWithLimit(t *testing.T) {
	opt := WithLimit(100)

	options := &HistoryOptions{}
	opt(options)

	if options.Limit != 100 {
		t.Errorf("Limit = %d, want 100", options.Limit)
	}
}

func TestWithLimitZero(t *testing.T) {
	opt := WithLimit(0)

	options := &HistoryOptions{}
	opt(options)

	if options.Limit != 0 {
		t.Errorf("Limit = %d, want 0", options.Limit)
	}
}

func TestWithLimitNegative(t *testing.T) {
	opt := WithLimit(-10)

	options := &HistoryOptions{}
	opt(options)

	if options.Limit != -10 {
		t.Errorf("Limit = %d, want -10", options.Limit)
	}
}

// Helper function to create a pointer to time.Time
func timePtr(t time.Time) *time.Time {
	return &t
}
