package ryugo

import (
	"github.com/egoist/mygo/ui"
)

// SplitButtonOptions are the options of SplitButton: its face is a Button's, of the same color, variant and size.
type SplitButtonOptions struct {
	// Color, if set, tints both segments with that color; Variant, with Color: Soft (the default) or Solid.
	Color   ui.Color
	Variant Variant
	// Icon is the icon in front of the label.
	Icon *ui.SVG
	// Size, if not set, comes from the enclosing scope.
	Size Size
	// Disabled disables the whole control; ActionDisabled the action alone, the menu still opening.
	Disabled, ActionDisabled bool
	// MenuLabel names the arrow for assistive technology: MyGo's "More options" when empty.
	MenuLabel string
}

// SplitButton is an action joined to an arrow that opens a menu of its variants, as New page beside the pages of
// the other templates: MyGo's SplitButton, one stop of Tab whose segments Left and Right move between, Enter and
// Space clicking the action and opening the menu, Up and Down and Alt+Down opening it too, a group of a button and a
// menu button to assistive technology; with a Button's face, its press wave and its sound. Each segment lights up
// apart; the outer corners are the pill's, the inner ones square, and a line parts them. Button.Clicked reports the
// action; Arrow is given the menu, with Menu or PopoverMenuButton, or opens one of the app's on Arrow.Clicked.
//
//	sb := ryugo.SplitButton(c.Key("new"), "New page", ryugo.SplitButtonOptions{Icon: icons.Plus, Color: t.Colors.Accent})
//	sb.Arrow.PopoverMenuButton(func(m *ui.Menu) { … })
//	if sb.Button.Clicked() { … }
func SplitButton(c *ui.Context, label string, o SplitButtonOptions) ui.SplitButtonParts {
	t := sizedFrom(c, o.Size)
	m := t.Metrics
	p := ui.SplitButton(c, "") // its label is the Button's content, below
	g, act, arrow := p.Group, p.Button, p.Arrow
	r := t.Pill(m.BtnH)
	// the face is the segments', each its own: the group, MyGo's, draws none
	g.Background(ui.Color{}).Border(0, ui.Color{}).Radius(r).Height(m.BtnH).Disabled(o.Disabled)
	act.Label(label).Disabled(o.Disabled || o.ActionDisabled)
	if o.MenuLabel != "" {
		arrow.Label(o.MenuLabel)
	}
	segment := func(e ui.Element, off bool, radius ...float32) (text ui.Color) {
		s := raisedState{hovered: e.Hovered() && !off, pressed: e.Pressed() && !off, focus: e.FocusVisible(), disabled: off}
		text, solid := t.buttonFace(e, o.Color, o.Variant, false, s)
		e.Radius(radius...).Height(m.BtnH).Cursor(ui.CursorPointer)
		if x, w, ok := ripple(c, e, s.pressed, e.Clicked() && !off); ok && t.sounds {
			k := tonePlain
			if solid {
				k = toneSolid
			}
			pressSound(k, x, w)
		}
		return text
	}
	text := segment(act, o.Disabled || o.ActionDisabled, r, 0, 0, r)
	act.Center().Gap(m.BtnGap).Padding(0, m.BtnPx).TextColor(text).FontSize(m.BtnFs).FontWeight(600).LetterSpacing(-0.01 * m.BtnFs)
	act.Children(func() {
		if o.Icon != nil {
			ui.Icon(c, o.Icon)
		}
		ui.Text(c, label).NoWrap()
	})
	segment(arrow, o.Disabled, 0, r, r, 0)
	arrow.Center().Width(m.BtnH * 0.9).TextColor(text)
	return p
}
