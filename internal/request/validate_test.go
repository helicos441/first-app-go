package request

import "testing"

func TestValidateFullBookRequest_ValidInput(t *testing.T) {
	// Create FullBookRequest br
	br := FullBookRequest{
		Title:  "Valid Book",
		Author: "Valid Author",
		Year:   1999,
	}

	// errors := ValidateFullBookRequest(br)
	errors := ValidateFullBookRequest(&br)

	// Check errors is empty (len == 0)
	if len(errors) > 0 {
		t.Errorf("expected no validation errors, got %d: %v", len(errors), errors)
	}
}
