package ryugo

import (
	"slices"
	"time"

	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"
)

// Layers float above the page: dialogs and sheets on the card surface over
// a scrim, popovers, menus and toasts on frosted glass. They move on
// springs, quick to arrive and quick to leave, and content scrolling in a
// dialog or a sheet fades into its header and footer.

var iconClose = Svg(16, 14, "M4.5 4.5l7 7M11.5 4.5l-7 7")

// DialogOptions are the options of Dialog and Sheet.
type DialogOptions struct {
	Title       string
	Description string
	// Width is the panel's width: 480 for a dialog, 420 for a sheet.
	Width float32
	// Footer builds the buttons at the bottom: aligned to the end in a
	// dialog, to the start in a sheet, below a groove.
	Footer func()
	// NoClose hides the close button in the header; Escape and a click on
	// the scrim still close.
	NoClose bool
	// Fixed keeps a dialog where it opens: without it, its header drags
	// it around the window.
	Fixed bool
	// Bare drops the header, the body's padding and scroll, and the footer:
	// Bare builds the whole card, as columns running edge to edge, say. It
	// is handed grip, to which it gives the element that drags the dialog
	// (unless Fixed), and close, which closes it. Title, Description and
	// Footer are not used.
	Bare func(grip func(ui.Element), close func())
}

// The springs of the layers: they settle fast, with a hint of overshoot.
var (
	springPanel = &ui.Animation{Spring: ui.SpringOptions{Duration: 300 * time.Millisecond, Bounce: 0.12}}
	springSheet = &ui.Animation{Spring: ui.SpringOptions{Duration: 380 * time.Millisecond, Bounce: 0.06}}
	springPop   = &ui.Animation{Spring: ui.SpringOptions{Duration: 220 * time.Millisecond, Bounce: 0.1}}
	// a tooltip following the pointer, a little behind it
	springFollow = ui.Animation{Spring: ui.SpringOptions{Duration: 260 * time.Millisecond, Bounce: 0.04}}
	springRest   = ui.Animation{Spring: ui.SpringOptions{Duration: 420 * time.Millisecond, Bounce: 0.18}}
	springTilt   = ui.Animation{Spring: ui.SpringOptions{Duration: 240 * time.Millisecond}}
	// leaving is quicker than arriving, and never overshoots
	exitPanel = ui.Animation{Duration: 150 * time.Millisecond, Ease: ui.EaseOut}
	exitSheet = ui.Animation{Duration: 220 * time.Millisecond, Ease: ui.CubicBezier(0.4, 0, 1, 1)}
)

// presence builds a dialog's layer while *open is true, and on until it
// has played out its exit: p goes from 0 to 1 as it opens (with enter) and
// back to 0 as it closes (with exit), for build to move it by. A layer
// whose parent appears with it gets no Transition, so the layer is moved
// by hand. A click on the scrim or Escape sets *open to false. The state
// lives on a host element that takes no room.
func presence(c *ui.Context, open *bool, enter, exit ui.Animation, build func(backdrop, panel ui.Element, p float32)) ui.Element {
	host := ui.Box(c).Absolute().Size(0, 0).PassThrough()
	shown := ui.Local(host, "shown", func() bool { return false })
	if *open && !*shown {
		*shown = true
		host.Jump("p", 0)
	}
	a := enter
	if !*open {
		a = exit
	}
	p := host.AnimateTo("p", b2f(*open), a)
	if !*open && host.Settled("p") {
		*shown = false
	}
	if !*shown {
		return ui.Element{}
	}
	// while it closes, it is still built, but no longer *open: the pointer
	// goes through it to the page, so that a click there is not lost
	vis, closing := true, !*open
	panel := ui.DialogBase(c, &vis, func(backdrop, panel ui.Element) {
		build(backdrop, panel, p)
		if closing {
			backdrop.PassThrough().IgnoreInput(true)
			panel.PassThrough().IgnoreInput(true)
		}
	})
	if !vis {
		*open = false
	}
	return panel
}

// Dialog shows a modal panel while *open is true: a card with the pop
// shadow over a scrim that fades in, that springs in from 94% and 10px
// below, and fades out on the way down as it closes. Its
// header drags it around the window (unless Fixed): it tilts with the
// pointer's speed, a throw glides on, and it springs back inside the
// window when let go; a double click on the header brings it back to the
// middle. Escape, a click on the scrim or the close button set *open to
// false, which the returned panel's Dismissed reports. It returns the
// panel, or an empty element while closed.
//
//	ryugo.Dialog(c.Key("delete"), &app.confirm, ryugo.DialogOptions{
//		Title: "Delete the report?",
//		Footer: func() {
//			ryugo.Button(c.Key("cancel"), "Cancel", ryugo.ButtonOptions{})
//			ryugo.Button(c.Key("ok"), "Delete", ryugo.ButtonOptions{Danger: true})
//		},
//	}, func() { ui.Text(c, "This cannot be undone.") })
func Dialog(c *ui.Context, open *bool, o DialogOptions, body func()) ui.Element {
	t := From(c)
	return presence(c, open, *springPanel, exitPanel, func(backdrop, panel ui.Element, p float32) {
		backdrop.Padding(24, 16).Background(t.Colors.Scrim.Alpha(clamp01(p)))
		panel.Width(or(o.Width, 480)).MaxWidthPercent(100).MaxHeight(720)
		card, head := t.panelContent(c, open, o, body, false, t.R(RadiusLg))
		s := 0.94 + 0.06*p
		card.Translate(0, 10*(1-p)).Scale(s, s).Opacity(clamp01(p))
		if !o.Fixed && head != (ui.Element{}) {
			t.float(backdrop, panel, head)
		}
	})
}

