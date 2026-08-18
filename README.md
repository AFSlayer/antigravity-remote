<div align="center">

# Antigravity Remote

### Your real Antigravity, on your phone. Not a clone.

[![release](https://img.shields.io/github/v/release/AFSlayer/antigravity-remote?style=flat-square&color=4f7cff)](https://github.com/AFSlayer/antigravity-remote/releases/latest)
[![ci](https://img.shields.io/github/actions/workflow/status/AFSlayer/antigravity-remote/ci.yml?branch=main&style=flat-square)](https://github.com/AFSlayer/antigravity-remote/actions/workflows/ci.yml)
[![license](https://img.shields.io/badge/license-Apache--2.0-blue?style=flat-square)](LICENSE)
[![platforms](https://img.shields.io/badge/macOS%20%C2%B7%20Windows%20%C2%B7%20Linux-single%20binary-24292f?style=flat-square)](#install)

<img src="docs/assets/hero.png" width="320" alt="Antigravity running in a phone browser" />

**[Quick start](#30-seconds-to-your-phone) · [Server mode](#server-mode-your-own-cloud-antigravity) · [Security](SECURITY.md) · [한국어](README.ko.md)**

</div>

---

Antigravity's desktop app already contains a complete web IDE. The
`language_server` binary that ships with it serves that entire UI over HTTPS —
it just refuses to listen anywhere except `127.0.0.1`.

`agy-remote` is one small Go binary that opens that door safely: a
password-protected reverse proxy, a QR code that signs your phone in without
typing, and a dozen surgical patches that make a desktop IDE usable with a thumb.

**You get the real thing.** File explorer, terminal, artifacts, browser agent,
model picker with reasoning levels, everything — including every future
Antigravity feature, because there is no second UI to maintain.

## 30 seconds to your phone

**1. Install and start it.** One line, nothing to configure:

```bash
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install-desktop.sh | bash
```

```powershell
# Windows (PowerShell)
irm https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install-desktop.ps1 | iex
```

**2. Scan the QR code** in the control panel that opens. That's it — you're in,
on your phone, with no password typed.

<div align="center">
<img src="docs/assets/control-panel.png" width="330" alt="The agy-remote control panel" />
</div>

Antigravity doesn't even have to be running: `agy-remote` starts the app for you,
flips on remote control, finds the language server's port, and proxies it to your
network. After the first run, just type `agy-remote`.

> **Add it to your home screen.** On the phone, tap Share → *Add to Home Screen*.
> It opens fullscreen with the Antigravity icon, no browser chrome — that's what
> the screenshots on this page look like.

### Installing without warnings

These builds are not code-signed, because Apple and Microsoft both charge yearly
for that. If you download an archive in your browser instead of using the command
above, the OS will quarantine it:

- **macOS** — *"agy-remote cannot be opened because the developer cannot be
  verified."* Right-click the file → **Open** → **Open** again. Or clear the flag:
  `xattr -d com.apple.quarantine agy-remote`
- **Windows** — *"Windows protected your PC."* Click **More info** → **Run anyway**.

The one-line installers download with `curl` / `Invoke-WebRequest`, which does not
set the quarantine flag, so you never see either warning. That is the only reason
they exist.

## Server mode: your own cloud Antigravity

One command on any Linux box — a $5 VPS, a home server, an Oracle free-tier ARM
instance:

```bash
curl -fsSL https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install.sh | bash
```

The installer asks for a domain and a workspace folder, then:

- downloads the **official** Antigravity build straight from Google's
  `storage.googleapis.com` and extracts only the 165 MB `language_server` — no
  Google binary is redistributed by this project;
- writes a systemd unit so Antigravity is always up;
- sets up Caddy for automatic HTTPS on your domain;
- generates an access password.

Now you have an agent that keeps working while your laptop is closed.

> **One manual step.** Antigravity signs in through a `localhost` OAuth callback,
> which a remote server cannot receive. Copy the token from a computer where you
> already use the desktop app:
>
> ```bash
> scp ~/.gemini/jetski-standalone-oauth-token you@your-server:~/.gemini/
> ```
>
> `agy-remote` prints this command for you if the token is missing.

## Why not one of the other projects?

There are a dozen "Antigravity on your phone" projects. They all build a UI:
their own chat panel, or a screen mirror over the Chrome DevTools Protocol.

This one builds none.

|  | Other projects | Antigravity Remote |
| --- | --- | --- |
| The UI | Reimplemented or mirrored | **Antigravity's own web UI** |
| Feature coverage | Whatever was ported | **Everything** — terminal, artifacts, browser agent, diffs |
| New Antigravity features | Wait for the maintainer | **Work the day Google ships them** |
| Runtime | Usually Node.js + a package install | **One static binary, one dependency** |
| Headless server hosting | Rare | **First-class** |

The tradeoff is honest: because we patch strings in Google's bundle, an
Antigravity update can break a patch. So `agy-remote` checks every patch on every
start and **tells you** instead of failing silently. Run `agy-remote doctor` any
time to see exactly which patches matched.

## Mobile fixes

A desktop IDE in a phone browser has sharp edges. These get sanded down — each
one is a named, inspectable patch in
[`internal/patches/registry.go`](internal/patches/registry.go):

| What was wrong | Fix |
| --- | --- |
| The app talked to `https://127.0.0.1:<port>`, unreachable from a phone | Point it at the browser's own origin |
| Enter sent the message mid-sentence | **Enter inserts a newline** on touch devices; Cmd/Ctrl+Enter sends |
| The iOS home bar covered the composer | Honour `safe-area-inset-bottom`, and collapse the gap while the keyboard is open |
| Tapping a model picked "medium" and closed the menu | Tap opens the **reasoning-effort submenu** instead |
| A mic button that could never work (no transcription in standalone mode) | Hidden |
| New projects started in `/` | Start in your configured workspace root |
| No app icon, 300 ms tap delay, browser chrome everywhere | Official icon, instant taps, and a web manifest so **Add to Home Screen** opens fullscreen |

Prefer the UI untouched? `agy-remote --no-mobile-patches`.

<div align="center">
<table>
<tr>
<td align="center" width="33%"><img src="docs/assets/patch-models.png" width="200" alt="Model picker on a phone" /></td>
<td align="center" width="33%"><img src="docs/assets/patch-effort.png" width="200" alt="Reasoning effort submenu on a phone" /></td>
<td align="center" width="33%"><img src="docs/assets/settings.png" width="200" alt="Antigravity settings on a phone" /></td>
</tr>
<tr>
<td align="center"><sub>Every model, tappable</sub></td>
<td align="center"><sub>…and its reasoning level</sub></td>
<td align="center"><sub>Real settings, not a subset</sub></td>
</tr>
</table>
</div>

## Security

Access to Antigravity is access to a shell. This is treated seriously:

- password hashed with PBKDF2-SHA256 (200k iterations), never stored in plain text;
- 256-bit session tokens, **stored only as hashes**, revocable from the control panel;
- login rate limiting with exponential lockout, per-IP and global;
- `HttpOnly` + `SameSite=Lax` cookies, `Secure` whenever the request is HTTPS;
- the QR code is a **single-use** 10-minute enrollment token;
- the control panel — QR, password, shutdown — lives on a **separate loopback-only
  port** and is never exposed to the network.

Read [SECURITY.md](SECURITY.md) before putting this on a public address. Short
version: put the host on Tailscale if you can, and use HTTPS if you can't.

## Commands

```
agy-remote                     Share the desktop app on your network
agy-remote serve               Run headless on a server
agy-remote doctor              Check everything and say what's wrong
agy-remote config [flags]      Save options to config.json
agy-remote passwd [password]   Set the access password
agy-remote sessions [revoke]   List or sign out devices
```

Useful flags: `--port`, `--bind`, `--public-url`, `--workspace-root`,
`--trusted-proxies`, `--session-days`, `--no-mobile-patches`,
`--language-server`. Every one has an `AGY_*` environment equivalent —
`agy-remote help` lists them.

<details>
<summary><b>agy-remote doctor</b> output</summary>

```
  Antigravity Remote v0.1.0

  ✓ Data directory /Users/you/.agy-remote
  ✓ Access password is set
  ✓ 2 device(s) signed in
  ✓ Remote control is enabled in Antigravity settings
  ✓ language_server found at /Applications/Antigravity.app/Contents/Resources/bin/language_server
  ✓ Language server running (pid 62348, port 52856)
  ✓ CSRF token is present
  ✓ patch base-url-origin
  ✓ patch skip-onboarding
  ✓ patch mobile-enter-newline
  ✓ patch model-effort-submenu
  ✓ patch hide-mic-button
  – patch folder-picker-initial-path (disabled)
  ✓ patch app-icons
  ✓ patch pwa-manifest
  ✓ patch touch-action
  ✓ patch safe-area-insets
  ✓ patch keyboard-detect
  ✓ patch cache-bust
  ✓ Reachable on your network at http://192.168.1.20:8765
  ✓ Port 8765 is free

  ────────────────────────────────────────────────────
  ✓ Everything looks good.
```

</details>

## How it works

```
  your phone                    your machine / server
 ┌──────────┐            ┌──────────────────────────────────┐
 │ browser  │            │  agy-remote                      │
 │          │  password  │  ┌────────────────────────────┐  │
 │  Anti-   │◄──────────►│  │ password + sessions        │  │
 │ gravity  │   :8765    │  │ rate limiting              │  │
 │   UI     │            │  ├────────────────────────────┤  │
 └──────────┘            │  │ patch main.js / index.html │  │
                         │  └─────────────┬──────────────┘  │
                         │                │ https           │
                         │  ┌─────────────▼──────────────┐  │
                         │  │ language_server            │  │
                         │  │ --standalone  (Google)     │  │
                         │  └─────────────┬──────────────┘  │
                         └────────────────┼─────────────────┘
                                          │
                                   Google CloudCode
```

`agy-remote` never sees your prompts or code as anything but proxied bytes, and
never sends them anywhere except the language server on the same host.

## FAQ

**Does this need the Antigravity IDE or the CLI?**
No — the **desktop app** ("Antigravity", not "Antigravity IDE"). Only it can run
`language_server --standalone`, which is what serves the web UI. The CLI binary
embeds the bundle but exposes no flag to serve it.

**Will this survive Antigravity updates?**
The proxy will. Individual patches may not — `agy-remote` reports which ones
stopped matching, and `base-url-origin` is the only one remote access truly needs.
[Open an issue](https://github.com/AFSlayer/antigravity-remote/issues) when one
breaks and it gets fixed.

**Is my code sent to a third party?**
No. The proxy is on your machine, talking to a language server on your machine.
Antigravity's own traffic to Google is unchanged.

**Can two people use it at once?**
Each device gets its own session, but they share one Antigravity instance and one
Google account — treat it as your own multi-device setup, not a team server.

**Why is HTTP fine on my LAN but not on the internet?**
Phone browsers reject self-signed certificates and you cannot get a real
certificate for `192.168.x.x`. On a trusted network that is an acceptable
tradeoff; on a public address, use `--public-url` behind Caddy or a tunnel so
everything is HTTPS.

## Contributing

```bash
go test ./...            # patch engine, auth, proxy
go run ./cmd/agy-remote  # local mode
```

The patch engine is the interesting part:
[`internal/patches`](internal/patches). Every patch is a declarative struct with
an anchor string, and adding one to `All()` is the only step needed — the tests,
the doctor report, the control panel and the cache key all derive from that list.

Patches are verified at two levels, because none of Antigravity's bundle is
vendored into this repository:

- `patches_test.go` builds a **synthetic** document from the registry and checks
  the engine: matching, ordering, head injection, disabled patches, cache keys.
- `live_test.go` fetches the **real** bundle from a running language server and
  asserts each anchor matches exactly once. It skips when nothing is running, so
  start the Antigravity desktop app before a release:

  ```bash
  go test ./internal/patches -run Live -v
  ```

`agy-remote doctor` runs the same live check for users, which is how a broken
anchor gets reported instead of silently doing nothing.

## License

[Apache-2.0](LICENSE). Not affiliated with Google — please read
[DISCLAIMER.md](DISCLAIMER.md).
