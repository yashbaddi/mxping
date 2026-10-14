package output

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/yashbaddi/mxping/internal/verifier"
)

func sampleResult(status verifier.Status) verifier.Result {
	return verifier.Result{
		Email:  "user@example.com",
		Status: status,
		Syntax: verifier.SyntaxResult{
			IsValid:   true,
			Address:   "user@example.com",
			LocalPart: "user",
			Domain:    "example.com",
		},
		DNS: verifier.DNSResult{
			HasMX: true,
			MXRecords: []verifier.MXRecord{
				{Host: "mail.example.com", Pref: 10},
			},
		},
		SMTP: verifier.SMTPResult{
			Connected:    true,
			TargetMX:     "mail.example.com",
			ResponseCode: 250,
			ResponseMsg:  "Recipient OK",
		},
		Duration:   120 * time.Millisecond,
		SummaryMsg: "Mailbox exists and is deliverable",
	}
}

func TestFormatter_PrintResult_Text(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewFormatter(buf, false, true)

	r := sampleResult(verifier.StatusValid)
	formatter.PrintResult(r)

	out := buf.String()
	if !bytes.Contains([]byte(out), []byte("user@example.com")) {
		t.Errorf("expected output to contain email, got: %s", out)
	}
	if !bytes.Contains([]byte(out), []byte("VALID")) {
		t.Errorf("expected output to contain VALID badge, got: %s", out)
	}
}

func TestFormatter_PrintResult_JSON(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewFormatter(buf, true, true)

	r := sampleResult(verifier.StatusValid)
	formatter.PrintResult(r)

	var decoded verifier.Result
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v", err)
	}
	if decoded.Email != "user@example.com" {
		t.Errorf("expected email 'user@example.com', got '%s'", decoded.Email)
	}
	if decoded.Status != verifier.StatusValid {
		t.Errorf("expected status valid, got %v", decoded.Status)
	}
}

func TestFormatter_PrintBatchSummary(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewFormatter(buf, false, true)

	results := []verifier.Result{
		sampleResult(verifier.StatusValid),
		sampleResult(verifier.StatusInvalid),
	}
	formatter.PrintBatchSummary(results)

	out := buf.String()
	if !bytes.Contains([]byte(out), []byte("Total: 2")) {
		t.Errorf("expected batch summary total 2, got: %s", out)
	}
}

