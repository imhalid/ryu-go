package mizan

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// Palettes in windows of their own, as Pixelmator Classic's: each palette
// a borderless, transparent window over the app's (its child), which the
// palette draws itself in, with its shadow in a margin round it. Dragged by
// its title bar, it follows the pointer, settling onto the edges of the
// display's work area, of the app's window (inside and out) and of the
// other palettes, as it does over a board. A palette's window redraws the
// app's and the other palettes' while the pointer is over it, and the
// app's redraws them while the pointer is over it: what changes in one
// shows in the others at once, and no window redraws another that redraws
// it back.

// PaletteHost shows a board's palettes in windows of their own. mizan/
// palettewin is one of MyGo's windows.
type PaletteHost interface {
	// Free reports whether the palettes can float in windows of their own:
	// where the system says where windows and the pointer are, and puts a
	// window where asked (not on Wayland).
	Free() bool
	// Board is where the content of the board's window is on the screen,
	// and Work the work area of its display.
	Board() ui.Rect
	Work() ui.Rect
	// Pointer is where the pointer is on the screen.
	Pointer() ui.Point
	// Show shows palette id in a window of its own, its content at r on the
	// screen, view building it; asked again, it moves or resizes it.
	Show(id string, r ui.Rect, view func(c *ui.Context))
	// Hide closes palette id's window.
	Hide(id string)
	// Redraw redraws the board's window and the palettes' windows, but
	// except: a palette's ID, or "main" for the board's window.
	Redraw(except string)
}

// floatPad is the transparent margin round a palette in its window, for
// its shadow.
const floatPad = 16

// paletteWin is a palette's window: whether it shows, its palette's height
// as it last laid it out, its drag, and when the last drag that moved it
// ended. estimated: placed from the board's bottom before its height was
// known.
type paletteWin struct {
	open      bool
	h         float32
	estimated bool
	drag      *paletteDrag
	moved     time.Time
}

func paletteWidth(o PaletteOptions) float32 {
	if o.Width > 0 {
		return o.Width
	}
	return 220
}

// rect is the window of palette s, w wide and h tall, on the screen.
func floatRect(s *PaletteState, w, h float32) ui.Rect {
	return ui.Rect{X: s.SX - floatPad, Y: s.SY - floatPad, W: w + 2*floatPad, H: max(h, paletteTitleH) + 2*floatPad}
}

// float shows the board's palettes in windows of their own: first where
// they would float over the board (root), then where they are taken.
func (b *Board) float(c *ui.Context, root ui.Element, items []paletteItem) {
	h := b.Host
	if b.wins == nil {
		b.wins = map[string]*paletteWin{}
	}
	b.items = map[string]paletteItem{}
	board, win := root.LayoutBounds(), h.Board()
	for _, it := range items {
		b.items[it.id] = it
		s := b.state(it.id)
		fw := b.wins[it.id]
		if fw == nil {
			fw = &paletteWin{}
			b.wins[it.id] = fw
		}
		w := paletteWidth(it.o)
		if s.Closed {
			// its window shrinks it away, then goes
			continue
		}
		if !s.SPlaced {
			if board.W == 0 {
				continue
			}
			x, y := it.o.At.X, it.o.At.Y
			if x < 0 {
				x = board.W + x - w
			}
			if y < 0 {
				// from the bottom: its height is a guess until it is laid out
				hh := fw.h
				if hh == 0 {
					hh, fw.estimated = 160, true
				}
				y = board.H + y - hh
			}
			s.SX, s.SY, s.SPlaced = win.X+board.X+x, win.Y+board.Y+y, true
		}
		fw.open = true
		h.Show(it.id, floatRect(s, w, fw.h), b.floatView(it.id))
	}
	for id, fw := range b.wins {
		if _, ok := b.items[id]; !ok && fw.open {
			fw.open = false
			h.Hide(id)
		}
	}
}

// dark is the palettes' theme: the board's, dark.
func (b *Board) dark() *Tokens {
	t := b.tokens
	if t == nil {
		return tokensFor(true, Md, CornersMd, ui.Color{})
	}
	return tokensFor(true, t.Size, t.Corners, t.accent)
}

