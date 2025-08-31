package ghestimatebot_test

import (
	"testing"

	"github.com/aziflaj/ghestimatebot/internal/ghestimatebot"
)

func TestEstimateRegex(t *testing.T) {
	tests := []struct {
		in string
		ok bool
	}{
		{"Estimate: 2 days", true},
		{"estimate: 5 days", true},
		{"Estimate:2 days", true},
		{"Estimate:2days", true},
		{"Estimate: 0 days", false},
		{"Estimate: some days", false},
		{"Estimate: days", false},
		{"Estimate: 3 day", false},
		{"Estimate: 1 day", true},
		{"No estimate here", false},
	}
	for _, tt := range tests {
		if got := ghestimatebot.HasEstimate(tt.in); got != tt.ok {
			t.Fatalf("%q => got %v want %v", tt.in, got, tt.ok)
		}
	}
}
