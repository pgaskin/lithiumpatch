package fonts

import (
	"testing"
)

func TestEmbeddedFonts(t *testing.T) {
	all := All()
	if len(all) == 0 {
		t.Fatalf("expected embedded fonts, got none")
	}

	t.Logf("Total loaded font families: %d", len(all))

	for _, f := range all {
		if f.Name == "" {
			t.Errorf("font has empty Name: %+v", f)
		}
		if f.Base == "" {
			t.Errorf("font %q has empty Base", f.Name)
		}
		if f.Regular == nil {
			t.Errorf("font %q is missing Regular face", f.Name)
		}
		if f.Script == 0 {
			t.Errorf("font %q has no supported scripts", f.Name)
		}
		t.Logf("✓ %s", f.String())
	}
}
