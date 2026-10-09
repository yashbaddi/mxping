package verifier

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/textproto"
	"strings"
	"testing"
	"time"
)

// mockSMTPServer starts an in-memory test SMTP server on an ephemeral port.
func startMockSMTPServer(t *testing.T, rcptResponses map[string]string) (string, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start mock SMTP server: %v", err)
	}

	stop := make(chan struct{})

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-stop:
					return
				default:
					return
				}
			}
			go handleMockSMTP(conn, rcptResponses)
		}
	}()

	cleanup := func() {
		close(stop)
		_ = ln.Close()
	}

	return ln.Addr().String(), cleanup
}

func handleMockSMTP(conn net.Conn, rcptResponses map[string]string) {
	defer conn.Close()
	reader := textproto.NewReader(bufio.NewReader(conn))
	writer := textproto.NewWriter(bufio.NewWriter(conn))

	// Send 220 greeting
	_ = writer.PrintfLine("220 mock.mail.server ESMTP Ready")

	for {
		line, err := reader.ReadLine()
		if err != nil {
			return
		}
		cmd := strings.ToUpper(strings.TrimSpace(line))
		switch {
		case strings.HasPrefix(cmd, "EHLO") || strings.HasPrefix(cmd, "HELO"):
			_ = writer.PrintfLine("250-mock.mail.server Hello")
			_ = writer.PrintfLine("250 OK")
		case strings.HasPrefix(cmd, "MAIL FROM:"):
			_ = writer.PrintfLine("250 2.1.0 Ok")
		case strings.HasPrefix(cmd, "RCPT TO:"):
			target := strings.TrimPrefix(line, "RCPT TO:")
			target = strings.Trim(target, "<> ")
			if resp, ok := rcptResponses[target]; ok {
				_ = writer.PrintfLine("%s", resp)
			} else {
				_ = writer.PrintfLine("550 5.1.1 User unknown")
			}
		case strings.HasPrefix(cmd, "RSET"):
			_ = writer.PrintfLine("250 2.0.0 Reset OK")
		case strings.HasPrefix(cmd, "QUIT"):
			_ = writer.PrintfLine("221 2.0.0 Bye")
			return
		default:
			_ = writer.PrintfLine("500 5.5.1 Command unrecognized")
		}
	}
}

func TestSMTPProber_MockResponses(t *testing.T) {
	rcptResponses := map[string]string{
		"valid@example.com":      "250 2.1.5 Recipient OK",
		"invalid@example.com":    "550 5.1.1 User unknown",
		"greylisted@example.com": "450 4.2.0 Greylisted, try again later",
	}

	addr, cleanup := startMockSMTPServer(t, rcptResponses)
	defer cleanup()

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("failed to split host port: %v", err)
	}
	var port int
	_, _ = fmt.Sscanf(portStr, "%d", &port)

	prober := NewSMTPProber(SMTPOptions{
		Port:     port,
		Timeout:  2 * time.Second,
		HeloName: "test.local",
		MailFrom: "probe@test.local",
	})

	ctx := context.Background()

	t.Run("Valid Mailbox", func(t *testing.T) {
		res, err := prober.ProbeRecipient(ctx, host, "valid@example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.ResponseCode != 250 {
			t.Errorf("expected 250, got %d (msg: %s)", res.ResponseCode, res.ResponseMsg)
		}
	})

	t.Run("Invalid Mailbox", func(t *testing.T) {
		res, err := prober.ProbeRecipient(ctx, host, "invalid@example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.ResponseCode != 550 {
			t.Errorf("expected 550, got %d (msg: %s)", res.ResponseCode, res.ResponseMsg)
		}
	})

	t.Run("Greylisted Mailbox", func(t *testing.T) {
		res, err := prober.ProbeRecipient(ctx, host, "greylisted@example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.ResponseCode != 450 {
			t.Errorf("expected 450, got %d (msg: %s)", res.ResponseCode, res.ResponseMsg)
		}
	})
}

