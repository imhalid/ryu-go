package mizan

import "github.com/egoist/mygo/ui"

// Mode, temadır: açık, koyu ya da sistemin görünümü. CSS'te data-theme.
type Mode uint8

const (
	// ModeInherit çevreden alır; pencerenin kökünde sistemi izler.
	ModeInherit Mode = iota
	Light
	Dark
)

// Skin, yüzeyin dilidir (data-skin): ışık çizgisi, gölge, gömük alan, topuz.
// Renk teması değildir; açık ve koyu her malzemede çalışır.
type Skin uint8

const (
	// SkinInherit çevreden alır; kökte SkinMizan.
	SkinInherit Skin = iota
	SkinMizan
	// SkinBevel, kabarık kenarlı dil: üstte ışık, altta gölge çizgisi, metal topuz.
	SkinBevel
)

const skinCount = 2

// Size, yoğunluk kademesidir (data-size). Sıfır çevreden alır; kökte Md.
type Size uint8

const (
	SizeInherit Size = iota
	Sm
	Md
	Lg
)

// Options, bir bölgenin teması, malzemesi ve boyutudur. Verilmeyen alan
// çevreden gelir; en yakın kapsam kazanır.
type Options struct {
	Mode Mode
	Skin Skin
	Size Size
}

// Tokens, bir kapsamdaki çözülmüş değerlerdir: renkler, gölgeler ve boyut
// kademesinin ölçüleri. Bileşenler değerleri yalnız buradan okur.
type Tokens struct {
	Dark bool
	Skin Skin
	Size Size

	Colors  Colors
	Shadows Shadows
	Metrics Metrics

	// ControlHover, kabarık kontrolün hover parlaklığıdır: açıkta hafif
	// kararır, koyuda aydınlanır (CSS --control-hover).
	ControlHover float32

	// theme, bu kapsamın MyGo teması: yerleşik widget'lar da Mizan renklerini
	// alır, Context.Theme bu işaretçiyi verdiği için kapsam ondan bulunur.
	theme *ui.Theme
}

// Font aileleri: sistem yazısı (tek aile) ve eş aralıklı.
const (
	FontSans = "SF Pro Text, Helvetica Neue, Inter, system-ui"
	FontMono = "SF Mono, JetBrains Mono, Menlo, Consolas, monospace"
)

// cache[malzeme][koyu][kademe]; registry temanın işaretçisinden kapsamı bulur.
// Görünümler ana iş parçacığında kurulur, kilit gerekmez.
var (
	cache    [skinCount][2][3]*Tokens
	registry = map[*ui.Theme]*Tokens{}
)

func tokensFor(skin Skin, dark bool, size Size) *Tokens {
	si, di, zi := int(skin-1), 0, int(size-1)
	if dark {
		di = 1
	}
	if t := cache[si][di][zi]; t != nil {
		return t
	}
	t := &Tokens{
		Dark:    dark,
		Skin:    skin,
		Size:    size,
		Colors:  genColors[si][di],
		Shadows: genShadows[si][di],
		Metrics: genMetrics[zi],
	}
	t.ControlHover = 0.98
	if dark {
		t.ControlHover = 1.12
	}
	t.theme = t.uiTheme()
	cache[si][di][zi] = t
	registry[t.theme] = t
	return t
}

// uiTheme, MyGo'nun yerleşik widget'larına (kaydırma çubuğu, menü, ipucu,
// odak halkası) Mizan'ın renklerini ve ölçülerini verir.
func (t *Tokens) uiTheme() *ui.Theme {
	c := t.Colors
	th := ui.LightTheme()
	if t.Dark {
		th = ui.DarkTheme()
	}
	th.Background = c.Bg
	th.Surface = c.Well
	th.SurfaceHover = Mix(c.Well, c.Text, 0.06)
	th.SurfacePressed = Mix(c.Well, c.Text, 0.1)
	th.Border = c.ControlBorder
	th.Text = c.Text
	th.TextMuted = c.Muted
	th.Accent = c.Accent
	th.AccentHover = c.ColorAccent500
	th.AccentPressed = c.ColorAccent700
	th.AccentText = c.OnAccent
	th.Danger = c.Up
	th.Warning = c.Warn
	th.Success = c.Down
	th.Selection = c.ColorAccent500.Alpha(0.3)
	th.Focus = c.ColorAccent500.Alpha(0.6)
	th.Radius = RadiusControl
	th.FontSize = FsBase
	th.Font = FontSans
	return th
}

