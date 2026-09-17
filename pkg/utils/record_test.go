package utils

import "testing"

func TestParseTransactionRecord(t *testing.T) {
	tests := []struct {
		name   string
		record string
		want   TransactionFields
	}{
		{
			name: "every field present",
			record: `D1/5'23
U-10.00
T-10.00
CX
N1234
PGrocery Store
MWeekly shop
LFood:Groceries`,
			want: TransactionFields{
				Date: "1/5'23", Amount: "-10.00", Cleared: "X", Number: "1234",
				Payee: "Grocery Store", Memo: "Weekly shop", Category: "Food:Groceries",
			},
		},
		{
			name: "amount falls back to T when U is absent",
			record: `D1/6'23
T-20.00
PNo U field`,
			want: TransactionFields{Date: "1/6'23", Amount: "-20.00", Payee: "No U field"},
		},
		{
			name: "U wins when both are present",
			record: `D1/6'23
U-11.11
T-22.22`,
			want: TransactionFields{Date: "1/6'23", Amount: "-11.11"},
		},
		{
			name: "optional cleared and category may be missing",
			record: `D1/8'23
U-40.00
T-40.00
PNo L field`,
			want: TransactionFields{Date: "1/8'23", Amount: "-40.00", Payee: "No L field"},
		},
		{
			name: "field order does not matter",
			record: `D1/9'23
U-50.00
LFood:Dining
PReordered after L`,
			want: TransactionFields{
				Date: "1/9'23", Amount: "-50.00",
				Payee: "Reordered after L", Category: "Food:Dining",
			},
		},
		{
			name: "split and address lines are ignored",
			record: `D1/5'23
T-60.00
PSplit Payee
LFood:Dining
SFood:Groceries
EGroceries portion
$-40.00
SFood:Dining
EDining portion
$-20.00
A123 Main St`,
			want: TransactionFields{
				Date: "1/5'23", Amount: "-60.00",
				Payee: "Split Payee", Category: "Food:Dining",
			},
		},
		{
			name: "carriage returns are trimmed",
			record:  "D1/5'23\r\nT-70.00\r\nPTrailing CR\r",
			want: TransactionFields{Date: "1/5'23", Amount: "-70.00", Payee: "Trailing CR"},
		},
		{
			name:   "a record with no date yields no date",
			record: "NChecking Account\nTBank",
			want:   TransactionFields{Number: "Checking Account", Amount: "Bank"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseTransactionRecord(tt.record); got != tt.want {
				t.Errorf("ParseTransactionRecord() =\n  %+v\nwant\n  %+v", got, tt.want)
			}
		})
	}
}

func TestSplitRecords(t *testing.T) {
	block := `D1/5'23
T-10.00
^
D1/6'23
T-20.00
^
D1/7'23
T-30.00
^`

	records := SplitRecords(block)
	if len(records) != 3 {
		t.Fatalf("SplitRecords returned %d records, want 3: %q", len(records), records)
	}
	for i, want := range []string{"-10.00", "-20.00", "-30.00"} {
		if got := ParseTransactionRecord(records[i]).Amount; got != want {
			t.Errorf("record %d amount = %q, want %q", i, got, want)
		}
	}
}

func TestSplitRecordsIgnoresTrailingContent(t *testing.T) {
	if got := SplitRecords("\n  \n^\n\n"); len(got) != 0 {
		t.Errorf("SplitRecords on blank input returned %d records, want 0: %q", len(got), got)
	}
}

func TestParseQIFDateField(t *testing.T) {
	tests := []struct{ field, want string }{
		{"3/15/99", "1999-03-15"},
		{"1/5'23", "2023-01-05"},
		{" 3/15' 5", "2005-03-15"},
		{"12/31/1985", "1985-12-31"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			got, err := ParseQIFDateField(tt.field)
			if err != nil {
				t.Fatalf("ParseQIFDateField(%q) returned error: %v", tt.field, err)
			}
			if formatted := got.Format("2006-01-02"); formatted != tt.want {
				t.Errorf("ParseQIFDateField(%q) = %s, want %s", tt.field, formatted, tt.want)
			}
		})
	}
}

func TestParseQIFDateFieldRejectsMalformedFields(t *testing.T) {
	for _, field := range []string{"", "not a date", "3-15-23", "3/15", "13/45'23"} {
		if _, err := ParseQIFDateField(field); err == nil {
			t.Errorf("ParseQIFDateField(%q) succeeded, want an error", field)
		}
	}
}
