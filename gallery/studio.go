package gallery

import (
	"fmt"
	"math"

	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/ryu-go"
)

// studioPage is the state of the studio page: a picture of layers on a
// canvas, palettes floating over it (tools, layers, swatches, and the
// brightness and contrast HUD) as in Pixelmator Classic.
type studioPage struct {
	Board    mz.Board
	Layers   []studioLayer
	Sel      int // the layer chosen
	Tool     int
	Bright   float64 // -100 to 100
	Contrast float64
	// what brightness and contrast were as the HUD opened: Cancel goes back
	keep [2]float64
}

// studioLayer is a layer of the picture: a shape in a color.
type studioLayer struct {
	Name    string
	Shape   int // 0 the background, 1 a sun, 2 hills, 3 a frame
	Color   ui.Color
	Visible bool
	Opacity float64 // 0 to 100
}

// The tools' icons.
var (
	iconMove    = mz.Svg(16, 14, "M8 1.5v13M1.5 8h13M8 1.5L6 3.5M8 1.5l2 2M8 14.5l-2-2M8 14.5l2-2M1.5 8l2-2M1.5 8l2 2M14.5 8l-2-2M14.5 8l-2 2")
	iconMarquee = mz.Svg(16, 14, "M2.5 4.5v-2h2M7 2.5h2M11.5 2.5h2v2M13.5 7v2M13.5 11.5v2h-2M9 13.5H7M4.5 13.5h-2v-2M2.5 9V7")
	iconLasso   = mz.Svg(16, 14, "M8 3c3.6 0 6 1.6 6 3.6S11.6 10 8 10 2 8.6 2 6.6 4.4 3 8 3zM5 9.5c-.8 1.2-.6 2.6.6 3.3 1 .6 2.4.2 2.6-1")
	iconBrush   = mz.Svg(16, 14, "M13.5 2.5L7 9M7 9c-1.5-.5-3 .5-3 2 0 1-.7 1.7-1.5 2 2.5.5 5-.5 5-3")
	iconEraser  = mz.Svg(16, 14, "M9 3l4 4-6 6H4l-2-2zM6 6l4 4M7 13h7")
	iconFill    = mz.Svg(16, 14, "M3 8l5-5 5 5-5 5zM3 8h10M13.5 10.5c.6.9 1 1.6 1 2a1 1 0 0 1-2 0c0-.4.4-1.1 1-2z")
	iconType    = mz.Svg(16, 14, "M3 3.5h10M8 3.5v9M6 12.5h4")
	iconShape   = mz.Svg(16, 14, "M2.5 6.5h7v7h-7zM8.5 2.5a3 3 0 1 1 3 6")
)

// newStudio is the picture the studio starts with.
func newStudio() []studioLayer {
	return []studioLayer{
		{Name: "Background", Shape: 0, Color: ui.RGB(82, 139, 214), Visible: true, Opacity: 100},
		{Name: "Sun", Shape: 1, Color: ui.RGB(255, 176, 59), Visible: true, Opacity: 100},
		{Name: "Hills", Shape: 2, Color: ui.RGB(46, 125, 84), Visible: true, Opacity: 100},
		{Name: "Frame", Shape: 3, Color: ui.RGB(250, 246, 238), Visible: true, Opacity: 100},
	}
}

func (s *studioPage) view(c *ui.Context, t *mz.Tokens) {
	if s.Layers == nil {
		s.Layers = newStudio()
		s.Sel = 1
	}
	mz.Card(c, mz.CardOptions{Title: "Studio", Actions: func() {
		mz.Badge(c, "Experimental", mz.ToneWarn)
	}, Subtitle: "Experimental: palettes floating over a canvas, as Pixelmator Classic's. Drag a palette by its title; it settles onto the board's edges and the palettes near it. A double click on a title folds it up; its close button closes it.", Pad: true}, func() {
		ui.Column(c).Gap(12).Children(func() {
			s.toolbar(c)
			ui.Box(c).FillWidth().Height(620).Radius(t.R(mz.RadiusControl)).Clip().Children(func() {
				mz.Palettes(c.Key("board"), &s.Board, func(p *mz.PaletteSet) {
					s.canvas(c, t)
					p.Palette("tools", mz.PaletteOptions{Title: "Tools", Width: 74, At: ui.Point{X: 12, Y: 12}}, s.tools)
					p.Palette("layers", mz.PaletteOptions{Title: "Layers", Width: 236, At: ui.Point{X: -12, Y: 12}}, s.layers)
					p.Palette("swatches", mz.PaletteOptions{Title: "Swatches", Width: 196, At: ui.Point{X: 12, Y: -12}}, s.swatches)
					p.Palette("adjust", mz.PaletteOptions{Title: "Brightness and Contrast", Width: 300, At: ui.Point{X: 120, Y: 60}}, s.adjust)
				}).Fill()
			})
		})
	})
}

