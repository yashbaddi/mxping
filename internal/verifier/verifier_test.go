package verifier

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"
)

func TestCatchAllDetection(t *testing.T) {
	// A server that accepts every RCPT TO (catch-all)
	acceptAll := map[string]string{}
	addr, cleanup := startMockSMTPServer(t, acceptAll)
	defer cleanup()

	host, portStr, _ := net.SplitHostPort(addr)
	var port int
	_, _ = fmt.Sscanf(portStr, "%d", &port)

	prober := NewSMTPProber(SMTPOptions{
		Port:    port,
		Timeout: 2 * time.Second,
	})

	// Override mock server behavior: let's test with a server that returns 250 for anything
	// startMockSMTPServer returns 550 unless in map, let's test when address is in map vs not
	t.Run("Generate random probe format", func(t *testing.T) {
		probe, err := GenerateRandomProbeAddress("example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(probe) == 0 {
			t.Fatalf("expected non-empty probe address")
		}
	})

	t.Run("Non Catch-All Server", func(t *testing.T) {
		// Normal server returns 550 for unknown random address
		isCatchAll, _, err := prober.CheckCatchAll(context.Background(), host, "example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if isCatchAll {
			t.Errorf("expected isCatchAll=false for normal mock server")
		}
	})
}

func TestVerifier_SyntaxAndDNS(t *testing.T) {
	cfg := DefaultConfig()
	cfg.CheckSMTP = false // Test syntax + DNS without connecting
	v := NewVerifier(cfg)
	ctx := context.Background()

	t.Run("Invalid Syntax", func(t *testing.T) {
		res := v.Verify(ctx, "invalid-email-address")
		if res.Status != StatusInvalid {
			t.Errorf("expected StatusInvalid, got %v", res.Status)
		}
	})

	t.Run("Non existent domain", func(t *testing.T) {
		res := v.Verify(ctx, "test@non-existent-domain-xyz-12345.invalid")
		if res.Status != StatusInvalid {
			t.Errorf("expected StatusInvalid, got %v", res.Status)
		}
	})
}