// dragState is where a dragged dialog is: pos is its offset at rest or
// under the pointer, before the rubber band.
type dragState struct{ pos ui.Point }

// float lets head drag panel around backdrop: the panel follows the
// pointer at once, past the window's edges less and less far; let go, it
// glides on with the throw and springs to rest inside the window.
func (t *Tokens) float(backdrop, panel, head ui.Element) {
	st := ui.Local(panel, "drag", func() dragState { return dragState{} })
	area, box := backdrop.Bounds(), panel.LayoutBounds()
	// how far the panel may go each way and stay in the window
	const margin = 8
	lo := ui.Point{X: area.X + margin - box.X, Y: area.Y + margin - box.Y}
	hi := ui.Point{X: area.X + area.W - margin - box.X - box.W, Y: area.Y + area.H - margin - box.Y - box.H}
	if area.W == 0 || box.W == 0 {
		lo, hi = ui.Point{}, ui.Point{}
	}
	clamp := func(p ui.Point) ui.Point {
		return ui.Point{X: min(max(p.X, lo.X), max(hi.X, lo.X)), Y: min(max(p.Y, lo.Y), max(hi.Y, lo.Y))}
	}
	rubber := func(v, lo, hi float32) float32 {
		switch {
		case v < lo:
			return lo - rubberband(lo-v)
		case v > hi && hi >= lo:
			return hi + rubberband(v-hi)
		}
		return v
	}
	head.Cursor(ui.CursorGrab)
	var x, y float32
	dx, dy, dragging := head.Dragged()
	vx, _ := head.DragVelocity()
	if dragging {
		st.pos.X += dx
		st.pos.Y += dy
		head.Cursor(ui.CursorGrabbing)
		x = panel.Jump("x", rubber(st.pos.X, lo.X, hi.X))
		y = panel.Jump("y", rubber(st.pos.Y, lo.Y, hi.Y))
	} else {
		if rvx, rvy, ok := head.DragEnded(); ok {
			// the throw carries it on a little, as far as it would glide
			st.pos = clamp(ui.Point{X: st.pos.X + rvx*0.12, Y: st.pos.Y + rvy*0.12})
		}
		if head.DoubleClicked() {
			st.pos = ui.Point{}
		}
		st.pos = clamp(st.pos)
		x = panel.AnimateTo("x", st.pos.X, springRest)
		y = panel.AnimateTo("y", st.pos.Y, springRest)
	}
	// it leans into the motion, up to 2 degrees, and lifts a little while held
	tilt := panel.AnimateTo("tilt", min(max(vx*0.0016, -2), 2), springTilt)
	lift := panel.AnimateTo("lift", b2f(dragging), springTilt)
	s := 1 + 0.012*lift
	panel.Translate(x, y).Rotate(tilt).Scale(s, s)
}

// rubberband is how far past an edge an element pulled d past it goes, as
// iOS holds content pulled past its edges.
func rubberband(d float32) float32 {
	const dim = 120
	return (1 - 1/(d*0.55/dim+1)) * dim
}

// Sheet shows a panel along the window's end edge while *open is true,
// over a scrim that fades in, as Dialog's: it springs in from past the
// edge, tilted a little, and straightens as it lands; closing, it
// speeds back out past the edge. It closes as Dialog does.
func Sheet(c *ui.Context, open *bool, o DialogOptions, body func()) ui.Element {
	t := From(c)
	w := or(o.Width, 420)
	return presence(c, open, *springSheet, exitSheet, func(backdrop, panel ui.Element, p float32) {
		backdrop.AlignItems(ui.End).Justify(ui.Stretch).Background(t.Colors.Scrim.Alpha(clamp01(p)))
		panel.Width(w).MaxWidthPercent(100).FillHeight()
		card, _ := t.panelContent(c, open, o, body, true, 0)
		card.TransformOrigin(0, 0.5).Translate((w+24)*(1-p), 0).Rotate(2 * (1 - p))
	})
}

// layer paints a floating panel: the card surface, the pop shadow.
func (t *Tokens) layer(e ui.Element, radius float32) {
	e.Radius(radius).Background(t.Colors.Card).TextColor(t.Colors.Text).Clip()
	t.Shadows.ShadowPop.Apply(e)
}

// panelContent builds a dialog's or a sheet's card, inside its panel: its
// header, scrolling body and footer. It returns the card and the header.
func (t *Tokens) panelContent(c *ui.Context, open *bool, o DialogOptions, body func(), sheet bool, radius float32) (card, head ui.Element) {
	cl := t.Colors
	card = ui.Column(c).FillWidth().Grow(1).MinHeight(0)
	t.layer(card, radius)
	if o.Bare != nil {
		card.Children(func() {
			o.Bare(func(e ui.Element) { head = e }, func() { *open = false })
		})
		return card, head
	}
	sc := ui.Local(card, "scroll", func() ui.ScrollState { return ui.ScrollState{} })
	card.Children(func() {
		head = ui.Row(c).AlignItems(ui.Start).Gap(12).Padding(22, 22, 0, 22)
		if sheet {
			head.AlignItems(ui.Center).Padding(18, 18, 6, 22)
		}
		head.Children(func() {
			ui.Column(c).Grow(1).MinWidth(0).Gap(4).Children(func() {
				if o.Title != "" {
					ui.Text(c, o.Title).FontSize(FsXl).FontWeight(600).LetterSpacing(-0.02 * FsXl).Wrap().Role(ui.RoleHeading).Level(2)
				}
				if o.Description != "" {
					ui.Text(c, o.Description).TextColor(cl.Muted).FontSize(FsMd).Wrap()
				}
			})
			if !o.NoClose {
				b := ghostButton(c.Key("close"), t, iconClose, "Close", 30)
				if !sheet {
					b.Margin(-4, -6, 0, 0)
				}
				if b.Clicked() {
					*open = false
				}
			}
		})
		bp := []float32{16, 22}
		if sheet {
			bp = []float32{10, 22, 22, 22}
		}
		ui.Box(c).Grow(1).MinHeight(0).Children(func() {
			ui.Scroll(c).Fill().TrackScroll(sc).Children(func() {
				ui.Column(c).FillWidth().Padding(bp...).TextColor(cl.Text2).FontSize(FsBase).Children(body)
			})
			// content scrolling past the body's edges fades into the
			// header and footer, the more the nearer the edge
			t.scrollEdge(c.Key("top"), false, sc.Y)
			t.scrollEdge(c.Key("bottom"), true, sc.MaxY-sc.Y)
		})
		if o.Footer == nil {
			return
		}
		f := ui.Row(c).Gap(GapControl).Padding(14, 22, 22, 22).Wrap()
		if sheet {
			// a groove above the footer: a dark line, then a light one
			f.InsetShadow(0, 1, 0, 0, cl.SepDark).InsetShadow(0, 2, 0, 0, cl.SepLight)
		} else {
			f.Justify(ui.End)
		}
		f.Children(o.Footer)
	})
	return card, head
}

