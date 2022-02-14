package common

import "time"

// DefaultTimeFormat is the time format to use for formatting and parsing time values.
var DefaultTimeFormat = time.RFC3339Nano

// FormatTime format time to string RFC3339
func FormatTime(t time.Time) string {
	return t.Format(DefaultTimeFormat)
}

// GetCurrentTime get time as string formatted RFC3339
func GetCurrentTime() string {
	return time.Now().Format(DefaultTimeFormat)
}

// GetCurrentTimestamp get unix timestamp in milliseconds
func GetCurrentTimestamp() int64 {
	return time.Now().UnixMilli()
}

// TimestampToTime convert unix timestamp to time
func TimestampToTime(timestamp int64) time.Time {
	return time.UnixMilli(timestamp)
}
