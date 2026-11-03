package mizan

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// Yüzeyin iki dili vardır: kabarık (düğme, Select tetikleyicisi) ışığı
// üstten alır; gömük (Input, Checkbox, Segmented rayı) opak bir kuyudur.
// Bileşenler bu iki fonksiyonla boyanır, kendi gradyanını yazmaz.

// raisedState, kabarık yüzeyin etkileşim hâlidir.
type raisedState struct {
	hovered, pressed, focus, disabled bool
}

// controlGradient, kabarık kontrolün zeminidir (CSS --control). Mizan'da
// en koyu bant altın %72'sinde, alt kenar yeniden biraz açılır (yansıyan
// ışık); bevel'de düz zemin ve üstten sönen parlama.
func (t *Tokens) controlGradient(top, bottom ui.Color) ui.LinearGradient {
	if t.Skin == SkinBevel {
		c := t.Colors
		return ui.LinearGradient{Angle: 180, Oklab: true, Stops: []ui.GradientStop{
			{Color: over(c.BevelHiSoft, c.Raised), At: 0},
			{Color: c.Raised, At: 0.7},
			{Color: c.Raised, At: 1},
		}}
	}
	return ui.LinearGradient{Angle: 180, Oklab: true, Stops: []ui.GradientStop{
		{Color: top, At: 0},
		{Color: bottom, At: 0.72},
		{Color: Mix(bottom, top, 0.45), At: 1},
	}}
}

// over, yarı saydam a'yı opak b'nin üstüne koyar.
func over(a, b ui.Color) ui.Color {
	return Mix(b, ui.Color{R: a.R, G: a.G, B: a.B, A: 255}, float32(a.A)/255)
}

// brighten, bir gradyanın bütün duraklarına filter: brightness uygular.
func brighten(g ui.LinearGradient, f float32) ui.LinearGradient {
	if f == 1 {
		return g
	}
	stops := make([]ui.GradientStop, len(g.Stops))
	for i, s := range g.Stops {
		stops[i] = ui.GradientStop{Color: Brightness(s.Color, f), At: s.At}
	}
	g.Stops = stops
	return g
}

// raised, e'yi kabarık kontrol olarak boyar: gradyan zemin, 1px kenar,
// üstte iç parlama ve temas gölgesi. Hover parlatır (koyuda aydınlatır),
// basmak 1px aşağı indirir ve karartır; klavye odağı halkayı gölgenin
// yerine koyar.
func (t *Tokens) raised(e ui.Element, g ui.LinearGradient, border ui.Color, shadow Shadow, s raisedState) {
	f := float32(1)
	switch {
	case s.disabled:
	case s.pressed:
		f = 0.94
		e.Translate(0, 1)
	case s.hovered:
		f = t.ControlHover
	}
	e.LinearGradient(brighten(g, f)).Border(1, Brightness(border, f))
	if s.focus && !s.disabled {
		shadow = t.Shadows.FocusRing
	}
	shadow.Apply(e)
	if s.disabled {
		e.Opacity(0.4)
	}
	e.FocusRing(false).Transition(ui.ElementTransition{Duration: 150 * time.Millisecond, Colors: true, Ease: ui.EaseOut})
}

// well, e'yi gömük alan olarak boyar: opak zemin, ince kenar; odakta yüzey
// rengine açılır, kenar vurgu rengini, çevresi yumuşak halkayı alır.
func (t *Tokens) well(e ui.Element, focus, invalid bool) {
	c := t.Colors
	bg, border, shadow := c.Well, c.WellBorder, t.Shadows.WellShadow
	if focus {
		bg, border, shadow = c.Surface, c.ColorAccent500, t.Shadows.Ring
	}
	if invalid {
		border = c.Up
		shadow = Shadow{{Spread: 4, Color: Mix(c.Up, ui.Color{}, 0.82)}}
	}
	e.Background(bg).Border(1, border)
	shadow.Apply(e)
	e.Transition(ui.ElementTransition{Duration: 200 * time.Millisecond, Colors: true})
}
