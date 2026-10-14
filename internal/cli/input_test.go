package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestReadEmails_FromArgs(t *testing.T) {
	args := []string{"user1@example.com", "  user2@example.com  ", ""}
	emails, err := ReadEmails(args, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(emails) != 2 {
		t.Fatalf("expected 2 emails, got %d", len(emails))
	}
	if emails[0] != "user1@example.com" || emails[1] != "user2@example.com" {
		t.Errorf("unexpected emails: %v", emails)
	}
}

func TestReadEmails_FromStdin(t *testing.T) {
	input := `
# Comment line
test1@domain.com

test2@domain.com
`
	stdin := strings.NewReader(input)
	emails, err := ReadEmails(nil, stdin)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(emails) != 2 {
		t.Fatalf("expected 2 emails, got %d", len(emails))
	}
	if emails[0] != "test1@domain.com" || emails[1] != "test2@domain.com" {
		t.Errorf("unexpected emails: %v", emails)
	}
}

func TestParseOptions(t *testing.T) {
	buf := &bytes.Buffer{}
	args := []string{"-port", "2525", "-skip-smtp", "-json", "user@example.com"}

	opts, remaining, err := ParseOptions(args, buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.SMTPPort != 2525 {
		t.Errorf("expected port 2525, got %d", opts.SMTPPort)
	}
	if !opts.SkipSMTP {
		t.Errorf("expected SkipSMTP=true")
	}
	if !opts.JSON {
		t.Errorf("expected JSON=true")
	}
	if len(remaining) != 1 || remaining[0] != "user@example.com" {
		t.Errorf("unexpected remaining args: %v", remaining)
	}
}
