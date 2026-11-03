package mizan

import (
	"math"
	"slices"
	"time"

	"github.com/egoist/mygo/ui"
)

// Drag and drop speaks in two parts: what is dragged is the raised skin (it
// can be held), where it is dropped is an inset well. A card picked up
// lifts off its place and follows the pointer; the cards around where it
// would go slide apart to make room for it, and close the room it left; let
// go, it flies into its place. A column taking it glows as if a cloud's
// shadow drifted across it, and a light runs up both its sides as it lands.

var (
	iconUpload = Svg(16, 20, "M8 10.5V2.5M5 5.5l3-3 3 3M2.5 10v2A1.5 1.5 0 004 13.5h8a1.5 1.5 0 001.5-1.5v-2")
	iconFile   = Svg(16, 14, "M9 1.5H4.5a1.5 1.5 0 0 0-1.5 1.5v10a1.5 1.5 0 0 0 1.5 1.5h7a1.5 1.5 0 0 0 1.5-1.5V5.5zM9 1.5v4h4")
)

// The springs of a drag: cards making room settle without a wobble, the
// card landing with a hint of one.
var (
	springShift = ui.Animation{Spring: ui.SpringOptions{Duration: 320 * time.Millisecond, Bounce: 0.06}}
	springLand  = ui.Animation{Spring: ui.SpringOptions{Duration: 380 * time.Millisecond, Bounce: 0.16}}
	springLift  = ui.Animation{Spring: ui.SpringOptions{Duration: 260 * time.Millisecond, Bounce: 0.2}}
)

// dragSession is the drag in progress: there is one pointer, so one drag
// at a time, which a card starts and the lists it goes over read.
type dragSession struct {
	active bool
	value  any
	w, h   float32  // the card's size
	grab   ui.Point // where the pointer holds the card, from its top left
	pos    ui.Point // the pointer, in window DIPs
	vx     float32  // its sideways speed, which the card leans with
	over   *dropList
	// a drop the next frame applies, before any list builds, so that every
	// list builds the new order at once; gen counts the drops applied
	pending func()
	gen     int
	// the card let go, which goes from where the copy was (land) to its
	// place, once the frame after the drop laid it out (after settled)
	landing any
	land    ui.Point
	tilt    float32
	settled time.Time
}

var drag dragSession

// building is the stack of lists being built, which a card joins.
var building []*dropList

// dropList is a list cards are dragged in and out of, as a board's column
// or a Sortable: its cards this frame, in order, and where the card
// dragged over it would go.
type dropList struct {
	cards   []dropCard
	accept  func(any) bool
	handler func(v any, index int) any
	gen     int
	at      int // the index the card over it would take, -1 for none
	heights []float32
	flash   bool // a card landed: the light runs up its sides
}

type dropCard struct {
	e     ui.Element
	value any
}

// settleDrop applies the drop let go in an earlier frame: lists call it
// before they build. Never in the frame it was let go in, where the lists
// built before the card already built the old order: the first list of the
// next frame applies it, so that every list builds the new order together.
func settleDrop(c *ui.Context) {
	if drag.pending == nil || !c.Now().After(drag.settled) {
		return
	}
	p := drag.pending
	drag.pending = nil
	drag.gen++
	drag.settled = c.Now()
	p()
	c.AnimationFrame()
}