// scrollEdge is the fade at a scroll body's top (or bottom) edge: it shows
// as the content goes under that edge, by how far there is to scroll on
// that side.
func (t *Tokens) scrollEdge(c *ui.Context, bottom bool, past float32) {
	const h = 40
	angle := float32(180)
	if bottom {
		angle = 0
	}
	e := ui.Box(c).Absolute().Left(0).Right(0).Height(h).PassThrough()
	if bottom {
		e.Bottom(0)
	} else {
		e.Top(0)
	}
	e.Opacity(e.AnimateTo("show", clamp01(past/24), ui.Animation{Duration: 120 * time.Millisecond}))
	card := t.Colors.Card
	e.Draw(func(p *ui.Painter, r ui.Rect) {
		p.FillGradient(r, ui.LinearGradient{Angle: angle, Oklab: true, Stops: []ui.GradientStop{
			{Color: card.Alpha(0.97), At: 0}, {Color: card.Alpha(0.7), At: 0.35}, {Color: card.Alpha(0), At: 1},
		}}, 0)
	})
}

// ghostButton is a round, flat icon button: no fill until hovered.
func ghostButton(c *ui.Context, t *Tokens, icon *ui.SVG, label string, d float32) ui.Element {
	b := ui.ButtonBase(c).Label(label).Center().Size(d, d).Radius(RadiusPill).Shrink(0).Cursor(ui.CursorPointer).FocusRing(false)
	color := t.Colors.Muted
	if b.Hovered() {
		b.Background(t.Colors.Fill)
		color = t.Colors.Text
	}
	if b.FocusVisible() {
		t.Shadows.FocusRing.Apply(b)
	}
	b.TextColor(color).Children(func() { ui.Icon(c, icon) })
	return b
}

// Popover shows a panel below anchor while *open is true, on the surface
// Select's list uses. A press outside it or Escape closes it. Build anchor
// before it, in the same frame.
//
//	b := ryugo.Button(c.Key("filter"), "Filter", ryugo.ButtonOptions{})
//	if b.Clicked() {
//		app.filtering = !app.filtering
//	}
//	ryugo.Popover(c.Key("filters"), b, &app.filtering, func() { ... })
func Popover(c *ui.Context, anchor ui.Element, open *bool, body func()) ui.Element {
	t := From(c)
	return ui.PopoverBase(c, anchor, open, func(panel ui.Element) {
		popover(t, panel)
		panel.MinWidth(200).MaxWidth(360).Padding(14).Gap(10).FontSize(FsMd).Children(body)
	})
}

// MenuList builds a Menu's items.
type MenuList struct {
	c     *ui.Context
	t     *Tokens
	open  *bool
	first bool
	n     int
}

// ItemOptions are the options of a menu item.
type ItemOptions struct {
	Icon *ui.SVG
	// Shortcut is shown at the item's end, as "⌘R".
	Shortcut string
	// Checked shows a check before the item; set it on every item of a
	// group that can be checked, so that their labels line up.
	Checked   bool
	Checkable bool
	// Danger colors the item Up: its highlight is a tint, not the accent,
	// so that the menu's loudest item is not its most dangerous.
	Danger   bool
	Disabled bool
}

// Item adds an item and reports whether it was chosen, which closes the
// menu. The highlight, on hover or with the arrows, is the accent gradient.
func (m *MenuList) Item(label string, o ItemOptions) bool {
	c, t := m.c.Key(m.n), m.t
	m.n++
	cl := t.Colors
	e := ui.ButtonBase(c).Label(label).Role(ui.RoleMenuItem).FillWidth().AlignItems(ui.Center).Gap(10).
		MinHeight(34).Padding(0, 12, 0, 10).Radius(Concentric(t.R(RadiusPop), PadPop)).FontSize(FsMd).Cursor(ui.CursorPointer).
		FocusRing(false).Disabled(o.Disabled)
	if m.first && !o.Disabled {
		e.AutoFocus()
		m.first = false
	}
	hl := (e.Hovered() || e.FocusVisible()) && !o.Disabled
	text, muted := cl.Text, cl.Muted
	switch {
	case o.Danger && hl:
		e.Background(Mix(ui.Color{}, cl.Up, 0.14))
		text, muted = cl.Up, cl.Up
	case o.Danger:
		text, muted = cl.Up, cl.Up
	case hl:
		e.LinearGradient(t.accentFill())
		text, muted = cl.OnAccent, cl.OnAccentMuted
	}
	if e.Pressed() && !o.Disabled {
		e.Scale(0.98, 0.98)
	}
	if o.Disabled {
		e.Opacity(0.4).Cursor(ui.CursorDefault)
	}
	e.TextColor(text).Children(func() {
		if o.Checkable {
			checkMark(c, 14, IconStroke, o.Checked, false, true, text)
		}
		if o.Icon != nil {
			ui.Icon(c, o.Icon).TextColor(muted)
		}
		ui.Text(c, label).Grow(1).NoWrap()
		if o.Shortcut != "" {
			ui.Text(c, o.Shortcut).TextColor(muted).FontSize(FsXs).Padding(0, 0, 0, 16).NoWrap()
		}
	})
	if e.Clicked() && !o.Disabled {
		*m.open = false
		return true
	}
	return false
}

