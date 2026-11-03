package mizan

import (
	"time"

	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"
)

// Layers float above the page: dialogs and sheets on the card surface in a
// halo of progressive blur, popovers, menus and toasts on frosted glass. They move
// on springs, quick to arrive and quick to leave, and content scrolling in
// a dialog or a sheet dissolves into its header and footer under a
// progressive blur.

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
}

// The springs of the layers: they settle fast, with a hint of overshoot.
var (
	springPanel = &ui.Animation{Spring: ui.SpringOptions{Duration: 300 * time.Millisecond, Bounce: 0.12}}
	springSheet = &ui.Animation{Spring: ui.SpringOptions{Duration: 380 * time.Millisecond, Bounce: 0.06}}
	springPop   = &ui.Animation{Spring: ui.SpringOptions{Duration: 220 * time.Millisecond, Bounce: 0.1}}
	springRest  = ui.Animation{Spring: ui.SpringOptions{Duration: 420 * time.Millisecond, Bounce: 0.18}}
	springTilt  = ui.Animation{Spring: ui.SpringOptions{Duration: 240 * time.Millisecond}}
	springToast = &ui.Animation{Spring: ui.SpringOptions{Duration: 360 * time.Millisecond, Bounce: 0.16}}
)

// halo paints a progressive blur around a panel, in place of a scrim
// over the whole window: what is behind blurs most along the panel's edge
// and less and less farther out, to nothing 64 DIPs away, under a faint
// haze of the page's color. Rings of a light blur, each larger than the
// last and rounded as the panel is, add up where they overlap, as blurs
// do (their variances add), so the blur falls off smoothly and follows
// the panel's corners. The halo is painted by the panel, so it moves,
// tilts and fades with it.
func (t *Tokens) halo(panel ui.Element, radius float32) {
	const rings, step, blur = 8, 8, 1.8
	haze := t.Colors.Bg.Alpha(0.05)
	panel.Draw(func(p *ui.Painter, r ui.Rect) {
		for i := rings; i >= 1; i-- {
			k := float32(i * step)
			ring := ui.Rect{X: r.X - k, Y: r.Y - k, W: r.W + 2*k, H: r.H + 2*k}
			rr := radius + k
			glass.Blur{Radius: blur}.PaintMaterial(p, ring, [4]float32{rr, rr, rr, rr})
			p.Fill(ring, haze, rr)
		}
	})
}