// toolbar opens and closes the palettes, and the HUD.
func (s *studioPage) toolbar(c *ui.Context) {
	if s.Board.Palettes == nil {
		// the HUD shows when asked for
		s.Board.Close("adjust")
	}
	ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
		mz.ToggleGroup(c, "Palettes", func() {
			for _, p := range []struct{ id, name string }{{"tools", "Tools"}, {"layers", "Layers"}, {"swatches", "Swatches"}} {
				on := s.Board.IsOpen(p.id)
				was := on
				mz.Toggle(c.Key(p.id), &on, p.name, mz.ToggleOptions{Size: mz.Sm})
				if on != was {
					if on {
						s.Board.Open(p.id)
					} else {
						s.Board.Close(p.id)
					}
				}
			}
		})
		if mz.Button(c.Key("adjust"), "Brightness and contrast…", mz.ButtonOptions{Size: mz.Sm}).Clicked() {
			s.keep = [2]float64{s.Bright, s.Contrast}
			s.Board.Open("adjust")
		}
		if mz.Button(c.Key("reset"), "Reset palettes", mz.ButtonOptions{Size: mz.Sm}).Clicked() {
			s.Board.Reset()
		}
	})
}

// adjusted is c with the brightness and contrast set.
func (s *studioPage) adjusted(col ui.Color) ui.Color {
	k := 1 + s.Contrast/100
	b := s.Bright / 200
	ch := func(v uint8) uint8 {
		f := (float64(v)/255-0.5)*k + 0.5 + b
		return uint8(math.Round(min(max(f, 0), 1) * 255))
	}
	return ui.Color{R: ch(col.R), G: ch(col.G), B: ch(col.B), A: col.A}
}

// canvas is the board's desk and the picture on it, drawn from its layers.
func (s *studioPage) canvas(c *ui.Context, t *mz.Tokens) {
	desk := ui.Box(c).Fill().Center().LinearGradient(ui.LinearGradient{Angle: 180, Oklab: true,
		From: ui.RGB(46, 84, 140), To: ui.RGB(24, 42, 78)})
	desk.Children(func() {
		pic := ui.Box(c).Size(460, 320).Shadow(0, 18, 48, 0, ui.RGBA(0, 0, 0, 0.45)).Shadow(0, 0, 0, 1, ui.RGBA(0, 0, 0, 0.3))
		pic.Draw(func(p *ui.Painter, r ui.Rect) {
			p.Clip(r, 0, func() {
				for _, l := range s.Layers {
					if l.Visible {
						s.draw(p, r, l)
					}
				}
			})
		})
	})
}

// draw draws layer l on the picture r.
func (s *studioPage) draw(p *ui.Painter, r ui.Rect, l studioLayer) {
	col := s.adjusted(l.Color)
	a := float32(l.Opacity / 100)
	switch l.Shape {
	case 0:
		p.FillGradient(r, ui.LinearGradient{Angle: 180, Oklab: true, From: col.Alpha(a), To: mz.Mix(col, ui.RGB(255, 255, 255), 0.55).Alpha(a)}, 0)
	case 1:
		cx, cy, rad := r.X+r.W*0.68, r.Y+r.H*0.36, r.H*0.17
		p.FillRadialGradient(ui.Rect{X: cx - rad*2, Y: cy - rad*2, W: rad * 4, H: rad * 4}, ui.RadialGradient{Center: ui.Point{X: 0.5, Y: 0.5}, Radius: rad * 2, Oklab: true,
			Stops: []ui.GradientStop{{Color: col.Alpha(0.5 * a), At: 0}, {Color: col.Alpha(0), At: 1}}}, 0)
		var c ui.Path
		c.Circle(cx, cy, rad)
		p.FillPath(&c, col.Alpha(a))
	case 2:
		var h ui.Path
		h.MoveTo(r.X, r.Y+r.H*0.72).
			CubeTo(r.X+r.W*0.22, r.Y+r.H*0.52, r.X+r.W*0.42, r.Y+r.H*0.62, r.X+r.W*0.58, r.Y+r.H*0.7).
			CubeTo(r.X+r.W*0.76, r.Y+r.H*0.78, r.X+r.W*0.9, r.Y+r.H*0.56, r.X+r.W, r.Y+r.H*0.6).
			LineTo(r.X+r.W, r.Y+r.H).LineTo(r.X, r.Y+r.H).Close()
		p.FillPath(&h, col.Alpha(a))
	case 3:
		p.Stroke(ui.Rect{X: r.X + 6, Y: r.Y + 6, W: r.W - 12, H: r.H - 12}, col.Alpha(a), 0, 12)
	}
}

// tools is the tools palette: the tools, two by two, and the layer's color.
func (s *studioPage) tools(c *ui.Context) {
	icons := []*ui.SVG{iconMove, iconMarquee, iconLasso, iconBrush, iconEraser, iconFill, iconType, iconShape}
	names := []string{"Move", "Select", "Lasso", "Brush", "Eraser", "Fill", "Type", "Shape"}
	ui.Column(c).Gap(4).AlignItems(ui.Center).Children(func() {
		for row := 0; row < len(icons); row += 2 {
			ui.Row(c).Gap(4).Children(func() {
				for i := row; i < row+2; i++ {
					on := s.Tool == i
					mz.Toggle(c.Key(names[i]), &on, names[i], mz.ToggleOptions{Icon: icons[i], IconOnly: true, Size: mz.Sm})
					if on && s.Tool != i {
						s.Tool = i
					}
				}
			})
		}
		// the chosen layer's color
		ui.Box(c).Size(36, 36).Margin(6, 0, 0, 0).Children(func() {
			ui.Box(c).Absolute().Left(12).Top(12).Size(22, 22).Radius(4).Background(ui.RGB(255, 255, 255)).Border(1, ui.RGBA(0, 0, 0, 0.4))
			ui.Box(c).Absolute().Left(0).Top(0).Size(22, 22).Radius(4).Background(s.Layers[s.Sel].Color).Border(1, ui.RGBA(255, 255, 255, 0.5))
		})
	})
}

