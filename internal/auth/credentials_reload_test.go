package auth

import (
	"path/filepath"
	"testing"
)

// A password set by the passwd command reaches an already-running server only if
// Verify re-reads the file, so this guards against the old password continuing to
// work after a rotation.
func TestVerifyPicksUpAPasswordSetByAnotherProcess(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials.json")

	serving, _, err := LoadOrCreateCredentials(path, "first-password")
	if err != nil {
		t.Fatal(err)
	}
	if !serving.Verify("first-password") {
		t.Fatal("the password it was created with should verify")
	}

	rotating, err := LoadCredentials(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := rotating.Set("second-password"); err != nil {
		t.Fatal(err)
	}

	if serving.Verify("first-password") {
		t.Error("the rotated-away password still verifies")
	}
	if !serving.Verify("second-password") {
		t.Error("the newly set password does not verify")
	}
}
