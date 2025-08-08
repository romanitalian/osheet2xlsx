package osheet

import "testing"

func TestInferCell_GoldenNumbersAndDates(t *testing.T) {
	cases := []struct {
		in   string
		want ValueType
	}{
		{"1\u00A0234,56", ValueNumber},
		{"1 234 567.89", ValueNumber},
		{"$1,234.50", ValueNumber},
		{"(1 234,50)", ValueNumber},
		{"12%", ValueNumber},
		{"02.01.2024", ValueDateTime},
		{"02/01/2024", ValueDateTime},
		{"2024-01-02T12:34:56+03:00", ValueDateTime},
		{"1704067200000", ValueDateTime},
	}
	for _, tc := range cases {
		if got := inferCell(tc.in); got.Type != tc.want {
			t.Fatalf("inferCell(%q) = %v, want %v", tc.in, got.Type, tc.want)
		}
	}
}
