package mizan

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// Seçim kontrolleri gömük başlar (kuyu, axis kenarı) ve seçilince vurgu
// gradyanıyla dolar; topuzlar (Switch, Slider) kuyunun üstünde duran
// beyaz ya da bevel'de metal disklerdir.

var iconDash = Svg(16, 12, "M4 8h8")

// accentFill, seçim işaretinin dolgusudur (CSS --accent-fill).
func (t *Tokens) accentFill() ui.LinearGradient {
	return ui.LinearGradient{Angle: 180, Oklab: true, From: t.Colors.ColorAccent500, To: t.Colors.ColorAccent600}
}

// mark, Checkbox ve Radio'nun kutusunu boyar: boşken kuyu, seçiliyken
// vurgu gradyanı, üstte parlama ve altta vurguya çalan temas gölgesi.
// Basılıyken %92'ye küçülür; klavye odağı halkayı gölgenin yerine koyar.
func (t *Tokens) mark(e ui.Element, on, pressed, focus bool) {
	cl := t.Colors
	sh := t.Shadows.WellShadow
	if on {
		e.LinearGradient(t.accentFill()).Border(1, cl.ColorAccent600)
		sh = t.Shadows.SheenAccent.With(ShadowLayer{Y: 1, Blur: 2, Color: Mix(cl.ColorAccent700, ui.Color{}, 0.6)})
	} else {
		e.Background(cl.Well).Border(1, cl.Axis)
	}
	if focus {
		sh = t.Shadows.FocusRing
	}
	sh.Apply(e)
	if pressed {
		e.Scale(0.92, 0.92)
	}
	e.Transition(ui.ElementTransition{Duration: 150 * time.Millisecond, Colors: true, Ease: ui.EaseOut})
}

// pop, işaretin içindekini (tik, nokta) yarı boyuttan, saydam başlatıp
// yayla yerine oturtur.
func pop(e ui.Element, on bool) {
	v := e.Spring("pop", b2f(on), ui.SpringOptions{Duration: 220 * time.Millisecond, Bounce: 0.35})
	s := 0.45 + 0.55*v
	e.Scale(s, s).Opacity(clamp01(v))
}

func b2f(b bool) float32 {
	if b {
		return 1
	}
	return 0
}

func clamp01(v float32) float32 { return min(max(v, 0), 1) }

// CheckOptions, Checkbox'ın seçenekleridir.
type CheckOptions struct {
	// Description, etiketin altındaki soluk açıklamadır.
	Description string
	// Indeterminate, kutuyu yarım (çizgi) gösterir: bazı alt seçenekler seçili.
	Indeterminate bool
	Disabled      bool
}

// Checkbox, 18px kare onay kutusu ve etiketidir: boşken gömük, seçiliyken
// vurgu gradyanı ve beyaz tik. Satırın tamamı tıklanır.
//
//	mizan.Checkbox(c.Key("mail"), &app.mail, "E-posta bildirimleri", mizan.CheckOptions{})
func Checkbox(c *ui.Context, checked *bool, label string, o CheckOptions) ui.Element {
	t := From(c)
	e := ui.CheckboxBase(c, checked).Disabled(o.Disabled)
	if label != "" {
		e.Label(label)
	}
	on := *checked || o.Indeterminate
	t.choiceRow(c, e, label, o.Description, o.Disabled, func() {
		box := ui.Box(c).Size(18, 18).Radius(6).Center().Shrink(0)
		t.mark(box, on, e.Pressed() && !o.Disabled, e.FocusVisible())
		box.Children(func() {
			icon := iconCheck
			if o.Indeterminate {
				icon = iconDash
			}
			i := ui.Icon(c, icon).Size(12, 12).TextColor(t.Colors.OnAccent)
			pop(i, on)
		})
	})
	return e
}

// Radio, 18px yuvarlak seçim düğmesi ve etiketidir: seçiliyken vurgu
// gradyanı ortasında 7px beyaz nokta. ui.RadioGroup içinde kullanın.
//
//	ui.RadioGroup(c, func() {
//		mizan.Radio(c.Key("d"), &app.plan, "day", "Günlük", mizan.CheckOptions{})
//		mizan.Radio(c.Key("w"), &app.plan, "week", "Haftalık", mizan.CheckOptions{})
//	}).Label("Plan")
func Radio[T comparable](c *ui.Context, selected *T, value T, label string, o CheckOptions) ui.Element {
	t := From(c)
	e := ui.RadioBase(c, selected, value).Disabled(o.Disabled)
	if label != "" {
		e.Label(label)
	}
	on := *selected == value
	t.choiceRow(c, e, label, o.Description, o.Disabled, func() {
		box := ui.Box(c).Size(18, 18).Radius(9).Center().Shrink(0)
		t.mark(box, on, e.Pressed() && !o.Disabled, e.FocusVisible())
		box.Children(func() {
			pop(ui.Box(c).Size(7, 7).Radius(3.5).Background(t.Colors.OnAccent), on)
		})
	})
	return e
}

