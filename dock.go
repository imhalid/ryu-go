package mizan

import (
	"slices"
	"time"

	"github.com/egoist/mygo/ui"
)

// A Dock shows a window of a Layout as Blender shows its areas: the window
// filled edge to edge with areas on the card surface, a thin gap between
// them. The gaps are edges: dragging one resizes the areas on both sides,
// and every other edge lined up with it along the same line moves with it,
// as one; where edges meet, the crossing drags both ways at once. See DOCK.md for the plan this is the start of.

// DockOptions are the options of Dock.
type DockOptions struct {
	// Window is the window of the layout shown: "main" when empty.
	Window string
	// Kinds are what an area can show, in the order of its header's menu;
	// they build the panes, by Pane.Kind. A kind it doesn't know shows the
	// pane's title alone, and stays in the layout.
	Kinds []DockKind
	// OnChange runs when the user changed the layout (an edge let go, sizes
	// shared out): the moment to save it.
	OnChange func()
	// MinSize is the least an area shrinks to, in DIPs: 96 by default.
	MinSize float32
	// Host connects the Dock to the windows of the app: given one, a tab
	// dragged out of the window tears off into a window of its own, or
	// docks into another window of the layout it is let go over, and the
	// layout's windows open and close with it. mizan/dockwin makes one of
	// MyGo's windows.
	Host DockHost
}

// DockHost shows a layout's windows in the app's windows.
type DockHost interface {
	// Content is where the content of the layout's window id is on the
	// screen; ok is false where the system doesn't say (Wayland), and a tab
	// dragged out of a window then only tears off.
	Content(window string) (r ui.Rect, ok bool)
	// Sync opens a window for each of the layout's windows that has none,
	// closes those of windows it no longer holds, and redraws them all.
	Sync()
	// Pointer is where the pointer is on the screen; ok is false where the
	// system doesn't say, and a window is then moved by the system alone.
	Pointer() (p ui.Point, ok bool)
	// Move puts the content of window id at p on the screen, faded while
	// it is carried over another window, so that what that window shows of
	// where it would dock shows through.
	Move(window string, p ui.Point, faded bool)
}

// DockKind is a kind of pane: what an area shows (a table, a chart, a
// console…), named in the header's menu.
type DockKind struct {
	Kind  string // Pane.Kind
	Title string // in the menu, and the title of a pane made of this kind
	// Icon shows on the header's menu button, and by the kind in its menu;
	// a generic one when nil.
	Icon  *ui.SVG
	Build func(c *ui.Context, p *Pane)
}

// kind returns the kind named k, nil when there is none.
func (o *DockOptions) kind(k string) *DockKind {
	for i := range o.Kinds {
		if o.Kinds[i].Kind == k {
			return &o.Kinds[i]
		}
	}
	return nil
}

// dockGap is the width of an edge between two areas.
const dockGap = 4

// dockEdge is an edge between children i and i+1 of a split, as the last
// frame laid it out, for finding the edges lined up with one.
type dockEdge struct {
	split    string
	i        int
	vertical bool // between children side by side
	rect     ui.Rect
	length   float32 // the split's room for its children, without the edges
}

// dockDrag is an edge or a crossing being dragged, with the edges it moves.
type dockDrag struct {
	by       any        // what was pressed: an edgeKey or a crossKey
	seeds    []dockEdge // the edges pressed
	edges    []dockEdge // they and, unless alone, the edges lined up with them
	alone    bool       // Option held as it was pressed: the edges pressed alone
	finding  bool       // pressed in this frame: the edges lined up are not found yet
	p, press ui.Point   // the pointer, and where it pressed, in the window

	// an edge's line: where it was pressed across it (start), the span of
	// the edges moving along it (lo, hi), and how far they moved across it
	// (moved); settled onto snapTo
	start, lo, hi float32
	moved         float32
	snapped       bool
	snapTo        float32
	snapAt        time.Time
}

