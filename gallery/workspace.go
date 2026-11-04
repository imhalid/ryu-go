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
	// Host shows the layout's windows in the app's windows, when the gallery
	// runs in one (cmd/gallery); OnChange runs when the layout changed.
	Host     mz.DockHost
	OnChange func()
}

// layout is the workspace's layout, made the first time.
func (w *workspacePage) layout() *mz.Layout {
	if w.Layout == nil {
		w.Layout = newWorkspace()
	}
	return w.Layout
}

// Layout returns the workspace's layout, for a host of its windows.
func (s *State) Layout() *mz.Layout { return s.Work.layout() }

// WindowView builds window id of the workspace's layout, torn off into a
// window of its own: a Dock of it, in the gallery's theme.
func (s *State) WindowView(id string) func(*ui.Context) {
	return func(c *ui.Context) {
		t := mz.Root(c, mz.Options{Mode: s.Mode, Accent: s.Accent, Size: steps[s.Density], Corners: corners[s.Round]})
		// the areas fill the window to its edges: their corners are its own
		ui.Box(c).Fill().Background(t.Colors.Bg).Children(func() {
			s.Work.dock(c, t, id).Fill()
		})
	}
}

// dock is the Dock of window id of the workspace.
func (w *workspacePage) dock(c *ui.Context, t *mz.Tokens, id string) ui.Element {
	return mz.Dock(c.Key("dock"), w.layout(), mz.DockOptions{Window: id, Kinds: w.kinds(t.Colors), Host: w.Host, OnChange: w.OnChange})
}

// kinds are what the workspace's areas show.
func (w *workspacePage) kinds(cl mz.Colors) []mz.DockKind {
	return []mz.DockKind{
		{Kind: "files", Title: "Files", Icon: iconFolder, Build: lines(cl, []string{"cmd/", "docs/", "gallery/", "accent.go", "controls.go", "dock.go", "dock_layout.go", "forms.go", "table.go", "tokens.go"})},
		{Kind: "editor", Title: "Editor", Icon: iconCode, Build: code(cl)},
		{Kind: "console", Title: "Console", Icon: iconPrompt, Build: lines(cl, []string{"$ go test ./...", "ok   github.com/imhalid/mizan-go   0.4s", "ok   github.com/imhalid/mizan-go/gallery   61.2s", "$ go vet ./...", "$"})},
		{Kind: "outline", Title: "Outline", Icon: iconList, Build: lines(cl, []string{"Layout", "DockWindow", "Node", "Pane", "Dock", "dockView.node", "dockView.edge", "dockView.lined", "dockView.move"})},
		{Kind: "props", Title: "Properties", Icon: iconSliders, Build: func(c *ui.Context, p *mz.Pane) {
			ui.Column(c).Padding(12).Gap(10).Children(func() {
				mz.Slider(c.Key("vol"), &w.Volume, 0, 100, mz.SliderOptions{Label: "Opacity"})
				mz.Switch(c.Key("snap"), new(bool), mz.SwitchOptions{Label: "Snap"})
			})
		}},
	}
}

// The icons of the workspace's kinds.
var (
	iconFolder  = mz.Svg(16, 14, "M2.5 4.5A1.5 1.5 0 0 1 4 3h2.6l1.4 1.5h4A1.5 1.5 0 0 1 13.5 6v5.5A1.5 1.5 0 0 1 12 13H4a1.5 1.5 0 0 1-1.5-1.5z")
	iconCode    = mz.Svg(16, 14, "M5.5 4.5L2 8l3.5 3.5M10.5 4.5L14 8l-3.5 3.5")
	iconPrompt  = mz.Svg(16, 14, "M3 4.5L6.5 8 3 11.5M8 11.5h5")
	iconList    = mz.Svg(16, 14, "M2.5 4h.01M2.5 8h.01M2.5 12h.01M5.5 4h8M5.5 8h8M5.5 12h8")
	iconSliders = mz.Svg(16, 14, "M2.5 4.5h4M9.5 4.5h4M8 3v3M2.5 11.5h7M12.5 11.5h1M11 10v3")
)

func newWorkspace() *mz.Layout {
	pane := func(kind, title string) *mz.Pane { return &mz.Pane{Kind: kind, Title: title} }
	return mz.NewLayout(mz.DockSplit(mz.DockRow, []float32{0.2, 0.56, 0.24},
		mz.DockArea(pane("files", "Files")),
		mz.DockSplit(mz.DockColumn, []float32{0.66, 0.34},
			mz.DockArea(pane("editor", "report.go"), pane("editor", "layout.go"), pane("editor", "theme.go")),
			mz.DockArea(pane("console", "Console"), pane("console", "Output")),
		),
		// its edge lines up with the middle column's: they move as one
		mz.DockSplit(mz.DockColumn, []float32{0.66, 0.34},
			mz.DockArea(pane("outline", "Outline")),
			mz.DockArea(pane("props", "Properties")),
		),
	))
}

func (w *workspacePage) view(c *ui.Context, t *mz.Tokens) {
	w.layout()
	cl := t.Colors
	mz.Card(c, mz.CardOptions{Title: "Workspace", Subtitle: "Drag the gaps between areas; the edges lined up with one move along (Option: it alone). Where edges meet, drag both ways. A double click shares the room out. Drag an area's corner inward to split it, or into a neighbor to join them. Drag a tab within its stack, onto another area's tabs or middle, or onto one of its sides to dock it there.", Pad: true}, func() {
		ui.Column(c).Gap(12).Children(func() {
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				if mz.Button(c.Key("save"), "Save layout", mz.ButtonOptions{Size: mz.Sm}).Clicked() {
					w.Saved, _ = w.Layout.MarshalJSON()
				}
				if mz.Button(c.Key("restore"), "Restore", mz.ButtonOptions{Size: mz.Sm, Disabled: w.Saved == nil}).Clicked() {
					if l, err := mz.ParseLayout(w.Saved); err == nil {
						w.Layout = l
						w.changed()
					}
				}
				if mz.Button(c.Key("undo"), "Undo", mz.ButtonOptions{Size: mz.Sm, Disabled: !w.Layout.CanUndo()}).Clicked() {
					w.Layout.Undo()
					w.changed()
				}
				if mz.Button(c.Key("redo"), "Redo", mz.ButtonOptions{Size: mz.Sm, Disabled: !w.Layout.CanRedo()}).Clicked() {
					w.Layout.Redo()
					w.changed()
				}
				if mz.Button(c.Key("reset"), "Reset", mz.ButtonOptions{Size: mz.Sm}).Clicked() {
					w.Layout = newWorkspace()
					w.changed()
				}
				if w.Saved != nil {
					ui.Text(c, fmt.Sprintf("Saved, %d bytes", len(w.Saved))).FontSize(mz.FsSm).TextColor(cl.Muted)
				}
			})
			ui.Box(c).FillWidth().Height(560).Children(func() {
				w.dock(c, t, "main").Fill()
			})
		})
	})
}

func (w *workspacePage) changed() {
	if w.OnChange != nil {
		w.OnChange()
	}
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
