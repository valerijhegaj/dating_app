package timestamp

import (
	"strconv"
	"time"
)

func ToTimestamp(age string) string {
	ageInt, err := strconv.Atoi(age)
	if err != nil || ageInt > 150 {
		return ""
	}
	today := time.Now()
	birthDay := today.AddDate(-ageInt, 0, 0)
	return birthDay.Format("2006-01-02")
}

func ToAge(timestamp string) string {
	birthDay, err := time.Parse("2006-01-02", timestamp[:10])
	if err != nil {
		return ""
	}
	today := time.Now()
	ans := strconv.Itoa(int(today.Sub(birthDay).Hours() / 365 / 24))
	return ans
}