type dockState struct {
	edges   []dockEdge // this frame's, as the last frame laid them out
	drag    *dockDrag
	shown   []edgeKey   // the edges whose grips show: those that would move, or move
	segment dockSegment // the part of an edge Option would cut off, under the pointer

	areas  map[string]ui.Rect // this frame's areas, as the last frame laid them out
	nodes  map[string]dockPlace
	room   map[string]float32 // each split's room for its children, along it
	corner *dockCorner
	guide  dockGuide
	fold   *dockFold
	ghosts []dockGhost // areas gone from the tree, folding away where they were
	nghost int

	tabs     map[string][]tabSlot // this frame's tabs of each area, as the last frame laid them out
	tab      *tabDrag
	tabLand  *tabLanding
	tabApply func()                // the move a tab let go asked for, as the next frame starts
	applyAt  time.Time             // when it was let go: the move applies after it
	tabGen   int                   // counts the moves applied
	zone     dockZone              // the last region shown, to fade where it was
	compass  compassAt             // where the compass shows, from the last frame's aim
	closing  map[string][]tabGhost // tabs closed, by area, closing their room
	settled  map[string]bool       // nodes a cut made, which show at once where they are

	maxLast string // the area last maximized: it shrinks back after Max is cleared
	focused string // the area holding the keyboard focus in this frame
	hovered string // the area under the pointer in this frame
	focusTo string // the area to give the focus to in this frame

	layout  *Layout     // the layout shown: another one starts afresh
	tokens  *Tokens     // the theme it was shown in, to redraw the other windows in another
	windows string      // the layout's windows, to open and close the app's as they change
	landed  *tabLanding // a tab another window let go here, landing as the next frame starts

	// where a tab the system drags over the Dock would go, and its copy's
	// place, kept for the drop, which comes once it left
	nativeZone dockZone
	nativeAt   ui.Point
}

