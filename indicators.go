package ryugo

import (
	"time"
	"unicode"

	"github.com/egoist/mygo/ui"
)

// Tone is the tone of small status-colored items (Badge, StatusDot). Color
// alone carries no meaning: text or a name always stands beside it.
type Tone uint8

const (
	// ToneDefault is the text color.
	ToneDefault Tone = iota
	ToneUp
	ToneDown
	ToneWarn
	ToneMuted
	// ToneNeutral is a gray label: its background comes from the fill, not the muted text.
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

// Badge is a pill on a background of 10% of its text color: whatever the
// tone, the background matches the text. Numbers are tabular.
//
//	ryugo.Badge(c, "+4.2%", ryugo.ToneDown)
func Badge(c *ui.Context, text string, tone Tone) ui.Element {
	t := From(c)
	color := t.tone(tone)
	bg := Mix(ui.Color{}, color, 0.1)
	if tone == ToneNeutral {
		bg = t.Colors.Fill
	}
	return ui.Box(c).Padding(2, 8).Radius(t.Pill(20)).Background(bg).Shrink(0).Children(func() {
		ui.Text(c, text).TextColor(color).FontSize(FsXs).FontWeight(600).NoWrap().FontFeatures("tnum")
	})
}

// Tag is the large neutral pill next to a page title.
func Tag(c *ui.Context, text string) ui.Element {
	t := From(c)
	return ui.Box(c).Padding(6, 12).Radius(t.Pill(32)).Background(t.Colors.Fill).Shrink(0).Children(func() {
		ui.Text(c, text).TextColor(t.Colors.Muted).FontSize(FsSm).FontWeight(500).NoWrap()
	})
}

// DotState is the state of StatusDot.
type DotState uint8

const (
	DotOff DotState = iota
	// DotLive is a green dot with a ring that spreads outward and fades.
	DotLive
	DotDelayed
)

// StatusDot is a status dot and its text: when live, the dot spreads
// outward and fades every 1.8 seconds (it stops when motion is reduced).
//
//	ryugo.StatusDot(c, ryugo.DotLive, "Live")
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
			// the ring: the dot's own shadow, fading as its radius opens to 2.6 times
			p := d.Keyframes("pulse", []ui.Keyframe{{At: 0, Value: 0}, {At: 1800 * time.Millisecond, Value: 1, Ease: ui.EaseOut}},
				ui.KeyframeOptions{Repeat: ui.RepeatForever})
			d.Shadow(0, 0, 0, 6.4*p, color.Alpha(0.6*(1-p)))
		}
		ui.Text(c, text).TextColor(t.Colors.Text2).FontSize(FsSm).NoWrap()
	})
}

// Kbd shows keyboard keys: the same raised skin as a button, scaled to the
// text. In a combined shortcut, the keys sit side by side.
//
//	ryugo.Kbd(c, "⌘", "K")
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

// AvatarOptions are the options of Avatar.
type AvatarOptions struct {
	// Image is the picture shown instead of the initials.
	Image *ui.Bitmap
	// Color, if unset, is one of the series colors (C1…C8) picked from the
	// name: the same name gets the same color everywhere.
	Color ui.Color
	// Solid makes the background the color itself and the initials white.
	Solid bool
	// Online shows a green status dot at the bottom right.
	Online bool
	// Size is the diameter (default 32).
	Size float32
}

// Avatar is a person's picture or initials: the background is 16% of the
// color over the card, the letters 62% of the color over the text (legible
// in both themes).
//
//	ryugo.Avatar(c, "Jane Smith", ryugo.AvatarOptions{Online: true})
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

// series picks a series color from the FNV-1a hash of s.
func (t *Tokens) series(s string) ui.Color {
	h := uint32(2166136261)
	for i := 0; i < len(s); i++ {
		h = (h ^ uint32(s[i])) * 16777619
	}
	c := t.Colors
	return [...]ui.Color{c.C1, c.C2, c.C3, c.C4, c.C5, c.C6, c.C7, c.C8}[h%8]
}

// initials returns the initials of the name's first and last words.
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

// EmptyStateOptions are the options of EmptyState.
type EmptyStateOptions struct {
	Icon        *ui.SVG
	Title       string
	Description string
	// Action builds the single action at the bottom (usually a raised Button).
	Action func()
}

// EmptyState is an empty, error or first-use state: a centered column with
// no color accent (the icon too sits in a neutral well); the state is told
// in text.
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
