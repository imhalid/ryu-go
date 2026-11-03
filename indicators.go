package mizan

import (
	"time"
	"unicode"

	"github.com/egoist/mygo/ui"
)

// Tone, durum renkli küçük öğelerin (Badge, StatusDot) tonudur. Renk tek
// başına anlam taşımaz: yanında her zaman metin ya da ad durur.
type Tone uint8

const (
	// ToneDefault, metin rengidir.
	ToneDefault Tone = iota
	ToneUp
	ToneDown
	ToneWarn
	ToneMuted
	// ToneNeutral, gri etikettir: zemin soluk metinden değil dolgudan.
	ToneNeutral
)

func (t *Tokens) tone(n Tone) ui.Color {
	c := t.Colors
	switch n {
	case ToneUp:
		return c.Up
	case ToneDown:
		return c.Down
	case ToneWarn:
		return c.Warn
	case ToneMuted, ToneNeutral:
		return c.Muted
	}
	return c.Text
}

// Badge, metin renginin %10'u zeminli haptır: hangi tonda olursa olsun
// zemin metinle uyumlu. Sayılar tabular.
//
//	mizan.Badge(c, "+4,2%", mizan.ToneDown)
func Badge(c *ui.Context, text string, tone Tone) ui.Element {
	t := From(c)
	color := t.tone(tone)
	bg := Mix(ui.Color{}, color, 0.1)
	if tone == ToneNeutral {
		bg = t.Colors.Fill
	}
	return ui.Box(c).Padding(2, 8).Radius(RadiusPill).Background(bg).Shrink(0).Children(func() {
		ui.Text(c, text).TextColor(color).FontSize(FsXs).FontWeight(600).NoWrap().FontFeatures("tnum")
	})
}

// Tag, sayfa başlığının yanındaki büyük nötr haptır.
func Tag(c *ui.Context, text string) ui.Element {
	t := From(c)
	return ui.Box(c).Padding(6, 12).Radius(RadiusPill).Background(t.Colors.Fill).Shrink(0).Children(func() {
		ui.Text(c, text).TextColor(t.Colors.Muted).FontSize(FsSm).FontWeight(500).NoWrap()
	})
}

// DotState, StatusDot'un durumudur.
type DotState uint8

const (
	DotOff DotState = iota
	// DotLive, yeşil nokta ve dışa yayılıp sönen halka.
	DotLive
	DotDelayed
)

// StatusDot, durum noktası ve metnidir: canlıda nokta 1,8 saniyede bir
// dışa yayılıp söner (hareket azaltılınca durur).
//
//	mizan.StatusDot(c, mizan.DotLive, "Canlı")
func StatusDot(c *ui.Context, state DotState, text string) ui.Element {
	t := From(c)
	color := t.Colors.Muted
	switch state {
	case DotLive:
		color = t.Colors.Down
	case DotDelayed:
		color = t.Colors.C3
	}
	return ui.Row(c).AlignItems(ui.Center).Gap(7).Children(func() {
		d := ui.Box(c).Size(8, 8).Radius(RadiusPill).Background(color).Shrink(0)
		if state == DotLive {
			// halka: noktanın kendi gölgesi, yarıçapı 2,6 katına açılırken söner
			p := d.Keyframes("pulse", []ui.Keyframe{{At: 0, Value: 0}, {At: 1800 * time.Millisecond, Value: 1, Ease: ui.EaseOut}},
				ui.KeyframeOptions{Repeat: ui.RepeatForever})
			d.Shadow(0, 0, 0, 6.4*p, color.Alpha(0.6*(1-p)))
		}
		ui.Text(c, text).TextColor(t.Colors.Text2).FontSize(FsSm).NoWrap()
	})
}

// Kbd, klavye tuşlarıdır: düğmeyle aynı kabarık malzeme, yazıya göre
// ölçeklenir. Birleşik kısayolda tuşlar yan yana.
//
//	mizan.Kbd(c, "⌘", "K")
func Kbd(c *ui.Context, keys ...string) ui.Element {
	t := From(c)
	cl := t.Colors
	fs := float32(FsBase) * 0.8
	return ui.Row(c).AlignItems(ui.Center).Gap(0.3 * fs).Children(func() {
		for _, k := range keys {
			e := ui.Box(c).Center().MinWidth(1.8*fs).Height(1.8*fs).Padding(0, 0.45*fs).Radius(0.4*fs).
				LinearGradient(ui.LinearGradient{Angle: 180, Oklab: true, From: cl.ControlTop, To: cl.ControlBottom}).
				Border(0.5, cl.ControlBorder)
			t.Shadows.ControlShadow.Apply(e)
			e.Children(func() {
				ui.Text(c, k).TextColor(cl.Text2).FontSize(fs).FontWeight(500).NoWrap().FontFeatures("tnum")
			})
		}
	})
}

