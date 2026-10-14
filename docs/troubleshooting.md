# Troubleshooting & Common Issues

This guide explains how to troubleshoot common network, DNS, and mail server behaviors encountered when using `mxping`.

---

## 1. Outbound Port 25 Blocked by ISP or Cloud Provider

### Symptoms
- `mxping` reports: `[⊘ BLOCKED] connection timed out on port 25 (likely blocked by your ISP or cloud provider)`.
- Exit code is `2`.

### Cause
Most residential Internet Service Providers (Comcast, AT&T, Vodafone, Jio, Airtel, etc.) and major cloud platforms (AWS EC2, Google Cloud Platform, Azure, DigitalOcean) block outbound traffic on TCP port 25 to prevent spam bots from sending bulk unauthenticated emails.

### Solutions
1. **Use `-skip-smtp`**: If you only need to ensure the email address has valid syntax and an active mail exchange, run:
   ```bash
   mxping -skip-smtp user@example.com
   ```
2. **Run on a VPS with Unblocked Port 25**: Run `mxping` on a server or VPS provider where port 25 is open (e.g., Linode/Akamai, Hetzner after requesting unblock, or a dedicated server).
3. **Check with `nc` or `telnet`**: Test manual connectivity:
   ```bash
   nc -zv -w 5 gmail-smtp-in.l.google.com 25
   ```

---

## 2. Catch-All Domains (`StatusCatchAll`)

### Symptoms
- Status shows `? CATCH-ALL`.
- Summary reads: `Mail server accepted recipient, but domain is a Catch-All server (accepts any email)`.

### Cause
Some companies configure their mail servers (e.g., Postfix, Google Workspace, Microsoft 365) to accept email sent to *any* mailbox under their domain (e.g. `anything@company.com` forwards to an administrator or catch-all inbox). 

In this case, the mail server responds with `250 OK` for both valid mailboxes and non-existent random addresses. Therefore, `mxping` cannot guarantee whether the specific user exists.

### Workaround
- If you want to bypass the secondary random probe and save latency:
  ```bash
  mxping -skip-catchall user@example.com
  ```

---

## 3. Greylisting & Temporary Defers (`StatusGreylisted`)

### Symptoms
- Status shows `⏳ GREYLISTED`.
- SMTP response code is `450`, `451`, or `421`.

### Cause
Greylisting is an anti-spam defense mechanism. When an unknown mail client connects for the first time, the receiving server temporarily rejects the message with a `4xx` code, instructing the client to retry after a short delay (typically 5 to 15 minutes). Legitimate mail transfer agents (MTAs) retry automatically, whereas spam bots often move on.

### Solutions
- Retry the verification command after 5-10 minutes.
- Supply a legitimate sender identity via `-from` and `-helo` to improve trust scores.

---

## 4. Mail Server Rejection (`554 Transaction Failed / Access Denied`)

### Symptoms
- Server immediately terminates connection or rejects `EHLO`/`MAIL FROM` with `554` or `550`.

### Cause
Receiving mail servers often run IP reputation checks (Spamhaus, Sorbs, Barracuda) or Reverse DNS (PTR) lookups. If your IP address has no PTR record or is on a residential IP blocklist, the server will refuse conversation.

### Solutions
- Specify a valid domain in `-helo` and `-from`:
  ```bash
  mxping -helo mail.yourdomain.com -from verify@yourdomain.com user@example.com
  ```
- Run verification from a clean static IP with matching PTR / reverse DNS records.

---

## 5. Null MX Records (RFC 7505)

### Symptoms
- Status is `✗ INVALID`.
- Reason mentions: `domain publishes a Null MX record (RFC 7505) and rejects all email`.

### Cause
RFC 7505 defines a standard way for domains that do not accept email to advertise this fact by publishing an MX record containing a single dot (`.`) with preference `0`. `mxping` detects this immediately and avoids attempting SMTP handshakes.

