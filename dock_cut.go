package mizan

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// Cutting an edge with Option, and settling an edge onto another.
//
// The tree makes an edge between two splits one edge, all along: in a row
// of two columns whose edges line up, the edge between the columns runs
// the whole height. Option-dragged, it moves alone, so it is cut where the
// edges across it meet it from both sides: the two columns become a column
// of rows, one per part, and the part under the pointer moves. The layout
// shows the same before and after; only what moves changes.
//
// An edge dragged settles onto an edge it would line up with (one whose
// end meets its end, along the same line) within snapIn, holds there until
// the pointer is snapOut away, and a light runs along it as it lands, as
// the split line does.

// dockSegment is the part of an edge, along it, from lo to hi in the
// window, that Option would cut off.
type dockSegment struct {
	key    edgeKey
	lo, hi float32
}

// dockCutPlan is how to cut the edge between children i and i+1 of split
// s: pairs are the edges of the two that line up (an index in each), k
// the part the pointer is in, from lo to hi along the edge.
type dockCutPlan struct {
	s      string
	i      int
	pairs  [][2]int
	k      int
	lo, hi float32
}

// spanOf is where e runs along its line, in the window.
func spanOf(e dockEdge) (float32, float32) {
	if e.vertical {
		return e.rect.Y, e.rect.Y + e.rect.H
	}
	return e.rect.X, e.rect.X + e.rect.W
}

// across is where e is across its line, in the window.
func across(e dockEdge) float32 {
	if e.vertical {
		return e.rect.X + e.rect.W/2
	}
	return e.rect.Y + e.rect.H/2
}

// cutPlan finds how Option would cut edge e at the pointer p: when the
// splits on both sides of it run across it and some of their edges line
// up, it is cut there.
func (d *dockView) cutPlan(e dockEdge, p ui.Point) (dockCutPlan, bool) {
	s := d.l.Find(e.split)
	if s == nil || e.i+1 >= len(s.Children) || e.rect.W == 0 && e.rect.H == 0 {
		return dockCutPlan{}, false
	}
	a, b := s.Children[e.i], s.Children[e.i+1]
	if a.IsArea() || b.IsArea() || a.Dir == s.Dir || b.Dir == s.Dir {
		return dockCutPlan{}, false
	}
	// the edges of each side, where they are along e
	at := func(n *Node) map[int]float32 {
		out := map[int]float32{}
		for _, o := range d.st.edges {
			if o.split == n.ID {
				out[o.i] = across(o)
			}
		}
		return out
	}
	ea, eb := at(a), at(b)
	plan := dockCutPlan{s: s.ID, i: e.i}
	var cuts []float32
	for i := range len(a.Children) - 1 {
		x, ok := ea[i]
		if !ok {
			continue
		}
		for j := range len(b.Children) - 1 {
			if y, ok := eb[j]; ok && abs32(x-y) <= 1.5 {
				plan.pairs = append(plan.pairs, [2]int{i, j})
				cuts = append(cuts, x)
			}
		}
	}
	if len(plan.pairs) == 0 {
		return dockCutPlan{}, false
	}
	along := p.Y
	if !e.vertical {
		along = p.X
	}
	plan.lo, plan.hi = spanOf(e)
	for _, x := range cuts {
		if along > x {
			plan.k++
			plan.lo = x + dockGap/2
		} else {
			plan.hi = min(plan.hi, x-dockGap/2)
		}
	}
	return plan, true
}

// cut cuts the edge as plan says, and returns the part under the pointer,
// the edge to drag now.
func (d *dockView) cut(plan dockCutPlan) (dockEdge, bool) {
	s := d.l.Find(plan.s)
	if s == nil {
		return dockEdge{}, false
	}
	a := s.Children[plan.i]
	room := d.st.room[s.ID] * (s.Sizes[plan.i] + s.Sizes[plan.i+1])
	row := d.l.cutEdge(plan.s, plan.i, plan.pairs, plan.k, d.st.room[a.ID])
	if row == nil {
		return dockEdge{}, false
	}
	// what the cut made shows where the parts were, with no glide
	if s = d.l.Find(plan.s); s != nil {
		walk(s, func(n *Node) { d.st.settled[n.ID] = true })
	}
	d.c.AnimationFrame()
	return dockEdge{split: row.ID, i: 0, vertical: row.Dir == DockRow, length: room}, true
}

