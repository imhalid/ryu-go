package ryugo

import (
	"slices"
	"time"

	"github.com/egoist/mygo/ui"
)

// Palettes are tool panels floating over a board, as Pixelmator Classic's
// over the photo: a dark glass HUD in both themes, a thin title bar with a
// round close button, compact controls. A palette moves by its title bar:
// it lifts, follows the pointer, and settles onto the board's edges and
// onto the edges of the palettes near it, as a magnet would (it holds there
// until pulled 16px away, a light running along the edge that settles). A
// press brings it to the front; a double click on its title folds it up to
// its title bar; closed, it shrinks away, and opens again growing.
//
// Experimental, as the Dock: they float in the window, so that they work
// the same everywhere (Wayland says nowhere where windows are).

// PaletteState is how a palette shows: plain data, saved with its Board.
type PaletteState struct {
	X      float32 `json:"x"`
	Y      float32 `json:"y"`
	Placed bool    `json:"placed,omitempty"` // X and Y are set: by At, or by a drag
	// SX and SY are its place on the screen, as a window of its own (see
	// PaletteHost); SPlaced, that they are set
	SX        float32 `json:"sx,omitempty"`
	SY        float32 `json:"sy,omitempty"`
	SPlaced   bool    `json:"splaced,omitempty"`
	Collapsed bool    `json:"collapsed,omitempty"`
	Closed    bool    `json:"closed,omitempty"`
	Z         int     `json:"z,omitempty"` // the palette with the greatest is in front
}

// Board is the state of a board's palettes, by ID: plain data the app
// keeps, and saves as JSON.
type Board struct {
	Palettes map[string]*PaletteState `json:"palettes"`
	// Host, given, floats the palettes in windows of their own, as
	// Pixelmator Classic's, where the system lets windows go where asked
	// (not Wayland); elsewhere they float over the board.
	Host PaletteHost `json:"-"`

	// what the palettes' windows build from: the board's palettes and its
	// theme as its last frame had them, and the windows' own state
	items  map[string]paletteItem
	tokens *Tokens
	wins   map[string]*paletteWin
}

// Reset puts every palette back where it first showed, open.
func (b *Board) Reset() { b.Palettes = nil }

func (b *Board) state(id string) *PaletteState {
	if b.Palettes == nil {
		b.Palettes = map[string]*PaletteState{}
	}
	s := b.Palettes[id]
	if s == nil {
		s = &PaletteState{}
		b.Palettes[id] = s
	}
	return s
}

// front puts palette id in front of the others.
func (b *Board) front(id string) {
	top := 0
	for k, s := range b.Palettes {
		if k != id {
			top = max(top, s.Z)
		}
	}
	if s := b.state(id); s.Z <= top {
		s.Z = top + 1
	}
}

// IsOpen reports whether palette id shows.
func (b *Board) IsOpen(id string) bool { return !b.state(id).Closed }

// Open shows palette id again, in front, where it was.
//
//	if ryugo.Button(c, "Layers", ryugo.ButtonOptions{}).Clicked() {
//		app.board.Open("layers")
//	}
func (b *Board) Open(id string) {
	b.state(id).Closed = false
	b.front(id)
}

// Close closes palette id, as its close button does.
func (b *Board) Close(id string) { b.state(id).Closed = true }

// PaletteOptions are the options of a palette.
type PaletteOptions struct {
	Title string
	// Width is the palette's width, 220 by default.
	Width float32
	// At is where it first shows, from the board's top left; a negative X
	// or Y counts from the board's right or bottom edge to its own.
	At ui.Point
	// NoClose leaves out the close button.
	NoClose bool
}

// PaletteSet is what a board's builder adds its palettes to.
type PaletteSet struct {
	items []paletteItem
}

type paletteItem struct {
	id   string
	o    PaletteOptions
	body func(c *ui.Context)
}

