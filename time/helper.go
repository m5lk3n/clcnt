package time

import (
	"strconv"
	"time"
)

// DefaultTimestamp tries to convert given string into timestamp, defaults to current Unix epoch time
func DefaultTimestamp(s string) int64 {
	s = s[1:] // chop leading /

	n, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return n
	}

	return time.Now().Unix()
}

func getStartOfTodayAsUnix() int64 {
	t := time.Now()
	m := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	return m.Unix()
}

// GetDaysAgoAsUnix converts given number of days into the Unix Epoch time
func GetDaysAgoAsUnix(d int) int64 {
	t := getStartOfTodayAsUnix()

	return t - int64((d-1)*24*60*60)
}
