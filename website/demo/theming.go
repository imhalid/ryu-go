package main

import (
	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/mizan-go"
)

// accents are the brand colors the theming demo offers, Mizan's blue first.
var accents = []struct {
	Name  string
	Color ui.Color
}{
	{"Blue", ui.Color{}},
	{"Violet", ui.Oklch(0.56, 0.2, 295)},
	{"Green", ui.Oklch(0.6, 0.15, 155)},
	{"Orange", ui.Oklch(0.68, 0.18, 50)},
	{"Rose", ui.Oklch(0.6, 0.2, 10)},
}

func init() {
	accent, size, round := 0, 1, 1
	on, plan, query := true, "week", ""
	demo("theming", func(c *ui.Context, t *mz.Tokens) {
		ui.Column(c).Gap(10).AlignItems(ui.Center).Children(func() {
			ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
				setting(c, t, "Size", &size)
				setting(c, t, "Radius", &round)
			})
			names := make([]string, len(accents))
			for i, a := range accents {
				names[i] = a.Name
			}
			mz.Segmented(c.Key("accent"), &accent, names, mz.SegmentedOptions{Label: "Accent", Size: mz.Sm})
		})
		mz.Scope(c.Key("scope"), mz.Options{
			Accent:  accents[accent].Color,
			Size:    [...]mz.Size{mz.Sm, mz.Md, mz.Lg}[size],
			Corners: [...]mz.Corners{mz.CornersSm, mz.CornersMd, mz.CornersLg}[round],
		}).Children(func() {
			s := mz.From(c)
			ui.Column(c).Width(380).Children(func() {
				mz.Card(c, mz.CardOptions{Title: "Notifications", Pad: true}, func() {
					ui.Column(c).Gap(14).Children(func() {
						mz.Input(c.Key("q"), &query, mz.InputOptions{Label: "Search", Placeholder: "Search settings", PrefixIcon: iconSearch})
						mz.Switch(c.Key("on"), &on, mz.SwitchOptions{Label: "Email me a digest"})
						ui.RadioGroup(c, func() {
							ui.Row(c).Gap(16).Children(func() {
								mz.Radio(c.Key("day"), &plan, "day", "Daily", mz.CheckOptions{})
								mz.Radio(c.Key("week"), &plan, "week", "Weekly", mz.CheckOptions{})
							})
						}).Label("Frequency")
						ui.Row(c).Gap(mz.GapControl).Justify(ui.End).Children(func() {
							mz.Button(c.Key("cancel"), "Cancel", mz.ButtonOptions{})
							mz.Button(c.Key("save"), "Save", mz.ButtonOptions{Color: s.Colors.Accent, Variant: mz.Solid})
						})
					})
				})
			})
		})
	})
}

// setting is a small, muted label and a compact S / M / L choice.
func setting(c *ui.Context, t *mz.Tokens, label string, at *int) {
	ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
		ui.Text(c, label).FontSize(mz.FsXs).FontWeight(500).TextColor(t.Colors.Muted).NoWrap()
		mz.Segmented(c.Key(label), at, []string{"S", "M", "L"}, mz.SegmentedOptions{Label: label, Size: mz.Sm})
	})
}
