package mizan

import (
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
)

// A Dock's windows (DOCK.md, step 5). The layout is the model of every
// window the app shows it in, each with a Dock of its own (DockOptions.
// Window); a DockHost opens and closes the app's windows as the layout's
// come and go. A tab dragged out of its window, let go over no window,
// tears off into a window of its own there; over another window of the
// layout, that window's Dock shows where it would dock (its region, its
// compass) as if the tab were its own, and takes it. Where the system
// doesn't say where windows are (Wayland), a tab dragged out only tears
// off, and a tab's menu moves it back to the main window. A window closed
// gives what it holds back to the main window (Layout.DockBack).

// follow notes, for the tab dragged here, whether it left the window,
// and, when it did, which window of the layout it is over, for that
// window's Dock to show it.
func (d *dockView) follow() {
	st, l, h := d.st, d.l, d.o.Host
	dr := st.tab
	if h == nil {
		return
	}
	me, known := h.Content(d.w.ID)
	dr.out = me.W > 0 && !inside(ui.Rect{W: me.W, H: me.H}, dr.pos)
	if !dr.out {
		if t := l.travel; t != nil && t.from == d.w.ID {
			l.travel = nil
			h.Sync()
		}
		return
	}
	dr.zone, dr.known = dockZone{}, known
	t := l.travel
	if t == nil || t.from != d.w.ID || t.pane != dr.pane {
		t = &dockTravel{pane: dr.pane, from: d.w.ID, title: dr.title, w: dr.w, h: dr.h, grab: dr.grab}
		l.travel = t
	}
	over, at := "", ui.Point{}
	if known {
		dr.screen = ui.Point{X: me.X + dr.pos.X, Y: me.Y + dr.pos.Y}
		for _, w := range l.Windows {
			if w.ID == d.w.ID {
				continue
			}
			if r, ok := h.Content(w.ID); ok && inside(r, dr.screen) {
				over, at = w.ID, ui.Point{X: dr.screen.X - r.X, Y: dr.screen.Y - r.Y}
				break
			}
		}
	}
	if over != t.over {
		t.op = nil
	}
	t.over, t.at = over, at
	// the window it is over shows it in its next frame
	h.Sync()
}

// letOut lets go of the tab dragged out of the window: over a window of
// the layout, it docks where that window shows; over none, it tears off
// into a window of its own there, the size of the area it leaves.
func (d *dockView) letOut() {
	st, l, h := d.st, d.l, d.o.Host
	dr := st.tab
	t := l.travel
	l.travel = nil
	pane := dr.pane
	switch {
	case t != nil && t.over != "" && t.op != nil:
		op := t.op
		st.tabApply = func() {
			op()
			l.pruneWindows()
		}
	case t != nil && t.over != "":
		// over a window that takes nothing here: it flies back
		st.tabLand = &tabLanding{pane: pane, at: ui.Point{X: dr.pos.X - dr.grab.X, Y: dr.pos.Y - dr.grab.Y}, tilt: dr.tilt, after: d.c.Now()}
		st.tab = nil
		h.Sync()
		return
	default:
		a := d.st.areas[dr.from]
		r := ui.Rect{W: min(max(a.W, 320), 900), H: min(max(a.H, 240), 700)}
		if dr.known {
			// its content where the copy was, a little up and left
			r.X, r.Y = max(dr.screen.X-dr.grab.X-12, 1), max(dr.screen.Y-dr.grab.Y-8, 1)
		}
		st.tabApply = func() {
			l.remember()
			l.tearOff(pane, r)
			l.pruneWindows()
		}
	}
	d.c.AnimationFrame()
	h.Sync()
}

// visit shows a tab dragged from another window over this one as if it
// were dragged here: its copy, the stack making room, the region and the
// compass.
func (d *dockView) visit() {
	st, l := d.st, d.l
	t := l.travel
	if t == nil || t.over != d.w.ID || t.from == d.w.ID {
		if st.tab != nil && st.tab.remote && !st.tab.native {
			st.tab = nil
		}
		return
	}
	dr := st.tab
	if dr == nil || !dr.remote || dr.pane != t.pane {
		dr = &tabDrag{pane: t.pane, title: t.title, w: t.w, h: t.h, grab: t.grab, remote: true}
		st.tab = dr
	}
	dr.pos = t.at
	d.c.AnimationFrame()
}