// Dialog shows a modal panel while *open is true: a card with the pop
// shadow and a halo of progressive blur around it (the page around it is
// not dimmed), that springs in from 94% and 10px below. Its
// header drags it around the window (unless Fixed): it tilts with the
// pointer's speed, a throw glides on, and it springs back inside the
// window when let go; a double click on the header brings it back to the
// middle. Escape, a click on the scrim or the close button set *open to
// false, which the returned panel's Dismissed reports. It returns the
// panel, or an empty element while closed.
//
//	mizan.Dialog(c.Key("delete"), &app.confirm, mizan.DialogOptions{
//		Title: "Delete the report?",
//		Footer: func() {
//			mizan.Button(c.Key("cancel"), "Cancel", mizan.ButtonOptions{})
//			mizan.Button(c.Key("ok"), "Delete", mizan.ButtonOptions{Danger: true})
//		},
//	}, func() { ui.Text(c, "This cannot be undone.") })
func Dialog(c *ui.Context, open *bool, o DialogOptions, body func()) ui.Element {
	t := From(c)
	return ui.DialogBase(c, open, func(backdrop, panel ui.Element) {
		backdrop.Padding(24, 16)
		panel.Width(or(o.Width, 480)).MaxWidthPercent(100).MaxHeight(720)
		t.halo(panel, RadiusLg)
		panel.Transition(ui.ElementTransition{
			Animation: springPanel,
			Enter:     &ui.Motion{Scale: 0.94, Y: 10},
			Exit:      &ui.Motion{Scale: 0.97, Y: 6},
		})
		head := t.panelContent(c, open, o, body, false, RadiusLg)
		if !o.Fixed {
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
// with the halo of progressive blur along its inner edge, as Dialog's: it
// springs in from past the edge, tilted a little, and straightens as it
// lands. It closes as Dialog does.
func Sheet(c *ui.Context, open *bool, o DialogOptions, body func()) ui.Element {
	t := From(c)
	w := or(o.Width, 420)
	return ui.DialogBase(c, open, func(backdrop, panel ui.Element) {
		backdrop.AlignItems(ui.End).Justify(ui.Stretch)
		panel.Width(w).MaxWidthPercent(100).FillHeight()
		t.halo(panel, 0)
		panel.TransformOrigin(0, 0.5).Transition(ui.ElementTransition{
			Animation: springSheet,
			Enter:     &ui.Motion{X: w + 24, Rotate: 2, Opacity: 1},
			Exit:      &ui.Motion{X: w + 24, Rotate: 2, Opacity: 1},
		})
		t.panelContent(c, open, o, body, true, 0)
	})
}

// layer paints a floating panel: the card surface, the pop shadow.
func (t *Tokens) layer(e ui.Element, radius float32) {
	e.Radius(radius).Background(t.Colors.Card).TextColor(t.Colors.Text).Clip()
	t.Shadows.ShadowPop.Apply(e)
}

// panelContent builds a dialog's or a sheet's card, inside the panel that
// paints its halo: its header, scrolling body and footer. It returns the
// header.
func (t *Tokens) panelContent(c *ui.Context, open *bool, o DialogOptions, body func(), sheet bool, radius float32) ui.Element {
	cl := t.Colors
	var head ui.Element
	panel := ui.Column(c).FillWidth().Grow(1).MinHeight(0)
	t.layer(panel, radius)
	sc := ui.Local(panel, "scroll", func() ui.ScrollState { return ui.ScrollState{} })
	panel.Children(func() {
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
			// content scrolling past the body's edges dissolves into the
			// header and footer: blurred more and more, and faded into the
			// card, the nearer the edge
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
	return head
}

// scrollEdge is the progressive blur at a scroll body's top (or bottom)
// edge: it shows as the content goes under that edge, by how far there is
// to scroll on that side.
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
	e.Material(glass.Blur{Radius: 7, Mask: &ui.LinearGradient{From: ui.RGB(0, 0, 0), To: ui.Transparent, Angle: angle}})
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
//	b := mizan.Button(c.Key("filter"), "Filter", mizan.ButtonOptions{})
//	if b.Clicked() {
//		app.filtering = !app.filtering
//	}
//	mizan.Popover(c.Key("filters"), b, &app.filtering, func() { ... })
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
		MinHeight(34).Padding(0, 12, 0, 10).Radius(RadiusItem).FontSize(FsMd).Cursor(ui.CursorPointer).
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
			i := ui.Icon(c, iconCheck).Size(14, 14).TextColor(text)
			if !o.Checked {
				i.Invisible()
			}
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

// Menu shows a list of actions below anchor while *open is true: the
// arrows move between the items, Enter or a click chooses one, Escape or
// a press outside closes it.
//
//	mizan.Menu(c.Key("more"), b, &app.menu, func(m *mizan.MenuList) {
//		if m.Item("Rename", mizan.ItemOptions{Shortcut: "⌘R"}) {
//			app.rename()
//		}
//		m.Separator()
//		if m.Item("Delete", mizan.ItemOptions{Danger: true}) {
//			app.delete()
//		}
//	})
func Menu(c *ui.Context, anchor ui.Element, open *bool, build func(m *MenuList)) ui.Element {
	t := From(c)
	return ui.PopoverBase(c, anchor, open, func(panel ui.Element) {
		popover(t, panel)
		panel.Role(ui.RoleMenu).MinWidth(max(200, anchor.Bounds().W)).MaxHeight(360).Gap(0).FocusGroup(ui.Vertical)
		panel.Children(func() { build(&MenuList{c: c, t: t, open: open, first: true}) })
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
// the keyboard focus: a small label of dark frosted glass (light in the
// dark theme) that rises 3px as it fades in. Give anchor a Description too: assistive
// technology does not read tooltips.
func Tooltip(c *ui.Context, anchor ui.Element, text string) ui.Element {
	t := From(c)
	cl := t.Colors
	return ui.TooltipBase(c, anchor, func(tip ui.Element) {
		tip.Margin(0, 0, 8, 0).MaxWidth(260).Padding(6, 10)
		t.frost(tip, cl.Text.Alpha(0.9), RadiusItem, 12)
		tip.TransformOrigin(0.5, 1).Transition(ui.ElementTransition{
			Duration: 120 * time.Millisecond, Ease: ui.EaseOut,
			Enter: &ui.Motion{Y: 3, Scale: 0.97}, Exit: &ui.Motion{Y: 2, Scale: 0.98},
		})
		tip.Children(func() {
			ui.Text(c, text).TextColor(cl.Bg).FontSize(FsXs).FontWeight(500).LineHeight(1.4).TextAlign(ui.Center).Wrap()
		})
	})
}

// Toast types: the tone of the toast's mark and of the tint in its corner.
// Show a toast with MyGo's Context.AddToast, giving Type one of these.
//
//	c.AddToast(ui.Toast{Title: "Saved", Type: mizan.ToastSuccess})
const (
	ToastInfo    = ""
	ToastSuccess = "success"
	ToastWarning = "warning"
	ToastError   = "error"
)

// Toaster draws the window's toasts in its bottom right corner: frosted
// glass with the pop shadow, a tint of their tone in the top left corner, a dot of
// the tone, the title and description, an action and a close button. The
// newest is at the bottom. Call it once in the view, wherever.
func Toaster(c *ui.Context) ui.Element {
	t := From(c)
	cl := t.Colors
	return ui.ToastViewportBase(c, func(viewport ui.Element, toasts []ui.Toast) {
		viewport.Padding(16).AlignItems(ui.End).Justify(ui.End).Gap(8)
		for _, tt := range toasts {
			tone := cl.Accent
			switch tt.Type {
			case ToastSuccess:
				tone = cl.Down
			case ToastWarning:
				tone = cl.Warn
			case ToastError:
				tone = cl.Up
			}
			p := ui.ToastBase(c.Key(tt.ID), tt)
			r := p.Root.Row().Width(380).MaxWidthPercent(100).AlignItems(ui.Start).Gap(10).Padding(12, 10, 12, 14)
			// frosted glass: the surface at 86% over what is behind, blurred,
			// with a tint of the tone spreading from the top left corner
			glow := Mix(ui.Color{}, tone, 0.12)
			r.Radius(RadiusPop).Material(glass.Blur{Radius: 20}).Draw(func(p *ui.Painter, b ui.Rect) {
				p.Fill(b, cl.Surface.Alpha(0.86), RadiusPop)
				p.FillRadialGradient(b, ui.RadialGradient{Center: ui.Point{X: 0.001, Y: 0.001}, Radius: b.W * 0.7, Oklab: true,
					Stops: []ui.GradientStop{{Color: glow, At: 0}, {Color: glow.Alpha(0), At: 1}}}, RadiusPop)
			})
			t.Shadows.ShadowPop.Apply(r)
			r.FontSize(FsMd).TextColor(cl.Text).TransformOrigin(0.5, 1).Transition(ui.ElementTransition{
				Animation: springToast, Position: true,
				Enter: &ui.Motion{Y: 28, Scale: 0.96}, Exit: &ui.Motion{Y: 6, Scale: 0.95},
			})
			r.Children(func() {
				ui.Box(c).Size(8, 8).Radius(4).Background(tone).Margin(6, 0, 0, 0).Shrink(0)
				ui.Column(c).Grow(1).MinWidth(0).Gap(2).Children(func() {
					ui.Text(c, tt.Title).FontWeight(600).Wrap()
					if tt.Description != "" {
						ui.Text(c, tt.Description).TextColor(cl.Muted).FontSize(FsSm).Wrap()
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
				x.TextColor(color).Children(func() { ui.Icon(c, iconClear) })
			})
		}
	})
}
