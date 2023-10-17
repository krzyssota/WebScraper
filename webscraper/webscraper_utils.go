package webscraper

import "unicode"

// at least one character is NOT whitespace
func isNotAllWhitespace(s string) bool {
	for _, c := range s {
		if !unicode.IsSpace(c) {
			return true
		}
	}
	return false
}

// check if in at least of one of the defined invisible tags
func inOneOfInvisibleTags(inInvisibleTags map[string]bool) bool {
	for _, inTag := range inInvisibleTags {
		if inTag {
			return true
		}
	}
	return false
}
