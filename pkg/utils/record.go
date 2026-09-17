package utils

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// TransactionFields holds the raw field values of one QIF transaction record.
// Values are as they appeared in the file, with surrounding space removed.
type TransactionFields struct {
	Date     string // D
	Amount   string // U, falling back to T
	Cleared  string // C
	Number   string // N
	Payee    string // P
	Memo     string // M
	Category string // L
}

// SplitRecords divides a QIF block into records on the "^" terminator that
// ends each one. Blank records are dropped, so trailing content after the
// final terminator does not produce an empty entry.
func SplitRecords(block string) []string {
	var records []string
	var current []string

	flush := func() {
		record := strings.Join(current, "\n")
		current = nil
		if strings.TrimSpace(record) != "" {
			records = append(records, record)
		}
	}

	for _, line := range strings.Split(block, "\n") {
		if strings.TrimSpace(line) == "^" {
			flush()
			continue
		}
		current = append(current, line)
	}
	flush()

	return records
}

// ParseTransactionRecord reads the field lines of a single QIF transaction
// record. QIF is line oriented: each line carries a one letter field code
// followed by its value. Field order is not significant and most fields are
// optional, so unknown codes are ignored rather than treated as an error.
func ParseTransactionRecord(record string) TransactionFields {
	var fields TransactionFields
	var fallbackAmount string
	var haveAmount bool

	for _, line := range strings.Split(record, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}

		code := line[0]
		value := strings.TrimSpace(line[1:])

		switch code {
		case 'D':
			fields.Date = value
		case 'U':
			// U is Quicken's own amount field and wins over T when both appear.
			fields.Amount = value
			haveAmount = true
		case 'T':
			fallbackAmount = value
		case 'C':
			fields.Cleared = value
		case 'N':
			fields.Number = value
		case 'P':
			fields.Payee = value
		case 'M':
			fields.Memo = value
		case 'L':
			fields.Category = value
		}
	}

	if !haveAmount {
		fields.Amount = fallbackAmount
	}

	return fields
}

// qifDateField matches a QIF date such as "3/15/99", "1/5'23" or "12/31/1985",
// capturing the separator that precedes the year because it carries the
// century. Quicken pads narrow fields with spaces.
var qifDateField = regexp.MustCompile(
	`^\s*(\d{1,2})\/\s*(\d{1,2})(['/])\s*(\d{1,4})\s*$`)

// ParseQIFDateField parses a whole QIF date field, such as the value of a D line.
func ParseQIFDateField(field string) (time.Time, error) {
	parts := qifDateField.FindStringSubmatch(field)
	if parts == nil {
		return time.Time{}, fmt.Errorf("%q is not a QIF date", field)
	}
	return ParseQIFDate(parts[1], parts[2], parts[3], parts[4])
}
