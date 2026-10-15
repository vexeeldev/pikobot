package utils

import "testing"

func TestFormatRupiah(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "Rp 0"},
		{500, "Rp 500"},
		{5000, "Rp 5.000"},
		{50000, "Rp 50.000"},
		{1500000, "Rp 1.500.000"},
		{1000000000, "Rp 1.000.000.000"},
		{-50000, "-Rp 50.000"},
	}

	for _, tt := range tests {
		got := FormatRupiah(tt.input)
		if got != tt.expected {
			t.Errorf("FormatRupiah(%d) = %s, expected %s", tt.input, got, tt.expected)
		}
	}
}

func TestParseRupiah(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		hasError bool
	}{
		{"50000", 50000, false},
		{"50.000", 50000, false},
		{"Rp 50.000", 50000, false},
		{"rp 50000", 50000, false},
		{"50k", 50000, false},
		{"100rb", 100000, false},
		{"50ribu", 50000, false},
		{"1jt", 1000000, false},
		{"1.5jt", 1500000, false},
		{"2.5juta", 2500000, false},
		{"abc", 0, true},
	}

	for _, tt := range tests {
		got, err := ParseRupiah(tt.input)
		if (err != nil) != tt.hasError {
			t.Errorf("ParseRupiah(%s) error = %v, expected error: %v", tt.input, err, tt.hasError)
		}
		if err == nil && got != tt.expected {
			t.Errorf("ParseRupiah(%s) = %d, expected %d", tt.input, got, tt.expected)
		}
	}
}
