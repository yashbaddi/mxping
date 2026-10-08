package verifier

import (
	"context"
	"testing"
	"time"
)

func TestDNSResolver_Lookup_ValidMX(t *testing.T) {
	resolver := NewDNSResolver(WithDNSTimeout(5 * time.Second))
	ctx := context.Background()

	res, err := resolver.Lookup(ctx, "google.com")
	if err != nil {
		t.Fatalf("unexpected error looking up google.com: %v", err)
	}

	if !res.HasMX {
		t.Errorf("expected HasMX=true for google.com")
	}

	if len(res.MXRecords) == 0 {
		t.Fatalf("expected at least one MX record for google.com")
	}

	// Verify MX records are sorted ascending by preference
	for i := 1; i < len(res.MXRecords); i++ {
		if res.MXRecords[i].Pref < res.MXRecords[i-1].Pref {
			t.Errorf("MX records not sorted properly: %v before %v", res.MXRecords[i-1], res.MXRecords[i])
		}
	}
}

func TestDNSResolver_Lookup_NonExistentDomain(t *testing.T) {
	resolver := NewDNSResolver(WithDNSTimeout(3 * time.Second))
	ctx := context.Background()

	// Use an unresolvable domain according to RFC 2606 / RFC 6761 or non-existent random domain
	domain := "non-existent-domain-mxping-test-xyz-987654321.invalid"
	res, err := resolver.Lookup(ctx, domain)
	if err == nil {
		t.Fatalf("expected error for non-existent domain, got nil (result: %+v)", res)
	}

	if res.HasMX || res.HasA {
		t.Errorf("expected HasMX=false and HasA=false for non-existent domain")
	}
}

func TestDNSResolver_Lookup_EmptyDomain(t *testing.T) {
	resolver := NewDNSResolver()
	ctx := context.Background()

	res, err := resolver.Lookup(ctx, "")
	if err == nil {
		t.Fatalf("expected error for empty domain")
	}
	if res.HasMX {
		t.Errorf("expected HasMX=false for empty domain")
	}
}

