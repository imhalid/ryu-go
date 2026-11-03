package mizan

import (
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

// Sohbet Mizan'da yok; dili buradan kurulur: kendi mesajın solid düğmenin
// malzemesi (vurgu gradyanı, beyaz yazı), karşınınki ikinci yüzey ve saç
// teli kenar. Etiketler alan etiketi gibi (fs-sm 500 soluk), saatler
// tabular. Aynı kişinin ardışık mesajları bir grup: ad yalnız ilkinde,
// avatar ve kuyruk köşesi yalnız sonuncusunda.

var (
	iconSend        = Svg(16, 16, "M8 12.5v-9M4.5 7L8 3.5 11.5 7")
	iconAttach      = Svg(16, 16, "M13 7.5l-5.2 5.2a3.2 3.2 0 0 1-4.5-4.5l5.5-5.5a2.1 2.1 0 0 1 3 3L6.3 11.2a1 1 0 0 1-1.5-1.5L9.5 5")
	iconSent        = Svg(16, 13, "M3.5 8.5l3 3 6-7")
	iconRead        = Svg(16, 13, "M1.5 8.5l3 3 6-7M8 11l.5.5 6-7")
	iconPending     = Svg(16, 12, "M8 4.5V8l2 1.5M14 8A6 6 0 1 1 2 8a6 6 0 0 1 12 0z")
	bubbleR, tailR  = float32(18), float32(6)
	bubbleAvatarDia = float32(28)
)

// Delivery, kendi mesajının iletim durumudur.
type Delivery uint8

const (
	DeliveryNone Delivery = iota
	// Sending, saat ikonu.
	Sending
	// Sent, tek onay.
	Sent
	// Read, vurgu renginde çift onay.
	Read
)

// ChatMessage, sohbetteki bir mesajdır.
type ChatMessage struct {
	Author string
	Mine   bool
	Text   string
	// Body verilirse Text yerine balonun içini kurar (Markdown, resim…).
	Body   func()
	Time   string
	Status Delivery
	// Avatar, karşı tarafın resmi ve rengi (yoksa addan).
	Avatar AvatarOptions
}

// Group, mesajın grubundaki yeridir: ilk mesaj adı, son mesaj avatarı ve
// kuyruğu taşır. GroupOf hesaplar.
type Group struct{ First, Last bool }

// GroupOf, msgs[i]'nin grubundaki yerini verir: aynı yazarın ardışık
// mesajları bir gruptur.
func GroupOf(msgs []ChatMessage, i int) Group {
	same := func(j int) bool {
		return j >= 0 && j < len(msgs) && msgs[j].Mine == msgs[i].Mine && msgs[j].Author == msgs[i].Author
	}
	return Group{First: !same(i - 1), Last: !same(i + 1)}
}

// Message, mesajı balonuyla, adıyla ve saatiyle kurar. Kendi mesajın sağda,
// karşınınki solda avatarıyla.
//
//	for i, m := range app.msgs {
//		mizan.Message(c.Key(i), m, mizan.GroupOf(app.msgs, i))
//	}
func Message(c *ui.Context, m ChatMessage, g Group) ui.Element {
	t := From(c)
	cl := t.Colors
	row := ui.Row(c).FillWidth().AlignItems(ui.End).Gap(8)
	if m.Mine {
		row.Justify(ui.End)
	}
	if g.Last {
		row.Margin(0, 0, 10, 0)
	}
	row.Children(func() {
		if !m.Mine {
			if g.Last {
				a := m.Avatar
				a.Size = or(a.Size, bubbleAvatarDia)
				// avatar balonun alt kenarına, saat satırının üstüne oturur
				Avatar(c, m.Author, a).Margin(0, 0, 18, 0)
			} else {
				ui.Box(c).Size(bubbleAvatarDia, 0).Shrink(0)
			}
		}
		col := ui.Column(c).Gap(3).MaxWidthPercent(72).MinWidth(0)
		if m.Mine {
			col.AlignItems(ui.End)
		} else {
			col.AlignItems(ui.Start)
		}
		col.Children(func() {
			if g.First && !m.Mine && m.Author != "" {
				ui.Text(c, m.Author).TextColor(cl.Muted).FontSize(FsSm).FontWeight(500).Padding(0, 12)
			}
			b := bubble(c, t, m.Mine, g.Last)
			b.Children(func() {
				if m.Body != nil {
					m.Body()
					return
				}
				ui.Text(c, m.Text).Wrap().FontSize(FsBase).LineHeight(1.4)
			})
			if g.Last && (m.Time != "" || m.Status != DeliveryNone) {
				ui.Row(c).AlignItems(ui.Center).Gap(4).Padding(0, 6).Children(func() {
					ui.Text(c, m.Time).TextColor(cl.Muted).FontSize(FsXs).FontFeatures("tnum").NoWrap()
					switch m.Status {
					case Sending:
						ui.Icon(c, iconPending).TextColor(cl.Muted).Label("Gönderiliyor")
					case Sent:
						ui.Icon(c, iconSent).TextColor(cl.Muted).Label("Gönderildi")
					case Read:
						ui.Icon(c, iconRead).TextColor(cl.Accent).Label("Okundu")
					}
				})
			}
		})
	})
	return row
}

// bubble, balonun yüzeyidir. Grubun son balonunun konuşan tarafa bakan alt
// köşesi kuyruktur (küçük yarıçap).
func bubble(c *ui.Context, t *Tokens, mine, tail bool) ui.Element {
	cl := t.Colors
	b := ui.Column(c).Padding(8, 12).MinWidth(0)
	bl, br := bubbleR, bubbleR
	if tail && mine {
		br = tailR
	} else if tail {
		bl = tailR
	}
	b.Radius(bubbleR, bubbleR, br, bl)
	if mine {
		// solid düğmeyle aynı: açıklık 0,52/0,46'da sınırlı, beyaz yazı okunur
		top := capLightness(cl.Accent, 0.52)
		bottom := capLightness(lighten(cl.Accent, -0.06), 0.46)
		b.LinearGradient(ui.LinearGradient{Angle: 180, Oklab: true, From: top, To: bottom}).TextColor(cl.OnAccent)
		t.Shadows.Sheen.With(ShadowLayer{Y: 1, Blur: 2, Color: Mix(cl.Accent, ui.Color{}, 0.55)}).Apply(b)
	} else {
		b.Background(cl.Surface2).Border(1, cl.Hairline).TextColor(cl.Text)
	}
	return b
}

// Typing, karşı tarafın yazdığını gösteren balondur: üç nokta sırayla
// kalkıp iner.
func Typing(c *ui.Context, author string, o AvatarOptions) ui.Element {
	t := From(c)
	cl := t.Colors
	return ui.Row(c).AlignItems(ui.End).Gap(8).Margin(0, 0, 10, 0).Children(func() {
		o.Size = or(o.Size, bubbleAvatarDia)
		Avatar(c, author, o)
		b := bubble(c, t, false, true).Label(author + " yazıyor")
		b.Padding(0, 14).Height(36).Justify(ui.Center).Children(func() {
			ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
				for i := range 3 {
					d := ui.Box(c.Key(i)).Size(6, 6).Radius(3).Background(cl.Muted)
					v := d.Keyframes("bob", []ui.Keyframe{
						{At: 0, Value: 0},
						{At: 300 * time.Millisecond, Value: 1, Ease: ui.EaseOut},
						{At: 600 * time.Millisecond, Value: 0, Ease: ui.EaseIn},
						{At: 1200 * time.Millisecond, Value: 0},
					}, ui.KeyframeOptions{Repeat: ui.RepeatForever, Delay: time.Duration(i) * 150 * time.Millisecond})
					d.Translate(0, -3*v).Opacity(0.45 + 0.55*v)
				}
			})
		})
	})
}