// layers is the layers palette: the layers, the top one first, each with
// its visibility, and the chosen one's opacity.
func (s *studioPage) layers(c *ui.Context) {
	t := mz.From(c)
	cl := t.Colors
	ui.Column(c).Gap(2).Children(func() {
		for i := len(s.Layers) - 1; i >= 0; i-- {
			l := &s.Layers[i]
			row := ui.Row(c.Key(l.Name)).Gap(8).Padding(4, 6).AlignItems(ui.Center).Radius(t.R(6))
			if i == s.Sel {
				row.Background(cl.FillSelected)
			}
			row.Children(func() {
				mz.Checkbox(c.Key("visible"), &l.Visible, "", mz.CheckOptions{})
				b := ui.ButtonBase(c.Key("pick")).Label(l.Name).Grow(1).Gap(8).AlignItems(ui.Center).Justify(ui.Start).Cursor(ui.CursorPointer).FocusRing(false)
				if b.Clicked() {
					s.Sel = i
				}
				b.Children(func() {
					ui.Box(c).Size(30, 22).Radius(3).Background(s.adjusted(l.Color)).Border(1, ui.RGBA(255, 255, 255, 0.25))
					ui.Text(c, l.Name).FontSize(mz.FsSm).TextColor(cl.Text)
				})
			})
		}
	})
	l := &s.Layers[s.Sel]
	mz.Slider(c.Key("opacity"), &l.Opacity, 0, 100, mz.SliderOptions{Label: "Opacity", Format: func(v float64) string { return fmt.Sprintf("%.0f%%", v) }})
}

// swatches is the swatches palette: a grid of colors; one clicked colors
// the chosen layer.
func (s *studioPage) swatches(c *ui.Context) {
	hues := []float32{25, 55, 95, 145, 190, 235, 275, 330}
	ui.Column(c).Gap(3).Children(func() {
		for row, l := range []float32{0.86, 0.72, 0.58, 0.42} {
			ui.Row(c.Key(row)).Gap(3).Children(func() {
				for _, h := range hues {
					col := ui.Oklch(l, 0.15, h)
					b := ui.ButtonBase(c.Key(h)).Label(fmt.Sprintf("Color %.0f %.0f", l*100, h)).Size(20, 20).Radius(4).
						Background(col).Border(1, ui.RGBA(255, 255, 255, 0.2)).Cursor(ui.CursorPointer).FocusRing(false)
					if b.Hovered() {
						b.Border(1, ui.RGBA(255, 255, 255, 0.8))
					}
					if b.Clicked() {
						s.Layers[s.Sel].Color = col
					}
				}
			})
		}
	})
}

// adjust is the brightness and contrast HUD: two sliders, applied as they
// move; Cancel goes back, OK keeps them.
func (s *studioPage) adjust(c *ui.Context) {
	signed := func(v float64) string { return fmt.Sprintf("%+.0f", v) }
	cl := mz.From(c).Colors
	// a slider's label names it for assistive technology: the HUD shows it
	// on its left, as Pixelmator's does
	for _, a := range []struct {
		name string
		v    *float64
	}{{"Brightness", &s.Bright}, {"Contrast", &s.Contrast}} {
		ui.Row(c.Key(a.name)).Gap(10).AlignItems(ui.Center).Children(func() {
			ui.Text(c, a.name+":").FontSize(mz.FsSm).TextColor(cl.Text2).Width(74).Shrink(0)
			ui.Box(c).Grow(1).MinWidth(0).Children(func() {
				mz.Slider(c.Key("slider"), a.v, -100, 100, mz.SliderOptions{Label: a.name, Format: signed})
			})
		})
	}
	ui.Row(c).Gap(8).Children(func() {
		if mz.Button(c.Key("revert"), "Reset", mz.ButtonOptions{Size: mz.Sm}).Clicked() {
			s.Bright, s.Contrast = 0, 0
		}
		ui.Box(c).Grow(1)
		if mz.Button(c.Key("cancel"), "Cancel", mz.ButtonOptions{Size: mz.Sm}).Clicked() {
			s.Bright, s.Contrast = s.keep[0], s.keep[1]
			s.Board.Close("adjust")
		}
		if mz.Button(c.Key("ok"), "OK", mz.ButtonOptions{Size: mz.Sm, Color: mz.From(c).Colors.Accent, Variant: mz.Solid}).Clicked() {
			s.Board.Close("adjust")
		}
	})
}
