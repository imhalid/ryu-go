package mizan

import (
	"slices"
	"time"

	"github.com/egoist/mygo/ui"
)

// An area's panes are tabs in its header. A tab is moved by dragging it:
// it lifts off its place as a raised tile and follows the pointer, and the
// tabs of the stack it is over slide apart where it would go, closing the
// room it left. Over another area, a translucent region shows where it
// would dock: its left, right, top or bottom half (a new area beside it),
// or the whole of it (a tab in its stack); the region slides between them
// on a spring. Let go, the pane moves there: the tab flies from where it
// was let go into its place, a new area opens on a spring as its neighbor
// makes room, and an area left without panes folds away.

// tabGap is the room between two tabs.
const tabGap = 2

// tabSlot is a tab as the last frame laid it out.
type tabSlot struct {
	pane string
	rect ui.Rect
}

type zoneSide uint8

const (
	zoneNone   zoneSide = iota
	zoneStack           // a tab in the area's stack
	zoneLeft            // a new area on the area's left
	zoneRight           // on its right
	zoneTop             // above it
	zoneBottom          // under it
)

// dockZone is where a tab dragged would go.
type dockZone struct {
	area  string
	side  zoneSide
	strip bool // over the area's tabs: index is where among them
	index int
	rect  ui.Rect // the region shown, in the window; none when empty
	// target: chosen on a target of the compass, or of the window's edges
	// (edge: area is the window's root, and the tab docks along its side)
	target, edge bool
}

// tabDrag is a tab being dragged, from the moment it moved until the move
// it asked for applies (let).
type tabDrag struct {
	pane, from string
	title      string
	w, h       float32
	grab, pos  ui.Point // where the pointer holds the copy, and where it is
	vx         float32
	tilt       float32
	zone       dockZone
	let        bool
	// out: the pointer left the window; screen is where it is on the
	// screen, when known. remote: dragged from another window, over this one.
	out, remote bool
	screen      ui.Point
	known       bool
	// carry: the tab is its window's only one, which moves with it, held
	// at anchor (the pointer in the window's content)
	carry  bool
	anchor ui.Point
	// native: the system drags it (Wayland), its copy its own
	native bool
}

// tabLanding is a tab let go, flying from where its copy was to its place,
// once a frame after after laid it out.
type tabLanding struct {
	pane  string
	at    ui.Point
	tilt  float32
	after time.Time
}

// tabGhost is a tab closed, closing the room it left where it was.
type tabGhost struct {
	pane, title string
	index       int
	w           float32
}

