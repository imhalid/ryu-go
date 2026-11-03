package gallery

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/mizan-go"
)

func (s *State) choices(c *ui.Context, t *mz.Tokens) {
	mz.Card(c, mz.CardOptions{Title: "Selection", Subtitle: "Inset when empty, accent gradient when selected; knobs sit above the well.", Pad: true}, func() {
		ui.Column(c).Gap(18).Children(func() {
			ui.Row(c).Gap(28).AlignItems(ui.Start).Children(func() {
				ui.Column(c).Gap(10).Children(func() {
					mz.Checkbox(c.Key("mail"), &s.Mail, "Newsletter", mz.CheckOptions{})
					mz.Checkbox(c.Key("push"), &s.Push, "Push notifications", mz.CheckOptions{Description: "Only on this device."})
					mz.Checkbox(c.Key("all"), &s.Digest, "All", mz.CheckOptions{Indeterminate: s.Mail != s.Push})
					mz.Checkbox(c.Key("off"), new(bool), "Disabled", mz.CheckOptions{Disabled: true})
				})
				ui.RadioGroup(c, func() {
					ui.Column(c).Gap(10).Children(func() {
						mz.Radio(c.Key("day"), &s.Plan, "day", "Daily", mz.CheckOptions{})
						mz.Radio(c.Key("week"), &s.Plan, "week", "Weekly", mz.CheckOptions{})
						mz.Radio(c.Key("month"), &s.Plan, "month", "Monthly", mz.CheckOptions{})
					})
				}).Label("Digest frequency")
				ui.Column(c).Gap(12).Children(func() {
					mz.Switch(c.Key("sync"), &s.Sync, mz.SwitchOptions{Label: "Sync"})
					mz.Switch(c.Key("led"), &s.LED, mz.SwitchOptions{Label: "Dark mode"})
				})
			})
			ui.RadioGroup(c, func() {
				ui.Row(c).Gap(8).Children(func() {
					for _, v := range [][3]string{{"free", "Free", "3 projects"}, {"pro", "Pro", "Unlimited projects"}, {"team", "Enterprise", "Collaboration"}} {
						ui.Column(c.Key(v[0])).Grow(1).Basis(0).MinWidth(0).Children(func() {
							mz.Choice(c, &s.Tier, v[0], v[1], v[2], mz.CheckOptions{})
						})
					}
				})
			}).Label("Plan")
			mz.Segmented(c.Key("view"), &s.ViewMode, []string{"List", "Grid", "Calendar", "Board"}, mz.SegmentedOptions{Label: "View"})
			mz.Slider(c.Key("vol"), &s.Volume, 0, 100, mz.SliderOptions{Label: "Volume", Step: 1,
				Format: func(v float64) string { return fmt.Sprintf("%.0f%%", v) }})
			mz.Slider(c.Key("zoom"), &s.Zoom, 0.5, 2, mz.SliderOptions{Label: "Zoom", Step: 0.05,
				Format: func(v float64) string { return fmt.Sprintf("%.2f×", v) }})
		})
	})
}
