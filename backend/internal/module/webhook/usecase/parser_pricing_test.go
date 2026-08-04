package usecase

import (
	"testing"
)

// TestParseWhatsAppStatuses_WithPricing / _WithoutPricing exercise the
// "statuses" array shape from Frente A of the WhatsApp 2026 adaptation plan
// (docs/WHATSAPP-2026-ADAPTATION-PLAN.md §2.2) — pure parsing, no DB needed.
// Same caveat as the engagement webhook parsers already in this package: the
// shape is inferred from public documentation, never confirmed against a
// real Meta app (see backend/ARCHITECTURE.md §8).
func TestParseWhatsAppStatuses_WithPricing(t *testing.T) {
	raw := []byte(`{
		"entry": [{"changes": [{"value": {
			"statuses": [{
				"id": "wamid.TEST_STATUS_1",
				"status": "delivered",
				"timestamp": "1700000000",
				"pricing": {"billable": true, "pricing_model": "CBP", "category": "utility"}
			}]
		}}]}]
	}`)

	statuses, err := parseWhatsAppStatuses(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(statuses) != 1 {
		t.Fatalf("expected 1 status event, got %d", len(statuses))
	}

	s := statuses[0]
	if s.MetaMessageID != "wamid.TEST_STATUS_1" {
		t.Fatalf("unexpected meta_message_id: %q", s.MetaMessageID)
	}
	if s.Status != "delivered" {
		t.Fatalf("unexpected status: %q", s.Status)
	}
	if s.Category == nil || *s.Category != "utility" {
		t.Fatalf("expected category=utility, got %v", s.Category)
	}
	if s.Billable == nil || !*s.Billable {
		t.Fatalf("expected billable=true, got %v", s.Billable)
	}
	if s.PricingModel == nil || *s.PricingModel != "CBP" {
		t.Fatalf("expected pricing_model=CBP, got %v", s.PricingModel)
	}
}

// TestParseWhatsAppStatuses_WithoutPricing covers a plain "sent" status
// event with no pricing object (e.g. Meta hasn't billed it yet) — Category/
// Billable/PricingModel must stay nil, not zero-valued, so the repository
// layer knows not to overwrite existing pricing data (see
// UpdateMessagePricing's COALESCE usage).
func TestParseWhatsAppStatuses_WithoutPricing(t *testing.T) {
	raw := []byte(`{
		"entry": [{"changes": [{"value": {
			"statuses": [{"id": "wamid.TEST_STATUS_2", "status": "sent", "timestamp": "1700000000"}]
		}}]}]
	}`)

	statuses, err := parseWhatsAppStatuses(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(statuses) != 1 {
		t.Fatalf("expected 1 status event, got %d", len(statuses))
	}
	if statuses[0].Category != nil {
		t.Fatalf("expected nil category for a status event without a pricing object, got %v", statuses[0].Category)
	}
}
