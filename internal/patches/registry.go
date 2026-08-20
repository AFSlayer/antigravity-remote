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
		// The titlebar user profile icon is a dead placeholder in standalone mode.
		// Replacing the component with a function returning null hides it cleanly.
		{
			ID:      "hide-user-profile-button",
			Desc:    "Hide the non-functional user profile placeholder button on mobile",
			Target:  MainJS,
			Kind:    Literal,
			Enabled: mobile,
			Find:    `function wmb({className:a=""}={}){return x.createElement("a",{href:"#",onClick:b=>{b.preventDefault()},className:` + "`w-6 h-6 rounded-full overflow-hidden shrink-0 flex items-center justify-center bg-transparent text-muted-foreground ${a}`" + `,"aria-label":"User Profile (Placeholder)"`,
			Replace: `function wmb(){return null};function wmbDisabled({className:a=""}={}){return x.createElement("a",{href:"#",onClick:b=>{b.preventDefault()},className:` + "`w-6 h-6 rounded-full overflow-hidden shrink-0 flex items-center justify-center bg-transparent text-muted-foreground ${a}`" + `,"aria-label":"User Profile (Placeholder)"`,
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

		// When a project or standalone section is selected on touch devices (e.g. via + button),
		// show the empty conversation view and composer directly rather than the full list.
		// Uses replace:true for the optimistic conversation creation so history back returns to the list.
		{
			ID:      "mobile-new-convo-view",
			Desc:    "Show the empty conversation composer view on touch devices when a project is selected",
			Target:  MainJS,
			Kind:    Literal,
			Enabled: mobile,
			Find:    `const tub=()=>{var a=yM(),b=IT();return(0,x.useCallback)((c,e)=>{b(HT.map(f=>({trigger:f,ran:!1})));a(c,{section:e})},[a,b])};` + "\n" + `var uub=()=>{var a=tub(),{q:b}=dM({strict:!1});return x.createElement("div",{className:"w-full h-full flex flex-col min-h-0 animate-fade-in"},x.createElement("div",{className:"flex-1 min-h-0 overflow-y-auto flex flex-col gap-6 pt-3"},x.createElement(sub,{surface:"background"})),x.createElement("div",{className:"shrink-0 p-2"},x.createElement(e_,{cascadeId:void 0},x.createElement(b_,{conversationId:void 0,isLoading:!1,dropdownPlacement:"top-start",openConversationOptimistically:a,showBottomToolbar:!0,` + "\n" + `aboveContent:x.createElement(s_,null),initialQuery:b}))))};`,
			Replace: `const tub=()=>{var a=yM(),b=IT();return(0,x.useCallback)((c,e)=>{b(HT.map(f=>({trigger:f,ran:!1})));a(c,{section:e,replace:!0})},[a,b])};var uub=()=>{var a=tub(),{q:b,section:sec}=dM({strict:!1}),isMobileNew=Boolean(window.matchMedia&&window.matchMedia("(pointer:coarse)").matches&&sec);return x.createElement("div",{className:"w-full h-full flex flex-col min-h-0 animate-fade-in"},isMobileNew?x.createElement("div",{className:"flex-1 min-h-0 flex flex-col items-center justify-center gap-3 select-none"},x.createElement(T,{name:"auto_awesome",size:32,className:"text-muted-foreground/30"}),x.createElement("span",{className:"text-xs text-muted-foreground/60"},"Start a new conversation")):x.createElement("div",{className:"flex-1 min-h-0 overflow-y-auto flex flex-col gap-6 pt-3"},x.createElement(sub,{surface:"background"})),x.createElement("div",{className:"shrink-0 p-2"},x.createElement(e_,{cascadeId:void 0},x.createElement(b_,{conversationId:void 0,isLoading:!1,dropdownPlacement:"top-start",openConversationOptimistically:a,showBottomToolbar:!0,aboveContent:x.createElement(s_,null),initialQuery:b}))))};`,
		},

		// On mobile, show the back button in the main titlebar when a project is selected
		// and ensure back navigation always clears the active section.
		{
			ID:      "mobile-new-convo-header",
			Desc:    "Show the back button in the titlebar when in new conversation mode and clear section on back",
			Target:  MainJS,
			Kind:    Literal,
			Enabled: mobile,
			Find:    `CM=()=>QL({select:a=>a.location.pathname==="/"})`,
			Replace: `CM=()=>QL({select:a=>a.location.pathname==="/"&&!a.location.search?.section})`,
		},
		{
			ID:      "mobile-back-clears-section",
			Desc:    "Ensure the mobile back button always clears the selected section to return to the root conversation list",
			Target:  MainJS,
			Kind:    Literal,
			Enabled: mobile,
			Find:    `x.createElement(gZ,{iconName:"arrow_back",onClick:()=>c(),"aria-label":"Back to home",dataTestId:"mobile-back-to-home"})`,
			Replace: `x.createElement(gZ,{iconName:"arrow_back",onClick:()=>c({clearSection:!0}),"aria-label":"Back to home",dataTestId:"mobile-back-to-home"})`,
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
		// A phone has no console, and the shell's geometry during the keyboard
		// animation is the only thing that explains the remaining layout bugs.
		// Off unless AGY_DEBUG is set.
		{
			ID:      "mobile-debug",
			Desc:    "Record the viewport and shell geometry around every keyboard event",
			Target:  HTML,
			Kind:    InjectHead,
			Enabled: func(o Options) bool { return o.Debug },
			Replace: mobileDebug,
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
input,textarea,select{font-size:16px !important}
</style>`

const safeArea = `<style id="agy-safe-area">
html,
body {
  position: fixed !important;
  inset: 0 !important;
  width: 100% !important;
  height: 100% !important;
  overflow: hidden !important;
  overscroll-behavior: none !important;
  margin: 0 !important;
  padding: 0 !important;
}

@supports (padding-bottom: env(safe-area-inset-bottom)) {
  .relative.w-screen.h-\[100dvh\] {
    position: fixed !important;
    inset: 0 !important;
    width: 100vw !important;
    height: 100% !important;
    max-height: 100% !important;
    padding: 0 !important;
    overflow: hidden !important;
  }
  div.h-\[100dvh\].w-screen.flex.flex-col {
    position: absolute !important;
    top: 0 !important;
    left: 0 !important;
    right: 0 !important;
    bottom: var(--agy-bottom, env(safe-area-inset-bottom, 0px)) !important;
    height: auto !important;
    max-height: none !important;
    padding-top: 0 !important;
    padding-bottom: 0 !important;
    box-sizing: border-box !important;
  }
  .aux-drawer-popup {
    padding-bottom: var(--agy-bottom, env(safe-area-inset-bottom, 0px)) !important;
  }
  .fixed.bottom-3 {
    bottom: calc(0.75rem + var(--agy-bottom, env(safe-area-inset-bottom, 0px))) !important;
  }
}
</style>`

const keyboardDetect = `<script id="agy-keyboard-detect">
(function () {
  var vv = window.visualViewport;
  if (!vv) return;

  // The messages live in a scroller nested inside the conversation view, which is
  // itself never taller than its content. Searching the subtree keeps the home
  // screen's history list out of it -- scrolling that one threw its virtualised
  // sticky headers off -- without depending on which element is the scroller.
  function chatScroller() {
    var root = document.querySelector('[data-testid="conversation-view"]');
    if (!root) return null;
    if (root.scrollHeight > root.clientHeight + 20) return root;

    var nodes = root.querySelectorAll("*");
    for (var i = 0; i < nodes.length && i < 400; i++) {
      var el = nodes[i];
      if (el.scrollHeight > el.clientHeight + 20 &&
          /auto|scroll/.test(getComputedStyle(el).overflowY)) {
        return el;
      }
    }
    return null;
  }

  function scrollChatToBottom() {
    var el = chatScroller();
    if (el) el.scrollTop = el.scrollHeight;
  }

  // html is position:fixed, so clientHeight is the layout viewport and does not
  // move with Safari's toolbar the way innerHeight can.
  function base() {
    return document.documentElement.clientHeight || window.innerHeight;
  }

  // Safari reveals the focused composer by panning the layout viewport, and it
  // reports that pan as a document scroll even here, where the document is fixed
  // and has nothing to scroll. Fixed elements move with it, so the whole shell
  // slides off the top of the screen until the offset is put back. Undoing it
  // once, mid-animation, is what made the shell lurch; doing it every frame keeps
  // the offset from ever being on screen for longer than one frame.
  function unpan() {
    var de = document.documentElement;
    if (window.scrollY === 0 && de.scrollTop === 0) return;
    window.scrollTo(0, 0);
    if (de.scrollTop !== 0) de.scrollTop = 0;
  }

  // The keyboard slides up over roughly this long, while visualViewport reports
  // its final height in a single step at the start.
  var OPEN_MS = 250;

  // Safari decides whether to pan about 40-80ms after focusin, before it reports
  // the new viewport height. Shrinking the shell to the height the keyboard had
  // last time gets the composer out of the way first, so there is no pan to
  // undo -- undoing one races Safari's own animation, which is what made the
  // shell lurch. The measurement is kept across page loads because the first
  // focus of a session is the one with nothing to go on.
  var predicted = 0;
  try {
    predicted = parseInt(localStorage.getItem("agy-kb"), 10) || 0;
  } catch (e) {}
  var holdUntil = 0;

  var applied = 0;
  var goal = 0;
  var from = 0;
  var moveAt = 0;
  var settled = true;
  var raf = 0;
  var deadline = 0;

  function write(kb) {
    if (Math.abs(kb - applied) < 1) return;

    var opening = applied === 0 && kb > 0;
    applied = kb;
    if (kb) {
      document.documentElement.style.setProperty("--agy-bottom", kb + "px");
    } else {
      document.documentElement.style.removeProperty("--agy-bottom");
    }
    if (opening) scrollChatToBottom();
  }

  function frame() {
    unpan();

    var target = Math.max(0, Math.round(base() - vv.height));
    if (target <= 20) target = 0;

    if (target > 0) {
      holdUntil = 0;
      if (target !== predicted) {
        predicted = target;
        try {
          localStorage.setItem("agy-kb", String(target));
        } catch (e) {}
      }
    } else if (performance.now() < holdUntil) {
      // Hold the predicted shrink until Safari reports the keyboard. If it never
      // does -- a hardware keyboard, say -- the hold expires and the shell
      // springs back on its own.
      target = predicted;
    }

    if (target !== goal) {
      goal = target;
      from = applied;
      moveAt = performance.now();
      settled = false;
    }

    if (goal <= from) {
      // Closing: the keyboard is already on its way out, and following it
      // immediately is what the shell did smoothly before.
      write(goal);
    } else {
      var p = Math.min(1, (performance.now() - moveAt) / OPEN_MS);
      write(Math.round(from + (goal - from) * (1 - Math.pow(1 - p, 3))));
    }

    // The chat has to be pulled to the bottom once the shell has stopped moving:
    // doing it only while the shell shrinks leaves it short of the last message,
    // because the scrollable distance is still growing.
    if (!settled && applied === goal) {
      settled = true;
      if (goal > 0) scrollChatToBottom();
    }

    if (performance.now() < deadline) {
      raf = requestAnimationFrame(frame);
      return;
    }
    raf = 0;
    write(goal);
    if (applied > 0) scrollChatToBottom();
  }

  function track(ms) {
    unpan();
    var until = performance.now() + ms;
    if (until > deadline) deadline = until;
    if (!raf) raf = requestAnimationFrame(frame);
  }

  vv.addEventListener("resize", function () { track(700); });
  vv.addEventListener("scroll", function () { track(400); });

  window.addEventListener("focusin", function (e) {
    var t = e.target;
    if (t && (t.tagName === "INPUT" || t.tagName === "TEXTAREA" || t.isContentEditable)) {
      if (predicted > 20 && applied === 0) {
        holdUntil = performance.now() + 500;
        goal = from = predicted;
        moveAt = performance.now();
        write(predicted);
      }
      track(900);
    }
  });
})();
</script>`

const mobileDebug = `<script id="agy-debug">
(function () {
  var vv = window.visualViewport;
  if (!vv) return;

  var ENDPOINT = "/__agy/api/debug/log";

  // The two selectors the safe-area patch relies on. Their match count is logged
  // on every sample because a selector that matches nothing, or several nested
  // shells, still reports as "applied" in the patch report.
  var OUTER = ".relative.w-screen.h-\\[100dvh\\]";
  var INNER = "div.h-\\[100dvh\\].w-screen.flex.flex-col";

  var session = Math.random().toString(36).slice(2, 8);
  var episodes = 0;

  function n(v) { return Math.round(v); }
  function pad(v, w) { v = String(v); while (v.length < w) v = " " + v; return v; }
  function padR(v, w) { v = String(v); while (v.length < w) v += " "; return v; }

  function all(sel) {
    try { return document.querySelectorAll(sel); } catch (e) { return []; }
  }
  function one(sel) {
    try { return document.querySelector(sel); } catch (e) { return null; }
  }

  function box(el) {
    if (!el) return "-";
    var b = el.getBoundingClientRect();
    return n(b.top) + ".." + n(b.bottom) + "/h" + n(b.height);
  }

  function insets() {
    var probe = document.createElement("div");
    probe.style.cssText =
      "position:fixed;left:0;top:0;width:0;height:0;visibility:hidden;padding:" +
      "env(safe-area-inset-top) env(safe-area-inset-right) " +
      "env(safe-area-inset-bottom) env(safe-area-inset-left)";
    document.documentElement.appendChild(probe);
    var s = getComputedStyle(probe);
    var out = [s.paddingTop, s.paddingRight, s.paddingBottom, s.paddingLeft].join("/");
    probe.parentNode.removeChild(probe);
    return out.replace(/px/g, "");
  }

  // In a conversation the composer's scroller is the conversation view; on the
  // home screen the history list is virtualised and its scroller is an ancestor.
  function scroller() {
    var el = one('[data-testid="conversation-view"]');
    if (el) return el;
    el = one('[data-testid^="conversation-list-"]');
    while (el && el !== document.body) {
      if (/auto|scroll/.test(getComputedStyle(el).overflowY)) return el;
      el = el.parentElement;
    }
    return null;
  }

  // Focusing an input makes the browser reveal it by scrolling ancestors, and an
  // overflow:hidden box still scrolls programmatically. An offset left behind
  // there moves the whole shell without touching the document scroll, so name
  // every ancestor of the composer that is not at zero.
  function scrolled() {
    var out = [];
    var el = one('[contenteditable="true"]');
    for (var i = 0; el && i < 15 && el !== document.documentElement; i++) {
      if (el.scrollTop || el.scrollLeft) {
        out.push(
          el.tagName.toLowerCase() +
          (el.getAttribute("data-testid") ? "[" + el.getAttribute("data-testid") + "]" : "") +
          "." + String(el.className || "").split(/\s+/).slice(0, 2).join(".") +
          "=" + n(el.scrollTop) + "," + n(el.scrollLeft));
      }
      el = el.parentElement;
    }
    return out.length ? out.join(" ") : "-";
  }

  function heads() {
    var nodes = all('[data-testid="section-header"]');
    var out = [];
    for (var i = 0; i < nodes.length; i++) {
      var wrap = nodes[i].closest("[data-index]") || nodes[i];
      var s = getComputedStyle(wrap);
      out.push(
        (nodes[i].getAttribute("data-title") || "?") +
        " " + s.position + " top:" + s.top + " z:" + s.zIndex + " " + box(wrap));
    }
    return out.length ? out.join(" | ") : "-";
  }

  // Which element actually holds the messages is not obvious from the class names,
  // and picking the wrong one is why a chat can stay scrolled away from its last
  // message. List every scroller that has something to scroll.
  function scrollers() {
    var nodes = document.querySelectorAll("body *");
    var out = [];
    for (var i = 0; i < nodes.length && out.length < 5; i++) {
      var el = nodes[i];
      if (el.scrollHeight <= el.clientHeight + 4) continue;
      if (!/auto|scroll/.test(getComputedStyle(el).overflowY)) continue;
      out.push(
        (el.getAttribute("data-testid") || el.tagName.toLowerCase() +
          "." + String(el.className || "").split(/\s+/)[0]) +
        ":st" + n(el.scrollTop) + "/sh" + el.scrollHeight + "/ch" + el.clientHeight);
    }
    return out.length ? out.join(" ") : "-";
  }

  function stickyOffset() {
    var el = one('[data-testid="history-search-input"]');
    el = el && el.closest(".sticky");
    return el ? n(el.getBoundingClientRect().height) : "-";
  }

  function state() {
    var de = document.documentElement;
    var sc = scroller();
    return [
      "vv=" + n(vv.width) + "x" + n(vv.height) + "+" + n(vv.offsetTop) + "," + n(vv.offsetLeft),
      "pageTop=" + n(vv.pageTop),
      "scale=" + vv.scale,
      "win=" + window.innerWidth + "x" + window.innerHeight,
      "dch=" + de.clientHeight,
      "dsh=" + de.scrollHeight,
      "dst=" + de.scrollTop,
      "sy=" + n(window.scrollY),
      "kb=" + (de.style.getPropertyValue("--agy-bottom") || "-"),
      "outer=" + all(OUTER).length + ":" + box(one(OUTER)),
      "inner=" + all(INNER).length + ":" + box(one(INNER)),
      "nav=" + box(one('[data-testid="mobile-open-settings"]')),
      "comp=" + box(one('[contenteditable="true"]')),
      "scr=" + (sc ? "st" + n(sc.scrollTop) + "/sh" + sc.scrollHeight + "/ch" + sc.clientHeight : "-"),
      "panned=" + scrolled(),
      "scrollers=[" + scrollers() + "]",
      "stickyTop=" + stickyOffset(),
      "heads=[" + heads() + "]"
    ].join(" ");
  }

  var lines = null;
  var startedAt = 0;
  var deadline = 0;
  var raf = 0;
  var last = "";

  function push(ev) {
    var s = state();
    if (ev === "raf" && s === last) return;
    last = s;
    // Safari drops a keepalive body over 64KB, so stay well inside one request.
    if (lines.length > 120) return;
    lines.push("  t=" + pad(n(performance.now() - startedAt), 5) + " ev=" + padR(ev, 9) + " " + s);
  }

  function flush() {
    var body = lines.join("\n") + "\n";
    lines = null;
    try {
      fetch(ENDPOINT, {
        method: "POST",
        headers: { "Content-Type": "text/plain" },
        credentials: "same-origin",
        body: body
      });
    } catch (e) {}
  }

  function loop() {
    push("raf");
    if (performance.now() < deadline) {
      raf = requestAnimationFrame(loop);
      return;
    }
    raf = 0;
    flush();
  }

  function begin(ev, ms) {
    if (!lines) {
      lines = [];
      startedAt = performance.now();
      last = "";
      episodes++;
      lines.push(
        "=== " + new Date().toISOString() +
        " session=" + session + " ep=" + episodes + " trigger=" + ev +
        " standalone=" + (navigator.standalone === true) +
        "/" + window.matchMedia("(display-mode:standalone)").matches +
        " screen=" + screen.width + "x" + screen.height +
        " dpr=" + window.devicePixelRatio +
        " env(t/r/b/l)=" + insets() +
        " ua=" + navigator.userAgent);
    }
    push(ev);
    var until = performance.now() + ms;
    if (until > deadline) deadline = until;
    if (!raf) raf = requestAnimationFrame(loop);
  }

  function editable(t) {
    return !!t && (t.tagName === "INPUT" || t.tagName === "TEXTAREA" || t.isContentEditable);
  }

  vv.addEventListener("resize", function () { begin("vv-resize", 700); });
  vv.addEventListener("scroll", function () { begin("vv-scroll", 400); });
  window.addEventListener("focusin", function (e) {
    if (editable(e.target)) begin("focusin", 1500);
  });
  window.addEventListener("focusout", function (e) {
    if (editable(e.target)) begin("focusout", 1200);
  });
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
