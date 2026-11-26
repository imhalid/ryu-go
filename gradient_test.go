package ryugo

import (
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
)

// A blend in OKLCH keeps a color's chroma where the blend in Oklab grays it: halfway from red to cyan it is a vivid
// hue between theirs, not gray; a darker step of a blue keeps its hue; a fade to clear keeps the color; and the hue
// goes the shorter way round.
func TestMixOklch(t *testing.T) {
	red, cyan := ui.Oklch(0.63, 0.25, 29), ui.Oklch(0.9, 0.15, 195)
	_, cOklab, _ := toOklch(Mix(red, cyan, 0.5))
	_, cOklch, _ := toOklch(MixOklch(red, cyan, 0.5))
	if cOklch < 0.12 || cOklch < cOklab*1.5 {
		t.Fatalf("red to cyan halfway: chroma %.3f in OKLCH, %.3f in Oklab", cOklch, cOklab)
	}
	blue, deep := ui.Oklch(0.62, 0.2, 256), ui.Oklch(0.48, 0.16, 256)
	if _, _, h := toOklch(MixOklch(blue, deep, 0.5)); math.Abs(h-256) > 2 {
		t.Fatalf("a blue's steps: hue %.1f", h)
	}
	if l, c, h := toOklch(MixOklch(blue, ui.Color{}, 0.9)); math.Abs(h-256) > 2 || c < 0.15 || math.Abs(l-0.62) > 0.03 {
		t.Fatalf("a fade to clear is not the blue's: l %.2f c %.3f h %.1f", l, c, h)
	}
	a, b := ui.Oklch(0.7, 0.15, 350), ui.Oklch(0.7, 0.15, 10)
	if _, _, h := toOklch(MixOklch(a, b, 0.5)); h > 5 && h < 355 {
		t.Fatalf("the hue went the long way round: %.1f", h)
	}
}

// InOklch keeps a gradient's ends where they were, its From and To, Start and End among its stops, within MyGo's 32.
func TestInOklch(t *testing.T) {
	g := InOklch(ui.LinearGradient{Angle: 90, From: ui.Oklch(0.6, 0.2, 256), To: ui.Oklch(0.5, 0.2, 262), Start: 0.2, End: 0.8})
	if n := len(g.Stops); n < 3 || n > 32 {
		t.Fatalf("%d stops", n)
	}
	if first, last := g.Stops[0], g.Stops[len(g.Stops)-1]; first.At != 0.2 || last.At != 0.8 || first.Color != ui.Oklch(0.6, 0.2, 256) {
		t.Fatalf("ends %+v %+v", first, last)
	}
	many := make([]ui.GradientStop, 12)
	for i := range many {
		many[i] = ui.GradientStop{Color: ui.Oklch(0.6, 0.15, float32(i*30)), At: float32(i) / 11}
	}
	if n := len(InOklch(ui.LinearGradient{Stops: many}).Stops); n > 32 {
		t.Fatalf("%d stops, MyGo draws 32", n)
	}
}
