package theme

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

//go:embed examples/*.theme.yaml
var exampleThemes embed.FS

const manifestName = ".github-butler-theme-manifest.yaml"

type seedManifest struct {
	Files map[string]string `yaml:"files"`
}

func SeedExamples(directory string) error {
	dir, err := expandHome(directory)
	if err != nil {
		return err
	}
	if dir == "" {
		return fmt.Errorf("theme directory is empty")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating theme directory %s: %w", dir, err)
	}

	manifest, err := loadManifest(dir)
	if err != nil {
		return err
	}
	changed := false

	entries, err := exampleThemes.ReadDir("examples")
	if err != nil {
		return fmt.Errorf("reading embedded themes: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !isThemeFile(entry.Name()) {
			continue
		}
		data, err := exampleThemes.ReadFile(filepath.ToSlash(filepath.Join("examples", entry.Name())))
		if err != nil {
			return fmt.Errorf("reading embedded theme %s: %w", entry.Name(), err)
		}
		wrote, err := seedThemeFile(dir, manifest, entry.Name(), data)
		if err != nil {
			return err
		}
		changed = changed || wrote
	}

	if changed {
		return saveManifest(dir, manifest)
	}
	return nil
}

func seedThemeFile(dir string, manifest seedManifest, name string, data []byte) (bool, error) {
	target := filepath.Join(dir, name)
	newSum := checksum(data)
	oldSum := manifest.Files[name]

	current, err := os.ReadFile(target)
	switch {
	case os.IsNotExist(err):
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return false, fmt.Errorf("writing theme %s: %w", target, err)
		}
		manifest.Files[name] = newSum
		return true, nil
	case err != nil:
		return false, fmt.Errorf("reading theme %s: %w", target, err)
	}

	currentSum := checksum(current)
	if currentSum == newSum {
		if oldSum != newSum {
			manifest.Files[name] = newSum
			return true, nil
		}
		return false, nil
	}
	if oldSum != "" && currentSum == oldSum {
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return false, fmt.Errorf("updating theme %s: %w", target, err)
		}
		manifest.Files[name] = newSum
		return true, nil
	}

	nextPath, err := newThemePath(target, data)
	if err != nil {
		return false, err
	}
	if nextPath == "" {
		return false, nil
	}
	if err := os.WriteFile(nextPath, data, 0o644); err != nil {
		return false, fmt.Errorf("writing updated theme example %s: %w", nextPath, err)
	}
	return false, nil
}

func loadManifest(dir string) (seedManifest, error) {
	manifest := seedManifest{Files: map[string]string{}}
	data, err := os.ReadFile(filepath.Join(dir, manifestName))
	if os.IsNotExist(err) {
		return manifest, nil
	}
	if err != nil {
		return seedManifest{}, fmt.Errorf("reading theme manifest: %w", err)
	}
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return seedManifest{}, fmt.Errorf("parsing theme manifest: %w", err)
	}
	if manifest.Files == nil {
		manifest.Files = map[string]string{}
	}
	return manifest, nil
}

func saveManifest(dir string, manifest seedManifest) error {
	data, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("marshaling theme manifest: %w", err)
	}
	path := filepath.Join(dir, manifestName)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing theme manifest: %w", err)
	}
	return nil
}

func newThemePath(target string, data []byte) (string, error) {
	for i := 0; i < 100; i++ {
		suffix := ".new"
		if i > 0 {
			suffix = fmt.Sprintf(".new.%d", i)
		}
		candidate := target + suffix
		existing, err := os.ReadFile(candidate)
		if os.IsNotExist(err) {
			return candidate, nil
		}
		if err != nil {
			return "", fmt.Errorf("reading updated theme example %s: %w", candidate, err)
		}
		if checksum(existing) == checksum(data) {
			return "", nil
		}
	}
	return "", fmt.Errorf("could not find available .new path for %s", target)
}

func checksum(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
