package main

import (
	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/mizan-go"
)

func init() {
	tab := 0
	demo("tabs", func(c *ui.Context, t *mz.Tokens) {
		ui.Column(c).Gap(14).Width(440).Children(func() {
			mz.Tabs(c.Key("tabs"), &tab, []mz.Tab{{Label: "Overview"}, {Label: "Activity", Badge: "12"}, {Label: "Members"}, {Label: "Settings"}},
				mz.TabsOptions{Label: "Report"})
			mz.TabPanel(c, tab, func() {
				mz.Text(c, []string{
					"Weekly active users rose 4.2% after the new onboarding.",
					"12 comments and 3 exports since Monday.",
					"Maya Chen, Leo Park and 4 others can edit.",
					"Exports, retention and sharing.",
				}[tab], mz.TBody).TextColor(t.Colors.Text2)
			})
		})
	})

	details := false
	demo("link", func(c *ui.Context, t *mz.Tokens) {
		ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "The export is ready.").TextColor(t.Colors.Text2).FontSize(mz.FsMd)
			mz.Link(c, "Read the changelog", "https://github.com/imhalid/mizan-go").FontSize(mz.FsMd)
		})
		label := "Show details"
		if details {
			label = "Hide details"
		}
		if mz.Link(c, label, "").Clicked() {
			details = !details
		}
		if details {
			status(c, t, "A link without a URL is a button that reads as a link.")
		}
	})

	page := 4
	demo("pagination", func(c *ui.Context, t *mz.Tokens) {
		ui.Column(c).Width(480).Children(func() {
			mz.Pagination(c.Key("pages"), &page, 240, mz.PaginationOptions{PageSize: 20})
		})
	})

	advanced, verbose, beta := false, false, false
	demo("collapsible", func(c *ui.Context, t *mz.Tokens) {
		ui.Column(c).Width(360).Children(func() {
			mz.Collapsible(c.Key("adv"), "Advanced options", &advanced, mz.CollapsibleOptions{Description: "Logging and beta features"}, func() {
				mz.Checkbox(c.Key("verbose"), &verbose, "Verbose logging", mz.CheckOptions{})
				mz.Checkbox(c.Key("beta"), &beta, "Beta features", mz.CheckOptions{Description: "They may change without notice."})
			})
		})
	})

	billing, security, export := true, false, false
	demo("accordion", func(c *ui.Context, t *mz.Tokens) {
		ui.Column(c).Width(400).Children(func() {
			mz.Accordion(c.Key("acc"), mz.AccordionOptions{Single: true}, func(a *mz.AccordionList) {
				a.Item("Billing", &billing, func() {
					ui.Text(c, "Invoices go out on the first of the month, to the address on file.").Wrap()
				})
				a.ItemWith("Security", &security, mz.CollapsibleOptions{Description: "Two-factor sign-in, sessions"}, func() {
					ui.Text(c, "Two-factor sign-in is on for every member of the team.").Wrap()
				})
				a.Item("Export", &export, func() {
					ui.Text(c, "Exports are kept for 30 days, then deleted.").Wrap()
				})
			})
		})
	})

	bold, italic, under, left, center, pin := true, false, false, true, false, false
	demo("toggle", func(c *ui.Context, t *mz.Tokens) {
		ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
			mz.ToggleGroup(c.Key("style"), "Text style", func() {
				mz.Toggle(c.Key("b"), &bold, "Bold", mz.ToggleOptions{Icon: iconBold, IconOnly: true})
				mz.Toggle(c.Key("i"), &italic, "Italic", mz.ToggleOptions{Icon: iconItalic, IconOnly: true})
				mz.Toggle(c.Key("u"), &under, "Underline", mz.ToggleOptions{Icon: iconUnder, IconOnly: true})
				mz.Separator(c, mz.SeparatorOptions{Vertical: true})
				if mz.Toggle(c.Key("l"), &left, "Align left", mz.ToggleOptions{Icon: iconAlignL, IconOnly: true}).Changed() && left {
					center = false
				}
				if mz.Toggle(c.Key("c"), &center, "Align center", mz.ToggleOptions{Icon: iconAlignC, IconOnly: true}).Changed() && center {
					left = false
				}
			})
			mz.Toggle(c.Key("pin"), &pin, "Pin", mz.ToggleOptions{Icon: iconPin})
		})
		status(c, t, "The arrows move between the toolbar's toggles.")
	})

	demo("separator", func(c *ui.Context, t *mz.Tokens) {
		ui.Column(c).Gap(14).Width(320).Children(func() {
			mz.Button(c.Key("google"), "Continue with Google", mz.ButtonOptions{}).AlignSelf(ui.Stretch)
			mz.Separator(c, mz.SeparatorOptions{Label: "or"})
			mz.Input(c.Key("email"), new(string), mz.InputOptions{Label: "Email", Placeholder: "you@example.com"})
			mz.Separator(c, mz.SeparatorOptions{})
			ui.Row(c).Gap(10).Height(20).AlignItems(ui.Center).Justify(ui.Center).Children(func() {
				mz.Link(c, "Terms", "https://example.com/terms").FontSize(mz.FsSm)
				mz.Separator(c, mz.SeparatorOptions{Vertical: true})
				mz.Link(c, "Privacy", "https://example.com/privacy").FontSize(mz.FsSm)
			})
		})
	})
}