// Dock shows window o.Window of l, its areas built by o.Kinds. Give it a
// size: Fill it, or Grow it in its parent.
//
//	mizan.Dock(c.Key("work"), app.layout, mizan.DockOptions{
//		Kinds:    map[string]func(*ui.Context, *mizan.Pane){"files": app.files, "editor": app.editor},
//		OnChange: app.saveLayout,
//	}).Fill()
func Dock(c *ui.Context, l *Layout, o DockOptions) ui.Element {
	t := From(c)
	root := ui.Box(c).MinWidth(0).MinHeight(0)
	st := *ui.Local(root, "dock", func() *dockState {
		return &dockState{closing: map[string][]tabGhost{}, settled: map[string]bool{}}
	})
	d := &dockView{t: t, c: c, l: l, o: o, st: st, root: root}
	if o.Host != nil {
		_, known := o.Host.Pointer()
		d.native = !known
	}
	d.windows()
	w := l.Window(or(o.Window, "main"))
	if w == nil || w.Root == nil {
		return root
	}
	d.w = w
	last := slices.Clone(st.edges)
	st.edges = st.edges[:0]
	prev := st.nodes
	st.areas, st.room, st.tabs = map[string]ui.Rect{}, map[string]float32{}, map[string][]tabSlot{}
	st.nodes = map[string]dockPlace{}
	st.focused, st.hovered = "", ""
	if d.o.MinSize <= 0 {
		d.o.MinSize = 96
	}
	// maximized, an area grows over the window; brought back, the same area
	// shrinks to its place
	if w.Max != "" && l.Find(w.Max) == nil {
		w.Max = ""
	}
	if w.Max != "" {
		st.maxLast = w.Max
	}
	d.max = root.AnimateTo("max", b2f(w.Max != ""), springOpen)
	if st.maxLast != "" && d.max > 0.001 && l.Find(st.maxLast) != nil {
		d.maxed = st.maxLast
	}
	folded := ""
	if f := st.fold; f != nil && f.done {
		// a join folded away: the layout changes as the frame starts, so that
		// every split builds the new tree
		st.fold, st.guide.mode, folded = nil, guideNone, f.gone
		if f.empty {
			// an area left without panes: Undo goes back to before it was emptied
			l.removeArea(f.gone)
			d.changed()
		} else if l.joinArea(f.keep, f.gone) {
			d.changed()
		}
	}
	d.applyTab()
	d.arrive()
	d.visit()
	d.nativeTarget()
	if w.Max == "" {
		d.vanished(prev, folded)
	}
	root.Children(func() {
		d.node(w.Root).Fill()
		d.hits()
		for _, x := range d.crossings() {
			d.cross(x)
		}
		d.ghosts()
		if d.maxed != "" {
			d.maximized()
		}
		d.guide()
		d.zonePreview()
		d.tabCopy()
		d.compass()
	})
	d.keys()
	if dr := st.tab; dr != nil && !dr.let && dr.carry {
		d.carry()
	} else if dr != nil && !dr.let {
		before := dr.zone
		d.aimTab()
		if dr.native {
			d.st.nativeZone, d.st.nativeAt = dr.zone, ui.Point{X: dr.pos.X - dr.grab.X, Y: dr.pos.Y - dr.grab.Y}
		} else if dr.remote {
			d.offer()
		} else {
			d.follow()
		}
		if dr.zone != before {
			c.AnimationFrame()
		}
	} else if dr != nil && st.tabApply == nil && dr.let {
		d.letTab()
	}
	if k := st.corner; k != nil {
		before := st.guide
		if k.let {
			d.let()
		} else {
			d.aim()
		}
		if st.guide != before {
			c.AnimationFrame()
		}
	}
	if dd := st.drag; dd != nil && dd.finding {
		dd.edges, dd.finding = d.spread(dd.seeds, dd.alone), false
		if _, edge := dd.by.(edgeKey); edge && len(dd.seeds) > 0 {
			lo, hi := spanOf(dd.edges[0])
			for _, e := range dd.edges {
				a, b := spanOf(e)
				lo, hi = min(lo, a), max(hi, b)
			}
			dd.span(lo, hi, dd.seeds[0])
		}
	}
	// the grips show on the edges that move, or would move under the pointer:
	// a line through several splits says it is one before it is dragged
	var shown []edgeKey
	switch {
	case st.drag != nil:
		shown = keysOf(st.drag.edges)
	case len(d.hover) > 0:
		alone := c.Modifiers()&ui.Alt != 0
		shown = keysOf(d.spread(d.hover, alone))
		// with Option, an edge edges across meet from both sides is cut
		// there: the grip shows on the part that would move
		seg := dockSegment{}
		if plan, ok := d.cutPlan(d.hover[0], d.hoverAt); ok && alone && len(d.hover) == 1 {
			seg = dockSegment{key: edgeKey{d.hover[0].split, d.hover[0].i}, lo: plan.lo, hi: plan.hi}
		}
		if seg != st.segment {
			st.segment = seg
			c.AnimationFrame()
		}
	}
	if len(d.hover) == 0 && st.drag == nil && st.segment != (dockSegment{}) {
		st.segment = dockSegment{}
		c.AnimationFrame()
	}
	if !slices.Equal(last, st.edges) {
		// the crossings were placed where the edges were: once the layout
		// holds still, they are where the edges are
		c.AnimationFrame()
	}
	if !sameKeys(shown, st.shown) {
		// the edges were built with the last ones: one more frame shows these
		st.shown = shown
		c.AnimationFrame()
	}
	return root
}

type dockView struct {
	t       *Tokens
	c       *ui.Context
	l       *Layout
	o       DockOptions
	st      *dockState
	root    ui.Element
	w       *DockWindow
	hover   []dockEdge // the edges under the pointer in this frame
	hoverAt ui.Point   // and the pointer
	max     float32    // how far the area maximized has grown over the window
	maxed   string     // the area shown maximized, growing or shrinking
	// native: tabs are dragged by the system, where it doesn't say where
	// windows are (Wayland)
	native bool
}

