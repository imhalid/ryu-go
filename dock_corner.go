package mizan

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// Splitting and joining from an area's corners, as in Blender. Dragged
// inward, a corner splits its area: a line grows out from the pointer
// across the area, follows it, and settles onto a lined-up edge or the
// middle when near one; let go, the area splits there, the new area (on the
// corner's side, showing a copy of the pane) fading in. Dragged into a
// neighbor sharing a whole edge, it joins them: the neighbor darkens and an
// arrow over it points the way; let go, the area grows over it on a spring
// as it folds away.

// cornerSize is the side of the square in an area's corner that takes the
// pointer.
const cornerSize = 14

// springGuide moves the split line onto a lined-up edge, and grows it.
var springGuide = ui.Animation{Spring: ui.SpringOptions{Duration: 200 * time.Millisecond, Bounce: 0.1}}

// cornerKey keys corner k (0 top left, 1 top right, 2 bottom left, 3 bottom
// right) of an area.
type cornerKey struct {
	area string
	k    int
}

// dockCorner is a corner being dragged.
type dockCorner struct {
	by        cornerKey
	area      string
	left, top bool     // which corner
	start, p  ui.Point // where it was pressed, and where the pointer is, in the window
	decided   bool     // the split's axis is chosen
	vertical  bool     // the split line is vertical: the area splits side by side
	let       bool     // let go in this frame
}

type guideMode uint8

const (
	guideNone  guideMode = iota
	guideSplit           // a split line over the area
	guideJoin            // the neighbor to be taken in, darkened, with an arrow
	guideFold            // the neighbor folding away after a join
)

// dockGuide is what the previews show, kept after a gesture ends so that
// they fade where they were.
type dockGuide struct {
	mode     guideMode
	area     ui.Rect // the area split
	vertical bool
	line     float32 // where the pointer puts the split line, across its axis
	snapTo   float32 // the edge or middle it settles onto
	snapped  bool
	perp     float32 // the pointer along the line: it grows out from there
	turn     int     // counts the times the axis was chosen, for the growth
	target   string  // the area to be taken in
	rect     ui.Rect // and where it is
	dir      ui.Point
}

// dockFold is a join playing: gone shrinks to nothing as keep grows over
// it; then the layout changes.
type dockFold struct {
	split, keep, gone string
	start             time.Time
	done              bool
}

// corner builds corner k of area n.
func (d *dockView) corner(n *Node, k int) {
	c, st, t := d.c, d.st, d.t
	left, top := k%2 == 0, k < 2
	key := cornerKey{n.ID, k}
	e := ui.Box(c.Key(key)).Absolute().Width(cornerSize).Height(cornerSize).Cursor(ui.CursorCrosshair)
	if left {
		e.Left(0)
	} else {
		e.Right(0)
	}
	if top {
		e.Top(0)
	} else {
		e.Bottom(0)
	}
	_, _, pressed := e.Dragged()
	mine := st.corner != nil && st.corner.by == key
	at := func() ui.Point {
		b := e.LayoutBounds()
		px, py, _ := e.PointerPosition()
		return ui.Point{X: b.X + px, Y: b.Y + py}
	}
	switch {
	case pressed && st.corner == nil && st.drag == nil && st.fold == nil:
		p := at()
		st.corner = &dockCorner{by: key, area: n.ID, left: left, top: top, start: p, p: p}
		mine = true
	case pressed && mine:
		st.corner.p = at()
	case !pressed && mine:
		st.corner.let = true
	}

	// under the pointer, a short bracket draws itself along the area's
	// corner, concentric with it, from its middle out
	hot := e.AnimateTo("hot", b2f(e.Hovered() || mine), ui.Animation{Duration: 180 * time.Millisecond, Ease: ui.EaseOut})
	if hot < 0.01 {
		return
	}
	R := t.R(RadiusControl)
	color := t.Colors.Muted.Alpha(0.8 * hot)
	e.Draw(func(p *ui.Painter, r ui.Rect) {
		const inset, leg, kappa = 3, 5, 0.5523
		cx, cy, sx, sy := r.X, r.Y, float32(1), float32(1)
		if !left {
			cx, sx = r.X+r.W, -1
		}
		if !top {
			cy, sy = r.Y+r.H, -1
		}
		rr := max(R, inset)
		q := (rr - inset) * kappa
		at := func(x, y float32) (float32, float32) { return cx + sx*x, cy + sy*y }
		var path ui.Path
		path.MoveTo(at(inset, rr+leg))
		path.LineTo(at(inset, rr))
		x1, y1 := at(inset, rr-q)
		x2, y2 := at(rr-q, inset)
		x3, y3 := at(rr, inset)
		path.CubeTo(x1, y1, x2, y2, x3, y3)
		path.LineTo(at(rr+leg, inset))
		path.Trim(0.5-0.5*hot, 0.5+0.5*hot)
		p.StrokePath(&path, 1.5, color)
	})
}

