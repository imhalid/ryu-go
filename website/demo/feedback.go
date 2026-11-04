package main

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/mizan-go"
)

func init() {
	upload := 0.64
	demo("progress", func(c *ui.Context, t *mz.Tokens) {
		ui.Column(c).Gap(18).Width(420).Children(func() {
			ui.Row(c).Gap(12).AlignItems(ui.End).Children(func() {
				ui.Box(c).Grow(1).MinWidth(0).Children(func() {
					mz.Progress(c, upload, mz.ProgressOptions{Label: "Uploading Q3-report.pdf", ShowValue: true})
				})
				if mz.Button(c.Key("less"), "−10%", mz.ButtonOptions{Size: mz.Sm}).Clicked() {
					upload = max(0, upload-0.1)
				}
				if mz.Button(c.Key("more"), "+10%", mz.ButtonOptions{Size: mz.Sm}).Clicked() {
					upload = min(1, upload+0.1)
				}
			})
			mz.Progress(c, -1, mz.ProgressOptions{Label: "Syncing contacts", Size: mz.Sm})
		})
	})

	demo("meter", func(c *ui.Context, t *mz.Tokens) {
		gb := func(v float64) string { return fmt.Sprintf("%.0f of 64 GB", v) }
		ui.Column(c).Gap(18).Width(420).Children(func() {
			mz.Meter(c, 41, 0, 64, mz.MeterOptions{Label: "Storage", Format: gb, Warning: 48, Critical: 60})
			mz.Meter(c, 53, 0, 64, mz.MeterOptions{Label: "Backups", Format: gb, Warning: 48, Critical: 60})
			mz.Meter(c, 62, 0, 64, mz.MeterOptions{Label: "Photos", Format: gb, Warning: 48, Critical: 60})
		})
	})

	demo("spinner", func(c *ui.Context, t *mz.Tokens) {
		ui.Row(c).Gap(16).AlignItems(ui.Center).Children(func() {
			mz.Spinner(c, mz.SpinnerOptions{Size: 14})
			mz.Spinner(c, mz.SpinnerOptions{})
			mz.Spinner(c, mz.SpinnerOptions{Size: 24})
			mz.Spinner(c, mz.SpinnerOptions{Color: t.Colors.Muted})
			mz.Button(c.Key("saving"), "Saving", mz.ButtonOptions{Loading: true})
		})
	})

	loaded := false
	demo("skeleton", func(c *ui.Context, t *mz.Tokens) {
		ui.Column(c).Gap(16).Width(380).Children(func() {
			ui.Row(c).Gap(12).AlignItems(ui.Start).Children(func() {
				if loaded {
					mz.Avatar(c, "Maya Chen", mz.AvatarOptions{})
					ui.Column(c).Grow(1).MinWidth(0).Gap(4).Children(func() {
						ui.Text(c, "Maya Chen").FontWeight(600).FontSize(mz.FsMd)
						ui.Text(c, "Shared Q3-report.pdf with the team and left two comments on the retention chart.").
							TextColor(t.Colors.Text2).FontSize(mz.FsMd).Wrap()
					})
					return
				}
				mz.Skeleton(c, 36, 36).Radius(mz.RadiusPill)
				ui.Column(c).Grow(1).MinWidth(0).Gap(10).Padding(4, 0, 0, 0).Children(func() {
					mz.Skeleton(c, 120, 12)
					mz.SkeletonText(c, 2)
				})
			})
			label := "Load"
			if loaded {
				label = "Show placeholders"
			}
			if mz.Button(c.Key("load"), label, mz.ButtonOptions{Size: mz.Sm}).AlignSelf(ui.Start).Clicked() {
				loaded = !loaded
			}
		})
	})
}