// node builds n: an area, or a split of its children and the edges
// between them.
func (d *dockView) node(n *Node) ui.Element {
	if n.IsArea() {
		return d.area(n)
	}
	c := d.c
	var box ui.Element
	if n.Dir == DockColumn {
		box = ui.Column(c.Key(n.ID))
	} else {
		box = ui.Row(c.Key(n.ID))
	}
	box.MinWidth(0).MinHeight(0).AlignItems(ui.Stretch)
	// the room the children share, without the edges between them
	b := box.LayoutBounds()
	length := b.W
	if n.Dir == DockColumn {
		length = b.H
	}
	length -= dockGap * float32(len(n.Children)-1)
	d.st.room[n.ID] = length
	held := d.held(n)
	glide := held && d.c.Now().Sub(d.st.drag.snapAt) < snapGlide
	if glide {
		d.c.AnimationFrame()
	}
	// a join folding: the area taken in shrinks to nothing, its edge with it,
	// as its neighbor grows over it
	fold := d.st.fold
	gone, keep := -1, -1
	if fold != nil && fold.split == n.ID {
		for i, ch := range n.Children {
			switch ch.ID {
			case fold.gone:
				gone = i
			case fold.keep:
				keep = i
			}
		}
	}
	fs := make([]float32, len(n.Children))
	for i, ch := range n.Children {
		// the sizes show through springs, so that a change of them (shared
		// out, an area joining) glides; under the pointer they follow it at
		// once
		size := n.Sizes[i]
		if gone >= 0 && keep >= 0 {
			switch i {
			case gone:
				size = 0
			case keep:
				size += n.Sizes[gone]
			}
		}
		if st := d.st; st.settled[ch.ID] {
			// a node a cut made shows where its parts were: no glide
			delete(st.settled, ch.ID)
			box.Jump("f:"+ch.ID, size)
		}
		if d.l.opening[ch.ID] {
			// an area a tab docked into opens from nothing, its neighbors
			// making room
			delete(d.l.opening, ch.ID)
			box.Jump("f:"+ch.ID, 0)
		}
		switch {
		case held && glide:
			// settling onto an edge, or leaving it, the sizes glide there
			fs[i] = box.AnimateTo("f:"+ch.ID, size, springGuide)
		case held:
			fs[i] = box.Jump("f:"+ch.ID, size)
		default:
			fs[i] = box.AnimateTo("f:"+ch.ID, size, springShift)
		}
	}
	if gone >= 0 && keep >= 0 && (fs[gone] < 0.002 || d.c.Now().Sub(fold.start) > 700*time.Millisecond) {
		fold.done = true
		d.c.AnimationFrame()
	}
	box.Children(func() {
		for i, ch := range n.Children {
			if i > 0 {
				scale := float32(1)
				if gone >= 0 && (i == gone || i-1 == gone) && (i == keep || i-1 == keep) {
					scale = min(max(fs[gone]/n.Sizes[gone], 0), 1)
				}
				d.edge(n, i-1, length, scale)
			}
			el := d.node(ch).Basis(0).Grow(max(fs[i], 0.0001)).Shrink(1)
			d.st.nodes[ch.ID] = dockPlace{rect: el.LayoutBounds(), parent: n.ID, area: ch.IsArea()}
		}
	})
	return box
}

// held reports whether an edge of split n is being dragged.
func (d *dockView) held(n *Node) bool {
	if d.st.drag == nil {
		return false
	}
	for _, e := range d.st.drag.edges {
		if e.split == n.ID {
			return true
		}
	}
	return false
}

