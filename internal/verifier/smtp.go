package verifier

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"strconv"
	"time"
)

var (
	ErrSMTPConnectFailed = errors.New("failed to connect to SMTP server")
	ErrSMTPTimeout       = errors.New("SMTP connection timed out")
	ErrPort25Blocked     = errors.New("port 25 appears blocked by ISP or firewall")
)

// SMTPOptions configures the SMTP probe behavior.
type SMTPOptions struct {
	Port        int
	Timeout     time.Duration
	HeloName    string
	MailFrom    string
	VerboseLog  func(format string, args ...any)
}

// DefaultSMTPOptions returns recommended default options for SMTP probing.
func DefaultSMTPOptions() SMTPOptions {
	return SMTPOptions{
		Port:     25,
		Timeout:  10 * time.Second,
		HeloName: "mxping.local",
		MailFrom: "check@mxping.local",
	}
}

// SMTPProber performs low-level SMTP handshakes to test recipient existence.
type SMTPProber struct {
	opts SMTPOptions
}

// NewSMTPProber creates a new SMTP prober with the given options.
func NewSMTPProber(opts SMTPOptions) *SMTPProber {
	if opts.Port <= 0 {
		opts.Port = 25
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 10 * time.Second
	}
	if opts.HeloName == "" {
		opts.HeloName = "mxping.local"
	}
	if opts.MailFrom == "" {
		opts.MailFrom = "check@mxping.local"
	}
	return &SMTPProber{opts: opts}
}

// ProbeRecipient attempts an SMTP transaction against the specified MX host for targetEmail.
func (p *SMTPProber) ProbeRecipient(ctx context.Context, mxHost string, targetEmail string) (SMTPResult, error) {
	result := SMTPResult{
		TargetMX: mxHost,
	}

	addr := net.JoinHostPort(mxHost, strconv.Itoa(p.opts.Port))
	p.log("Connecting to %s (timeout %v)...", addr, p.opts.Timeout)

	dialer := net.Dialer{Timeout: p.opts.Timeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		p.log("Connection failed: %v", err)
		result.Connected = false
		result.Reason = err.Error()

		// Diagnose potential port 25 block (very common on residential/cloud ISPs)
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() && p.opts.Port == 25 {
			result.PortBlocked = true
			result.Reason = "connection timed out on port 25 (likely blocked by your ISP or cloud provider)"
			return result, ErrPort25Blocked
		}
		return result, fmt.Errorf("%w: %v", ErrSMTPConnectFailed, err)
	}
	defer conn.Close()

	result.Connected = true
	_ = conn.SetDeadline(time.Now().Add(p.opts.Timeout))

	reader := textproto.NewReader(bufio.NewReader(conn))
	writer := textproto.NewWriter(bufio.NewWriter(conn))

	// 1. Read Initial Greeting Banner (expecting 220)
	code, msg, err := reader.ReadResponse(220)
	if err != nil {
		result.ResponseCode = code
		result.ResponseMsg = msg
		result.Reason = fmt.Sprintf("invalid initial greeting: %v", err)
		return result, err
	}
	p.log("< %d %s", code, msg)

	// 2. Send EHLO (fallback to HELO if rejected)
	p.log("> EHLO %s", p.opts.HeloName)
	if err := writer.PrintfLine("EHLO %s", p.opts.HeloName); err != nil {
		return result, err
	}
	code, msg, err = reader.ReadResponse(250)
	if err != nil {
		// Try fallback HELO
		p.log("< %d %s (EHLO rejected, trying HELO)", code, msg)
		p.log("> HELO %s", p.opts.HeloName)
		if err := writer.PrintfLine("HELO %s", p.opts.HeloName); err != nil {
			return result, err
		}
		code, msg, err = reader.ReadResponse(250)
		if err != nil {
			result.ResponseCode = code
			result.ResponseMsg = msg
			result.Reason = fmt.Sprintf("HELO rejected: %v", err)
			return result, err
		}
	}
	p.log("< %d %s", code, msg)

	// 3. Send MAIL FROM
	p.log("> MAIL FROM:<%s>", p.opts.MailFrom)
	if err := writer.PrintfLine("MAIL FROM:<%s>", p.opts.MailFrom); err != nil {
		return result, err
	}
	code, msg, err = reader.ReadResponse(250)
	if err != nil {
		result.ResponseCode = code
		result.ResponseMsg = msg
		result.Reason = fmt.Sprintf("MAIL FROM rejected: %v (code %d)", msg, code)
		p.log("< %d %s", code, msg)
		p.sendQuit(writer, reader)
		return result, nil
	}
	p.log("< %d %s", code, msg)

	// 4. Send RCPT TO
	p.log("> RCPT TO:<%s>", targetEmail)
	if err := writer.PrintfLine("RCPT TO:<%s>", targetEmail); err != nil {
		return result, err
	}
	code, msg, err = reader.ReadCodeLine(0)
	if err != nil && !isSMTPExpectedError(err) {
		result.Reason = fmt.Sprintf("failed to read RCPT response: %v", err)
		p.sendQuit(writer, reader)
		return result, err
	}
	p.log("< %d %s", code, msg)

	result.ResponseCode = code
	result.ResponseMsg = msg

	// 5. Clean teardown with RSET & QUIT
	p.sendRset(writer, reader)
	p.sendQuit(writer, reader)

	return result, nil
}

func (p *SMTPProber) sendRset(w *textproto.Writer, r *textproto.Reader) {
	_ = w.PrintfLine("RSET")
	_, _, _ = r.ReadCodeLine(0)
}

func (p *SMTPProber) sendQuit(w *textproto.Writer, r *textproto.Reader) {
	_ = w.PrintfLine("QUIT")
	_, _, _ = r.ReadCodeLine(0)
}

func (p *SMTPProber) log(format string, args ...any) {
	if p.opts.VerboseLog != nil {
		p.opts.VerboseLog(format, args...)
	}
}

func isSMTPExpectedError(err error) bool {
	if err == nil || errors.Is(err, io.EOF) {
		return false
	}
	var protoErr *textproto.Error
	return errors.As(err, &protoErr)
}
