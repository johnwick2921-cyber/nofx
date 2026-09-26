// FIX-KNOBS P1-C (DS-105, 2026-09-26) — activation must refuse the live FLAT
// layout (NOFX_RELEASE_DIR unset ⇒ repo root nofx-bin + deploy/RELEASE +
// web/dist). The live install has no versioned releases, and moving its files
// is forbidden (main-tree law), so a mutating activation verb must fail LOUD
// with the v6 deploy path printed — never a cryptic manifest.json error.
package activation

import (
	"strings"
	"testing"

	"nofx/internal/installpath"
)

func TestRefuseFlatLayout(t *testing.T) {
	t.Run("flat layout refused with v6 path printed", func(t *testing.T) {
		installpath.ResetReleaseDirForTest()
		t.Setenv("NOFX_RELEASE_DIR", "")
		err := RefuseFlatLayout()
		if err == nil {
			t.Fatal("expected refusal on the flat layout, got nil")
		}
		for _, want := range []string{
			"FLAT",
			"NOFX_RELEASE_DIR",
			"nofx-bin",
			"deploy/RELEASE",
			"web/dist",
			"v6 deploy path",
			"kill -9",
		} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("refusal must print %q; got:\n%s", want, err)
			}
		}
	})

	t.Run("versioned release dir is allowed", func(t *testing.T) {
		installpath.ResetReleaseDirForTest()
		t.Setenv("NOFX_RELEASE_DIR", t.TempDir())
		if err := RefuseFlatLayout(); err != nil {
			t.Fatalf("expected nil for a versioned install, got: %v", err)
		}
	})
}