// cutEdge turns children i and i+1 of split id, two splits across it
// whose edges pairs line up, into one split across it of the parts between
// those edges, each a split like id of the two sides' children there; it
// returns part k. room is the first side's room for its children, which
// gives the parts their lengths.
func (l *Layout) cutEdge(id string, i int, pairs [][2]int, k int, room float32) *Node {
	s := l.Find(id)
	if s == nil || i+1 >= len(s.Children) || len(pairs) == 0 {
		return nil
	}
	a, b := s.Children[i], s.Children[i+1]
	// a side's children from one cut to the next: one alone, or a split
	part := func(n *Node, from, to int) *Node {
		if to-from == 1 {
			return n.Children[from]
		}
		return &Node{Dir: n.Dir, Children: append([]*Node(nil), n.Children[from:to]...), Sizes: append([]float32(nil), n.Sizes[from:to]...)}
	}
	var rows []*Node
	var lengths []float32
	var hit *Node
	fa, fb := 0, 0
	for p := 0; p <= len(pairs); p++ {
		ta, tb := len(a.Children), len(b.Children)
		if p < len(pairs) {
			ta, tb = pairs[p][0]+1, pairs[p][1]+1
		}
		row := &Node{Dir: s.Dir, Children: []*Node{part(a, fa, ta), part(b, fb, tb)}, Sizes: []float32{s.Sizes[i], s.Sizes[i+1]}}
		// its length: its children's, and the edges between them
		var f float32
		for _, x := range a.Sizes[fa:ta] {
			f += x
		}
		lengths = append(lengths, f*room+dockGap*float32(ta-fa-1))
		rows = append(rows, row)
		if p == k {
			hit = row
		}
		fa, fb = ta, tb
	}
	col := &Node{Dir: a.Dir, Children: rows, Sizes: lengths}
	if len(s.Children) == 2 {
		// the split itself turns, keeping its place
		s.Dir, s.Children, s.Sizes = col.Dir, col.Children, col.Sizes
	} else {
		size := s.Sizes[i] + s.Sizes[i+1]
		s.Children = append(append(append([]*Node(nil), s.Children[:i]...), col), s.Children[i+2:]...)
		s.Sizes = append(append(append([]float32(nil), s.Sizes[:i]...), size), s.Sizes[i+2:]...)
	}
	l.fix()
	return hit
}

// span sets where the edges dragged run along their line, lo to hi, and
// where the line was pressed, from seed.
func (dd *dockDrag) span(lo, hi float32, seed dockEdge) {
	dd.lo, dd.hi, dd.start = lo, hi, across(seed)
}

// slide moves the edges dragged with the pointer: they follow it across
// their line, from where it pressed (read from where it is, not by adding
// up its moves: what takes the pointer moves with the edge), settling onto
// an edge they would line up with.
func (d *dockView) slide(dd *dockDrag) {
	vertical := dd.edges[0].vertical
	way := dd.p.Y - dd.press.Y
	if vertical {
		way = dd.p.X - dd.press.X
	}
	pos := dd.start + way
	// the edges along the same line whose ends meet the dragged ones' ends
	was := dd.snapped
	if dd.snapped && abs32(dd.snapTo-pos) >= snapOut {
		dd.snapped = false
	}
	if !dd.snapped {
		best := float32(snapIn)
		for _, o := range d.st.edges {
			if o.vertical != vertical || dd.has(edgeKey{o.split, o.i}) || o.rect.W == 0 && o.rect.H == 0 {
				continue
			}
			a, b := spanOf(o)
			// across the gap between them, which rounding to pixels widens
			meets := abs32(b-dd.lo) <= dockGap+2 || abs32(a-dd.hi) <= dockGap+2
			if x := across(o); meets && abs32(x-pos) < best {
				best, dd.snapTo, dd.snapped = abs32(x-pos), x, true
			}
		}
	}
	if dd.snapped != was {
		dd.snapAt = d.c.Now()
		if dd.snapped {
			d.root.Jump("edge-spark", 0)
		}
	}
	if dd.snapped {
		pos = dd.snapTo
	}
	by := pos - dd.start - dd.moved
	if by == 0 {
		return
	}
	var moved float32
	for k, e := range dd.edges {
		if m := d.move(e, by); k == 0 {
			moved = m
		}
	}
	dd.moved += moved
}

// snapGlide is how long the sizes glide after the edge settles or leaves.
const snapGlide = 240 * time.Millisecond

// sparkCurve plays the light that runs along a line as it settles.
var sparkCurve = ui.Animation{Duration: 620 * time.Millisecond, Ease: ui.EaseInOut}

// sparkLight is the color of that light: the accent, lifted toward white
// in the dark theme, as the table's sorted light.
func (t *Tokens) sparkLight() ui.Color {
	if t.Dark {
		return Mix(t.Colors.ColorAccent400, ui.RGB(255, 255, 255), 0.35)
	}
	return t.Colors.ColorAccent500
}

// sparks draws the light running along a line (vertical, or not) at pos
// across it, at f from 0 to 1: two sparks with fading tails leave perp
// and run to from and to, a core as wide as the line in a narrow glow,
// fading as they reach the ends; shown fades it all.
func sparks(p *ui.Painter, vertical bool, pos, perp, from, to, f, shown float32, light ui.Color) {
	if f <= 0 || f >= 1 {
		return
	}
	bar := func(s0, s1, w float32) ui.Rect {
		if vertical {
			return ui.Rect{X: pos - w/2, Y: s0, W: w, H: s1 - s0}
		}
		return ui.Rect{X: s0, Y: pos - w/2, W: s1 - s0, H: w}
	}
	fade := min(1, (1-f)*2.5) * min(1, f*8) * shown
	for _, end := range []float32{from, to} {
		head := perp + (end-perp)*f
		tail := head - (end-perp)*0.35*min(f*3, 1)
		lo, hi := min(head, tail), max(head, tail)
		if hi-lo < 1 {
			continue
		}
		// the gradient goes from the tail to the head
		angle := float32(180)
		if !vertical {
			angle = 90
		}
		if end < perp {
			angle += 180
		}
		grad := func(a float32) ui.LinearGradient {
			return ui.LinearGradient{From: light.Alpha(0), To: light.Alpha(a * fade), Angle: angle, Oklab: true}
		}
		p.FillGradient(bar(lo, hi, 8), grad(0.2), 4)
		p.FillGradient(bar(lo, hi, 2), grad(1), 1)
	}
}
