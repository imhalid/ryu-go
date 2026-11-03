package mizan

import "github.com/egoist/mygo/ui"

// InputOptions, Input'un seçenekleridir.
type InputOptions struct {
	Placeholder string
	// Prefix ve Suffix, alanın içindeki soluk ekler: ikon ya da metin önde,
	// birim arkada ("₺", "kg").
	PrefixIcon *ui.SVG
	Prefix     string
	Suffix     string
	// Clearable, alan doluyken temizle düğmesi gösterir.
	Clearable bool
	// Password, yazıyı gizler; göster/gizle düğmesi ekler.
	Password bool
	// Label, alanın erişilebilirlik adıdır; Field içinde Field'ın etiketi.
	Label    string
	Size     Size
	Disabled bool
	// Invalid, alanı hatalı gösterir: kenar ve halka Up renginde.
	Invalid bool
}

// Input, gömük metin alanıdır: opak kuyu, ince kenar; odakta yüzeye açılır
// ve vurgu halkası alır. Yazının kendisini (Changed, Submitted, Focus için)
// döndürür.
//
//	mizan.Input(c.Key("q"), &app.query, mizan.InputOptions{Placeholder: "Ara", Clearable: true})
func Input(c *ui.Context, value *string, o InputOptions) ui.Element {
	t := sizedFrom(c, o.Size)
	m := t.Metrics
	cl := t.Colors
	fs := m.InputFs
	if fs == 0 {
		// md'de yazı çevreden kalıtılır: gövde metni
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
			if inputButton(c.Key("clear"), t, iconClear, "Temizle").Clicked() {
				*value = ""
				field.Focus()
			}
		}
		if o.Password {
			icon, label := iconEye, "Göster"
			if *reveal {
				icon, label = iconEyeOff, "Gizle"
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

// adorn, alanın içindeki soluk, tabular ektir.
func adorn(c *ui.Context, s string, cl Colors) {
	ui.Text(c, s).TextColor(cl.Muted).NoWrap().Unselectable().FontFeatures("tnum")
}

// inputButton, gömük alanın içindeki küçük, düz, yuvarlak düğmedir
// (temizle, göster/gizle).
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
