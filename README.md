<div align="center">

# Antigravity Remote

**Use the Antigravity desktop app from your phone.**

[![release](https://img.shields.io/github/v/release/AFSlayer/antigravity-remote?style=flat-square&color=4f7cff)](https://github.com/AFSlayer/antigravity-remote/releases/latest)
[![ci](https://img.shields.io/github/actions/workflow/status/AFSlayer/antigravity-remote/ci.yml?branch=main&style=flat-square)](https://github.com/AFSlayer/antigravity-remote/actions/workflows/ci.yml)
[![license](https://img.shields.io/badge/license-Apache--2.0-blue?style=flat-square)](LICENSE)

<img src="docs/assets/demo.gif" width="300" alt="Asking an agent for the server state from a phone" />

[한국어](README.ko.md) · [中文](README.zh-CN.md) · [日本語](README.ja.md) · [Português](README.pt-BR.md) · [Español](README.es.md)

</div>

## What it is

The Antigravity desktop app ships a binary called `language_server`. It's what
talks to Google, and with `--standalone` it also serves the whole Antigravity UI as
a web app. It only listens on `127.0.0.1`.

`agy-remote` puts a password in front of it and forwards it to your network. It
also rewrites a few strings in the JS bundle on the way through, since a desktop
IDE in a phone browser has some rough edges.

Other projects that do this build their own UI, or mirror the screen over CDP. This
one serves Antigravity's own, so the terminal, file tree, artifacts and browser agent
all work, and new Google features appear without any work here.

The catch: patching a minified bundle is fragile. An Antigravity update can break a
patch. `agy-remote` checks all of them at startup and tells you which stopped
matching.

## Install

```bash
# macOS, Linux
curl -fsSL https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install-desktop.sh | bash
```

```powershell
# Windows
irm https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install-desktop.ps1 | iex
```

It installs and starts. A control panel opens with a QR code. Scan it and you're
in, no password to type. The code holds a one-time token good for ten minutes.

<div align="center">
<img src="docs/assets/control-panel.png" width="320" alt="Control panel" />
</div>

Antigravity doesn't need to be open first. If it isn't, `agy-remote` starts it,
turns on remote control, waits for the language server, and works out which of its
ports serves the UI.

On the phone use Share → *Add to Home Screen*. You get the Antigravity icon on the
home screen.

The binaries aren't code-signed, so downloading an archive in a browser gets it
quarantined. macOS: right-click, **Open**, **Open** again. Windows: **More info** →
**Run anyway**. The install commands above use `curl`, which doesn't set the
quarantine flag, so they skip all of that.

## On a server

On a Linux box, Antigravity keeps working with your laptop closed. A cheap VPS or a
free-tier ARM instance is enough.

```bash
curl -fsSL https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install.sh | bash
```

It asks for a domain and a workspace folder, pulls the official Antigravity build
from Google's `storage.googleapis.com`, and takes out just the 165 MB
`language_server`. No Google binary is redistributed here. Then it writes a systemd
unit, sets up Caddy for HTTPS, and makes a password.

Since it's the same web UI, a laptop browser works too, and it looks and behaves
like the desktop app. Conversations, workspaces and running agents live on the
server, so you can start something on your phone on the train and keep going on a
desktop when you get in. Nothing syncs because there's only one instance.

One step isn't automatic. Antigravity signs in through an OAuth callback on
`localhost`, which a remote server can't receive. Copy the token from a machine
where you already use the desktop app:

```bash
scp ~/.gemini/jetski-standalone-oauth-token you@your-server:~/.gemini/
```

`agy-remote` prints that command when the token is missing.

## What gets patched

Fourteen patches, each described in
[`internal/patches/registry.go`](internal/patches/registry.go). `agy-remote doctor`
says which applied.

| Problem | Patch |
| --- | --- |
| The bundle calls `https://127.0.0.1:<port>`, which from a phone is the phone | Use the browser's origin |
| Enter sends mid-sentence | On touch, Enter is a newline and Cmd/Ctrl+Enter sends |
| Tapping a model picks medium and closes the menu | Tap opens the effort submenu |
| "Enable Notifications" banner on the first reply | Skip it on touch devices |
| A mic button that can't work, since standalone has no transcription | Hide it |
| New projects start in `/` | Start in the workspace folder you set |
| No icon, 300 ms tap delay | The Antigravity icon, instant taps |
| iOS home bar overlap & virtual keyboard pushing view offscreen | Safe area insets & viewport height sync |
| Standalone build cannot sign in from a remote browser | Connect Settings sign-in button to web auth flow |

`agy-remote --no-mobile-patches` leaves the UI alone.

<div align="center">
<table><tr>
<td align="center"><img src="docs/assets/patch-models.png" width="190" alt="Model picker" /></td>
<td align="center"><img src="docs/assets/patch-effort.png" width="190" alt="Effort submenu" /></td>
<td align="center"><img src="docs/assets/settings.png" width="190" alt="Settings" /></td>
</tr></table>
</div>

## Security

Whoever gets in can read your files and run commands, so this is closer to shell
access than to sharing a doc.

- Passwords go through PBKDF2-SHA256 at 200k iterations. Nothing is stored in plain
  text.
- Session tokens are 256 random bits and only hashes hit the disk, so a copied
  `sessions.json` is useless.
- Five failed logins per IP per five minutes, then a lockout doubling to 30 minutes.
  It applies to the right password too.
- The control panel, where the QR code and shutdown button live, listens on loopback
  on its own port. It never reaches the network.

Behind a reverse proxy, say which peers you trust or forwarded headers get ignored:

```bash
agy-remote serve --public-url https://agy.example.com --trusted-proxies 127.0.0.1/32
```

Rest is in [SECURITY.md](SECURITY.md). Use Tailscale if you can and HTTPS if you
can't.

## Commands

```
agy-remote                     share the desktop app on your network
agy-remote serve               run headless on a server
agy-remote doctor              check everything, say what's wrong
agy-remote config [flags]      write options to config.json
agy-remote passwd [password]   set the password
agy-remote sessions [revoke]   list or sign out devices
```

`agy-remote help` lists the flags. Each has an `AGY_*` environment variable.

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

Prompts and code go to the language server on the same host and nowhere else.

## FAQ

**Does the IDE or the CLI work?** No. It has to be the desktop app, the one called
just "Antigravity", because only that runs `language_server --standalone`. The CLI
has the bundle compiled in but no flag to serve it.

**Will it survive updates?** The proxy will, individual patches might not.
`base-url-origin` is the only one remote access needs. File an issue when something
breaks.

**Does my code leave the machine?** No. Proxy and language server both run locally.

**Two people?** Each device gets a session, but they share one Antigravity and one
Google account. It's for your devices, not a team.

## Building

```bash
go test ./...
go run ./cmd/agy-remote
```

The interesting directory is [`internal/patches`](internal/patches). A patch is a
struct with an anchor string; add one to `All()` and the tests, `doctor`, the
control panel and the cache key pick it up.

Antigravity's bundle isn't in this repo, so `patches_test.go` tests the engine
against a synthetic document and `live_test.go` checks the anchors against a real
running language server, skipping if there isn't one. Before tagging a release,
open the desktop app and run `go test ./internal/patches -run Live -v`.

## License

[Apache-2.0](LICENSE). Not a Google project. See [DISCLAIMER.md](DISCLAIMER.md).
