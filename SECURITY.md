# Security

## What you are exposing

Antigravity is a coding agent. Anyone who reaches its UI can read and write
files on the host and run commands on it. Treat access to Antigravity Remote as
equivalent to shell access.

## How access is protected

| Control | Detail |
| --- | --- |
| Password | Stored as a PBKDF2-HMAC-SHA256 hash (200,000 iterations, random 16-byte salt) in `~/.agy-remote/credentials.json`, mode `0600`. Never stored in plain text. |
| Sessions | 256-bit random tokens. Only their SHA-256 hashes are written to disk, so a stolen `sessions.json` grants nothing. Default lifetime 30 days (`--session-days`). |
| Cookies | `HttpOnly`, `SameSite=Lax`, and `Secure` whenever the request arrived over HTTPS. |
| Rate limiting | 5 failed attempts per IP per 5 minutes, then an exponential lockout up to 30 minutes. A global limiter (60 failures/minute) blunts distributed guessing. The lockout applies even to the correct password. |
| Revocation | `agy-remote sessions revoke` or "Sign out all" in the control panel invalidates every device immediately. |
| Admin surface | The control panel (QR code, password, shutdown) listens on a **separate loopback-only port** and is never routed through the public listener. |
| Login QR | The QR code carries a single-use enrollment token valid for 10 minutes, so you never type the password on a phone. Reuse is rejected. |

## Running behind a reverse proxy

If a proxy terminates TLS in front of `agy-remote`, tell it which peers to trust:

```bash
agy-remote serve --public-url https://agy.example.com --trusted-proxies 127.0.0.1/32,::1/128
```

Without `--trusted-proxies`, `X-Forwarded-For` and `X-Forwarded-Proto` are
ignored — which is the safe default, but it means rate limiting sees only the
proxy's address and cookies are not marked `Secure`. `agy-remote doctor` warns
when `--public-url` is set without trusted proxies.

Never list a CIDR you do not control. A trusted peer can claim any client IP.

## Recommended deployments

**Best: no public exposure at all.** Put the host on
[Tailscale](https://tailscale.com/) or a WireGuard network and reach it over the
private address. `agy-remote` detects and displays a Tailscale address
automatically. The password then becomes a second layer rather than the only one.

**Good: a domain behind Cloudflare Tunnel or Caddy with HTTPS**, plus a strong
password. `scripts/install.sh --domain agy.example.com` sets up Caddy with
automatic certificates and an HTTP-to-HTTPS redirect.

**Avoid: plain HTTP on a public IP.** The password and session cookie travel in
cleartext. On a trusted LAN this is acceptable; on the internet it is not.

## Local (same-network) mode

Local mode serves plain HTTP on your LAN, because browsers on phones reject
self-signed certificates and there is no way to get a real certificate for a
private address. The password still applies, but anyone already on your network
can observe the traffic. Do not use local mode on untrusted Wi-Fi — use
Tailscale instead.

## Reporting a vulnerability

Please open a [private security advisory](https://github.com/AFSlayer/antigravity-remote/security/advisories/new)
rather than a public issue. Include the version (`agy-remote version`), the
deployment shape, and reproduction steps. Expect a first response within a week.

## Not in scope

- Vulnerabilities in Antigravity itself or in Google's `language_server` —
  report those to Google.
- Anyone with an interactive session on the host: they can read
  `~/.agy-remote/` and the Antigravity OAuth token directly.