// Separator adds a groove between groups of items.
func (m *MenuList) Separator() {
	cl := m.t.Colors
	ui.Box(m.c).Height(2).Margin(4, 8).InsetShadow(0, 1, 0, 0, cl.SepDark).InsetShadow(0, -1, 0, 0, cl.SepLight)
}

// Label adds a small, muted heading over a group of items.
func (m *MenuList) Label(text string) {
	ui.Text(m.c, text).Padding(6, 10, 2, 10).FontSize(FsXs).FontWeight(500).TextColor(m.t.Colors.Muted)
}

// menuMaxH is the most a menu's panel grows to: a longer menu scrolls.
const menuMaxH = 440

// Menu shows a list of actions below anchor while *open is true: the
// arrows move between the items, Enter or a click chooses one, Escape or
// a press outside closes it.
//
//	ryugo.Menu(c.Key("more"), b, &app.menu, func(m *ryugo.MenuList) {
//		if m.Item("Rename", ryugo.ItemOptions{Shortcut: "⌘R"}) {
//			app.rename()
//		}
//		m.Separator()
//		if m.Item("Delete", ryugo.ItemOptions{Danger: true}) {
//			app.delete()
//		}
//	})
func Menu(c *ui.Context, anchor ui.Element, open *bool, build func(m *MenuList)) ui.Element {
	t := From(c)
	return ui.PopoverBase(c, anchor, open, func(panel ui.Element) {
		popover(t, panel)
		panel.Role(ui.RoleMenu).MinWidth(max(200, anchor.Bounds().W)).Gap(0).FocusGroup(ui.Vertical)
		// past its height the menu scrolls inside its panel, never out of it
		panel.Children(func() {
			t.scrollList(c, menuMaxH-2*PadPop, func() {
				build(&MenuList{c: c, t: t, open: open, first: true})
			})
		})
	})
}

// MenuButton is a Button that opens a Menu below itself.
func MenuButton(c *ui.Context, label string, o ButtonOptions, build func(m *MenuList)) ui.Element {
	b := Button(c, label, o)
	open := ui.Local(b, "menu", func() bool { return false })
	if b.Clicked() {
		*open = !*open
	}
	Menu(c, b, open, build)
	return b
}