// Palette adds palette id, which body builds with the context it is given:
// over the board, the board's, in the order of the palettes' Z (the board
// builds them after its content); in a window of its own, that window's.
func (ps *PaletteSet) Palette(id string, o PaletteOptions, body func(c *ui.Context)) {
	ps.items = append(ps.items, paletteItem{id, o, body})
}

// paletteMargin is how far from the board's edges a palette settles, and
// paletteGap how far from another palette beside it.
const paletteMargin, paletteGap = 8, 6

// paletteTitleH is the height of a palette's title bar.
const paletteTitleH = 24

// boardState is a board's: the palettes as the last frame laid them out,
// and the one dragged.
type boardState struct {
	rects map[string]ui.Rect
	drag  *paletteDrag
	// moved is when a drag that moved a palette ended: the release is no
	// first click of a double click
	moved time.Time
}

// paletteDrag is a palette dragged by its title bar: where the pointer
// holds it, and the edges it settled onto, across (x) and down (y).
type paletteDrag struct {
	id     string
	anchor ui.Point
	from   ui.Point // where it was as it was pressed
	press  ui.Point // and the pointer
	moving bool     // the pointer moved a few DIPs: a press alone moves nothing
	x, y   paletteSnap
}

// paletteSnap is a palette settled onto an edge along one axis: where, and
// which of its edges (0 its left or top, 1 its right or bottom), and when.
type paletteSnap struct {
	on   bool
	at   float32
	edge int
	when time.Time
}

// Palettes builds a board filling its parent: build its content (a canvas)
// in it, and add the palettes that float over it with p.Palette. b keeps
// where they are.
//
//	ryugo.Palettes(c.Key("studio"), &app.board, func(p *ryugo.PaletteSet) {
//		app.canvas(c)
//		p.Palette("layers", ryugo.PaletteOptions{Title: "Layers", At: ui.Point{X: -12, Y: 12}}, app.layers)
//	}).Fill()
func Palettes(c *ui.Context, b *Board, build func(p *PaletteSet)) ui.Element {
	root := ui.Box(c).MinWidth(0).MinHeight(0).Clip()
	bs := ui.Local(root, "board", func() *boardState { return &boardState{} })
	st := *bs
	ps := &PaletteSet{}
	free := b.Host != nil && b.Host.Free()
	root.Children(func() {
		build(ps)
		if free {
			b.float(c, root, ps.items)
			return
		}
		items := slices.Clone(ps.items)
		// the one in front last: an absolute element paints over those before it
		slices.SortStableFunc(items, func(a, z paletteItem) int { return b.state(a.id).Z - b.state(z.id).Z })
		rects := map[string]ui.Rect{}
		for _, it := range items {
			if r, ok := palette(c, root, b, st, it); ok {
				rects[it.id] = r
			}
		}
		st.rects = rects
	})
	if free {
		// what changed here shows at once in the palettes' windows (they
		// redraw when the pointer is here only: no window redraws another
		// that redraws it back)
		if _, _, over := c.Root().PointerPosition(); over || b.tokens != From(c) {
			b.Host.Redraw("main")
		}
		b.tokens = From(c)
	}
	return root
}