// AvatarOptions, Avatar'ın seçenekleridir.
type AvatarOptions struct {
	// Image, baş harflerin yerine gösterilen resimdir.
	Image *ui.Bitmap
	// Color verilmezse addan seri renklerinden (C1…C8) biri seçilir: aynı
	// ad her yerde aynı renk.
	Color ui.Color
	// Solid, zemini rengin kendisi, baş harfleri beyaz yapar.
	Solid bool
	// Online, sağ altta yeşil durum noktası gösterir.
	Online bool
	// Size, çaptır (varsayılan 32).
	Size float32
}

// Avatar, kişinin resmi ya da baş harfleridir: zemin rengin %16'sı kart
// üstünde, harfler rengin %62'si metin üstünde (iki temada okunur).
//
//	mizan.Avatar(c, "Ayşe Yılmaz", mizan.AvatarOptions{Online: true})
func Avatar(c *ui.Context, name string, o AvatarOptions) ui.Element {
	t := From(c)
	cl := t.Colors
	size := or(o.Size, 32)
	av := o.Color
	if av == (ui.Color{}) {
		av = t.series(name)
	}
	bg, fg := Mix(cl.Card, av, 0.16), Mix(cl.Text, av, 0.62)
	if o.Solid {
		bg, fg = av, cl.OnAccent
	}
	e := ui.Box(c).Size(size, size).Radius(size / 2).Center().Shrink(0).Background(bg).Role(ui.RoleImage).Label(name)
	e.Children(func() {
		if o.Image != nil {
			ui.Image(c, o.Image).Fit(ui.Cover).Size(size, size).Radius(size / 2)
		} else {
			ui.Text(c, initials(name)).TextColor(fg).FontSize(size * 0.4).FontWeight(600).SingleLine().Unselectable()
		}
		if o.Online {
			s := size * 0.3
			ui.Box(c).Absolute().Right(-1).Bottom(-1).Size(s, s).Radius(s/2).Background(cl.Down).Shadow(0, 0, 0, 2, cl.Card)
		}
	})
	return e
}

// series, s'nin FNV-1a özetinden bir seri rengi seçer.
func (t *Tokens) series(s string) ui.Color {
	h := uint32(2166136261)
	for i := 0; i < len(s); i++ {
		h = (h ^ uint32(s[i])) * 16777619
	}
	c := t.Colors
	return [...]ui.Color{c.C1, c.C2, c.C3, c.C4, c.C5, c.C6, c.C7, c.C8}[h%8]
}

// initials, adın ilk ve son sözcüğünün baş harfleridir.
func initials(name string) string {
	var first, last rune
	in := false
	for _, r := range name {
		letter := unicode.IsLetter(r) || unicode.IsDigit(r)
		if letter && !in {
			if first == 0 {
				first = r
			} else {
				last = r
			}
		}
		in = letter
	}
	s := ""
	if first != 0 {
		s += string(unicode.ToUpper(first))
	}
	if last != 0 {
		s += string(unicode.ToUpper(last))
	}
	return s
}

// EmptyStateOptions, EmptyState'in seçenekleridir.
type EmptyStateOptions struct {
	Icon        *ui.SVG
	Title       string
	Description string
	// Action, altta tek eylemi (çoğunlukla kabarık bir Button) kurar.
	Action func()
}

// EmptyState, boş, hata ya da ilk kullanım durumudur: ortalı sütun, renk
// vurgusu yok (ikon da nötr bir kuyuda); durum metinle anlatılır.
func EmptyState(c *ui.Context, o EmptyStateOptions) ui.Element {
	cl := From(c).Colors
	return ui.Column(c).AlignItems(ui.Center).Gap(6).Padding(32, 20).Children(func() {
		if o.Icon != nil {
			ui.Box(c).Size(44, 44).Radius(22).Center().Background(cl.Well).Margin(0, 0, 6, 0).TextColor(cl.Muted).
				Children(func() { ui.Icon(c, o.Icon) })
		}
		if o.Title != "" {
			ui.Text(c, o.Title).TextColor(cl.Text).FontSize(FsBase).FontWeight(600).TextAlign(ui.Center)
		}
		if o.Description != "" {
			ui.Text(c, o.Description).TextColor(cl.Muted).FontSize(FsMd).Wrap().TextAlign(ui.Center).MaxWidth(44 * FsMd * 0.5)
		}
		if o.Action != nil {
			ui.Row(c).Margin(10, 0, 0, 0).Children(o.Action)
		}
	})
}
