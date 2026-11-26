package ryugo

import (
	"math"

	"github.com/egoist/mygo/ui"
)

// InOklch returns g blended in OKLCH, as CSS's linear-gradient(in oklch, …): between two of its stops, lightness,
// chroma and hue each go their own way, the hue the shorter way round. A darker step keeps its color, where a blend in
// Oklab or sRGB grays it, and two hues blend through the vivid ones between them, not through gray. MyGo blends a
// gradient in Oklab at most, so the stops between g's own are made here, in OKLCH and in the display's widest gamut
// (ui.Oklch), and MyGo blends the short steps between them. g's From and To, with Start and End, are stops too.
func InOklch(g ui.LinearGradient) ui.LinearGradient {
	g.Stops = oklchStops(gradientStops(g.From, g.To, g.Start, g.End, g.Stops))
	g.From, g.To, g.Start, g.End, g.Oklab = ui.Color{}, ui.Color{}, 0, 0, true
	return g
}

// RadialInOklch is InOklch for a radial gradient.
func RadialInOklch(g ui.RadialGradient) ui.RadialGradient {
	g.Stops, g.Oklab = oklchStops(g.Stops), true
	return g
}

// gradientStops are a gradient's stops: its Stops, or From at Start and To at End (0 and 1 when both are 0).
func gradientStops(from, to ui.Color, start, end float32, stops []ui.GradientStop) []ui.GradientStop {
	if len(stops) > 0 {
		return stops
	}
	if start == 0 && end == 0 {
		end = 1
	}
	return []ui.GradientStop{{Color: from, At: start}, {Color: to, At: end}}
}

// maxOklchStops is how many stops a gradient blended in OKLCH has at most: MyGo draws 32.
const maxOklchStops = 32

// oklchStops are stops with stops between each two of them, blended in OKLCH.
func oklchStops(stops []ui.GradientStop) []ui.GradientStop {
	if len(stops) < 2 {
		return stops
	}
	each := min(8, (maxOklchStops-1)/(len(stops)-1)) // the steps between two stops
	out := make([]ui.GradientStop, 0, (len(stops)-1)*each+1)
	for i := 0; i+1 < len(stops); i++ {
		a, b := stops[i], stops[i+1]
		for k := range each {
			t := float32(k) / float32(each)
			out = append(out, ui.GradientStop{Color: MixOklch(a.Color, b.Color, t), At: a.At + (b.At-a.At)*t})
		}
	}
	return append(out, stops[len(stops)-1])
}

// MixOklch is CSS's color-mix(in oklch, a, b t): from a towards b by t, lightness, chroma and hue each on its own, the
// hue the shorter way round, in the display's widest gamut. A gray has no hue, nor a color with no opacity: it takes
// the other's, as CSS's powerless hue, so that a fade to clear or to gray keeps the color's hue; a clear color takes
// the other's lightness and chroma too, as a premultiplied blend does.
func MixOklch(a, b ui.Color, t float32) ui.Color {
	t = max(0, min(t, 1))
	switch {
	case t == 0:
		return a
	case t == 1:
		return b
	}
	la, ca, ha := toOklch(a)
	lb, cb, hb := toOklch(b)
	if a.A == 0 {
		la, ca = lb, cb
	}
	if b.A == 0 {
		lb, cb = la, ca
	}
	const gray = 0.02 // under this chroma the hue is noise
	switch {
	case ca < gray && cb >= gray, a.A == 0:
		ha = hb
	case cb < gray && ca >= gray, b.A == 0:
		hb = ha
	}
	dh := math.Mod(hb-ha+540, 360) - 180 // the shorter way round
	tt := float64(t)
	l, c, h := la+(lb-la)*tt, ca+(cb-ca)*tt, math.Mod(ha+dh*tt+360, 360)
	alpha := float32(a.A)/255 + (float32(b.A)/255-float32(a.A)/255)*t
	return ui.Oklch(float32(l), float32(c), float32(h)).Alpha(alpha)
}

// OklchOf is the lightness, chroma and hue of c, as CSS's oklch() writes them.
func OklchOf(c ui.Color) (l, ch, h float64) { return toOklch(c) }

// ConicInOklch is InOklch for a conic gradient.
func ConicInOklch(g ui.ConicGradient) ui.ConicGradient {
	g.Stops, g.Oklab = oklchStops(g.Stops), true
	return g
}