// strip builds the tabs of area n, focused when it has the keyboard focus.
func (d *dockView) strip(n *Node, focused bool) {
	c, st, t := d.c, d.st, d.t
	s := ui.Row(c.Key("tabs")).Grow(1).Shrink(1).MinWidth(0).Height(headerH).AlignItems(ui.Center).Gap(tabGap)
	gen := ui.Local(s, "gen", func() int { return st.tabGen })
	settle := *gen != st.tabGen
	*gen = st.tabGen
	sb := s.LayoutBounds()

	// the tab dragged, when it is one of these, and where it would go here
	dr := st.tab
	if dr != nil && dr.carry {
		dr = nil
	}
	from, at := -1, -1
	var room float32
	if dr != nil {
		room = dr.w + tabGap
		for i, p := range n.Panes {
			if p.ID == dr.pane {
				from = i
			}
		}
		if z := dr.zone; z.area == n.ID && z.side == zoneStack && z.strip {
			at = z.index
		}
	}
	sel := min(n.Active, len(n.Panes)-1)
	var pill ui.Rect // the chosen tab, from the strip's top left, as it shows
	held, hidden := false, false
	var slots []tabSlot
	// a tab closed while the stack builds changes n.Panes: build this frame's
	panes := slices.Clone(n.Panes)
	s.Children(func() {
		m := 0
		ghosts := st.closing[n.ID]
		for i, p := range panes {
			for _, g := range ghosts {
				if g.index == i {
					d.closingTab(n.ID, g)
				}
			}
			var shift float32
			if dr != nil && i != from {
				if from >= 0 && i > from {
					shift -= room
				}
				if at >= 0 && m >= at {
					shift += room
				}
			}
			if i != from {
				m++
			}
			tb := d.tab(n, i, p, focused, shift, settle)
			b := tb.e.LayoutBounds()
			slots = append(slots, tabSlot{p.ID, b})
			if i == sel && b.W > 0 && sb.W > 0 {
				pill = ui.Rect{X: b.X - sb.X + tb.x + tb.lx, Y: b.Y - sb.Y + tb.ly, W: b.W, H: b.H}
				held, hidden = tb.pressed, tb.hidden
			}
		}
		for _, g := range ghosts {
			if g.index >= len(panes) {
				d.closingTab(n.ID, g)
			}
		}
		// in a window without a title bar, the room after the tabs moves the
		// window, as a title bar does
		if d.w.ID != "main" && d.o.Host != nil {
			ui.Box(c.Key("drag-window")).Grow(1).AlignSelf(ui.Stretch).MinWidth(0).DragWindow()
		}
	})
	st.tabs[n.ID] = slots

	// the chosen tab sits on a raised pill, Mizan's control, as Segmented's
	// chosen segment does: it slides to a new choice on a spring (straight
	// there the first time), gives under a press, and goes while its tab is
	// picked up
	if pill.W == 0 || hidden {
		return
	}
	placed := ui.Local(s, "placed", func() bool { return false })
	var px, pw float32
	if !*placed {
		px, pw = s.Jump("pill-x", pill.X), s.Jump("pill-w", pill.W)
		*placed = true
	} else {
		px, pw = s.AnimateTo("pill-x", pill.X, springPill), s.AnimateTo("pill-w", pill.W, springPill)
	}
	sq := s.AnimateTo("squeeze", b2f(held), springSqueeze)
	r := ui.Rect{X: px + 2*sq, Y: pill.Y + sq, W: pw - 4*sq, H: pill.H - 2*sq}
	s.Draw(func(p *ui.Painter, b ui.Rect) {
		t.paintRaised(p, ui.Rect{X: b.X + r.X, Y: b.Y + r.Y, W: r.W, H: r.H}, r.H/2, t.Shadows.ShadowRaised)
	})
}

// builtTab is a tab built: its place, how far it is moved to make room
// (x) and landing (lx, ly), whether it is pressed, and whether it is
// hidden, picked up.
type builtTab struct {
	e               ui.Element
	x, lx, ly       float32
	pressed, hidden bool
}

