package patches

import (
	"encoding/json"
	"regexp"
)

// The minifier wraps long lines, so the anchor must tolerate a newline between
// the assignment and the byEffort lookup.
var modelEffortRe = regexp.MustCompile(`,onClick:\(\)=>\{var y=[\r\n\s]*v\.byEffort\.get\(w\);y&&b\(y\)\}`)

func mobile(o Options) bool { return o.MobileUX }

// All returns every patch in a stable order. Adding a patch here is the only
// step needed: the tests, the doctor report, the control panel and the cache key
// all derive from this list.
func All() []Patch {
	return []Patch{
		// Without this the phone's browser would call https://127.0.0.1:<port>,
		// which resolves to the phone itself. Nothing works until it is fixed.
		{
			ID:       "base-url-origin",
			Desc:     "Point the web app at the browser origin instead of https://127.0.0.1",
			Target:   MainJS,
			Kind:     Literal,
			Required: true,
			Find:     "get baseUrl(){return`https://127.0.0.1:${this.port}`}",
			Replace:  "get baseUrl(){return typeof window!==\"undefined\"?window.location.origin:`https://127.0.0.1:${this.port}`}",
		},
		{
			ID:       "skip-onboarding",
			Desc:     "Skip the desktop onboarding redirect on remote clients",
			Target:   MainJS,
			Kind:     Literal,
			Required: true,
			Find:     `c.hasOnboardingScreens&&e!==2&&RK({to:"/onboarding",replace:!0,throw:!0})`,
			Replace:  `void 0`,
		},
		// Returning false from the Lexical ENTER command handler lets the editor
		// insert its default newline. There are three registerCommand(FE, ...)
		// call sites; only this one is the message composer.
		{
			ID:      "mobile-enter-newline",
			Desc:    "Enter inserts a newline on touch devices; Cmd/Ctrl+Enter sends",
			Target:  MainJS,
			Kind:    Literal,
			Enabled: mobile,
			Find:    `registerCommand(FE,k=>{if(!k)return!1;k.preventDefault();`,
			Replace: `registerCommand(FE,k=>{if(!k)return!1;if(window.matchMedia&&window.matchMedia("(pointer:coarse)").matches&&!k.metaKey&&!k.ctrlKey)return!1;k.preventDefault();`,
		},
		// On a desktop the effort submenu opens on hover, so the row's onClick is
		// a convenience that picks the default effort. A tap fires both, closing
		// the popup before the submenu can be used. Removing the handler leaves
		// the submenu reachable.
		{
			ID:      "model-effort-submenu",
			Desc:    "Tapping a model opens its reasoning-effort submenu instead of picking medium",
			Target:  MainJS,
			Kind:    Regexp,
			Enabled: mobile,
			FindRe:  modelEffortRe,
			Replace: "",
		},
		// Replacing the component with a function rather than an arrow keeps the
		// anchor "var vz=(" out of the replacement, so the rewrite cannot match
		// its own output.
		{
			ID:      "hide-mic-button",
			Desc:    "Hide the voice-recording button (transcription is unavailable in standalone mode)",
			Target:  MainJS,
			Kind:    Literal,
			Enabled: mobile,
			Find:    `uz.displayName="GutterHoverCommentButton";var vz=(`,
			Replace: `uz.displayName="GutterHoverCommentButton";var vz=function(){return null};var vzDisabled=(`,
		},
		// The prompt is an in-app banner shown once when notificationPermission is
		// still "default". Making the "have we asked yet" flag read as true on
		// touch devices skips it without touching the granted path, so a desktop
		// browser can still turn notifications on.
		{
			ID:      "mobile-skip-notification-prompt",
			Desc:    "Skip the Enable Notifications banner on touch devices",
			Target:  MainJS,
			Kind:    Literal,
			Enabled: mobile,
			Find:    `var e=!!this.storageService.get("didAskForNotificationPermission");`,
			Replace: `var e=(window.matchMedia&&window.matchMedia("(pointer:coarse)").matches)||!!this.storageService.get("didAskForNotificationPermission");`,
		},

		// Only the fallback is replaced: when a workspace is already open the
		// picker keeps starting from it.
		{
			ID:     "folder-picker-initial-path",
			Desc:   "Start the folder picker at the configured workspace root",
			Target: MainJS,
			Kind:   Literal,
			Enabled: func(o Options) bool {
				return o.WorkspaceRoot != ""
			},
			Find: `initialPath:b?b.fsPath:Sf?"C:/":"/",fetchDirectoryContents:`,
			ReplaceFn: func(o Options) string {
				return `initialPath:b?b.fsPath:` + jsString(o.WorkspaceRoot) + `,fetchDirectoryContents:`
			},
		},

		{
			ID:      "app-icons",
			Desc:    "Serve the official Antigravity favicon and home-screen icon",
			Target:  HTML,
			Kind:    InjectHead,
			Replace: appIcons,
		},
		{
			ID:      "pwa-manifest",
			Desc:    "Add a web app manifest so Add to Home Screen opens fullscreen",
			Target:  HTML,
			Kind:    InjectHead,
			Replace: pwaManifest,
		},
		{
			ID:      "touch-action",
			Desc:    "Remove the 300ms tap delay and tap highlight on controls",
			Target:  HTML,
			Kind:    InjectHead,
			Replace: touchAction,
		},
		{
			ID:      "safe-area-insets",
			Desc:    "Keep the composer and toasts clear of the iOS home bar",
			Target:  HTML,
			Kind:    InjectHead,
			Enabled: mobile,
			Replace: safeArea,
		},
		{
			ID:      "keyboard-detect",
			Desc:    "Collapse the safe-area gap while the on-screen keyboard is open",
			Target:  HTML,
			Kind:    InjectHead,
			Enabled: mobile,
			Replace: keyboardDetect,
		},
		{
			ID:     "cache-bust",
			Desc:   "Invalidate cached bundles when the applied patch set changes",
			Target: HTML,
			Kind:   Literal,
			Find:   `src="/main.js"`,
			ReplaceFn: func(o Options) string {
				if o.CacheKey == "" {
					return `src="/main.js"`
				}
				return `src="/main.js?agy=` + o.CacheKey + `"`
			},
		},
	}
}

