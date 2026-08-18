<div align="center">

# Antigravity Remote

**Use the Antigravity desktop app from your phone.**

[![release](https://img.shields.io/github/v/release/AFSlayer/antigravity-remote?style=flat-square&color=4f7cff)](https://github.com/AFSlayer/antigravity-remote/releases/latest)
[![ci](https://img.shields.io/github/actions/workflow/status/AFSlayer/antigravity-remote/ci.yml?branch=main&style=flat-square)](https://github.com/AFSlayer/antigravity-remote/actions/workflows/ci.yml)
[![license](https://img.shields.io/badge/license-Apache--2.0-blue?style=flat-square)](LICENSE)

<img src="docs/assets/demo.gif" width="300" alt="Asking an agent for the server state from a phone" />

<sub>Antigravity in mobile Safari, running a shell command on a server.</sub>

[한국어](README.ko.md) · [中文](README.zh-CN.md) · [日本語](README.ja.md) · [Português](README.pt-BR.md) · [Español](README.es.md)

</div>

## What this is

The Antigravity desktop app ships a binary called `language_server`. It does the
actual work of talking to Google, and with `--standalone` it also serves the whole
Antigravity UI as a web app. It only listens on `127.0.0.1`, so nothing but the
desktop app can reach it.

`agy-remote` puts a password in front of that server and forwards it to your
network. It also rewrites a few strings in the JS bundle as it passes through,
because a desktop IDE in a phone browser has some rough edges.

There are already a dozen projects for using Antigravity from a phone, and they all
build a UI: their own chat panel, or a screen mirror over CDP. This one serves
Antigravity's own instead. The terminal works, the file tree works, artifacts and
the browser agent work, and new Google features show up on their own. I didn't
write any of it and I don't have to keep up with it.

The tradeoff is that patching a minified bundle is fragile. An Antigravity update
can break a patch. `agy-remote` checks every patch on startup and says which ones
stopped matching.

## Install

```bash
# macOS, Linux
curl -fsSL https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install-desktop.sh | bash
```

```powershell
# Windows
irm https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install-desktop.ps1 | iex
```

That installs the binary and starts it. A control panel opens in your browser with
a QR code. Scan it and you're in. There's no password to type, because the code
carries a one-time token good for ten minutes.

<div align="center">
<img src="docs/assets/control-panel.png" width="320" alt="Control panel" />
</div>

Antigravity doesn't need to be open first. If it isn't running, `agy-remote` starts
it, turns on remote control in the settings, waits for the language server, and
works out which of its ports serves the UI.

On the phone, tap Share and *Add to Home Screen*. That gives you a fullscreen app
with the Antigravity icon and no browser bar, which is what the screenshots here
show.

### The security warnings

The binaries aren't code-signed, since Apple and Microsoft both charge yearly for
that. If you download an archive in a browser, the OS quarantines it:

- macOS says the developer can't be verified. Right-click the file, **Open**, then
  **Open** again. Or run `xattr -d com.apple.quarantine agy-remote`.
- Windows shows SmartScreen. **More info** → **Run anyway**.

`curl` and `Invoke-WebRequest` don't set the quarantine flag, so the install
commands above avoid all of this.

## Putting it on a server

On a Linux box this becomes an Antigravity that keeps working with your laptop
closed. A cheap VPS is enough. Mine is an Oracle free-tier ARM instance.

```bash
curl -fsSL https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install.sh | bash
```

It asks for a domain and a workspace folder, downloads the official Antigravity
build from Google's `storage.googleapis.com`, and extracts just the 165 MB
`language_server`. No Google binary is redistributed here. Then it writes a systemd
unit, configures Caddy for automatic HTTPS, and generates a password.

One step can't be automated. Antigravity signs in through an OAuth callback on
`localhost`, which a remote server can't receive. Copy the token from a machine
where you already use the desktop app:

```bash
scp ~/.gemini/jetski-standalone-oauth-token you@your-server:~/.gemini/
```

`agy-remote` prints that command for you when the token is missing.

## What gets patched

Twelve patches, each with a description in
[`internal/patches/registry.go`](internal/patches/registry.go). `agy-remote doctor`
reports which ones applied.

| Problem | Patch |
| --- | --- |
| The bundle calls `https://127.0.0.1:<port>`, which from a phone is the phone | Use the browser's origin |
| Enter sends the message mid-sentence | Enter is a newline on touch devices, Cmd/Ctrl+Enter sends |
| The iOS home bar covers the composer | Respect `safe-area-inset-bottom`, drop the gap while the keyboard is up |
| Tapping a model picks medium and closes the menu | Tap opens the effort submenu |
| A mic button that can't work, as standalone mode has no transcription | Hide it |
| New projects start in `/` | Start in your configured workspace folder |
| No app icon, 300 ms tap delay, browser chrome | Icon, instant taps, and a manifest for fullscreen Add to Home Screen |

For the UI untouched, use `agy-remote --no-mobile-patches`.

<div align="center">
<table><tr>
<td align="center"><img src="docs/assets/patch-models.png" width="190" alt="Model picker" /></td>
<td align="center"><img src="docs/assets/patch-effort.png" width="190" alt="Effort submenu" /></td>
<td align="center"><img src="docs/assets/settings.png" width="190" alt="Settings" /></td>
</tr></table>
</div>

## Security

Anyone who reaches Antigravity can read your files and run commands, so treat
access as equivalent to a shell.

- Passwords are hashed with PBKDF2-SHA256 at 200k iterations and never stored in
  plaintext.
- Session tokens are 256 random bits, and only their hashes are written to disk, so
  a copied `sessions.json` is useless. `agy-remote sessions revoke` signs out
  everything.
- Login is limited to five failures per IP per five minutes, then a lockout that
  doubles up to 30 minutes. A global limiter covers distributed attempts. The
  lockout also applies to the correct password.
- Cookies are `HttpOnly`, `SameSite=Lax`, and `Secure` when the request arrived
  over HTTPS.
- The control panel with the QR code and shutdown button listens on loopback only,
  on a separate port. It is never exposed to the network.

Behind a reverse proxy, name the peers you trust, otherwise forwarded headers are
ignored:

```bash
agy-remote serve --public-url https://agy.example.com --trusted-proxies 127.0.0.1/32
```

[SECURITY.md](SECURITY.md) covers the rest. In short: put the host on Tailscale if
you can, and use HTTPS if you can't.

## Commands

```
agy-remote                     share the desktop app on your network
agy-remote serve               run headless on a server
agy-remote doctor              check everything, say what's wrong
agy-remote config [flags]      write options to config.json
agy-remote passwd [password]   set the password
agy-remote sessions [revoke]   list or sign out devices
```

Useful flags: `--port`, `--public-url`, `--workspace-root`, `--trusted-proxies`,
`--session-days`, `--no-mobile-patches`, `--language-server`. Each one has an
`AGY_*` environment variable, and `agy-remote help` lists them.

<details>
<summary>What <code>doctor</code> prints</summary>

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
  phone                          your machine or server
┌──────────┐              ┌────────────────────────────────┐
│ browser  │   password   │ agy-remote                     │
│          │◄────────────►│   sessions, rate limiting      │
│  Anti-   │    :8765     │   patch main.js / index.html   │
│ gravity  │              │              │ https           │
│   UI     │              │   language_server --standalone │
└──────────┘              └──────────────┼─────────────────┘
                                         ▼
                                  Google CloudCode
```

Prompts and code pass through as proxied bytes and go nowhere but the language
server on the same host. Antigravity's own traffic to Google is unchanged.

## FAQ

**Can I use the IDE or the CLI instead?**
No, it needs the desktop app, the one called just "Antigravity". Only that runs
`language_server --standalone`, the mode that serves the web UI. The CLI binary has
the bundle compiled in but no flag to serve it.

**Will it survive Antigravity updates?**
The proxy will. Individual patches may not. `base-url-origin` is the only one remote
access needs, and the others are conveniences. Open an issue when one breaks.

**Does my code go anywhere?**
No. The proxy and the language server both run on your machine.

**Can two people use it?**
Each device gets a session, but they share one Antigravity and one Google account.
It's meant for your own devices, not a team.

**Why is HTTP fine on a LAN but not on the internet?**
Phone browsers reject self-signed certificates and you can't get a real one for
`192.168.x.x`. On a trusted network that's an acceptable trade. On a public address
it isn't, so use `--public-url` behind Caddy or a tunnel.

## Building it

```bash
go test ./...
go run ./cmd/agy-remote
```

The part worth reading is [`internal/patches`](internal/patches). A patch is a
struct with an anchor string, and adding one to `All()` is enough: the tests,
`doctor`, the control panel and the cache key all read from that list.

None of Antigravity's bundle is in this repo, so patches are tested at two levels.
`patches_test.go` builds a synthetic document from the registry to test the engine.
`live_test.go` fetches the real bundle from a running language server and asserts
each anchor matches exactly once, skipping when nothing is running. Before a
release, open the desktop app and run:

```bash
go test ./internal/patches -run Live -v
```

## License

[Apache-2.0](LICENSE). Not a Google project and not affiliated with Google. See
[DISCLAIMER.md](DISCLAIMER.md).
