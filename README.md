<p align="center">
  <img src="assets/logo.svg" alt="mxping logo" width="360" />
</p>

<p align="center">
  <strong>Fast, accurate email address verification & mailbox pinging CLI with zero external dependencies.</strong>
</p>

<p align="center">
  <a href="https://github.com/yashbaddi/mxping/releases"><img src="https://img.shields.io/github/v/release/yashbaddi/mxping?color=blue&style=flat-square" alt="Release"></a>
  <a href="https://github.com/yashbaddi/mxping/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/yashbaddi/mxping/ci.yml?branch=main&label=CI&style=flat-square" alt="CI Status"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-green.svg?style=flat-square" alt="License"></a>
  <a href="https://goreportcard.com/report/github.com/yashbaddi/mxping"><img src="https://goreportcard.com/badge/github.com/yashbaddi/mxping?style=flat-square" alt="Go Report Card"></a>
</p>

---

`mxping` verifies whether an email address actually exists and can receive mail without delivering any messages.

## Features

- **RFC 5322 Syntax Validation**: Strict local-part, domain, length constraints, and illegal character checks.
- **DNS & MX Resolution**: MX record querying sorted by preference/priority, RFC 7505 Null MX detection, and RFC 5321 fallback to domain `A`/`AAAA` records.
- **SMTP Handshake Probing**: Connects via TCP port 25, runs `EHLO/HELO` $\rightarrow$ `MAIL FROM` $\rightarrow$ `RCPT TO` $\rightarrow$ `RSET/QUIT`. Never delivers emails.
- **Catch-All Detection**: Automatically detects domains configured to accept all mailboxes (preventing false positives).
- **Network / Port 25 Diagnostics**: Identifies when outbound port 25 is blocked by residential or cloud ISPs.
- **Batch Processing & Stdin Piping**: Supports verifying multiple emails concurrently with worker pools.
- **Structured JSON Output**: Ideal for integration into scripts and backend pipelines.

---

## Documentation

For full command references, scripting recipes, and troubleshooting:
- 📖 **[CLI Usage Guide](docs/cli-usage.md)**: Every flag, batch piping, `jq` recipes, and CI/CD setup.
- 🛠️ **[Troubleshooting Guide](docs/troubleshooting.md)**: Handling ISP Port 25 blocking, Greylisting, and Catch-All servers.

---

## Installation

### One-Line Install (macOS & Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/yashbaddi/mxping/main/install.sh | bash
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
```

---

## Usage

### Single Email Check

```bash
mxping user@example.com
```

### JSON Output

```bash
mxping -json user@example.com
```

### Batch Verification

Pass multiple addresses as arguments:

```bash
mxping alice@example.com bob@example.com charlie@company.org
```

Or pipe from a file or standard input:

```bash
cat emails.txt | mxping
```

### Fast Syntax & DNS Check (Skip SMTP)

Useful if your ISP blocks port 25 or if you only need syntax and DNS validation:

```bash
mxping -skip-smtp user@example.com
```

### Verbose SMTP Dialogue

```bash
mxping -v user@example.com
```

---

## CLI Options

| Flag | Default | Description |
|---|---|---|
| `-from` | `check@mxping.local` | Sender address for the `MAIL FROM` command |
| `-helo` | `mxping.local` | Domain name sent in `EHLO`/`HELO` |
| `-port` | `25` | SMTP port to connect to |
| `-timeout` | `10s` | Network timeout per operation |
| `-dns` | `""` | Custom DNS nameserver (e.g. `8.8.8.8:53`) |
| `-skip-smtp` | `false` | Validate syntax & DNS only without SMTP probe |
| `-skip-catchall` | `false` | Skip catch-all probe detection |
| `-json` | `false` | Output results in JSON format |
| `-no-color` | `false` | Disable ANSI terminal color codes |
| `-v` | `false` | Verbose SMTP dialogue logging to stderr |
| `-c` | `4` | Concurrency limit for batch checks |
| `-version` | `false` | Print version information |

---

## Exit Codes

- `0`: Valid email address (deliverable)
- `1`: Invalid email, rejected mailbox, or batch containing invalid addresses
- `2`: Network blocked (e.g. ISP Port 25 blocked), usage error, or fatal connection failure

---

## License

MIT
