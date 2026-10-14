# mxping Documentation

Welcome to the documentation for **mxping**, a fast and accurate CLI tool written in Go to verify email address deliverability and mailbox existence.

## Documentation Index

- **[CLI Command Usage Guide](cli-usage.md)**
  - Detailed examples for all CLI commands, modes, and flags.
  - Single check, batch checking, and stdin piping recipes.
  - JSON output formatting and `jq` filters.
  - Performance and concurrency tuning (`-c`, `-skip-catchall`).
  - Shell scripting and CI/CD integration examples.
- **[Troubleshooting Guide](troubleshooting.md)**
  - Outbound Port 25 firewall blocks by ISPs and cloud hosts.
  - Understanding Catch-All domains and Greylisting.
  - Anti-spam heuristics, IP reputation, and Null MX records (RFC 7505).

---

## Quick Reference Table

| Goal | Command |
|---|---|
| Single address verification | `mxping user@example.com` |
| Output in JSON | `mxping -json user@example.com` |
| Pipe addresses from file | `cat emails.txt \| mxping` |
| Fast check (Syntax + DNS only) | `mxping -skip-smtp user@example.com` |
| Verbose SMTP conversation | `mxping -v user@example.com` |
| Custom DNS server | `mxping -dns 8.8.8.8:53 user@example.com` |
| Concurrency for batch lists | `cat emails.txt \| mxping -c 16 -skip-catchall` |
| Plain text without colors | `mxping -no-color user@example.com` |