// palette builds palette it over board root, and returns where the last
// frame laid it out, in the board.
func palette(c *ui.Context, root ui.Element, b *Board, st *boardState, it paletteItem) (ui.Rect, bool) {
	s := b.state(it.id)
	// it shrinks away closed, and grows in opened
	var shown float32
	if s.Closed {
		shown = root.AnimateTo("shown:"+it.id, 0, ui.Animation{Duration: 160 * time.Millisecond, Ease: ui.EaseOut})
	} else {
		shown = root.AnimateTo("shown:"+it.id, 1, *springPop)
	}
	if s.Closed && shown < 0.01 {
		return ui.Rect{}, false
	}
	board := root.LayoutBounds()
	w := it.o.Width
	if w <= 0 {
		w = 220
	}
	last := st.rects[it.id]
	if !s.Placed && board.W > 0 && (it.o.At.Y >= 0 || last.H > 0) {
		// where At puts it, once the board, and for the bottom its own
		// height, are known
		s.X, s.Y = it.o.At.X, it.o.At.Y
		if s.X < 0 {
			s.X = board.W + it.o.At.X - w
		}
		if s.Y < 0 {
			s.Y = board.H + it.o.At.Y - last.H
		}
		s.Placed = true
	}

	dr := st.drag
	held := dr != nil && dr.id == it.id
	if held {
		follow(c, root, b, st, it, s, w)
	}
	t := scoped(c, Options{Mode: Dark})
	cl := t.Colors
	radius := t.R(RadiusControl)
	e := ui.Themed(c.Key(paletteKey(it.id)), t.theme).Absolute().Width(w).Clip()
	// under the pointer it goes straight where it is held; settling onto an
	// edge, or leaving it, it glides there
	glide := held && (c.Now().Sub(dr.x.when) < snapGlide || c.Now().Sub(dr.y.when) < snapGlide)
	if glide {
		e.Left(e.AnimateTo("x", s.X, springGuide)).Top(e.AnimateTo("y", s.Y, springGuide))
		c.AnimationFrame()
	} else {
		e.Left(e.Jump("x", s.X)).Top(e.Jump("y", s.Y))
	}
	if !s.Placed {
		e.Opacity(0)
	}
	if e.Pressed() {
		b.front(it.id)
	}
	// settling onto an edge, the light runs again
	type settled struct{ x, y paletteSnap }
	was := ui.Local(e, "settled", func() settled { return settled{} })
	if held {
		if dr.x.on && (!was.x.on || was.x.at != dr.x.at) {
			e.Jump("spark-x", 0)
		}
		if dr.y.on && (!was.y.on || was.y.at != dr.y.at) {
			e.Jump("spark-y", 0)
		}
		*was = settled{dr.x, dr.y}
	} else {
		*was = settled{}
	}
	lift := e.AnimateTo("lift", b2f(held), springLift)
	sc := (0.96 + 0.04*shown) * (1 + 0.012*lift)
	e.Scale(sc, sc).Opacity(min(shown, 1)).TransformOrigin(0.5, 0)
	if s.Closed {
		// going, it lets the pointer through
		e.IgnoreInput(true)
	}
	// the HUD: a darker glass than a menu's, the pop shadow, a hairline
	t.frost(e, Mix(cl.Pop, ui.RGB(0, 0, 0), 0.3).Alpha(0.84), radius, 24)
	t.Shadows.ShadowPop.Apply(e)
	if lift > 0.01 {
		e.Shadow(0, 18*lift, 40*lift, 0, ui.RGBA(0, 0, 0, 0.35*lift))
	}
	e.DrawOver(func(p *ui.Painter, r ui.Rect) { p.Stroke(r, cl.Hairline, radius, 1) })

	// settling onto an edge, a light runs along it from the pointer's side
	var sparkX, sparkY float32
	if held {
		sparkX = e.AnimateTo("spark-x", 1, sparkCurve)
		sparkY = e.AnimateTo("spark-y", 1, sparkCurve)
	}

	e.Children(func() {
		paletteTitle(c, t, b, it, s, st.moved, func(bar ui.Element) { movePalette(c, root, b, st, it, s, bar) })
		paletteBody(c, t, it, s)
	})

	if held && (sparkX > 0 && sparkX < 1 || sparkY > 0 && sparkY < 1) {
		light := t.sparkLight()
		r := last
		e.DrawOver(func(p *ui.Painter, b ui.Rect) {
			if dr.x.on && sparkX > 0 && sparkX < 1 {
				x := b.X + float32(dr.x.edge)*b.W
				sparks(p, true, x, b.Y+r.H/2, b.Y, b.Y+b.H, sparkX, 1, light)
			}
			if dr.y.on && sparkY > 0 && sparkY < 1 {
				y := b.Y + float32(dr.y.edge)*b.H
				sparks(p, false, y, b.X+b.W/2, b.X, b.X+b.W, sparkY, 1, light)
			}
		})
	}
	lb := e.LayoutBounds()
	return ui.Rect{X: lb.X - board.X, Y: lb.Y - board.Y, W: lb.W, H: lb.H}, lb.W > 0
}

