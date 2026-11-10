package ryugo

import "github.com/egoist/mygo/ui"

// Mode is the theme: light, dark or the system's appearance. data-theme in CSS.
type Mode uint8

const (
	// ModeInherit inherits it; at the window's root it follows the system.
	ModeInherit Mode = iota
	Light
	Dark
)

// Size is the density step (data-size). Zero inherits it; Md at the root.
type Size uint8

const (
	SizeInherit Size = iota
	Sm
	Md
	Lg
)

// Corners is how round the corners are: the radii of cards, fields, menus
// and tables at half, at, or one and a half times Ryugo's. Pills (buttons,
// badges) and circles keep their shape.
type Corners uint8

const (
	// CornersInherit inherits it; CornersMd at the root.
	CornersInherit Corners = iota
	CornersSm
	CornersMd
	CornersLg
)

// Options are a region's theme, size, corners and accent. Fields left unset are
// inherited; the nearest scope wins.
type Options struct {
	Mode    Mode
	Size    Size
	Corners Corners
	// CornerScale sets the corners as a measure of the design's own, in place
	// of Corners' three steps: 1 is the design, 0.5 half as round, 2 twice. It
	// also takes the pills (buttons, switches, badges) along: below 1 they
	// round less, until square; at 1 and over they stay pills. Zero keeps
	// Corners, and the pills as they are. For square, use a small value, 0.01.
	CornerScale float32
	// Accent is the brand color: the accent ramp, the selection fills and
	// the rings take its hue and chroma, at Ryugo's lightness steps (so
	// contrast holds in both themes). Zero keeps Ryugo's blue. Keep it to a
	// few colors: every one is cached for the life of the program.
	Accent ui.Color
	// Sounds makes presses heard: a Button pressed plays a short drop, the
	// wave it sends through itself made audible, from the side it was
	// pressed on. Off by default; a Scope inside a scope with sounds keeps
	// them.
	//
	//	ryugo.Root(c, ryugo.Options{Sounds: app.prefs.sounds})
	Sounds bool
	// Waves sends a wave through a Button pressed, from where it was pressed
	// (ripple.go): the press seen as motion. Off by default, as Sounds; a
	// Scope inside a scope with waves keeps them, and less motion in the
	// system's settings stops them whatever this is.
	//
	//	ryugo.Root(c, ryugo.Options{Waves: app.prefs.waves})
	Waves bool
}

// Tokens are the resolved values of a scope: colors, shadows and the size
// step's dimensions. Components read values only from here.
type Tokens struct {
	Dark    bool
	Size    Size
	Corners Corners
	// corner is the factor the radii take for Corners (or CornerScale).
	corner float32
	// scale is CornerScale, 0 when unset; the pills follow it below 1.
	scale float32
	// accent is the brand color the scope was made with, zero for Ryugo's.
	accent ui.Color
	// sounds is Options.Sounds: presses are heard.
	sounds bool
	// waves is Options.Waves: presses send a wave through the button.
	waves bool

	Colors  Colors
	Shadows Shadows
	Metrics Metrics

	// ControlHover is the hover brightness of a raised control: slightly
	// darker in light, lighter in dark (CSS --control-hover).
	ControlHover float32

	// theme is this scope's MyGo theme: the built-in widgets get Ryugo's
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
	dark    bool
	size    Size
	corners Corners
	accent  ui.Color
	scale   float32
	sounds  bool
	waves   bool
}

// cache holds the scopes made so far; registry finds the scope from the
// theme's pointer. Views are built on the main thread, so no lock is needed.
var (
	cache    = map[scopeKey]*Tokens{}
	registry = map[*ui.Theme]*Tokens{}
)

func tokensFor(dark bool, size Size, corners Corners, accent ui.Color) *Tokens {
	return tokensScaled(dark, size, corners, accent, 0)
}

func tokensScaled(dark bool, size Size, corners Corners, accent ui.Color, scale float32) *Tokens {
	return tokensOf(scopeKey{dark, size, corners, accent, scale, false, false})
}

