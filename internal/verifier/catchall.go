package verifier

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateRandomProbeAddress generates an improbable email address for catch-all testing.
func GenerateRandomProbeAddress(domain string) (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random probe bytes: %w", err)
	}
	randomPart := hex.EncodeToString(bytes)
	return fmt.Sprintf("mxping-probe-%s@%s", randomPart, domain), nil
}

// CheckCatchAll probes the MX host with a random non-existent address to determine if the server accepts all recipients.
func (p *SMTPProber) CheckCatchAll(ctx context.Context, mxHost string, domain string) (bool, SMTPResult, error) {
	probeEmail, err := GenerateRandomProbeAddress(domain)
	if err != nil {
		return false, SMTPResult{}, err
	}

	p.log("Testing catch-all with random address: %s", probeEmail)
	res, err := p.ProbeRecipient(ctx, mxHost, probeEmail)
	if err != nil {
		return false, res, err
	}

	// If the server accepts the random address (code 250 or 251), it is a catch-all server.
	if res.ResponseCode == 250 || res.ResponseCode == 251 {
		res.IsCatchAll = true
		return true, res, nil
	}

	res.IsCatchAll = false
	return false, res, nil
}
