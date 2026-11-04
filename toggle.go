package ryugo

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// ToggleOptions are the options of Toggle.
type ToggleOptions struct {
	// Icon is the icon in front of the label; with IconOnly, a square
	// toggle, and the label is only the accessibility name.
	Icon     *ui.SVG
	IconOnly bool
	Size     Size
	Disabled bool
}

// Toggle is a button that stays pressed while *on, as Bold in an editor's
// toolbar: off it is flat, in the muted color, and takes a soft fill under
// the pointer; on it is pressed into the surface, an inset well with the
// label in the text color and the icon in the accent. A click or Space
// turns it over, which Changed reports; assistive technology sees a toggle
// button. It is as tall as a button of its size step.
//
//	ryugo.Toggle(c.Key("bold"), &app.bold, "Bold", ryugo.ToggleOptions{Icon: boldIcon, IconOnly: true})
func Toggle(c *ui.Context, on *bool, label string, o ToggleOptions) ui.Element {
	t := sizedFrom(c, o.Size)
	m := t.Metrics
	cl := t.Colors
	radius := t.R(RadiusItem)
	if groupInner >= 0 {
		radius = groupInner
	}
	e := ui.ToggleBase(c, on).Label(label).Center().Height(m.BtnH).Gap(m.BtnGap).Radius(radius).
		Cursor(ui.CursorPointer).FocusRing(false).Disabled(o.Disabled).Shrink(0)
	if o.IconOnly {
		e.Width(m.BtnH)
	} else {
		e.Padding(0, m.BtnPx-4)
	}
	text, icon := cl.Muted, cl.Muted
	var sh Shadow
	switch {
	case *on:
		e.Background(cl.Well)
		text, icon = cl.Text, cl.ColorAccent500
	case e.Hovered() && !o.Disabled:
		e.Background(cl.Fill)
		text, icon = cl.Text, cl.Text
	}
	if e.FocusVisible() {
		sh = t.Shadows.FocusRing
	}
	sh.Apply(e)
	if e.Pressed() && !o.Disabled {
		e.Scale(0.96, 0.96)
	}
	if o.Disabled {
		e.Opacity(0.4).Cursor(ui.CursorDefault)
	}
	e.TextColor(text).FontSize(m.BtnFs).FontWeight(600).
		Transition(ui.ElementTransition{Duration: 150 * time.Millisecond, Colors: true, Ease: ui.EaseOut})
	e.Children(func() {
		if o.Icon != nil {
			ui.Icon(c, o.Icon).TextColor(icon)
		}
		if !o.IconOnly {
			ui.Text(c, label).NoWrap()
		}
	})
	return e
}

// ToggleGroup is a toolbar of toggles and buttons on one raised bar, as a
// text editor's B / I / U: the arrows move the focus between them, Tab
// leaves the bar. Separator{Vertical: true} divides groups within it. label
// names it to assistive technology.
//
//	ryugo.ToggleGroup(c, "Text style", func() {
//		ryugo.Toggle(c.Key("b"), &app.bold, "Bold", ryugo.ToggleOptions{Icon: boldIcon, IconOnly: true})
//		ryugo.Toggle(c.Key("i"), &app.italic, "Italic", ryugo.ToggleOptions{Icon: italicIcon, IconOnly: true})
//	})
func ToggleGroup(c *ui.Context, label string, build func()) ui.Element {
	t := From(c)
	cl := t.Colors
	bar := ui.Row(c).Role(ui.RoleToolbar).Label(label).FocusGroup(ui.Horizontal).AlignItems(ui.Center).
		Gap(2).Padding(3).Radius(t.R(RadiusControl)).Shrink(0).
		LinearGradient(t.controlGradient(cl.ControlTop, cl.ControlBottom)).Border(0.5, cl.ControlBorder)
	t.Shadows.ControlShadow.Apply(bar)
	// the toggles inside, 3px in, take the bar's corners' centers
	prev := groupInner
	groupInner = Concentric(t.R(RadiusControl), 3)
	bar.Children(build)
	groupInner = prev
	return bar
}

// groupInner is the radius of the toggles of the ToggleGroup being built,
// -1 outside one (0 is a square corner).
var groupInner float32 = -1

// SeparatorOptions are the options of Separator.
type SeparatorOptions struct {
	// Vertical divides a row, as the groups of a toolbar; it is as tall as
	// the row.
	Vertical bool
	// Label is written in the middle of a horizontal separator, as "or".
	Label string
}

// Separator is a groove dividing content: a dark line and a light one
// beside it, as Menu's and Sheet's footer's, cut into the surface rather
// than drawn on it.
//
//	ryugo.Separator(c, ryugo.SeparatorOptions{Label: "or"})
func Separator(c *ui.Context, o SeparatorOptions) ui.Element {
	cl := From(c).Colors
	groove := func() ui.Element {
		if o.Vertical {
			return ui.Box(c).Width(2).AlignSelf(ui.Stretch).Margin(4, 4).Shrink(0).
				InsetShadow(1, 0, 0, 0, cl.SepDark).InsetShadow(-1, 0, 0, 0, cl.SepLight)
		}
		return ui.Box(c).Height(2).Grow(1).MinWidth(0).
			InsetShadow(0, 1, 0, 0, cl.SepDark).InsetShadow(0, -1, 0, 0, cl.SepLight)
	}
	if o.Label == "" {
		g := groove()
		if !o.Vertical {
			g.FillWidth()
		}
		return g
	}
	return ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(12).Children(func() {
		groove()
		ui.Text(c, o.Label).TextColor(cl.Muted).FontSize(FsXs).FontWeight(500).NoWrap()
		groove()
	})
}

// Link is a link inside text: in the accent color, deep enough to read
// as text, with an underline of a third of it that turns full under the
// pointer. A click or Enter opens url in the browser (in a Router, a path
// goes there). Inside a RichText, it is a link within the paragraph.
//
//	ryugo.Link(c, "Read the changelog", "https://example.com/changelog")
func Link(c *ui.Context, label, url string) ui.Element {
	t := From(c)
	color := t.Colors.ColorAccent600
	if t.Dark {
		color = t.Colors.ColorAccent400
	}
	e := ui.Link(c, label, url).TextColor(color).Underline().Cursor(ui.CursorPointer)
	line := color.Alpha(0.35)
	if e.Hovered() {
		line = color
	}
	return e.DecorationColor(line)
}
