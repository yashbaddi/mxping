# Release Notes - mxping v1.0.0

> **Release Tag**: `v1.0.0`  
> **Release Date**: October 2026  
> **Repository**: [github.com/yashbaddi/mxping](https://github.com/yashbaddi/mxping)

---

## 🚀 Overview

We are excited to announce the initial release of **`mxping` v1.0.0** — a fast, accurate, and dependency-free CLI tool written in Go to verify email address deliverability and mailbox existence without sending any emails.

`mxping` performs a non-intrusive multi-stage verification pipeline, connecting to target mail exchange servers over SMTP and cleanly terminating sessions right after recipient validation.

---

## ✨ Key Features & Capabilities

### 1. Multi-Stage Verification Pipeline
- **RFC 5322 Syntax Checking**: Validates local-part length ($\le 64$ chars), domain length ($\le 255$ chars), overall length ($\le 254$ chars), illegal characters, dot placement rules, and numeric TLD detection.
- **DNS & MX Record Resolution**: Resolves and sorts MX records by ascending preference/priority.
- **RFC 7505 Null MX Detection**: Identifies domains that explicitly publish `.` MX records to decline all inbound email.
- **RFC 5321 Fallback**: Automatically falls back to domain `A`/`AAAA` records when no MX records are published.
- **Low-Level SMTP Probing**: Performs initial handshake (`EHLO/HELO` $\rightarrow$ `MAIL FROM` $\rightarrow$ `RCPT TO` $\rightarrow$ `RSET/QUIT`) without ever delivering a message.
- **Catch-All (Accept-All) Detection**: Probes a cryptographically generated non-existent mailbox to identify servers that accept all email addresses, avoiding false positives.

### 2. Intelligent Network & ISP Diagnostics
- **Port 25 Block Detection**: Specifically diagnoses when outbound TCP port 25 is blocked by residential Internet Service Providers (ISPs) or cloud hosting providers (AWS, GCP, Azure, DigitalOcean), alerting the user with actionable next steps.

### 3. Flexible CLI & Execution Modes
- **Single Check**: `mxping user@example.com` with formatted terminal cards.
- **Batch Processing**: Check multiple addresses simultaneously (`mxping a@x.com b@y.com`).
- **Stdin & Pipeline Streaming**: Pipe addresses directly from files or shell pipelines (`cat leads.txt | mxping`), automatically filtering blank lines and `#` comment lines.
- **Parallel Worker Pools**: Built-in concurrency control (`-c 8`) for high-throughput batch checks.
- **Structured JSON Mode**: Emit clean JSON (`-json`) for seamless scripting and piping into `jq`.
- **Fast Syntax & DNS Check**: Skip SMTP connection overhead with `-skip-smtp`.
- **Verbose Protocol Dialogue**: Trace raw SMTP client/server transcripts with `-v`.

### 4. Zero External Dependencies & Cross-Platform
- Built entirely using the Go standard library.
- Available as static binaries for **Linux**, **macOS (Darwin)**, and **Windows** on both **`amd64`** and **`arm64`** (Apple Silicon).
- Native Linux packages available for `.deb` (Debian/Ubuntu), `.rpm` (RHEL/CentOS/Fedora), and `.apk` (Alpine).

---

## 📦 Installation & Getting Started

### One-Line Install (macOS & Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/yashbaddi/mxping/main/install.sh | bash
```

### Pre-Built Binaries

Download the appropriate archive for your operating system from the [Releases page](https://github.com/yashbaddi/mxping/releases):

| OS | Architecture | Package |
|---|---|---|
| **macOS** | Apple Silicon (`arm64`) | `mxping_0.1.0_darwin_arm64.tar.gz` |
| **macOS** | Intel (`amd64`) | `mxping_0.1.0_darwin_amd64.tar.gz` |
| **Linux** | 64-bit (`amd64`) | `mxping_0.1.0_linux_amd64.tar.gz` / `.deb` / `.rpm` / `.apk` |
| **Linux** | ARM 64-bit (`arm64`) | `mxping_0.1.0_linux_arm64.tar.gz` / `.deb` / `.rpm` / `.apk` |
| **Windows** | 64-bit (`amd64`) | `mxping_0.1.0_windows_amd64.zip` |
| **Windows** | ARM 64-bit (`arm64`) | `mxping_0.1.0_windows_arm64.zip` |

### Install via Debian / Ubuntu (.deb)

```bash
sudo dpkg -i mxping_0.1.0_linux_amd64.deb
```

### Install via Go

```bash
go install github.com/yashbaddi/mxping/cmd/mxping@latest
```

### Build from Source

```bash
git clone https://github.com/yashbaddi/mxping.git
cd mxping
make build
# Binary created at bin/mxping
```

---

## 🛠️ CLI Flag Reference

| Flag | Default | Description |
|---|---|---|
| `-from` | `check@mxping.local` | Sender address used in the `MAIL FROM` command |
| `-helo` | `mxping.local` | Domain name sent in `EHLO`/`HELO` identification |
| `-port` | `25` | SMTP port to connect to |
| `-timeout` | `10s` | Network timeout per operation |
| `-dns` | `""` | Custom DNS nameserver (e.g. `8.8.8.8:53`, `1.1.1.1:53`) |
| `-skip-smtp` | `false` | Validate syntax & DNS only without connecting to mail servers |
| `-skip-catchall` | `false` | Skip catch-all probe detection (saves latency in bulk lists) |
| `-json` | `false` | Output results in JSON format |
| `-no-color` | `false` | Disable ANSI terminal color codes |
| `-v` | `false` | Verbose SMTP dialogue logging to `stderr` |
| `-c` | `4` | Concurrency limit for batch checks |
| `-version` | `false` | Print version information and exit |

---

## 🚦 Exit Codes

`mxping` adheres to standard POSIX conventions for scripting:

- **`0`**: Success — email address is valid and deliverable (or all addresses in a batch are valid).
- **`1`**: Invalid — address is non-existent, rejected by server (`550`), or syntax is malformed.
- **`2`**: System/Network Error — connection blocked (e.g., Port 25 ISP block), invalid flags, or fatal I/O error.

---

## 📚 Documentation Links

- [CLI Command Usage Guide](docs/cli-usage.md)
- [Troubleshooting & Port 25 Guide](docs/troubleshooting.md)
- [Makefile Automation Reference](Makefile)
- [GoReleaser Workflow](.goreleaser.yaml)