// edge builds the edge between children i and i+1 of split n: the gap
// between them, with its grip. What takes the pointer is built over the
// window (hit), so that a press outlives a change of the tree. A short grip
// grows in it while it would move, longer as it is held; no color comes
// with it.
func (d *dockView) edge(n *Node, i int, length, scale float32) {
	c, st := d.c, d.st
	cl := d.t.Colors
	vertical := n.Dir == DockRow
	k := edgeKey{n.ID, i}
	e := ui.Box(c.Key(k)).Shrink(0).AlignSelf(ui.Stretch)
	if vertical {
		e.Width(dockGap * scale)
	} else {
		e.Height(dockGap * scale)
	}
	me := dockEdge{split: n.ID, i: i, vertical: vertical, rect: e.LayoutBounds(), length: length}
	st.edges = append(st.edges, me)

	// the grip: grows in on every edge that would move, longer as it is held;
	// on the part of an edge Option would cut off, in that part's middle
	dd := st.drag
	moving := dd != nil && dd.has(k)
	hot := e.AnimateTo("hot", b2f(slices.Contains(st.shown, k)), ui.Animation{Duration: 160 * time.Millisecond, Ease: ui.EaseOut})
	held := e.AnimateTo("held", b2f(moving), springSqueeze)
	along := me.rect.Y + me.rect.H/2
	if !vertical {
		along = me.rect.X + me.rect.W/2
	}
	if sg := st.segment; sg.key == k {
		along = (sg.lo + sg.hi) / 2
	}
	mid := e.AnimateTo("grip-at", along, ui.Animation{Duration: 160 * time.Millisecond, Ease: ui.EaseOut})
	// settled onto a lined-up edge, a light runs along it from the pointer
	var spark float32
	if moving && dd.snapped {
		spark = d.root.AnimateTo("edge-spark", 1, sparkCurve)
	}
	if hot < 0.01 && spark <= 0 {
		return
	}
	color := cl.Muted.Alpha(0.5 * hot)
	light := d.t.sparkLight()
	e.Draw(func(p *ui.Painter, r ui.Rect) {
		long := (14 + 18*held) * hot
		if vertical {
			y := r.Y + mid - me.rect.Y
			if me.rect.H == 0 {
				y = r.Y + r.H/2
			}
			p.Fill(ui.Rect{X: r.X + r.W/2 - 1, Y: y - long/2, W: 2, H: long}, color, 1)
		} else {
			x := r.X + mid - me.rect.X
			if me.rect.W == 0 {
				x = r.X + r.W/2
			}
			p.Fill(ui.Rect{X: x - long/2, Y: r.Y + r.H/2 - 1, W: long, H: 2}, color, 1)
		}
		if spark > 0 && spark < 1 {
			if vertical {
				perp := min(max(dd.p.Y-me.rect.Y+r.Y, r.Y), r.Y+r.H)
				sparks(p, true, r.X+r.W/2, perp, r.Y, r.Y+r.H, spark, 1, light)
			} else {
				perp := min(max(dd.p.X-me.rect.X+r.X, r.X), r.X+r.W)
				sparks(p, false, r.Y+r.H/2, perp, r.X, r.X+r.W, spark, 1, light)
			}
		}
	})
}

// hits builds what takes the pointer on the edges, over the window, where
// the last frame laid them out: dragged, an edge moves, and every edge
// lined up with it moves along (with Option held, it goes alone, cut off
// where edges across it meet it from both sides); a double click shares
// its split's room out equally. The edge held keeps taking the pointer
// when a change of the tree took it away.
func (d *dockView) hits() {
	c, st := d.c, d.st
	at := d.root.LayoutBounds()
	built := false
	for _, me := range st.edges {
		k := edgeKey{me.split, me.i}
		r := me.rect
		e := ui.Box(c.Key(k)).Absolute().Left(r.X - at.X).Top(r.Y - at.Y).Width(r.W).Height(r.H)
		vertical := me.vertical
		if dd := st.drag; dd != nil && dd.by == k && len(dd.edges) > 0 {
			// the edge held may now be another: the cursor says what moves
			vertical = dd.edges[0].vertical
		}
		if vertical {
			e.Cursor(ui.CursorResizeEW)
		} else {
			e.Cursor(ui.CursorResizeNS)
		}
		d.handle(e, k, []dockEdge{me})
		if dd := st.drag; dd != nil && dd.by == k {
			built = true
		}
		if e.DoubleClicked() {
			if n := d.l.Find(me.split); n != nil {
				for j := range n.Sizes {
					n.Sizes[j] = 1 / float32(len(n.Sizes))
				}
				d.changed()
			}
		}
	}
	if dd := st.drag; dd != nil && !built {
		if k, ok := dd.by.(edgeKey); ok {
			e := ui.Box(c.Key(k)).Absolute().Left(0).Top(0).Width(0).Height(0)
			if len(dd.edges) > 0 && !dd.edges[0].vertical {
				e.Cursor(ui.CursorResizeNS)
			} else {
				e.Cursor(ui.CursorResizeEW)
			}
			d.handle(e, k, dd.seeds)
		}
	}
}

