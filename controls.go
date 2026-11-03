package mizan

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// Choice controls start inset (a well, an axis border) and fill with the
// accent gradient when selected; knobs (Switch, Slider) are discs sitting
// on the well, white, or metal in bevel.

// accentFill is the fill of a selection mark (CSS --accent-fill).
func (t *Tokens) accentFill() ui.LinearGradient {
	return ui.LinearGradient{Angle: 180, Oklab: true, From: t.Colors.ColorAccent500, To: t.Colors.ColorAccent600}
}

// mark paints the box of Checkbox and Radio: a well when empty; when
// selected, the accent gradient, a sheen on top and a contact shadow tinted
// toward the accent below. It shrinks to 92% while pressed; keyboard focus
// puts the ring in place of the shadow.
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

// pop starts the mark's contents (check, dot) at half size and transparent,
// and springs them into place.
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

// CheckOptions are the options of Checkbox.
type CheckOptions struct {
	// Description is the muted description below the label.
	Description string
	// Indeterminate shows the box half-checked (a dash): some sub-options are selected.
	Indeterminate bool
	Disabled      bool
}

// Checkbox is an 18px square checkbox and its label: inset when empty, the
// accent gradient and a white check when checked. The whole row is
// clickable.
//
//	mizan.Checkbox(c.Key("mail"), &app.mail, "Email notifications", mizan.CheckOptions{})
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
		box.Children(func() { checkMark(c, 12, 1.75, on, o.Indeterminate, false, t.Colors.OnAccent) })
	})
	return e
}

// checkMark draws a check (or a dash) size DIPs wide in a stroke width
// wide, as a pen would: on, the stroke runs from the short arm through the
// long one, easing in and out; off, it draws itself back the way it came,
// quicker, and fades with the box's fill, so that no white stroke is left
// on the empty well. drawIn draws a mark that shows from the start too,
// once its pop-up list has opened (a list closes as an option is chosen,
// so the change itself is seldom seen). It takes its room even off.
func checkMark(c *ui.Context, size, width float32, on, dash, drawIn bool, color ui.Color) ui.Element {
	m := ui.Box(c).Size(size, size).Shrink(0)
	key := "check"
	if dash {
		key = "dash"
	}
	if drawIn {
		if seen := ui.Local(m, "seen", func() bool { return false }); !*seen {
			*seen = true
			m.Jump(key, 0)
		}
	}
	a := ui.Animation{Duration: 260 * time.Millisecond, Ease: ui.CubicBezier(0.65, 0, 0.35, 1)}
	if drawIn {
		a.Delay = 90 * time.Millisecond
	}
	if !on {
		a = ui.Animation{Duration: 120 * time.Millisecond, Ease: ui.EaseOut}
	}
	drawn := m.AnimateTo(key, b2f(on), a)
	if drawn <= 0.001 {
		return m
	}
	if !on {
		m.Opacity(drawn * drawn)
	}
	m.Draw(func(p *ui.Painter, r ui.Rect) {
		// the icon's path, in a 16-unit box scaled to the mark
		at := func(x, y float32) (float32, float32) { return r.X + x*r.W/16, r.Y + y*r.H/16 }
		var path ui.Path
		if dash {
			path.MoveTo(at(4, 8)).LineTo(at(12, 8))
		} else {
			path.MoveTo(at(3.5, 8.5)).LineTo(at(6.5, 11.5)).LineTo(at(12.5, 4.5))
		}
		p.StrokePath(path.Trim(0, drawn), width, color)
	})
	return m
}

// Radio is an 18px round radio button and its label: when selected, the
// accent gradient with a 7px white dot in its middle. Use it inside
// ui.RadioGroup.
//
//	ui.RadioGroup(c, func() {
//		mizan.Radio(c.Key("d"), &app.plan, "day", "Daily", mizan.CheckOptions{})
//		mizan.Radio(c.Key("w"), &app.plan, "week", "Weekly", mizan.CheckOptions{})
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

