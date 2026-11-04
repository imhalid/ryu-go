package mizan

import (
	"math"
	"time"

	"github.com/egoist/mygo/ui"
)

// The docking compass: while a tab is dragged, the area under the pointer
// shows five targets in its middle, a plus of raised tiles on a plate:
// the middle puts the tab in the area's stack, the arms dock it beside the
// area, on that side. Four more targets sit in the middle of the window's
// edges and dock it along the whole of that edge. Each target draws what
// it does: a small pane with the part the tab would take filled in the
// accent. The compass glides from area to area on a spring; the target
// under the pointer grows a little and fills in, and the region it would
// take shows behind, as it does when the pointer is near an area's side.

const (
	compassTile = 30 // a target's side
	compassGap  = 4  // between targets
	compassPad  = 5  // the plate around them
	// compassExt is from the compass's middle to the end of an arm
	compassExt = 1.5*compassTile + compassGap + compassPad
)

// compassAt says which targets the compass shows over area: center, then
// left, right, top, bottom.
type compassAt struct {
	area string
	at   ui.Point // its middle, in the window
	on   [5]bool
}

// compassTargets are the targets' boxes around the middle at.
func compassTargets(at ui.Point) [5]ui.Rect {
	const s, step = compassTile, compassTile + compassGap
	box := func(dx, dy float32) ui.Rect {
		return ui.Rect{X: at.X + dx*step - s/2, Y: at.Y + dy*step - s/2, W: s, H: s}
	}
	return [5]ui.Rect{box(0, 0), box(-1, 0), box(1, 0), box(0, -1), box(0, 1)}
}

// edgeTargets are the boxes of the targets on the window's edges, w its
// Dock in the window: left, right, top, bottom.
func edgeTargets(w ui.Rect) [4]ui.Rect {
	// the top one under the headers, so that it hides no tab
	const s, in, top = compassTile, 12, headerH + 10
	mx, my := w.X+w.W/2-s/2, w.Y+w.H/2-s/2
	return [4]ui.Rect{
		{X: w.X + in, Y: my, W: s, H: s},
		{X: w.X + w.W - in - s, Y: my, W: s, H: s},
		{X: mx, Y: w.Y + top, W: s, H: s},
		{X: mx, Y: w.Y + w.H - in - s, W: s, H: s},
	}
}

// sides are the zones of the compass's arms, and of the window's edges, in
// their order.
var sides = [4]zoneSide{zoneLeft, zoneRight, zoneTop, zoneBottom}

// aimCompass places the compass over the area under the pointer, and
// returns the zone of the target under it: one of the compass's, or of
// the window's edges.
func (d *dockView) aimCompass(p ui.Point) (dockZone, bool) {
	st := d.st
	dr := st.tab
	st.compass = compassAt{}
	reach := func(r ui.Rect) bool {
		const m = compassGap / 2
		return p.X >= r.X-m && p.X < r.X+r.W+m && p.Y >= r.Y-m && p.Y < r.Y+r.H+m
	}
	// the window's edges first: they sit over the areas along them
	win := d.root.LayoutBounds()
	for i, r := range edgeTargets(win) {
		if reach(r) {
			return d.edgeZone(sides[i], win), true
		}
	}
	for id, a := range st.areas {
		if !inside(a, p) {
			continue
		}
		n := d.l.Find(id)
		if n == nil || !n.IsArea() {
			return dockZone{}, false
		}
		own := id == dr.from
		body := ui.Rect{X: a.X, Y: a.Y + headerH, W: a.W, H: a.H - headerH}
		const span = 3*compassTile + 2*compassGap + 2*compassPad + 16
		if (own && len(n.Panes) == 1) || body.W < span || body.H < span {
			return dockZone{}, false
		}
		wide, tall := a.W >= 2*d.o.MinSize+dockGap, a.H >= 2*d.o.MinSize+dockGap
		c := compassAt{area: id, at: ui.Point{X: body.X + body.W/2, Y: body.Y + body.H/2}, on: [5]bool{!own, wide, wide, tall, tall}}
		// it keeps clear of the window's edge targets, moving away from them
		for i, r := range edgeTargets(win) {
			const clear = compassExt + 6
			if abs32(r.X+r.W/2-c.at.X) >= clear+r.W/2 || abs32(r.Y+r.H/2-c.at.Y) >= clear+r.H/2 {
				continue
			}
			switch sides[i] {
			case zoneLeft:
				c.at.X = r.X + r.W + clear
			case zoneRight:
				c.at.X = r.X - clear
			case zoneTop:
				c.at.Y = r.Y + r.H + clear
			case zoneBottom:
				c.at.Y = r.Y - clear
			}
		}
		st.compass = c
		inner := ui.Rect{X: a.X + 3, Y: a.Y + 3, W: a.W - 6, H: a.H - 6}
		for i, r := range compassTargets(c.at) {
			if !c.on[i] || !reach(r) {
				continue
			}
			if i == 0 {
				return dockZone{area: id, side: zoneStack, index: len(n.Panes), rect: inner, target: true}, true
			}
			return dockZone{area: id, side: sides[i-1], rect: half(inner, sides[i-1]), target: true}, true
		}
		return dockZone{}, false
	}
	return dockZone{}, false
}

