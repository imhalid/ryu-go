package mizan

import "github.com/egoist/mygo/ui"

// Variant, renkli düğmenin dolgusudur.
type Variant uint8

const (
	// Soft, renge çalan kabarık yüzey; etiket renk ile metin arasında.
	Soft Variant = iota + 1
	// Solid, renkli gradyan, beyaz etiket.
	Solid
)

// ButtonOptions, Button'ın seçenekleridir; sıfır değer kabarık nötr düğmedir.
type ButtonOptions struct {
	// Color verilirse düğme o renge çalar (Colors.C1…C8, Accent, Up, Down).
	Color ui.Color
	// Variant, Color ile: Soft (varsayılan) ya da Solid.
	Variant Variant
	// Danger, yıkıcı eylem: Up renginde, varsayılan Solid. Color verilirse yok sayılır.
	Danger bool
	// Icon, etiketin önündeki ikon; IconOnly ile kare ikon düğmesi
	// (kenarı control-h), etiket yalnız erişilebilirlik adı olur.
	Icon     *ui.SVG
	IconOnly bool
	// Size verilmezse çevredeki kapsamdan.
	Size Size
	// Loading, dönen işaret gösterir ve tıklamayı kapatır; soluklaşmaz.
	Loading  bool
	Disabled bool
	// Block, düğmeyi satır boyunca genişletir.
	Block bool
}

// Button, kabarık hap düğmedir: tek tür, boyu aynı kademedeki alanlarla
// aynı (control-h). Clicked ile basıldığını bildirir.
//
//	if mizan.Button(c.Key("export"), "Dışa aktar", mizan.ButtonOptions{Color: t.Colors.Accent, Variant: mizan.Solid}).Clicked() {
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
		// dolgunun açıklığı OKLCH'de sınırlanır (üst 0,52, alt 0,46): ton ve
		// canlılık kalır, beyaz yazı ≥ 4,5:1 okunur
		top := capLightness(tint, 0.52)
		bottom := capLightness(lighten(tint, -0.06), 0.46)
		g := ui.LinearGradient{Angle: 180, Oklab: true, From: top, To: bottom}
		sh := t.Shadows.Sheen.With(ShadowLayer{Y: 1, Blur: 2, Color: Mix(tint, ui.Color{}, 0.55)})
		t.raised(e, g, Mix(bottom, oklch(0, 0, 0, 1), 0.2), sh, s)
		text = cl.OnAccent
	default:
		g := t.controlGradient(Mix(cl.ControlTop, tint, 0.1), Mix(cl.ControlBottom, tint, 0.2))
		t.raised(e, g, Mix(cl.ControlBorder, tint, 0.32), t.Shadows.ControlShadow, s)
		// renk ile metin rengi yarı yarıya: iki temada ≥ 4,5:1
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

// sizedFrom, kapsamın token'larını bileşenin kendi boyut kademesiyle verir
// (bileşenin Size'ı çevreden geleni ezer). Sarmalayıcı öğe kurmaz: c'nin
// anahtarı (c.Key) bileşenin kendi öğesine kalır.
func sizedFrom(c *ui.Context, s Size) *Tokens {
	if s == SizeInherit {
		return From(c)
	}
	return scoped(c, Options{Size: s})
}
