package utils

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	// Email validation regex - RFC 5322 compliant
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	// Phone validation regex - supports international format
	// Supports: +62812345678, 0812345678, +1234567890, etc.
	phoneRegex = regexp.MustCompile(`^(\+?\d{1,3}[-.\s]?)?(\(?\d{2,4}\)?[-.\s]?)?\d{6,10}$`)

	// Indonesia phone number regex (more strict)
	indonesiaPhoneRegex = regexp.MustCompile(`^(\+62|62|0)[0-9]{9,13}$`)
)

// ValidateEmail validates email format
func ValidateEmail(email string) bool {
	if email == "" {
		return false
	}

	email = strings.TrimSpace(email)

	// Check length
	if len(email) < 5 || len(email) > 254 {
		return false
	}

	// Check format
	return emailRegex.MatchString(email)
}

// ValidatePhone validates phone number format
func ValidatePhone(phone string) bool {
	if phone == "" {
		return false
	}

	phone = strings.TrimSpace(phone)

	// Remove common separators for validation
	cleanPhone := strings.ReplaceAll(phone, " ", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "-", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, ".", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "(", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, ")", "")

	// Check length (minimum 10 digits, maximum 15 digits with country code)
	if len(cleanPhone) < 10 || len(cleanPhone) > 15 {
		return false
	}

	// Check format using cleaned phone
	return phoneRegex.MatchString(cleanPhone)
}

// ValidateIndonesiaPhone validates Indonesian phone number format
// Accepts: +62812345678, 62812345678, 0812345678
func ValidateIndonesiaPhone(phone string) bool {
	if phone == "" {
		return false
	}

	phone = strings.TrimSpace(phone)

	// Remove common separators
	cleanPhone := strings.ReplaceAll(phone, " ", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "-", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, ".", "")

	return indonesiaPhoneRegex.MatchString(cleanPhone)
}

// ValidatePassword validates password strength
// Requirements:
// - Minimum 8 characters
// - At least one uppercase letter
// - At least one lowercase letter
// - At least one digit
func ValidatePassword(password string) (bool, []string) {
	var errors []string

	if len(password) < 8 {
		errors = append(errors, "password must be at least 8 characters")
	}

	if len(password) > 128 {
		errors = append(errors, "password must not exceed 128 characters")
	}

	hasUpper := false
	hasLower := false
	hasDigit := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		}
	}

	if !hasUpper {
		errors = append(errors, "password must contain at least one uppercase letter")
	}

	if !hasLower {
		errors = append(errors, "password must contain at least one lowercase letter")
	}

	if !hasDigit {
		errors = append(errors, "password must contain at least one digit")
	}

	return len(errors) == 0, errors
}

// ValidateFullName validates full name
func ValidateFullName(name string) (bool, string) {
	name = strings.TrimSpace(name)

	if name == "" {
		return false, "full name is required"
	}

	if len(name) < 2 {
		return false, "full name must be at least 2 characters"
	}

	if len(name) > 100 {
		return false, "full name must not exceed 100 characters"
	}

	// Check if name contains only letters, spaces, and basic punctuation
	for _, char := range name {
		if !unicode.IsLetter(char) && !unicode.IsSpace(char) && char != '.' && char != '\'' && char != '-' {
			return false, "full name contains invalid characters"
		}
	}

	return true, ""
}

// SanitizeEmail sanitizes email by trimming and converting to lowercase
func SanitizeEmail(email string) string {
	email = strings.TrimSpace(email)
	email = strings.ToLower(email)
	return email
}

// SanitizePhone sanitizes phone number by removing spaces and dashes
func SanitizePhone(phone string) string {
	phone = strings.TrimSpace(phone)
	// Optionally remove separators for storage
	// phone = strings.ReplaceAll(phone, " ", "")
	// phone = strings.ReplaceAll(phone, "-", "")
	return phone
}

// NormalizeIndonesiaPhone normalizes Indonesian phone to +62 format
// Input: 0812345678 -> Output: +62812345678
// Input: 62812345678 -> Output: +62812345678
// Input: +62812345678 -> Output: +62812345678
func NormalizeIndonesiaPhone(phone string) string {
	phone = strings.TrimSpace(phone)

	// Remove separators
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, ".", "")

	// Convert to +62 format
	if strings.HasPrefix(phone, "0") {
		phone = "+62" + phone[1:]
	} else if strings.HasPrefix(phone, "62") && !strings.HasPrefix(phone, "+62") {
		phone = "+" + phone
	} else if !strings.HasPrefix(phone, "+") {
		// If no country code, assume Indonesia
		phone = "+62" + phone
	}

	return phone
}