// DaySeparator, mesajlar arasındaki gün ayracıdır: iki yanı saç teli,
// ortada soluk etiket.
func DaySeparator(c *ui.Context, label string) ui.Element {
	cl := From(c).Colors
	return ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(12).Padding(8, 0, 14, 0).Children(func() {
		ui.Box(c).Grow(1).Height(1).Background(cl.Hairline)
		ui.Text(c, label).TextColor(cl.Muted).FontSize(FsXs).FontWeight(500).NoWrap()
		ui.Box(c).Grow(1).Height(1).Background(cl.Hairline)
	})
}

// ComposerOptions, Composer'ın seçenekleridir.
type ComposerOptions struct {
	Placeholder string
	// Attach verilirse soldaki ataç düğmesine basılınca çağrılır.
	Attach   func()
	Disabled bool
}

// Composer, mesaj yazma alanıdır: Input'un gömük kuyusu, sağda solid
// yuvarlak gönder düğmesi. Enter ya da düğme gönderir: gönderildiği
// çerçevede true döner; metni boşaltmak çağıranındır.
//
//	if mizan.Composer(c.Key("composer"), &app.draft, mizan.ComposerOptions{Placeholder: "Mesaj"}) {
//		app.send()
//	}
func Composer(c *ui.Context, value *string, o ComposerOptions) bool {
	t := From(c)
	m := t.Metrics
	cl := t.Colors
	empty := strings.TrimSpace(*value) == ""
	sent := false
	h := m.FieldH + 8
	box := ui.Row(c).FillWidth().AlignItems(ui.Center).Height(h).Gap(6).Radius(h/2).Padding(0, 4, 0, 6).Cursor(ui.CursorText)
	var field ui.Element
	box.Children(func() {
		if o.Attach != nil {
			if inputButton(c.Key("attach"), t, iconAttach, "Dosya ekle").Margin(0).Clicked() {
				o.Attach()
			}
		} else {
			ui.Box(c).Size(6, 0)
		}
		field = ui.TextInputBase(c, value).Grow(1).MinWidth(0).Label("Mesaj").
			Placeholder(o.Placeholder).PlaceholderColor(cl.Muted).TextColor(cl.Text).FontSize(FsBase).
			Disabled(o.Disabled).FocusRing(false)
		if field.Submitted() && !empty {
			sent = true
		}
		d := h - 8
		send := ui.ButtonBase(c.Key("send")).Label("Gönder").Center().Size(d, d).Radius(d / 2).Cursor(ui.CursorPointer).
			Disabled(empty || o.Disabled)
		off := empty || o.Disabled
		top := capLightness(cl.Accent, 0.52)
		bottom := capLightness(lighten(cl.Accent, -0.06), 0.46)
		g := ui.LinearGradient{Angle: 180, Oklab: true, From: top, To: bottom}
		sh := t.Shadows.Sheen.With(ShadowLayer{Y: 1, Blur: 2, Color: Mix(cl.Accent, ui.Color{}, 0.55)})
		t.raised(send, g, Mix(bottom, oklch(0, 0, 0, 1), 0.2), sh,
			raisedState{hovered: send.Hovered() && !off, pressed: send.Pressed() && !off, focus: send.FocusVisible(), disabled: off})
		send.TextColor(cl.OnAccent).Children(func() { ui.Icon(c, iconSend) })
		if send.Clicked() && !off {
			sent = true
		}
	})
	t.well(box, field.Focused(), false)
	if o.Disabled {
		box.Opacity(0.4)
	} else if box.Clicked() {
		field.Focus()
	}
	if sent {
		field.Focus()
	}
	return sent
}
