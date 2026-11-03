package mizan

import (
	"math"
	"slices"

	"github.com/egoist/mygo/ui"
)

// Mizan's accent is a ramp of one hue (256, blue): steps 400 to 700 and
// the colors made of them (the selected-row fill, the field and focus
// rings). A brand accent replaces that hue and scales the
// chroma; the lightness of every step stays Mizan's, so white text on a
// solid button and the rings keep their contrast in both themes.

const (
	// baseHue and baseChroma are those of Mizan's own accent (step 600).
	baseHue    = 256
	baseChroma = 0.19
)

// accentShift turns colors of Mizan's accent family into those of another
// accent: it rotates their hue by the difference between the two hues and
// scales their chroma by the ratio of the two chromas.
type accentShift struct{ dh, k float64 }

func newAccentShift(accent ui.Color) accentShift {
	_, c, h := toOklch(accent)
	return accentShift{dh: h - baseHue, k: c / baseChroma}
}

// color returns c in the new accent, reduced in chroma until it fits sRGB.
func (s accentShift) color(c ui.Color) ui.Color {
	l, ch, h := toOklch(c)
	return fromOklch(l, ch*s.k, h+s.dh, float64(c.A)/255)
}

func (s accentShift) shadow(sh Shadow, layers ...int) Shadow {
	out := append(Shadow(nil), sh...)
	for i := range out {
		if len(layers) == 0 || slices.Contains(layers, i) {
			out[i].Color = s.color(out[i].Color)
		}
	}
	return out
}

// withAccent gives t the brand accent: the accent ramp and every color
// derived from it.
func (t *Tokens) withAccent(accent ui.Color) {
	s := newAccentShift(accent)
	c := &t.Colors
	for _, p := range []*ui.Color{
		&c.ColorAccent400, &c.ColorAccent500, &c.ColorAccent600, &c.ColorAccent700, &c.Accent,
		&c.FillSelected,
	} {
		*p = s.color(*p)
	}
	sh := &t.Shadows
	sh.Ring = s.shadow(sh.Ring)
	// the first layer of the focus ring is the gap in the background color
	sh.FocusRing = s.shadow(sh.FocusRing, 1)
}

func toOklch(c ui.Color) (l, ch, h float64) {
	l, a, b := toOklab(c)
	ch = math.Hypot(a, b)
	h = math.Atan2(b, a) * 180 / math.Pi
	if h < 0 {
		h += 360
	}
	return l, ch, h
}

// fromOklch returns the sRGB color of oklch(l c h / alpha); when it lies
// outside sRGB, the chroma is reduced until it fits (as CSS gamut mapping
// keeps the lightness and the hue).
func fromOklch(l, ch, h, alpha float64) ui.Color {
	rad := h * math.Pi / 180
	ab := func(c float64) (float64, float64) { return c * math.Cos(rad), c * math.Sin(rad) }
	if a, b := ab(ch); inGamut(l, a, b) {
		return fromOklab(l, a, b, alpha)
	}
	lo, hi := 0.0, ch
	for range 24 {
		mid := (lo + hi) / 2
		if a, b := ab(mid); inGamut(l, a, b) {
			lo = mid
		} else {
			hi = mid
		}
	}
	a, b := ab(lo)
	return fromOklab(l, a, b, alpha)
}

func inGamut(l, a, b float64) bool {
	lm := l + 0.3963377774*a + 0.2158037573*b
	mm := l - 0.1055613458*a - 0.0638541728*b
	sm := l - 0.0894841775*a - 1.2914855480*b
	lm, mm, sm = lm*lm*lm, mm*mm*mm, sm*sm*sm
	const eps = 1e-4
	for _, v := range []float64{
		4.0767416621*lm - 3.3077115913*mm + 0.2309699292*sm,
		-1.2684380046*lm + 2.6097574011*mm - 0.3413193965*sm,
		-0.0041960863*lm - 0.7034186147*mm + 1.7076147010*sm,
	} {
		if v < -eps || v > 1+eps {
			return false
		}
	}
	return true
}

// Swatch returns the accent that a brand color gives in t's theme: what
// Options.Accent would make Colors.Accent, for showing a choice of colors.
func Swatch(t *Tokens, accent ui.Color) ui.Color {
	base := themeColors[b2i(t.Dark)].Accent
	if accent == (ui.Color{}) {
		return base
	}
	return newAccentShift(accent).color(base)
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}
