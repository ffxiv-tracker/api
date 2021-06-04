package utils

import "time"

func GetFFWeekYear(t time.Time) (int, int) {
	return GetMostRecentTuesday(t).ISOWeek()
}

func GetMostRecentTuesday(t time.Time) time.Time {
	offset := time.Duration(time.Tuesday - t.Weekday())
	if offset > 0 {
		offset -= 7
	}

	return t.Add(offset * 24 * time.Hour)
}
