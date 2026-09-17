package utils

import "testing"

func TestSanitizeFileName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"a name needing no change", "Normal Account", "Normal Account"},
		{"forward slash", "Amex / Joint", "Amex _ Joint"},
		{"back slash", "Back\\Slash", "Back_Slash"},
		{"colon, which NTFS reads as an alternate data stream", "Fidelity: Roth", "Fidelity_ Roth"},
		{"asterisk and question mark", "Savings*2024?", "Savings_2024_"},
		{"pipe", "Joint|Acct", "Joint_Acct"},
		{"double quote", "Quote\"Name", "Quote_Name"},
		{"angle brackets", "Angle<>Name", "Angle__Name"},
		{"control characters", "Tab\tName", "Tab_Name"},
		{"an empty name still yields something usable", "", "account"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeFileName(tt.input); got != tt.want {
				t.Errorf("SanitizeFileName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