// choiceRow, işaret ile etiketi yan yana dizer: açıklama varsa işaret ilk
// satırın ortasında durur.
func (t *Tokens) choiceRow(c *ui.Context, e ui.Element, label, desc string, disabled bool, mark func()) {
	cl := t.Colors
	e.Gap(10).AlignItems(ui.Start).Cursor(ui.CursorPointer).FocusRing(false)
	if disabled {
		e.Opacity(0.4).Cursor(ui.CursorDefault)
	}
	e.Children(func() {
		ui.Box(c).Height(20).Center().Children(mark)
		if label == "" && desc == "" {
			return
		}
		ui.Column(c).Gap(2).MinWidth(0).Children(func() {
			if label != "" {
				ui.Text(c, label).TextColor(cl.Text).FontSize(FsMd).LineHeight(20.0 / FsMd)
			}
			if desc != "" {
				ui.Text(c, desc).TextColor(cl.Muted).FontSize(FsXs).Wrap()
			}
		})
	})
}

// Choice, kart biçiminde seçim düğmesidir: kuyuda durur, seçilince
// kabarık malzemeye geçer (gölgesi ve kenarıyla öne çıkar). Seçenekler
// ızgarada yan yana, ui.RadioGroup içinde.
func Choice[T comparable](c *ui.Context, selected *T, value T, title, desc string, o CheckOptions) ui.Element {
	t := From(c)
	cl := t.Colors
	e := ui.RadioBase(c, selected, value).Disabled(o.Disabled).Label(title)
	on := *selected == value
	e.Gap(10).AlignItems(ui.Start).Padding(12, 14).Radius(RadiusControl).Cursor(ui.CursorPointer).FocusRing(false).MinWidth(0)
	switch {
	case on:
		e.LinearGradient(t.controlGradient(cl.ControlTop, cl.ControlBottom)).Border(1, cl.ControlBorder)
		t.Shadows.ShadowRaised.Apply(e)
	case e.Hovered() && !o.Disabled:
		e.Background(Mix(cl.Well, cl.Text, 0.06)).Border(1, ui.Color{})
		t.Shadows.WellShadow.Apply(e)
	default:
		e.Background(cl.Well).Border(1, ui.Color{})
		t.Shadows.WellShadow.Apply(e)
	}
	if e.FocusVisible() {
		t.Shadows.FocusRing.Apply(e)
	}
	if e.Pressed() && !o.Disabled {
		e.Scale(0.98, 0.98)
	}
	if o.Disabled {
		e.Opacity(0.4).Cursor(ui.CursorDefault)
	}
	e.Transition(ui.ElementTransition{Duration: 200 * time.Millisecond, Colors: true})
	e.Children(func() {
		ui.Box(c).Height(20).Center().Children(func() {
			box := ui.Box(c).Size(18, 18).Radius(9).Center().Shrink(0)
			t.mark(box, on, false, false)
			box.Children(func() { pop(ui.Box(c).Size(7, 7).Radius(3.5).Background(cl.OnAccent), on) })
		})
		ui.Column(c).Gap(2).MinWidth(0).Grow(1).Children(func() {
			ui.Text(c, title).TextColor(cl.Text).FontSize(FsMd).FontWeight(600).LineHeight(20.0 / FsMd)
			if desc != "" {
				ui.Text(c, desc).TextColor(cl.Muted).FontSize(FsXs).Wrap()
			}
		})
	})
	return e
}

// knob, topuzun yüzeyidir: Mizan'da beyaz disk, bevel'de yukarıdan aşağı
// koyulaşan metal.
func (t *Tokens) knob(e ui.Element) {
	if t.Skin == SkinBevel {
		e.LinearGradient(ui.LinearGradient{Angle: 180, Oklab: true, From: t.Colors.KnobTop, To: t.Colors.KnobBot})
		return
	}
	e.Background(t.Colors.Thumb)
}

// SwitchOptions, Switch'in seçenekleridir.
type SwitchOptions struct {
	Label       string
	Description string
	Disabled    bool
}