// tab builds tab i of area n, pane p, shift DIPs from its place to make
// room (straight there after a move applied, settle).
func (d *dockView) tab(n *Node, i int, p *Pane, focused bool, shift float32, settle bool) builtTab {
	c, st, t := d.c, d.st, d.t
	cl := t.Colors
	e := ui.Row(c.Key(p.ID)).Height(headerH - 8).Shrink(1).MinWidth(44).MaxWidth(180).AlignItems(ui.Stretch)
	if settle {
		e.Jump("shift", 0)
	}
	x := e.AnimateTo("shift", shift, springShift)
	e.Translate(x, 0)
	selected := i == n.Active
	dr := st.tab
	// carried, the tab stays: its window is what moves
	picked := dr != nil && dr.pane == p.ID && !dr.carry

	// let go, it flies from where its copy was to its place, once a frame
	// laid it out there
	land := st.tabLand
	landing := land != nil && land.pane == p.ID
	if landing && dr == nil && c.Now().After(land.after) {
		if b := e.LayoutBounds(); b.W > 0 {
			e.Jump("lx", land.at.X-b.X)
			e.Jump("ly", land.at.Y-b.Y)
			e.Jump("lift", 1)
			e.Jump("tilt", land.tilt)
			st.tabLand, landing = nil, false
		}
	}
	lx, ly := e.AnimateTo("lx", 0, springLand), e.AnimateTo("ly", 0, springLand)
	lift := e.AnimateTo("lift", 0, springLand)
	tilt := e.AnimateTo("tilt", 0, springLand)
	if picked || landing {
		// its place stays, empty, while the copy is held
		e.Opacity(0)
	}
	if landing {
		c.AnimationFrame()
	}

	// as Segmented's: the chosen one in the text color on its pill (the
	// second text color while its area isn't focused), the others muted,
	// darkening under the pointer
	hover := e.Hovered() && dr == nil
	text := cl.Muted
	switch {
	case selected && !focused:
		text = cl.Text2
	case selected, hover:
		text = cl.Text
	}
	out := builtTab{e: e, x: x, lx: lx, ly: ly, hidden: picked || landing}
	e.Children(func() {
		body := ui.Row(c).Grow(1).MinWidth(0).AlignItems(ui.Center).Radius(RadiusPill).
			Translate(lx, ly).Scale(1+0.03*lift, 1+0.03*lift).Rotate(tilt)
		body.Children(func() {
			b := ui.ButtonBase(c.Key("title")).Label(p.Title).Grow(1).MinWidth(0).Padding(0, 4, 0, 10).
				AlignItems(ui.Center).FocusRing(false).Cursor(ui.CursorPointer)
			if d.native {
				// the system drags the tab, where it doesn't say where windows
				// are (Wayland): from window to window too
				d.nativeSource(n, p, e, b)
			} else {
				d.pick(n, p, e, b)
			}
			if b.Clicked() && !picked && n.Active != i {
				n.Active = i
				d.changed()
			}
			// a double click maximizes the area, and brings it back
			if b.DoubleClicked() {
				d.toggleMax(n.ID)
			}
			// a right click opens the tab's menu
			menu := ui.Local(e, "menu", func() bool { return false })
			if b.RightClicked() {
				*menu = true
			}
			d.tabMenu(n, p, e, menu)
			if b.FocusVisible() {
				t.Shadows.FocusRing.Apply(body)
			}
			// pressed, the chosen tab's pill gives; another tab shrinks a little
			if b.Pressed() && !picked {
				out.pressed = selected
				if !selected {
					body.Scale(0.97, 0.97)
				}
			}
			b.Children(func() {
				ui.Text(c, p.Title).FontSize(FsSm).FontWeight(500).TextColor(text).SingleLine().Ellipsis("…").MinWidth(0).
					Transition(ui.ElementTransition{Duration: 160 * time.Millisecond, Colors: true})
			})
			d.closeButton(n, i, p, selected || hover, e.LayoutBounds().W)
		})
	})
	return out
}

// closeButton is a tab's close button: it shows on the tab chosen and the
// one under the pointer.
func (d *dockView) closeButton(n *Node, i int, p *Pane, shown bool, w float32) {
	c, st, t := d.c, d.st, d.t
	b := ui.ButtonBase(c.Key("close")).Label("Close "+p.Title).Size(16, 16).Center().Margin(0, 3, 0, 0).
		Radius(RadiusPill).Cursor(ui.CursorPointer).FocusRing(false).Shrink(0)
	on := b.AnimateTo("shown", b2f(shown || b.FocusVisible()), ui.Animation{Duration: 140 * time.Millisecond, Ease: ui.EaseOut})
	hot := b.AnimateTo("hot", b2f(b.Hovered()), ui.Animation{Duration: 120 * time.Millisecond, Ease: ui.EaseOut})
	if hot > 0.01 {
		b.Background(t.Colors.Fill.Alpha(hot))
	}
	press := b.AnimateTo("press", b2f(b.Pressed()), springSqueeze)
	b.Opacity(on).Scale(1-0.08*press, 1-0.08*press)
	if b.FocusVisible() {
		t.Shadows.FocusRing.Apply(b)
	}
	b.Children(func() {
		ui.Icon(c, iconClose).Size(12, 12).TextColor(t.Colors.Muted)
	})
	if b.Clicked() && st.tab == nil && d.l.closePane(p.ID) {
		// the tab goes at once; where it was, a copy closes its room
		st.closing[n.ID] = append(st.closing[n.ID], tabGhost{pane: p.ID, title: p.Title, index: i, w: w})
		d.changed()
		c.AnimationFrame()
	}
}

