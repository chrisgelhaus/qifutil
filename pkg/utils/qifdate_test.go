package utils

import "testing"

func TestParseQIFDate(t *testing.T) {
	tests := []struct {
		name                   string
		month, day, sep, year  string
		want                   string
	}{
		{"slash means the 1900s", "3", "15", "/", "99", "1999-03-15"},
		{"apostrophe means the 2000s", "1", "5", "'", "23", "2023-01-05"},
		{"apostrophe is honoured for high years", "3", "15", "'", "99", "2099-03-15"},
		{"slash is honoured for low years", "3", "15", "/", "23", "1923-03-15"},
		{"four digit year carries its own century", "3", "15", "/", "1999", "1999-03-15"},
		{"four digit year after an apostrophe", "3", "15", "'", "2023", "2023-03-15"},
		{"quicken pads narrow fields with spaces", " 3", " 15", "'", " 5", "2005-03-15"},
		{"single digit year", "12", "31", "'", "0", "2000-12-31"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseQIFDate(tt.month, tt.day, tt.sep, tt.year)
			if err != nil {
				t.Fatalf("ParseQIFDate(%q, %q, %q, %q) returned error: %v",
					tt.month, tt.day, tt.sep, tt.year, err)
			}
			if formatted := got.Format("2006-01-02"); formatted != tt.want {
				t.Errorf("ParseQIFDate(%q, %q, %q, %q) = %s, want %s",
					tt.month, tt.day, tt.sep, tt.year, formatted, tt.want)
			}
		})
	}
}

func TestParseQIFDateRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name                  string
		month, day, sep, year string
	}{
		{"month above twelve", "13", "15", "'", "23"},
		{"month of zero", "0", "15", "'", "23"},
		{"day above thirty one", "3", "32", "'", "23"},
		{"day of zero", "3", "0", "'", "23"},
		{"unknown separator", "3", "15", "x", "23"},
		{"empty year", "3", "15", "'", ""},
		{"non numeric month", "abc", "15", "'", "23"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseQIFDate(tt.month, tt.day, tt.sep, tt.year); err == nil {
				t.Errorf("ParseQIFDate(%q, %q, %q, %q) succeeded, want an error",
					tt.month, tt.day, tt.sep, tt.year)
			}
		})
	}
}
