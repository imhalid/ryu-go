package main

import (
	"time"

	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/ryu-go"
)

var comments = []string{
	"The retention chart needs the October numbers.",
	"Can we split the cohort by plan?",
	"Looks good to me; shipping it Friday.",
	"The export dropped two columns again.",
	"Moved the summary above the fold.",
	"Please double-check the totals in section 3.",
}

func init() {
	open, last := false, ""
	demo("dialog", func(c *ui.Context, t *mz.Tokens) {
		if mz.Button(c.Key("open"), "Delete report…", mz.ButtonOptions{Danger: true, Variant: mz.Soft}).Clicked() {
			open = true
		}
		if last != "" {
			status(c, t, last)
		}
		mz.Dialog(c.Key("dialog"), &open, mz.DialogOptions{
			Title:       "Delete the report?",
			Description: "Q3-report.pdf will be removed for everyone.",
			Footer: func() {
				if mz.Button(c.Key("cancel"), "Cancel", mz.ButtonOptions{}).Clicked() {
					open, last = false, "Canceled"
				}
				if mz.Button(c.Key("confirm"), "Delete report", mz.ButtonOptions{Danger: true}).Clicked() {
					open, last = false, "Deleted"
				}
			},
		}, func() {
			ui.Column(c).Gap(12).Children(func() {
				ui.Text(c, "The report and its comments cannot be restored once deleted. Drag the dialog by its header.").Wrap()
				for i, who := range []string{"Maya Chen", "Leo Park", "Sam Rivera", "Noah Kim", "Ivy Brooks", "Maya Chen"} {
					ui.Row(c.Key(i)).Gap(10).AlignItems(ui.Start).Children(func() {
						mz.Avatar(c, who, mz.AvatarOptions{Size: 28})
						ui.Column(c).Grow(1).MinWidth(0).Gap(2).Children(func() {
							ui.Text(c, who).FontSize(mz.FsSm).FontWeight(600).TextColor(t.Colors.Text)
							ui.Text(c, comments[i%len(comments)]).FontSize(mz.FsMd).Wrap()
						})
					})
				}
			})
		})
	})

	sheet, title, keep, layout := false, "", true, 0
	demo("sheet", func(c *ui.Context, t *mz.Tokens) {
		if mz.Button(c.Key("open"), "Report settings", mz.ButtonOptions{}).Clicked() {
			sheet = true
		}
		mz.Sheet(c.Key("sheet"), &sheet, mz.DialogOptions{
			Title:       "Report settings",
			Description: "Changes apply to the next export.",
			Footer: func() {
				if mz.Button(c.Key("discard"), "Discard", mz.ButtonOptions{}).Clicked() {
					sheet = false
				}
				if mz.Button(c.Key("save"), "Save", mz.ButtonOptions{Color: t.Colors.Accent, Variant: mz.Solid}).Clicked() {
					sheet = false
				}
			},
		}, func() {
			ui.Column(c).Gap(16).Children(func() {
				mz.Field(c, mz.FieldOptions{Label: "Title"}, func() {
					mz.Input(c.Key("title"), &title, mz.InputOptions{Label: "Title", Placeholder: "Untitled report"})
				})
				mz.Switch(c.Key("sync"), &keep, mz.SwitchOptions{Label: "Keep in sync", Description: "Refresh the data every hour."})
				mz.Segmented(c.Key("layout"), &layout, []string{"List", "Grid", "Board"}, mz.SegmentedOptions{Label: "Layout", Fill: true})
			})
		})
	})

	pop, mine, archived := false, true, false
	demo("popover", func(c *ui.Context, t *mz.Tokens) {
		b := mz.Button(c.Key("filter"), "Filter", mz.ButtonOptions{})
		if b.Clicked() {
			pop = !pop
		}
		mz.Popover(c.Key("filters"), b, &pop, func() {
			mz.Text(c, "Filters", mz.TLabel)
			mz.Checkbox(c.Key("mine"), &mine, "Only mine", mz.CheckOptions{})
			mz.Checkbox(c.Key("archived"), &archived, "Include archived", mz.CheckOptions{})
		})
	})

	menu, sort, action := false, "Name", "nothing yet"
	demo("menu", func(c *ui.Context, t *mz.Tokens) {
		ui.Row(c).Gap(mz.GapControl).Children(func() {
			mz.MenuButton(c.Key("sort"), "Sort", mz.ButtonOptions{}, func(m *mz.MenuList) {
				m.Label("Sort by")
				for _, by := range []string{"Name", "Date", "Size"} {
					if m.Item(by, mz.ItemOptions{Checkable: true, Checked: sort == by}) {
						sort = by
					}
				}
			})
			mb := mz.Button(c.Key("more"), "More", mz.ButtonOptions{Icon: iconMore, IconOnly: true})
			if mb.Clicked() {
				menu = !menu
			}
			mz.Menu(c.Key("moreMenu"), mb, &menu, func(m *mz.MenuList) {
				if m.Item("Rename", mz.ItemOptions{Icon: iconEdit, Shortcut: "⌘R"}) {
					action = "Rename"
				}
				if m.Item("Share", mz.ItemOptions{Icon: iconShare, Shortcut: "⌘S"}) {
					action = "Share"
				}
				m.Item("Duplicate", mz.ItemOptions{Disabled: true})
				m.Separator()
				if m.Item("Delete", mz.ItemOptions{Icon: iconTrash, Danger: true}) {
					action = "Delete"
				}
			})
		})
		status(c, t, "Sorted by "+sort+" · last action: "+action)
	})

	demo("tooltip", func(c *ui.Context, t *mz.Tokens) {
		// run the pointer along the toolbar: one tooltip slides from tool to tool
		ui.Row(c).Gap(4).Children(func() {
			for _, f := range []struct {
				name, tip string
				icon      *ui.SVG
			}{
				{"Bold", "Bold  ⌘B", iconBold},
				{"Italic", "Italic  ⌘I", iconItalic},
				{"Underline", "Underline  ⌘U", iconUnder},
				{"Align left", "Align left", iconAlignL},
				{"Align center", "Align center", iconAlignC},
				{"Share", "Share the report", iconShare},
			} {
				b := mz.Button(c.Key(f.name), f.name, mz.ButtonOptions{Icon: f.icon, IconOnly: true}).Description(f.tip)
				mz.Tooltip(c.Key(f.name+" tip"), b, f.tip)
			}
		})
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			for _, p := range []struct {
				name   string
				online bool
			}{{"Maya Chen", true}, {"Leo Park", false}, {"Sam Rivera", true}} {
				a := mz.Avatar(c.Key(p.name), p.name, mz.AvatarOptions{Online: p.online})
				tip := p.name + " · away"
				if p.online {
					tip = p.name + " · online"
				}
				mz.Tooltip(c.Key(p.name+" tip"), a.Focusable().Description(tip), tip)
			}
		})
	})

	var uploadAt time.Time
	demo("toast", func(c *ui.Context, t *mz.Tokens) {
		cl := t.Colors
		ui.Row(c).Gap(mz.GapControl).Wrap().Justify(ui.Center).Children(func() {
			if mz.Button(c.Key("info"), "Notify", mz.ButtonOptions{}).Clicked() {
				c.AddToast(ui.Toast{ID: "info", Title: "Report shared", Description: "Maya Chen can now view it."})
			}
			if mz.Button(c.Key("ok"), "Succeed", mz.ButtonOptions{Color: cl.Down}).Clicked() {
				c.AddToast(ui.Toast{ID: "ok", Title: "Export finished", Description: "Q3-report.pdf, 2.4 MB", Type: mz.ToastSuccess})
			}
			if mz.Button(c.Key("undo"), "Archive", mz.ButtonOptions{Color: cl.Warn}).Clicked() {
				c.AddToast(ui.Toast{ID: "undo", Title: "3 reports archived", Type: mz.ToastWarning, Action: "Undo", OnAction: func() {}})
			}
			if mz.Button(c.Key("err"), "Fail", mz.ButtonOptions{Danger: true, Variant: mz.Soft}).Clicked() {
				c.AddToast(ui.Toast{ID: "err", Title: "Upload failed", Description: "The connection was lost.", Type: mz.ToastError, Timeout: -1})
			}
			if mz.Button(c.Key("load"), "Upload", mz.ButtonOptions{}).Clicked() {
				c.AddToast(ui.Toast{ID: "upload", Title: "Uploading Q3-report.pdf", Description: "2.4 MB", Type: mz.ToastLoading, Timeout: -1})
				uploadAt = c.Now()
			}
		})
		if !uploadAt.IsZero() {
			// a stand-in for the upload finishing: the same toast turns into its result
			const took = 2200 * time.Millisecond
			if left := took - c.Now().Sub(uploadAt); left > 0 {
				c.After(left)
			} else {
				uploadAt = time.Time{}
				c.AddToast(ui.Toast{ID: "upload", Title: "Q3-report.pdf uploaded", Description: "Shared with the finance team.", Type: mz.ToastSuccess})
			}
		}
	})
}
