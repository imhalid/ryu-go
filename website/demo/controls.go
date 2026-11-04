package main

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/mizan-go"
)

func init() {
	presses := 0
	demo("button", func(c *ui.Context, t *mz.Tokens) {
		cl := t.Colors
		ui.Row(c).Gap(mz.GapControl).Wrap().Justify(ui.Center).AlignItems(ui.Center).Children(func() {
			if mz.Button(c.Key("default"), "Default", mz.ButtonOptions{}).Clicked() {
				presses++
			}
			mz.Button(c.Key("search"), "Search", mz.ButtonOptions{Icon: iconSearch})
			mz.Button(c.Key("iconOnly"), "Search", mz.ButtonOptions{Icon: iconSearch, IconOnly: true})
			mz.Button(c.Key("soft"), "Accent", mz.ButtonOptions{Color: cl.Accent})
			mz.Button(c.Key("solid"), "Export", mz.ButtonOptions{Color: cl.Accent, Variant: mz.Solid})
		})
		ui.Row(c).Gap(mz.GapControl).Wrap().Justify(ui.Center).AlignItems(ui.Center).Children(func() {
			mz.Button(c.Key("danger"), "Delete", mz.ButtonOptions{Danger: true})
			mz.Button(c.Key("dangerSoft"), "Remove", mz.ButtonOptions{Danger: true, Variant: mz.Soft})
			mz.Button(c.Key("loading"), "Saving", mz.ButtonOptions{Loading: true})
			mz.Button(c.Key("disabled"), "Disabled", mz.ButtonOptions{Disabled: true})
		})
		status(c, t, fmt.Sprintf("Default pressed %d times", presses))
	})

	demo("button-colors", func(c *ui.Context, t *mz.Tokens) {
		cl := t.Colors
		ui.Row(c).Gap(mz.GapControl).Wrap().Justify(ui.Center).Children(func() {
			for i, col := range []ui.Color{cl.C1, cl.C2, cl.C3, cl.C4, cl.C5, cl.C6, cl.C7, cl.C8} {
				mz.Button(c.Key(i), fmt.Sprintf("C%d", i+1), mz.ButtonOptions{Color: col, Size: mz.Sm})
			}
		})
		ui.Row(c).Gap(mz.GapControl).AlignItems(ui.Center).Children(func() {
			for _, z := range []mz.Size{mz.Sm, mz.Md, mz.Lg} {
				mz.Button(c.Key(z), [...]string{mz.Sm: "Small", mz.Md: "Medium", mz.Lg: "Large"}[z], mz.ButtonOptions{Size: z})
			}
		})
	})

	demo("copy-button", func(c *ui.Context, t *mz.Tokens) {
		ui.Row(c).Gap(mz.GapControl).AlignItems(ui.Center).Children(func() {
			mz.Text(c, "https://mizan.example/q3-report", mz.TCode).TextColor(t.Colors.Text2)
			mz.CopyButton(c.Key("copy"), "https://mizan.example/q3-report", mz.CopyOptions{Label: "Copy link"})
		})
	})

	query, email, pass := "", "ada@", "secret"
	demo("input", func(c *ui.Context, t *mz.Tokens) {
		ui.Column(c).Gap(12).Width(320).Children(func() {
			mz.Input(c.Key("q"), &query, mz.InputOptions{Label: "Search", Placeholder: "Patients, files, reports…", PrefixIcon: iconSearch, Clearable: true})
			mz.Input(c.Key("email"), &email, mz.InputOptions{Label: "Email", Invalid: true})
			mz.Input(c.Key("pass"), &pass, mz.InputOptions{Label: "Password", Password: true})
			mz.Input(c.Key("price"), new(string), mz.InputOptions{Label: "Price", Prefix: "$", Suffix: "USD", Placeholder: "0.00"})
		})
	})

	period, team := 30, ""
	demo("select", func(c *ui.Context, t *mz.Tokens) {
		ui.Row(c).Gap(12).Children(func() {
			ui.Column(c).Width(200).Children(func() {
				mz.Select(c.Key("range"), &period, []mz.Option[int]{
					{Value: 7, Label: "Last 7 days"},
					{Value: 30, Label: "Last 30 days", Description: "default"},
					{Value: 90, Label: "Last 90 days"},
					{Value: 365, Label: "Last year", Disabled: true},
				}, mz.SelectOptions{Label: "Period"})
			})
			ui.Column(c).Width(200).Children(func() {
				mz.Select(c.Key("team"), &team, []mz.Option[string]{
					{Value: "web", Label: "Web", Dot: t.Colors.C1},
					{Value: "ios", Label: "iOS", Dot: t.Colors.C2},
					{Value: "and", Label: "Android", Dot: t.Colors.C3},
				}, mz.SelectOptions{Label: "Team", Placeholder: "Choose a team"})
			})
		})
		chosen := team
		if chosen == "" {
			chosen = "none"
		}
		status(c, t, fmt.Sprintf("Period: %d days · team: %s", period, chosen))
	})

	name, mail, shake := "", "ada@", 0
	demo("field", func(c *ui.Context, t *mz.Tokens) {
		ui.Column(c).Gap(16).Width(320).Children(func() {
			mz.Field(c, mz.FieldOptions{Label: "Name", Help: "As it appears on invoices."}, func() {
				mz.Input(c.Key("name"), &name, mz.InputOptions{Label: "Name", Placeholder: "Ada Lovelace"})
			})
			mz.Field(c, mz.FieldOptions{Label: "Email", Error: "Enter a valid email address.", Required: true, Shake: shake}, func() {
				mz.Input(c.Key("email"), &mail, mz.InputOptions{Label: "Email", Invalid: true})
			})
			// a check that finds the email invalid shakes its field
			if mz.Button(c.Key("check"), "Check form", mz.ButtonOptions{Size: mz.Sm}).AlignSelf(ui.Start).Clicked() {
				shake++
			}
		})
	})

	demo("card", func(c *ui.Context, t *mz.Tokens) {
		ui.Column(c).Width(380).Children(func() {
			mz.Card(c, mz.CardOptions{Title: "Weekly report", Subtitle: "Updated 5 minutes ago", Pad: true, Actions: func() {
				mz.Button(c.Key("share"), "Share", mz.ButtonOptions{Icon: iconShare, IconOnly: true, Size: mz.Sm})
			}}, func() {
				ui.Row(c).Gap(24).Children(func() {
					for _, kv := range [][2]string{{"Active users", "12,480"}, {"Retention", "41.8%"}} {
						ui.Column(c.Key(kv[0])).Gap(2).Children(func() {
							mz.Text(c, kv[0], mz.TCaption).TextColor(t.Colors.Muted)
							mz.Text(c, kv[1], mz.TFigure)
						})
					}
				})
			})
		})
	})
}
