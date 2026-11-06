package ryugo

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// CornerScale sets the radii continuously and takes the pills with it below 1.
func TestCornerScale(t *testing.T) {
	base := tokensFor(false, Sm, CornersMd, ui.Color{})
	if base.Pill(28) != RadiusPill {
		t.Fatal("a pill is not a pill without CornerScale")
	}
	half := tokensScaled(false, Sm, CornersMd, ui.Color{}, 0.5)
	if got := half.R(12); got != 6 {
		t.Fatalf("R(12) at 0.5 = %v", got)
	}
	if got := half.Pill(28); got != 7 {
		t.Fatalf("Pill(28) at 0.5 = %v", got)
	}
	square := tokensScaled(false, Sm, CornersMd, ui.Color{}, 0.01)
	if square.Pill(28) > 0.2 {
		t.Fatalf("not square: %v", square.Pill(28))
	}
	for _, s := range []float32{1, 1.5, 2} {
		if tokensScaled(false, Sm, CornersMd, ui.Color{}, s).Pill(28) != RadiusPill {
			t.Fatalf("at %v the pills are not pills", s)
		}
	}
	if got := tokensScaled(false, Sm, CornersMd, ui.Color{}, 2).R(12); got != 24 {
		t.Fatalf("R(12) at 2 = %v", got)
	}
}
