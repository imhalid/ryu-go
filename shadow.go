package ryugo

import "github.com/egoist/mygo/ui"

// ShadowLayer is one layer of a CSS box-shadow.
type ShadowLayer struct {
	X, Y, Blur, Spread float32
	Color              ui.Color
	// Inset draws the layer inside the box (box-shadow: inset).
	Inset bool
}

// Shadow is a stack of shadow layers, in CSS order: the first on top.
type Shadow []ShadowLayer

// Apply gives the shadow to e and returns e.
func (s Shadow) Apply(e ui.Element) ui.Element {
	for _, l := range s {
		if l.Color.A == 0 {
			continue
		}
		if l.Inset {
			e.InsetShadow(l.X, l.Y, l.Blur, l.Spread, l.Color)
		} else {
			e.Shadow(l.X, l.Y, l.Blur, l.Spread, l.Color)
		}
	}
	return e
}

// With appends o's layers after s's (listing them with commas in CSS).
func (s Shadow) With(o ...ShadowLayer) Shadow {
	return append(append(Shadow(nil), s...), o...)
}
