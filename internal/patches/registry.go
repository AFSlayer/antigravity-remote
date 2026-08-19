package patches

import (
	"encoding/json"
	"regexp"
)

// The minifier wraps long lines, so the anchor must tolerate a newline between
// the assignment and the byEffort lookup.
var modelEffortRe = regexp.MustCompile(`,onClick:\(\)=>\{var y=[\r\n\s]*v\.byEffort\.get\(w\);y&&b\(y\)\}`)

// The Settings > Account button calls showLoginFlow, which the standalone build
// wires to a stub that only writes to the console. The minifier may wrap the line
// between the arrow and the call.
var signInButtonRe = regexp.MustCompile(`onClick:\(\)=>[\r\n\s]*\w+\.showLoginFlow\(\)`)

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
		// Google's standalone build cannot sign in from a browser: its auth service
		// is a stub, and its OAuth client only accepts loopback redirect URIs. Point
		// the button at a page that can actually complete the flow instead of
		// leaving it dead.
		{
			ID:      "sign-in-button",
			Desc:    "Make the Settings > Account sign-in button work over the network",
			Target:  MainJS,
			Kind:    Regexp,
			FindRe:  signInButtonRe,
			Replace: `onClick:()=>{window.location.href="/__agy/signin"}`,
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
			ID:      "mobile-signin-banner",
			Desc:    "Show a sign-in prompt on touch devices, which Antigravity omits there",
			Target:  HTML,
			Kind:    InjectHead,
			Enabled: mobile,
			Replace: signInBanner,
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

const touchAction = `<style id="agy-touch-action">
button,input,textarea,select{touch-action:manipulation;-webkit-tap-highlight-color:transparent}
</style>`

const safeArea = `<style id="agy-safe-area">
@supports (padding-bottom: env(safe-area-inset-bottom)) {
  .relative.w-screen.h-\[100dvh\] {
    height: 100dvh !important;
    padding: 0 !important;
  }
  div.h-\[100dvh\].w-screen.flex.flex-col {
    height: 100% !important;
    padding-top: 0 !important;
    padding-bottom: env(safe-area-inset-bottom, 0px) !important;
    box-sizing: border-box !important;
  }
  html.agy-keyboard-open div.h-\[100dvh\].w-screen.flex.flex-col {
    padding-bottom: 0 !important;
  }
  .aux-drawer-popup {
    padding-bottom: env(safe-area-inset-bottom, 0px) !important;
  }
  html.agy-keyboard-open .aux-drawer-popup {
    padding-bottom: 0 !important;
  }
  .fixed.bottom-3 {
    bottom: calc(0.75rem + env(safe-area-inset-bottom, 0px)) !important;
  }
  html.agy-keyboard-open .fixed.bottom-3 {
    bottom: 0.75rem !important;
  }
}
</style>`

const keyboardDetect = `<script id="agy-keyboard-detect">
(function () {
  if (!window.visualViewport) return;
  var base = window.visualViewport.height;
  function check() {
    var height = window.visualViewport.height;
    if (height / base < 0.85 || (window.innerHeight - height > 150)) {
      document.documentElement.classList.add("agy-keyboard-open");
    } else {
      document.documentElement.classList.remove("agy-keyboard-open");
      base = height;
    }
  }
  window.visualViewport.addEventListener("resize", check);
  window.visualViewport.addEventListener("scroll", check);
})();
</script>`

const signInBanner = `<style id="agy-signin-banner-style">
#agy-signin-banner-el {
  position: fixed;
  z-index: 40;
  display: none;
  text-decoration: none;
  animation: agy-banner-in 0.18s ease-out;
}
@keyframes agy-banner-in {
  from { opacity: 0; transform: translateY(-4px); }
  to { opacity: 1; transform: none; }
}
</style>
<script id="agy-signin-banner">
(function () {
  if (!(window.matchMedia && window.matchMedia("(pointer:coarse)").matches)) return;
  if (location.pathname.indexOf("/__agy/") === 0) return;

  // Antigravity's own auth banner, reproduced with its classes and icon so it themes
  // with the app. Its mobile layout omits the real one, which on desktop sits above
  // the composer card.
  //
  // The banner lives on document.body and is positioned over that spot rather than
  // inserted next to the card: the card is inside React's tree, so anything put
  // there is removed on the next render, and re-adding it in a loop thrashes the
  // layout badly enough to break the app's own keyboard handling.
  var ICON =
    '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 -960 960 960"' +
    ' fill="currentColor" class="h-4 w-4 shrink-0 text-yellow-500" aria-hidden="true">' +
    '<path d="M74.62-140L480-840L885.38-140H74.62ZM178-200H782L480-720L178-200Zm324.92-57.08q9.38-9.38 ' +
    '9.38-22.92t-9.38-22.92T480-312.31t-22.92,9.38T447.69-280t9.38,22.92T480-247.69t22.92-9.38ZM450-352.31h60v-200H450v200ZM480-460Z"/>' +
    "</svg>";

  var banner = document.createElement("a");
  banner.id = "agy-signin-banner-el";
  banner.href = "/__agy/signin";
  banner.className =
    "bg-muted px-3 min-h-[30px] py-1.5 flex items-center gap-2 text-sm border rounded-lg";
  banner.innerHTML =
    ICON +
    '<span class="text-foreground"><span>To use the agent, please login </span>' +
    '<span class="text-current underline">here</span></span>';

  // The composer is the only editable region on screen. Its card is the outermost
  // ancestor that still has a visible margin on both sides, since the wrappers above
  // it span the full width. Requiring an actual inset rather than merely "not quite
  // full width" is what keeps the banner from stretching edge to edge. Matching on
  // width rather than height keeps it detectable while the keyboard is open, which
  // shrinks innerHeight and broke an earlier ratio-based rule.
  function composerCard() {
    var editable = document.querySelector('[contenteditable="true"]');
    if (!editable) return null;

    var card = null;
    var node = editable;
    var viewport = window.innerWidth;

    for (var i = 0; i < 10 && node.parentElement && node.parentElement !== document.body; i++) {
      node = node.parentElement;
      var box = node.getBoundingClientRect();
      if (box.left >= 6 && box.right <= viewport - 6 && box.width >= viewport * 0.5 && box.height > 24) {
        card = node;
      }
    }
    return card;
  }

  // Antigravity does render its own banner inside a chat, just not on the project
  // list, so showing ours unconditionally puts two of them on screen. Detect the real
  // one by the copy it shares with the desktop layout and stand down when it is
  // there. If Google restyles it this stops matching and the duplicate comes back,
  // which is the mild failure mode of the two.
  function nativeBanner() {
    var nodes = document.querySelectorAll('[class*="bg-muted"]');
    for (var i = 0; i < nodes.length; i++) {
      if (nodes[i] !== banner && nodes[i].textContent.indexOf("please login") >= 0) {
        return true;
      }
    }
    return false;
  }

  var lastKey = "";

  function sync() {
    var card = nativeBanner() ? null : composerCard();
    if (!card) {
      if (banner.style.display !== "none") banner.style.display = "none";
      lastKey = "";
      return;
    }

    var box = card.getBoundingClientRect();
    if (box.width <= 0) return;

    if (banner.style.display === "none") banner.style.display = "flex";

    // Both getBoundingClientRect and position:fixed resolve against the layout
    // viewport, so tracking the card needs no adjustment for Safari's pan: the two
    // move together. Subtracting the pan here pushed the banner off-screen instead.
    var height = banner.offsetHeight || 32;
    var top = Math.max(4, box.top - height - 8);
    var key = box.left + ":" + box.width + ":" + top;
    if (key === lastKey) return;
    lastKey = key;

    banner.style.left = box.left + "px";
    banner.style.width = box.width + "px";
    banner.style.top = top + "px";
  }

  function start() {
    document.body.appendChild(banner);
    sync();

    window.addEventListener("resize", sync, { passive: true });
    window.addEventListener("scroll", sync, { passive: true });
    if (window.visualViewport) {
      window.visualViewport.addEventListener("resize", sync, { passive: true });
      window.visualViewport.addEventListener("scroll", sync, { passive: true });
    }
    setInterval(sync, 1000);
  }

  function check() {
    fetch("/__agy/api/signin/status", { credentials: "same-origin" })
      .then(function (r) { return r.ok ? r.json() : null; })
      .then(function (d) { if (d && !d.signedIn) start(); })
      .catch(function () {});
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", check);
  } else {
    check();
  }
})();
</script>`
