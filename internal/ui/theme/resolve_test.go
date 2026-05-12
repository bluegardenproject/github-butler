package theme

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveBuiltinsAndAuto(t *testing.T) {
	dark, err := Resolve(Options{Selected: BuiltinDark})
	if err != nil {
		t.Fatal(err)
	}
	if got := dark.Colors["neon_pink"]; got != "#FF10F0" {
		t.Fatalf("dark neon_pink = %q, want current palette", got)
	}

	light, err := Resolve(Options{Selected: BuiltinLight})
	if err != nil {
		t.Fatal(err)
	}
	if got := light.Colors["selected_fg"]; got != "#FFFFFF" {
		t.Fatalf("light selected_fg = %q, want #FFFFFF", got)
	}

	autoDark, err := Resolve(Options{Selected: BuiltinAuto, DetectDark: func() bool { return true }})
	if err != nil {
		t.Fatal(err)
	}
	if got := autoDark.Colors["neon_pink"]; got != dark.Colors["neon_pink"] {
		t.Fatalf("auto dark neon_pink = %q, want %q", got, dark.Colors["neon_pink"])
	}

	autoLight, err := Resolve(Options{Selected: BuiltinAuto, DetectDark: func() bool { return false }})
	if err != nil {
		t.Fatal(err)
	}
	if got := autoLight.Colors["neon_pink"]; got != light.Colors["neon_pink"] {
		t.Fatalf("auto light neon_pink = %q, want %q", got, light.Colors["neon_pink"])
	}
}

func TestResolveConfigColorAliases(t *testing.T) {
	resolved, err := Resolve(Options{
		Selected: BuiltinDark,
		Colors: map[string]string{
			"accent":              "#123456",
			"selected-background": "#654321",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := resolved.Colors["neon_magenta"]; got != "#123456" {
		t.Fatalf("accent override = %q, want #123456", got)
	}
	if got := resolved.Colors["selected_bg"]; got != "#654321" {
		t.Fatalf("selected background override = %q, want #654321", got)
	}
}

func TestResolveUserThemeFile(t *testing.T) {
	dir := t.TempDir()
	writeTheme(t, dir, "ocean.theme.yaml", `
name: Ocean
base: neon-light
colors:
  accent: "#123456"
gradients:
  title:
    - "#123456"
    - "#654321"
`)

	resolved, err := Resolve(Options{Selected: "ocean", Directory: dir})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Name != "Ocean" {
		t.Fatalf("Name = %q, want Ocean", resolved.Name)
	}
	if got := resolved.Colors["neon_magenta"]; got != "#123456" {
		t.Fatalf("accent override = %q, want #123456", got)
	}
	if got := resolved.Gradients["title"][0]; got != "#123456" {
		t.Fatalf("title gradient first stop = %q, want #123456", got)
	}

	dark, err := Resolve(Options{Selected: BuiltinDark})
	if err != nil {
		t.Fatal(err)
	}
	if got := dark.Colors["neon_magenta"]; got != "#FF00FF" {
		t.Fatalf("custom theme mutated built-in dark accent: %q", got)
	}
}

func TestResolveUserThemeRejectsInvalidColor(t *testing.T) {
	dir := t.TempDir()
	writeTheme(t, dir, "bad.theme.yaml", `
name: Bad
base: neon-dark
colors:
  accent: "blue"
`)

	if _, err := Resolve(Options{Selected: "bad", Directory: dir}); err == nil {
		t.Fatal("Resolve succeeded, want invalid color error")
	}
}

func TestChoicesIncludesCustomThemesAfterBuiltins(t *testing.T) {
	dir := t.TempDir()
	writeTheme(t, dir, "ocean.theme.yaml", "name: Ocean\nbase: neon-light\n")

	choices := Choices(dir)
	if len(choices) < 4 {
		t.Fatalf("Choices length = %d, want at least 4", len(choices))
	}
	if choices[0].ID != BuiltinAuto || choices[1].ID != BuiltinDark || choices[2].ID != BuiltinLight {
		t.Fatalf("built-in choices = %#v", choices[:3])
	}
	if got := choices[3]; got.ID != "ocean" || got.Name != "Ocean" || got.Kind != "custom" {
		t.Fatalf("custom choice = %#v, want ocean/Ocean/custom", got)
	}
}

func TestSeedExamplesPreservesModifiedFiles(t *testing.T) {
	dir := t.TempDir()
	if err := SeedExamples(dir); err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(dir, "midnight-neon.theme.yaml")
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("seeded theme missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, manifestName)); err != nil {
		t.Fatalf("manifest missing: %v", err)
	}

	custom := []byte("name: User Modified\nbase: neon-dark\n")
	if err := os.WriteFile(target, custom, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SeedExamples(dir); err != nil {
		t.Fatal(err)
	}

	current, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(current) != string(custom) {
		t.Fatalf("modified theme was overwritten:\n%s", current)
	}
	if _, err := os.Stat(target + ".new"); err != nil {
		t.Fatalf("updated example was not written as .new: %v", err)
	}
}

func writeTheme(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
