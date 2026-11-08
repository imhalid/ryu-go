package ryugo

import (
	"slices"

	"github.com/egoist/mygo/ui"
)

// Something dragged into the Dock from outside it (DockOptions.Drop): a row of the app's list, a
// page, held over the Dock as a tab is. While it is held it is a tab drag of its own kind (outer),
// aimed as one is: the compass shows over the area under the pointer, the window's edge targets
// along its edges, the region it would take behind them, and the tabs of a strip it is over make
// room. Let go over a place, its pane goes in there at once; the app draws what it carries.

// dropW is the room a pane dragged in takes among the tabs of a strip.
const dropW = 140

// dropIn follows o.Drop: it starts, aims, places or forgets the drag in from outside.
func (d *dockView) dropIn() {
	st, o := d.st, d.o.Drop
	dr := st.tab
	if o == nil || o.Pane == nil || o.Pane.ID == "" {
		if dr != nil && dr.outer { // the app let it go elsewhere, or dropped it
			st.tab = nil
			d.c.AnimationFrame()
		}
		return
	}
	if dr == nil || !dr.outer {
		if dr != nil || st.tabApply != nil || st.drag != nil || st.corner != nil || st.fold != nil || d.maxed != "" {
			return // the Dock is busy with a drag of its own
		}
		h := float32(headerH - 8)
		dr = &tabDrag{pane: o.Pane.ID, title: o.Pane.Title, w: dropW, h: h, grab: ui.Point{X: dropW / 2, Y: h / 2}, outer: true}
		st.tab = dr
	}
	dr.pos, dr.title = o.At, o.Pane.Title
	before := dr.zone
	if inside(d.root.LayoutBounds(), o.At) {
		d.aimTab()
	} else {
		dr.zone = dockZone{}
	}
	if dr.zone != before {
		d.c.AnimationFrame()
	}
	if !o.Let {
		return
	}
	z := dr.zone
	st.tab = nil
	if z.side == zoneNone {
		return
	}
	if d.place(z, o.Pane) {
		o.Placed = true
		o.Where = [...]string{zoneStack: "stack", zoneLeft: "left", zoneRight: "right", zoneTop: "top", zoneBottom: "bottom"}[z.side]
		st.tabGen++
		d.changed()
	}
	d.c.AnimationFrame()
}

// place puts pane in zone z: in an area's stack, or in a new area beside it or along the window's
// edge, as a tab let go there goes. A pane the layout has moves there.
func (d *dockView) place(z dockZone, pane *Pane) bool {
	l := d.l
	if l.FindPane(pane.ID) != nil {
		op := d.dropOp(z, pane.ID)
		if op == nil {
			return false
		}
		op()
		return true
	}
	to := l.Find(z.area)
	if to == nil {
		return false
	}
	if z.side == zoneStack {
		if !to.IsArea() {
			return false
		}
		l.remember()
		i := min(max(z.index, 0), len(to.Panes))
		to.Panes = slices.Insert(to.Panes, i, pane)
		to.Active = i
		return true
	}
	r := d.st.areas[z.area]
	if z.edge {
		r = d.root.LayoutBounds()
	}
	dir, first, length := DockRow, z.side == zoneLeft, r.W
	if z.side == zoneTop || z.side == zoneBottom {
		dir, first, length = DockColumn, z.side == zoneTop, r.H
	}
	keep, made := (length-dockGap)/2, (length-dockGap)/2
	if z.edge {
		made = d.edgeLength(z.side, r)
		keep = length - made - dockGap
	}
	var room float32
	if _, p, _ := l.where(z.area); p != nil {
		room = d.st.room[p.ID]
	}
	l.remember()
	n := l.beside(to, pane, dir, first, keep, made, room)
	if n != nil {
		l.mark(n.ID, true)
	}
	return n != nil
}
