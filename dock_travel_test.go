package ryugo

import (
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// twoWindows is a host showing two windows of a layout side by side in a
// tester's one: main on the left, 500 wide, and w1 on the right. Both
// share the tester's coordinates, so w1's content says it starts at the
// window's top left, and spans it all: whatever is right of main is over
// it. syncs counts the calls to Sync.
type twoWindows struct {
	syncs  int
	beside bool // w1 is there; else the pointer right of main is over no window
	// the pointer, which the test keeps; the moves of a window carried,
	// and whether the last one faded it
	p     ui.Point
	moves int
	faded bool
}

func (h *twoWindows) Pointer() (ui.Point, bool) { return h.p, true }

func (h *twoWindows) Move(id string, p ui.Point, faded bool) {
	h.moves++
	h.faded = faded
}

func (h *twoWindows) Content(id string) (ui.Rect, bool) {
	switch {
	case id == "main":
		return ui.Rect{W: 500, H: 700}, true
	case id == "w1" && h.beside:
		return ui.Rect{W: 1100, H: 700}, true
	}
	return ui.Rect{}, false
}

func (h *twoWindows) Sync() { h.syncs++ }

func twoLayout() *Layout {
	l := NewLayout(DockSplit(DockRow, nil,
		DockArea(&Pane{Kind: "k", Title: "files"}),
		DockArea(&Pane{Kind: "k", Title: "editor"}, &Pane{Kind: "k", Title: "notes"}),
	))
	l.Windows = append(l.Windows, &DockWindow{ID: "w1", Root: DockArea(&Pane{Kind: "k", Title: "console"})})
	l.fix()
	return l
}

func twoTester(l *Layout, h DockHost, show2 bool) *ui.Tester {
	kinds := []DockKind{{Kind: "k", Title: "Kind", Build: func(c *ui.Context, p *Pane) {}}}
	tt := ui.NewTester(func(c *ui.Context) {
		Root(c, Options{})
		ui.Row(c).Fill().Children(func() {
			ui.Box(c).Width(500).Height(700).Children(func() {
				Dock(c.Key("main"), l, DockOptions{Kinds: kinds, Host: h}).Fill()
			})
			if show2 {
				ui.Box(c).Width(600).Height(700).Children(func() {
					Dock(c.Key("w1"), l, DockOptions{Window: "w1", Kinds: kinds, Host: h}).Fill()
				})
			}
		})
	}, 1100, 700)
	tt.Advance(300 * time.Millisecond)
	return tt
}

// pull drags from (x, y) to (tx, ty) a few points at a time, and lets go.
func pull(tt *ui.Tester, x, y, tx, ty float32) {
	tt.Press(x, y)
	for i := 1; i <= 12; i++ {
		f := float32(i) / 12
		tt.Move(x+(tx-x)*f, y+(ty-y)*f)
		tt.Advance(16 * time.Millisecond)
	}
	tt.Advance(100 * time.Millisecond)
	tt.Release(tx, ty)
	tt.Advance(800 * time.Millisecond)
}

func titlesOf(n *Node) []string {
	var out []string
	for _, p := range n.Panes {
		out = append(out, p.Title)
	}
	return out
}

// A tab dragged out of its window over another window of the layout docks
// there, in that window's stack.
func TestDockBetweenWindows(t *testing.T) {
	l := twoLayout()
	h := &twoWindows{beside: true}
	tt := twoTester(l, h, true)
	r, _ := tt.Find("Close notes")
	// onto the middle of the compass, in the middle of w1's area's body
	pull(tt, r.X-12, r.Y+r.H/2, 800, 28+(700-28)/2)
	w1 := l.Window("w1").Root
	if !slices.Equal(titlesOf(w1), []string{"console", "notes"}) {
		t.Fatalf("w1 holds %v", titlesOf(w1))
	}
	if ed := l.Window("main").Root.Children[1]; !slices.Equal(titlesOf(ed), []string{"editor"}) {
		t.Fatalf("main's editor area holds %v", titlesOf(ed))
	}
	if l.travel != nil || h.syncs == 0 {
		t.Fatalf("travel %v, syncs %d", l.travel, h.syncs)
	}
}

// A tab let go over no window tears off into a window of its own; Undo
// brings it back, and Redo tears it off again.
func TestDockTearOff(t *testing.T) {
	l := twoLayout()
	h := &twoWindows{}
	tt := twoTester(l, h, false)
	r, _ := tt.Find("Close notes")
	pull(tt, r.X-12, r.Y+r.H/2, 800, 330)
	if len(l.Windows) != 3 {
		t.Fatalf("%d windows, want a third torn off", len(l.Windows))
	}
	w := l.Windows[2]
	if !slices.Equal(titlesOf(w.Root), []string{"notes"}) || w.X == 0 || w.Width < 320 {
		t.Fatalf("the window torn off: %+v holding %v", w, titlesOf(w.Root))
	}
	if !l.Undo() || len(l.Windows) != 2 || len(l.Window("main").Root.Children[1].Panes) != 2 {
		t.Fatal("undo did not bring the tab back")
	}
	if !l.Redo() || len(l.Windows) != 3 {
		t.Fatal("redo did not tear it off again")
	}
}

// A window closed gives what it holds back to the main window, along its
// right edge; a window left without panes goes.
func TestLayoutDockBack(t *testing.T) {
	l := twoLayout()
	if !l.DockBack("w1") || l.Window("w1") != nil {
		t.Fatal("w1 is still there")
	}
	root := l.Window("main").Root
	if len(root.Children) != 3 || !slices.Equal(titlesOf(root.Children[2]), []string{"console"}) {
		t.Fatalf("console is not along main's right edge: %d children", len(root.Children))
	}
	if !l.opening[root.Children[2].ID] {
		t.Fatal("the area back is not marked to open")
	}
	l.Windows = append(l.Windows, &DockWindow{ID: "w2", Root: DockArea()})
	if !l.pruneWindows() || l.Window("w2") != nil {
		t.Fatal("an empty window stayed")
	}
}

// A tab's menu splits the area with it.
func TestDockTabMenuSplit(t *testing.T) {
	l := twoLayout()
	tt := twoTester(l, nil, false)
	x, y := float32(0), float32(0)
	if r, ok := tt.Find("Close notes"); ok {
		x, y = r.X-12, r.Y+r.H/2
	}
	tt.Press(x, y)
	tt.Release(x, y)
	tt.Advance(100 * time.Millisecond)
	tt.RightClickAt(x, y)
	tt.Advance(300 * time.Millisecond)
	if err := tt.Click("Split down"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(600 * time.Millisecond)
	right := l.Window("main").Root.Children[1]
	if right.IsArea() || right.Dir != DockColumn || !slices.Equal(titlesOf(right.Children[1]), []string{"notes"}) {
		t.Fatalf("notes did not go under its stack: %+v", right)
	}
}

// A torn-off window's only tab carries its window; over the main window,
// faded, it docks where the main window shows, and the window, left
// empty, goes.
func TestDockCarryWindow(t *testing.T) {
	l := twoLayout()
	h := &twoWindows{beside: true}
	tt := twoTester(l, h, true)
	r, _ := tt.Find("Close console")
	x, y := r.X-12, r.Y+r.H/2
	// onto the middle of the compass over main's right area (250 to 500)
	tx, ty := float32(375), float32(28+(700-28)/2)
	h.p = ui.Point{X: x, Y: y}
	tt.Press(x, y)
	for i := 1; i <= 12; i++ {
		f := float32(i) / 12
		h.p = ui.Point{X: x + (tx-x)*f, Y: y + (ty-y)*f}
		// the window moves with the pointer: in it, the pointer hardly moves
		tt.Move(x+8*f, y)
		tt.Advance(16 * time.Millisecond)
	}
	tt.Advance(100 * time.Millisecond)
	if h.moves == 0 || !h.faded {
		t.Fatalf("the window was not carried, faded over main: %d moves, faded %v", h.moves, h.faded)
	}
	tt.Release(x+8, y)
	tt.Advance(800 * time.Millisecond)
	if l.Window("w1") != nil {
		t.Fatal("w1, left empty, is still there")
	}
	if ed := l.Window("main").Root.Children[1]; !slices.Contains(titlesOf(ed), "console") {
		t.Fatalf("console did not dock into main's right area: %v", titlesOf(ed))
	}
}

// A tab's menu moves it into another window torn off: two windows become
// one, the one left empty going.
func TestDockMenuMergeWindows(t *testing.T) {
	l := twoLayout()
	l.Windows = append(l.Windows, &DockWindow{ID: "w2", Root: DockArea(&Pane{Kind: "k", Title: "props"})})
	l.fix()
	h := &twoWindows{beside: true}
	tt := twoTester(l, h, true)
	r, _ := tt.Find("Close console")
	tt.RightClickAt(r.X-12, r.Y+r.H/2)
	tt.Advance(300 * time.Millisecond)
	if err := tt.Click("Move to props window"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(300 * time.Millisecond)
	if l.Window("w1") != nil {
		t.Fatal("w1, left empty, is still there")
	}
	if got := titlesOf(l.Window("w2").Root); !slices.Equal(got, []string{"props", "console"}) {
		t.Fatalf("w2 holds %v", got)
	}
}

// Moving a window's only tab back to the main window from its menu closes
// the window, as its menu still builds.
func TestDockMenuToMain(t *testing.T) {
	l := twoLayout()
	l.Windows = append(l.Windows, &DockWindow{ID: "w2", Root: DockArea(&Pane{Kind: "k", Title: "props"})})
	l.fix()
	tt := twoTester(l, &twoWindows{beside: true}, true)
	r, _ := tt.Find("Close console")
	tt.RightClickAt(r.X-12, r.Y+r.H/2)
	tt.Advance(300 * time.Millisecond)
	if err := tt.Click("Move to main window"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(300 * time.Millisecond)
	if l.Window("w1") != nil || !slices.Contains(titlesOf(l.Window("main").Root.Children[2]), "console") {
		t.Fatal("console did not go back to the main window")
	}
}