// closingTab is a tab closed, where it was: it narrows to nothing as it
// fades, and the tabs after it slide over.
func (d *dockView) closingTab(area string, g tabGhost) {
	c, st, cl := d.c, d.st, d.t.Colors
	e := ui.Row(c.Key("closing:" + g.pane)).Height(headerH - 8).Shrink(0).AlignItems(ui.Center).Clip()
	if seen := ui.Local(e, "seen", func() bool { return false }); !*seen {
		*seen = true
		e.Jump("k", 1)
	}
	k := e.AnimateTo("k", 0, ui.Animation{Duration: 200 * time.Millisecond, Ease: ui.EaseOut})
	// the gap after it goes with it
	e.Width(max(g.w*k, 0)).Margin(0, (k-1)*tabGap, 0, 0).Opacity(k)
	e.Children(func() {
		ui.Text(c, g.title).FontSize(FsSm).FontWeight(500).TextColor(cl.Muted).SingleLine().Padding(0, 8).Shrink(0)
	})
	if k < 0.01 {
		gs := st.closing[area]
		for j := range gs {
			if gs[j].pane == g.pane {
				st.closing[area] = append(gs[:j:j], gs[j+1:]...)
				break
			}
		}
		if len(st.closing[area]) == 0 {
			delete(st.closing, area)
		}
		// one more frame, without it
		c.AnimationFrame()
	}
}

// pick makes b, the title of tab e (pane p of area n), pick the tab up
// when dragged a few DIPs, and follows the pointer while it is held.
func (d *dockView) pick(n *Node, p *Pane, e, b ui.Element) {
	c, st := d.c, d.st
	moved := ui.Local(b, "moved", func() float32 { return 0 })
	dx, dy, pressed := b.Dragged()
	dr := st.tab
	mine := dr != nil && dr.pane == p.ID && !dr.let
	at := func() ui.Point {
		bb := b.LayoutBounds()
		px, py, _ := b.PointerPosition()
		return ui.Point{X: bb.X + px, Y: bb.Y + py}
	}
	switch {
	case pressed && mine:
		dr.pos = at()
		dr.vx, _ = b.DragVelocity()
	case pressed && dr == nil && st.tabLand == nil && st.drag == nil && st.corner == nil && st.fold == nil && d.maxed == "":
		// a press becomes a drag once it moved a few DIPs, not a click
		if *moved += abs32(dx) + abs32(dy); *moved > 4 {
			eb := e.LayoutBounds()
			pos := at()
			st.tab = &tabDrag{pane: p.ID, from: n.ID, title: p.Title, w: eb.W, h: eb.H,
				grab: ui.Point{X: pos.X - eb.X, Y: pos.Y - eb.Y}, pos: pos}
			d.carryWindow(st.tab)
			c.AnimationFrame()
		}
	case !pressed:
		*moved = 0
	}
	if mine && !pressed {
		dr.let = true
	}
	if mine || (dr != nil && dr.pane == p.ID) {
		b.Cursor(ui.CursorGrabbing)
	}
}

// aimTab finds where the tab dragged would go, the pointer where it is: in
// a stack (among its tabs when over them, else at its end), or beside an
// area, on the side the pointer is nearest; nowhere over its own area when
// it is the only pane there.
func (d *dockView) aimTab() {
	st := d.st
	dr := st.tab
	p := dr.pos
	// a target of the compass, or of the window's edges, wins
	if z, ok := d.aimCompass(p); ok {
		dr.zone, st.zone = z, z
		return
	}
	z := dockZone{}
	for id, a := range st.areas {
		if !inside(a, p) {
			continue
		}
		n := d.l.Find(id)
		if n == nil || !n.IsArea() {
			break
		}
		own := id == dr.from
		inner := ui.Rect{X: a.X + 3, Y: a.Y + 3, W: a.W - 6, H: a.H - 6}
		if p.Y < a.Y+headerH {
			// among its tabs: before the first whose middle is past the copy's
			mid := p.X - dr.grab.X + dr.w/2
			idx, from := 0, -1
			for i, s := range st.tabs[id] {
				if s.pane == dr.pane {
					from = i
				}
			}
			for i, s := range st.tabs[id] {
				if i == from {
					continue
				}
				x := s.rect.X + s.rect.W/2
				if from >= 0 && i > from {
					x -= dr.w + tabGap
				}
				if x < mid {
					idx++
				}
			}
			z = dockZone{area: id, side: zoneStack, strip: true, index: idx}
			if !own {
				z.rect = inner
			}
			break
		}
		if own && len(n.Panes) == 1 {
			break
		}
		u := (p.X - a.X) / a.W
		v := min(max((p.Y-a.Y-headerH)/max(a.H-headerH, 1), 0), 1)
		near := min(u, 1-u, v, 1-v)
		wide, tall := a.W >= 2*d.o.MinSize+dockGap, a.H >= 2*d.o.MinSize+dockGap
		switch {
		case near == u && wide && u < 0.3:
			z = dockZone{area: id, side: zoneLeft, rect: ui.Rect{X: inner.X, Y: inner.Y, W: inner.W / 2, H: inner.H}}
		case near == 1-u && wide && u > 0.7:
			z = dockZone{area: id, side: zoneRight, rect: ui.Rect{X: inner.X + inner.W/2, Y: inner.Y, W: inner.W / 2, H: inner.H}}
		case near == v && tall && v < 0.3:
			z = dockZone{area: id, side: zoneTop, rect: ui.Rect{X: inner.X, Y: inner.Y, W: inner.W, H: inner.H / 2}}
		case near == 1-v && tall && v > 0.7:
			z = dockZone{area: id, side: zoneBottom, rect: ui.Rect{X: inner.X, Y: inner.Y + inner.H/2, W: inner.W, H: inner.H / 2}}
		case !own:
			z = dockZone{area: id, side: zoneStack, index: len(n.Panes), rect: inner}
		}
		break
	}
	dr.zone = z
	if z.rect.W > 0 {
		st.zone = z
	}
}

