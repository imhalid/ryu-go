package mizan

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// Layers float above the page on the card surface with the pop shadow:
// dialogs and sheets over a scrim, popovers and menus beside their anchor,
// toasts in the window's corner. MyGo has no backdrop blur, so the glass
// surfaces of the web library (pop, toast) are opaque here.

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
}

// Dialog shows a modal panel over a scrim while *open is true: a card
// with the pop shadow that comes in from 96% and 8px below. Escape, a
// click on the scrim or the close button set *open to false, which the
// returned panel's Dismissed reports. It returns the panel, or an empty
// element while closed.
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
		backdrop.Background(t.Colors.Scrim).Padding(24, 16).
			Transition(ui.ElementTransition{Duration: 280 * time.Millisecond, Enter: &ui.Motion{}, Exit: &ui.Motion{}})
		panel.Width(or(o.Width, 480)).MaxWidthPercent(100).MaxHeight(720)
		t.layer(panel, RadiusLg)
		panel.Transition(ui.ElementTransition{
			Duration: 280 * time.Millisecond, Ease: ui.EaseOut,
			Enter: &ui.Motion{Scale: 0.96, Y: 8}, Exit: &ui.Motion{Scale: 0.96, Y: 8},
		})
		t.panelContent(c, open, o, body, false)
	})
}

// Sheet shows a panel along the window's end edge, over a soft scrim,
// while *open is true: it slides in from past the edge, tilted a little,
// and straightens as it lands. It closes as Dialog does.
func Sheet(c *ui.Context, open *bool, o DialogOptions, body func()) ui.Element {
	t := From(c)
	w := or(o.Width, 420)
	return ui.DialogBase(c, open, func(backdrop, panel ui.Element) {
		backdrop.Background(t.Colors.ScrimSoft).AlignItems(ui.End).Justify(ui.Stretch).
			Transition(ui.ElementTransition{Duration: 240 * time.Millisecond, Enter: &ui.Motion{}, Exit: &ui.Motion{}})
		panel.Width(w).MaxWidthPercent(100).FillHeight()
		t.layer(panel, 0)
		panel.Transition(ui.ElementTransition{
			Duration: 360 * time.Millisecond, Ease: ui.EaseOut,
			Enter: &ui.Motion{X: w + 16, Rotate: 2, Opacity: 1}, Exit: &ui.Motion{X: w + 16, Rotate: 2, Opacity: 1},
		})
		t.panelContent(c, open, o, body, true)
	})
}

// layer paints a floating panel: the card surface, the pop shadow.
func (t *Tokens) layer(e ui.Element, radius float32) {
	e.Radius(radius).Background(t.Colors.Card).TextColor(t.Colors.Text).Clip()
	t.Shadows.ShadowPop.Apply(e)
}

// panelContent builds a dialog's or a sheet's header, scrolling body and
// footer.
func (t *Tokens) panelContent(c *ui.Context, open *bool, o DialogOptions, body func(), sheet bool) {
	cl := t.Colors
	panel := ui.Column(c).FillWidth().Grow(1).MinHeight(0)
	panel.Children(func() {
		head := ui.Row(c).AlignItems(ui.Start).Gap(12).Padding(22, 22, 0, 22)
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
		ui.Scroll(c).Grow(1).MinHeight(0).Children(func() {
			ui.Column(c).FillWidth().Padding(bp...).TextColor(cl.Text2).FontSize(FsBase).Children(body)
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
// the keyboard focus: a small dark label (light in the dark theme) that
// rises 3px as it fades in. Give anchor a Description too: assistive
// technology does not read tooltips.
func Tooltip(c *ui.Context, anchor ui.Element, text string) ui.Element {
	t := From(c)
	cl := t.Colors
	return ui.TooltipBase(c, anchor, func(tip ui.Element) {
		tip.Margin(0, 0, 8, 0).MaxWidth(260).Padding(6, 10).Radius(RadiusItem).Background(cl.Text.Alpha(0.92)).
			Transition(ui.ElementTransition{Duration: 150 * time.Millisecond, Enter: &ui.Motion{Y: 3}, Exit: &ui.Motion{Y: 3}})
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

// Toaster draws the window's toasts in its bottom right corner: cards with
// the pop shadow, a tint of their tone in the top left corner, a dot of
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
			r := p.Root.Row().Width(380).MaxWidthPercent(100).AlignItems(ui.Start).Gap(10).Padding(12, 10, 12, 14).
				Radius(RadiusPop).Background(over(Mix(ui.Color{}, tone, 0.12), cl.Card))
			r.RadialGradient(ui.RadialGradient{Center: ui.Point{X: 0.001, Y: 0.001}, Radius: 220, Oklab: true, Stops: []ui.GradientStop{
				{Color: over(Mix(ui.Color{}, tone, 0.12), cl.Card), At: 0}, {Color: cl.Card, At: 0.6},
			}})
			t.Shadows.ShadowPop.Apply(r)
			r.FontSize(FsMd).TextColor(cl.Text).Transition(ui.ElementTransition{
				Duration: 250 * time.Millisecond, Ease: ui.EaseOut, Position: true,
				Enter: &ui.Motion{Y: 40}, Exit: &ui.Motion{Y: 8, Scale: 0.96},
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
