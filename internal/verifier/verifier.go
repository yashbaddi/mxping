package verifier

import (
	"context"
	"fmt"
	"time"
)

// Config configures the overall verification pipeline.
type Config struct {
	DNSServer       string
	DNSTimeout      time.Duration
	SMTPPort        int
	SMTPTimeout     time.Duration
	HeloName        string
	MailFrom        string
	CheckCatchAll   bool
	CheckSMTP       bool
	VerboseLog      func(format string, args ...any)
}

// DefaultConfig returns safe and sensible defaults for email verification.
func DefaultConfig() Config {
	return Config{
		DNSTimeout:    5 * time.Second,
		SMTPPort:      25,
		SMTPTimeout:   10 * time.Second,
		HeloName:      "mxping.local",
		MailFrom:      "check@mxping.local",
		CheckCatchAll: true,
		CheckSMTP:     true,
	}
}

// Verifier orchestrates the end-to-end verification pipeline.
type Verifier struct {
	config Config
	dns    *DNSResolver
	smtp   *SMTPProber
}

// NewVerifier creates a new Verifier instance.
func NewVerifier(cfg Config) *Verifier {
	dnsOpts := []ResolverOption{
		WithDNSTimeout(cfg.DNSTimeout),
	}
	if cfg.DNSServer != "" {
		dnsOpts = append(dnsOpts, WithDNSServer(cfg.DNSServer))
	}

	smtpOpts := SMTPOptions{
		Port:       cfg.SMTPPort,
		Timeout:    cfg.SMTPTimeout,
		HeloName:   cfg.HeloName,
		MailFrom:   cfg.MailFrom,
		VerboseLog: cfg.VerboseLog,
	}

	return &Verifier{
		config: cfg,
		dns:    NewDNSResolver(dnsOpts...),
		smtp:   NewSMTPProber(smtpOpts),
	}
}

// Verify executes the full verification lifecycle for a given email address.
func (v *Verifier) Verify(ctx context.Context, email string) Result {
	startTime := time.Now()
	result := Result{
		Email:      email,
		Status:     StatusUnknown,
		ExecutedAt: startTime,
	}

	// Step 1: Syntax Validation
	syntaxRes, err := ValidateSyntax(email)
	result.Syntax = syntaxRes
	if err != nil {
		result.Status = StatusInvalid
		result.SummaryMsg = fmt.Sprintf("Syntax error: %s", syntaxRes.Reason)
		result.Duration = time.Since(startTime)
		return result
	}

	// Step 2: DNS & MX Records Lookup
	dnsRes, err := v.dns.Lookup(ctx, syntaxRes.Domain)
	result.DNS = dnsRes
	if err != nil {
		result.Status = StatusInvalid
		result.SummaryMsg = fmt.Sprintf("DNS resolution failed: %s", dnsRes.Reason)
		result.Duration = time.Since(startTime)
		return result
	}

	// If SMTP checking is disabled, evaluate status based on DNS validity
	if !v.config.CheckSMTP {
		result.Status = StatusValid
		result.SummaryMsg = "Syntax and DNS records are valid (SMTP check skipped)"
		result.Duration = time.Since(startTime)
		return result
	}

	// Step 3: SMTP Handshake & RCPT TO Probe
	// Try resolved MX hosts in order of priority
	var lastSMTPRes SMTPResult
	var lastErr error
	smtpSuccess := false

	for _, mx := range dnsRes.MXRecords {
		smtpRes, probeErr := v.smtp.ProbeRecipient(ctx, mx.Host, syntaxRes.Address)
		lastSMTPRes = smtpRes
		lastErr = probeErr

		// If connected and received a valid SMTP response code
		if smtpRes.Connected && smtpRes.ResponseCode > 0 {
			smtpSuccess = true
			break
		}
	}

	result.SMTP = lastSMTPRes

	if !smtpSuccess {
		if lastSMTPRes.PortBlocked {
			result.Status = StatusBlocked
			result.SummaryMsg = "Cannot verify mailbox: outbound port 25 is blocked by network or ISP"
		} else if lastErr != nil {
			result.Status = StatusUnknown
			result.SummaryMsg = fmt.Sprintf("Failed to establish SMTP connection: %v", lastErr)
		} else {
			result.Status = StatusUnknown
			result.SummaryMsg = "Unable to connect to any MX server"
		}
		result.Duration = time.Since(startTime)
		return result
	}

	// Step 4: Evaluate SMTP Response Code
	switch {
	case lastSMTPRes.ResponseCode == 250 || lastSMTPRes.ResponseCode == 251:
		// Step 5: Catch-All Detection
		if v.config.CheckCatchAll {
			isCatchAll, _, catchErr := v.smtp.CheckCatchAll(ctx, lastSMTPRes.TargetMX, syntaxRes.Domain)
			if catchErr == nil && isCatchAll {
				result.SMTP.IsCatchAll = true
				result.Status = StatusCatchAll
				result.SummaryMsg = "Mail server accepted recipient, but domain is a Catch-All server (accepts any email)"
				result.Duration = time.Since(startTime)
				return result
			}
		}

		result.Status = StatusValid
		result.SummaryMsg = "Mailbox exists and is deliverable"

	case lastSMTPRes.ResponseCode >= 550 && lastSMTPRes.ResponseCode <= 554:
		result.Status = StatusInvalid
		result.SummaryMsg = fmt.Sprintf("Mailbox does not exist (rejected by server: %d %s)", lastSMTPRes.ResponseCode, lastSMTPRes.ResponseMsg)

	case lastSMTPRes.ResponseCode >= 420 && lastSMTPRes.ResponseCode <= 459:
		result.Status = StatusGreylisted
		result.SummaryMsg = fmt.Sprintf("Server temporarily deferred request (greylisted or rate limited: %d %s)", lastSMTPRes.ResponseCode, lastSMTPRes.ResponseMsg)

	default:
		result.Status = StatusUnknown
		result.SummaryMsg = fmt.Sprintf("Server returned unexpected SMTP status %d: %s", lastSMTPRes.ResponseCode, lastSMTPRes.ResponseMsg)
	}

	result.Duration = time.Since(startTime)
	return result
}

