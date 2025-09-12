package utils

import "testing"

func TestSplitCategoryAndTag(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantCategory string
		wantTag      string
	}{
		{
			name:         "no tag",
			input:        "Food:Groceries",
			wantCategory: "Food:Groceries",
			wantTag:      "",
		},
		{
			name:         "with tag",
			input:        "Food:Groceries/Monthly",
			wantCategory: "Food:Groceries",
			wantTag:      "Monthly",
		},
		{
			name:         "empty string",
			input:        "",
			wantCategory: "",
			wantTag:      "",
		},
		{
			name:         "multiple slashes",
			input:        "Food:Groceries/Monthly/Extra",
			wantCategory: "Food:Groceries",
			wantTag:      "Monthly/Extra",
		},
		{
			name:         "with whitespace",
			input:        "Food:Groceries / Monthly ",
			wantCategory: "Food:Groceries",
			wantTag:      "Monthly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCategory, gotTag := SplitCategoryAndTag(tt.input)
			if gotCategory != tt.wantCategory {
				t.Errorf("SplitCategoryAndTag() category = %q, want %q", gotCategory, tt.wantCategory)
			}
			if gotTag != tt.wantTag {
				t.Errorf("SplitCategoryAndTag() tag = %q, want %q", gotTag, tt.wantTag)
			}
		})
	}
}
