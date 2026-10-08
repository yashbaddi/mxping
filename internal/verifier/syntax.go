package verifier

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
)

var (
	// domainRegex validates standard domain label format per RFC 1035 / RFC 5321.
	domainLabelRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?$`)
	
	// localPartRegex performs a sanity check on unquoted local parts.
	// Matches standard ASCII letters, digits, and RFC-allowed printable special characters.
	localPartRegex = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+$`)
)

// ValidateSyntax validates the email address against RFC standards and length constraints.
func ValidateSyntax(email string) (SyntaxResult, error) {
	trimmed := strings.TrimSpace(email)
	if trimmed == "" {
		return SyntaxResult{
			IsValid: false,
			Address: email,
			Reason:  "email address is empty",
		}, errors.New("email address is empty")
	}

	// RFC 5321: An email address cannot exceed 254 octets in total.
	if len(trimmed) > 254 {
		return SyntaxResult{
			IsValid: false,
			Address: trimmed,
			Reason:  "email address exceeds maximum length of 254 characters",
		}, errors.New("email address exceeds maximum length of 254 characters")
	}

	// Parse using standard net/mail parser
	parsed, err := mail.ParseAddress(trimmed)
	if err != nil {
		return SyntaxResult{
			IsValid: false,
			Address: trimmed,
			Reason:  "failed to parse email format: " + err.Error(),
		}, err
	}

	// Extract local and domain parts
	addr := parsed.Address
	parts := strings.Split(addr, "@")
	if len(parts) != 2 {
		return SyntaxResult{
			IsValid: false,
			Address: addr,
			Reason:  "invalid address structure: must contain exactly one '@'",
		}, errors.New("invalid address structure")
	}

	localPart := parts[0]
	domain := strings.ToLower(parts[1])

	// Validate Local Part length (RFC 5321 limit is 64 characters)
	if len(localPart) == 0 {
		return SyntaxResult{
			IsValid: false,
			Address: addr,
			Reason:  "local part cannot be empty",
		}, errors.New("local part is empty")
	}
	if len(localPart) > 64 {
		return SyntaxResult{
			IsValid:   false,
			Address:   addr,
			LocalPart: localPart,
			Domain:    domain,
			Reason:    "local part exceeds maximum length of 64 characters",
		}, errors.New("local part exceeds 64 characters")
	}

	// Check for leading, trailing, or consecutive dots in local part
	if strings.HasPrefix(localPart, ".") || strings.HasSuffix(localPart, ".") {
		return SyntaxResult{
			IsValid:   false,
			Address:   addr,
			LocalPart: localPart,
			Domain:    domain,
			Reason:    "local part cannot start or end with a dot",
		}, errors.New("local part cannot start or end with a dot")
	}
	if strings.Contains(localPart, "..") {
		return SyntaxResult{
			IsValid:   false,
			Address:   addr,
			LocalPart: localPart,
			Domain:    domain,
			Reason:    "local part cannot contain consecutive dots",
		}, errors.New("local part cannot contain consecutive dots")
	}

	// Validate local part characters if not quoted
	if !strings.HasPrefix(localPart, "\"") {
		if !localPartRegex.MatchString(localPart) {
			return SyntaxResult{
				IsValid:   false,
				Address:   addr,
				LocalPart: localPart,
				Domain:    domain,
				Reason:    "local part contains invalid characters",
			}, errors.New("local part contains invalid characters")
		}
	}

	// Validate Domain
	if len(domain) == 0 {
		return SyntaxResult{
			IsValid:   false,
			Address:   addr,
			LocalPart: localPart,
			Reason:    "domain cannot be empty",
		}, errors.New("domain is empty")
	}
	if len(domain) > 255 {
		return SyntaxResult{
			IsValid:   false,
			Address:   addr,
			LocalPart: localPart,
			Domain:    domain,
			Reason:    "domain exceeds maximum length of 255 characters",
		}, errors.New("domain exceeds 255 characters")
	}

	// Check domain labels
	labels := strings.Split(domain, ".")
	if len(labels) < 2 {
		return SyntaxResult{
			IsValid:   false,
			Address:   addr,
			LocalPart: localPart,
			Domain:    domain,
			Reason:    "domain must contain at least one dot separating labels (e.g., domain.com)",
		}, errors.New("domain must contain at least one dot")
	}

	for _, label := range labels {
		if len(label) == 0 {
			return SyntaxResult{
				IsValid:   false,
				Address:   addr,
				LocalPart: localPart,
				Domain:    domain,
				Reason:    "domain contains empty label (e.g., consecutive dots or trailing dot)",
			}, errors.New("domain contains empty label")
		}
		if len(label) > 63 {
			return SyntaxResult{
				IsValid:   false,
				Address:   addr,
				LocalPart: localPart,
				Domain:    domain,
				Reason:    "domain label exceeds maximum length of 63 characters",
			}, errors.New("domain label exceeds 63 characters")
		}
		if !domainLabelRegex.MatchString(label) {
			return SyntaxResult{
				IsValid:   false,
				Address:   addr,
				LocalPart: localPart,
				Domain:    domain,
				Reason:    "domain label contains invalid characters or leading/trailing hyphen",
			}, errors.New("domain label contains invalid characters")
		}
	}

	// TLD check (last label cannot be entirely numeric)
	tld := labels[len(labels)-1]
	allNumeric := true
	for _, ch := range tld {
		if ch < '0' || ch > '9' {
			allNumeric = false
			break
		}
	}
	if allNumeric {
		return SyntaxResult{
			IsValid:   false,
			Address:   addr,
			LocalPart: localPart,
			Domain:    domain,
			Reason:    "TLD cannot be entirely numeric",
		}, errors.New("TLD cannot be numeric")
	}

	return SyntaxResult{
		IsValid:   true,
		Address:   addr,
		LocalPart: localPart,
		Domain:    domain,
	}, nil
}

