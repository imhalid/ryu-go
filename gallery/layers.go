package gallery

import (
	"time"

	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/mizan-go"
)

var (
	more  = mz.Svg(16, 15, "M3.5 8h.01M8 8h.01M12.5 8h.01")
	share = mz.Svg(16, 15, "M8 10V2.5M5 5.5l3-3 3 3M3 9.5v3A1.5 1.5 0 0 0 4.5 14h7a1.5 1.5 0 0 0 1.5-1.5v-3")
	edit  = mz.Svg(16, 15, "M10.5 2.5l3 3-8 8h-3v-3z")
	trash = mz.Svg(16, 15, "M2.5 4.5h11M6 4.5V3h4v1.5M4 4.5l.7 9h6.6l.7-9")
	info  = mz.Svg(16, 14, "M8 14.5a6.5 6.5 0 1 0 0-13 6.5 6.5 0 0 0 0 13z", "M8 7.5v3.5M8 5h.01")
)

// The formatting toolbar's tools: their icons and what their tooltips say.
var formatTools = []struct {
	name, tip string
	icon      *ui.SVG
}{
	{"Bold", "Bold  ⌘B", mz.Svg(16, 15, "M4.5 2.5h4.25a2.75 2.75 0 0 1 0 5.5H4.5zM4.5 8h5a2.75 2.75 0 0 1 0 5.5h-5z")},
	{"Italic", "Italic  ⌘I", mz.Svg(16, 15, "M10 2.5H6.5M9.5 13.5H6M9.25 2.5l-2.5 11")},
	{"Underline", "Underline  ⌘U", mz.Svg(16, 15, "M4.5 2.5v5a3.5 3.5 0 0 0 7 0v-5M3.5 14h9")},
	{"Strikethrough", "Strikethrough  ⇧⌘X", mz.Svg(16, 15, "M3 8h10M11 4.5c-.5-1.2-1.7-2-3.2-2-1.9 0-3.3 1-3.3 2.6M5 11.5c.5 1.2 1.7 2 3.2 2 1.9 0 3.3-1 3.3-2.6")},
	{"Link", "Add a link  ⌘K", mz.Svg(16, 15, "M6.5 9.5l3-3M7 4.5l1-1a2.8 2.8 0 0 1 4 4l-1 1M9 11.5l-1 1a2.8 2.8 0 0 1-4-4l1-1")},
	{"Bulleted list", "Bulleted list", mz.Svg(16, 15, "M6 4h7.5M6 8h7.5M6 12h7.5M2.5 4h.01M2.5 8h.01M2.5 12h.01")},
}

var comments = []string{
	"The retention chart needs the October numbers.",
	"Can we split the cohort by plan?",
	"Looks good to me; shipping it Friday.",
	"The export dropped two columns again.",
	"Moved the summary above the fold.",
	"Please double-check the totals in section 3.",
}