// dragSource makes e, a card in the list being built, picked up by a drag
// of the pointer: look styles a card and builds its content, on e and on
// the copy following the pointer (lifted). value is what it carries,
// comparable.
func (t *Tokens) dragSource(c *ui.Context, e ui.Element, value any, look func(e ui.Element, lifted bool)) {
	var list *dropList
	if n := len(building); n > 0 {
		list = building[n-1]
		list.cards = append(list.cards, dropCard{e, value})
	}
	moved := ui.Local(e, "moved", func() float32 { return 0 })
	dx, dy, pressed := e.Dragged()
	mine := drag.active && drag.value == value
	switch {
	case pressed && mine:
		// where the pointer is, from the card's place, which stays put
		b := e.LayoutBounds()
		px, py, _ := e.PointerPosition()
		drag.pos = ui.Point{X: b.X + px, Y: b.Y + py}
		drag.vx, _ = e.DragVelocity()
	case pressed && !drag.active:
		// a press becomes a drag once it moved a few DIPs, not a click
		if *moved += abs32(dx) + abs32(dy); *moved > 4 {
			b := e.LayoutBounds()
			px, py, _ := e.PointerPosition()
			pp := ui.Point{X: px, Y: py}
			drag = dragSession{active: true, value: value, w: b.W, h: b.H, grab: pp,
				pos: ui.Point{X: b.X + pp.X, Y: b.Y + pp.Y}, gen: drag.gen}
			mine = true
		}
	case !pressed:
		*moved = 0
	}
	if mine && !pressed {
		// let go: on a list, the drop applies in the next frame
		drag.active = false
		drag.landing, drag.land, drag.settled = value, ui.Point{X: drag.pos.X - drag.grab.X, Y: drag.pos.Y - drag.grab.Y}, c.Now()
		if l := drag.over; l != nil && l.handler != nil {
			v, at, h := value, l.at, l.handler
			l.flash = true
			drag.pending = func() { drag.landing = h(v, at) }
		}
		mine = false
		c.AnimationFrame()
	}

	// the card on its way from where it was let go to its place
	landing := !drag.active && drag.landing == value
	if landing && c.Now().After(drag.settled) && drag.pending == nil {
		b := e.LayoutBounds()
		e.Jump("lx", drag.land.X-b.X)
		e.Jump("ly", drag.land.Y-b.Y)
		e.Jump("lift", 1)
		e.Jump("tilt", drag.tilt)
		drag.landing = nil
		landing = false
	}
	lift := e.AnimateTo("lift", 0, springLand)
	x, y := e.AnimateTo("lx", 0, springLand), e.AnimateTo("ly", 0, springLand)
	s := 1 + 0.03*lift
	tilt := e.AnimateTo("tilt", 0, springLand)
	// the card inside its place: the list moves the place to make room,
	// the card moves within it as it lands
	e.Children(func() {
		in := ui.Column(c).FillWidth().MinWidth(0).Translate(x, y).Scale(s, s).Rotate(tilt)
		look(in, lift > 0.01)
	})
	if landing {
		c.AnimationFrame()
	}
	if mine || landing {
		// its place stays, empty, while a copy is held
		e.Opacity(0).Cursor(ui.CursorGrabbing)
		t.ghost(c, look)
	}
}

// ghost is the copy of the card dragged, lifted off the page: it follows
// the pointer, a little larger, leaning into the motion, and stays where
// it was let go until the card takes over.
func (t *Tokens) ghost(c *ui.Context, look func(e ui.Element, lifted bool)) {
	ui.Overlay(c, func() {
		at := drag.land
		if drag.active {
			at = ui.Point{X: drag.pos.X - drag.grab.X, Y: drag.pos.Y - drag.grab.Y}
		}
		g := ui.Column(c).Absolute().Left(at.X).Top(at.Y).Width(drag.w).PassThrough()
		picked := ui.Local(g, "picked", func() bool { return false })
		if !*picked {
			*picked = true
			g.Jump("lift", 0)
		}
		lift := g.AnimateTo("lift", 1, springLift)
		tilt := g.AnimateTo("tilt", min(max(drag.vx*0.0016, -2.5), 2.5), springTilt)
		drag.tilt = tilt
		s := 1 + 0.03*lift
		g.Scale(s, s).Rotate(tilt).Cursor(ui.CursorGrabbing)
		look(g, true)
	})
}

