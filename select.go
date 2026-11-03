package mizan

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// Option, seçimdeki bir seçenektir; her yerde aynı biçim: değer ve etiket.
type Option[T comparable] struct {
	Value T
	Label string
	// Description, etiketin yanındaki soluk açıklamadır.
	Description string
	// Dot, seçeneğin önündeki seri rengi noktasıdır (Colors.C1…C8).
	Dot      ui.Color
	Disabled bool
}

// SelectOptions, Select'in seçenekleridir.
type SelectOptions struct {
	// Placeholder, değer seçenekler arasında değilken görünen soluk metindir.
	Placeholder string
	Label       string
	Size        Size
	Disabled    bool
	Invalid     bool
}

// Select, aramasız seçim listesidir: tetikleyici kabarık alan (düğmeyle
// aynı malzeme, alanın ölçüleri), liste cam açılır panel; gezilen seçenek
// vurgu gradyanıyla dolar, seçilenin yanında onay işareti durur.
//
//	mizan.Select(c.Key("range"), &app.rangeDays, []mizan.Option[int]{
//		{Value: 7, Label: "Son 7 gün"},
//		{Value: 30, Label: "Son 30 gün"},
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
		// açık alan: vurgu kenarı ve yumuşak halka (Input'un odağı gibi)
		tr.BorderColor(cl.ColorAccent500)
		t.Shadows.Ring.Apply(tr)
	}
	if tr.Pressed() && !o.Disabled {
		// alan basınca aşağı inmez, hafifçe küçülür
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

// dot, 8px seri rengi noktasıdır (Chips ile aynı).
func dot(c *ui.Context, color ui.Color) ui.Element {
	return ui.Box(c).Size(8, 8).Radius(RadiusPill).Background(color).Shrink(0)
}

// popover, açılır yüzeyi boyar: cam zemin, köşe radius-pop, 0,5px çizgi ve
// iki katman yayılan gölge; tetikleyiciden gap-field uzakta, %96 boyuttan
// saydam başlayarak açılır.
func popover(t *Tokens, panel ui.Element) {
	// Mizan'da cam: --pop yarı saydam + backdrop-filter bulanıklığı. MyGo'da
	// arkayı bulanıklaştıran katman yok; saydam bırakılırsa alttaki yazı
	// okunurdu, renk kart zemininin üstünde opak hâle getirilir.
	panel.Margin(GapField, 0, 0, 0).Padding(PadPop).Radius(RadiusPop).Background(over(t.Colors.Pop, t.Colors.Card)).
		TextColor(t.Colors.Text)
	t.Shadows.ShadowPop.Apply(panel)
	panel.Transition(ui.ElementTransition{
		Duration: 180 * time.Millisecond,
		Enter:    &ui.Motion{Opacity: 0, Scale: 0.96},
	})
}
