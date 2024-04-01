package tools

import "time"

func TimestampToTime(sec *int) *time.Time {
	if sec == nil {
		return nil
	}
	utcTime := time.Unix(int64(*sec), 0).UTC()
	return &utcTime
}
