package theme

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/bluegardenproject/github-butler/internal/config"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

const (
	BuiltinAuto  = "auto"
	BuiltinDark  = "neon-dark"
	BuiltinLight = "neon-light"
)

type Options struct {
	Selected   string
	Directory  string
	Colors     map[string]string
	DetectDark func() bool
}

type Choice struct {
	ID   string
	Name string
	Kind string
}

type Resolved struct {
	ID        string
	Name      string
	Colors    map[string]string
	Gradients map[string][]string
}

type spec struct {
	Name      string              `yaml:"name"`
	Base      string              `yaml:"base"`
	Colors    map[string]string   `yaml:"colors"`
	Gradients map[string][]string `yaml:"gradients"`
}

var hexColor = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
var themeIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$`)

var builtinThemes = map[string]spec{
	BuiltinDark: {
		Name: "Neon Dark",
		Colors: map[string]string{
			"neon_pink":    "#FF10F0",
			"neon_cyan":    "#00F0FF",
			"neon_magenta": "#FF00FF",
			"neon_lime":    "#39FF14",
			"neon_purple":  "#BF00FF",
			"neon_orange":  "#FF6A00",
			"neon_yellow":  "#F5FF00",
			"neon_blue":    "#1B03FF",
			"hot_pink":     "#FF2A6D",
			"black":        "#000000",
			"white":        "#FFFFFF",
			"dim":          "#6C6C80",
			"dark_bg":      "#120018",
			"selected_fg":  "#F5FF00",
			"selected_bg":  "#BF00FF",
			"chip_fg":      "#000000",
		},
		Gradients: map[string][]string{
			"title":     {"#FF10F0", "#FF00FF", "#BF00FF", "#00F0FF"},
			"countdown": {"#FF10F0", "#BF00FF", "#00F0FF"},
			"header":    {"#00F0FF", "#FF10F0"},
		},
	},
	BuiltinLight: {
		Name: "Neon Light",
		Colors: map[string]string{
			"neon_pink":    "#A0007A",
			"neon_cyan":    "#005F73",
			"neon_magenta": "#7B1FA2",
			"neon_lime":    "#006B3C",
			"neon_purple":  "#5A2D82",
			"neon_orange":  "#9A4D00",
			"neon_yellow":  "#7A5C00",
			"neon_blue":    "#0033A0",
			"hot_pink":     "#B00020",
			"black":        "#111111",
			"white":        "#FFFFFF",
			"dim":          "#5F6470",
			"dark_bg":      "#FFFFFF",
			"selected_fg":  "#FFFFFF",
			"selected_bg":  "#005F73",
			"chip_fg":      "#FFFFFF",
		},
		Gradients: map[string][]string{
			"title":     {"#A0007A", "#7B1FA2", "#005F73"},
			"countdown": {"#A0007A", "#5A2D82", "#005F73"},
			"header":    {"#005F73", "#A0007A"},
		},
	},
}

var colorAliases = map[string]string{
	"pink":                "neon_pink",
	"cyan":                "neon_cyan",
	"magenta":             "neon_magenta",
	"lime":                "neon_lime",
	"purple":              "neon_purple",
	"orange":              "neon_orange",
	"yellow":              "neon_yellow",
	"blue":                "neon_blue",
	"danger":              "hot_pink",
	"success":             "neon_lime",
	"warning":             "neon_yellow",
	"info":                "neon_cyan",
	"accent":              "neon_magenta",
	"background":          "dark_bg",
	"selected_foreground": "selected_fg",
	"selected_background": "selected_bg",
	"chip_foreground":     "chip_fg",
}

func OptionsFromConfig(cfg config.Theme) Options {
	return Options{
		Selected:   cfg.Selected,
		Directory:  cfg.Directory,
		Colors:     cfg.Colors,
		DetectDark: lipgloss.HasDarkBackground,
	}
}

func Activate(opts Options) error {
	resolved, err := Resolve(opts)
	if err != nil {
		return err
	}
	apply(resolved)
	return nil
}

func Resolve(opts Options) (Resolved, error) {
	selected := strings.TrimSpace(opts.Selected)
	if selected == "" {
		selected = BuiltinAuto
	}
	if opts.DetectDark == nil {
		opts.DetectDark = lipgloss.HasDarkBackground
	}

	baseID := selected
	if selected == BuiltinAuto {
		baseID = builtinForBackground(opts.DetectDark())
	}

	base, ok := builtinThemes[baseID]
	if !ok {
		user, err := loadUserTheme(opts.Directory, selected)
		if err != nil {
			return Resolved{}, err
		}
		if user.Base == "" {
			user.Base = BuiltinDark
		}
		baseID = user.Base
		if baseID == BuiltinAuto {
			baseID = builtinForBackground(opts.DetectDark())
		}
		base, ok = builtinThemes[baseID]
		if !ok {
			return Resolved{}, fmt.Errorf("theme %q has unknown base %q", selected, user.Base)
		}
		base = cloneSpec(base)
		if err := mergeSpec(&base, user); err != nil {
			return Resolved{}, fmt.Errorf("theme %q: %w", selected, err)
		}
		if base.Name == "" {
			base.Name = selected
		}
	} else if selected == BuiltinAuto {
		base = cloneSpec(base)
		base.Name = "Auto (" + base.Name + ")"
	} else {
		base = cloneSpec(base)
	}

	resolved := Resolved{
		ID:        selected,
		Name:      base.Name,
		Colors:    cloneColors(base.Colors),
		Gradients: cloneGradients(base.Gradients),
	}
	if err := mergeColors(resolved.Colors, opts.Colors); err != nil {
		return Resolved{}, err
	}
	if err := validateResolved(resolved); err != nil {
		return Resolved{}, err
	}
	return resolved, nil
}

func Choices(directory string) []Choice {
	choices := []Choice{
		{ID: BuiltinAuto, Name: "Auto", Kind: "built-in"},
		{ID: BuiltinDark, Name: builtinThemes[BuiltinDark].Name, Kind: "built-in"},
		{ID: BuiltinLight, Name: builtinThemes[BuiltinLight].Name, Kind: "built-in"},
	}

	dir, err := expandHome(directory)
	if err != nil {
		return choices
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return choices
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, entry := range entries {
		if entry.IsDir() || !isThemeFile(entry.Name()) {
			continue
		}
		id := themeID(entry.Name())
		if !validThemeID(id) {
			continue
		}
		name := id
		if parsed, err := readThemeFile(filepath.Join(dir, entry.Name())); err == nil && parsed.Name != "" {
			name = parsed.Name
		}
		choices = append(choices, Choice{ID: id, Name: name, Kind: "custom"})
	}
	return choices
}

func builtinForBackground(dark bool) string {
	if dark {
		return BuiltinDark
	}
	return BuiltinLight
}

func mergeSpec(dst *spec, override spec) error {
	if override.Name != "" {
		dst.Name = override.Name
	}
	if err := mergeColors(dst.Colors, override.Colors); err != nil {
		return err
	}
	for key, stops := range override.Gradients {
		normalized := normalizeKey(key)
		if !isKnownGradient(normalized) {
			return fmt.Errorf("unknown gradient %q", key)
		}
		if len(stops) == 0 {
			return fmt.Errorf("gradient %q must contain at least one color", key)
		}
		for _, stop := range stops {
			if !isHexColor(stop) {
				return fmt.Errorf("gradient %q contains invalid color %q", key, stop)
			}
		}
		dst.Gradients[normalized] = normalizeHexSlice(stops)
	}
	return nil
}

func mergeColors(dst map[string]string, override map[string]string) error {
	for key, value := range override {
		normalized, ok := canonicalColorKey(key)
		if !ok {
			return fmt.Errorf("unknown color %q", key)
		}
		if !isHexColor(value) {
			return fmt.Errorf("color %q must be a #RRGGBB hex color", key)
		}
		dst[normalized] = strings.ToUpper(strings.TrimSpace(value))
	}
	return nil
}

func loadUserTheme(directory, selected string) (spec, error) {
	if strings.ContainsAny(selected, `/\`) || !validThemeID(themeID(selected)) {
		return spec{}, fmt.Errorf("theme %q is not a valid theme id", selected)
	}
	dir, err := expandHome(directory)
	if err != nil {
		return spec{}, err
	}
	candidates := []string{selected}
	if !isThemeFile(selected) {
		candidates = append(candidates, selected+".theme.yaml", selected+".theme.yml")
	}
	for _, candidate := range candidates {
		path := filepath.Join(dir, candidate)
		parsed, err := readThemeFile(path)
		if err == nil {
			return parsed, nil
		}
		if !os.IsNotExist(err) {
			return spec{}, err
		}
	}
	return spec{}, fmt.Errorf("theme %q was not found in %s", selected, dir)
}

func readThemeFile(path string) (spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return spec{}, err
	}
	var parsed spec
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return spec{}, fmt.Errorf("parsing %s: %w", path, err)
	}
	return parsed, nil
}

func apply(resolved Resolved) {
	NeonPink = lipgloss.Color(resolved.Colors["neon_pink"])
	NeonCyan = lipgloss.Color(resolved.Colors["neon_cyan"])
	NeonMagenta = lipgloss.Color(resolved.Colors["neon_magenta"])
	NeonLime = lipgloss.Color(resolved.Colors["neon_lime"])
	NeonPurple = lipgloss.Color(resolved.Colors["neon_purple"])
	NeonOrange = lipgloss.Color(resolved.Colors["neon_orange"])
	NeonYellow = lipgloss.Color(resolved.Colors["neon_yellow"])
	NeonBlue = lipgloss.Color(resolved.Colors["neon_blue"])
	HotPink = lipgloss.Color(resolved.Colors["hot_pink"])
	Black = lipgloss.Color(resolved.Colors["black"])
	White = lipgloss.Color(resolved.Colors["white"])
	Dim = lipgloss.Color(resolved.Colors["dim"])
	DarkBg = lipgloss.Color(resolved.Colors["dark_bg"])
	selectedForeground = lipgloss.Color(resolved.Colors["selected_fg"])
	selectedBackground = lipgloss.Color(resolved.Colors["selected_bg"])
	chipForeground = lipgloss.Color(resolved.Colors["chip_fg"])

	TitleStops = toLipglossColors(resolved.Gradients["title"])
	CountdownStops = toLipglossColors(resolved.Gradients["countdown"])
	HeaderStops = toLipglossColors(resolved.Gradients["header"])
	rebuildStyles()
}

func validateResolved(resolved Resolved) error {
	for key := range builtinThemes[BuiltinDark].Colors {
		value, ok := resolved.Colors[key]
		if !ok {
			return fmt.Errorf("theme is missing color %q", key)
		}
		if !isHexColor(value) {
			return fmt.Errorf("color %q must be a #RRGGBB hex color", key)
		}
	}
	for key := range builtinThemes[BuiltinDark].Gradients {
		stops := resolved.Gradients[key]
		if len(stops) == 0 {
			return fmt.Errorf("theme is missing gradient %q", key)
		}
		for _, stop := range stops {
			if !isHexColor(stop) {
				return fmt.Errorf("gradient %q contains invalid color %q", key, stop)
			}
		}
	}
	return nil
}

func canonicalColorKey(key string) (string, bool) {
	normalized := normalizeKey(key)
	if alias, ok := colorAliases[normalized]; ok {
		normalized = alias
	}
	_, ok := builtinThemes[BuiltinDark].Colors[normalized]
	return normalized, ok
}

func isKnownGradient(key string) bool {
	_, ok := builtinThemes[BuiltinDark].Gradients[key]
	return ok
}

func isHexColor(value string) bool {
	return hexColor.MatchString(strings.TrimSpace(value))
}

func normalizeKey(key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	key = strings.ReplaceAll(key, "-", "_")
	return key
}

func normalizeHexSlice(values []string) []string {
	out := make([]string, len(values))
	for i, value := range values {
		out[i] = strings.ToUpper(strings.TrimSpace(value))
	}
	return out
}

func cloneColors(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func cloneGradients(in map[string][]string) map[string][]string {
	out := make(map[string][]string, len(in))
	for key, values := range in {
		out[key] = append([]string(nil), values...)
	}
	return out
}

func cloneSpec(in spec) spec {
	in.Colors = cloneColors(in.Colors)
	in.Gradients = cloneGradients(in.Gradients)
	return in
}

func toLipglossColors(values []string) []lipgloss.Color {
	out := make([]lipgloss.Color, len(values))
	for i, value := range values {
		out[i] = lipgloss.Color(value)
	}
	return out
}

func isThemeFile(name string) bool {
	return strings.HasSuffix(name, ".theme.yaml") || strings.HasSuffix(name, ".theme.yml")
}

func themeID(name string) string {
	name = strings.TrimSuffix(name, ".theme.yaml")
	return strings.TrimSuffix(name, ".theme.yml")
}

func validThemeID(id string) bool {
	return themeIDPattern.MatchString(id)
}

func expandHome(path string) (string, error) {
	if path == "" || !strings.HasPrefix(path, "~") {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, strings.TrimPrefix(path, "~")), nil
}
