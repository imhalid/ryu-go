package mizan

import (
	"strconv"
	"time"

	"github.com/egoist/mygo/ui"
)

// Feedback shows how far work has gone or how full something is: bars in
// an inset well, filled with the accent gradient (Progress) or a tone
// (Meter); a ring turning while the length is unknown (Spinner); and
// placeholders shimmering where content is still loading (Skeleton).

// The bar fills on a spring that does not overshoot: a bar past its value
// would read as more than there is.
var springBar = ui.Animation{Spring: ui.SpringOptions{Duration: 500 * time.Millisecond}}

// ProgressOptions are the options of Progress.
type ProgressOptions struct {
	// Label names the bar, above it on the start side and to assistive
	// technology.
	Label string
	// ShowValue writes the value as a percentage above the bar's end.
	ShowValue bool
	// Color fills the bar in its place of the accent gradient.
	Color ui.Color
	// Size sets the bar's thickness: 4, 6 or 8 DIPs.
	Size Size
}

// Progress is a progress bar filled to value between 0 and 1: an inset
// well, as Slider's track, filling with the accent gradient on a spring as
// value grows. A negative value is work of unknown length: a segment of
// the fill sweeps across the well, again and again. As value reaches 1, a
// light passes once over the full bar. Assistive technology sees a
// progress bar.
//
//	mizan.Progress(c, app.uploaded, mizan.ProgressOptions{Label: "Uploading", ShowValue: true})
func Progress(c *ui.Context, value float64, o ProgressOptions) ui.Element {
	t := sizedFrom(c, o.Size)
	fill := t.accentFill()
	if o.Color != (ui.Color{}) {
		fill = ui.LinearGradient{Angle: 180, Oklab: true, From: o.Color, To: lighten(o.Color, -0.06)}
	}
	value = min(value, 1)
	text := ""
	if o.ShowValue && value >= 0 {
		text = strconv.Itoa(int(value*100+0.5)) + "%"
	}
	return t.bar(c, o.Label, text, float32(value), fill, true, func(e ui.Element) {
		e.Role(ui.RoleProgress)
		if value >= 0 {
			e.Range(0, 1, value)
		}
	})
}

// MeterOptions are the options of Meter.
type MeterOptions struct {
	Label string
	// Format writes the value above the bar's end, as "42 GB"; nil writes
	// nothing.
	Format func(v float64) string
	// Warning and Critical are the levels from which the bar turns the
	// Warn and the Up color; zero is no such level. Below them it is the
	// Down color, which reads as fine.
	Warning, Critical float64
	Size              Size
}

// Meter shows how full something is, value between lo and hi, as a disk
// or a quota: Progress's well, filled in a tone that turns from Down to
// Warn to Up as value passes Warning and Critical. Assistive technology
// sees a level indicator.
//
//	mizan.Meter(c, app.used, 0, app.quota, mizan.MeterOptions{Label: "Storage", Warning: 0.8 * app.quota, Critical: 0.95 * app.quota})
func Meter(c *ui.Context, value, lo, hi float64, o MeterOptions) ui.Element {
	t := sizedFrom(c, o.Size)
	cl := t.Colors
	tone := cl.Down
	switch {
	case o.Critical != 0 && value >= o.Critical:
		tone = cl.Up
	case o.Warning != 0 && value >= o.Warning:
		tone = cl.Warn
	}
	frac := float32(0)
	if hi > lo {
		frac = clamp01(float32((value - lo) / (hi - lo)))
	}
	text := ""
	if o.Format != nil {
		text = o.Format(value)
	}
	fill := ui.LinearGradient{Angle: 180, Oklab: true, From: tone, To: lighten(tone, -0.06)}
	return t.bar(c, o.Label, text, frac, fill, false, func(e ui.Element) {
		e.Role(ui.RoleMeter).Range(lo, hi, value)
	})
}

// bar builds Progress and Meter: the label and the value over the well,
// and the well filled to frac (negative: sweeping), which describe names
// to assistive technology. With sheen, a light passes once over the bar
// as it fills to the end. It returns the well.
func (t *Tokens) bar(c *ui.Context, label, value string, frac float32, fill ui.LinearGradient, sheen bool, describe func(ui.Element)) ui.Element {
	cl := t.Colors
	h := map[Size]float32{Sm: 4, Md: 6, Lg: 8}[t.Size]
	if h == 0 {
		h = 6
	}
	var well ui.Element
	ui.Column(c).FillWidth().MinWidth(80).Gap(8).Children(func() {
		if label != "" || value != "" {
			ui.Row(c).AlignItems(ui.End).Gap(12).Children(func() {
				ui.Text(c, label).Grow(1).MinWidth(0).TextColor(cl.Text2).FontSize(FsSm).FontWeight(500).NoWrap()
				if value != "" {
					ui.Text(c, value).TextColor(cl.Muted).FontSize(FsSm).FontFeatures("tnum").NoWrap()
				}
			})
		}
		well = ui.Box(c).FillWidth().Height(h).Radius(RadiusPill).Background(cl.Well).Clip()
		if label != "" {
			well.Label(label)
		}
		describe(well)
		t.Shadows.WellEdge.Apply(well)
		well.Children(func() {
			f := ui.Box(c).Absolute().Top(0).Bottom(0).Radius(RadiusPill).LinearGradient(fill)
			if frac < 0 {
				// a third of the well, from past its start to past its end
				p := f.Keyframes("sweep", []ui.Keyframe{{At: 0, Value: 0}, {At: 1400 * time.Millisecond, Value: 1, Ease: ui.EaseInOut}},
					ui.KeyframeOptions{Repeat: ui.RepeatForever})
				f.LeftPercent(-35 + 135*p).WidthPercent(35)
				return
			}
			v := f.AnimateTo("v", frac, springBar)
			f.Left(0).WidthPercent(100 * clamp01(v))
			if v > 0.001 {
				// a sliver still shows as a pill, not a hairline
				f.MinWidth(h)
			}
			if sheen {
				t.finish(c, well, frac >= 1)
			}
		})
	})
	return well
}