func jsString(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		return `""`
	}
	return string(b)
}

const appIcons = `<link rel="icon" type="image/x-icon" href="/favicon.ico">
<link rel="apple-touch-icon" href="/apple-touch-icon.png">`

const pwaManifest = `<link rel="manifest" href="/__agy/manifest.webmanifest">`

const touchAction = `<style id="agy-touch-action">
button,input,textarea,select{touch-action:manipulation;-webkit-tap-highlight-color:transparent}
</style>`

const safeArea = `<style id="agy-safe-area">
@supports (padding-bottom: env(safe-area-inset-bottom)) {
  .h-\[100dvh\] { height: calc(100dvh - env(safe-area-inset-bottom, 0px)) !important; }
  html.agy-keyboard-open .h-\[100dvh\] { height: 100dvh !important; }
  .aux-drawer-popup { padding-bottom: env(safe-area-inset-bottom, 0px) !important; }
  html.agy-keyboard-open .aux-drawer-popup { padding-bottom: 0 !important; }
  .fixed.bottom-3 { bottom: calc(0.75rem + env(safe-area-inset-bottom, 0px)) !important; }
}
</style>`

const keyboardDetect = `<script id="agy-keyboard-detect">
(function () {
  if (!window.visualViewport) return;
  var base = window.visualViewport.height;
  window.visualViewport.addEventListener("resize", function () {
    if (window.visualViewport.height / base < 0.85) {
      document.documentElement.classList.add("agy-keyboard-open");
    } else {
      document.documentElement.classList.remove("agy-keyboard-open");
      base = window.visualViewport.height;
    }
  });
})();
</script>`
