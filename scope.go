package mizan

import "github.com/egoist/mygo/ui"

// Mode is the theme: light, dark or the system's appearance. data-theme in CSS.
type Mode uint8

const (
	// ModeInherit inherits it; at the window's root it follows the system.
	ModeInherit Mode = iota
	Light
	Dark
)

// Skin is the surface's language (data-skin): light line, shadow, inset
// well, knob. It is not a color theme; light and dark work in every skin.
type Skin uint8

const (
	// SkinInherit inherits it; SkinMizan at the root.
	SkinInherit Skin = iota
	SkinMizan
	// SkinBevel is the language of raised edges: a light line at the top, a shadow line at the bottom, a metal knob.
	SkinBevel
)

const skinCount = 2

// Size is the density step (data-size). Zero inherits it; Md at the root.
type Size uint8

const (
	SizeInherit Size = iota
	Sm
	Md
	Lg
)

// Options are a region's theme, skin and size. Fields left unset are
// inherited; the nearest scope wins.
type Options struct {
	Mode Mode
	Skin Skin
	Size Size
	// Accent is the brand color: the accent ramp, the selection fills and
	// the rings take its hue and chroma, at Mizan's lightness steps (so
	// contrast holds in both themes). Zero keeps Mizan's blue. Keep it to a
	// few colors: every one is cached for the life of the program.
	Accent ui.Color
}

// Tokens are the resolved values of a scope: colors, shadows and the size
// step's dimensions. Components read values only from here.
type Tokens struct {
	Dark bool
	Skin Skin
	Size Size
	// accent is the brand color the scope was made with, zero for Mizan's.
	accent ui.Color

	Colors  Colors
	Shadows Shadows
	Metrics Metrics

	// ControlHover is the hover brightness of a raised control: slightly
	// darker in light, lighter in dark (CSS --control-hover).
	ControlHover float32

	// theme is this scope's MyGo theme: the built-in widgets get Mizan's
	// colors too, and since Context.Theme returns this pointer, the scope is
	// found from it.
	theme *ui.Theme
}

// Font families: the system font (a single family) and monospace.
const (
	FontSans = "SF Pro Text, Helvetica Neue, Inter, system-ui"
	FontMono = "SF Mono, JetBrains Mono, Menlo, Consolas, monospace"
)

// scopeKey tells apart the cached scopes.
type scopeKey struct {
	skin   Skin
	dark   bool
	size   Size
	accent ui.Color
}

// cache holds the scopes made so far; registry finds the scope from the
// theme's pointer. Views are built on the main thread, so no lock is needed.
var (
	cache    = map[scopeKey]*Tokens{}
	registry = map[*ui.Theme]*Tokens{}
)

func tokensFor(skin Skin, dark bool, size Size, accent ui.Color) *Tokens {
	key := scopeKey{skin, dark, size, accent}
	if t := cache[key]; t != nil {
		return t
	}
	si, di, zi := int(skin-1), 0, int(size-1)
	if dark {
		di = 1
	}
	t := &Tokens{
		Dark:    dark,
		Skin:    skin,
		Size:    size,
		accent:  accent,
		Colors:  skinColors[si][di],
		Shadows: skinShadows[si][di],
		Metrics: sizeMetrics[zi],
	}
	if accent != (ui.Color{}) {
		t.withAccent(accent)
	}
	t.ControlHover = 0.98
	if dark {
		t.ControlHover = 1.12
	}
	t.theme = t.uiTheme()
	cache[key] = t
	registry[t.theme] = t
	return t
}

// uiTheme gives MyGo's built-in widgets (scroll bar, menu, tooltip, focus
// ring) Mizan's colors and dimensions.
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

// From returns the tokens of c's scope: the nearest Scope's, otherwise
// Root's. If Root was not called, the defaults in the system's appearance.
func From(c *ui.Context) *Tokens {
	if t := registry[c.Theme()]; t != nil {
		return t
	}
	return tokensFor(SkinMizan, c.Theme().Dark, Md, ui.Color{})
}

// Painting returns the scope's tokens while painting (Element.Draw).
func Painting(p *ui.Painter) *Tokens {
	if t := registry[p.Theme()]; t != nil {
		return t
	}
	return tokensFor(SkinMizan, p.Theme().Dark, Md, ui.Color{})
}

// Root gives Mizan to the whole window; it is called at the start of the
// view, every frame. ModeInherit follows the system's light/dark appearance.
//
//	func (a *app) view(c *ui.Context) {
//		mizan.Root(c, mizan.Options{})
//		...
//	}
func Root(c *ui.Context, o Options) *Tokens {
	dark := c.Theme().Dark
	if prev := registry[c.Theme()]; prev != nil {
		// second call in the same frame: the system's appearance is in the previous call
		dark = prev.Dark
	}
	switch o.Mode {
	case Light:
		dark = false
	case Dark:
		dark = true
	}
	t := tokensFor(or(o.Skin, SkinMizan), dark, or(o.Size, Md), o.Accent)
	c.SetTheme(t.theme)
	return t
}

// Scope is a column whose children are built with another theme, skin or
// size (giving a container data-theme, data-skin, data-size in CSS). Fields
// left unset are inherited. It paints no background; if the region needs
// one, give it with Background(mizan.From(c).Colors.Bg).
//
//	mizan.Scope(c, mizan.Options{Size: mizan.Sm}).Gap(8).Children(func() {
//		mizan.Button(c, "Export", mizan.ButtonOptions{})
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
	return tokensFor(or(o.Skin, p.Skin), dark, or(o.Size, p.Size), or(o.Accent, p.accent))
}

func or[T comparable](v, d T) T {
	var zero T
	if v == zero {
		return d
	}
	return v
}

// TypeStyle is a text style of the type scale in tokens.go.
type TypeStyle struct {
	Size, LineHeight float32
	Weight           int
	// Tracking is the letter spacing, in em (-0.02 = -0.02em).
	Tracking float32
	Mono     bool
}

// Apply gives the style to a text element.
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

// Text writes s in a type style.
//
//	mizan.Text(c, "Weekly active users", mizan.TTitle)
func Text(c *ui.Context, s string, style TypeStyle) ui.Element {
	return style.Apply(ui.Text(c, s))
}