// finish passes a light once over a bar that just filled to the end, as
// the fill settles: a soft band of white crossing it from start to end.
// A bar full from the start, or full again before the light is gone,
// gets no new one.
func (t *Tokens) finish(c *ui.Context, well ui.Element, full bool) {
	done := ui.Local(well, "done", func() bool { return full })
	if full && !*done {
		well.Jump("sheen", 0)
	}
	*done = full
	p := well.AnimateTo("sheen", 1, ui.Animation{Duration: 1100 * time.Millisecond, Delay: 380 * time.Millisecond, Ease: ui.EaseInOut})
	if p >= 1 {
		return
	}
	light := t.Colors.OnAccent.Alpha(0.55)
	ui.Box(c).Absolute().Top(0).Bottom(0).WidthPercent(30).LeftPercent(-30 + 130*p).PassThrough().
		LinearGradient(ui.LinearGradient{Angle: 90, Oklab: true, Stops: []ui.GradientStop{
			{Color: light.Alpha(0), At: 0}, {Color: light, At: 0.5}, {Color: light.Alpha(0), At: 1},
		}})
}

// SpinnerOptions are the options of Spinner.
type SpinnerOptions struct {
	// Size is the ring's diameter: 16 DIPs by default.
	Size float32
	// Color is the turning arc's; the accent by default.
	Color ui.Color
	// Label names it to assistive technology: "Loading" by default.
	Label string
}

// Spinner is a ring for work of unknown length: a faint track, and over it
// an arc of a quarter turn in the accent color that turns once every 0.8
// seconds. Assistive technology sees a progress indicator.
//
//	mizan.Spinner(c, mizan.SpinnerOptions{})
func Spinner(c *ui.Context, o SpinnerOptions) ui.Element {
	t := From(c)
	cl := t.Colors
	d := or(o.Size, 16)
	color := or(o.Color, cl.ColorAccent500)
	e := ui.Box(c).Size(d, d).Shrink(0).Role(ui.RoleProgress).Label(or(o.Label, "Loading"))
	turn := e.Keyframes("turn", []ui.Keyframe{{At: 0, Value: 0}, {At: 800 * time.Millisecond, Value: 360}},
		ui.KeyframeOptions{Repeat: ui.RepeatForever})
	w := max(1.5, d/9)
	return e.Rotate(turn).Draw(func(p *ui.Painter, r ui.Rect) {
		cx, cy, rad := r.X+r.W/2, r.Y+r.H/2, r.W/2-w/2
		var track, arc ui.Path
		track.Circle(cx, cy, rad)
		p.StrokePath(&track, w, cl.WellStrong)
		arc.Circle(cx, cy, rad)
		p.StrokePath(arc.Trim(0, 0.28), w, color)
	})
}

// Skeleton is a placeholder w by h DIPs (w zero: as wide as its row) where
// content is still loading: a fill of the strong well's tone, rounded as a
// field, with a sheen sweeping across it every 1.8 seconds. Round it
// otherwise with Radius; the sheen keeps to its corners.
//
//	mizan.Skeleton(c, 40, 40).Radius(mizan.RadiusPill) // an avatar's
func Skeleton(c *ui.Context, w, h float32) ui.Element {
	t := From(c)
	cl := t.Colors
	e := ui.Box(c).Height(h).Radius(t.R(RadiusTrack)).Background(cl.WellStrong).Clip().Shrink(0)
	if w > 0 {
		e.Width(w)
	} else {
		e.FillWidth()
	}
	sheen := cl.Surface.Alpha(0.5)
	return e.Children(func() {
		s := ui.Box(c).Absolute().Top(0).Bottom(0).WidthPercent(60).LinearGradient(ui.LinearGradient{
			Angle: 90, Oklab: true, Stops: []ui.GradientStop{
				{Color: sheen.Alpha(0), At: 0}, {Color: sheen, At: 0.5}, {Color: sheen.Alpha(0), At: 1},
			}})
		p := s.Keyframes("sheen", []ui.Keyframe{{At: 0, Value: 0}, {At: 1600 * time.Millisecond, Value: 1, Ease: ui.EaseInOut}},
			ui.KeyframeOptions{Repeat: ui.RepeatForever, RepeatDelay: 200 * time.Millisecond})
		s.LeftPercent(-60 + 160*p)
	})
}

// SkeletonText is a placeholder for a paragraph of lines lines: bars as
// tall as a line of text, the last one shorter.
func SkeletonText(c *ui.Context, lines int) ui.Element {
	return ui.Column(c).FillWidth().Gap(9).Children(func() {
		for i := range lines {
			s := Skeleton(c.Key(i), 0, 10)
			if i == lines-1 && lines > 1 {
				s.WidthPercent(60)
			}
		}
	})
}
