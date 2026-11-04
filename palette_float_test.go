package mizan

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// fakePalettes shows palettes' windows as testers of their own: the board's
// window's content is at (100, 100) on the screen, 1000 by 700, on a work
// area of 1920 by 1080.
type fakePalettes struct {
	p       ui.Point // the pointer, which the test keeps
	shown   map[string]ui.Rect
	views   map[string]func(*ui.Context)
	hidden  []string
	redrawn map[string]int // by what was left out
}

func newFakePalettes() *fakePalettes {
	return &fakePalettes{shown: map[string]ui.Rect{}, views: map[string]func(*ui.Context){}, redrawn: map[string]int{}}
}

func (f *fakePalettes) Free() bool        { return true }
func (f *fakePalettes) Board() ui.Rect    { return ui.Rect{X: 100, Y: 100, W: 1000, H: 700} }
func (f *fakePalettes) Work() ui.Rect     { return ui.Rect{W: 1920, H: 1080} }
func (f *fakePalettes) Pointer() ui.Point { return f.p }
func (f *fakePalettes) Show(id string, r ui.Rect, view func(*ui.Context)) {
	f.shown[id] = r
	if view != nil && f.views[id] == nil {
		f.views[id] = view
	}
}
func (f *fakePalettes) Hide(id string)       { f.hidden = append(f.hidden, id); delete(f.shown, id) }
func (f *fakePalettes) Redraw(except string) { f.redrawn[except]++ }

// floatTesters builds a board of two palettes in windows of their own,
// and a tester for the window of a.
func floatTesters(t *testing.T) (*Board, *fakePalettes, *ui.Tester, *ui.Tester) {
	t.Helper()
	f := newFakePalettes()
	b := &Board{Host: f}
	main := ui.NewTester(func(c *ui.Context) {
		Root(c, Options{})
		Palettes(c, b, func(p *PaletteSet) {
			ui.Box(c).Fill()
			p.Palette("a", PaletteOptions{Title: "A", Width: 200, At: ui.Point{X: 12, Y: 12}}, func(c *ui.Context) {
				ui.Text(c, "in a")
			})
			p.Palette("b", PaletteOptions{Title: "B", Width: 200, At: ui.Point{X: 400, Y: 12}}, func(c *ui.Context) {
				ui.Text(c, "in b")
			})
		}).Fill()
	}, 1000, 700)
	main.Advance(300 * time.Millisecond)
	main.Frame()
	if f.views["a"] == nil {
		t.Fatal("no window for a")
	}
	r := f.shown["a"]
	win := ui.NewTester(f.views["a"], int(r.W), 400)
	win.Advance(400 * time.Millisecond)
	return b, f, main, win
}

// Palettes open in windows of their own where they would float over the
// board, the window as tall as the palette.
func TestPaletteWindows(t *testing.T) {
	b, f, _, win := floatTesters(t)
	s := b.state("a")
	if s.SX != 112 || s.SY != 112 {
		t.Fatalf("a at %v, %v on the screen, want 112, 112", s.SX, s.SY)
	}
	if !win.HasText("in a") || !win.HasText("A") {
		t.Fatalf("a's window shows %v", win.Texts())
	}
	if r := f.shown["a"]; r.X != 112-floatPad || r.H <= 2*floatPad+paletteTitleH {
		t.Fatalf("a's window is %+v", r)
	}
}

// Its title bar carries a palette's window with the pointer, settling onto
// the board's window's edge inside it.
func TestPaletteWindowCarry(t *testing.T) {
	b, f, _, win := floatTesters(t)
	s := b.state("a")
	// the title bar, in the window: under the margin
	x, y := float32(floatPad+100), float32(floatPad+paletteTitleH/2)
	f.p = ui.Point{X: s.SX + 100, Y: s.SY + paletteTitleH/2}
	win.Press(x, y)
	win.Advance(16 * time.Millisecond)
	// 60 right, then back to 2 from the window's edge inside
	for _, dx := range []float32{20, 40, 60} {
		f.p.X += 20
		win.Move(x+dx, y)
		win.Advance(16 * time.Millisecond)
	}
	if s.SX != 172 || f.shown["a"].X != 172-floatPad {
		t.Fatalf("a at %v, its window at %v; want 172", s.SX, f.shown["a"].X)
	}
	f.p.X -= 62
	win.Move(x-2, y)
	win.Advance(16 * time.Millisecond)
	win.Release(x-2, y)
	win.Advance(100 * time.Millisecond)
	if s.SX != 108 {
		t.Fatalf("a at %v, want it settled 8 inside the window's edge, at 108", s.SX)
	}
	if f.redrawn["a"] == 0 {
		t.Fatal("the pointer over a's window redrew no other window")
	}
}

// Its close button shrinks a palette away, and its window goes.
func TestPaletteWindowClose(t *testing.T) {
	b, f, _, win := floatTesters(t)
	if err := win.Click("Close A"); err != nil {
		t.Fatal(err)
	}
	win.Advance(400 * time.Millisecond)
	if !b.state("a").Closed || len(f.hidden) != 1 || f.hidden[0] != "a" {
		t.Fatalf("closed %v, hidden %v", b.state("a").Closed, f.hidden)
	}
}