// arrange moves the cards of l, built in area, to make room for the card
// dragged over it and to close the room it left, gap DIPs apart, and
// returns how much taller the list grows for a card coming from elsewhere.
func (l *dropList) arrange(c *ui.Context, area ui.Element, gap float32) float32 {
	settle := l.gen != drag.gen
	l.gen = drag.gen
	held := drag.active || drag.pending != nil
	d := -1 // the dragged card, when it is one of these
	if held {
		for i, cd := range l.cards {
			if cd.value == drag.value {
				d = i
			}
		}
	}
	room := drag.h + gap
	if drag.active {
		b := area.LayoutBounds()
		inside := drag.pos.X >= b.X && drag.pos.X < b.X+b.W && drag.pos.Y >= b.Y && drag.pos.Y < b.Y+b.H &&
			(l.accept == nil || l.accept(drag.value))
		l.at = -1
		switch {
		case inside:
			drag.over = l
			// before the first card whose middle is below the dragged one's
			mid := drag.pos.Y - drag.grab.Y + drag.h/2
			l.at = 0
			for i, cd := range l.cards {
				if i == d {
					continue
				}
				cb := cd.e.LayoutBounds()
				y := cb.Y + cb.H/2
				if d >= 0 && i > d {
					y -= room
				}
				if y < mid {
					l.at++
				}
			}
		case drag.over == l:
			drag.over = nil
		}
	} else if !held {
		l.at = -1
	}
	m := 0
	for i, cd := range l.cards {
		if i == d {
			continue
		}
		var s float32
		if held {
			if d >= 0 && i > d {
				s -= room
			}
			if l.at >= 0 && m >= l.at {
				s += room
			}
		}
		if settle {
			cd.e.Jump("shift", 0)
		}
		cd.e.Translate(0, cd.e.AnimateTo("shift", s, springShift))
		m++
	}
	if held && l.at >= 0 && d < 0 {
		return room
	}
	return 0
}

// spacer is the room a list grows by, after its last card.
func spacer(c *ui.Context, room, gap float32, settle bool) {
	e := ui.Box(c).FillWidth().Margin(-gap, 0, 0, 0).Shrink(0)
	if settle {
		e.Jump("room", room)
	}
	e.Height(max(0, e.AnimateTo("room", room, springShift)))
}

func abs32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

// grip is the grip: dots in two columns of three rows.
func grip(c *ui.Context, color ui.Color) {
	ui.Row(c).Gap(3).Shrink(0).Children(func() {
		for range 2 {
			ui.Column(c).Gap(3).Children(func() {
				for range 3 {
					ui.Box(c).Size(2.5, 2.5).Radius(1.25).Background(color)
				}
			})
		}
	})
}

// tile is the raised tile being dragged: the button's skin, the control's
// corner radius; lifted, it takes the pop shadow.
func (t *Tokens) tile(e ui.Element, hovered, lifted bool) {
	cl := t.Colors
	e.Radius(RadiusControl).Cursor(ui.CursorGrab)
	sh := t.Shadows.ControlShadow
	if lifted {
		sh = t.Shadows.ShadowPop
	}
	t.raised(e, t.controlGradient(cl.ControlTop, cl.ControlBottom), cl.ControlBorder, sh, raisedState{hovered: hovered})
}

// SortableOptions are the options of Sortable.
type SortableOptions struct {
	// Label is a row's text: its accessible name.
	Label func(i int) string
	// Key is a row's identity; its index if unset. Provide it so that rows'
	// state (local values, focus) is kept as they move.
	Key func(i int) any
}

// sortValue is what a Sortable's row carries: its list and its key.
type sortValue struct {
	list *ui.ListState
	key  any
}