// aim finds what the corner dragged would do where the pointer is: split
// its area, join a neighbor, or nothing.
func (d *dockView) aim() {
	k, g := d.st.corner, &d.st.guide
	a, ok := d.st.areas[k.area]
	g.mode = guideNone
	if !ok {
		return
	}
	p := k.p
	if inside(a, p) {
		dx, dy := abs32(p.X-k.start.X), abs32(p.Y-k.start.Y)
		// the axis is the one the pointer leaves the corner by; it turns only
		// when the other one clearly wins
		switch {
		case !k.decided && max(dx, dy) < 8:
			return
		case !k.decided:
			k.decided, k.vertical = true, dx >= dy
			g.turn++
		case k.vertical && dy > 1.6*dx, !k.vertical && dx > 1.6*dy:
			k.vertical = !k.vertical
			g.turn++
		}
		lo, hi, pos, perp := a.X, a.X+a.W, p.X, p.Y
		if !k.vertical {
			lo, hi, pos, perp = a.Y, a.Y+a.H, p.Y, p.X
		}
		least := d.o.MinSize + dockGap/2
		if hi-lo < 2*least {
			return // too small to split
		}
		pos = min(max(pos, lo+least), hi-least)
		// it settles onto the middle, or onto an edge it would line up with
		g.snapped = false
		best := float32(8)
		try := func(at float32) {
			if at >= lo+least && at <= hi-least && abs32(at-pos) < best {
				best, g.snapTo, g.snapped = abs32(at-pos), at, true
			}
		}
		try((lo + hi) / 2)
		for _, e := range d.st.edges {
			if e.vertical == k.vertical {
				if e.vertical {
					try(e.rect.X + e.rect.W/2)
				} else {
					try(e.rect.Y + e.rect.H/2)
				}
			}
		}
		g.mode, g.area, g.vertical, g.line, g.perp = guideSplit, a, k.vertical, pos, perp
		return
	}
	for id, r := range d.st.areas {
		if id == k.area || !inside(r, p) || !d.l.joinable(k.area, id) {
			continue
		}
		g.mode, g.target, g.rect = guideJoin, id, r
		switch {
		case r.X >= a.X+a.W:
			g.dir = ui.Point{X: 1}
		case r.X+r.W <= a.X:
			g.dir = ui.Point{X: -1}
		case r.Y >= a.Y+a.H:
			g.dir = ui.Point{Y: 1}
		default:
			g.dir = ui.Point{Y: -1}
		}
		return
	}
}

// let does what the corner let go asks: splits, or starts a join's fold.
func (d *dockView) let() {
	st := d.st
	k, g := st.corner, &st.guide
	st.corner = nil
	switch g.mode {
	case guideSplit:
		pos := g.line
		if g.snapped {
			pos = g.snapTo
		}
		a := g.area
		dir, lo, hi, first := DockRow, a.X, a.X+a.W, k.left
		if !g.vertical {
			dir, lo, hi, first = DockColumn, a.Y, a.Y+a.H, k.top
		}
		// the new area is on the corner's side; the edge's middle is where
		// the line was
		before, after := pos-lo-dockGap/2, hi-pos-dockGap/2
		keep, made := after, before
		if !first {
			keep, made = before, after
		}
		var room float32
		if _, p, _ := d.l.where(k.area); p != nil {
			room = st.room[p.ID]
		}
		if n := d.l.splitArea(k.area, dir, first, keep, made, room); n != nil {
			st.born[n.ID] = true
			d.changed()
		}
		g.line, g.snapped = pos, false
		g.snapTo = pos
		g.mode = guideNone
	case guideJoin:
		if _, p, _ := d.l.where(g.target); p != nil {
			st.fold = &dockFold{split: p.ID, keep: k.area, gone: g.target, start: d.c.Now()}
			g.mode = guideFold
		}
	}
}

func (d *dockView) changed() {
	if d.o.OnChange != nil {
		d.o.OnChange()
	}
}

func inside(r ui.Rect, p ui.Point) bool {
	return p.X >= r.X && p.X < r.X+r.W && p.Y >= r.Y && p.Y < r.Y+r.H
}

