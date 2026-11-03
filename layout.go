package mizan

import "github.com/egoist/mygo/ui"

// CardOptions, Card'ın seçenekleridir.
type CardOptions struct {
	Title    string
	Subtitle string
	// Pad, panele pad-card (28px) iç dolgu verir; yoksa yalnız 10px dikey
	// (içine tablo ya da liste kenardan kenara oturur).
	Pad bool
	// Actions, başlığın sağındaki düğmeleri kurar.
	Actions func()
}

// Card, başlığı panelin dışında, üstünde duran yüzeydir. Panel kart renginde,
// radius-lg köşeli; kenarı üstte belirgin, aşağı doğru sönen bir gradyandır
// (koyu temada gölge görünmez, kenar ayırır).
//
//	mizan.Card(c, mizan.CardOptions{Title: "Aktif kullanıcı", Pad: true}, func() {
//		...
//	})
func Card(c *ui.Context, o CardOptions, body func()) ui.Element {
	t := From(c)
	cl := t.Colors
	card := ui.Column(c).MinWidth(0)
	card.Children(func() {
		if o.Title != "" || o.Actions != nil {
			ui.Row(c).AlignItems(ui.Start).Justify(ui.SpaceBetween).Gap(16).Padding(12, 4, 14, 4).Children(func() {
				ui.Column(c).Grow(1).MinWidth(0).Gap(3).Children(func() {
					if o.Title != "" {
						Text(c, o.Title, TTitle).Level(2).Role(ui.RoleHeading)
					}
					if o.Subtitle != "" {
						ui.Text(c, o.Subtitle).TextColor(cl.Muted).FontSize(FsMd)
					}
				})
				if o.Actions != nil {
					ui.Row(c).AlignItems(ui.Center).Gap(GapControl).Children(o.Actions)
				}
			})
		}
		panel := ui.Column(c).MinWidth(0).Radius(RadiusLg).Background(cl.Card)
		if o.Pad {
			panel.Padding(PadCard)
		} else {
			panel.Padding(10, 0)
		}
		panel.Border(1, ui.Color{}).BorderLinearGradient(t.edge()).Children(body)
	})
	return card
}

// edge, kart kenarının gradyanıdır: %10'da en belirgin, %90'da en soluk.
func (t *Tokens) edge() ui.LinearGradient {
	c := t.Colors
	return ui.LinearGradient{Angle: 180, Stops: []ui.GradientStop{
		{Color: c.GradientBorder10, At: 0.1},
		{Color: c.GradientBorder25, At: 0.25},
		{Color: c.GradientBorder75, At: 0.75},
		{Color: c.GradientBorder90, At: 0.9},
	}}
}

// FieldOptions, Field'ın seçenekleridir.
type FieldOptions struct {
	Label string
	// Help, alanın altındaki soluk açıklamadır; Error verilince onun yerine
	// Up renginde hata yazılır.
	Help  string
	Error string
	// Required, etiketin yanına soluk bir yıldız koyar.
	Required bool
}

// Field, etiket, kontrol ve açıklamadan oluşan dikey gruptur: etiket fs-sm
// 500 soluk, kontrolle arası gap-field. Hatalı alanda etiket Up rengine
// döner; kontrolün kendisi Invalid ile işaretlenir.
//
//	mizan.Field(c, mizan.FieldOptions{Label: "E-posta", Error: app.emailErr}, func() {
//		mizan.Input(c.Key("email"), &app.email, mizan.InputOptions{Label: "E-posta", Invalid: app.emailErr != ""})
//	})
func Field(c *ui.Context, o FieldOptions, control func()) ui.Element {
	cl := From(c).Colors
	f := ui.Column(c).Gap(GapField).MinWidth(0)
	label := cl.Muted
	if o.Error != "" {
		label = cl.Up
	} else if f.FocusWithin() {
		label = cl.Text
	}
	f.Children(func() {
		ui.Row(c).Gap(3).Children(func() {
			ui.Text(c, o.Label).FontSize(FsSm).FontWeight(500).TextColor(label)
			if o.Required {
				ui.Text(c, "*").FontSize(FsSm).TextColor(cl.Muted)
			}
		})
		control()
		switch {
		case o.Error != "":
			ui.Text(c, o.Error).FontSize(FsXs).TextColor(cl.Up)
		case o.Help != "":
			ui.Text(c, o.Help).FontSize(FsXs).TextColor(cl.Muted)
		}
	})
	return f
}