// choiceRow lays out the mark and the label side by side: with a
// description, the mark stays centered on the first line.
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

// Choice is a radio button shaped as a card: it sits in a well and switches
// to the raised skin when selected (standing out with its shadow and
// border). Options sit side by side in a grid, inside ui.RadioGroup.
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

// knob is a knob's surface: a white disc in Mizan, metal that darkens from
// top to bottom in bevel.
func (t *Tokens) knob(e ui.Element) {
	if t.Skin == SkinBevel {
		e.LinearGradient(ui.LinearGradient{Angle: 180, Oklab: true, From: t.Colors.KnobTop, To: t.Colors.KnobBot})
		return
	}
	e.Background(t.Colors.Thumb)
}

// SwitchOptions are the options of Switch.
type SwitchOptions struct {
	Label       string
	Description string
	Disabled    bool
}

// Switch is a 38×22 switch: the track is a dark well that fills with the
// accent gradient when on (in bevel the track stays the same and a 6px LED
// lights up on the side the knob leaves); the knob slides on a spring and
// stretches into a short capsule while held.
//
//	mizan.Switch(c.Key("sync"), &app.sync, mizan.SwitchOptions{Label: "Sync"})
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
		// the knob's start and width move on one spring, so that it goes
		// straight the new way: pressed, it stretches toward the middle
		// of the track, and an "on" knob's end stays put as it does
		spring := ui.SpringOptions{Duration: 300 * time.Millisecond, Bounce: 0.25}
		stretch := 4 * b2f(pressed)
		w := track.Spring("w", 18+stretch, spring)
		x := track.Spring("x", 16*b2f(*on)-stretch*b2f(*on), spring)
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

// SegmentedOptions are the options of Segmented.
type SegmentedOptions struct {
	// Label is the group's accessible name.
	Label string
	Size  Size
	// Fill spreads the track across the row; the options share it equally.
	Fill     bool
	Disabled bool
}

