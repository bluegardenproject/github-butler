package update

import (
	"path/filepath"
	"testing"
)

func TestCompare(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{"equal plain", "0.1.0", "0.1.0", 0},
		{"equal with v", "v0.1.0", "0.1.0", 0},
		{"a less patch", "0.1.0", "0.1.1", -1},
		{"a greater patch", "0.1.2", "0.1.1", 1},
		{"minor bump", "0.1.9", "0.2.0", -1},
		{"major bump", "0.9.0", "1.0.0", -1},
		{"different lengths", "1.0", "1.0.0", 0},
		{"different lengths newer", "1.0", "1.0.1", -1},
		{"v prefix both", "v1.2.3", "v1.2.4", -1},
		{"dev vs released", "dev", "0.1.0", -1},
		{"released vs dev", "v0.1.0", "dev", 1},
		{"both dev", "dev", "dev", 0},
		{"unknown vs released", "unknown", "0.1.0", -1},
		{"empty vs released", "", "0.1.0", -1},
		{"two-digit numeric", "0.10.0", "0.9.0", 1},
		{"two-digit lex would be wrong", "0.2.0", "0.10.0", -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Compare(tt.a, tt.b); got != tt.want {
				t.Errorf("Compare(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestCompareSymmetry(t *testing.T) {
	pairs := [][2]string{
		{"0.1.0", "0.2.0"},
		{"v1.0.0", "v1.0.1"},
		{"dev", "0.1.0"},
	}
	for _, p := range pairs {
		ab := Compare(p[0], p[1])
		ba := Compare(p[1], p[0])
		if ab != -ba {
			t.Errorf("Compare(%q, %q)=%d but Compare(%q, %q)=%d (not symmetric)",
				p[0], p[1], ab, p[1], p[0], ba)
		}
	}
}

func TestInstalledBinaryPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got, err := InstalledBinaryPath()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(home, ".github-butler", "github-butler"); got != want {
		t.Errorf("InstalledBinaryPath = %q; want %q", got, want)
	}
}
