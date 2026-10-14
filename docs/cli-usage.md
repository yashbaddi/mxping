# CLI Command Usage Guide

`mxping` is a command-line tool written in Go to verify email address deliverability and mailbox existence without sending any emails.

---

## Table of Contents

1. [Basic Syntax](#basic-syntax)
2. [Single Email Verification](#1-single-email-verification)
3. [Batch Verification via Positional Arguments](#2-batch-verification-via-positional-arguments)
4. [Batch Verification via Stdin & Piping](#3-batch-verification-via-stdin--piping)
5. [Structured JSON Output & `jq` Integration](#4-structured-json-output--jq-integration)
6. [Fast Syntax & DNS Only (`-skip-smtp`)](#5-fast-syntax--dns-only--skip-smtp)
7. [Verbose SMTP Handshake Tracing (`-v`)](#6-verbose-smtp-handshake-tracing--v)
8. [Customizing Identity & Sender (`-from`, `-helo`)](#7-customizing-identity--sender--from--helo)
9. [Custom DNS Resolver (`-dns`)](#8-custom-dns-resolver--dns)
10. [Network Port & Timeout Tuning (`-port`, `-timeout`)](#9-network-port--timeout-tuning--port--timeout)
11. [Concurrency Tuning for Large Lists (`-c`)](#10-concurrency-tuning-for-large-lists--c)
12. [Disabling Color & CI/CD Support (`-no-color`)](#11-disabling-color--cicd-support--no-color)
13. [Exit Codes & Shell Scripting](#12-exit-codes--shell-scripting)

---

## Basic Syntax

```bash
mxping [flags] <email> [email2 ...]
cat emails.txt | mxping [flags]
```

Run `mxping -h` or `mxping --help` at any time to view all available flags:

```text
Usage:
  mxping [flags] <email> [email2 ...]
  cat emails.txt | mxping [flags]

Flags:
  -c int
        Concurrency limit for batch email checks (default 4)
  -dns string
        Custom DNS resolver address (e.g. 8.8.8.8:53)
  -from string
        Sender address for MAIL FROM probe (default "check@mxping.local")
  -helo string
        HELO/EHLO domain to identify as (default "mxping.local")
  -json
        Output results in JSON format
  -no-color
        Disable colored terminal output
  -port int
        SMTP port to connect to (default 25)
  -skip-catchall
        Skip catch-all probe detection
  -skip-smtp
        Validate syntax and DNS only; skip SMTP connection
  -timeout duration
        Network timeout duration (default 10s)
  -v    Enable verbose SMTP dialogue logging (printed to stderr)
  -version
        Print version information and exit
```

---

## 1. Single Email Verification

To check a single email address:

```bash
mxping alex@example.com
```

### Example Terminal Output

```text
 ✓ VALID   alex@example.com (342ms)
   Summary: Mailbox exists and is deliverable

  [Syntax]
   • Status:      Valid
   • Local Part:  alex
   • Domain:      example.com

  [DNS / Mail Exchange]
   • Records:     Found 2 MX record(s)
     1. mail.example.com (priority 10)
     2. backup.example.com (priority 20)

  [SMTP Handshake]
   • Connected:   Yes (mail.example.com)
   • Response:    Code 250
   • Message:     2.1.5 Recipient OK
   • Catch-All:   No
```

---

## 2. Batch Verification via Positional Arguments

Pass multiple email addresses separated by spaces:

```bash
mxping support@github.com invalid.user.123987@gmail.com contact@stripe.com
```

### Output Summary Table

When verifying multiple emails, `mxping` outputs a clean table followed by summary counters:

```text
EMAIL                               STATUS          DURATION     DETAILS
────────────────────────────────────────────────────────────────────────────────
support@github.com                  valid           280ms        Mailbox exists and is deliv...
invalid.user.123987@gmail.com       invalid         190ms        Mailbox does not exist (rej...
contact@stripe.com                  catch_all       310ms        Mail server accepted recipi...
────────────────────────────────────────────────────────────────────────────────
Total: 3 | Valid: 1 | Invalid: 1 | Catch-All: 1 | Other: 0
```

---

## 3. Batch Verification via Stdin & Piping

You can pipe lists of emails from files, scripts, or standard input streams. Empty lines and comment lines starting with `#` are automatically ignored.

### Pipe from a File

```bash
cat leads.txt | mxping
```

### Filter with Grep / Sed Before Pinging

```bash
grep "@company.org" subscribers.csv | cut -d',' -f2 | mxping
```

### Combine with Echo

```bash
echo "dev@example.com" | mxping
```

---

## 4. Structured JSON Output & `jq` Integration

Pass the `-json` flag to emit raw, parseable JSON.

### Single Check JSON

```bash
mxping -json alex@example.com
```

Output:
```json
{
  "email": "alex@example.com",
  "status": "valid",
  "syntax": {
    "is_valid": true,
    "address": "alex@example.com",
    "local_part": "alex",
    "domain": "example.com"
  },
  "dns": {
    "has_mx": true,
    "has_a": false,
    "mx_records": [
      {
        "host": "mail.example.com",
        "preference": 10
      }
    ]
  },
  "smtp": {
    "connected": true,
    "target_mx": "mail.example.com",
    "response_code": 250,
    "response_msg": "2.1.5 Recipient OK",
    "is_catch_all": false,
    "port_blocked": false
  },
  "duration_ms": 342000000,
  "executed_at": "2026-10-14T16:50:00Z",
  "summary_msg": "Mailbox exists and is deliverable"
}
```

### Batch Check JSON Array

```bash
cat emails.txt | mxping -json
```

Output:
```json
[
  {
    "email": "user1@example.com",
    "status": "valid",
    ...
  },
  {
    "email": "user2@example.com",
    "status": "invalid",
    ...
  }
]
```

### Useful `jq` Filter Recipes

Extract only deliverable emails:
```bash
cat emails.txt | mxping -json | jq -r '.[] | select(.status == "valid") | .email'
```

Extract mail exchange hosts for a list of emails:
```bash
cat emails.txt | mxping -json | jq -r '.[] | "\(.email) -> \(.smtp.target_mx)"'
```

Count status breakdown:
```bash
cat emails.txt | mxping -json | jq -r '.[].status' | sort | uniq -c
```

---

## 5. Fast Syntax & DNS Only (`-skip-smtp`)

If outbound port 25 is blocked by your ISP, or if you only want to verify that the domain has active mail exchanges without contacting mail servers:

```bash
mxping -skip-smtp user@example.com
```

- Verifies RFC 5322 compliance.
- Resolves MX and A/AAAA fallback records.
- Completes in a few milliseconds without TCP connection overhead.

---

## 6. Verbose SMTP Handshake Tracing (`-v`)

The `-v` flag logs the raw SMTP network protocol exchange (client requests and server responses) directly to `stderr`:

```bash
mxping -v user@example.com
```

### Example Verbose Transcript

```text
2026/10/14 16:52:10 [SMTP] Connecting to mail.example.com:25 (timeout 10s)...
2026/10/14 16:52:10 [SMTP] < 220 mail.example.com ESMTP Postfix
2026/10/14 16:52:10 [SMTP] > EHLO mxping.local
2026/10/14 16:52:10 [SMTP] < 250-mail.example.com
2026/10/14 16:52:10 [SMTP] > MAIL FROM:<check@mxping.local>
2026/10/14 16:52:10 [SMTP] < 250 2.1.0 Ok
2026/10/14 16:52:10 [SMTP] > RCPT TO:<user@example.com>
2026/10/14 16:52:10 [SMTP] < 250 2.1.5 Recipient OK
2026/10/14 16:52:10 [SMTP] Testing catch-all with random address: mxping-probe-a49e21f83c@example.com
2026/10/14 16:52:10 [SMTP] > RCPT TO:<mxping-probe-a49e21f83c@example.com>
2026/10/14 16:52:11 [SMTP] < 550 5.1.1 User unknown
```

---

## 7. Customizing Identity & Sender (`-from`, `-helo`)

Some strictly configured enterprise mail servers reject probes if the `EHLO` host or `MAIL FROM` domain is unrecognizable or fails reverse DNS checks.

You can configure legitimate domain values using:

```bash
mxping -helo mail.yourdomain.com -from verify@yourdomain.com user@targetdomain.com
```

---

## 8. Custom DNS Resolver (`-dns`)

By default, `mxping` uses the system's DNS resolver. To bypass local DNS cache or use public nameservers (e.g., Google or Cloudflare DNS):

```bash
mxping -dns 8.8.8.8:53 user@example.com
```

Or Cloudflare DNS:

```bash
mxping -dns 1.1.1.1:53 user@example.com
```

---

## 9. Network Port & Timeout Tuning (`-port`, `-timeout`)

To adjust the per-operation timeout (DNS queries and SMTP dialogues):

```bash
mxping -timeout 5s user@example.com
```

If connecting to a custom SMTP server running on a non-standard port (e.g. testing against local or staging mail servers):

```bash
mxping -port 2525 user@localtest.internal
```

---

## 10. Concurrency Tuning for Large Lists (`-c`)

When validating lists of hundreds or thousands of emails, use `-c` to control the worker pool size:

```bash
cat mailing-list.txt | mxping -c 16 -skip-catchall
```

> **Tip**: Skipping catch-all (`-skip-catchall`) halves the number of SMTP requests per address, doubling verification throughput for bulk checks.

---

## 11. Disabling Color & CI/CD Support (`-no-color`)

To prevent ANSI escape sequences in plain text log files or CI/CD output:

```bash
mxping -no-color user@example.com
```

`mxping` also automatically honors the [`NO_COLOR`](https://no-color.org/) standard environment variable:

```bash
NO_COLOR=1 mxping user@example.com
```

---

## 12. Exit Codes & Shell Scripting

`mxping` returns standard POSIX exit codes:

| Code | Meaning |
|:---:|---|
| `0` | **Success**: Email is valid and deliverable (or all emails in batch are valid) |
| `1` | **Invalid / Rejected**: Address does not exist, syntax is invalid, or server returned 550 |
| `2` | **System / Network Error**: Network blocked (e.g. Port 25 blocked), bad flags, or I/O failure |

### Bash Conditional Example

```bash
if mxping -skip-smtp "$USER_EMAIL" > /dev/null 2>&1; then
    echo "Email domain is valid, proceeding with signup..."
else
    echo "Invalid email address or domain, aborting."
    exit 1
fi
```

### GitHub Actions CI Workflow Step

```yaml
- name: Verify Contact Email
  run: |
    go install github.com/yashbaddi/mxping/cmd/mxping@latest
    mxping -skip-smtp "${{ github.event.inputs.contact_email }}"
```

