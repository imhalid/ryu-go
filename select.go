package mizan

import (
	"time"

	"github.com/egoist/mygo/ui"
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
//	mizan.Select(c.Key("range"), &app.rangeDays, []mizan.Option[int]{
//		{Value: 7, Label: "Last 7 days"},
//		{Value: 30, Label: "Last 30 days"},
//	}, mizan.SelectOptions{})
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
		panel.MaxHeight(300).Gap(0)
		for _, op := range options {
			item := sel.Item(op.Value).AlignItems(ui.Center).Gap(6).Padding(6, 14, 6, 8).
				Radius(RadiusItem).FontSize(FsMd).Disabled(op.Disabled).Cursor(ui.CursorPointer)
			hl := item.Highlighted() && !op.Disabled
			muted, check := cl.Muted, cl.Accent
			if hl {
				item.LinearGradient(ui.LinearGradient{Angle: 180, Oklab: true, From: cl.ColorAccent500, To: cl.ColorAccent600}).
					TextColor(cl.OnAccent)
				muted, check = cl.OnAccentMuted, cl.OnAccentMuted
			} else {
				item.TextColor(cl.Text)
			}
			if op.Disabled {
				item.Opacity(0.35).Cursor(ui.CursorDefault)
			}
			item.Children(func() {
				icon := ui.Icon(c, iconCheck).TextColor(check)
				if op.Value != *value {
					icon.Invisible()
				}
				if op.Dot != (ui.Color{}) {
					d := dot(c, op.Dot)
					if hl {
						d.Shadow(0, 0, 0, 1.5, cl.OnAccentMuted)
					}
				}
				ui.Text(c, op.Label).NoWrap()
				if op.Description != "" {
					ui.Text(c, op.Description).TextColor(muted).FontSize(FsSm).NoWrap()
				}
			})
		}
	})
	return tr
}

// dot is the 8px series-color dot (the same as in Chips).
func dot(c *ui.Context, color ui.Color) ui.Element {
	return ui.Box(c).Size(8, 8).Radius(RadiusPill).Background(color).Shrink(0)
}

// popover paints the popup surface: glass background, radius-pop corners, a
// 0.5px line and a two-layer spreading shadow; it sits gap-field away from
// the trigger and opens from 96% size, starting transparent.
func popover(t *Tokens, panel ui.Element) {
	// In Mizan, glass is a translucent --pop plus a backdrop-filter blur. MyGo
	// has no layer that blurs what is behind; left translucent, the text
	// underneath would show through, so the color is made opaque over the card
	// background.
	panel.Margin(GapField, 0, 0, 0).Padding(PadPop).Radius(RadiusPop).Background(over(t.Colors.Pop, t.Colors.Card)).
		TextColor(t.Colors.Text)
	t.Shadows.ShadowPop.Apply(panel)
	panel.Transition(ui.ElementTransition{
		Duration: 180 * time.Millisecond,
		Enter:    &ui.Motion{Opacity: 0, Scale: 0.96},
	})
}