// Tooltip shows text above anchor while the pointer rests on it or it has
// the keyboard focus: a small raised label, lit from above as a button is,
// that rises 3px as it fades in and follows the pointer along anchor,
// lagging a little behind it. It lingers a moment after the pointer leaves
// (but goes at once when the pointer comes onto it), and going from one
// element to the next it doesn't close and open again: the new one slides
// from where the last one was. Give anchor a Description too: assistive
// technology does not read tooltips.
func Tooltip(c *ui.Context, anchor ui.Element, text string) ui.Element {
	t := From(c)
	now := c.Now()
	trail := ui.Local(c.Root(), "ryugo-tip", func() *tipTrail { return &tipTrail{} })
	tr := *trail
	me := ui.Local(anchor, "ryugo-tip", func() tipOwn { return tipOwn{} })
	if anchor.Pressed() && (me.live || now.Sub(me.seen) < tipLinger) {
		// a press hides it, as MyGo's does: at once, without lingering
		me.seen, me.live = now.Add(-tipLinger), false
	}
	tip := ui.TooltipBase(c, anchor, func(tip ui.Element) {
		// its own key, not its place among what the overlay layer holds: the
		// last one's lingering copy leaving that layer, or anything else
		// coming or going there, would make it a new element for a frame,
		// unplaced, and so drawn invisible
		tip.Key(tipKey{me})
		tip.Margin(0, 0, 8, 0)
		t.tipLook(tip)
		if tr.owner != me {
			// the last one holds where it is until this one is placed, if it
			// still shows: one gone stays gone
			tr.from = nil
			if tr.owner != nil && tr.owner.live || now.Sub(tr.at) < tipHeld {
				tr.from = tr.owner
			}
			tr.owner, tr.placed = me, false
		}
		me.seen, me.live = now, true
		tb := tip.LayoutBounds()
		if tb.W == 0 {
			// placed in the next frame
			c.AnimationFrame()
		}
		// it follows the pointer along anchor, its middle never past
		// anchor's ends nor the tooltip past the window's
		var want float32
		if px, _, ok := anchor.PointerPosition(); ok {
			ab, win := anchor.LayoutBounds(), c.Root().LayoutBounds()
			want = min(max(px-ab.W/2, -ab.W/2), ab.W/2)
			if tb.W > 0 && win.W > 0 {
				want = min(max(want, win.X+4-tb.X), win.X+win.W-4-tb.X-tb.W)
			}
		}
		fx := tip.AnimateTo("follow", want, springFollow)
		// going from one element to the next, a tooltip doesn't close and open
		// again: the new one comes from where the last one was, sliding
		slide := ui.Local(tip, "slide", func() int8 {
			if tr.r.W > 0 && now.Sub(tr.at) < tipSlideWithin {
				return 1
			}
			return -1
		})
		var dx, dy float32
		if *slide >= 1 {
			switch {
			case tb.W == 0:
				// placed in the next frame: from there it slides, while the
				// last one holds where it was
				tip.Opacity(0)
			case *slide == 1:
				*slide = 2
				tip.Jump("dx", tr.r.X+tr.r.W/2-tb.X-tb.W/2-fx)
				tip.Jump("dy", tr.r.Y+tr.r.H/2-tb.Y-tb.H/2)
			}
			dx, dy = tip.AnimateTo("dx", 0, *springPop), tip.AnimateTo("dy", 0, *springPop)
		} else {
			tip.TransformOrigin(0.5, 1).Transition(ui.ElementTransition{
				Duration: 120 * time.Millisecond, Ease: ui.EaseOut, Enter: &ui.Motion{Y: 3, Scale: 0.97},
			})
		}
		tip.Translate(dx+fx, dy)
		if tb.W > 0 && *slide != 1 {
			// where it shows, which the next one slides from and a lingering
			// copy holds
			at := ui.Rect{X: tb.X + dx + fx, Y: tb.Y + dy, W: tb.W, H: tb.H}
			tr.r, tr.at, tr.placed = at, now, true
			me.r = at
		}
		t.tipText(c, tip, text)
	})
	if tip == (ui.Element{}) && me.live {
		// it showed until this frame, however long ago the last one was: it
		// lingers from now
		me.seen, me.live = now, false
		if px, py, ok := c.Root().PointerPosition(); ok && px >= me.r.X && px < me.r.X+me.r.W && py >= me.r.Y && py < me.r.Y+me.r.H {
			// left for the tooltip itself: it goes now, as the ghost below
			// does when the pointer comes onto it
			me.seen = now.Add(-tipLinger)
		}
	}
	if tip != (ui.Element{}) || me.r.W == 0 {
		return tip
	}
	// gone from under the pointer, it lingers where it was, then fades;
	// taken over by the next one, it holds until that one is placed
	var show float32
	switch since := now.Sub(me.seen); {
	case tr.from == me && !tr.placed:
		show = 1
	case tr.owner != me:
	case since < tipLinger:
		show = 1
		c.After(tipLinger - since)
	case since < tipLinger+tipFade:
		show = 1 - float32(since-tipLinger)/float32(tipFade)
		c.AnimationFrame()
	}
	if show <= 0 {
		return tip
	}
	if tr.owner == me {
		tr.at = now
	}
	r := me.r
	ui.Overlay(c, func() {
		ghost := ui.Box(c.Key(me)).Absolute().Left(r.X).Top(r.Y).Width(r.W).Height(r.H).PassThrough()
		// the pointer come onto it, it doesn't stay in the way: it fades
		// now, without lingering
		if _, _, over := ghost.PointerPosition(); over && now.Sub(me.seen) < tipLinger {
			me.seen = now.Add(-tipLinger)
			c.AnimationFrame()
		}
		t.tipLook(ghost)
		k := 1 - show
		ghost.Opacity(show).TransformOrigin(0.5, 1).Translate(0, 2*k).Scale(1-0.02*k, 1-0.02*k)
		t.tipText(c, ghost, text)
	})
	return tip
}

// tipLook is a tooltip's surface: a button's, lit from above, lifted a
// little off the page. It takes no pointer, its words neither: what is
// under it is hovered, and its tooltip shows.
func (t *Tokens) tipLook(tip ui.Element) {
	cl := t.Colors
	tip.MaxWidth(260).Padding(6, 10).Radius(t.R(RadiusItem)).IgnoreInput(true)
	t.raised(tip, t.controlGradient(cl.ControlTop, cl.ControlBottom), cl.ControlBorder,
		t.Shadows.ControlShadow.With(t.Shadows.ShadowRaised...), raisedState{})
}

// tipText is a tooltip's words.
func (t *Tokens) tipText(c *ui.Context, tip ui.Element, text string) {
	tip.Children(func() {
		ui.Text(c, text).TextColor(t.Colors.Text).FontSize(FsXs).FontWeight(500).LineHeight(1.4).TextAlign(ui.Center).Wrap()
	})
}

// tipTrail is a window's last tooltip: where it showed and when, whose it
// is, and whose it was before (until it is placed).
type tipTrail struct {
	r           ui.Rect
	at          time.Time
	owner, from *tipOwn
	placed      bool
}

// tipKey is the key of an element's tooltip in the overlay layer, apart
// from that of its lingering copy (keyed by the tipOwn itself).
type tipKey struct{ own *tipOwn }

// tipOwn is where an element's tooltip last showed, and when; live while
// it shows as itself, not lingering. A tooltip at rest draws no frames, so
// seen can be long past while it still shows.
type tipOwn struct {
	r    ui.Rect
	seen time.Time
	live bool
}

const (
	// tipLinger is how long a tooltip stays after the pointer leaves its
	// element, and tipFade how long it then takes to go.
	tipLinger = 350 * time.Millisecond
	tipFade   = 120 * time.Millisecond
	// tipSlideWithin is how soon after the last tooltip showed a new one
	// slides from it, rather than coming in where it is.
	tipSlideWithin = 900 * time.Millisecond
	// tipHeld is how lately a lingering tooltip must have shown to hold
	// where it is while the next one is placed: a frame or two. One that
	// shows as itself holds however long ago its last frame was.
	tipHeld = 50 * time.Millisecond
)

