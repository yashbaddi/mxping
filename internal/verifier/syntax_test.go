package verifier

import (
	"strings"
	"testing"
)

func TestValidateSyntax(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		wantValid bool
		wantLocal string
		wantDom   string
	}{
		// Valid cases
		{
			name:      "Simple valid email",
			email:     "user@example.com",
			wantValid: true,
			wantLocal: "user",
			wantDom:   "example.com",
		},
		{
			name:      "Email with tags and subdomains",
			email:     "john.doe+newsletter@mail.company.co.uk",
			wantValid: true,
			wantLocal: "john.doe+newsletter",
			wantDom:   "mail.company.co.uk",
		},
		{
			name:      "Valid special characters in local part",
			email:     "dis!count_123#test$@example.org",
			wantValid: true,
			wantLocal: "dis!count_123#test$",
			wantDom:   "example.org",
		},
		{
			name:      "Email with leading/trailing spaces (trimmed)",
			email:     "   alice@example.com  ",
			wantValid: true,
			wantLocal: "alice",
			wantDom:   "example.com",
		},
		{
			name:      "Domain with uppercase letters (normalized to lowercase)",
			email:     "Alice@EXAMPLE.COM",
			wantValid: true,
			wantLocal: "Alice",
			wantDom:   "example.com",
		},

		// Invalid cases
		{
			name:      "Empty email",
			email:     "",
			wantValid: false,
		},
		{
			name:      "Missing @ symbol",
			email:     "plainaddress",
			wantValid: false,
		},
		{
			name:      "Multiple @ symbols",
			email:     "user@foo@bar.com",
			wantValid: false,
		},
		{
			name:      "Missing local part",
			email:     "@example.com",
			wantValid: false,
		},
		{
			name:      "Missing domain",
			email:     "user@",
			wantValid: false,
		},
		{
			name:      "Local part starts with a dot",
			email:     ".user@example.com",
			wantValid: false,
		},
		{
			name:      "Local part ends with a dot",
			email:     "user.@example.com",
			wantValid: false,
		},
		{
			name:      "Consecutive dots in local part",
			email:     "user..name@example.com",
			wantValid: false,
		},
		{
			name:      "Domain without dot (no TLD)",
			email:     "user@localhost",
			wantValid: false,
		},
		{
			name:      "Domain with leading hyphen",
			email:     "user@-example.com",
			wantValid: false,
		},
		{
			name:      "Domain with trailing hyphen",
			email:     "user@example-.com",
			wantValid: false,
		},
		{
			name:      "All-numeric TLD",
			email:     "user@example.123",
			wantValid: false,
		},
		{
			name:      "Local part exceeding 64 chars",
			email:     strings.Repeat("a", 65) + "@example.com",
			wantValid: false,
		},
		{
			name:      "Total length exceeding 254 chars",
			email:     strings.Repeat("a", 64) + "@" + strings.Repeat("b", 60) + "." + strings.Repeat("c", 60) + "." + strings.Repeat("d", 60) + ".example.com",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := ValidateSyntax(tt.email)
			if tt.wantValid {
				if err != nil {
					t.Fatalf("expected valid, got error: %v", err)
				}
				if !res.IsValid {
					t.Fatalf("expected IsValid=true, got false (reason: %s)", res.Reason)
				}
				if res.LocalPart != tt.wantLocal {
					t.Errorf("expected localPart %q, got %q", tt.wantLocal, res.LocalPart)
				}
				if res.Domain != tt.wantDom {
					t.Errorf("expected domain %q, got %q", tt.wantDom, res.Domain)
				}
			} else {
				if err == nil && res.IsValid {
					t.Fatalf("expected invalid for %q, but got valid", tt.email)
				}
			}
		})
	}
}
