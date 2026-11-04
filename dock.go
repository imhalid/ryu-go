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
	// Kinds build the panes, by Pane.Kind. A kind it doesn't know shows
	// the pane's title alone, and stays in the layout.
	Kinds map[string]func(c *ui.Context, p *Pane)
	// OnChange runs when the user changed the layout (an edge let go, sizes
	// shared out): the moment to save it.
	OnChange func()
	// MinSize is the least an area shrinks to, in DIPs: 96 by default.
	MinSize float32
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
	by      any        // what was pressed: an edgeKey or a crossKey
	seeds   []dockEdge // the edges pressed
	edges   []dockEdge // they and, unless alone, the edges lined up with them
	alone   bool       // Option held as it was pressed: the edges pressed alone
	finding bool       // pressed in this frame: the edges lined up are not found yet
}

type dockState struct {
	edges []dockEdge // this frame's, as the last frame laid them out
	drag  *dockDrag
	shown []edgeKey // the edges whose grips show: those that would move, or move

	areas  map[string]ui.Rect // this frame's areas, as the last frame laid them out
	room   map[string]float32 // each split's room for its children, along it
	corner *dockCorner
	guide  dockGuide
	fold   *dockFold
	born   map[string]bool // areas a split made, to fade in
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
	w := l.Window(or(o.Window, "main"))
	if w == nil || w.Root == nil {
		return root
	}
	st := *ui.Local(root, "dock", func() *dockState { return &dockState{born: map[string]bool{}} })
	last := slices.Clone(st.edges)
	st.edges = st.edges[:0]
	st.areas, st.room = map[string]ui.Rect{}, map[string]float32{}
	d := &dockView{t: t, c: c, l: l, o: o, st: st, root: root}
	if d.o.MinSize <= 0 {
		d.o.MinSize = 96
	}
	if f := st.fold; f != nil && f.done {
		// a join folded away: the layout changes as the frame starts, so that
		// every split builds the new tree
		st.fold, st.guide.mode = nil, guideNone
		if l.joinArea(f.keep, f.gone) {
			d.changed()
		}
	}
	root.Children(func() {
		d.node(w.Root).Fill()
		for _, x := range d.crossings() {
			d.cross(x)
		}
		d.guide()
	})
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
	}
	// the grips show on the edges that move, or would move under the pointer:
	// a line through several splits says it is one before it is dragged
	var shown []edgeKey
	switch {
	case st.drag != nil:
		shown = keysOf(st.drag.edges)
	case len(d.hover) > 0:
		shown = keysOf(d.spread(d.hover, c.Modifiers()&ui.Alt != 0))
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
	t     *Tokens
	c     *ui.Context
	l     *Layout
	o     DockOptions
	st    *dockState
	root  ui.Element
	hover []dockEdge // the edges under the pointer in this frame
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
		if held {
			fs[i] = box.Jump("f:"+ch.ID, size)
		} else {
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
			d.node(ch).Basis(0).Grow(max(fs[i], 0.0001)).Shrink(1)
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

// area builds area n: the card surface, the shown pane's title over its
// content.
func (d *dockView) area(n *Node) ui.Element {
	c, t := d.c, d.t
	cl := t.Colors
	a := ui.Column(c.Key(n.ID)).MinWidth(0).MinHeight(0).Radius(t.R(RadiusControl)).Background(cl.Card).Clip().
		DrawOver(func(p *ui.Painter, r ui.Rect) { p.Stroke(r, cl.Hairline, t.R(RadiusControl), 1) })
	d.st.areas[n.ID] = a.LayoutBounds()
	if f := d.st.fold; f != nil && f.gone == n.ID {
		d.folding(a)
	}
	// an area a split made: its surface was there, the area's it came out
	// of; what it shows fades in
	if d.st.born[n.ID] {
		delete(d.st.born, n.ID)
		a.Jump("in", 0)
	}
	in := a.AnimateTo("in", 1, ui.Animation{Duration: 240 * time.Millisecond, Ease: ui.EaseOut})
	a.Children(func() {
		defer func() {
			for k := range 4 {
				d.corner(n, k)
			}
		}()
		if len(n.Panes) == 0 {
			return
		}
		p := n.Panes[min(n.Active, len(n.Panes)-1)]
		ui.Row(c).Height(28).Padding(0, 12).AlignItems(ui.Center).Shrink(0).InsetShadow(0, -1, 0, 0, cl.Hairline.Alpha(in)).Opacity(in).Children(func() {
			ui.Text(c, p.Title).FontSize(FsSm).FontWeight(500).TextColor(cl.Text2).SingleLine().Ellipsis("…")
		})
		body := ui.Column(c.Key(p.ID)).Grow(1).MinHeight(0).MinWidth(0).Opacity(in)
		body.Children(func() {
			if build := d.o.Kinds[p.Kind]; build != nil {
				build(c, p)
			}
		})
	})
	return a
}

// edge builds the edge between children i and i+1 of split n: dragged,
// it moves, and every edge lined up with it moves along (with Option held,
// it goes alone); a double click shares the split's room out equally. A
// short grip grows in it while it would move, longer as it is held; no color
// comes with it.
func (d *dockView) edge(n *Node, i int, length, scale float32) {
	c, st := d.c, d.st
	cl := d.t.Colors
	vertical := n.Dir == DockRow
	e := ui.Box(c.Key(edgeKey{n.ID, i})).Shrink(0).AlignSelf(ui.Stretch)
	if vertical {
		e.Width(dockGap * scale).Cursor(ui.CursorResizeEW)
	} else {
		e.Height(dockGap * scale).Cursor(ui.CursorResizeNS)
	}
	me := dockEdge{split: n.ID, i: i, vertical: vertical, rect: e.LayoutBounds(), length: length}
	st.edges = append(st.edges, me)

	d.handle(e, edgeKey{n.ID, i}, []dockEdge{me})
	if e.DoubleClicked() {
		for k := range n.Sizes {
			n.Sizes[k] = 1 / float32(len(n.Sizes))
		}
		if d.o.OnChange != nil {
			d.o.OnChange()
		}
	}

	// the grip: grows in on every edge that would move, longer as it is held
	k := edgeKey{n.ID, i}
	moving := st.drag != nil && st.drag.has(k)
	hot := e.AnimateTo("hot", b2f(slices.Contains(st.shown, k)), ui.Animation{Duration: 160 * time.Millisecond, Ease: ui.EaseOut})
	held := e.AnimateTo("held", b2f(moving), springSqueeze)
	if hot < 0.01 {
		return
	}
	color := cl.Muted.Alpha(0.5 * hot)
	e.Draw(func(p *ui.Painter, r ui.Rect) {
		long := (14 + 18*held) * hot
		if vertical {
			p.Fill(ui.Rect{X: r.X + r.W/2 - 1, Y: r.Y + r.H/2 - long/2, W: 2, H: long}, color, 1)
		} else {
			p.Fill(ui.Rect{X: r.X + r.W/2 - long/2, Y: r.Y + r.H/2 - 1, W: long, H: 2}, color, 1)
		}
	})
}

// handle makes e drag the edges seeds, as by: pressed, it takes the drag
// (with Option held, the seeds alone); while held, the vertical edges follow
// the pointer across and the others along; let go, the layout is saved.
func (d *dockView) handle(e ui.Element, by any, seeds []dockEdge) {
	st := d.st
	dx, dy, pressed := e.Dragged()
	mine := st.drag != nil && st.drag.by == by
	switch {
	case pressed && st.drag == nil && st.corner == nil && st.fold == nil:
		// the edges lined up are found once the frame built them all
		alone := d.c.Modifiers()&ui.Alt != 0
		st.drag = &dockDrag{by: by, seeds: seeds, edges: seeds, alone: alone, finding: true}
		mine = true
	case !pressed && mine:
		st.drag = nil
		if d.o.OnChange != nil {
			d.o.OnChange()
		}
		return
	}
	if mine {
		for _, le := range st.drag.edges {
			if le.vertical && dx != 0 {
				d.move(le, dx)
			} else if !le.vertical && dy != 0 {
				d.move(le, dy)
			}
		}
	}
	if st.drag == nil && e.Hovered() {
		d.hover = append(d.hover, seeds...)
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
// both sides.
func (d *dockView) move(e dockEdge, by float32) {
	n := d.l.Find(e.split)
	if n == nil || e.i+1 >= len(n.Sizes) || e.length <= 0 {
		return
	}
	a, b := n.Sizes[e.i], n.Sizes[e.i+1]
	least := min(d.o.MinSize/e.length, (a+b)/2)
	na := min(max(a+by/e.length, least), a+b-least)
	n.Sizes[e.i], n.Sizes[e.i+1] = na, a+b-na
}