// letTab asks for the move the tab let go wants, which applies as the next
// frame starts, and sends the tab flying to its place; nowhere, it flies
// back.
func (d *dockView) letTab() {
	st, l := d.st, d.l
	dr := st.tab
	z := dr.zone
	st.applyAt = d.c.Now()
	if dr.carry {
		d.letCarry()
		return
	}
	if dr.out && d.o.Host != nil {
		d.letOut()
		return
	}
	st.tabLand = &tabLanding{pane: dr.pane, at: ui.Point{X: dr.pos.X - dr.grab.X, Y: dr.pos.Y - dr.grab.Y}, tilt: dr.tilt, after: d.c.Now()}
	if a, i := l.paneAt(dr.pane); z.side == zoneStack && a != nil && a.ID == z.area && i == z.index {
		// back where it was
		st.tab = nil
		return
	}
	if st.tabApply = d.dropOp(z, dr.pane); st.tabApply == nil {
		st.tab = nil
		return
	}
	d.c.AnimationFrame()
}

// dropOp is what letting the tab of pane go in zone z does to the layout;
// nil when z takes nothing.
func (d *dockView) dropOp(z dockZone, pane string) func() {
	st, l := d.st, d.l
	to := z.area
	switch z.side {
	case zoneNone:
		return nil
	case zoneStack:
		return func() {
			l.remember()
			l.movePane(pane, to, z.index)
		}
	}
	r := st.areas[to]
	if z.edge {
		r = d.root.LayoutBounds()
	}
	dir, first, length := DockRow, z.side == zoneLeft, r.W
	if z.side == zoneTop || z.side == zoneBottom {
		dir, first, length = DockColumn, z.side == zoneTop, r.H
	}
	keep, made := (length-dockGap)/2, (length-dockGap)/2
	if z.edge {
		// along the window's edge, a quarter of it
		made = d.edgeLength(z.side, r)
		keep = length - made - dockGap
	}
	var room float32
	if _, p, _ := l.where(to); p != nil {
		room = st.room[p.ID]
	}
	return func() {
		l.remember()
		if n := l.dockPane(pane, to, dir, first, keep, made, room); n != nil {
			l.mark(n.ID, true)
		}
	}
}

// applyTab applies the move a tab let go asked for, as the frame starts,
// so that every stack builds the new order together.
func (d *dockView) applyTab() {
	st := d.st
	if st.tabApply == nil || !d.c.Now().After(st.applyAt) {
		return
	}
	st.tabApply()
	st.tabApply, st.tab = nil, nil
	st.tabGen++
	// the tab lands once a frame laid out its new place
	if st.tabLand != nil {
		st.tabLand.after = d.c.Now()
	}
	d.changed()
	d.c.AnimationFrame()
}

