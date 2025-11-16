package day

import "strings"

// MaxDays returns the maximum amount of days available for the provided month and year.
func MaxDays(month string, year int) int {
	upper := strings.ToUpper(month)
	switch upper {
	case "APRIL", "JUNE", "SEPTEMBER", "NOVEMBER":
		return 30
	case "JANUARY", "MARCH", "MAY", "JULY", "AUGUST", "OCTOBER", "DECEMBER":
		return 31
	case "FEBRUARY":
		if isLeapYear(year) {
			return 29
		}
		return 28
	default:
		return 31
	}
}

func isLeapYear(year int) bool {
	if year%400 == 0 {
		return true
	}

	if year%100 == 0 {
		return false
	}

	return year%4 == 0
}
