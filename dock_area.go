package mizan

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// An area: the card surface, a header (the kind's menu, the pane's title,
// the maximize button) over the pane. An area holds the keyboard focus as
// a whole: a press anywhere in it that no control takes gives it the focus,
// and Ctrl+Alt+arrows move it to the next area that way. Ctrl+Space, or
// the header's button, maximizes the area under the pointer (else the one
// focused) over the window, growing it from where it was, and brings it
// back.

var (
	iconMaximize = Svg(16, 13, "M3 6.5V3h3.5M9.5 3H13v3.5M13 9.5V13H9.5M6.5 13H3V9.5")
	iconRestore  = Svg(16, 13, "M6.5 3v3.5H3M13 6.5H9.5V3M9.5 13V9.5H13M3 9.5h3.5V13")
)

// headerH is the height of an area's header.
const headerH = 28

// area builds area n in the tree. Maximized, it keeps its place as a bare
// surface, its content showing over the window (maximized).
func (d *dockView) area(n *Node) ui.Element {
	st := d.st
	a := d.surface(ui.Column(d.c.Key(n.ID)))
	st.areas[n.ID] = a.LayoutBounds()
	if a.Hovered() {
		st.hovered = n.ID
	}
	if f := st.fold; f != nil && f.gone == n.ID {
		d.folding(a)
	}
	// an area a split made: its surface was there, the area's it came out
	// of; what it shows fades in
	if st.born[n.ID] {
		delete(st.born, n.ID)
		a.Jump("in", 0)
	}
	in := a.AnimateTo("in", 1, ui.Animation{Duration: 240 * time.Millisecond, Ease: ui.EaseOut})
	if d.maxed == n.ID {
		return a
	}
	d.content(n, a, in, true)
	return a
}

// surface makes a an area's surface: the card, its hairline drawn over
// what it holds, and the keyboard focus.
func (d *dockView) surface(a ui.Element) ui.Element {
	t := d.t
	cl := t.Colors
	R := t.R(RadiusControl)
	return a.MinWidth(0).MinHeight(0).Radius(R).Background(cl.Card).Clip().Focusable().FocusRing(false).
		DrawOver(func(p *ui.Painter, r ui.Rect) { p.Stroke(r, cl.Hairline, R, 1) })
}

// content builds what area n shows into a: its header and its pane, at in
// (fading in); corners puts the corners that split and join on it.
func (d *dockView) content(n *Node, a ui.Element, in float32, corners bool) {
	c, st, t := d.c, d.st, d.t
	cl := t.Colors
	if st.focusTo == n.ID {
		st.focusTo = ""
		a.Focus()
	}
	active := a.FocusWithin()
	if active {
		st.focused = n.ID
	}
	a.Children(func() {
		if corners {
			defer func() {
				for k := range 4 {
					d.corner(n, k)
				}
			}()
		}
		if len(n.Panes) == 0 {
			return
		}
		p := n.Panes[min(n.Active, len(n.Panes)-1)]
		kind := d.o.kind(p.Kind)
		h := ui.Row(c.Key("header")).Height(headerH).Padding(0, 4).Gap(2).AlignItems(ui.Center).Shrink(0).
			InsetShadow(0, -1, 0, 0, cl.Hairline.Alpha(in)).Opacity(in)
		h.Children(func() {
			d.kindMenu(n, p, kind, active)
			if kind == nil || p.Title != kind.Title {
				ui.Text(c, p.Title).FontSize(FsSm).TextColor(cl.Muted).SingleLine().Ellipsis("…").Shrink(1).MinWidth(0).Padding(0, 4)
			}
			ui.Box(c).Grow(1)
			d.maxButton(n, h.Hovered())
		})
		body := ui.Column(c.Key(p.ID)).Grow(1).MinHeight(0).MinWidth(0).Opacity(in)
		body.Children(func() {
			if kind != nil && kind.Build != nil {
				kind.Build(c, p)
			}
		})
	})
}

// headerButton is a quiet button in an area's header: a fill under the
// pointer, a little shrink when pressed, its corner concentric with the
// area's.
func (d *dockView) headerButton(b ui.Element) ui.Element {
	t := d.t
	cl := t.Colors
	b.Height(headerH - 8).Radius(Concentric(t.R(RadiusControl), 4)).Cursor(ui.CursorPointer).FocusRing(false).Shrink(0)
	hot := b.AnimateTo("hot", b2f(b.Hovered() || b.FocusVisible()), ui.Animation{Duration: 120 * time.Millisecond, Ease: ui.EaseOut})
	if hot > 0.01 {
		b.Background(cl.Fill.Alpha(hot))
	}
	press := b.AnimateTo("press", b2f(b.Pressed()), springSqueeze)
	if press > 0.001 {
		b.Scale(1-0.04*press, 1-0.04*press)
	}
	if b.FocusVisible() {
		t.Shadows.FocusRing.Apply(b)
	}
	return b
}