// Sortable is a list whose rows are reordered by dragging: each row is a
// raised tile with a grip. A row picked up lifts and follows the pointer,
// the others slide apart where it would go, and let go it flies into its
// place; Alt+Up and Alt+Down move the focused row. The list reorders
// *items itself. s identifies the list and must stay the same across
// frames.
//
//	mizan.Sortable(c, &app.list, &app.tasks, mizan.SortableOptions{}, func(t Task) {
//		ui.Text(c, t.Title)
//	})
func Sortable[T any](c *ui.Context, s *ui.ListState, items *[]T, o SortableOptions, row func(T)) ui.Element {
	t := From(c)
	cl := t.Colors
	const gap = 6
	key := func(i int) any {
		if o.Key != nil {
			return o.Key(i)
		}
		return i
	}
	settleDrop(c)
	list := ui.Column(c).FillWidth().Gap(gap).Padding(2)
	l := ui.Local(list, "list", func() *dropList { return &dropList{at: -1, gen: drag.gen} })
	settle := (*l).gen != drag.gen
	prev := (*l).heights
	(*l).cards = (*l).cards[:0]
	(*l).accept = func(v any) bool { sv, ok := v.(sortValue); return ok && sv.list == s }
	(*l).handler = func(v any, at int) any {
		from := -1
		for i := range *items {
			if key(i) == v.(sortValue).key {
				from = i
			}
		}
		if from < 0 {
			return v
		}
		to := at
		if at > from {
			to++
		}
		*items = move(*items, []int{from}, to)
		return sortValue{s, key(at)}
	}
	// a row moved by the keys comes from its old place
	nudge := ui.Local(list, "nudge", func() map[any]float32 { return nil })
	focus := ui.Local(list, "focus", func() any { return nil })
	// the move Alt+Up or Alt+Down asked for in the last frame
	kbd := ui.Local(list, "kbd", func() [2]int { return [2]int{} })
	if from, to := (*kbd)[0], (*kbd)[1]; from != to && from < len(prev) && to < len(prev) {
		// the two rows trade places, each coming from where it was
		step := float32(to - from)
		h, other := prev[from], prev[to]
		*items = move(*items, []int{from}, to+max(to-from, 0))
		*nudge = map[any]float32{key(to): -step * (other + gap), key(from): step * (h + gap)}
		*focus = key(to)
	}
	*kbd = [2]int{}
	building = append(building, *l)
	list.Children(func() {
		for i := range *items {
			k := key(i)
			e := ui.Box(c.Key(k)).FillWidth().Focusable().FocusRing(false)
			if o.Label != nil {
				e.Label(o.Label(i))
			}
			if *focus == k {
				e.Focus()
				*focus = nil
			}
			if off, ok := (*nudge)[k]; ok {
				e.Jump("ly", off)
				delete(*nudge, k)
			}
			for _, step := range []int{-1, 1} {
				j, k := i+step, ui.KeyUp
				if step > 0 {
					k = ui.KeyDown
				}
				if j >= 0 && j < len(*items) && e.Shortcut(ui.Alt, k) {
					*kbd = [2]int{i, j}
					c.AnimationFrame()
				}
			}
			focused := e.FocusVisible()
			t.dragSource(c, e, sortValue{s, k}, func(r ui.Element, lifted bool) {
				r.Row().AlignItems(ui.Center).Gap(12).Padding(10, 14, 10, 12).MinHeight(t.Metrics.ControlH + 8)
				t.tile(r, r.Hovered() && !drag.active, lifted)
				if focused && !lifted {
					t.Shadows.FocusRing.Apply(r)
				}
				r.TextColor(cl.Text).FontSize(FsMd).Children(func() {
					grip(c, cl.Muted)
					ui.Row(c).Grow(1).MinWidth(0).AlignItems(ui.Center).Gap(8).Children(func() { row((*items)[i]) })
				})
			})
		}
		spacer(c, (*l).arrange(c, list, gap), gap, settle)
	})
	building = building[:len(building)-1]
	// the rows' heights, for the keys to move them by in the next frame
	(*l).heights = (*l).heights[:0]
	for _, cd := range (*l).cards {
		(*l).heights = append((*l).heights, cd.e.LayoutBounds().H)
	}
	return list
}

