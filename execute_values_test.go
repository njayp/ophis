package ophis

import (
	"encoding/json"
	"testing"
)

func TestFormatFlagValueKeepsLargeIntegersIntact(t *testing.T) {
	cases := map[any]string{
		float64(9588350222588):       "9588350222588",
		float64(42):                  "42",
		float64(-7):                  "-7",
		float64(0):                   "0",
		float64(1.5):                 "1.5",
		float64(1e21):                "1000000000000000000000",
		json.Number("9588350222588"): "9588350222588",
		"already-a-string":           "already-a-string",
		int64(12):                    "12",
	}
	for in, want := range cases {
		if got := formatFlagValue(in); got != want {
			t.Errorf("formatFlagValue(%#v) = %q, want %q", in, got, want)
		}
	}
}

func TestParseFlagArgValueUsesPlainIntegers(t *testing.T) {
	got := parseFlagArgValue("productId", float64(9588350222588))
	if len(got) != 2 || got[1] != "9588350222588" {
		t.Fatalf("got %v", got)
	}
}
