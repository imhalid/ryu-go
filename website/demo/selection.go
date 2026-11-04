package main

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/ryu-go"
)

func init() {
	news, push := true, false
	demo("checkbox", func(c *ui.Context, t *mz.Tokens) {
		ui.Column(c).Gap(10).Children(func() {
			all := news && push
			if mz.Checkbox(c.Key("all"), &all, "All notifications", mz.CheckOptions{Indeterminate: news != push}).Changed() {
				news, push = all, all
			}
			ui.Column(c).Gap(10).Padding(0, 0, 0, 26).Children(func() {
				mz.Checkbox(c.Key("mail"), &news, "Newsletter", mz.CheckOptions{})
				mz.Checkbox(c.Key("push"), &push, "Push notifications", mz.CheckOptions{Description: "Only on this device."})
			})
			mz.Checkbox(c.Key("off"), new(bool), "Disabled", mz.CheckOptions{Disabled: true})
		})
	})

	plan := "week"
	demo("radio", func(c *ui.Context, t *mz.Tokens) {
		ui.RadioGroup(c, func() {
			ui.Column(c).Gap(10).Children(func() {
				mz.Radio(c.Key("day"), &plan, "day", "Daily", mz.CheckOptions{})
				mz.Radio(c.Key("week"), &plan, "week", "Weekly", mz.CheckOptions{})
				mz.Radio(c.Key("month"), &plan, "month", "Monthly", mz.CheckOptions{Description: "On the first of the month."})
				mz.Radio(c.Key("never"), &plan, "never", "Never", mz.CheckOptions{Disabled: true})
			})
		}).Label("Digest frequency")
		status(c, t, "Digest: "+plan)
	})

	tier := "pro"
	demo("choice", func(c *ui.Context, t *mz.Tokens) {
		ui.RadioGroup(c, func() {
			ui.Row(c).Gap(8).Width(560).Children(func() {
				for _, v := range [][3]string{{"free", "Free", "3 projects"}, {"pro", "Pro", "Unlimited projects"}, {"team", "Enterprise", "Collaboration"}} {
					ui.Column(c.Key(v[0])).Grow(1).Basis(0).MinWidth(0).Children(func() {
						mz.Choice(c, &tier, v[0], v[1], v[2], mz.CheckOptions{})
					})
				}
			})
		}).Label("Plan")
	})

	sync, dark := true, false
	demo("switch", func(c *ui.Context, t *mz.Tokens) {
		ui.Column(c).Gap(14).Children(func() {
			mz.Switch(c.Key("sync"), &sync, mz.SwitchOptions{Label: "Sync", Description: "Refresh the data every hour."})
			mz.Switch(c.Key("dark"), &dark, mz.SwitchOptions{Label: "Dark mode"})
			mz.Switch(c.Key("off"), new(bool), mz.SwitchOptions{Label: "Disabled", Disabled: true})
		})
	})

	view, size := 0, 1
	demo("segmented", func(c *ui.Context, t *mz.Tokens) {
		ui.Column(c).Gap(12).AlignItems(ui.Center).Children(func() {
			mz.Segmented(c.Key("view"), &view, []string{"List", "Grid", "Calendar", "Board"}, mz.SegmentedOptions{Label: "View"})
			mz.Segmented(c.Key("size"), &size, []string{"S", "M", "L"}, mz.SegmentedOptions{Label: "Size", Size: mz.Sm})
		})
	})

	volume, zoom := 64.0, 1.25
	demo("slider", func(c *ui.Context, t *mz.Tokens) {
		ui.Column(c).Gap(18).Width(340).Children(func() {
			mz.Slider(c.Key("vol"), &volume, 0, 100, mz.SliderOptions{Label: "Volume", Step: 1,
				Format: func(v float64) string { return fmt.Sprintf("%.0f%%", v) }})
			mz.Slider(c.Key("zoom"), &zoom, 0.5, 2, mz.SliderOptions{Label: "Zoom", Step: 0.05,
				Format: func(v float64) string { return fmt.Sprintf("%.2f×", v) }})
		})
	})
}