// kindMenu is the header's menu of what the area shows: the kind's title
// and a chevron; a kind chosen replaces the pane with a new one of it
// (Undo brings the old one back).
func (d *dockView) kindMenu(n *Node, p *Pane, kind *DockKind, active bool) {
	c, t := d.c, d.t
	cl := t.Colors
	title := p.Kind
	if kind != nil {
		title = kind.Title
	}
	b := d.headerButton(ui.ButtonBase(c.Key("kind")).Label(title).Padding(0, 6).Gap(4))
	open := ui.Local(b, "menu", func() bool { return false })
	if b.Clicked() {
		*open = !*open
	}
	on := b.AnimateTo("active", b2f(active), ui.Animation{Duration: 160 * time.Millisecond, Ease: ui.EaseOut})
	b.Children(func() {
		ui.Text(c, title).FontSize(FsSm).FontWeight(500).TextColor(Mix(cl.Text2, cl.Text, on)).SingleLine()
		ui.Icon(c, iconChevronEnd).Size(10, 10).TextColor(cl.Muted).Rotate(90 + 180*b2f(*open))
	})
	Menu(c, b, open, func(m *MenuList) {
		for _, k := range d.o.Kinds {
			if m.Item(k.Title, ItemOptions{Checkable: true, Checked: k.Kind == p.Kind}) && k.Kind != p.Kind {
				d.l.remember()
				*p = Pane{Kind: k.Kind, Title: k.Title}
				d.l.fix()
				d.changed()
			}
		}
	})
}

// maxButton is the header's button that maximizes area n, and brings it
// back; it shows while the pointer is over the header, or the area is
// maximized.
func (d *dockView) maxButton(n *Node, over bool) {
	c := d.c
	maxed := d.w.Max == n.ID
	label := "Maximize"
	icon := iconMaximize
	if maxed {
		label, icon = "Restore", iconRestore
	}
	b := d.headerButton(ui.ButtonBase(c.Key("max")).Label(label).Width(headerH - 8).Center())
	shown := b.AnimateTo("shown", b2f(over || maxed || b.FocusVisible()), ui.Animation{Duration: 160 * time.Millisecond, Ease: ui.EaseOut})
	b.Opacity(shown)
	if b.Clicked() {
		d.toggleMax(n.ID)
	}
	b.Children(func() {
		ui.Icon(c, icon).TextColor(d.t.Colors.Muted)
	})
}

// toggleMax maximizes area id, or brings back the area maximized.
func (d *dockView) toggleMax(id string) {
	if d.w.Max != "" {
		d.w.Max = ""
	} else if id != "" {
		d.w.Max = id
	}
	d.changed()
	d.c.AnimationFrame()
}

// maximized builds the area maximized over the window: it grows from its
// place in the tree to the whole of it, and shrinks back there.
func (d *dockView) maximized() {
	n := d.l.Find(d.maxed)
	at := d.root.LayoutBounds()
	full := ui.Rect{W: at.W, H: at.H}
	from := full
	if r, ok := d.st.areas[n.ID]; ok && r.W > 0 {
		from = ui.Rect{X: r.X - at.X, Y: r.Y - at.Y, W: r.W, H: r.H}
	}
	m := d.max
	r := ui.Rect{
		X: from.X + (full.X-from.X)*m, Y: from.Y + (full.Y-from.Y)*m,
		W: from.W + (full.W-from.W)*m, H: from.H + (full.H-from.H)*m,
	}
	a := d.surface(ui.Column(d.c.Key("max"))).Absolute().Left(r.X).Top(r.Y).Width(r.W).Height(r.H)
	// it lifts while it moves, as a card held does, and lies flat again
	if lift := 4 * m * (1 - m); lift > 0.01 {
		for _, l := range d.t.Shadows.ShadowPop {
			if !l.Inset {
				a.Shadow(l.X, l.Y, l.Blur, l.Spread, l.Color.Alpha(min(lift, 1)))
			}
		}
	}
	d.content(n, a, 1, false)
}

// step moves the keyboard focus from the area focused (or under the
// pointer) to the nearest one in direction (dx, dy).
func (d *dockView) step(dx, dy float32) {
	st := d.st
	from := or(st.focused, st.hovered)
	a, ok := st.areas[from]
	if !ok {
		return
	}
	best, score := "", float32(0)
	for id, r := range st.areas {
		if id == from {
			continue
		}
		// how far it is that way, and how far off the line
		var ahead, off float32
		switch {
		case dx > 0:
			ahead, off = r.X-(a.X+a.W), abs32(r.Y+r.H/2-(a.Y+a.H/2))
		case dx < 0:
			ahead, off = a.X-(r.X+r.W), abs32(r.Y+r.H/2-(a.Y+a.H/2))
		case dy > 0:
			ahead, off = r.Y-(a.Y+a.H), abs32(r.X+r.W/2-(a.X+a.W/2))
		default:
			ahead, off = a.Y-(r.Y+r.H), abs32(r.X+r.W/2-(a.X+a.W/2))
		}
		if ahead < -1 {
			continue
		}
		if s := ahead + off/2; best == "" || s < score {
			best, score = id, s
		}
	}
	if best != "" {
		st.focusTo = best
		d.c.AnimationFrame()
	}
}

// keys handles the Dock's keys: Ctrl+Space maximizes and brings back,
// Ctrl+Alt+arrows move the focus between areas.
func (d *dockView) keys() {
	c := d.c
	if c.Shortcut(ui.Ctrl, ui.KeySpace) {
		d.toggleMax(or(d.st.hovered, d.st.focused))
	}
	if d.w.Max != "" {
		return
	}
	for _, k := range []struct {
		key    ui.Key
		dx, dy float32
	}{{ui.KeyLeft, -1, 0}, {ui.KeyRight, 1, 0}, {ui.KeyUp, 0, -1}, {ui.KeyDown, 0, 1}} {
		if c.Shortcut(ui.Ctrl|ui.Alt, k.key) {
			d.step(k.dx, k.dy)
		}
	}
}
