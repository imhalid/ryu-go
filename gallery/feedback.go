package gallery

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/mizan-go"
)

// feedbackPage is the state of the feedback page: progress, disclosure,
// toggles and loading.
type feedbackPage struct {
	Upload                    float64
	Advanced, Verbose, Beta   bool
	Billing, Security, Export bool
	Bold, Italic, Underline   bool
	Left, Center, Pin         bool
	Loaded                    bool
}

func newFeedback() feedbackPage {
	return feedbackPage{Upload: 0.64, Billing: true, Bold: true, Left: true}
}

var (
	iconBold      = mz.Svg(16, 15, "M5 3h3.8a2.5 2.5 0 0 1 0 5H5zM5 8h4.3a2.5 2.5 0 0 1 0 5H5z")
	iconItalic    = mz.Svg(16, 15, "M7 3h5M4 13h5M9.5 3l-3 10")
	iconUnderline = mz.Svg(16, 15, "M5 3v4.5a3 3 0 0 0 6 0V3M4 13.5h8")
	iconAlignL    = mz.Svg(16, 15, "M3 4h10M3 8h6M3 12h8")
	iconAlignC    = mz.Svg(16, 15, "M3 4h10M5 8h6M4 12h8")
	iconPin       = mz.Svg(16, 15, "M6 2.5h4M7 2.5v4L4.5 9h7L9 6.5v-4M8 9v4.5")
)

func (f *feedbackPage) view(c *ui.Context, t *mz.Tokens) {
	ui.Row(c).Gap(mz.GapSection).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(mz.GapSection).Children(func() {
			f.progress(c, t)
			f.disclosure(c, t)
		})
		ui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(mz.GapSection).Children(func() {
			f.toolbar(c, t)
			f.loading(c, t)
		})
	})
}

func (f *feedbackPage) progress(c *ui.Context, t *mz.Tokens) {
	mz.Card(c, mz.CardOptions{Title: "Progress and meter", Subtitle: "Inset wells that fill on a spring.", Pad: true}, func() {
		ui.Column(c).Gap(18).Children(func() {
			ui.Row(c).Gap(12).AlignItems(ui.End).Children(func() {
				ui.Box(c).Grow(1).MinWidth(0).Children(func() {
					mz.Progress(c, f.Upload, mz.ProgressOptions{Label: "Uploading Q3-report.pdf", ShowValue: true})
				})
				if mz.Button(c.Key("less"), "−10%", mz.ButtonOptions{Size: mz.Sm}).Clicked() {
					f.Upload = max(0, f.Upload-0.1)
				}
				if mz.Button(c.Key("more"), "+10%", mz.ButtonOptions{Size: mz.Sm}).Clicked() {
					f.Upload = min(1, f.Upload+0.1)
				}
			})
			mz.Progress(c, -1, mz.ProgressOptions{Label: "Syncing contacts", Size: mz.Sm})
			gb := func(v float64) string { return fmt.Sprintf("%.0f of 64 GB", v) }
			mz.Meter(c, 41, 0, 64, mz.MeterOptions{Label: "Storage", Format: gb, Warning: 48, Critical: 60})
			mz.Meter(c, 53, 0, 64, mz.MeterOptions{Label: "Backups", Format: gb, Warning: 48, Critical: 60})
			mz.Meter(c, 62, 0, 64, mz.MeterOptions{Label: "Photos", Format: gb, Warning: 48, Critical: 60})
			ui.Row(c).Gap(16).AlignItems(ui.Center).Children(func() {
				mz.Spinner(c, mz.SpinnerOptions{Size: 14})
				mz.Spinner(c, mz.SpinnerOptions{})
				mz.Spinner(c, mz.SpinnerOptions{Size: 24})
				mz.Spinner(c, mz.SpinnerOptions{Color: t.Colors.Muted})
				mz.Button(c.Key("saving"), "Saving", mz.ButtonOptions{Loading: true})
			})
		})
	})
}