// floatView builds palette id in its own window.
func (b *Board) floatView(id string) func(c *ui.Context) {
	return func(c *ui.Context) {
		it, ok := b.items[id]
		fw := b.wins[id]
		if !ok || fw == nil {
			return
		}
		s := b.state(id)
		h := b.Host
		t := b.dark()
		c.SetTheme(t.theme)
		c.Root().Background(ui.Transparent)
		cl := t.Colors
		w := paletteWidth(it.o)
		radius := t.R(RadiusControl)
		outer := ui.Box(c.Key("palette")).Fill().Padding(floatPad).AlignItems(ui.Start)
		// it grows in as its window opens, and shrinks away closed, its
		// window going after it
		if seen := ui.Local(outer, "seen", func() bool { return false }); !*seen {
			*seen = true
			outer.Jump("shown", 0)
		}
		var shown float32
		if s.Closed {
			shown = outer.AnimateTo("shown", 0, ui.Animation{Duration: 160 * time.Millisecond, Ease: ui.EaseOut})
		} else {
			shown = outer.AnimateTo("shown", 1, *springPop)
		}
		if s.Closed && shown < 0.01 {
			if fw.open {
				fw.open = false
				h.Hide(id)
			}
			return
		}
		dr := fw.drag
		held := dr != nil && dr.moving
		outer.Children(func() {
			e := ui.Column(c.Key("surface")).Width(w).Shrink(0).Clip().Radius(radius).
				Background(Mix(cl.Pop, ui.RGB(0, 0, 0), 0.3).Alpha(0.97))
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
			sc := 0.96 + 0.04*shown
			e.Scale(sc, sc).Opacity(min(shown, 1)).TransformOrigin(0.5, 0)
			// its shadow fits the margin; held, it lifts
			e.Shadow(0, 3+2*lift, 10+4*lift, 0, ui.RGBA(0, 0, 0, 0.42+0.1*lift)).Shadow(0, 0, 0, 0.5, ui.RGBA(0, 0, 0, 0.3))
			var sparkX, sparkY float32
			if held {
				sparkX = e.AnimateTo("spark-x", 1, sparkCurve)
				sparkY = e.AnimateTo("spark-y", 1, sparkCurve)
			}
			light := t.sparkLight()
			e.DrawOver(func(p *ui.Painter, r ui.Rect) {
				p.Stroke(r, cl.Hairline, radius, 1)
				if held && dr.x.on && sparkX > 0 && sparkX < 1 {
					sparks(p, true, r.X+float32(dr.x.edge)*r.W, r.Y+r.H/2, r.Y, r.Y+r.H, sparkX, 1, light)
				}
				if held && dr.y.on && sparkY > 0 && sparkY < 1 {
					sparks(p, false, r.Y+float32(dr.y.edge)*r.H, r.X+r.W/2, r.X, r.X+r.W, sparkY, 1, light)
				}
			})
			e.Children(func() {
				paletteTitle(c, t, b, it, s, fw.moved, func(bar ui.Element) { b.carry(c, id, bar) })
				paletteBody(c, t, it, s)
			})
			// the window takes the palette's height, as it folds up too
			if lb := e.LayoutBounds(); lb.H > 0 && abs32(lb.H-fw.h) > 0.5 {
				if fw.estimated {
					// placed from the board's bottom by a guess: now by its height
					s.SY += 160 - lb.H
					fw.estimated = false
				}
				fw.h = lb.H
				h.Show(id, floatRect(s, w, fw.h), nil)
			}
		})
		// what changed here shows at once in the other windows
		if _, _, over := c.Root().PointerPosition(); over {
			h.Redraw(id)
		}
	}
}

// carry makes bar, the title bar of palette id in its window, carry the
// window: pressed, it is held; dragged a few DIPs, it follows the pointer,
// settling onto the edges of the work area, of the board's window and of
// the other palettes.
func (b *Board) carry(c *ui.Context, id string, bar ui.Element) {
	h, fw, s := b.Host, b.wins[id], b.state(id)
	it := b.items[id]
	_, _, pressed := bar.Dragged()
	p := h.Pointer()
	switch {
	case pressed && fw.drag == nil:
		fw.drag = &paletteDrag{id: id, anchor: ui.Point{X: p.X - s.SX, Y: p.Y - s.SY}, from: ui.Point{X: s.SX, Y: s.SY}, press: p}
		c.AnimationFrame()
		return
	case !pressed && fw.drag != nil:
		if abs32(s.SX-fw.drag.from.X)+abs32(s.SY-fw.drag.from.Y) > 3 {
			fw.moved = c.Now()
		}
		fw.drag = nil
		return
	case fw.drag == nil:
		return
	}
	bar.Cursor(ui.CursorGrabbing)
	dr := fw.drag
	c.AnimationFrame()
	if !dr.moving && abs32(p.X-dr.press.X)+abs32(p.Y-dr.press.Y) < 4 {
		return
	}
	dr.moving = true
	w, hh := paletteWidth(it.o), fw.h
	x, y := p.X-dr.anchor.X, p.Y-dr.anchor.Y
	var xs, ys []snapTarget
	near := func(a0, a1, b0, b1 float32) bool { return b0-24 < a1 && a0 < b1+24 }
	// the work area's edges, inside
	wa := h.Work()
	xs = append(xs, snapTarget{wa.X + paletteMargin, 0}, snapTarget{wa.X + wa.W - paletteMargin - w, 1})
	ys = append(ys, snapTarget{wa.Y + paletteMargin, 0}, snapTarget{wa.Y + wa.H - paletteMargin - hh, 1})
	// the board's window's and the other palettes' edges: beside them,
	// outside, and lined up with them, inside
	add := func(o ui.Rect, in float32) {
		if near(y, y+hh, o.Y, o.Y+o.H) {
			xs = append(xs, snapTarget{o.X + in, 0}, snapTarget{o.X + o.W - in - w, 1},
				snapTarget{o.X + o.W + paletteGap, 0}, snapTarget{o.X - paletteGap - w, 1})
		}
		if near(x, x+w, o.X, o.X+o.W) {
			ys = append(ys, snapTarget{o.Y + in, 0}, snapTarget{o.Y + o.H - in - hh, 1},
				snapTarget{o.Y + o.H + paletteGap, 0}, snapTarget{o.Y - paletteGap - hh, 1})
		}
	}
	add(h.Board(), paletteMargin)
	for oid, ofw := range b.wins {
		if oid == id || !ofw.open || b.state(oid).Closed {
			continue
		}
		os := b.state(oid)
		add(ui.Rect{X: os.SX, Y: os.SY, W: paletteWidth(b.items[oid].o), H: ofw.h}, 0)
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
	s.SX, s.SY = x, y
	h.Show(id, floatRect(s, w, hh), nil)
}

// Hide closes the palettes' windows, keeping where they are and how they
// show: call it while the board isn't built (another page shows). Built
// again, the board opens them again.
func (b *Board) Hide() {
	for id, fw := range b.wins {
		if fw.open {
			fw.open = false
			b.Host.Hide(id)
		}
	}
}