type paletteKey string

// paletteBody builds a palette's body under its title bar: folded up, it
// closes to nothing.
func paletteBody(c *ui.Context, t *Tokens, it paletteItem, s *PaletteState) {
	cl := t.Colors
	wrap := ui.Column(c.Key("body")).Clip().MinHeight(0)
	open := wrap.AnimateTo("open", b2f(!s.Collapsed), springOpen)
	h := ui.Local(wrap, "h", func() float32 { return 0 })
	wrap.Children(func() {
		// its height as the last frame laid it out, whole
		inner := ui.Column(c.Key("inner")).Shrink(0).Padding(8).Gap(8)
		if ib := inner.LayoutBounds(); ib.H > 0 {
			*h = ib.H
		}
		inner.Children(func() { it.body(c) })
	})
	if open < 0.999 {
		wrap.Height(*h*max(open, 0)).InsetShadow(0, 1, 0, 0, cl.Hairline.Alpha(open))
	} else {
		wrap.InsetShadow(0, 1, 0, 0, cl.Hairline)
	}
	if open > 0.001 && open < 0.999 {
		c.AnimationFrame()
	}
}

// paletteTitle builds a palette's title bar: its close button on the
// left, its title in the middle; move makes it move the palette, and a
// double click folds it up (but just after a move, moved, whose release
// would be its first click).
func paletteTitle(c *ui.Context, t *Tokens, b *Board, it paletteItem, s *PaletteState, moved time.Time, move func(bar ui.Element)) {
	cl := t.Colors
	bar := ui.Row(c.Key("title")).Height(paletteTitleH).Padding(0, 6).AlignItems(ui.Center).Cursor(ui.CursorGrab)
	bar.Label(it.o.Title)
	move(bar)
	if bar.DoubleClicked() && c.Now().Sub(moved) > 600*time.Millisecond {
		s.Collapsed = !s.Collapsed
	}
	bar.Children(func() {
		if !it.o.NoClose {
			x := ui.ButtonBase(c.Key("close")).Label("Close "+it.o.Title).Size(13, 13).Radius(RadiusPill).Center().
				Cursor(ui.CursorPointer).FocusRing(false).Shrink(0)
			hot := x.AnimateTo("hot", b2f(x.Hovered() || x.FocusVisible()), ui.Animation{Duration: 120 * time.Millisecond, Ease: ui.EaseOut})
			x.Background(Mix(cl.Fill, cl.Up, hot*0.85))
			press := x.AnimateTo("press", b2f(x.Pressed()), springSqueeze)
			x.Scale(1-0.1*press, 1-0.1*press)
			if x.FocusVisible() {
				t.Shadows.FocusRing.Apply(x)
			}
			x.Children(func() {
				ui.Icon(c, iconClose).Size(9, 9).TextColor(Mix(cl.Muted, cl.OnAccent, hot))
			})
			if x.Clicked() {
				b.Close(it.id)
			}
		}
		// the title in the bar's middle, whatever is on its left
		ui.Box(c).Absolute().Left(0).Right(0).Top(0).Bottom(0).Center().PassThrough().Children(func() {
			ui.Text(c, it.o.Title).FontSize(FsXs).FontWeight(600).TextColor(cl.Text2).SingleLine()
		})
	})
}

