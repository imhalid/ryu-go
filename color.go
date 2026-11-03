package mizan

import (
	"math"

	"github.com/egoist/mygo/ui"
)

// oklch is CSS's oklch(l c h / a) color; the tokens are written with it.
func oklch(l, c, h, a float32) ui.Color { return ui.Oklch(l, c, h).Alpha(a) }

// Mix is CSS's color-mix(in oklab, a, b t): from a towards b by t.
// Translucent colors mix premultiplied (like color-mix): Mix(c, ui.Color{},
// 0.55) is c at 45% opacity, without darkening. It is not in oklch: neutrals
// carry h 286, so mixing red with gray would shift the hue towards purple.
func Mix(a, b ui.Color, t float32) ui.Color {
	t = max(0, min(t, 1))
	aa, ba := float64(a.A)/255, float64(b.A)/255
	alpha := aa*(1-float64(t)) + ba*float64(t)
	if alpha == 0 {
		return ui.Color{}
	}
	la, a1, b1 := toOklab(a)
	lb, a2, b2 := toOklab(b)
	w1, w2 := aa*(1-float64(t)), ba*float64(t)
	l := (la*w1 + lb*w2) / alpha
	x := (a1*w1 + a2*w2) / alpha
	y := (b1*w1 + b2*w2) / alpha
	return fromOklab(l, x, y, alpha)
}

// Brightness is CSS's filter: brightness(f): it multiplies the color channels
// by f. The button's hover and pressed states lighten or darken with it.
func Brightness(c ui.Color, f float32) ui.Color {
	m := func(v uint8) uint8 { return uint8(max(0, min(float32(v)*f+0.5, 255))) }
	return ui.Color{R: m(c.R), G: m(c.G), B: m(c.B), A: c.A}
}

// capLightness is CSS's oklch(from c min(l, maxL) c h): hue and chroma stay,
// and the color darkens only if it is lighter than maxL. It is the lightness
// limit at which white text stays readable on a solid button.
func capLightness(c ui.Color, maxL float64) ui.Color {
	l, a, b := toOklab(c)
	if l <= maxL {
		return c
	}
	return fromOklab(maxL, a, b, float64(c.A)/255)
}

// lighten adds d to the oklab lightness (oklch(from c calc(l + d) c h)).
func lighten(c ui.Color, d float64) ui.Color {
	l, a, b := toOklab(c)
	return fromOklab(max(0, min(l+d, 1)), a, b, float64(c.A)/255)
}

func toOklab(c ui.Color) (l, a, b float64) {
	lin := func(v uint8) float64 {
		x := float64(v) / 255
		if x <= 0.04045 {
			return x / 12.92
		}
		return math.Pow((x+0.055)/1.055, 2.4)
	}
	r, g, bl := lin(c.R), lin(c.G), lin(c.B)
	lm := math.Cbrt(0.4122214708*r + 0.5363325363*g + 0.0514459929*bl)
	mm := math.Cbrt(0.2119034982*r + 0.6806995451*g + 0.1073969566*bl)
	sm := math.Cbrt(0.0883024619*r + 0.2817188376*g + 0.6299787005*bl)
	return 0.2104542553*lm + 0.7936177850*mm - 0.0040720468*sm,
		1.9779984951*lm - 2.4285922050*mm + 0.4505937099*sm,
		0.0259040371*lm + 0.7827717662*mm - 0.8086757660*sm
}

func fromOklab(l, a, b, alpha float64) ui.Color {
	lm := l + 0.3963377774*a + 0.2158037573*b
	mm := l - 0.1055613458*a - 0.0638541728*b
	sm := l - 0.0894841775*a - 1.2914855480*b
	lm, mm, sm = lm*lm*lm, mm*mm*mm, sm*sm*sm
	r := 4.0767416621*lm - 3.3077115913*mm + 0.2309699292*sm
	g := -1.2684380046*lm + 2.6097574011*mm - 0.3413193965*sm
	bl := -0.0041960863*lm - 0.7034186147*mm + 1.7076147010*sm
	enc := func(x float64) uint8 {
		if x <= 0.0031308 {
			x *= 12.92
		} else {
			x = 1.055*math.Pow(x, 1/2.4) - 0.055
		}
		return uint8(max(0, min(math.Round(x*255), 255)))
	}
	return ui.Color{R: enc(r), G: enc(g), B: enc(bl), A: uint8(max(0, min(math.Round(alpha*255), 255)))}
}
