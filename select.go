package ryugo

import (
	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"
	"time"
)

// Option is an option of a selection, in the same shape everywhere: a value and a label.
type Option[T comparable] struct {
	Value T
	Label string
	// Description is the muted description next to the label.
	Description string
	// Dot is the series-color dot in front of the option (Colors.C1…C8).
	Dot      ui.Color
	Disabled bool
}

// SelectOptions are the options of Select.
type SelectOptions struct {
	// Placeholder is the muted text shown while the value is not among the options.
	Placeholder string
	Label       string
	Size        Size
	Disabled    bool
	Invalid     bool
}

// Select is a selection list without search: the trigger is a raised field
// (the button's skin, a field's dimensions), the list a glass popup panel;
// the highlighted option fills with the accent gradient, and a check mark
// stands next to the selected one.
//
//	ryugo.Select(c.Key("range"), &app.rangeDays, []ryugo.Option[int]{
//		{Value: 7, Label: "Last 7 days"},
//		{Value: 30, Label: "Last 30 days"},
//	}, ryugo.SelectOptions{})
func Select[T comparable](c *ui.Context, value *T, options []Option[T], o SelectOptions) ui.Element {
	t := sizedFrom(c, o.Size)
	m := t.Metrics
	cl := t.Colors

	sel := ui.SelectBase(c, value)
	tr := sel.Trigger
	tr.FillWidth().Height(m.FieldH).Padding(0, m.FieldPe, 0, m.FieldPs).Radius(m.FieldR).
		Justify(ui.SpaceBetween).AlignItems(ui.Center).Gap(6).Cursor(ui.CursorPointer).Disabled(o.Disabled)
	if o.Label != "" {
		tr.Label(o.Label)
	}
	open := sel.Open()
	s := raisedState{hovered: tr.Hovered() && !open, pressed: false, focus: tr.FocusVisible(), disabled: o.Disabled}
	t.raised(tr, t.controlGradient(cl.ControlTop, cl.ControlBottom), cl.ControlBorder, t.Shadows.ControlShadow, s)
	switch {
	case o.Invalid:
		tr.BorderColor(cl.Up)
	case open:
		// open field: accent border and a soft ring (like Input's focus)
		tr.BorderColor(cl.ColorAccent500)
		t.Shadows.Ring.Apply(tr)
	}
	if tr.Pressed() && !o.Disabled {
		// a pressed field does not move down; it shrinks slightly
		tr.Scale(0.99, 0.99)
	}

	current := -1
	for i, op := range options {
		if op.Value == *value {
			current = i
		}
	}
	tr.FontSize(m.FieldFs).FontWeight(500).TextColor(cl.Text).Children(func() {
		ui.Row(c).AlignItems(ui.Center).Gap(6).MinWidth(0).Shrink(1).Children(func() {
			if current < 0 {
				ui.Text(c, o.Placeholder).TextColor(cl.Muted).FontWeight(400).SingleLine().Ellipsis("…")
				return
			}
			op := options[current]
			if op.Dot != (ui.Color{}) {
				dot(c, op.Dot)
			}
			ui.Text(c, op.Label).SingleLine().Ellipsis("…").Shrink(1)
			if op.Description != "" {
				ui.Text(c, op.Description).TextColor(cl.Muted).FontSize(FsSm).FontWeight(400).SingleLine()
			}
		})
		ui.Icon(c, iconChevrons).TextColor(cl.Muted)
	})

	sel.Popup(func(panel ui.Element) {
		popover(t, panel)
		panel.Gap(0)
		// past its height the list scrolls inside its panel
		t.scrollList(c, 300-2*PadPop, func() {
			for _, op := range options {
				optionItem(c, t, sel.Item(op.Value).Disabled(op.Disabled), op, op.Value == *value)
			}
		})
	})
	return tr
}

// dot is the 8px series-color dot (the same as in Chips).
func dot(c *ui.Context, color ui.Color) ui.Element {
	return ui.Box(c).Size(8, 8).Radius(RadiusPill).Background(color).Shrink(0)
}

// popover paints a pop-up surface: frosted glass (what is behind blurs by
// 24 DIPs under the translucent pop color), radius-pop corners and two
// layers of spreading shadow; gap-field away from its trigger, it springs
// open from its top corner, from 95% and 4px up.
func popover(t *Tokens, panel ui.Element) {
	panel.Margin(GapField, 0, 0, 0).Padding(PadPop).TextColor(t.Colors.Text)
	t.frost(panel, t.Colors.Pop, t.R(RadiusPop), 24)
	t.Shadows.ShadowPop.Apply(panel)
	panel.TransformOrigin(0, 0).Transition(ui.ElementTransition{
		Animation: springPop,
		Enter:     &ui.Motion{Scale: 0.95, Y: -4},
		Exit:      &ui.Motion{Scale: 0.97, Y: -2},
	})
}

// scrollList builds a pop-up's list, which scrolls inside its panel past
// maxH: where there is more of it above or below, its edge fades into the
// panel, so that more is seen to be there.
func (t *Tokens) scrollList(c *ui.Context, maxH float32, build func()) {
	box := ui.Box(c.Key("list")).MinHeight(0).Shrink(1)
	sc := ui.Local(box, "scroll", func() *ui.ScrollState { return &ui.ScrollState{} })
	box.Children(func() {
		ui.Scroll(c).MaxHeight(maxH).TrackScroll(*sc).Children(build)
		above, below := (*sc).Y, (*sc).MaxY-(*sc).Y
		listFade(c, t, false, above)
		listFade(c, t, true, below)
	})
}

// listFade is the fade of a pop-up list's top (or bottom) edge, there as
// far as the list goes on past it, up to 16 DIPs.
func listFade(c *ui.Context, t *Tokens, bottom bool, past float32) {
	e := ui.Box(c).Absolute().Left(0).Right(0).Height(18).PassThrough()
	angle := float32(180)
	if bottom {
		e.Bottom(0)
		angle = 0
	} else {
		e.Top(0)
	}
	shown := e.AnimateTo("shown", clamp01(past/16), ui.Animation{Duration: 120 * time.Millisecond})
	if shown < 0.01 {
		return
	}
	tint := t.Colors.Pop
	e.Draw(func(p *ui.Painter, r ui.Rect) {
		p.FillGradient(r, ui.LinearGradient{Angle: angle, Oklab: true, Stops: []ui.GradientStop{
			{Color: tint.Alpha(0.95 * shown), At: 0}, {Color: tint.Alpha(0), At: 1},
		}}, 0)
	})
}

// frost fills e with frosted glass: what is behind it blurs by blur DIPs,
// under tint, in corners of radius.
func (t *Tokens) frost(e ui.Element, tint ui.Color, radius, blur float32) {
	e.Radius(radius).Material(glass.Blur{Radius: blur}).Draw(func(p *ui.Painter, r ui.Rect) { p.Fill(r, tint, radius) })
}
