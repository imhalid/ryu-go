package ryugo

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// Something dragged in from outside (DockOptions.Drop): over the window's right edge target it docks
// along that edge; over an area's middle target it goes into the area's stack; let go nowhere, nothing.
func TestDockDropFromOutside(t *testing.T) {
	l := NewLayout(DockArea(&Pane{ID: "a", Kind: "k", Title: "A"}))
	var drop *DockDrop
	kinds := []DockKind{{Kind: "k", Title: "Kind", Build: func(c *ui.Context, p *Pane) {}}}
	tt := ui.NewTester(func(c *ui.Context) {
		Root(c, Options{})
		Dock(c.Key("d"), l, DockOptions{Kinds: kinds, Drop: drop}).Fill()
	}, 1000, 700)
	tt.Frame()
	hold := func(id string, x, y float32) {
		drop = &DockDrop{Pane: &Pane{ID: id, Kind: "k", Title: id}, At: ui.Point{X: x, Y: y}}
		for range 3 {
			tt.Advance(50 * time.Millisecond)
			drop.At = ui.Point{X: x, Y: y}
		}
	}
	let := func() *DockDrop {
		d := drop
		d.Let = true
		tt.Frame()
		drop = nil
		tt.Advance(300 * time.Millisecond)
		return d
	}
	// the right edge target: the middle of the window's right edge
	r := edgeTargets(ui.Rect{W: 1000, H: 700})[1]
	hold("b", r.X+r.W/2, r.Y+r.H/2)
	if d := let(); !d.Placed || d.Where != "right" {
		t.Fatalf("not docked on the right: %+v", d)
	}
	if ids := l.PaneIDs(); len(ids) != 2 || ids[1] != "b" {
		t.Fatalf("panes: %v", ids)
	}
	// the middle target of the left area, which kept three quarters of the window: a tab of it
	tt.Frame()
	hold("c", (1000-250-dockGap)/2, headerH+(700-headerH)/2)
	if d := let(); !d.Placed || d.Where != "stack" {
		t.Fatalf("not a tab: %+v", d)
	}
	if a, _ := l.paneAt("c"); a == nil || len(a.Panes) != 2 || a.Panes[a.Active].ID != "c" {
		t.Fatalf("stack: %+v", a)
	}
	// nowhere: outside the window
	hold("d", -50, -50)
	if d := let(); d.Placed {
		t.Fatalf("placed from outside the window: %+v", d)
	}
}
