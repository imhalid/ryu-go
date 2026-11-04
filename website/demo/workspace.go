package main

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/ryu-go"
)

// The icons of the workspace's kinds.
var (
	iconFolder  = mz.Svg(16, 14, "M2.5 4.5A1.5 1.5 0 0 1 4 3h2.6l1.4 1.5h4A1.5 1.5 0 0 1 13.5 6v5.5A1.5 1.5 0 0 1 12 13H4a1.5 1.5 0 0 1-1.5-1.5z")
	iconCode    = mz.Svg(16, 14, "M5.5 4.5L2 8l3.5 3.5M10.5 4.5L14 8l-3.5 3.5")
	iconPrompt  = mz.Svg(16, 14, "M3 4.5L6.5 8 3 11.5M8 11.5h5")
	iconSliders = mz.Svg(16, 14, "M2.5 4.5h4M9.5 4.5h4M8 3v3M2.5 11.5h7M12.5 11.5h1M11 10v3")
)

// newWorkspace is the workspace's first layout: files, an editor over a
// console, and properties.
func newWorkspace() *mz.Layout {
	pane := func(kind, title string) *mz.Pane { return &mz.Pane{Kind: kind, Title: title} }
	return mz.NewLayout(mz.DockSplit(mz.DockRow, []float32{0.22, 0.54, 0.24},
		mz.DockArea(pane("files", "Files")),
		mz.DockSplit(mz.DockColumn, []float32{0.64, 0.36},
			mz.DockArea(pane("editor", "report.go"), pane("editor", "layout.go")),
			mz.DockArea(pane("console", "Console")),
		),
		mz.DockArea(pane("props", "Properties")),
	))
}

func init() {
	layout := newWorkspace()
	opacity, snap := 80.0, true
	demo("workspace", func(c *ui.Context, t *mz.Tokens) {
		cl := t.Colors
		kinds := []mz.DockKind{
			{Kind: "files", Title: "Files", Icon: iconFolder, Build: lines(cl, []string{"cmd/", "docs/", "gallery/", "accent.go", "controls.go", "dock.go", "forms.go", "table.go", "tokens.go"})},
			{Kind: "editor", Title: "Editor", Icon: iconCode, Build: source(cl)},
			{Kind: "console", Title: "Console", Icon: iconPrompt, Build: lines(cl, []string{"$ go test ./...", "ok   github.com/imhalid/ryu-go   0.4s", "$ go vet ./...", "$"})},
			{Kind: "props", Title: "Properties", Icon: iconSliders, Build: func(c *ui.Context, p *mz.Pane) {
				ui.Column(c).Padding(12).Gap(12).Children(func() {
					mz.Slider(c.Key("opacity"), &opacity, 0, 100, mz.SliderOptions{Label: "Opacity"})
					mz.Switch(c.Key("snap"), &snap, mz.SwitchOptions{Label: "Snap"})
				})
			}},
		}
		ui.Column(c).Gap(10).AlignSelf(ui.Stretch).Grow(1).MinHeight(0).Children(func() {
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				if mz.Button(c.Key("undo"), "Undo", mz.ButtonOptions{Size: mz.Sm, Disabled: !layout.CanUndo()}).Clicked() {
					layout.Undo()
				}
				if mz.Button(c.Key("redo"), "Redo", mz.ButtonOptions{Size: mz.Sm, Disabled: !layout.CanRedo()}).Clicked() {
					layout.Redo()
				}
				if mz.Button(c.Key("reset"), "Reset", mz.ButtonOptions{Size: mz.Sm}).Clicked() {
					layout = newWorkspace()
				}
				ui.Box(c).Grow(1)
				mz.Badge(c, "Experimental", mz.ToneWarn)
			})
			ui.Box(c).FillWidth().Grow(1).MinHeight(0).Children(func() {
				mz.Dock(c.Key("dock"), layout, mz.DockOptions{Kinds: kinds}).Fill()
			})
		})
	})
}

// lines is a pane kind showing rows of text, scrolled where the pane is.
func lines(cl mz.Colors, rows []string) func(*ui.Context, *mz.Pane) {
	return func(c *ui.Context, p *mz.Pane) {
		ui.Scroll(c).Fill().TrackScroll(&p.Scroll).Children(func() {
			ui.Column(c).Padding(8, 12).Gap(6).Children(func() {
				for _, r := range rows {
					ui.Text(c, r).FontSize(mz.FsSm).TextColor(cl.Text2).SingleLine()
				}
			})
		})
	}
}

// source is a pane kind showing Go, in the monospace font.
func source(cl mz.Colors) func(*ui.Context, *mz.Pane) {
	src := []string{
		"func (d *dockView) move(e dockEdge, by float32) {",
		"\tn := d.l.Find(e.split)",
		"\ta, b := n.Sizes[e.i], n.Sizes[e.i+1]",
		"\tleast := min(d.o.MinSize/e.length, (a+b)/2)",
		"\tna := min(max(a+by/e.length, least), a+b-least)",
		"\tn.Sizes[e.i], n.Sizes[e.i+1] = na, a+b-na",
		"}",
	}
	return func(c *ui.Context, p *mz.Pane) {
		ui.Scroll(c).Fill().TrackScroll(&p.Scroll).Children(func() {
			ui.Column(c).Padding(10, 14).Gap(4).Children(func() {
				for i, l := range src {
					ui.Row(c).Gap(14).Children(func() {
						ui.Text(c, fmt.Sprint(i+1)).FontSize(mz.FsSm).TextColor(cl.Muted).Font(mz.FontMono).Width(18)
						ui.Text(c, l).FontSize(mz.FsSm).TextColor(cl.Text).Font(mz.FontMono).SingleLine()
					})
				}
			})
		})
	}
}
