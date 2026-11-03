package mizan

import (
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
)

func near(a, b ui.Color, tol int) bool {
	d := func(x, y uint8) int { return int(math.Abs(float64(x) - float64(y))) }
	return d(a.R, b.R) <= tol && d(a.G, b.G) <= tol && d(a.B, b.B) <= tol && d(a.A, b.A) <= tol
}

// Mizan's own accent given as the brand changes nothing.
func TestAccentIdentity(t *testing.T) {
	for _, dark := range []bool{false, true} {
		base := tokensFor(SkinBevel, dark, Md, ui.Color{})
		same := tokensFor(SkinBevel, dark, Md, oklch(0.56, 0.19, 256, 1))
		for name, pair := range map[string][2]ui.Color{
			"500":      {base.Colors.ColorAccent500, same.Colors.ColorAccent500},
			"accent":   {base.Colors.Accent, same.Colors.Accent},
			"selected": {base.Colors.FillSelected, same.Colors.FillSelected},
			"ring":     {base.Shadows.Ring[0].Color, same.Shadows.Ring[0].Color},
			"focus":    {base.Shadows.FocusRing[1].Color, same.Shadows.FocusRing[1].Color},
			"led":      {base.Colors.LedColor, same.Colors.LedColor},
		} {
			if !near(pair[0], pair[1], 2) {
				t.Errorf("dark=%v %s: %v, want %v", dark, name, pair[1], pair[0])
			}
		}
	}
}

// Another accent takes its hue, keeps Mizan's lightness steps, and leaves
// the focus ring's gap and the neutrals alone.
func TestAccentDerives(t *testing.T) {
	orange := oklch(0.7, 0.18, 50, 1)
	for _, dark := range []bool{false, true} {
		base := tokensFor(SkinMizan, dark, Md, ui.Color{})
		got := tokensFor(SkinMizan, dark, Md, orange)
		for _, pair := range [][2]ui.Color{
			{base.Colors.ColorAccent400, got.Colors.ColorAccent400},
			{base.Colors.ColorAccent500, got.Colors.ColorAccent500},
			{base.Colors.ColorAccent600, got.Colors.ColorAccent600},
			{base.Colors.ColorAccent700, got.Colors.ColorAccent700},
		} {
			lb, _, _ := toOklch(pair[0])
			l, _, h := toOklch(pair[1])
			if math.Abs(l-lb) > 0.01 {
				t.Errorf("dark=%v: lightness %.3f, want %.3f", dark, l, lb)
			}
			if math.Abs(h-50) > 3 {
				t.Errorf("dark=%v: hue %.1f, want 50", dark, h)
			}
		}
		if got.Shadows.FocusRing[0].Color != base.Shadows.FocusRing[0].Color {
			t.Errorf("dark=%v: the focus ring's gap changed", dark)
		}
		if got.Colors.Text != base.Colors.Text || got.Colors.C1 != base.Colors.C1 {
			t.Errorf("dark=%v: neutrals or series changed", dark)
		}
		if got.theme.Accent != got.Colors.Accent {
			t.Errorf("dark=%v: MyGo's theme lacks the accent", dark)
		}
	}
}

// Scopes inherit the brand accent.
func TestAccentInherited(t *testing.T) {
	green := oklch(0.6, 0.15, 150, 1)
	var inner *Tokens
	ui.Render(func(c *ui.Context) {
		Root(c, Options{Accent: green})
		Scope(c, Options{Size: Sm}).Children(func() { inner = From(c) })
	}, 200, 100, 1)
	if inner == nil || inner.accent != green || inner.Size != Sm {
		t.Fatalf("scope did not inherit the accent: %+v", inner)
	}
}