// move moves the rows at the indices rows to before index to (keeping their order).
func move[T any](items []T, rows []int, to int) []T {
	moved := make([]T, 0, len(rows))
	rest := make([]T, 0, len(items))
	at := to
	for i, v := range items {
		if slices.Contains(rows, i) {
			moved = append(moved, v)
			if i < to {
				at--
			}
		} else {
			rest = append(rest, v)
		}
	}
	return slices.Insert(rest, at, moved...)
}

// DragCard is a raised tile dragged within its DropArea or to another;
// value, comparable, is what it carries to where it is dropped. Picked up,
// it lifts and follows the pointer, leaving its place empty.
//
//	mizan.DragCard(c.Key(task.ID), task, func() { ui.Text(c, task.Title) })
func DragCard(c *ui.Context, value any, body func()) ui.Element {
	t := From(c)
	e := ui.Column(c).FillWidth().MinWidth(0)
	t.dragSource(c, e, value, func(r ui.Element, lifted bool) {
		r.Gap(4).Padding(10, 12).MinWidth(0)
		t.tile(r, r.Hovered() && !drag.active, lifted)
		r.TextColor(t.Colors.Text).FontSize(FsMd).Children(body)
	})
	return e
}

// DropAreaOptions are the options of DropArea.
type DropAreaOptions struct {
	Title string
	// Count is the neutral count next to the title (not shown if empty).
	Count string
	// Accept says whether to take the dragged value (every value of T if
	// unset): to refuse a card from another board, for example.
	Accept func(v any) bool
	// Fill stretches the column to its container's height: columns side by
	// side stay the same height, and an empty column still has room to drop on.
	Fill bool
}

// DropArea is an inset column of DragCards taking values of type T, as a
// board's column: a card can be moved within it and in from another.
// While one is held over it, its cards slide apart where it would go, and
// the column glows, a cloud's shadow of the accent drifting across it; as
// the card lands, a light runs up both its sides. drop gets the value let
// go on it, and index, where it goes among the column's other cards: move
// it there, as it is applied before any column builds the next frame.
//
//	mizan.DropArea(c.Key("done"), mizan.DropAreaOptions{Title: "Done"}, func() {
//		for _, task := range app.done { mizan.DragCard(c.Key(task.ID), task, ...) }
//	}, func(task *Task, index int) {
//		app.moveTo(task, "done", index)
//	})
func DropArea[T any](c *ui.Context, o DropAreaOptions, body func(), drop func(v T, index int)) ui.Element {
	t := From(c)
	cl := t.Colors
	const gap = 8
	settleDrop(c)
	e := ui.Column(c).Gap(gap).Padding(10).Radius(RadiusLg).MinWidth(0)
	if o.Fill {
		e.Grow(1)
	}
	l := ui.Local(e, "list", func() *dropList { return &dropList{at: -1, gen: drag.gen} })
	settle := (*l).gen != drag.gen
	(*l).cards = (*l).cards[:0]
	(*l).accept = func(v any) bool {
		_, ok := v.(T)
		return ok && (o.Accept == nil || o.Accept(v))
	}
	(*l).handler = func(v any, at int) any {
		if drop != nil {
			drop(v.(T), at)
		}
		return v
	}
	over := e.AnimateTo("over", b2f(drag.active && drag.over == *l), ui.Animation{Duration: 260 * time.Millisecond, Ease: ui.EaseOut})
	if (*l).flash {
		(*l).flash = false
		e.Jump("flash", 0)
	}
	flash := e.AnimateTo("flash", 1, ui.Animation{Duration: 1500 * time.Millisecond, Ease: ui.EaseInOut})
	e.Background(Mix(cl.Well, cl.ColorAccent500, 0.04*over)).Border(1, Mix(cl.WellBorder, cl.ColorAccent500, 0.55*over))
	t.Shadows.WellShadow.Apply(e)
	t.cloud(e, over)
	t.rise(e, flash)
	building = append(building, *l)
	e.Children(func() {
		if o.Title != "" {
			ui.Row(c).AlignItems(ui.Center).Gap(8).Padding(2, 4, 4, 4).Children(func() {
				ui.Text(c, o.Title).TextColor(cl.Text2).FontSize(FsSm).FontWeight(600)
				if o.Count != "" {
					Badge(c, o.Count, ToneNeutral)
				}
			})
		}
		body()
		spacer(c, (*l).arrange(c, e, gap), gap, settle)
	})
	building = building[:len(building)-1]
	return e
}

