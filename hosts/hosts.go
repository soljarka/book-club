package hosts

import (
	"time"
)

type Host struct {
	Name string
	Book string
	Date string
}

var hostQueue = []string{"Ольга", "Михаил", "Андрей", "Яна", "Алексей"}
var startDate = time.Date(2024, time.November, 1, 0, 0, 0, 0, time.UTC)
var startHost = 0

var books = map[string]string{"Ольга": "Эльфриде Елинек - \"Пианистка\""}

func GetNthHost(n int) *Host {

	nextDate := time.Now()
	for i := 0; i < n; i++ {
		nextDate = nextSecondTuesday(nextDate)
		nextDate = nextDate.AddDate(0, 0, 1)
	}

	nextDate = nextDate.AddDate(0, 0, -1)
	hostIndex := (monthDiff(startDate, nextDate) + startHost) % len(hostQueue)
	host := hostQueue[hostIndex]
	book := books[host]
	return &Host{
		Name: hostQueue[hostIndex%len(hostQueue)],
		Book: book,
		Date: nextDate.Format("2006-01-02"),
	}
}

func monthDiff(t1, t2 time.Time) int {
	year1, month1, _ := t1.Date()
	year2, month2, _ := t2.Date()
	return (year2-year1)*12 + int(month2-month1)
}

func nextSecondTuesday(date time.Time) time.Time {
	for !isSecondTuesday(date) {
		date = date.AddDate(0, 0, 1)
	}
	return date
}

func isSecondTuesday(t time.Time) bool {
	if t.Weekday() == time.Tuesday && t.Day() > 7 && t.Day() < 15 {
		return true
	}
	return false
}
