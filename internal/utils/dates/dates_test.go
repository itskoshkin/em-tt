package dates

import "testing"

//

import (
	"time"
)

func TestString2Date(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantYear  int
		wantMonth time.Month
		wantErr   bool
	}{
		{
			name:      "valid date",
			input:     "01-2024",
			wantYear:  2024,
			wantMonth: time.January,
			wantErr:   false,
		},
		{
			name:      "valid date december",
			input:     "12-2025",
			wantYear:  2025,
			wantMonth: time.December,
			wantErr:   false,
		},
		{
			name:      "valid date with spaces",
			input:     "  07-2023  ",
			wantYear:  2023,
			wantMonth: time.July,
			wantErr:   false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "invalid format - wrong separator",
			input:   "01/2024",
			wantErr: true,
		},
		{
			name:    "invalid format - YYYY-MM",
			input:   "2024-01",
			wantErr: true,
		},
		{
			name:    "invalid month",
			input:   "13-2024",
			wantErr: true,
		},
		{
			name:    "invalid month zero",
			input:   "00-2024",
			wantErr: true,
		},
		{
			name:    "non-numeric",
			input:   "ab-cdef",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := String2Date(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("String2Date(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("String2Date(%q) unexpected error: %v", tt.input, err)
				return
			}

			if got.Year() != tt.wantYear {
				t.Errorf("String2Date(%q) year = %d, want %d", tt.input, got.Year(), tt.wantYear)
			}
			if got.Month() != tt.wantMonth {
				t.Errorf("String2Date(%q) month = %v, want %v", tt.input, got.Month(), tt.wantMonth)
			}
			if got.Day() != 1 {
				t.Errorf("String2Date(%q) day = %d, want 1", tt.input, got.Day())
			}
		})
	}
}