// cloud paints the glow of a column a card is held over, at strength
// over: three soft blots of the accent, as the shadows of clouds, drift
// across the well, each along its own smooth noise, so that the glow
// never repeats nor stands still.
func (t *Tokens) cloud(e ui.Element, over float32) {
	if over <= 0.001 {
		return
	}
	accent := t.Colors.ColorAccent500
	e.Draw(func(p *ui.Painter, r ui.Rect) {
		s := float64(p.Now().UnixMilli()%1_000_000) / 1000
		for i := range 3 {
			k := float64(i)*17.3 + 3.1
			cx := 0.5 + 0.42*noise(s*0.22+k)
			cy := 0.5 + 0.45*noise(s*0.17+k+40)
			rad := max(r.W, r.H) * float32(0.42+0.12*noise(s*0.3+k+80))
			a := over * float32(0.11+0.04*noise(s*0.4+k+120))
			p.FillRadialGradient(r, ui.RadialGradient{Center: ui.Point{X: float32(cx), Y: float32(cy)}, Radius: rad, Oklab: true,
				Stops: []ui.GradientStop{{Color: accent.Alpha(a), At: 0}, {Color: accent.Alpha(a * 0.45), At: 0.45}, {Color: accent.Alpha(0), At: 1}}}, RadiusLg)
		}
		p.AnimationFrame()
	})
}

// rise paints the light that runs up a column's sides as a card lands in
// it, at f from 0 to 1: two sparks leave the middle of its bottom edge,
// round its lower corners and climb its sides, each with a fading tail,
// to meet at the top as they fade out.
func (t *Tokens) rise(e ui.Element, f float32) {
	if f <= 0 || f >= 1 {
		return
	}
	cl := t.Colors
	light := Mix(cl.ColorAccent400, ui.RGB(255, 255, 255), 0.35)
	e.DrawOver(func(p *ui.Painter, r ui.Rect) {
		r = ui.Rect{X: r.X + 0.5, Y: r.Y + 0.5, W: r.W - 1, H: r.H - 1}
		head := f
		fade := min(1, (1-f)*2.5) * min(1, f*8)
		const tail, n = 0.28, 7
		for _, left := range []bool{true, false} {
			path := halfBorder(r, RadiusLg, left)
			for j := range n {
				from := max(0, head-tail*float32(n-j)/n)
				seg := *path
				p.StrokePath(seg.Trim(from, head), 3.5, light.Alpha(0.04*fade))
				seg = *path
				p.StrokePath(seg.Trim(from, head), 1, light.Alpha(0.22*fade))
			}
		}
	})
}

// halfBorder is half of r's rounded border, from the middle of its bottom
// edge round the left (or right) side to the middle of its top edge.
func halfBorder(r ui.Rect, radius float32, left bool) *ui.Path {
	rad := min(radius, r.W/2, r.H/2)
	cx, x, sx := r.X+r.W/2, r.X, float32(1)
	if !left {
		x, sx = r.X+r.W, -1
	}
	top, bottom := r.Y, r.Y+r.H
	var p ui.Path
	p.MoveTo(cx, bottom).LineTo(x+sx*rad, bottom).
		QuadTo(x, bottom, x, bottom-rad).LineTo(x, top+rad).
		QuadTo(x, top, x+sx*rad, top).LineTo(cx, top)
	return &p
}

// noise is smooth noise in [-1, 1]: values at whole numbers, eased
// between them.
func noise(x float64) float64 {
	i := math.Floor(x)
	f := x - i
	u := f * f * (3 - 2*f)
	h := func(n float64) float64 {
		v := math.Sin(n*127.1+311.7) * 43758.5453
		return v - math.Floor(v)
	}
	return (h(i)*(1-u)+h(i+1)*u)*2 - 1
}

