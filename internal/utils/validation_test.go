package utils

import "testing"

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"valid email", "test@example.com", true},
		{"valid email with subdomain", "user@mail.example.com", true},
		{"valid email with plus", "user+tag@example.com", true},
		{"valid email with dash", "user-name@example.com", true},
		{"invalid - no @", "testexample.com", false},
		{"invalid - no domain", "test@", false},
		{"invalid - no local", "@example.com", false},
		{"invalid - spaces", "test @example.com", false},
		{"invalid - empty", "", false},
		{"invalid - too short", "a@b.c", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateEmail(tt.email); got != tt.want {
				t.Errorf("ValidateEmail(%q) = %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}

func TestValidatePhone(t *testing.T) {
	tests := []struct {
		name  string
		phone string
		want  bool
	}{
		{"valid - Indonesia 08", "081234567890", true},
		{"valid - international +62", "+6281234567890", true},
		{"valid - with spaces", "+62 812 3456 7890", true},
		{"valid - with dashes", "+62-812-3456-7890", true},
		{"valid - US format", "+12345678901", true},
		{"invalid - too short", "12345", false},
		{"invalid - too long", "12345678901234567", false},
		{"invalid - letters", "081234abcd", false},
		{"invalid - empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidatePhone(tt.phone); got != tt.want {
				t.Errorf("ValidatePhone(%q) = %v, want %v", tt.phone, got, tt.want)
			}
		})
	}
}

func TestValidateIndonesiaPhone(t *testing.T) {
	tests := []struct {
		name  string
		phone string
		want  bool
	}{
		{"valid - 08 format", "081234567890", true},
		{"valid - +62 format", "+6281234567890", true},
		{"valid - 62 format", "6281234567890", true},
		{"invalid - wrong prefix", "181234567890", false},
		{"invalid - too short", "08123456", false},
		{"invalid - US format", "+12345678901", false},
		{"invalid - empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateIndonesiaPhone(tt.phone); got != tt.want {
				t.Errorf("ValidateIndonesiaPhone(%q) = %v, want %v", tt.phone, got, tt.want)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantOk   bool
	}{
		{"valid password", "Password123", true},
		{"valid with special", "Pass@word123", true},
		{"invalid - too short", "Pass1", false},
		{"invalid - no uppercase", "password123", false},
		{"invalid - no lowercase", "PASSWORD123", false},
		{"invalid - no digit", "Password", false},
		{"invalid - empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, _ := ValidatePassword(tt.password)
			if ok != tt.wantOk {
				t.Errorf("ValidatePassword(%q) = %v, want %v", tt.password, ok, tt.wantOk)
			}
		})
	}
}

func TestValidateFullName(t *testing.T) {
	tests := []struct {
		name     string
		fullName string
		wantOk   bool
	}{
		{"valid name", "John Doe", true},
		{"valid with apostrophe", "O'Brien", true},
		{"valid with dash", "Mary-Jane", true},
		{"valid with dot", "Dr. Smith", true},
		{"invalid - too short", "A", false},
		{"invalid - numbers", "John123", false},
		{"invalid - special chars", "John@Doe", false},
		{"invalid - empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, _ := ValidateFullName(tt.fullName)
			if ok != tt.wantOk {
				t.Errorf("ValidateFullName(%q) = %v, want %v", tt.fullName, ok, tt.wantOk)
			}
		})
	}
}

func TestSanitizeEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  string
	}{
		{"lowercase", "TEST@EXAMPLE.COM", "test@example.com"},
		{"trim spaces", "  test@example.com  ", "test@example.com"},
		{"already clean", "test@example.com", "test@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeEmail(tt.email); got != tt.want {
				t.Errorf("SanitizeEmail(%q) = %q, want %q", tt.email, got, tt.want)
			}
		})
	}
}

func TestNormalizeIndonesiaPhone(t *testing.T) {
	tests := []struct {
		name  string
		phone string
		want  string
	}{
		{"08 to +62", "081234567890", "+6281234567890"},
		{"62 to +62", "6281234567890", "+6281234567890"},
		{"already +62", "+6281234567890", "+6281234567890"},
		{"with spaces", "0812 3456 7890", "+6281234567890"},
		{"with dashes", "0812-3456-7890", "+6281234567890"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeIndonesiaPhone(tt.phone); got != tt.want {
				t.Errorf("NormalizeIndonesiaPhone(%q) = %q, want %q", tt.phone, got, tt.want)
			}
		})
	}
}