// Toast types: the tone of the toast's mark and of the tint in its corner.
// Show a toast with MyGo's Context.AddToast, giving Type one of these.
//
//	c.AddToast(ui.Toast{Title: "Saved", Type: ryugo.ToastSuccess})
//
// A ToastLoading toast shows a spinner in place of its mark; give it a
// Timeout of -1 and, when the work is done, add a toast of the same ID and
// another type: it turns into it where it is.
const (
	ToastInfo    = ""
	ToastSuccess = "success"
	ToastWarning = "warning"
	ToastError   = "error"
	ToastLoading = "loading"
)

// Toaster draws the window's toasts in its bottom right corner: frosted
// glass with the pop shadow, a tint of their tone in the top left corner, a
// dot of the tone (a spinner while loading), the title and description, an
// action and a close button. As one arrives it rises from below the
// stack, growing to its size, and Ryugo's border beam lights its border
// in its tone, two lights running half way round from opposite sides. The newest is in front, at the bottom, with the older ones
// tucked behind it until the pointer comes onto them (see stack). A toast
// swiped to the end side goes, flung on with the swipe; let go short of
// that, it springs back. Call it once in the view, wherever.
func Toaster(c *ui.Context) ui.Element {
	t := From(c)
	return ui.ToastViewportBase(c, func(viewport ui.Element, toasts []ui.Toast) {
		viewport.Padding(16).AlignItems(ui.End).Justify(ui.End)
		// the stack is as tall as the toasts it shows, so that the pointer
		// on it, between two spread toasts too, keeps it spread
		box := ui.Box(c).Width(380).MaxWidthPercent(100)
		open := box.Hovered() || box.FocusWithin()
		box.Children(func() {
			// the toasts first, to know whether the pointer is on one
			parts := make([]ui.ToastParts, len(toasts))
			roots := make([]ui.Element, len(toasts))
			for i, tt := range toasts {
				parts[i] = ui.ToastBase(c.Key(tt.ID), tt)
				roots[i] = parts[i].Root
				open = open || roots[i].Hovered() || roots[i].FocusWithin()
			}
			xs := make([]float32, len(toasts))
			for i, tt := range toasts {
				xs[i] = t.toast(c, parts[i], tt, open || i == len(toasts)-1)
			}
			stack(box, roots, xs, open)
		})
	})
}

// toast styles and builds one toast; shown is whether its content shows
// (a toast tucked behind shows only its glass). It returns how far it is
// swiped.
func (t *Tokens) toast(c *ui.Context, p ui.ToastParts, tt ui.Toast, shown bool) float32 {
	cl := t.Colors
	tone := t.toastTone(tt.Type)
	r := p.Root.Row().AlignItems(ui.Start).Gap(10).Padding(12, 10, 14, 14)

	// a new type, as a loading toast done: the beam turns again
	kind := ui.Local(r, "type", func() string { return tt.Type })
	if seen := ui.Local(r, "beam-seen", func() bool { return false }); !*seen || *kind != tt.Type {
		*seen = true
		r.Jump("beam", 0)
	}
	*kind = tt.Type
	beam := r.AnimateTo("beam", 1, ui.Animation{Duration: 3000 * time.Millisecond, Delay: 260 * time.Millisecond, Ease: ui.EaseInOut})

	swiped, gone := t.swipe(c, r, tt.ID)

	// frosted glass: the surface at 86% over what is behind, blurred, with
	// a tint of the tone spreading from the top left corner
	tint := Mix(ui.Color{}, tone, 0.12)
	ring, halo, beaming := t.beam(beam, tone)
	r.Radius(t.R(RadiusPop)).Material(glass.Blur{Radius: 20}).Draw(func(pt *ui.Painter, b ui.Rect) {
		if beaming {
			// the beam's glow under the glass: only what spills past its
			// edge shows, as behind Ryugo's opaque cards
			for _, w := range []float32{18, 11, 5} {
				pt.StrokePathConicGradient(roundBorder(b, t.R(RadiusPop)), w, halo)
			}
		}
		pt.Fill(b, cl.Surface.Alpha(0.86), t.R(RadiusPop))
		pt.FillRadialGradient(b, ui.RadialGradient{Center: ui.Point{X: 0.001, Y: 0.001}, Radius: b.W * 0.7, Oklab: true,
			Stops: []ui.GradientStop{{Color: tint, At: 0}, {Color: tint.Alpha(0), At: 1}}}, t.R(RadiusPop))
	})
	vis := r.AnimateTo("shown", b2f(shown), stackSpring)
	if beaming {
		r.DrawOver(func(pt *ui.Painter, b ui.Rect) {
			b = ui.Rect{X: b.X + 0.5, Y: b.Y + 0.5, W: b.W - 1, H: b.H - 1}
			pt.StrokePathConicGradient(roundBorder(b, t.R(RadiusPop)-0.5), 1, ring)
		})
	}
	t.Shadows.ShadowPop.Apply(r)
	// it leaves sinking and shrinking a little as it fades, from wherever
	// the stack had it; swiped, it goes on the way it was flung
	// swiped away, it is out of sight already as it closes
	var exit *ui.Motion
	if !gone {
		exit = &ui.Motion{Y: 16, Scale: 0.94}
	}
	r.FontSize(FsMd).TextColor(cl.Text).Transition(ui.ElementTransition{
		Duration: 340 * time.Millisecond, Ease: ui.CubicBezier(0.32, 0.72, 0, 1), Colors: true, Exit: exit,
	})
	r.Children(func() {
		ui.Row(c).Grow(1).MinWidth(0).AlignItems(ui.Start).Gap(10).Opacity(vis).Children(func() { t.toastBody(c, p, tt, tone) })
	})
	return swiped
}

// toastTone is the tone of a toast's type.
func (t *Tokens) toastTone(kind string) ui.Color {
	cl := t.Colors
	switch kind {
	case ToastSuccess:
		return cl.Down
	case ToastWarning:
		return cl.Warn
	case ToastError:
		return cl.Up
	}
	return cl.Accent
}

