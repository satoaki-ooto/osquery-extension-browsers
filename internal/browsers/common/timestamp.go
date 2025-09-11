package common

import (
	"time"
)

// WebKitTimestampToTime converts a WebKit timestamp to a time.Time
func WebKitTimestampToTime(webkitTimestamp int64) time.Time {
	// WebKit timestamps are in microseconds since January 1, 1601
	// Convert to nanoseconds since January 1, 1970
	unixTimestamp := (webkitTimestamp - 11644473600000000) * 1000
	return time.Unix(0, unixTimestamp)
}

// UnixTimestampToTime converts a Unix timestamp to a time.Time
func UnixTimestampToTime(unixTimestamp int64) time.Time {
	return time.Unix(unixTimestamp, 0)
}
