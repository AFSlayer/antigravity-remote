package patches

import (
	"bytes"
	"strings"
	"testing"
)

// syntheticBundle builds a document that contains every main.js anchor once,
// separated by filler. It exercises the patch engine without shipping any of
// Antigravity's compiled code in this repository; the anchors themselves are
// checked against the real bundle by TestAnchorsMatchLiveBundle and by
// "agy-remote doctor".
func syntheticBundle() []byte {
	var buf bytes.Buffer

	buf.WriteString("(function(){var filler=0;\n")
	for _, p := range All() {
		if p.Target != MainJS {
			continue
		}

		buf.WriteString("/* " + p.ID + " */\n")
		switch p.Kind {
		case Literal:
			buf.WriteString(p.Find)
		case Regexp:
			buf.WriteString(regexpFixtures[p.ID])
		}
		buf.WriteString("\nfiller++;\n")
	}
	buf.WriteString("})();\n")

	return buf.Bytes()
}

// regexpFixtures supplies a synthetic match for each regexp patch, since a
// regexp has no literal to reuse.
var regexpFixtures = map[string]string{
	"model-effort-submenu": ",onClick:()=>{var y=\nv.byEffort.get(w);y&&b(y)}",
	"sign-in-button":       `rightElement:x.createElement(tz,{variant:"primary",onClick:()=>` + "\n" + `b.showLoginFlow()},"Sign In")`,
}

func fullOptions() Options {
	return Options{MobileUX: true, WorkspaceRoot: "/home/ubuntu/workspace", CacheKey: "testkey"}
}

func TestEachAnchorAppearsOnceInSyntheticBundle(t *testing.T) {
	body := syntheticBundle()

	for _, p := range All() {
		if p.Target != MainJS {
			continue
		}

		switch p.Kind {
		case Literal:
			if n := bytes.Count(body, []byte(p.Find)); n != 1 {
				t.Errorf("%s: expected exactly 1 anchor match, got %d", p.ID, n)
			}
		case Regexp:
			if n := len(p.FindRe.FindAll(body, -1)); n != 1 {
				t.Errorf("%s: expected exactly 1 regexp match, got %d", p.ID, n)
			}
		}
	}
}

func TestAllMainJSPatchesApply(t *testing.T) {
	out, report := Apply(MainJS, syntheticBundle(), fullOptions())

	for _, r := range report.Missing() {
		t.Errorf("%s: %s", r.ID, r.Status)
	}

	for _, p := range All() {
		if p.Target != MainJS || p.Kind != Literal {
			continue
		}
		if bytes.Contains(out, []byte(p.Find)) {
			t.Errorf("%s: anchor still present after patching", p.ID)
		}
	}
}

func TestPatchedContentIsCorrect(t *testing.T) {
	out, _ := Apply(MainJS, syntheticBundle(), fullOptions())
	body := string(out)

	want := []string{
		`window.location.origin`,
		`window.matchMedia("(pointer:coarse)")`,
		`var vz=function(){return null};var vzDisabled=(`,
		`function wmb(){return null};function wmbDisabled(`,
		`initialPath:b?b.fsPath:"/home/ubuntu/workspace",fetchDirectoryContents:`,
		`var e=(window.matchMedia&&window.matchMedia("(pointer:coarse)").matches)||!!this.storageService`,
		`isMobileNew=Boolean(window.matchMedia&&window.matchMedia("(pointer:coarse)").matches&&sec)`,
		`!a.location.search?.section`,
		`onClick:()=>c({clearSection:!0})`,
	}
	for _, w := range want {
		if !strings.Contains(body, w) {
			t.Errorf("missing expected replacement: %s", w)
		}
	}

	if strings.Contains(body, `v.byEffort.get(w);y&&b(y)}`) {
		t.Error("model-effort onClick handler was not removed")
	}
	if !strings.Contains(body, `onClick:()=>{window.location.href="/__agy/signin"}`) {
		t.Error("sign-in button was not redirected")
	}
	if strings.Contains(body, "showLoginFlow()") {
		t.Error("stub showLoginFlow call still wired to the button")
	}
	if strings.Contains(body, `RK({to:"/onboarding"`) {
		t.Error("onboarding redirect was not removed")
	}
}

