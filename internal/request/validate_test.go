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

func TestValidateFullBookRequest_InvalidInput(t *testing.T) {
	tests := []struct {
		name     string          // A short label for the test case
		br       FullBookRequest // The input data to validate
		wantKeys []string        // The expected error keys we should get back
	}{
		{
			name:     "missing all fields",
			br:       FullBookRequest{},
			wantKeys: []string{"title", "author", "year"},
		},
		{
			name: "missing title",
			br: FullBookRequest{
				Author: "Valid Author", // Valid author
				Year:   1999,           // Valid year
			},
			wantKeys: []string{"title"}, // Only title should fail validation
		},
		{
			name: "missing author",
			br: FullBookRequest{
				Title: "Test Title", // Valid title
				Year:  1999,         // Valid year
			},
			wantKeys: []string{"author"}, // Only author should fail validation
		},
		{
			name: "missing author",
			br: FullBookRequest{
				Title:  "Test Title",   // Valid title
				Author: "Valid Author", // Valid author
			},
			wantKeys: []string{"year"}, // Only author should fail validation
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			errors := ValidateFullBookRequest(&tc.br)

			if len(errors) != len(tc.wantKeys) {
				t.Errorf("%s: expected %d validation errors; got %d", tc.name, len(tc.wantKeys), len(errors))
			}

			for _, key := range tc.wantKeys {
				if _, ok := errors[key]; !ok {
					t.Errorf("%s: expected error for %s but is missing", tc.name, key)
				}
			}
		})
	}
}