// offer says, for the tab dragged from another window, what letting it go
// where it is here does: the move, and its landing here.
func (d *dockView) offer() {
	st, l := d.st, d.l
	dr, t := st.tab, l.travel
	if t == nil || dr == nil || !dr.remote {
		return
	}
	op := d.dropOp(dr.zone, dr.pane)
	if op == nil {
		t.op = nil
		return
	}
	land := tabLanding{pane: dr.pane, at: ui.Point{X: dr.pos.X - dr.grab.X, Y: dr.pos.Y - dr.grab.Y}}
	t.op = func() {
		op()
		st.landed = &land
	}
}

// arrive lands, as the frame starts, a tab another window let go here.
func (d *dockView) arrive() {
	st := d.st
	if st.landed == nil {
		return
	}
	land := *st.landed
	land.after = d.c.Now()
	st.tabLand, st.landed, st.tab = &land, nil, nil
	st.tabGen++
	d.c.AnimationFrame()
}

// windows keeps the Dock in step with its layout: a layout put in place
// of another starts afresh, every node where it is; windows left without
// panes go; and when the layout's windows change, the host opens and
// closes the app's.
func (d *dockView) windows() {
	st, l := d.st, d.l
	// another theme (the app's appearance, accent, size or corners): the
	// layout's other windows take it at once, not when the pointer next
	// goes over them
	if st.tokens != d.t {
		if st.tokens != nil && d.o.Host != nil {
			d.o.Host.Sync()
		}
		st.tokens = d.t
	}
	if st.layout != l {
		*st = dockState{closing: map[string][]tabGhost{}, settled: map[string]bool{}, windows: st.windows, tokens: st.tokens}
		st.layout = l
		for _, w := range l.Windows {
			walk(w.Root, func(n *Node) { st.settled[n.ID] = true })
		}
	}
	if l.pruneWindows() {
		d.changed()
	}
	var ids []string
	for _, w := range l.Windows {
		ids = append(ids, w.ID)
	}
	if s := strings.Join(ids, " "); s != st.windows {
		st.windows = s
		if h := d.o.Host; h != nil {
			h.Sync()
		}
	}
}

// tabMenu is a tab's menu, on a right click: split the area with it, send
// it to a window of its own or back to the main window, close it or the
// others.
func (d *dockView) tabMenu(n *Node, p *Pane, anchor ui.Element, open *bool) {
	c, l := d.c, d.l
	host := d.o.Host
	Menu(c, anchor, open, func(m *MenuList) {
		many := len(n.Panes) > 1
		split := func(dir DockDir) {
			a := d.st.areas[n.ID]
			length := a.W
			if dir == DockColumn {
				length = a.H
			}
			half := (length - dockGap) / 2
			var room float32
			if _, par, _ := l.where(n.ID); par != nil {
				room = d.st.room[par.ID]
			}
			var made *Node
			if many {
				// the tab goes beside its stack
				l.remember()
				made = l.dockPane(p.ID, n.ID, dir, false, half, half, room)
			} else {
				// alone, it shows twice
				made = l.splitArea(n.ID, dir, false, half, half, room)
			}
			if made != nil {
				l.mark(made.ID, true)
			}
			d.changed()
		}
		if m.Item("Split right", ItemOptions{}) {
			split(DockRow)
		}
		if m.Item("Split down", ItemOptions{}) {
			split(DockColumn)
		}
		if host != nil {
			m.Separator()
			if m.Item("Open in new window", ItemOptions{}) {
				l.remember()
				l.tearOff(p.ID, d.windowRect(n))
				l.pruneWindows()
				d.changed()
			}
			if many && m.Item("Move area to new window", ItemOptions{}) {
				l.remember()
				l.tearOffArea(n.ID, d.windowRect(n))
				d.changed()
			}
			if d.w.ID != "main" && m.Item("Move to main window", ItemOptions{}) {
				l.remember()
				l.moveToMain(p.ID)
				l.pruneWindows()
				d.changed()
			}
			// into another window torn off, as a tab of its top right area:
			// two windows become one, where no drag can do it (Wayland)
			// a copy: an item chosen above may have closed a window, which
			// clears its slot
			for _, w := range slices.Clone(l.Windows) {
				if w == nil || w.ID == "main" || w.ID == d.w.ID {
					continue
				}
				if to := topRight(w.Root); to != nil && m.Item("Move to "+windowTitle(w), ItemOptions{}) {
					l.remember()
					l.movePane(p.ID, to.ID, len(to.Panes))
					l.pruneWindows()
					d.changed()
				}
			}
		}
		m.Separator()
		if m.Item("Close", ItemOptions{}) && l.closePane(p.ID) {
			d.changed()
		}
		if many && m.Item("Close others", ItemOptions{}) {
			l.remember()
			n.Panes, n.Active = []*Pane{p}, 0
			d.changed()
		}
	})
}