// swipe lets a toast be swiped to the end side: it follows the pointer,
// fading as it goes; let go past a third of its width, or flung, it goes on
// past the end edge as it fades and then closes, else it springs back.
// Pulled the other way, it gives a little, as a rubber band. It returns how
// far the toast shows swiped, and whether it is on its way out.
func (t *Tokens) swipe(c *ui.Context, r ui.Element, id string) (float32, bool) {
	type swipeState struct {
		x          float32
		held, gone bool
	}
	s := *ui.Local(r, "swipe", func() *swipeState { return &swipeState{} })
	w := max(r.LayoutBounds().W, 1)
	end := w + 60
	dx, _, pressed := r.Dragged()
	vis := func(x float32) float32 {
		if x < 0 {
			return -rubberband(-x) * 0.4
		}
		return x
	}
	fade := func(x float32) { r.Opacity(1 - clamp01(x/end)) }
	switch {
	case s.gone:
		x := r.AnimateTo("swipe", end, ui.Animation{Duration: 260 * time.Millisecond, Ease: ui.EaseOut})
		fade(x)
		if x >= end-1 {
			c.CloseToast(id)
		}
		return x, true
	case pressed:
		s.held = true
		s.x += dx
		x := r.Jump("swipe", vis(s.x))
		fade(x * 0.55)
		return x, false
	case s.held:
		s.held = false
		vx := float32(0)
		if v, _, ok := r.DragEnded(); ok {
			vx = v
		}
		if s.x > w/3 || vx > 700 {
			s.gone = true
			c.AnimationFrame()
			return r.AnimateTo("swipe", end, ui.Animation{Duration: 260 * time.Millisecond, Ease: ui.EaseOut}), true
		}
		s.x = 0
	}
	x := r.AnimateTo("swipe", 0, ui.Animation{Spring: ui.SpringOptions{Duration: 420 * time.Millisecond, Bounce: 0.22}})
	fade(max(x, 0) * 0.55)
	return x, false
}

// beam is Ryugo's border beam, run once as a toast arrives (f from 0 to
// 1), as two lights: one rises by the toast's dot, at the middle of its
// start edge, and runs along the top to the end edge; the other rises
// opposite it and runs along the bottom to the start edge. Each goes half
// way round, slowly, and they fade as they meet the other's start. It
// returns the gradients of the 1px ring along the border, each light a
// long band fading at both ends to a bright head (in the dark theme, the
// tone lifted toward white), and of their glow, the same bands wider and
// fainter; ok is false while it does not run.
func (t *Tokens) beam(f float32, tone ui.Color) (ring, glow ui.ConicGradient, ok bool) {
	if f <= 0 || f >= 1 {
		return ring, glow, false
	}
	head := tone
	if t.Dark {
		head = Mix(tone, ui.RGB(255, 255, 255), 0.65)
	}
	vis := min(1, f*5) * min(1, (1-f)*4)
	// a band's head is 16% round from the angle the gradient starts at,
	// the other's half a turn on: the first starts at the start edge (270
	// degrees, by the dot), and both turn half a turn
	from := 270 - 0.16*360 + 180*f
	at := func(c ui.Color, a, pos float32) ui.GradientStop {
		return ui.GradientStop{Color: c.Alpha(a * vis), At: pos}
	}
	// one band's stops, around its head at h, for the ring and the glow
	ringBand := func(h float32) []ui.GradientStop {
		return []ui.GradientStop{
			at(tone, 0, h-0.12), at(tone, 0.12, h-0.09), at(tone, 0.4, h-0.048),
			at(head, 0.75, h-0.015), at(head, 0.95, h), at(head, 0.75, h+0.015),
			at(tone, 0.4, h+0.048), at(tone, 0.12, h+0.09), at(tone, 0, h+0.12),
		}
	}
	const g = 0.1
	glowBand := func(h float32) []ui.GradientStop {
		return []ui.GradientStop{
			at(tone, 0, h-0.16), at(tone, 0.3*g, h-0.08), at(head, 0.9*g, h),
			at(tone, 0.3*g, h+0.08), at(tone, 0, h+0.16),
		}
	}
	gap := func(a, b float32) []ui.GradientStop { return []ui.GradientStop{at(tone, 0, a), at(tone, 0, b)} }
	ring = ui.ConicGradient{Angle: from, Oklab: true, Stops: slices.Concat(
		gap(0, 0.02), ringBand(0.16), gap(0.3, 0.52), ringBand(0.66), gap(0.8, 1))}
	glow = ui.ConicGradient{Angle: from, Oklab: true, Stops: slices.Concat(
		glowBand(0.16), gap(0.34, 0.48), glowBand(0.66), gap(0.84, 1))}
	return ring, glow, true
}

// roundBorder is r's rounded border, clockwise from the middle of its
// start edge.
func roundBorder(r ui.Rect, radius float32) *ui.Path {
	rad := min(radius, r.W/2, r.H/2)
	x0, y0, x1, y1 := r.X, r.Y, r.X+r.W, r.Y+r.H
	var p ui.Path
	p.MoveTo(x0, y0+r.H/2).LineTo(x0, y0+rad).QuadTo(x0, y0, x0+rad, y0).
		LineTo(x1-rad, y0).QuadTo(x1, y0, x1, y0+rad).
		LineTo(x1, y1-rad).QuadTo(x1, y1, x1-rad, y1).
		LineTo(x0+rad, y1).QuadTo(x0, y1, x0, y1-rad).
		LineTo(x0, y0+r.H/2)
	return &p
}

