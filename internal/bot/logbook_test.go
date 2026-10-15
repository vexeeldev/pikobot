package bot

import (
	"testing"
)

func TestParseJamInput(t *testing.T) {
	tests := []struct {
		input       string
		wantMulai   string
		wantSelesai string
		wantErr     bool
	}{
		{"08.00 - 16.00", "08.00", "16.00", false},
		{"08:00 - 16:00", "08.00", "16.00", false},
		{"8.00 - 16.00", "08.00", "16.00", false},
		{"08.30-17.00", "08.30", "17.00", false},
		{"09.00 sd 17.00", "09.00", "17.00", false},
		{"invalid format", "", "", true},
	}

	for _, tt := range tests {
		mulai, selesai, err := ParseJamInput(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseJamInput(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr {
			if mulai != tt.wantMulai || selesai != tt.wantSelesai {
				t.Errorf("ParseJamInput(%q) = (%q, %q), want (%q, %q)", tt.input, mulai, selesai, tt.wantMulai, tt.wantSelesai)
			}
		}
	}
}