func (f *feedbackPage) disclosure(c *ui.Context, t *mz.Tokens) {
	mz.Card(c, mz.CardOptions{Title: "Collapsible and accordion", Subtitle: "The chevron turns; the body grows into view.", Pad: true}, func() {
		ui.Column(c).Gap(14).Children(func() {
			mz.Collapsible(c.Key("adv"), "Advanced options", &f.Advanced, mz.CollapsibleOptions{}, func() {
				mz.Checkbox(c.Key("verbose"), &f.Verbose, "Verbose logging", mz.CheckOptions{})
				mz.Checkbox(c.Key("beta"), &f.Beta, "Beta features", mz.CheckOptions{Description: "They may change without notice."})
			})
			mz.Separator(c, mz.SeparatorOptions{})
			mz.Accordion(c.Key("acc"), mz.AccordionOptions{Single: true}, func(a *mz.AccordionList) {
				a.Item("Billing", &f.Billing, func() {
					ui.Text(c, "Invoices go out on the first of the month, to the address on file.").Wrap()
				})
				a.ItemWith("Security", &f.Security, mz.CollapsibleOptions{Description: "Two-factor sign-in, sessions"}, func() {
					ui.Text(c, "Two-factor sign-in is on for every member of the team.").Wrap()
				})
				a.Item("Export", &f.Export, func() {
					ui.Text(c, "Exports are kept for 30 days, then deleted.").Wrap()
				})
			})
		})
	})
}

func (f *feedbackPage) toolbar(c *ui.Context, t *mz.Tokens) {
	mz.Card(c, mz.CardOptions{Title: "Toggle and toolbar", Subtitle: "Pressed into the surface while on.", Pad: true}, func() {
		ui.Column(c).Gap(14).Children(func() {
			ui.Row(c).Gap(12).AlignItems(ui.Center).Wrap().Children(func() {
				mz.ToggleGroup(c.Key("style"), "Text style", func() {
					mz.Toggle(c.Key("b"), &f.Bold, "Bold", mz.ToggleOptions{Icon: iconBold, IconOnly: true})
					mz.Toggle(c.Key("i"), &f.Italic, "Italic", mz.ToggleOptions{Icon: iconItalic, IconOnly: true})
					mz.Toggle(c.Key("u"), &f.Underline, "Underline", mz.ToggleOptions{Icon: iconUnderline, IconOnly: true})
					mz.Separator(c, mz.SeparatorOptions{Vertical: true})
					if mz.Toggle(c.Key("l"), &f.Left, "Align left", mz.ToggleOptions{Icon: iconAlignL, IconOnly: true}).Changed() && f.Left {
						f.Center = false
					}
					if mz.Toggle(c.Key("c"), &f.Center, "Align center", mz.ToggleOptions{Icon: iconAlignC, IconOnly: true}).Changed() && f.Center {
						f.Left = false
					}
				})
				mz.Toggle(c.Key("pin"), &f.Pin, "Pin", mz.ToggleOptions{Icon: iconPin})
				mz.Toggle(c.Key("pin-sm"), &f.Pin, "Pinned", mz.ToggleOptions{Size: mz.Sm})
			})
			ui.Text(c, "B / I / U and the alignments are one toolbar: the arrows move between them.").
				TextColor(t.Colors.Muted).FontSize(mz.FsSm).Wrap()
		})
	})
}

func (f *feedbackPage) loading(c *ui.Context, t *mz.Tokens) {
	cl := t.Colors
	mz.Card(c, mz.CardOptions{Title: "Links, separators and skeletons", Subtitle: "Placeholders shimmer until the content arrives.", Pad: true}, func() {
		ui.Column(c).Gap(16).Children(func() {
			ui.Row(c).Gap(4).Wrap().AlignItems(ui.Center).Children(func() {
				ui.Text(c, "The export is ready.").TextColor(cl.Text2).FontSize(mz.FsMd)
				mz.Link(c, "Read the changelog", "https://github.com/imhalid/mizan-go").FontSize(mz.FsMd)
			})
			mz.Separator(c, mz.SeparatorOptions{Label: "or"})
			ui.Row(c).Gap(12).AlignItems(ui.Start).Children(func() {
				if f.Loaded {
					mz.Avatar(c, "Maya Chen", mz.AvatarOptions{})
					ui.Column(c).Grow(1).MinWidth(0).Gap(4).Children(func() {
						ui.Text(c, "Maya Chen").FontWeight(600).FontSize(mz.FsMd)
						ui.Text(c, "Shared Q3-report.pdf with the team and left two comments on the retention chart.").
							TextColor(cl.Text2).FontSize(mz.FsMd).Wrap()
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
			if f.Loaded {
				label = "Show placeholders"
			}
			if mz.Button(c.Key("load"), label, mz.ButtonOptions{Size: mz.Sm}).AlignSelf(ui.Start).Clicked() {
				f.Loaded = !f.Loaded
			}
		})
	})
}