// layers is the third page: dialogs, sheets, popovers, menus, tooltips and
// toasts.
func (s *State) layers(c *ui.Context, t *mz.Tokens) {
	cl := t.Colors
	ui.Row(c).Gap(mz.GapSection).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Grow(1).Basis(0).Gap(mz.GapSection).Children(func() {
			mz.Card(c, mz.CardOptions{Title: "Dialog and sheet", Subtitle: "Draggable panels that spring in over a soft scrim.", Pad: true}, func() {
				ui.Row(c).Gap(mz.GapControl).Children(func() {
					if mz.Button(c.Key("openDialog"), "Open dialog", mz.ButtonOptions{}).Clicked() {
						s.Dialog = true
					}
					if mz.Button(c.Key("openSheet"), "Open sheet", mz.ButtonOptions{}).Clicked() {
						s.Sheet = true
					}
				})
			})
			mz.Card(c, mz.CardOptions{Title: "Tabs", Subtitle: "The underline slides; a soft fill follows the pointer.", Pad: true}, func() {
				mz.Tabs(c.Key("tabs"), &s.Tab, []mz.Tab{{Label: "Overview"}, {Label: "Activity", Badge: "12"}, {Label: "Members"}, {Label: "Settings"}},
					mz.TabsOptions{Label: "Report"})
				mz.TabPanel(c, s.Tab, func() {
					mz.Text(c, []string{
						"Weekly active users rose 4.2% after the new onboarding.",
						"12 comments and 3 exports since Monday.",
						"Maya Chen, Leo Park and 4 others can edit.",
						"Exports, retention and sharing.",
					}[s.Tab], mz.TBody).TextColor(cl.Text2)
				})
			})
			mz.Card(c, mz.CardOptions{Title: "Toast", Subtitle: "Notices in the window's corner, tinted by their tone.", Pad: true}, func() {
				ui.Row(c).Gap(mz.GapControl).Wrap().Children(func() {
					if mz.Button(c.Key("toastInfo"), "Notify", mz.ButtonOptions{}).Clicked() {
						c.AddToast(ui.Toast{ID: "info", Title: "Report shared", Description: "Maya Chen can now view it."})
					}
					if mz.Button(c.Key("toastOK"), "Succeed", mz.ButtonOptions{Color: cl.Down}).Clicked() {
						c.AddToast(ui.Toast{ID: "ok", Title: "Export finished", Description: "Q3-report.pdf, 2.4 MB", Type: mz.ToastSuccess})
					}
					if mz.Button(c.Key("toastUndo"), "Archive", mz.ButtonOptions{Color: cl.Warn}).Clicked() {
						c.AddToast(ui.Toast{ID: "undo", Title: "3 reports archived", Type: mz.ToastWarning, Action: "Undo",
							OnAction: func() { s.Last = "Undo" }})
					}
					if mz.Button(c.Key("toastErr"), "Fail", mz.ButtonOptions{Danger: true, Variant: mz.Soft}).Clicked() {
						c.AddToast(ui.Toast{ID: "err", Title: "Upload failed", Description: "The connection was lost.", Type: mz.ToastError, Timeout: -1})
					}
					if mz.Button(c.Key("toastLoad"), "Upload", mz.ButtonOptions{}).Clicked() {
						c.AddToast(ui.Toast{ID: "upload", Title: "Uploading Q3-report.pdf", Description: "2.4 MB", Type: mz.ToastLoading, Timeout: -1})
						s.uploadAt = c.Now()
					}
					if !s.uploadAt.IsZero() {
						// a stand-in for the upload finishing: the same toast turns
						// into its result
						const took = 2200 * time.Millisecond
						if left := took - c.Now().Sub(s.uploadAt); left > 0 {
							c.After(left)
						} else {
							s.uploadAt = time.Time{}
							c.AddToast(ui.Toast{ID: "upload", Title: "Q3-report.pdf uploaded", Description: "Shared with the finance team.", Type: mz.ToastSuccess})
						}
					}
				})
			})
		})
		ui.Column(c).Grow(1).Basis(0).Gap(mz.GapSection).Children(func() {
			mz.Card(c, mz.CardOptions{Title: "Popover, menu and tooltip", Subtitle: "Beside their anchor, on the pop surface.", Pad: true}, func() {
				ui.Column(c).Gap(14).Children(func() {
					ui.Row(c).Gap(mz.GapControl).AlignItems(ui.Center).Children(func() {
						b := mz.Button(c.Key("filter"), "Filter", mz.ButtonOptions{})
						if b.Clicked() {
							s.Pop = !s.Pop
						}
						mz.Popover(c.Key("filters"), b, &s.Pop, func() {
							mz.Text(c, "Filters", mz.TLabel)
							mz.Checkbox(c.Key("f1"), &s.Mail, "Only mine", mz.CheckOptions{})
							mz.Checkbox(c.Key("f2"), &s.Push, "Include archived", mz.CheckOptions{})
						})
						mz.MenuButton(c.Key("sort"), "Sort", mz.ButtonOptions{}, func(m *mz.MenuList) {
							m.Label("Sort by")
							for _, by := range []string{"Name", "Date", "Size"} {
								if m.Item(by, mz.ItemOptions{Checkable: true, Checked: or(s.Sort, "Name") == by}) {
									s.Sort = by
								}
							}
						})
						mb := mz.Button(c.Key("more"), "More", mz.ButtonOptions{Icon: more, IconOnly: true})
						if mb.Clicked() {
							s.Menu = !s.Menu
						}
						mz.Menu(c.Key("moreMenu"), mb, &s.Menu, func(m *mz.MenuList) {
							if m.Item("Rename", mz.ItemOptions{Icon: edit, Shortcut: "⌘R"}) {
								s.Last = "Rename"
							}
							if m.Item("Share", mz.ItemOptions{Icon: share, Shortcut: "⌘S"}) {
								s.Last = "Share"
							}
							m.Item("Duplicate", mz.ItemOptions{Disabled: true})
							m.Separator()
							if m.Item("Delete", mz.ItemOptions{Icon: trash, Danger: true}) {
								s.Last = "Delete"
							}
						})
						tb := mz.Button(c.Key("share"), "Share link", mz.ButtonOptions{Icon: share, IconOnly: true}).Description("Share the report")
						mz.Tooltip(c.Key("shareTip"), tb, "Share the report")
						// from one tooltip to the next, it slides
						cb := mz.CopyButton(c.Key("copy"), "https://mizan.example/q3-report", mz.CopyOptions{Label: "Copy link"}).Description("Copy the link")
						mz.Tooltip(c.Key("copyTip"), cb, "Copy the link")
					})
					last := s.Last
					if last == "" {
						last = "nothing yet"
					}
					mz.Text(c, "Last action: "+last, mz.TCaption).TextColor(cl.Muted)
				})
			})
			mz.Card(c, mz.CardOptions{Title: "Tooltip", Subtitle: "On hover or keyboard focus; from one to the next, it slides.", Pad: true}, func() {
				ui.Column(c).Gap(16).Children(func() {
					// a toolbar: run the pointer along it and one tooltip
					// slides from tool to tool
					ui.Row(c).Gap(4).Children(func() {
						for _, f := range formatTools {
							b := mz.Button(c.Key(f.name), f.name, mz.ButtonOptions{Icon: f.icon, IconOnly: true}).Description(f.tip)
							if b.Clicked() {
								s.Last = f.name
							}
							mz.Tooltip(c.Key(f.name+" tip"), b, f.tip)
						}
					})
					// a longer one wraps, up to 260 wide
					ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
						mz.Text(c, "Retention", mz.TLabel)
						hint := ui.Box(c.Key("retentionHint")).Focusable().Role(ui.RoleImage).Label("About retention").
							Description("The share of users who came back within 30 days of signing up.").
							TextColor(cl.Muted).Children(func() { ui.Icon(c, info) })
						mz.Tooltip(c.Key("retentionTip"), hint, "The share of users who came back within 30 days of signing up, by the month they signed up in.")
						mz.Text(c, "41.8%", mz.TLabel).TextColor(cl.Text2)
					})
					// on people: their names, and whether they're here
					ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
						for _, p := range []struct {
							name   string
							online bool
						}{{"Maya Chen", true}, {"Leo Park", false}, {"Sam Rivera", true}, {"Ivy Brooks", false}} {
							a := mz.Avatar(c.Key(p.name), p.name, mz.AvatarOptions{Online: p.online})
							tip := p.name + " · away"
							if p.online {
								tip = p.name + " · online"
							}
							mz.Tooltip(c.Key(p.name+" tip"), a.Focusable().Description(tip), tip)
						}
						mz.Text(c, "can edit", mz.TCaption).TextColor(cl.Muted)
					})
				})
			})
		})
	})
	mz.Dialog(c.Key("dialog"), &s.Dialog, mz.DialogOptions{
		Title:       "Delete the report?",
		Description: "Q3-report.pdf will be removed for everyone.",
		Footer: func() {
			if mz.Button(c.Key("cancel"), "Cancel", mz.ButtonOptions{}).Clicked() {
				s.Dialog = false
			}
			if mz.Button(c.Key("confirm"), "Delete report", mz.ButtonOptions{Danger: true}).Clicked() {
				s.Dialog, s.Last = false, "Deleted"
			}
		},
	}, func() {
		ui.Column(c).Gap(12).Children(func() {
			ui.Text(c, "The report and its 12 comments cannot be restored once deleted. Drag the dialog by its header; scroll the comments to see them dissolve into its edges.").Wrap()
			for i, who := range []string{"Maya Chen", "Leo Park", "Sam Rivera", "Noah Kim", "Ivy Brooks", "Maya Chen",
				"Leo Park", "Sam Rivera", "Noah Kim", "Ivy Brooks", "Maya Chen", "Leo Park"} {
				ui.Row(c.Key(i)).Gap(10).AlignItems(ui.Start).Children(func() {
					mz.Avatar(c, who, mz.AvatarOptions{Size: 28})
					ui.Column(c).Grow(1).MinWidth(0).Gap(2).Children(func() {
						ui.Text(c, who).FontSize(mz.FsSm).FontWeight(600).TextColor(cl.Text)
						ui.Text(c, comments[i%len(comments)]).FontSize(mz.FsMd).Wrap()
					})
				})
			}
		})
	})
	mz.Sheet(c.Key("sheet"), &s.Sheet, mz.DialogOptions{
		Title:       "Report settings",
		Description: "Changes apply to the next export.",
		Footer: func() {
			if mz.Button(c.Key("sheetCancel"), "Discard", mz.ButtonOptions{}).Clicked() {
				s.Sheet = false
			}
			if mz.Button(c.Key("sheetSave"), "Save", mz.ButtonOptions{Color: cl.Accent, Variant: mz.Solid}).Clicked() {
				s.Sheet, s.Last = false, "Saved"
			}
		},
	}, func() {
		ui.Column(c).Gap(16).Children(func() {
			mz.Field(c, mz.FieldOptions{Label: "Title"}, func() {
				mz.Input(c.Key("title"), &s.Query, mz.InputOptions{Label: "Title", Placeholder: "Untitled report"})
			})
			mz.Switch(c.Key("sheetSync"), &s.Sync, mz.SwitchOptions{Label: "Keep in sync", Description: "Refresh the data every hour."})
			mz.Segmented(c.Key("sheetView"), &s.ViewMode, []string{"List", "Grid", "Board"}, mz.SegmentedOptions{Label: "Layout", Fill: true})
		})
	})
}

func or(s, d string) string {
	if s == "" {
		return d
	}
	return s
}
