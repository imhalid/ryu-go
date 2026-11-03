package mizan

import "github.com/egoist/mygo/ui"

// InputOptions are the options of Input.
type InputOptions struct {
	Placeholder string
	// Prefix and Suffix are muted adornments inside the field: an icon or
	// text in front, a unit behind ("$", "kg").
	PrefixIcon *ui.SVG
	Prefix     string
	Suffix     string
	// Clearable shows a clear button while the field is not empty.
	Clearable bool
	// Password hides the text and adds a show/hide button.
	Password bool
	// Label is the field's accessibility name; inside a Field, the Field's label.
	Label    string
	Size     Size
	Disabled bool
	// Invalid shows the field as invalid: border and ring in the Up color.
	Invalid bool
}

// Input is an inset text field: an opaque well with a thin border; on focus
// it lightens to the surface and gets an accent ring. It returns the text
// input itself (for Changed, Submitted, Focus).
//
//	mizan.Input(c.Key("q"), &app.query, mizan.InputOptions{Placeholder: "Search", Clearable: true})
func Input(c *ui.Context, value *string, o InputOptions) ui.Element {
	t := sizedFrom(c, o.Size)
	m := t.Metrics
	cl := t.Colors
	fs := m.InputFs
	if fs == 0 {
		// at md the font size is inherited: body text
		fs = FsBase
	}

	box := ui.Row(c).AlignItems(ui.Center).Height(m.FieldH).Gap(m.InputGap).
		Radius(m.FieldR).Padding(0, m.FieldPs).MinWidth(0).Cursor(ui.CursorText)
	reveal := ui.Local(box, "reveal", func() bool { return false })
	var field ui.Element
	box.Children(func() {
		if o.PrefixIcon != nil {
			ui.Icon(c, o.PrefixIcon).TextColor(cl.Muted)
		} else if o.Prefix != "" {
			adorn(c, o.Prefix, cl)
		}
		field = ui.TextInputBase(c, value).Grow(1).MinWidth(0).
			Placeholder(o.Placeholder).PlaceholderColor(cl.Muted).
			TextColor(cl.Text).FontSize(fs).Disabled(o.Disabled).FocusRing(false)
		if o.Label != "" {
			field.Label(o.Label)
		}
		if o.Password && !*reveal {
			field.Password()
		}
		if o.Suffix != "" {
			adorn(c, o.Suffix, cl)
		}
		if o.Clearable && *value != "" && !o.Disabled {
			if inputButton(c.Key("clear"), t, iconClear, "Clear").Clicked() {
				*value = ""
				field.Focus()
			}
		}
		if o.Password {
			icon, label := iconEye, "Show"
			if *reveal {
				icon, label = iconEyeOff, "Hide"
			}
			if inputButton(c.Key("reveal"), t, icon, label).Clicked() {
				*reveal = !*reveal
			}
		}
	})
	t.well(box, field.Focused(), o.Invalid)
	if o.Disabled {
		box.Opacity(0.4).Cursor(ui.CursorDefault)
	} else if box.Clicked() {
		field.Focus()
	}
	return field
}

// adorn is a muted, tabular adornment inside the field.
func adorn(c *ui.Context, s string, cl Colors) {
	ui.Text(c, s).TextColor(cl.Muted).NoWrap().Unselectable().FontFeatures("tnum")
}

// inputButton is the small, flat, round button inside the inset field
// (clear, show/hide).
func inputButton(c *ui.Context, t *Tokens, icon *ui.SVG, label string) ui.Element {
	m := t.Metrics
	b := ui.ButtonBase(c).Label(label).Center().Size(m.InputBtn, m.InputBtn).
		Radius(RadiusPill).Margin(0, -4, 0, 0).Cursor(ui.CursorPointer)
	color := t.Colors.Muted
	if b.Hovered() {
		b.Background(t.Colors.Fill)
		color = t.Colors.Text
	}
	b.TextColor(color).Children(func() { ui.Icon(c, icon) })
	return b
}
