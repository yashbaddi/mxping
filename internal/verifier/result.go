package verifier

import "time"

// Status represents the verification status of an email address.
type Status string

const (
	// StatusValid means the email address syntax is valid, MX records exist, and the mailbox is confirmed reachable.
	StatusValid Status = "valid"
	// StatusInvalid means the email is malformed, domain has no MX/A records, or the mailbox was rejected.
	StatusInvalid Status = "invalid"
	// StatusCatchAll means the mail server accepts all recipients, so individual existence cannot be verified.
	StatusCatchAll Status = "catch_all"
	// StatusGreylisted means the server requested a retry later (temporary failure / rate limit / greylist).
	StatusGreylisted Status = "greylisted"
	// StatusBlocked means the connection was blocked by anti-spam, blacklists, or firewall (e.g. Port 25 blocked).
	StatusBlocked Status = "blocked"
	// StatusUnknown means verification could not be completed with high confidence.
	StatusUnknown Status = "unknown"
)

// SyntaxResult holds the breakdown of email syntax checks.
type SyntaxResult struct {
	IsValid   bool   `json:"is_valid"`
	Address   string `json:"address"`
	LocalPart string `json:"local_part"`
	Domain    string `json:"domain"`
	Reason    string `json:"reason,omitempty"`
}

// MXRecord holds resolved mail exchange server info.
type MXRecord struct {
	Host string `json:"host"`
	Pref uint16 `json:"preference"`
}

// DNSResult holds DNS resolution details.
type DNSResult struct {
	HasMX     bool       `json:"has_mx"`
	HasA      bool       `json:"has_a"`
	MXRecords []MXRecord `json:"mx_records,omitempty"`
	Reason    string     `json:"reason,omitempty"`
}

// SMTPResult holds low-level SMTP interaction details.
type SMTPResult struct {
	Connected    bool   `json:"connected"`
	TargetMX     string `json:"target_mx,omitempty"`
	ResponseCode int    `json:"response_code,omitempty"`
	ResponseMsg  string `json:"response_msg,omitempty"`
	IsCatchAll   bool   `json:"is_catch_all"`
	PortBlocked  bool   `json:"port_blocked"`
	Reason       string `json:"reason,omitempty"`
}

// Result is the comprehensive output of the email verification pipeline.
type Result struct {
	Email       string        `json:"email"`
	Status      Status        `json:"status"`
	Syntax      SyntaxResult  `json:"syntax"`
	DNS         DNSResult     `json:"dns"`
	SMTP        SMTPResult    `json:"smtp"`
	Duration    time.Duration `json:"duration_ms"`
	ExecutedAt  time.Time     `json:"executed_at"`
	SummaryMsg  string        `json:"summary_msg"`
}