// Segmented is a choice between a few views: the track is inset and as
// tall as the field beside it (control-h); the selected option is a raised
// pill inside the track. Unselected options read as separate buttons,
// divided by grooved separators; the separators on both sides of the pill
// fade out.
//
//	mizan.Segmented(c.Key("view"), &app.view, []string{"List", "Grid"}, mizan.SegmentedOptions{Label: "View"})
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
		// the track is as wide as its contents: it doesn't spread across the column
		tr.AlignSelf(ui.Start)
	}
	if o.Disabled {
		tr.Opacity(0.4)
	}
	// The pill under the chosen segment slides to a new choice on a spring
	// (straight there with the keys, or the first time it shows). It is
	// painted by the track, under the segments: placed from where the last
	// frame laid them out. Until then, the chosen segment paints its own.
	segs := make([]ui.Element, len(labels))
	var seps []ui.Element
	keys := false
	tr.Children(func() {
		for i, l := range labels {
			if i > 0 {
				sep := ui.Box(c).Width(1).Margin(m.ControlH*0.25-3, 0).Background(cl.SepDark).Shadow(1, 0, 0, 0, cl.SepLight).Shrink(0)
				if *selected == i || *selected == i-1 {
					sep.Opacity(0)
				}
				sep.Transition(ui.ElementTransition{Duration: 150 * time.Millisecond, Colors: true})
				seps = append(seps, sep)
			}
			s := parts.Segment(i).Center().Padding(0, m.SegPx).Radius(RadiusInner).Cursor(ui.CursorPointer).FocusRing(false).Border(0.5, ui.Color{})
			segs[i] = s
			if o.Fill {
				s.Grow(1).Basis(0)
			}
			text := cl.Muted
			if *selected == i || s.Hovered() && !o.Disabled {
				text = cl.Text
			}
			if s.FocusVisible() {
				keys = true
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
	sel := *selected
	if sel < 0 || sel >= len(segs) {
		return tr
	}
	sb, tb := segs[sel].LayoutBounds(), tr.LayoutBounds()
	if sb.W == 0 || tb.W == 0 {
		segs[sel].LinearGradient(t.controlGradient(cl.ControlTop, cl.ControlBottom)).Border(0.5, cl.ControlBorder)
		t.Shadows.ShadowRaised.Apply(segs[sel])
		return tr
	}
	placed := ui.Local(tr, "placed", func() bool { return false })
	x, w := sb.X-tb.X, sb.W
	if keys || !*placed {
		x, w = tr.Jump("pill-x", x), tr.Jump("pill-w", w)
		*placed = true
	} else {
		x, w = tr.AnimateTo("pill-x", x, springPill), tr.AnimateTo("pill-w", w, springPill)
	}
	pill := ui.Rect{X: x, Y: sb.Y - tb.Y, W: w, H: sb.H}
	// the separators the pill passes over go out as it does
	for _, sep := range seps {
		if b := sep.LayoutBounds(); b.W > 0 && b.X-tb.X > pill.X-2 && b.X-tb.X < pill.X+pill.W+2 {
			sep.Opacity(0)
		}
	}
	tr.Draw(func(p *ui.Painter, r ui.Rect) {
		t.paintRaised(p, ui.Rect{X: r.X + pill.X, Y: r.Y + pill.Y, W: pill.W, H: pill.H}, RadiusInner, t.Shadows.ShadowRaised)
	})
	return tr
}

// springPill is how Segmented's pill slides: quick, with a hint of
// overshoot.
var springPill = ui.Animation{Spring: ui.SpringOptions{Duration: 280 * time.Millisecond, Bounce: 0.14}}

// paintRaised paints the raised material with a painter, as raised() styles
// an element: its shadows, the control gradient, the inset highlight and
// shade lines, and a 0.5px border.
func (t *Tokens) paintRaised(p *ui.Painter, r ui.Rect, radius float32, sh Shadow) {
	cl := t.Colors
	for _, l := range sh {
		if !l.Inset && l.Color.A > 0 {
			p.Shadow(r, radius, l.X, l.Y, l.Blur, l.Spread, l.Color)
		}
	}
	p.FillGradient(r, t.controlGradient(cl.ControlTop, cl.ControlBottom), radius)
	p.Clip(r, radius, func() {
		for _, l := range sh {
			if !l.Inset || l.Color.A == 0 {
				continue
			}
			// an inset line of 1px at the top (Y > 0) or the bottom (Y < 0)
			y := r.Y
			if l.Y < 0 {
				y = r.Y + r.H - 1
			}
			p.Fill(ui.Rect{X: r.X, Y: y, W: r.W, H: 1}, l.Color, 0)
		}
	})
	p.Stroke(r, cl.ControlBorder, radius, 0.5)
}

// SliderOptions are the options of Slider.
type SliderOptions struct {
	Label string
	// Step is the value's step (one hundredth of the range if unset).
	Step float64
	// Format, if set, writes the value on the right, at a fixed width, in tabular figures.
	Format   func(v float64) string
	Disabled bool
}

// Slider is a slider: a 6px inset track whose filled part is the accent
// gradient; the 22px knob grows 12% while pressed. With Format, its value
// stands on its right.
//
//	mizan.Slider(c.Key("vol"), &app.volume, 0, 100, mizan.SliderOptions{Label: "Volume"})
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
		// the column fits the widest possible text, so that the track doesn't
		// narrow and the knob slip from under the pointer when the value changes
		// its number of digits
		w := max(56, 8*float32(max(len(o.Format(lo)), len(o.Format(hi)))))
		ui.Text(c, o.Format(*value)).TextColor(cl.Text).FontSize(FsMd).FontWeight(650).FontFeatures("tnum").
			TextAlign(ui.End).Width(w).Shrink(0)
	})
	return s
}
