package main

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/mizan-go"
)

func init() {
	demo("badge", func(c *ui.Context, t *mz.Tokens) {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			mz.Badge(c, "+4.2%", mz.ToneDown)
			mz.Badge(c, "−1.8%", mz.ToneUp)
			mz.Badge(c, "Pending", mz.ToneWarn)
			mz.Badge(c, "Draft", mz.ToneNeutral)
			mz.Badge(c, "Default", mz.ToneDefault)
		})
		ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
			mz.Text(c, "Quarterly report", mz.TTitle)
			mz.Tag(c, "v2.4")
		})
	})

	demo("status-dot", func(c *ui.Context, t *mz.Tokens) {
		ui.Row(c).Gap(18).AlignItems(ui.Center).Children(func() {
			mz.StatusDot(c, mz.DotLive, "Live")
			mz.StatusDot(c, mz.DotDelayed, "Delayed")
			mz.StatusDot(c, mz.DotOff, "Offline")
		})
	})

	demo("kbd", func(c *ui.Context, t *mz.Tokens) {
		ui.Column(c).Gap(10).Children(func() {
			for _, k := range []struct {
				what string
				keys []string
			}{{"Search", []string{"⌘", "K"}}, {"New report", []string{"⌘", "N"}}, {"Close", []string{"Esc"}}, {"Move a row", []string{"⌥", "↑"}}} {
				ui.Row(c.Key(k.what)).Gap(12).Width(200).AlignItems(ui.Center).Children(func() {
					ui.Text(c, k.what).TextColor(t.Colors.Text2).FontSize(mz.FsMd).Grow(1)
					mz.Kbd(c, k.keys...)
				})
			}
		})
	})

	demo("avatar", func(c *ui.Context, t *mz.Tokens) {
		ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
			for _, n := range []string{"Maya Chen", "Leo Park", "Sam Rivera", "Noah Kim"} {
				mz.Avatar(c.Key(n), n, mz.AvatarOptions{Online: n == "Leo Park"})
			}
			mz.Avatar(c.Key("solid"), "Ivy Brooks", mz.AvatarOptions{Solid: true})
		})
		ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
			for _, z := range []float32{24, 32, 40, 52} {
				mz.Avatar(c.Key(z), "Ada Lovelace", mz.AvatarOptions{Size: z, Color: t.Colors.C5})
			}
		})
	})

	added := 0
	demo("empty-state", func(c *ui.Context, t *mz.Tokens) {
		ui.Column(c).Width(360).Children(func() {
			mz.EmptyState(c, mz.EmptyStateOptions{Icon: iconInbox, Title: "No files", Description: "Files you add are listed here.", Action: func() {
				if mz.Button(c.Key("add"), "Add a file", mz.ButtonOptions{Size: mz.Sm}).Clicked() {
					added++
				}
			}})
		})
		if added > 0 {
			status(c, t, fmt.Sprintf("Add pressed %d times", added))
		}
	})
}