// toastBody builds a toast's content: the tone's dot (a spinner while
// loading), the title and description, the action and the close button.
// When a loading toast turns into its result, the spinner fades as the dot
// opens in its place, and the new words rise in as the old ones rise away.
func (t *Tokens) toastBody(c *ui.Context, p ui.ToastParts, tt ui.Toast, tone ui.Color) {
	cl := t.Colors
	mark := ui.Box(c).Size(12, 12).Margin(4, 0, 0, -2).Shrink(0)
	done := mark.AnimateTo("done", b2f(tt.Type != ToastLoading), ui.Animation{Duration: 360 * time.Millisecond, Ease: ui.CubicBezier(0.2, 0.8, 0.2, 1)})
	mark.Children(func() {
		if done < 1 {
			Spinner(c.Key("spin"), SpinnerOptions{Size: 12, Color: cl.Accent}).Absolute().Left(0).Top(0).
				Opacity(1-done).Scale(1-0.4*done, 1-0.4*done)
		}
		if done > 0 {
			ui.Box(c.Key("dot")).Absolute().Left(2).Top(2).Size(8, 8).Radius(4).Background(tone).
				Scale(done, done).Opacity(clamp01(done * 2))
		}
	})
	words := ui.ElementTransition{Duration: 260 * time.Millisecond, Ease: ui.EaseOut,
		Enter: &ui.Motion{Y: 6}, Exit: &ui.Motion{Y: -6}}
	ui.Column(c).Grow(1).MinWidth(0).Gap(2).Children(func() {
		ui.Text(c.Key("t:"+tt.Title), tt.Title).FontWeight(600).Wrap().Transition(words)
		if tt.Description != "" {
			ui.Text(c.Key("d:"+tt.Description), tt.Description).TextColor(cl.Muted).FontSize(FsSm).Wrap().Transition(words)
		}
	})
	if tt.Action != "" {
		a := p.ActionButton().Center().Height(26).Padding(0, 10).Radius(RadiusPill).Cursor(ui.CursorPointer).FocusRing(false)
		s := raisedState{hovered: a.Hovered(), pressed: a.Pressed(), focus: a.FocusVisible()}
		t.raised(a, t.controlGradient(cl.ControlTop, cl.ControlBottom), cl.ControlBorder, t.Shadows.ControlShadow, s)
		a.TextColor(cl.Text).FontSize(FsSm).FontWeight(600).Children(func() { ui.Text(c, tt.Action).NoWrap() })
	}
	x := p.CloseButton().Label("Close").Center().Size(24, 24).Radius(RadiusPill).Cursor(ui.CursorPointer).FocusRing(false)
	color := cl.Muted
	if x.Hovered() {
		x.Background(cl.Fill)
		color = cl.Text
	}
	if x.Pressed() {
		x.Scale(0.9, 0.9)
	}
	x.TextColor(color).Children(func() { ui.Icon(c, iconClear) })
}

// stack lays the toasts out by hand, each pinned to the bottom of the
// stack's box and moved up from there, so that one arriving or leaving
// moves the others on their springs and never through the layout. The
// older toasts tuck behind the newest, each peeking 10 DIPs above the one
// in front, 5% smaller and no taller than it; with the pointer on the
// stack or the keyboard focus in a toast, they spread out into a list, 8
// DIPs apart. The box grows and shrinks with them. A new toast rises from
// below the stack, growing from 96% as it fades in; xs are how far each is
// swiped.
func stack(box ui.Element, roots []ui.Element, xs []float32, open bool) {
	const peek, gap, guess = 10, 8, 64
	n := len(roots)
	if n == 0 {
		box.Height(box.AnimateTo("h", 0, stackSpring))
		return
	}
	// the heights the last frame laid out; a new toast's is a guess, for a frame
	hs := make([]float32, n)
	for i, r := range roots {
		if hs[i] = r.LayoutBounds().H; hs[i] == 0 {
			hs[i] = guess
		}
	}
	front := hs[n-1]
	h := front + float32(n-1)*peek
	if open {
		h = float32(n-1) * gap
		for _, v := range hs {
			h += v
		}
	}
	box.Height(box.AnimateTo("h", h, stackSpring))
	below := float32(0) // how tall the spread toasts in front are
	for i := n - 1; i >= 0; i-- {
		r, d, b := roots[i], float32(n-1-i), hs[i]
		y, sx, sy := -below, float32(1), float32(1)
		if !open && d > 0 {
			sx = 1 - 0.05*d
			sy = sx * min(1, front/b)
			// its top d peeks above the front toast's top
			y = b*sy - front - d*peek
		}
		below += b + gap
		seen := ui.Local(r, "seen", func() bool { return false })
		if !*seen {
			*seen = true
			r.Jump("in", 0)
		}
		in := r.AnimateTo("in", 1, toastIn)
		y = r.AnimateTo("stack-y", y, stackSpring) + (1-in)*(b+20)
		grow := 0.96 + 0.04*in
		sx = r.AnimateTo("stack-sx", sx, stackSpring) * grow
		sy = r.AnimateTo("stack-sy", sy, stackSpring) * grow
		r.Absolute().Left(0).Right(0).Bottom(0).TransformOrigin(0.5, 1).Translate(xs[i], y).Scale(sx, sy)
		if in < 1 {
			r.Opacity(clamp01(in * 1.6))
		}
	}
}

var (
	// a toast arrives unhurried, with a soft settle; the stack moves at a
	// calm pace as it makes room
	stackSpring = ui.Animation{Spring: ui.SpringOptions{Duration: 560 * time.Millisecond, Bounce: 0.08}}
	toastIn     = ui.Animation{Spring: ui.SpringOptions{Duration: 700 * time.Millisecond, Bounce: 0.14}}
)
