package quote

import (
	"testing"
	"time"
)

func TestCalculateQuoteWithAccessorials(t *testing.T) {
	req := Request{
		Origin:       Location{City: "Vancouver", Province: "BC", PostalCode: "V6B 1A1"},
		Destination:  Location{City: "Calgary", Province: "AB", PostalCode: "T2P 1J9"},
		ShipmentType: "ltl_pallet",
		Pieces:       2,
		Pallets:      2,
		WeightLbs:    700,
		LengthIn:     48,
		WidthIn:      40,
		HeightIn:     60,
		ServiceLevel: "standard",
		Accessorials: []string{"liftgate", "residential"},
	}

	resp := Calculate(req, CalculateOptions{QuoteID: "Q-TEST", CreatedAt: time.Date(2026, 6, 2, 12, 0, 0, 0, time.UTC)})

	if resp.EstimatedLow <= 0 || resp.EstimatedHigh <= resp.EstimatedLow {
		t.Fatalf("unexpected estimate range: low=%v high=%v", resp.EstimatedLow, resp.EstimatedHigh)
	}
	if resp.Breakdown.AccessorialFees != 160 {
		t.Fatalf("expected liftgate + residential accessorials to equal 160, got %v", resp.Breakdown.AccessorialFees)
	}
	if resp.Status != "Ready to Quote" {
		t.Fatalf("expected ready status, got %q", resp.Status)
	}
	if resp.Confidence != "High" {
		t.Fatalf("expected high confidence, got %q", resp.Confidence)
	}
}

func TestCalculateQuoteWithMissingFields(t *testing.T) {
	req := Request{
		Origin:      Location{City: "Vancouver", Province: "BC"},
		Destination: Location{City: "Calgary", Province: "AB"},
		WeightLbs:   0,
	}

	resp := Calculate(req, CalculateOptions{QuoteID: "Q-MISSING"})

	if resp.Status == "Ready to Quote" {
		t.Fatal("expected quote with missing fields to require review")
	}
	if len(resp.MissingFields) == 0 {
		t.Fatal("expected missing fields to be reported")
	}
}
