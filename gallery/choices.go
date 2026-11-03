package gallery

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/mizan-go"
)

func (s *State) choices(c *ui.Context, t *mz.Tokens) {
	mz.Card(c, mz.CardOptions{Title: "Seçim", Subtitle: "Boşken gömük, seçilince vurgu gradyanı; topuzlar kuyunun üstünde.", Pad: true}, func() {
		ui.Column(c).Gap(18).Children(func() {
			ui.Row(c).Gap(28).AlignItems(ui.Start).Children(func() {
				ui.Column(c).Gap(10).Children(func() {
					mz.Checkbox(c.Key("mail"), &s.Mail, "Bülten", mz.CheckOptions{})
					mz.Checkbox(c.Key("push"), &s.Push, "Anlık bildirim", mz.CheckOptions{Description: "Yalnız bu cihazda."})
					mz.Checkbox(c.Key("all"), &s.Digest, "Tümü", mz.CheckOptions{Indeterminate: s.Mail != s.Push})
					mz.Checkbox(c.Key("off"), new(bool), "Kapalı", mz.CheckOptions{Disabled: true})
				})
				ui.RadioGroup(c, func() {
					ui.Column(c).Gap(10).Children(func() {
						mz.Radio(c.Key("day"), &s.Plan, "day", "Günlük", mz.CheckOptions{})
						mz.Radio(c.Key("week"), &s.Plan, "week", "Haftalık", mz.CheckOptions{})
						mz.Radio(c.Key("month"), &s.Plan, "month", "Aylık", mz.CheckOptions{})
					})
				}).Label("Özet sıklığı")
				ui.Column(c).Gap(12).Children(func() {
					mz.Switch(c.Key("sync"), &s.Sync, mz.SwitchOptions{Label: "Eşitleme"})
					mz.Switch(c.Key("led"), &s.LED, mz.SwitchOptions{Label: "Karanlık mod"})
				})
			})
			ui.RadioGroup(c, func() {
				ui.Row(c).Gap(8).Children(func() {
					for _, v := range [][3]string{{"free", "Ücretsiz", "3 proje"}, {"pro", "Pro", "Sınırsız proje"}, {"team", "Kurumsal", "Ortak çalışma"}} {
						ui.Column(c.Key(v[0])).Grow(1).Basis(0).MinWidth(0).Children(func() {
							mz.Choice(c, &s.Tier, v[0], v[1], v[2], mz.CheckOptions{})
						})
					}
				})
			}).Label("Paket")
			mz.Segmented(c.Key("view"), &s.ViewMode, []string{"Liste", "Izgara", "Takvim", "Pano"}, mz.SegmentedOptions{Label: "Görünüm"})
			mz.Slider(c.Key("vol"), &s.Volume, 0, 100, mz.SliderOptions{Label: "Ses", Step: 1,
				Format: func(v float64) string { return fmt.Sprintf("%%%.0f", v) }})
			mz.Slider(c.Key("zoom"), &s.Zoom, 0.5, 2, mz.SliderOptions{Label: "Yakınlık", Step: 0.05,
				Format: func(v float64) string { return fmt.Sprintf("%.2f×", v) }})
		})
	})
}
