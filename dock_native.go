package mizan

import (
	"github.com/egoist/mygo/transfer"
	"github.com/egoist/mygo/ui"
)

// Tabs dragged by the system, where it doesn't say where windows are
// (Wayland), so that the Dock can't follow a tab out of its window: the
// system's drag and drop carries a tab from window to window instead. Its
// copy is the system's; the Dock under the pointer shows the tab as its
// own (its stack making room, the region, the compass) and takes it where
// it is let go; let go over no window, it tears off into a window of its
// own, which the system places.

// dockTabFormat is what a tab dragged by the system is to other apps:
// nothing they take.
const dockTabFormat transfer.Format = "application/x-mizan-dock-tab"

// dockTab is a tab dragged by the system: what the Docks of its layout take.
type dockTab struct {
	l           *Layout
	pane, title string
	w, h        float32
}

// nativeSource makes b, the title of tab e (pane p of area n), a source
// of the system's drag and drop.
func (d *dockView) nativeSource(n *Node, p *Pane, e, b ui.Element) {
	l, h, onChange := d.l, d.o.Host, d.o.OnChange
	eb := e.LayoutBounds()
	pane := p.ID
	a := d.st.areas[n.ID]
	size := ui.Rect{W: min(max(a.W, 320), 900), H: min(max(a.H, 240), 700)}
	b.Drag(dockTab{l: l, pane: pane, title: p.Title, w: eb.W, h: eb.H}).
		DragDataFrom(func() transfer.Data {
			return transfer.New(transfer.NewItem(transfer.Bytes(dockTabFormat, []byte(pane))))
		}, transfer.DragOptions{Done: func(r transfer.Result) {
			// let go over no window (not given up): it tears off
			if !r.Canceled || r.Escaped {
				return
			}
			if a, _ := l.paneAt(pane); a == nil {
				return
			}
			l.remember()
			l.tearOff(pane, size)
			l.pruneWindows()
			if onChange != nil {
				onChange()
			}
			h.Sync()
		}})
	if b.Dragging() {
		// its place stays, empty, while the system shows its copy
		e.Opacity(0)
	}
}

// nativeTarget takes the tabs the system drags over the Dock: over it, the
// Dock shows one as its own tab dragged; let go, it goes where the Dock
// showed it would.
func (d *dockView) nativeTarget() {
	st, l, c := d.st, d.l, d.c
	if v, ok := ui.Drop[dockTab](d.root); ok && v.l == l {
		z := st.nativeZone
		st.tab = nil
		if a, i := l.paneAt(v.pane); z.side == zoneStack && a != nil && a.ID == z.area && i == z.index {
			return // back where it was
		}
		if op := d.dropOp(z, v.pane); op != nil {
			st.tabApply = func() {
				op()
				l.pruneWindows()
			}
			st.applyAt = c.Now()
			st.tabLand = &tabLanding{pane: v.pane, at: st.nativeAt, after: c.Now()}
			c.AnimationFrame()
		}
		return
	}
	v, ok := ui.DragOver[dockTab](d.root)
	if !ok || v.l != l {
		if st.tab != nil && st.tab.native {
			st.tab = nil
		}
		return
	}
	dr := st.tab
	if dr == nil || !dr.native || dr.pane != v.pane {
		// from an area of this window, its own rules hold (alone there, it
		// goes nowhere over it)
		from := ""
		if a, _ := l.paneAt(v.pane); a != nil {
			walk(d.w.Root, func(n *Node) {
				if n == a {
					from = a.ID
				}
			})
		}
		dr = &tabDrag{pane: v.pane, from: from, title: v.title, w: v.w, h: v.h,
			grab: ui.Point{X: v.w / 2, Y: v.h / 2}, remote: true, native: true}
		st.tab = dr
	}
	b := d.root.LayoutBounds()
	px, py, _ := d.root.PointerPosition()
	dr.pos = ui.Point{X: b.X + px, Y: b.Y + py}
	c.AnimationFrame()
}
