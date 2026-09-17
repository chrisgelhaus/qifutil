package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseQIFDate builds a date from the parts of a QIF date field.
//
// Quicken marks the century with the separator that precedes the year: a slash
// for the 1900s and an apostrophe for the 2000s, so "3/15/99" is 1999 while
// "3/15'99" is 2099. A four digit year carries its own century and the
// separator is then only a delimiter. Fields may be padded with spaces, as in
// "D 3/15' 5".
func ParseQIFDate(month, day, sep, year string) (time.Time, error) {
	monthNum, err := atoiField("month", month)
	if err != nil {
		return time.Time{}, err
	}
	dayNum, err := atoiField("day", day)
	if err != nil {
		return time.Time{}, err
	}
	yearField := strings.TrimSpace(year)
	yearNum, err := atoiField("year", yearField)
	if err != nil {
		return time.Time{}, err
	}

	if monthNum < 1 || monthNum > 12 {
		return time.Time{}, fmt.Errorf("month %d out of range", monthNum)
	}
	if dayNum < 1 || dayNum > 31 {
		return time.Time{}, fmt.Errorf("day %d out of range", dayNum)
	}

	// A four digit year is already absolute; a shorter one takes its century
	// from the separator.
	if len(yearField) < 4 {
		switch strings.TrimSpace(sep) {
		case "/":
			yearNum += 1900
		case "'":
			yearNum += 2000
		default:
			return time.Time{}, fmt.Errorf("unrecognised year separator %q", sep)
		}
	}

	return time.Date(yearNum, time.Month(monthNum), dayNum, 0, 0, 0, 0, time.UTC), nil
}

// atoiField parses one space padded numeric field of a QIF date.
func atoiField(name, value string) (int, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, fmt.Errorf("%s is empty", name)
	}
	number, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, fmt.Errorf("%s %q is not a number", name, trimmed)
	}
	return number, nil
}