// handle makes e drag the edges seeds, as by: pressed, it takes the drag
// (with Option held, the seeds alone, an edge cut off where edges across
// meet it); while held, the vertical edges follow the pointer across and
// the others along, settling onto an edge they would line up with; let go,
// the layout is saved.
func (d *dockView) handle(e ui.Element, by any, seeds []dockEdge) {
	st := d.st
	dx, dy, pressed := e.Dragged()
	mine := st.drag != nil && st.drag.by == by
	b := e.LayoutBounds()
	px, py, _ := e.PointerPosition()
	p := ui.Point{X: b.X + px, Y: b.Y + py}
	switch {
	case pressed && st.drag == nil && st.corner == nil && st.fold == nil && st.tab == nil && d.maxed == "":
		// the edges lined up are found once the frame built them all
		alone := d.c.Modifiers()&ui.Alt != 0
		dd := &dockDrag{by: by, seeds: seeds, edges: seeds, alone: alone, finding: true, p: p, press: p}
		if _, edge := by.(edgeKey); edge && alone && len(seeds) == 1 {
			if plan, ok := d.cutPlan(seeds[0], p); ok {
				if ne, ok := d.cut(plan); ok {
					dd.seeds, dd.edges, dd.finding = []dockEdge{ne}, []dockEdge{ne}, false
					dd.span(plan.lo, plan.hi, seeds[0])
				}
			}
		}
		st.drag = dd
		mine = true
	case !pressed && mine:
		st.drag, st.segment = nil, dockSegment{}
		d.changed()
		return
	}
	if mine {
		dd := st.drag
		dd.p = p
		if _, edge := by.(edgeKey); edge && !dd.finding && len(dd.edges) > 0 {
			d.slide(dd)
		} else {
			for _, le := range dd.edges {
				if le.vertical && dx != 0 {
					d.move(le, dx)
				} else if !le.vertical && dy != 0 {
					d.move(le, dy)
				}
			}
		}
	}
	if st.drag == nil && e.Hovered() {
		d.hover = append(d.hover, seeds...)
		d.hoverAt = p
	}
}

// dockCross is where edges of both directions meet: a vertical edge, and
// the edges across it ending at it, from one side or both.
type dockCross struct {
	key   crossKey
	at    ui.Point // its middle, in the window
	seeds []dockEdge
}

// crossKey keys a crossing by its vertical edge and the first edge across.
type crossKey struct{ v, h edgeKey }

// crossSize is the side of a crossing's square: wider than an edge, so the
// pointer finds it.
const crossSize = 12

// crossings are the places this frame's edges meet, as the last frame laid
// them out.
func (d *dockView) crossings() []dockCross {
	var out []dockCross
	for _, v := range d.st.edges {
		if !v.vertical || v.rect.H == 0 {
			continue
		}
		for _, h := range d.st.edges {
			if h.vertical || h.rect.W == 0 {
				continue
			}
			y := h.rect.Y + h.rect.H/2
			if y < v.rect.Y || y > v.rect.Y+v.rect.H {
				continue
			}
			ends := abs32(h.rect.X+h.rect.W-v.rect.X) <= 1 || abs32(h.rect.X-(v.rect.X+v.rect.W)) <= 1
			if !ends {
				continue
			}
			merged := false
			for k := range out {
				if out[k].key.v == (edgeKey{v.split, v.i}) && abs32(out[k].at.Y-y) <= dockGap+1 {
					out[k].seeds = append(out[k].seeds, h)
					merged = true
				}
			}
			if !merged {
				out = append(out, dockCross{
					key:   crossKey{edgeKey{v.split, v.i}, edgeKey{h.split, h.i}},
					at:    ui.Point{X: v.rect.X + v.rect.W/2, Y: y},
					seeds: []dockEdge{v, h},
				})
			}
		}
	}
	return out
}