// From, c'nin kapsamındaki token'ları verir: en yakın Scope'unki, yoksa
// Root'unki. Root çağrılmamışsa sistemin görünümünde varsayılanlar.
func From(c *ui.Context) *Tokens {
	if t := registry[c.Theme()]; t != nil {
		return t
	}
	return tokensFor(SkinMizan, c.Theme().Dark, Md)
}

// Painting, boyarken (Element.Draw) kapsamın token'larını verir.
func Painting(p *ui.Painter) *Tokens {
	if t := registry[p.Theme()]; t != nil {
		return t
	}
	return tokensFor(SkinMizan, p.Theme().Dark, Md)
}

// Root, pencerenin bütününe Mizan'ı verir; görünümün başında, her karede
// çağrılır. ModeInherit sistemin açık/koyu görünümünü izler.
//
//	func (a *app) view(c *ui.Context) {
//		mizan.Root(c, mizan.Options{})
//		...
//	}
func Root(c *ui.Context, o Options) *Tokens {
	dark := c.Theme().Dark
	if prev := registry[c.Theme()]; prev != nil {
		// aynı karede ikinci çağrı: sistemin görünümü önceki çağrıdadır
		dark = prev.Dark
	}
	switch o.Mode {
	case Light:
		dark = false
	case Dark:
		dark = true
	}
	t := tokensFor(or(o.Skin, SkinMizan), dark, or(o.Size, Md))
	c.SetTheme(t.theme)
	return t
}

// Scope, çocukları başka tema, malzeme ya da boyutla kurulan bir sütundur
// (CSS'te bir kaba data-theme, data-skin, data-size vermek). Verilmeyen alan
// çevreden gelir. Arka plan boyamaz; bölgenin zemini isteniyorsa
// Background(mizan.From(c).Colors.Bg) ile verilir.
//
//	mizan.Scope(c, mizan.Options{Size: mizan.Sm}).Gap(8).Children(func() {
//		mizan.Button(c, "Dışa aktar", mizan.ButtonOptions{})
//	})
func Scope(c *ui.Context, o Options) ui.Element {
	return ui.Themed(c, scoped(c, o).theme).Background(ui.Transparent)
}

func scoped(c *ui.Context, o Options) *Tokens {
	p := From(c)
	dark := p.Dark
	switch o.Mode {
	case Light:
		dark = false
	case Dark:
		dark = true
	}
	return tokensFor(or(o.Skin, p.Skin), dark, or(o.Size, p.Size))
}

func or[T comparable](v, d T) T {
	var zero T
	if v == zero {
		return d
	}
	return v
}

// TypeStyle, tokens.json › type'taki bir yazı stilidir.
type TypeStyle struct {
	Size, LineHeight float32
	Weight           int
	// Tracking, harf aralığıdır, em cinsinden (-0.02 = -0.02em).
	Tracking float32
	Mono     bool
}

// Apply, stili bir metin öğesine verir.
func (s TypeStyle) Apply(e ui.Element) ui.Element {
	e.FontSize(s.Size).FontWeight(s.Weight).LineHeight(s.LineHeight)
	if s.Tracking != 0 {
		e.LetterSpacing(s.Tracking * s.Size)
	}
	if s.Mono {
		e.Font(FontMono)
	}
	return e
}

// Text, metni bir yazı stiliyle yazar.
//
//	mizan.Text(c, "Haftalık aktif kullanıcı", mizan.TTitle)
func Text(c *ui.Context, s string, style TypeStyle) ui.Element {
	return style.Apply(ui.Text(c, s))
}
