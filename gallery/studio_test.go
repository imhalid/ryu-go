package gallery

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/ryu-go"
)

// studioTester opens the studio page, its palettes placed.
func studioTester() (*State, *ui.Tester) {
	s := New()
	s.Mode, s.Page = mz.Light, 7
	tt := ui.NewTester(s.View, 1180, 1000)
	tt.Advance(400 * time.Millisecond)
	tt.Frame()
	tt.Advance(400 * time.Millisecond)
	return s, tt
}

// titleOf is a point on palette title's bar, right of its close button.
func titleOf(t *testing.T, tt *ui.Tester, title string, in float32) (float32, float32) {
	t.Helper()
	r, ok := tt.Find("Close " + title)
	if !ok {
		t.Fatalf("no palette %q", title)
	}
	return r.X + in, r.Y + r.H/2
}

// A palette moves with its title bar, from where it is held.
func TestPaletteDrag(t *testing.T) {
	s, tt := studioTester()
	p := s.Studio.Board.Palettes["layers"]
	x0, y0 := p.X, p.Y
	x, y := titleOf(t, tt, "Layers", 60)
	dragEdge(tt, x, y, -200, 120)
	if d := p.X - (x0 - 200); d > 0.5 || d < -0.5 || p.Y-(y0+120) > 0.5 || p.Y-(y0+120) < -0.5 {
		t.Fatalf("layers at %v, %v; want %v, %v", p.X, p.Y, x0-200, y0+120)
	}
}

// Near the board's edge, a palette settles onto it, 8 DIPs in.
func TestPaletteSnapEdge(t *testing.T) {
	s, tt := studioTester()
	p := s.Studio.Board.Palettes["layers"]
	// placed 12 from the right edge: 4 more to the right settles at 8
	want := p.X + 4
	x, y := titleOf(t, tt, "Layers", 60)
	dragEdge(tt, x, y, 6, 40)
	if p.X != want {
		t.Fatalf("layers at x %v, want it settled at %v", p.X, want)
	}
}

// Near another palette, a palette settles beside it, 6 DIPs apart.
func TestPaletteSnapNeighbor(t *testing.T) {
	s, tt := studioTester()
	tools, p := s.Studio.Board.Palettes["tools"], s.Studio.Board.Palettes["layers"]
	// the tools palette is 74 wide: beside it, layers' left is at its right
	// and the gap
	want := tools.X + 74 + 6
	x, y := titleOf(t, tt, "Layers", 60)
	dragEdge(tt, x, y, want+4-p.X, 0)
	if p.X != want {
		t.Fatalf("layers at x %v, want it beside the tools at %v", p.X, want)
	}
}

// A double click on its title folds a palette up, and another opens it.
func TestPaletteCollapse(t *testing.T) {
	s, tt := studioTester()
	x, y := titleOf(t, tt, "Layers", 60)
	tt.ClickAt(x, y)
	tt.ClickAt(x, y)
	tt.Advance(500 * time.Millisecond)
	if !s.Studio.Board.Palettes["layers"].Collapsed {
		t.Fatal("not folded up")
	}
	// far enough from the last click not to make a third
	tt.Advance(time.Second)
	tt.ClickAt(x, y)
	tt.ClickAt(x, y)
	tt.Advance(500 * time.Millisecond)
	if s.Studio.Board.Palettes["layers"].Collapsed {
		t.Fatal("not opened again")
	}
}

// Its close button closes a palette; opened, it comes back in front.
func TestPaletteClose(t *testing.T) {
	s, tt := studioTester()
	if err := tt.Click("Close Layers"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(400 * time.Millisecond)
	if tt.HasText("Background") {
		t.Fatal("the layers palette still shows")
	}
	s.Studio.Board.Open("layers")
	tt.Frame()
	tt.Advance(400 * time.Millisecond)
	if !tt.HasText("Background") {
		t.Fatal("the layers palette did not open again")
	}
	b := s.Studio.Board.Palettes
	if b["layers"].Z <= b["tools"].Z || b["layers"].Z <= b["swatches"].Z {
		t.Fatalf("opened, layers is not in front: %v", b["layers"].Z)
	}
}

// The brightness and contrast HUD opens from the toolbar; Cancel takes the
// values back and closes it.
func TestPaletteHUD(t *testing.T) {
	s, tt := studioTester()
	if err := tt.Click("Brightness and contrast…"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(400 * time.Millisecond)
	if !s.Studio.Board.IsOpen("adjust") || !tt.HasText("Contrast") {
		t.Fatal("the HUD did not open")
	}
	s.Studio.Bright = 40
	if err := tt.Click("Cancel"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(400 * time.Millisecond)
	if s.Studio.Board.IsOpen("adjust") || s.Studio.Bright != 0 {
		t.Fatalf("cancel: open %v, brightness %v", s.Studio.Board.IsOpen("adjust"), s.Studio.Bright)
	}
}
