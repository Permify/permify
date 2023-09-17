package ristretto

import (
	"testing"
)

func TestMaxCostOption(t *testing.T) {
	// Create a Ristretto instance with a default value.
	r := &Ristretto{}

	// Apply the MaxCost option.
	MaxCost("100")(r)

	// Check if the maxCost field is set correctly.
	if r.maxCost != "100" {
		t.Errorf("Expected maxCost to be 100, but got %v", r.maxCost)
	}
}

func TestNumberOfCountersOption(t *testing.T) {
	// Create a Ristretto instance with a default value.
	r := &Ristretto{}

	// Apply the NumberOfCounters option.
	NumberOfCounters(50)(r)

	// Check if the numCounters field is set correctly.
	if r.numCounters != 50 {
		t.Errorf("Expected numCounters to be 50, but got %v", r.numCounters)
	}
}
