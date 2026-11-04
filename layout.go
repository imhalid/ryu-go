package mizan

import (
	"math"
	"time"

	"github.com/egoist/mygo/ui"
)

// CardOptions are the options of Card.
type CardOptions struct {
	Title    string
	Subtitle string
	// Pad gives the panel pad-card (28px) padding; otherwise only 10px
	// vertically (a table or a list inside sits edge to edge).
	Pad bool
	// Actions builds the buttons to the right of the title.
	Actions func()
}

// Card is a surface whose title stands outside the panel, above it. The panel
// is in the card color with radius-lg corners; its border is a gradient,
// strong at the top and fading downwards (in the dark theme shadows don't
// show, so the border separates).
//
//	mizan.Card(c, mizan.CardOptions{Title: "Active users", Pad: true}, func() {
//		...
//	})
func Card(c *ui.Context, o CardOptions, body func()) ui.Element {
	t := From(c)
	cl := t.Colors
	card := ui.Column(c).MinWidth(0)
	card.Children(func() {
		if o.Title != "" || o.Actions != nil {
			ui.Row(c).AlignItems(ui.Start).Justify(ui.SpaceBetween).Gap(16).Padding(12, 4, 14, 4).Children(func() {
				ui.Column(c).Grow(1).MinWidth(0).Gap(3).Children(func() {
					if o.Title != "" {
						Text(c, o.Title, TTitle).Level(2).Role(ui.RoleHeading)
					}
					if o.Subtitle != "" {
						ui.Text(c, o.Subtitle).TextColor(cl.Muted).FontSize(FsMd)
					}
				})
				if o.Actions != nil {
					ui.Row(c).AlignItems(ui.Center).Gap(GapControl).Children(o.Actions)
				}
			})
		}
		panel := ui.Column(c).MinWidth(0).Radius(t.R(RadiusLg)).Background(cl.Card)
		if o.Pad {
			panel.Padding(PadCard)
		} else {
			panel.Padding(10, 0)
		}
		panel.Border(1, ui.Color{}).BorderLinearGradient(t.edge()).Children(body)
	})
	return card
}

// edge is the card border's gradient: strongest at 10%, faintest at 90%.
func (t *Tokens) edge() ui.LinearGradient {
	c := t.Colors
	return ui.LinearGradient{Angle: 180, Stops: []ui.GradientStop{
		{Color: c.GradientBorder10, At: 0.1},
		{Color: c.GradientBorder25, At: 0.25},
		{Color: c.GradientBorder75, At: 0.75},
		{Color: c.GradientBorder90, At: 0.9},
	}}
}

// FieldOptions are the options of Field.
type FieldOptions struct {
	Label string
	// Help is the muted description under the field; when Error is set, the
	// error is written in the Up color instead.
	Help  string
	Error string
	// Required puts a muted asterisk next to the label.
	Required bool
	// Shake, changed, shakes the field gently sideways, motion only: bump it
	// as a submit finds the field invalid (and give its control the focus).
	Shake int
}

// Field is a vertical group of a label, a control and a description: the
// label is fs-sm 500 muted, gap-field away from the control. In an invalid
// field the label turns the Up color; the control itself is marked with
// Invalid.
//
//	mizan.Field(c, mizan.FieldOptions{Label: "Email", Error: app.emailErr}, func() {
//		mizan.Input(c.Key("email"), &app.email, mizan.InputOptions{Label: "Email", Invalid: app.emailErr != ""})
//	})
func Field(c *ui.Context, o FieldOptions, control func()) ui.Element {
	cl := From(c).Colors
	f := ui.Column(c).Gap(GapField).MinWidth(0)
	// shaken: a few swings sideways, each smaller, over 420 ms
	shaken := ui.Local(f, "shake", func() int { return o.Shake })
	if *shaken != o.Shake {
		*shaken = o.Shake
		f.Jump("shake", 0)
	}
	if p := f.AnimateTo("shake", 1, ui.Animation{Duration: 420 * time.Millisecond}); p < 1 {
		f.Translate(float32(math.Sin(float64(p)*math.Pi*5))*7*(1-p), 0)
	}
	label := cl.Muted
	if o.Error != "" {
		label = cl.Up
	} else if f.FocusWithin() {
		label = cl.Text
	}
	f.Children(func() {
		ui.Row(c).Gap(3).Children(func() {
			ui.Text(c, o.Label).FontSize(FsSm).FontWeight(500).TextColor(label)
			if o.Required {
				ui.Text(c, "*").FontSize(FsSm).TextColor(cl.Muted)
			}
		})
		control()
		switch {
		case o.Error != "":
			ui.Text(c, o.Error).FontSize(FsXs).TextColor(cl.Up)
		case o.Help != "":
			ui.Text(c, o.Help).FontSize(FsXs).TextColor(cl.Muted)
		}
	})
	return f
}