func TestMobileUXDisabledSkipsMobilePatches(t *testing.T) {
	_, report := Apply(MainJS, syntheticBundle(), Options{MobileUX: false})

	byID := map[string]Result{}
	for _, r := range report {
		byID[r.ID] = r
	}

	for _, id := range []string{
		"mobile-enter-newline", "hide-mic-button", "model-effort-submenu",
		"mobile-skip-notification-prompt",
	} {
		if got := byID[id].Status; got != StatusDisabled {
			t.Errorf("%s: want disabled, got %s", id, got)
		}
	}
	if got := byID["base-url-origin"].Status; got != StatusApplied {
		t.Errorf("base-url-origin: want applied, got %s", got)
	}
	if got := byID["folder-picker-initial-path"].Status; got != StatusDisabled {
		t.Errorf("folder-picker-initial-path: want disabled without workspace root, got %s", got)
	}
}

func TestMissingAnchorIsReported(t *testing.T) {
	_, report := Apply(MainJS, []byte("nothing to see here"), fullOptions())

	if len(report.Missing()) == 0 {
		t.Fatal("expected missing anchors to be reported")
	}
	if len(report.MissingRequired()) != 2 {
		t.Errorf("want 2 missing required patches, got %d", len(report.MissingRequired()))
	}
}

func TestHTMLInjection(t *testing.T) {
	html := []byte(`<!doctype html><html><head><title>x</title>` +
		`<meta name="viewport" content="width=device-width, initial-scale=1.0, viewport-fit=cover, maximum-scale=1.0" />` +
		`</head><body><script src="/main.js"></script></body></html>`)
	out, report := Apply(HTML, html, fullOptions())
	body := string(out)

	for _, r := range report.Missing() {
		t.Errorf("unexpected missing HTML patch %s", r.ID)
	}

	want := []string{
		`id="agy-touch-action"`,
		`id="agy-safe-area"`,
		`id="agy-keyboard-detect"`,
		`id="agy-signin-banner"`,
		`/__agy/api/signin/status`,
		`href="/apple-touch-icon.png"`,
		`src="/main.js?agy=testkey"`,
	}
	for _, w := range want {
		if !strings.Contains(body, w) {
			t.Errorf("missing %s in output", w)
		}
	}

	if strings.Index(body, "agy-touch-action") > strings.Index(body, "</head>") {
		t.Error("injected styles must land inside <head>")
	}
}

func TestHTMLWithoutHeadStillInjects(t *testing.T) {
	out, report := Apply(HTML, []byte(`<div id="root"></div>`), fullOptions())
	for _, r := range report.Missing() {
		if r.ID != "cache-bust" {
			t.Errorf("unexpected missing patch %s", r.ID)
		}
	}
	if !strings.Contains(string(out), "agy-touch-action") {
		t.Error("expected injection to fall back to prepending")
	}
}

func TestCacheKeyChangesWithPatchSet(t *testing.T) {
	a := CacheKey("1.0.0", Options{MobileUX: true})
	b := CacheKey("1.0.0", Options{MobileUX: false})
	c := CacheKey("1.0.1", Options{MobileUX: true})

	if a == b {
		t.Error("cache key must change when the enabled patch set changes")
	}
	if a == c {
		t.Error("cache key must change when the version changes")
	}
	if a != CacheKey("1.0.0", Options{MobileUX: true}) {
		t.Error("cache key must be stable for identical inputs")
	}
}

func TestEveryPatchIsWellFormed(t *testing.T) {
	seen := map[string]bool{}

	for _, p := range All() {
		if p.ID == "" {
			t.Error("every patch needs an ID")
			continue
		}
		if seen[p.ID] {
			t.Errorf("%s: duplicate patch ID", p.ID)
		}
		seen[p.ID] = true

		if p.Desc == "" {
			t.Errorf("%s: needs a user-facing Desc", p.ID)
		}

		switch p.Kind {
		case Literal:
			if p.Find == "" {
				t.Errorf("%s: literal patch needs Find", p.ID)
			}
		case Regexp:
			if p.FindRe == nil {
				t.Errorf("%s: regexp patch needs FindRe", p.ID)
			}
			if regexpFixtures[p.ID] == "" {
				t.Errorf("%s: regexp patch needs an entry in regexpFixtures", p.ID)
			}
		case InjectHead:
			if p.Replace == "" && p.ReplaceFn == nil {
				t.Errorf("%s: injection needs content", p.ID)
			}
		}
	}
}
