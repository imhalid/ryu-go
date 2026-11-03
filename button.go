package mizan

import "github.com/egoist/mygo/ui"

// Variant is the fill of a colored button.
type Variant uint8

const (
	// Soft is a raised surface tinted with the color; the label is between the color and the text.
	Soft Variant = iota + 1
	// Solid is a colored gradient with a white label.
	Solid
)

// ButtonOptions are the options of Button; the zero value is a raised neutral button.
type ButtonOptions struct {
	// Color, if set, tints the button with that color (Colors.C1…C8, Accent, Up, Down).
	Color ui.Color
	// Variant, with Color: Soft (the default) or Solid.
	Variant Variant
	// Danger marks a destructive action: in the Up color, Solid by default. Ignored if Color is set.
	Danger bool
	// Icon is the icon in front of the label; with IconOnly, a square icon
	// button (control-h on a side), and the label is only the accessibility name.
	Icon     *ui.SVG
	IconOnly bool
	// Size, if not set, comes from the enclosing scope.
	Size Size
	// Loading shows a spinner and disables clicks; the button does not fade.
	Loading  bool
	Disabled bool
	// Block stretches the button across the row.
	Block bool
}

// Button is a raised pill button: a single kind, as tall as the fields of
// the same size step (control-h). Clicked reports that it was pressed.
//
//	if mizan.Button(c.Key("export"), "Export", mizan.ButtonOptions{Color: t.Colors.Accent, Variant: mizan.Solid}).Clicked() {
//		...
//	}
func Button(c *ui.Context, label string, o ButtonOptions) ui.Element {
	t := sizedFrom(c, o.Size)
	m := t.Metrics
	cl := t.Colors

	e := ui.ButtonBase(c).Label(label).Center().Height(m.BtnH).Gap(m.BtnGap).Radius(RadiusPill).Cursor(ui.CursorPointer)
	if o.IconOnly {
		e.Width(m.BtnH)
	} else {
		e.Padding(0, m.BtnPx)
	}
	if o.Block {
		e.FillWidth()
	}
	off := o.Disabled || o.Loading
	e.Disabled(o.Disabled)
	s := raisedState{hovered: e.Hovered() && !off, pressed: e.Pressed() && !off, focus: e.FocusVisible(), disabled: o.Disabled}

	tint := o.Color
	solid := o.Variant == Solid
	if tint == (ui.Color{}) && o.Danger {
		tint = cl.Up
		solid = o.Variant != Soft
	}
	text := cl.Text
	switch {
	case tint == (ui.Color{}):
		t.raised(e, t.controlGradient(cl.ControlTop, cl.ControlBottom), cl.ControlBorder, t.Shadows.ControlShadow, s)
	case solid:
		// the fill's lightness is capped in OKLCH (top 0.52, bottom 0.46): hue
		// and chroma stay, and white text reads at ≥ 4.5:1
		top := capLightness(tint, 0.52)
		bottom := capLightness(lighten(tint, -0.06), 0.46)
		g := ui.LinearGradient{Angle: 180, Oklab: true, From: top, To: bottom}
		sh := t.Shadows.Sheen.With(ShadowLayer{Y: 1, Blur: 2, Color: Mix(tint, ui.Color{}, 0.55)})
		t.raised(e, g, Mix(bottom, oklch(0, 0, 0, 1), 0.2), sh, s)
		text = cl.OnAccent
	default:
		g := t.controlGradient(Mix(cl.ControlTop, tint, 0.1), Mix(cl.ControlBottom, tint, 0.2))
		t.raised(e, g, Mix(cl.ControlBorder, tint, 0.32), t.Shadows.ControlShadow, s)
		// half the color, half the text color: ≥ 4.5:1 in both themes
		text = Mix(cl.Text, tint, 0.5)
	}
	if o.Loading {
		e.Cursor(ui.CursorDefault)
	}

	e.TextColor(text).FontSize(m.BtnFs).FontWeight(600).LetterSpacing(-0.01 * m.BtnFs)
	e.Children(func() {
		switch {
		case o.Loading:
			ui.Spinner(c).Size(m.BtnFs+2, m.BtnFs+2)
		case o.Icon != nil:
			ui.Icon(c, o.Icon)
		}
		if !o.IconOnly {
			ui.Text(c, label).NoWrap()
		}
	})
	return e
}

// sizedFrom gives the scope's tokens at the component's own size step (the
// component's Size overrides the inherited one). It builds no wrapper
// element: c's key (c.Key) stays with the component's own element.
func sizedFrom(c *ui.Context, s Size) *Tokens {
	if s == SizeInherit {
		return From(c)
	}
	return scoped(c, Options{Size: s})
}
