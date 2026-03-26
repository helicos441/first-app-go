package request

func ValidatePartialBookRequest(br *PartialBookRequest) map[string]string {
	errors := make(map[string]string)
	if br.Title != nil && *br.Title == "" {
		errors["title"] = "title is required"
	}
	if br.Author != nil && *br.Author == "" {
		errors["author"] = "author is required"
	}
	if br.Year != nil && *br.Year < 1 {
		errors["year"] = "year must be a positive integer"
	}
	return errors
}

func ValidateFullBookRequest(br *FullBookRequest) map[string]string {
	errors := make(map[string]string)

	if br.Title == "" {
		errors["title"] = "title is required"
	}

	if br.Author == "" {
		errors["author"] = "author is required"
	}

	if br.Year < 1 {
		errors["year"] = "year must be a positive integer"
	}

	return errors
}
