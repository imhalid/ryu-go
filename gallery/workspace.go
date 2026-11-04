package gallery

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/mizan-go"
)

// workspacePage is the state of the workspace page: a Dock's layout, and
// the last one saved.
type workspacePage struct {
	Layout *mz.Layout
	Saved  []byte
	Volume float64
}

func newWorkspace() *mz.Layout {
	pane := func(kind, title string) *mz.Pane { return &mz.Pane{Kind: kind, Title: title} }
	return mz.NewLayout(mz.DockSplit(mz.DockRow, []float32{0.2, 0.56, 0.24},
		mz.DockArea(pane("files", "Files")),
		mz.DockSplit(mz.DockColumn, []float32{0.66, 0.34},
			mz.DockArea(pane("editor", "report.go")),
			mz.DockArea(pane("console", "Console")),
		),
		// its edge lines up with the middle column's: they move as one
		mz.DockSplit(mz.DockColumn, []float32{0.66, 0.34},
			mz.DockArea(pane("outline", "Outline")),
			mz.DockArea(pane("props", "Properties")),
		),
	))
}

func (w *workspacePage) view(c *ui.Context, t *mz.Tokens) {
	if w.Layout == nil {
		w.Layout = newWorkspace()
	}
	cl := t.Colors
	mz.Card(c, mz.CardOptions{Title: "Workspace", Subtitle: "Drag the gaps between areas; the edges lined up with one move along (Option: it alone). Where edges meet, drag both ways. A double click shares the room out. Drag an area's corner inward to split it, or into a neighbor to join them.", Pad: true}, func() {
		ui.Column(c).Gap(12).Children(func() {
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				if mz.Button(c.Key("save"), "Save layout", mz.ButtonOptions{Size: mz.Sm}).Clicked() {
					w.Saved, _ = w.Layout.MarshalJSON()
				}
				if mz.Button(c.Key("restore"), "Restore", mz.ButtonOptions{Size: mz.Sm, Disabled: w.Saved == nil}).Clicked() {
					if l, err := mz.ParseLayout(w.Saved); err == nil {
						w.Layout = l
					}
				}
				if mz.Button(c.Key("undo"), "Undo", mz.ButtonOptions{Size: mz.Sm, Disabled: !w.Layout.CanUndo()}).Clicked() {
					w.Layout.Undo()
				}
				if mz.Button(c.Key("reset"), "Reset", mz.ButtonOptions{Size: mz.Sm}).Clicked() {
					w.Layout = newWorkspace()
				}
				if w.Saved != nil {
					ui.Text(c, fmt.Sprintf("Saved, %d bytes", len(w.Saved))).FontSize(mz.FsSm).TextColor(cl.Muted)
				}
			})
			ui.Box(c).FillWidth().Height(560).Children(func() {
				mz.Dock(c.Key("dock"), w.Layout, mz.DockOptions{Kinds: []mz.DockKind{
					{Kind: "files", Title: "Files", Build: lines(cl, []string{"cmd/", "docs/", "gallery/", "accent.go", "controls.go", "dock.go", "dock_layout.go", "forms.go", "table.go", "tokens.go"})},
					{Kind: "editor", Title: "Editor", Build: code(cl)},
					{Kind: "console", Title: "Console", Build: lines(cl, []string{"$ go test ./...", "ok   github.com/imhalid/mizan-go   0.4s", "ok   github.com/imhalid/mizan-go/gallery   61.2s", "$ go vet ./...", "$"})},
					{Kind: "outline", Title: "Outline", Build: lines(cl, []string{"Layout", "DockWindow", "Node", "Pane", "Dock", "dockView.node", "dockView.edge", "dockView.lined", "dockView.move"})},
					{Kind: "props", Title: "Properties", Build: func(c *ui.Context, p *mz.Pane) {
						ui.Column(c).Padding(12).Gap(10).Children(func() {
							mz.Slider(c.Key("vol"), &w.Volume, 0, 100, mz.SliderOptions{Label: "Opacity"})
							mz.Switch(c.Key("snap"), new(bool), mz.SwitchOptions{Label: "Snap"})
						})
					}},
				}}).Fill()
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

// code is a pane kind showing source, in the monospace font.
func code(cl mz.Colors) func(*ui.Context, *mz.Pane) {
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