// DropZoneOptions are the options of DropZone.
type DropZoneOptions struct {
	// Title is the bold line ("Drop files here"); Hint the muted description
	// below it ("PNG, PDF · up to 10 MB").
	Title string
	Hint  string
	// Browse, if set, makes the zone clickable and is called when it is
	// clicked (for example, a file picker with mygo.Dialog.Open).
	Browse   func()
	Disabled bool
}

// DropZone is a well with a dashed border that takes files from other apps
// (Finder, Explorer): when files come over it, the border turns solid and
// takes the accent color. It returns the dropped files' paths, or nil.
//
//	if paths := mizan.DropZone(c.Key("upload"), mizan.DropZoneOptions{Title: "Drop files"}); paths != nil {
//		app.add(paths)
//	}
func DropZone(c *ui.Context, o DropZoneOptions) []string {
	t := From(c)
	cl := t.Colors
	var e ui.Element
	if o.Browse != nil {
		e = ui.ButtonBase(c).Label(or(o.Title, "Choose files")).Cursor(ui.CursorPointer).Disabled(o.Disabled).FocusRing(false)
	} else {
		e = ui.Column(c)
	}
	e.FillWidth().AlignItems(ui.Center).Gap(4).Padding(24, 16).Radius(RadiusControl)
	over := !o.Disabled && e.FileDragOver()
	var files []string
	if !o.Disabled {
		files = e.DroppedFiles()
	}
	switch {
	case over:
		e.Background(Mix(cl.Well, cl.ColorAccent500, 0.08)).Border(1, cl.ColorAccent500)
	case o.Browse != nil && e.FocusVisible():
		e.Background(cl.Well).Border(1, cl.ColorAccent500).BorderStyle(ui.BorderDashed)
		t.Shadows.Ring.Apply(e)
	case e.Hovered() && !o.Disabled:
		e.Background(cl.Well).Border(1, cl.Muted).BorderStyle(ui.BorderDashed)
	default:
		e.Background(cl.Well).Border(1, Mix(ui.Color{}, cl.Muted, 0.45)).BorderStyle(ui.BorderDashed)
	}
	e.Transition(ui.ElementTransition{Duration: 150 * time.Millisecond, Colors: true})
	if o.Disabled {
		e.Opacity(0.4)
	}
	e.TextColor(cl.Muted).FontSize(FsSm).Children(func() {
		ui.Icon(c, iconUpload).Size(20, 20).TextColor(cl.Muted).Margin(0, 0, 4, 0)
		if o.Title != "" {
			ui.Text(c, o.Title).TextColor(cl.Text).FontSize(FsMd).FontWeight(600).TextAlign(ui.Center)
		}
		if o.Hint != "" {
			ui.Text(c, o.Hint).TextAlign(ui.Center).Wrap()
		}
	})
	if o.Browse != nil && !o.Disabled && e.Clicked() {
		o.Browse()
	}
	return files
}

// FileItem is the row of an attached file: a well background, the name
// truncated, the size tabular, a remove button on the right. It returns
// true when Remove is pressed.
func FileItem(c *ui.Context, name, size string) bool {
	t := From(c)
	cl := t.Colors
	removed := false
	ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(10).MinHeight(36).Padding(0, 4, 0, 12).Radius(RadiusItem).
		Background(cl.Well).FontSize(FsSm).Children(func() {
		ui.Icon(c, iconFile).TextColor(cl.Muted)
		ui.Text(c, name).TextColor(cl.Text).Grow(1).MinWidth(0).SingleLine().Ellipsis("…")
		if size != "" {
			ui.Text(c, size).TextColor(cl.Muted).FontFeatures("tnum").NoWrap()
		}
		removed = inputButton(c.Key("remove"), t, iconClear, "Remove").Margin(0).Clicked()
	})
	return removed
}