// half is the half of r on side.
func half(r ui.Rect, side zoneSide) ui.Rect {
	switch side {
	case zoneLeft:
		r.W /= 2
	case zoneRight:
		r.X, r.W = r.X+r.W/2, r.W/2
	case zoneTop:
		r.H /= 2
	case zoneBottom:
		r.Y, r.H = r.Y+r.H/2, r.H/2
	}
	return r
}

// edgeLength is the length of a new area docked along an edge of the
// window w, across it.
func (d *dockView) edgeLength(side zoneSide, w ui.Rect) float32 {
	length := w.W
	if side == zoneTop || side == zoneBottom {
		length = w.H
	}
	return min(max(length*0.25, d.o.MinSize), length/2)
}

// edgeZone is the zone of the window's edge on side: a strip along it.
func (d *dockView) edgeZone(side zoneSide, w ui.Rect) dockZone {
	in := ui.Rect{X: w.X + 3, Y: w.Y + 3, W: w.W - 6, H: w.H - 6}
	l := d.edgeLength(side, w) - 3
	r := in
	switch side {
	case zoneLeft:
		r.W = l
	case zoneRight:
		r.X, r.W = in.X+in.W-l, l
	case zoneTop:
		r.H = l
	case zoneBottom:
		r.Y, r.H = in.Y+in.H-l, l
	}
	return dockZone{area: d.w.Root.ID, side: side, rect: r, target: true, edge: true}
}

// compass builds the compass and the window's edge targets over the
// window, above the tab's copy, so that it doesn't hide them.
func (d *dockView) compass() {
	c, st, t := d.c, d.st, d.t
	cl := t.Colors
	dr := st.tab
	dragging := dr != nil && !dr.let
	cp := st.compass
	ui.Overlay(c, func() {
		host := ui.Box(c.Key("compass")).Absolute().Left(0).Top(0).Width(0).Height(0).PassThrough()
		on := dragging && cp.area != ""
		// in on a spring, out quicker on a curve
		var show float32
		if on {
			show = host.AnimateTo("show", 1, *springPop)
		} else {
			show = host.AnimateTo("show", 0, ui.Animation{Duration: 140 * time.Millisecond, Ease: ui.EaseOut})
		}
		// from one area to the next it glides; it appears where it is
		last := ui.Local(host, "at", func() compassAt { return compassAt{} })
		if on {
			*last = cp
		}
		at := last.at
		if show < 0.01 {
			at = ui.Point{X: host.Jump("x", at.X), Y: host.Jump("y", at.Y)}
		} else {
			at = ui.Point{X: host.AnimateTo("x", at.X, *springPop), Y: host.AnimateTo("y", at.Y, *springPop)}
		}
		edges := host.AnimateTo("edges", b2f(dragging), ui.Animation{Duration: 160 * time.Millisecond, Ease: ui.EaseOut})
		var z dockZone
		if dr != nil {
			z = dr.zone
		}
		if show > 0.01 {
			d.plate(at, last.on, show, z)
		}
		if edges > 0.01 {
			win := d.root.LayoutBounds()
			for i, r := range edgeTargets(win) {
				hot := dragging && z.edge && z.side == sides[i]
				d.target(c.Key(edgeKey{"edge", i}), r, sides[i], true, true, hot, edges, cl)
			}
		}
	})
}

// plate builds the compass around at: the plate, a plus, and its targets.
func (d *dockView) plate(at ui.Point, on [5]bool, show float32, z dockZone) {
	c, t := d.c, d.t
	cl := t.Colors
	const ext = compassExt
	tileR := min(t.R(6), 8)
	plateR := tileR + compassPad
	s := 0.92 + 0.08*show
	plate := ui.Box(c.Key("plate")).Absolute().Left(at.X-ext).Top(at.Y-ext).Width(2*ext).Height(2*ext).
		PassThrough().Opacity(min(show, 1)).Scale(s, s)
	shadow := t.Shadows.ShadowPop
	plate.Draw(func(p *ui.Painter, r ui.Rect) {
		cx, cy := r.X+r.W/2, r.Y+r.H/2
		arm := float32(compassTile/2 + compassPad)
		bars := [2]ui.Rect{
			{X: r.X, Y: cy - arm, W: r.W, H: 2 * arm},
			{X: cx - arm, Y: r.Y, W: 2 * arm, H: r.H},
		}
		for _, l := range shadow {
			if l.Inset {
				continue
			}
			for _, b := range bars {
				p.Shadow(b, plateR, l.X, l.Y, l.Blur, l.Spread, l.Color)
			}
		}
		path := plusPath(cx, cy, arm, ext, plateR, 4)
		p.FillPath(path, over(cl.Pop, cl.Card))
		p.StrokePath(path, 1, cl.Hairline)
	})
	plate.Children(func() {
		targets := compassTargets(ui.Point{X: ext, Y: ext})
		for i, r := range targets {
			side := zoneStack
			if i > 0 {
				side = sides[i-1]
			}
			hot := z.target && !z.edge && z.side == side
			d.target(c.Key(edgeKey{"compass", i}), r, side, false, on[i], hot, 1, cl)
		}
	})
}