// guide builds the previews over the areas: the split line, and the area
// to be taken in, darkened, with its arrow.
func (d *dockView) guide() {
	c, st, t := d.c, d.st, d.t
	cl := t.Colors
	g := st.guide
	at := d.root.LayoutBounds()
	e := ui.Box(c.Key("guide")).Absolute().Left(0).Top(0).Right(0).Bottom(0).PassThrough()
	split := e.AnimateTo("split", b2f(g.mode == guideSplit), ui.Animation{Duration: 140 * time.Millisecond, Ease: ui.EaseOut})
	snap := e.AnimateTo("snap", b2f(g.snapped), springGuide)
	turn := ui.Local(e, "turn", func() int { return 0 })
	if *turn != g.turn {
		*turn = g.turn
		e.Jump("grow", 0)
	}
	grow := e.AnimateTo("grow", 1, springGuide)
	// let go, the area folding draws its own darkening (folding): here it
	// would be a frame behind it
	if g.mode == guideFold {
		e.Jump("join", 0)
		e.Jump("arrow", 0)
	}
	join := e.AnimateTo("join", b2f(g.mode == guideJoin), ui.Animation{Duration: 160 * time.Millisecond, Ease: ui.EaseOut})
	arrow := e.AnimateTo("arrow", b2f(g.mode == guideJoin), *springPop)
	// the darkened area moves from one neighbor to the next on a spring;
	// it starts where the area is
	r := g.rect
	if join < 0.01 {
		r = ui.Rect{X: e.Jump("rx", r.X), Y: e.Jump("ry", r.Y), W: e.Jump("rw", r.W), H: e.Jump("rh", r.H)}
	} else {
		r = ui.Rect{X: e.AnimateTo("rx", r.X, *springPop), Y: e.AnimateTo("ry", r.Y, *springPop), W: e.AnimateTo("rw", r.W, *springPop), H: e.AnimateTo("rh", r.H, *springPop)}
	}
	if split < 0.01 && join < 0.01 {
		return
	}
	R := t.R(RadiusControl)
	e.Draw(func(p *ui.Painter, b ui.Rect) {
		ox, oy := b.X-at.X, b.Y-at.Y
		if split > 0.01 {
			pos := g.line + (g.snapTo-g.line)*snap
			a, color := g.area, cl.Text.Alpha(0.45*split)
			if g.vertical {
				from, to := a.Y+6, a.Y+a.H-6
				y0, y1 := g.perp-(g.perp-from)*grow, g.perp+(to-g.perp)*grow
				p.Fill(ui.Rect{X: pos - 1 + ox, Y: y0 + oy, W: 2, H: y1 - y0}, color, 1)
			} else {
				from, to := a.X+6, a.X+a.W-6
				x0, x1 := g.perp-(g.perp-from)*grow, g.perp+(to-g.perp)*grow
				p.Fill(ui.Rect{X: x0 + ox, Y: pos - 1 + oy, W: x1 - x0, H: 2}, color, 1)
			}
		}
		if join < 0.01 || r.W < 1 || r.H < 1 {
			return
		}
		p.Fill(ui.Rect{X: r.X + ox, Y: r.Y + oy, W: r.W, H: r.H}, cl.Scrim.Alpha(join), R)
		if arrow < 0.01 {
			return
		}
		// it comes in from behind, growing
		joinArrow(p, ui.Rect{X: r.X + ox, Y: r.Y + oy, W: r.W, H: r.H}, g.dir, arrow, cl.Text.Alpha(0.75*join*min(arrow, 1)))
	})
}

// joinArrow draws the arrow over an area to be taken in, pointing the way
// its neighbor grows, at grown (0 to 1: smaller, and back from its place).
func joinArrow(p *ui.Painter, r ui.Rect, dir ui.Point, grown float32, color ui.Color) {
	ux, uy := dir.X, dir.Y
	long := min(max(min(r.W, r.H)*0.3, 28), 88) * (0.7 + 0.3*grown)
	back := (1 - grown) * -10
	mx, my := r.X+r.W/2+ux*back, r.Y+r.H/2+uy*back
	tx, ty := mx+ux*long/2, my+uy*long/2
	w := min(long*0.28, 14)
	var path ui.Path
	path.MoveTo(mx-ux*long/2, my-uy*long/2).LineTo(tx, ty)
	path.MoveTo(tx-ux*w-uy*w, ty-uy*w+ux*w).LineTo(tx, ty).LineTo(tx-ux*w+uy*w, ty-uy*w-ux*w)
	p.StrokePath(&path, 3, color)
}

// folding darkens area a as a join folds it away, as the preview did; the
// arrow fades.
func (d *dockView) folding(a ui.Element) {
	cl, g := d.t.Colors, d.st.guide
	if seen := ui.Local(a, "folding", func() bool { return false }); !*seen {
		*seen = true
		a.Jump("arrow", 1)
	}
	arrow := a.AnimateTo("arrow", 0, ui.Animation{Duration: 140 * time.Millisecond, Ease: ui.EaseOut})
	R := d.t.R(RadiusControl)
	a.DrawOver(func(p *ui.Painter, r ui.Rect) {
		p.Fill(r, cl.Scrim, R)
		if arrow > 0.01 {
			joinArrow(p, r, g.dir, 1, cl.Text.Alpha(0.75*arrow))
		}
	})
}