// cross builds crossing x: a square over where the edges meet, dragged in
// both directions at once; under the pointer a small cross grows in it.
func (d *dockView) cross(x dockCross) {
	c, st := d.c, d.st
	at := d.root.LayoutBounds()
	e := ui.Box(c.Key(x.key)).Absolute().
		Left(x.at.X - at.X - crossSize/2).Top(x.at.Y - at.Y - crossSize/2).
		Width(crossSize).Height(crossSize).Cursor(ui.CursorMove)
	d.handle(e, x.key, x.seeds)
	moving := st.drag != nil && st.drag.by == x.key
	hot := e.AnimateTo("hot", b2f(e.Hovered() || moving), ui.Animation{Duration: 160 * time.Millisecond, Ease: ui.EaseOut})
	if hot < 0.01 {
		return
	}
	color := d.t.Colors.Muted.Alpha(0.5 * hot)
	e.Draw(func(p *ui.Painter, r ui.Rect) {
		arm := 5 * hot
		mx, my := r.X+r.W/2, r.Y+r.H/2
		p.Fill(ui.Rect{X: mx - arm, Y: my - 1, W: 2 * arm, H: 2}, color, 1)
		p.Fill(ui.Rect{X: mx - 1, Y: my - arm, W: 2, H: 2 * arm}, color, 1)
	})
}

// edgeKey keys an edge apart from the nodes.
type edgeKey struct {
	split string
	i     int
}

func (dd *dockDrag) has(k edgeKey) bool {
	for _, e := range dd.edges {
		if (edgeKey{e.split, e.i}) == k {
			return true
		}
	}
	return false
}

// spread is seeds and, unless alone, the edges lined up with each.
func (d *dockView) spread(seeds []dockEdge, alone bool) []dockEdge {
	if alone {
		return seeds
	}
	var out []dockEdge
	for _, s := range seeds {
		for _, e := range d.lined(s) {
			if !slices.ContainsFunc(out, func(o dockEdge) bool { return o.split == e.split && o.i == e.i }) {
				out = append(out, e)
			}
		}
	}
	return out
}

func keysOf(edges []dockEdge) []edgeKey {
	var out []edgeKey
	for _, e := range edges {
		out = append(out, edgeKey{e.split, e.i})
	}
	return out
}

// sameKeys reports whether a and b hold the same keys, in any order.
func sameKeys(a, b []edgeKey) bool {
	if len(a) != len(b) {
		return false
	}
	for _, k := range a {
		if !slices.Contains(b, k) {
			return false
		}
	}
	return true
}

// lined is e and the edges lined up with it, of those this frame built, as
// the last frame laid them out: edges of the same direction on the same line, each touching one
// already found, so that a line running through several splits moves as
// one, as Blender's do.
func (d *dockView) lined(e dockEdge) []dockEdge {
	out := []dockEdge{e}
	if e.rect.W == 0 && e.rect.H == 0 {
		return out
	}
	mid := func(r ui.Rect, vertical bool) float32 {
		if vertical {
			return r.X + r.W/2
		}
		return r.Y + r.H/2
	}
	span := func(r ui.Rect, vertical bool) (float32, float32) {
		if vertical {
			return r.Y, r.Y + r.H
		}
		return r.X, r.X + r.W
	}
	line := mid(e.rect, e.vertical)
	for grew := true; grew; {
		grew = false
		for _, o := range d.st.edges {
			if o.vertical != e.vertical || abs32(mid(o.rect, o.vertical)-line) > 1 {
				continue
			}
			dup := false
			for _, f := range out {
				if f.split == o.split && f.i == o.i {
					dup = true
				}
			}
			if dup {
				continue
			}
			a0, a1 := span(o.rect, o.vertical)
			for _, f := range out {
				b0, b1 := span(f.rect, f.vertical)
				// touching, across the gap of an edge between them
				if a0 <= b1+dockGap+1 && b0 <= a1+dockGap+1 {
					out = append(out, o)
					grew = true
					break
				}
			}
		}
	}
	return out
}

// move moves edge e by by DIPs, within the least size of the areas on
// both sides, and returns how far it moved.
func (d *dockView) move(e dockEdge, by float32) float32 {
	n := d.l.Find(e.split)
	if n == nil || e.i+1 >= len(n.Sizes) || e.length <= 0 {
		return 0
	}
	a, b := n.Sizes[e.i], n.Sizes[e.i+1]
	least := min(d.o.MinSize/e.length, (a+b)/2)
	na := min(max(a+by/e.length, least), a+b-least)
	n.Sizes[e.i], n.Sizes[e.i+1] = na, a+b-na
	return (na - a) * e.length
}
