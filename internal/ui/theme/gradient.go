package theme

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/lipgloss"
)

// Gradient colors each rune of s by linearly interpolating RGB between
// stops. Zero-width runes (e.g. combining marks) are appended to the
// previous rune's styling so width stays correct.
func Gradient(s string, stops ...lipgloss.Color) string {
	runes := []rune(s)
	if len(runes) == 0 || len(stops) == 0 {
		return s
	}
	if len(stops) == 1 {
		return lipgloss.NewStyle().Foreground(stops[0]).Render(s)
	}

	rgbs := make([][3]int, len(stops))
	for i, c := range stops {
		rgbs[i] = hexToRGB(string(c))
	}

	total := len(runes)
	out := make([]byte, 0, total*20)
	segments := len(rgbs) - 1

	for i, r := range runes {
		var t float64
		if total > 1 {
			t = float64(i) / float64(total-1)
		}
		seg := int(t * float64(segments))
		if seg >= segments {
			seg = segments - 1
		}
		localT := t*float64(segments) - float64(seg)
		c := interp(rgbs[seg], rgbs[seg+1], localT)
		out = append(out, []byte(lipgloss.NewStyle().
			Foreground(lipgloss.Color(rgbToHex(c))).
			Render(string(r)))...)
	}
	return string(out)
}

func hexToRGB(h string) [3]int {
	if len(h) != 7 || h[0] != '#' {
		return [3]int{255, 255, 255}
	}
	r, _ := strconv.ParseInt(h[1:3], 16, 0)
	g, _ := strconv.ParseInt(h[3:5], 16, 0)
	b, _ := strconv.ParseInt(h[5:7], 16, 0)
	return [3]int{int(r), int(g), int(b)}
}

func rgbToHex(c [3]int) string {
	return fmt.Sprintf("#%02X%02X%02X", c[0], c[1], c[2])
}

func interp(a, b [3]int, t float64) [3]int {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return [3]int{
		int(float64(a[0]) + (float64(b[0])-float64(a[0]))*t),
		int(float64(a[1]) + (float64(b[1])-float64(a[1]))*t),
		int(float64(a[2]) + (float64(b[2])-float64(a[2]))*t),
	}
}