// movePalette makes bar move palette it: pressed, it comes to the front
// and is held; let go, it stays.
func movePalette(c *ui.Context, root ui.Element, b *Board, st *boardState, it paletteItem, s *PaletteState, bar ui.Element) {
	_, _, pressed := bar.Dragged()
	dr := st.drag
	switch {
	case pressed && dr == nil:
		b.front(it.id)
		p := boardPointer(root)
		st.drag = &paletteDrag{id: it.id, anchor: ui.Point{X: p.X - s.X, Y: p.Y - s.Y}, from: ui.Point{X: s.X, Y: s.Y}, press: p}
		c.AnimationFrame()
	case !pressed && dr != nil && dr.id == it.id:
		st.drag = nil
		if abs32(s.X-dr.from.X)+abs32(s.Y-dr.from.Y) > 3 {
			st.moved = c.Now()
		}
	}
	if st.drag != nil && st.drag.id == it.id {
		bar.Cursor(ui.CursorGrabbing)
	}
}

// boardPointer is where the pointer is on the board.
func boardPointer(root ui.Element) ui.Point {
	x, y, _ := root.PointerPosition()
	return ui.Point{X: x, Y: y}
}

// follow moves the palette held with the pointer (read from where it is,
// not by adding up its moves), settling it onto the board's edges and the
// edges of the palettes near it, beside it or lined up with it.
func follow(c *ui.Context, root ui.Element, b *Board, st *boardState, it paletteItem, s *PaletteState, w float32) {
	dr := st.drag
	board := root.LayoutBounds()
	h := st.rects[it.id].H
	p := boardPointer(root)
	if !dr.moving && abs32(p.X-dr.press.X)+abs32(p.Y-dr.press.Y) < 4 {
		return
	}
	dr.moving = true
	x, y := p.X-dr.anchor.X, p.Y-dr.anchor.Y
	var xs, ys []snapTarget
	xs = append(xs, snapTarget{paletteMargin, 0}, snapTarget{board.W - paletteMargin - w, 1})
	ys = append(ys, snapTarget{paletteMargin, 0}, snapTarget{board.H - paletteMargin - h, 1})
	for id, o := range st.rects {
		if id == it.id || b.state(id).Closed {
			continue
		}
		const near = 24
		if o.Y-near < y+h && y < o.Y+o.H+near {
			xs = append(xs, snapTarget{o.X, 0}, snapTarget{o.X + o.W - w, 1},
				snapTarget{o.X + o.W + paletteGap, 0}, snapTarget{o.X - paletteGap - w, 1})
		}
		if o.X-near < x+w && x < o.X+o.W+near {
			ys = append(ys, snapTarget{o.Y, 0}, snapTarget{o.Y + o.H - h, 1},
				snapTarget{o.Y + o.H + paletteGap, 0}, snapTarget{o.Y - paletteGap - h, 1})
		}
	}
	now := c.Now()
	settle(&dr.x, x, xs, now)
	settle(&dr.y, y, ys, now)
	if dr.x.on {
		x = dr.x.at
	}
	if dr.y.on {
		y = dr.y.at
	}
	// it stays on the board, by its title bar at least
	s.X = min(max(x, 0), max(board.W-w, 0))
	s.Y = min(max(y, 0), max(board.H-paletteTitleH, 0))
	s.Placed = true
}

// snapTarget is where an edge would settle, and which edge of the palette
// lines up there.
type snapTarget struct {
	at   float32
	edge int
}

// settle settles s onto the nearest of targets within snapIn of pos, and
// holds it there until pos is snapOut away; it reports whether s changed.
func settle(s *paletteSnap, pos float32, targets []snapTarget, now time.Time) bool {
	was := *s
	if s.on && abs32(s.at-pos) >= snapOut {
		s.on = false
	}
	if !s.on {
		best := float32(snapIn)
		for _, t := range targets {
			if d := abs32(t.at - pos); d < best {
				best, s.on, s.at, s.edge = d, true, t.at, t.edge
			}
		}
	}
	if s.on != was.on || s.on && s.at != was.at {
		s.when = now
		return true
	}
	return false
}
