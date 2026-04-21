package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// NormalizeSAPhone converts input into 27XXXXXXXXX format
func NormalizeSAPhone(input string) (string, error) {
	// Remove spaces, dashes, etc.
	phone := strings.TrimSpace(input)
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")

	// Remove "+"
	phone = strings.TrimPrefix(phone, "+")

	// Case 1: starts with 0 → replace with 27
	if strings.HasPrefix(phone, "0") {
		phone = "27" + phone[1:]
	}

	// Case 2: already starts with 27 → keep
	if !strings.HasPrefix(phone, "27") {
		return "", fmt.Errorf("only South African numbers are allowed")
	}

	// Must be exactly 11 digits (27 + 9 digits)
	if len(phone) != 11 {
		return "", fmt.Errorf("invalid phone length, expected 11 digits (27XXXXXXXXX)")
	}

	// Ensure numeric only
	match, _ := regexp.MatchString(`^27[0-9]{9}$`, phone)
	if !match {
		return "", fmt.Errorf("invalid phone format")
	}

	return phone, nil
}