// Switch, 38×22 anahtardır: iz koyu kuyu, açıkken vurgu gradyanıyla dolar
// (bevel'de iz aynı kalır, boşalan tarafta 6px LED yanar); topuz yayla
// kayar, basılı tutunca kısa bir kapsüle uzar.
//
//	mizan.Switch(c.Key("sync"), &app.sync, mizan.SwitchOptions{Label: "Eşitleme"})
func Switch(c *ui.Context, on *bool, o SwitchOptions) ui.Element {
	t := From(c)
	cl := t.Colors
	e := ui.SwitchBase(c, on).Disabled(o.Disabled)
	if o.Label != "" {
		e.Label(o.Label)
	}
	pressed := e.Pressed() && !o.Disabled
	t.choiceRow(c, e, o.Label, o.Description, o.Disabled, func() {
		track := ui.Box(c).Size(38, 22).Radius(RadiusPill).Background(cl.WellStrong).Shrink(0)
		sh := t.Shadows.WellShadow
		if e.FocusVisible() {
			sh = t.Shadows.FocusRing
		}
		sh.Apply(track)
		lit := track.Animate("on", b2f(*on), 200*time.Millisecond)
		x := track.Spring("x", 16*b2f(*on), ui.SpringOptions{Duration: 300 * time.Millisecond, Bounce: 0.25})
		w := track.Animate("w", 18+4*b2f(pressed), 200*time.Millisecond)
		if *on {
			// uzayan topuz satır başına doğru büyür
			x -= w - 18
		}
		track.Children(func() {
			if t.Skin == SkinBevel {
				led := ui.Box(c).Absolute().Left(7).Top(8).Size(6, 6).Radius(3).Opacity(lit).
					RadialGradient(ui.RadialGradient{Center: ui.Point{X: 0.5, Y: 0.35}, Oklab: true, Stops: []ui.GradientStop{
						{Color: cl.LedCore, At: 0}, {Color: cl.LedColor, At: 0.7},
					}})
				t.Shadows.SwitchOnGlow.Apply(led)
			} else {
				ui.Box(c).Absolute().Left(0).Top(0).Right(0).Bottom(0).Radius(RadiusPill).
					LinearGradient(t.accentFill()).Opacity(lit)
			}
			k := ui.Box(c).Absolute().Left(2).Top(2).Size(w, 18).Radius(RadiusPill).Translate(x, 0)
			t.knob(k)
			t.Shadows.KnobShadow.Apply(k)
		})
	})
	return e
}

// SegmentedOptions, Segmented'ın seçenekleridir.
type SegmentedOptions struct {
	// Label, grubun erişilebilirlik adıdır.
	Label string
	Size  Size
	// Fill, rayı satır boyunca yayar; seçenekler eşit paylaşır.
	Fill     bool
	Disabled bool
}

// Segmented, birkaç görünüm arasında seçimdir: ray gömük ve yanındaki
// alanla aynı boyda (control-h); seçili seçenek rayın içinde kabarık bir
// hap. Seçim yokken seçenekler oyuk ayraçlarla ayrı düğmeler gibi okunur;
// hapın iki yanındaki ayraç söner.
//
//	mizan.Segmented(c.Key("view"), &app.view, []string{"Liste", "Izgara"}, mizan.SegmentedOptions{Label: "Görünüm"})
func Segmented(c *ui.Context, selected *int, labels []string, o SegmentedOptions) ui.Element {
	t := sizedFrom(c, o.Size)
	m := t.Metrics
	cl := t.Colors
	parts := ui.SegmentedBase(c, selected, len(labels))
	tr := parts.Track.Height(m.ControlH).Padding(3).Radius(RadiusControl).Background(cl.Well).AlignItems(ui.Stretch).Disabled(o.Disabled)
	if o.Label != "" {
		tr.Label(o.Label)
	}
	t.Shadows.WellShadow.Apply(tr)
	if o.Fill {
		tr.FillWidth()
	} else {
		// ray içeriği kadar: sütunun genişliğine yayılmaz
		tr.AlignSelf(ui.Start)
	}
	if o.Disabled {
		tr.Opacity(0.4)
	}
	tr.Children(func() {
		for i, l := range labels {
			if i > 0 {
				sep := ui.Box(c).Width(1).Margin(m.ControlH*0.25-3, 0).Background(cl.SepDark).Shadow(1, 0, 0, 0, cl.SepLight).Shrink(0)
				if *selected == i || *selected == i-1 {
					sep.Opacity(0)
				}
				sep.Transition(ui.ElementTransition{Duration: 150 * time.Millisecond, Colors: true})
			}
			s := parts.Segment(i).Center().Padding(0, m.SegPx).Radius(RadiusInner).Cursor(ui.CursorPointer).FocusRing(false)
			if o.Fill {
				s.Grow(1).Basis(0)
			}
			on := *selected == i
			text := cl.Muted
			if on {
				text = cl.Text
				s.LinearGradient(t.controlGradient(cl.ControlTop, cl.ControlBottom)).Border(0.5, cl.ControlBorder)
				t.Shadows.ShadowRaised.Apply(s)
			} else {
				s.Border(0.5, ui.Color{})
				if s.Hovered() && !o.Disabled {
					text = cl.Text
				}
			}
			if s.FocusVisible() {
				s.Border(2, Mix(ui.Color{}, cl.ColorAccent500, 0.55))
			}
			if s.Pressed() && !o.Disabled {
				s.Scale(0.97, 0.97)
			}
			s.Transition(ui.ElementTransition{Duration: 200 * time.Millisecond, Colors: true})
			s.Children(func() {
				ui.Text(c, l).TextColor(text).FontSize(m.SegFs).FontWeight(500).NoWrap()
			})
		}
	})
	return tr
}