// target builds a target in r (in its parent): a raised tile showing a
// small pane with the part side takes filled in the accent (along the
// window's edge, edge), faded when off; under the pointer, hot, it grows
// a little and fills in.
func (d *dockView) target(key *ui.Context, r ui.Rect, side zoneSide, edge, on, hot bool, shown float32, cl Colors) {
	t := d.t
	e := ui.Box(key).Absolute().Left(r.X).Top(r.Y).Width(r.W).Height(r.H).PassThrough().Radius(min(t.R(6), 8))
	t.raised(e, t.controlGradient(cl.ControlTop, cl.ControlBottom), cl.ControlBorder, t.Shadows.ControlShadow, raisedState{hovered: hot})
	grow := e.AnimateTo("hot", b2f(hot), springSqueeze)
	s := 1 + 0.08*grow
	e.Scale(s, s)
	if !on {
		shown *= 0.35
	}
	e.Opacity(shown)
	fill := Mix(cl.ColorAccent500.Alpha(0.35), cl.ColorAccent500, grow)
	if !on {
		fill = cl.Muted.Alpha(0.3)
	}
	frame := Mix(cl.Muted, cl.ColorAccent500, grow)
	e.Draw(func(p *ui.Painter, b ui.Rect) {
		// the pane: 16 by 13, its header a line under the top
		pane := ui.Rect{X: b.X + b.W/2 - 8, Y: b.Y + b.H/2 - 6.5, W: 16, H: 13}
		part := pane
		switch {
		case side == zoneStack:
		case edge:
			part = edgeStrip(pane, side)
		default:
			part = half(pane, side)
		}
		p.Fill(part, fill, 1.5)
		p.Stroke(pane, frame, 2, 1)
		if edge {
			// an arrow toward the edge, beside the strip
			arrow(p, pane, side, frame)
		}
	})
}

// edgeStrip is the third of r along side.
func edgeStrip(r ui.Rect, side zoneSide) ui.Rect {
	switch side {
	case zoneLeft:
		r.W /= 3
	case zoneRight:
		r.X, r.W = r.X+r.W*2/3, r.W/3
	case zoneTop:
		r.H /= 3
	case zoneBottom:
		r.Y, r.H = r.Y+r.H*2/3, r.H/3
	}
	return r
}

// arrow draws a small chevron in r's middle, pointing to side.
func arrow(p *ui.Painter, r ui.Rect, side zoneSide, color ui.Color) {
	cx, cy := r.X+r.W/2+r.W/6, r.Y+r.H/2
	var dx, dy float32
	switch side {
	case zoneLeft:
		dx = -1
	case zoneRight:
		dx, cx = 1, r.X+r.W/2-r.W/6
	case zoneTop:
		dy, cx, cy = -1, r.X+r.W/2, r.Y+r.H/2+r.H/6
	case zoneBottom:
		dy, cx, cy = 1, r.X+r.W/2, r.Y+r.H/2-r.H/6
	}
	const a = 2.5
	var path ui.Path
	tx, ty := cx+dx*a/2, cy+dy*a/2
	path.MoveTo(tx-dx*a-dy*a, ty-dy*a-dx*a).LineTo(tx, ty).LineTo(tx-dx*a+dy*a, ty-dy*a+dx*a)
	p.StrokePath(&path, 1.2, color)
}

// plusPath is a plus around (cx, cy): arms arm from its middle to their
// sides and ext to their ends, its outer corners rounded by out and its
// inner ones by in.
func plusPath(cx, cy, arm, ext, out, in float32) *ui.Path {
	pts := [12]ui.Point{
		{X: -arm, Y: -ext}, {X: arm, Y: -ext}, {X: arm, Y: -arm},
		{X: ext, Y: -arm}, {X: ext, Y: arm}, {X: arm, Y: arm},
		{X: arm, Y: ext}, {X: -arm, Y: ext}, {X: -arm, Y: arm},
		{X: -ext, Y: arm}, {X: -ext, Y: -arm}, {X: -arm, Y: -arm},
	}
	toward := func(a, b ui.Point, d float32) ui.Point {
		l := float32(math.Hypot(float64(b.X-a.X), float64(b.Y-a.Y)))
		return ui.Point{X: a.X + (b.X-a.X)*d/l, Y: a.Y + (b.Y-a.Y)*d/l}
	}
	var path ui.Path
	for i := range pts {
		prev, cur, next := pts[(i+11)%12], pts[i], pts[(i+1)%12]
		// every third corner, from the third, is an inner one
		r := out
		if i%3 == 2 {
			r = in
		}
		a, b := toward(cur, prev, r), toward(cur, next, r)
		if i == 0 {
			path.MoveTo(cx+a.X, cy+a.Y)
		} else {
			path.LineTo(cx+a.X, cy+a.Y)
		}
		path.QuadTo(cx+cur.X, cy+cur.Y, cx+b.X, cy+b.Y)
	}
	path.Close()
	return &path
}
