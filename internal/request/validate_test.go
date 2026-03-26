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

func TestValidatePartialBookRequest_ValidInput(t *testing.T) {
	// Create PartialBookRequest br
	title := "Partial Book"
	br := PartialBookRequest{
		Title:  &title,
		Author: nil,
		Year:   nil,
	}

	errors := ValidatePartialBookRequest(&br)

	// Check errors is empty (len == 0)
	if len(errors) != 0 {
		t.Errorf("expected no validation errors, got %d: %v", len(errors), errors)
	}
}

func TestValidatePartialBookRequest_InvalidInput(t *testing.T) {
	title := "Partial Book"
	emptyTitle := ""
	emptyAuthor := ""
	zeroYear := 0
	// Table-driven tests: we define a list (slice) of test cases to loop over.
	tests := []struct {
		name     string             // A short label for the test case (helps identify failures)
		br       PartialBookRequest // The input data to validate
		wantKeys []string           // The expected error keys we should get back
	}{
		{
			name:     "empty title, author, and year",
			br:       PartialBookRequest{Title: &emptyTitle, Author: &emptyAuthor, Year: &zeroYear},
			wantKeys: []string{"title", "author", "year"},
		},
		{
			name:     "empty author",
			br:       PartialBookRequest{Title: &title, Author: &emptyAuthor, Year: nil},
			wantKeys: []string{"author"},
		},
		{
			name:     "invalid year",
			br:       PartialBookRequest{Title: &title, Author: nil, Year: &zeroYear},
			wantKeys: []string{"year"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			errors := ValidatePartialBookRequest(&tc.br)

			if len(errors) != len(tc.wantKeys) {
				t.Errorf("%s: expected %d validation errors, got %d", tc.name, len(tc.wantKeys), len(errors))
			}

			// Check if all expected keys exist in the error map
			for _, key := range tc.wantKeys {
				if _, ok := errors[key]; !ok {
					t.Errorf("%s: expected validation error for key %s", tc.name, key)
				}
			}
		})
	}
}