// tabCopy is the copy of the tab dragged, lifted off the header as a
// raised tile: it follows the pointer, leaning into the motion, and stays
// where it was let go until the tab takes over.
func (d *dockView) tabCopy() {
	c, st, t := d.c, d.st, d.t
	dr, land := st.tab, st.tabLand
	if dr == nil && land == nil || dr != nil && (dr.carry || dr.native) {
		return
	}
	// the copy keeps the size the tab had when it was picked up
	size := ui.Local(d.root, "copy-size", func() ui.Point { return ui.Point{} })
	title := ""
	var at ui.Point
	if dr != nil {
		title, *size = dr.title, ui.Point{X: dr.w, Y: dr.h}
		at = ui.Point{X: dr.pos.X - dr.grab.X, Y: dr.pos.Y - dr.grab.Y}
	}
	if land != nil && (dr == nil || dr.let) {
		at = land.at
	}
	if dr == nil {
		// flying back, or landing after the move applied: the pane says what
		if a, i := d.l.paneAt(land.pane); a != nil {
			title = a.Panes[i].Title
		}
	}
	w, h := size.X, size.Y
	ui.Overlay(c, func() {
		g := ui.Row(c.Key(d.over("tab-copy", 0))).Absolute().Left(at.X).Top(at.Y).Width(w).Height(h).PassThrough().
			AlignItems(ui.Center).Padding(0, 8)
		if seen := ui.Local(g, "picked", func() bool { return false }); !*seen {
			*seen = true
			g.Jump("lift", 0)
		}
		lift := g.AnimateTo("lift", 1, springLift)
		var vx float32
		if dr != nil && !dr.let {
			vx = dr.vx
		}
		tilt := g.AnimateTo("tilt", min(max(vx*0.0016, -2.5), 2.5), springTilt)
		if dr != nil {
			dr.tilt = tilt
		}
		s := 1 + 0.03*lift
		// over a target, it lets the target show through
		thin := g.AnimateTo("thin", b2f(dr != nil && !dr.let && dr.zone.target), ui.Animation{Duration: 120 * time.Millisecond, Ease: ui.EaseOut})
		g.Opacity(1 - 0.45*thin)
		t.tile(g, false, true)
		g.Radius(Concentric(t.R(RadiusControl), 4)).Scale(s, s).Rotate(tilt).Cursor(ui.CursorGrabbing)
		g.Children(func() {
			ui.Text(c, title).FontSize(FsSm).FontWeight(500).TextColor(t.Colors.Text).SingleLine().Ellipsis("…").MinWidth(0)
		})
	})
}

// zonePreview builds the region a tab dragged would dock into: a
// translucent wash of the accent with a hairline of it, sliding between
// zones and areas on a spring, fading where it was when the tab leaves.
func (d *dockView) zonePreview() {
	c, st, t := d.c, d.st, d.t
	cl := t.Colors
	at := d.root.LayoutBounds()
	e := ui.Box(c.Key("zone")).Absolute().Left(0).Top(0).Right(0).Bottom(0).PassThrough()
	on := st.tab != nil && !st.tab.let && st.tab.zone.rect.W > 0
	show := e.AnimateTo("show", b2f(on), ui.Animation{Duration: 140 * time.Millisecond, Ease: ui.EaseOut})
	r := st.zone.rect
	if show < 0.01 {
		r = ui.Rect{X: e.Jump("x", r.X), Y: e.Jump("y", r.Y), W: e.Jump("w", r.W), H: e.Jump("h", r.H)}
	} else {
		r = ui.Rect{X: e.AnimateTo("x", r.X, *springPop), Y: e.AnimateTo("y", r.Y, *springPop), W: e.AnimateTo("w", r.W, *springPop), H: e.AnimateTo("h", r.H, *springPop)}
	}
	if show < 0.01 || r.W < 1 || r.H < 1 {
		return
	}
	R := Concentric(t.R(RadiusControl), 3)
	wash, line := 0.1, 0.55
	if t.Dark {
		wash, line = 0.16, 0.7
	}
	e.Draw(func(p *ui.Painter, b ui.Rect) {
		rr := ui.Rect{X: r.X - at.X + b.X, Y: r.Y - at.Y + b.Y, W: r.W, H: r.H}
		p.Fill(rr, cl.ColorAccent500.Alpha(float32(wash)*show), R)
		p.Stroke(rr, cl.ColorAccent500.Alpha(float32(line)*show), R, 1)
	})
}