// windowRect is where a window torn off from area n opens: over the area,
// its size, when the system says where windows are; else where it puts it.
func (d *dockView) windowRect(n *Node) ui.Rect {
	a := d.st.areas[n.ID]
	r := ui.Rect{W: min(max(a.W, 320), 900), H: min(max(a.H, 240), 700)}
	if h := d.o.Host; h != nil {
		if me, ok := h.Content(d.w.ID); ok {
			r.X, r.Y = max(me.X+a.X+24, 1), max(me.Y+a.Y+24, 1)
		}
	}
	return r
}

// carryWindow makes the drag of a torn-off window's only tab carry the
// window: it moves with the pointer, as by a title bar, and over another
// window of the layout it fades, that window showing where it would dock.
// Where the system says nowhere where the pointer is, the tab is dragged
// as any other.
func (d *dockView) carryWindow(dr *tabDrag) {
	h := d.o.Host
	if h == nil || d.w.ID == "main" || countPanes(d.w.Root) != 1 {
		return
	}
	p, ok := h.Pointer()
	me, placed := h.Content(d.w.ID)
	if !ok || !placed {
		return
	}
	dr.carry, dr.anchor = true, ui.Point{X: p.X - me.X, Y: p.Y - me.Y}
}

// carry moves the window carried with the pointer, and notes over which
// window of the layout the pointer is, for that window's Dock to show.
func (d *dockView) carry() {
	st, l, h := d.st, d.l, d.o.Host
	dr := st.tab
	p, ok := h.Pointer()
	if !ok {
		return
	}
	t := l.travel
	if t == nil || t.from != d.w.ID || t.pane != dr.pane {
		t = &dockTravel{pane: dr.pane, from: d.w.ID, title: dr.title, w: dr.w, h: dr.h, grab: dr.grab}
		l.travel = t
	}
	over, at := "", ui.Point{}
	for _, w := range l.Windows {
		if w.ID == d.w.ID {
			continue
		}
		if r, ok := h.Content(w.ID); ok && inside(r, p) {
			over, at = w.ID, ui.Point{X: p.X - r.X, Y: p.Y - r.Y}
			break
		}
	}
	if over != t.over {
		t.op = nil
	}
	t.over, t.at = over, at
	dr.screen = p
	h.Move(d.w.ID, ui.Point{X: p.X - dr.anchor.X, Y: p.Y - dr.anchor.Y}, over != "")
	h.Sync()
	d.c.AnimationFrame()
}

// letCarry lets go of the window carried: over another window that takes
// its tab, the tab docks there and the window, left empty, closes; else
// the window stays where it was carried to.
func (d *dockView) letCarry() {
	st, l, h := d.st, d.l, d.o.Host
	dr := st.tab
	t := l.travel
	l.travel = nil
	if t != nil && t.over != "" && t.op != nil {
		op := t.op
		st.tabApply = func() {
			op()
			l.pruneWindows()
		}
		d.c.AnimationFrame()
	} else {
		st.tab = nil
		h.Move(d.w.ID, ui.Point{X: dr.screen.X - dr.anchor.X, Y: dr.screen.Y - dr.anchor.Y}, false)
	}
	h.Sync()
}

// topRight is the area in the top right corner of n: where a window
// without a title bar has its close button.
func topRight(n *Node) *Node {
	for n != nil && !n.IsArea() {
		if n.Dir == DockRow {
			n = n.Children[len(n.Children)-1]
		} else {
			n = n.Children[0]
		}
	}
	return n
}

// windowTitle names a window in a menu: the pane its first area shows.
func windowTitle(w *DockWindow) string {
	n := w.Root
	for n != nil && !n.IsArea() {
		n = n.Children[0]
	}
	if n == nil || len(n.Panes) == 0 {
		return w.ID
	}
	return n.Panes[min(n.Active, len(n.Panes)-1)].Title + " window"
}