func tokensOf(key scopeKey) *Tokens {
	dark, size, corners, accent, scale := key.dark, key.size, key.corners, key.accent, key.scale
	if t := cache[key]; t != nil {
		return t
	}
	di, zi := 0, int(size-1)
	if dark {
		di = 1
	}
	t := &Tokens{
		Dark:    dark,
		Size:    size,
		accent:  accent,
		Colors:  themeColors[di],
		Shadows: themeShadows[di],
		Metrics: sizeMetrics[zi],
		Corners: corners,
		corner:  [...]float32{CornersSm: 0.5, CornersMd: 1, CornersLg: 1.5}[corners],
		scale:   scale,
		sounds:  key.sounds,
		waves:   key.waves,
	}
	if scale > 0 {
		t.corner = scale
	}
	// the fields' own radii go with the rest
	for _, r := range []*float32{&t.Metrics.FieldR, &t.Metrics.ChipR, &t.Metrics.NumR} {
		*r = t.R(*r)
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
// ring) Ryugo's colors and dimensions.
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
	th.Radius = t.R(RadiusControl)
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
	return tokensFor(c.Theme().Dark, Md, CornersMd, ui.Color{})
}

// Painting returns the scope's tokens while painting (Element.Draw).
func Painting(p *ui.Painter) *Tokens {
	if t := registry[p.Theme()]; t != nil {
		return t
	}
	return tokensFor(p.Theme().Dark, Md, CornersMd, ui.Color{})
}

// Root gives Ryugo to the whole window; it is called at the start of the
// view, every frame. ModeInherit follows the system's light/dark appearance.
//
//	func (a *app) view(c *ui.Context) {
//		ryugo.Root(c, ryugo.Options{})
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
	t := tokensOf(scopeKey{dark, or(o.Size, Md), or(o.Corners, CornersMd), o.Accent, o.CornerScale, o.Sounds, o.Waves})
	c.SetTheme(t.theme)
	return t
}

// Scope is a column whose children are built with another theme, size or
// accent (giving a container data-theme, data-size in CSS). Fields
// left unset are inherited. It paints no background; if the region needs
// one, give it with Background(ryugo.From(c).Colors.Bg).
//
//	ryugo.Scope(c, ryugo.Options{Size: ryugo.Sm}).Gap(8).Children(func() {
//		ryugo.Button(c, "Export", ryugo.ButtonOptions{})
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
	return tokensOf(scopeKey{dark, or(o.Size, p.Size), or(o.Corners, p.Corners), or(o.Accent, p.accent), or(o.CornerScale, p.scale), o.Sounds || p.sounds, o.Waves || p.waves})
}

// Concentric is the radius of a surface inset pad DIPs inside one whose
// corners have radius outer, for their corners to share one center: outer
// less pad, square once pad reaches outer. Two radii scaled apart, or the
// same radius inside and out, leave the gap between the corners thicker
// than along the sides.
//
//	track := t.R(ryugo.RadiusControl)
//	pill := ryugo.Concentric(track, 3)
func Concentric(outer, pad float32) float32 { return max(outer-pad, 0) }

// R is radius r at the scope's Corners; pills stay pills.
func (t *Tokens) R(r float32) float32 {
	if r >= RadiusPill {
		return r
	}
	return r * t.corner
}

// Pill is the radius of a pill h high: a full pill, as a button is, until
// CornerScale takes the corners below the design's own, when it narrows with
// them, half the height at 1 and square at 0.
func (t *Tokens) Pill(h float32) float32 {
	if t.scale <= 0 || t.scale >= 1 {
		return RadiusPill
	}
	return h / 2 * t.scale
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
//	ryugo.Text(c, "Weekly active users", ryugo.TTitle)
func Text(c *ui.Context, s string, style TypeStyle) ui.Element {
	return style.Apply(ui.Text(c, s))
}
