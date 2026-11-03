package mizan

import "github.com/egoist/mygo/ui"

// ShadowLayer, CSS box-shadow'un bir katmanıdır.
type ShadowLayer struct {
	X, Y, Blur, Spread float32
	Color              ui.Color
	// Inset, katmanı kutunun içine çizer (box-shadow: inset).
	Inset bool
}

// Shadow, üst üste binen gölge katmanlarıdır, CSS'teki sırayla: ilki en üstte.
type Shadow []ShadowLayer

// Apply, gölgeyi e'ye verir ve e'yi döndürür.
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

// With, s'nin ardına o'nun katmanlarını ekler (CSS'te virgülle sıralamak).
func (s Shadow) With(o ...ShadowLayer) Shadow {
	return append(append(Shadow(nil), s...), o...)
}