// SliderOptions, Slider'ın seçenekleridir.
type SliderOptions struct {
	Label string
	// Step, değerin adımıdır (yoksa aralığın yüzde biri).
	Step float64
	// Format verilirse değer sağda, sabit genişlikte, tabular yazılır.
	Format   func(v float64) string
	Disabled bool
}

// Slider, kaydırıcıdır: 6px gömük iz, dolu kısmı vurgu gradyanı; 22px
// topuz basılıyken %12 büyür. Format ile sağında değeri durur.
//
//	mizan.Slider(c.Key("vol"), &app.volume, 0, 100, mizan.SliderOptions{Label: "Ses"})
func Slider(c *ui.Context, value *float64, lo, hi float64, o SliderOptions) ui.Element {
	t := From(c)
	cl := t.Colors
	var s ui.Element
	slider := func() {
		const k, track = 22, 6
		s = ui.SliderBase(c, value, lo, hi).Height(28).Padding(0, k/2).Grow(1).MinWidth(80).
			Cursor(ui.CursorPointer).FocusRing(false).Disabled(o.Disabled)
		if o.Step > 0 {
			s.Step(o.Step)
		}
		if o.Label != "" {
			s.Label(o.Label)
		}
		if o.Disabled {
			s.Opacity(0.4).Cursor(ui.CursorDefault)
		}
		frac := float32(0)
		if hi > lo {
			frac = clamp01(float32((*value - lo) / (hi - lo)))
		}
		pressed := s.Pressed() && !o.Disabled
		focus := s.FocusVisible()
		s.Children(func() {
			ui.Box(c).FillWidth().Height(28).Children(func() {
				y := float32(28-track) / 2
				well := ui.Box(c).Absolute().Left(-k / 2).Right(-k / 2).Top(y).Height(track).Radius(RadiusPill).Background(cl.Well)
				t.Shadows.WellEdge.Apply(well)
				ui.Box(c).Absolute().Left(-k / 2).RightPercent(100 * (1 - frac)).Top(y).Height(track).
					Radius(RadiusPill).LinearGradient(t.accentFill())
				kn := ui.Box(c).Absolute().LeftPercent(100*frac).Top(3).Size(k, k).Radius(k/2).Translate(-k/2, 0)
				t.knob(kn)
				sh := t.Shadows.KnobShadowLg
				if focus {
					sh = t.Shadows.FocusRing.With(t.Shadows.KnobDrop...)
				}
				sh.Apply(kn)
				sc := kn.Animate("held", 1+0.12*b2f(pressed), 150*time.Millisecond)
				kn.Scale(sc, sc)
			})
		})
	}
	if o.Format == nil {
		slider()
		return s
	}
	ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(12).Children(func() {
		slider()
		// sütun en geniş olası metne göre: değer basamak değiştirince iz
		// daralıp topuz imleçten kaymasın
		w := max(56, 8*float32(max(len(o.Format(lo)), len(o.Format(hi)))))
		ui.Text(c, o.Format(*value)).TextColor(cl.Text).FontSize(FsMd).FontWeight(650).FontFeatures("tnum").
			TextAlign(ui.End).Width(w).Shrink(0)
	})
	return s
}
