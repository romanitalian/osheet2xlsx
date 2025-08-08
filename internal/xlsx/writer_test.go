package xlsx

import "testing"

func TestColumnName(t *testing.T) {
	cases := map[int]string{1: "A", 2: "B", 26: "Z", 27: "AA", 52: "AZ", 53: "BA", 702: "ZZ", 703: "AAA"}
	for k, v := range cases {
		if got := columnName(k); got != v {
			t.Fatalf("columnName(%d) = %s, want %s", k, got, v)
		}
	}
}
