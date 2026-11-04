package gallery

import (
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/ryu-go"
)

// workspaceTester opens the workspace page.
func workspaceTester() (*State, *ui.Tester) {
	s := New()
	s.Mode, s.Page = mz.Light, 6
	tt := ui.NewTester(s.View, 1180, 900)
	tt.Advance(400 * time.Millisecond)
	return s, tt
}

// dragEdge drags from x, y by dx, dy, a few points at a time.
func dragEdge(tt *ui.Tester, x, y, dx, dy float32) {
	tt.Press(x, y)
	for i := 1; i <= 8; i++ {
		f := float32(i) / 8
		tt.Move(x+dx*f, y+dy*f)
		tt.Advance(16 * time.Millisecond)
	}
	tt.Release(x+dx, y+dy)
	tt.Advance(400 * time.Millisecond)
}

// The edge between Files and the middle column moves with the pointer.
func TestDockEdge(t *testing.T) {
	s, tt := workspaceTester()
	root := s.Work.Layout.Windows[0].Root
	before := root.Sizes[0]
	r, _ := tt.Find("Editor")
	// the editor's kind button sits 4 DIPs into its area; the edge is just left of it
	x, y := r.X-4-2, r.Y+100
	dragEdge(tt, x, y, 80, 0)
	if root.Sizes[0] <= before+0.03 {
		t.Fatalf("Files' size %v, was %v: the edge did not move", root.Sizes[0], before)
	}
	if sum := root.Sizes[0] + root.Sizes[1] + root.Sizes[2]; sum < 0.999 || sum > 1.001 {
		t.Fatalf("sizes add up to %v", sum)
	}
	// dragged far, it stops where Files is as small as an area goes
	r, _ = tt.Find("Editor")
	dragEdge(tt, r.X-6, r.Y+100, -600, 0)
	if w := root.Sizes[0] * (1180 - 2*68); w < 80 {
		t.Fatalf("Files shrank to about %v DIPs, under the least size", w)
	}
	// a double click on it shares the room out
	r, _ = tt.Find("Editor")
	tt.ClickAt(r.X-6, r.Y+100)
	tt.ClickAt(r.X-6, r.Y+100)
	tt.Advance(400 * time.Millisecond)
	if d := root.Sizes[0] - 1.0/3; d > 0.001 || d < -0.001 {
		t.Fatalf("double click: Files' size %v, want a third", root.Sizes[0])
	}
}

// The horizontal edges of the middle and right columns are one line:
// dragging one moves the other.
func TestDockLinedEdges(t *testing.T) {
	s, tt := workspaceTester()
	root := s.Work.Layout.Windows[0].Root
	mid, right := root.Children[1], root.Children[2]
	r, _ := tt.Find("Console")
	// the console's title is in the middle of its 28 DIP header; the edge is
	// just above the header
	y := r.Y + r.H/2 - 14 - 2
	dragEdge(tt, r.X+40, y, 0, -60)
	if mid.Sizes[0] >= 0.66-0.03 {
		t.Fatalf("the middle column's edge did not move: %v", mid.Sizes)
	}
	if d := mid.Sizes[0] - right.Sizes[0]; d > 0.01 || d < -0.01 {
		t.Fatalf("the edges lined up did not move together: %v and %v", mid.Sizes, right.Sizes)
	}
}

// consoleEdge is a point on the edge over the console, and the x of the
// edge between the middle and right columns.
func consoleEdge(tt *ui.Tester) (x, y, cross float32) {
	r, _ := tt.Find("Console")
	o, _ := tt.Find("Properties")
	return r.X + 40, r.Y + r.H/2 - 14 - 2, o.X - 4 - 2
}

// Where the edges meet, a drag moves the vertical edge across and the
// lined-up edges along, at once.
func TestDockCrossing(t *testing.T) {
	s, tt := workspaceTester()
	root := s.Work.Layout.Windows[0].Root
	mid, right := root.Children[1], root.Children[2]
	before := root.Sizes[1]
	_, y, x := consoleEdge(tt)
	dragEdge(tt, x, y, 50, -60)
	if root.Sizes[1] <= before+0.02 {
		t.Fatalf("the vertical edge did not move: %v", root.Sizes)
	}
	if mid.Sizes[0] >= 0.66-0.03 {
		t.Fatalf("the horizontal edge did not move: %v", mid.Sizes)
	}
	if d := mid.Sizes[0] - right.Sizes[0]; d > 0.01 || d < -0.01 {
		t.Fatalf("the lined-up edges parted: %v and %v", mid.Sizes, right.Sizes)
	}
}

// editorArea is the editor's area, from its title.
func editorArea(tt *ui.Tester) ui.Rect {
	r, _ := tt.Find("Editor")
	c, _ := tt.Find("Console")
	o, _ := tt.Find("Outline")
	top := r.Y + r.H/2 - 14
	bottom := c.Y + c.H/2 - 14 - 4
	return ui.Rect{X: r.X - 4, Y: top, W: o.X - 4 - 4 - (r.X - 4), H: bottom - top}
}

// Dragged inward from its top left corner, the editor splits side by side,
// the new area on the corner's side, the line where the pointer let go.
func TestDockCornerSplit(t *testing.T) {
	s, tt := workspaceTester()
	a := editorArea(tt)
	dragEdge(tt, a.X+5, a.Y+5, 200, 12)
	mid := s.Work.Layout.Windows[0].Root.Children[1]
	row := mid.Children[0]
	if row.Dir != mz.DockRow || len(row.Children) != 2 {
		t.Fatalf("the editor did not split side by side: %+v", row)
	}
	if row.Children[0].Panes[0].Title != "report.go" || row.Children[1].Panes[0].Title != "report.go" {
		t.Fatalf("the areas don't show the editor: %+v %+v", row.Children[0].Panes[0], row.Children[1].Panes[0])
	}
	if w := row.Sizes[0] * (a.W - 4); w < 195 || w > 215 {
		t.Fatalf("the new area is %v wide, want about 205", w)
	}
	if !s.Work.Layout.CanUndo() {
		t.Fatal("the split can't be undone")
	}
}

// Dragged from its bottom right corner into the console, the editor takes
// the console in: the middle column folds away into the editor alone.
func TestDockCornerJoin(t *testing.T) {
	s, tt := workspaceTester()
	a := editorArea(tt)
	dragEdge(tt, a.X+a.W-5, a.Y+a.H-5, 0, 80)
	tt.Advance(600 * time.Millisecond)
	root := s.Work.Layout.Windows[0].Root
	if len(root.Children) != 3 || !root.Children[1].IsArea() || root.Children[1].Panes[0].Title != "report.go" {
		t.Fatalf("the console was not taken in: %+v", root.Children[1])
	}
	if tt.HasText("Console") {
		t.Fatal("the console still shows")
	}
	if err := tt.Click("Undo"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(600 * time.Millisecond)
	if !tt.HasText("Console") {
		t.Fatal("undo did not bring the console back")
	}
}

// The header's menu changes what an area shows; Undo brings it back.
func TestDockKindMenu(t *testing.T) {
	s, tt := workspaceTester()
	files := s.Work.Layout.Windows[0].Root.Children[0]
	if err := tt.Click("Files"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(300 * time.Millisecond)
	// the items' names are on the headers too: choose with the keys, Files,
	// Editor, Console
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	tt.Advance(300 * time.Millisecond)
	if files.Panes[0].Kind != "console" || files.Panes[0].Title != "Console" {
		t.Fatalf("the area shows %+v, want the console", files.Panes[0])
	}
	if err := tt.Click("Undo"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(300 * time.Millisecond)
	if p := s.Work.Layout.Windows[0].Root.Children[0].Panes[0]; p.Kind != "files" {
		t.Fatalf("undo: the area shows %+v", p)
	}
}

// Ctrl+Space maximizes the area under the pointer over the window, growing
// it, and brings it back.
func TestDockMaximize(t *testing.T) {
	s, tt := workspaceTester()
	w := s.Work.Layout.Windows[0]
	editor := w.Root.Children[1].Children[0]
	before, _ := tt.Find("Editor")
	files, _ := tt.Find("Files")
	tt.Move(before.X+100, before.Y+100)
	tt.Advance(50 * time.Millisecond)
	tt.Key(ui.Ctrl, ui.KeySpace)
	if w.Max != editor.ID {
		t.Fatalf("maximized %q, want the editor %q", w.Max, editor.ID)
	}
	tt.Advance(100 * time.Millisecond)
	mid, _ := tt.Find("report.go")
	tt.Advance(500 * time.Millisecond)
	after, _ := tt.Find("report.go")
	// maximized, the editor's header starts where the files' did
	if after.X >= mid.X || mid.X >= before.X+80 || after.X < files.X || after.X > files.X+90 {
		t.Fatalf("the editor's title went %v → %v → %v, want it to glide to the window's left (%v)", before.X, mid.X, after.X, files.X)
	}
	tt.Key(ui.Ctrl, ui.KeySpace)
	tt.Advance(600 * time.Millisecond)
	back, _ := tt.Find("Editor")
	if w.Max != "" || abs(back.X-before.X) > 0.5 {
		t.Fatalf("not brought back: max %q, at %v, was %v", w.Max, back.X, before.X)
	}
}

// Ctrl+Alt+arrows move the focus between areas; Ctrl+Space then maximizes
// the one focused when the pointer is elsewhere.
func TestDockFocusStep(t *testing.T) {
	s, tt := workspaceTester()
	w := s.Work.Layout.Windows[0]
	console := w.Root.Children[1].Children[1]
	r, _ := tt.Find("Editor")
	tt.ClickAt(r.X+200, r.Y+150)
	tt.Advance(50 * time.Millisecond)
	tt.Key(ui.Ctrl|ui.Alt, ui.KeyDown)
	tt.Advance(50 * time.Millisecond)
	tt.Move(2, 2)
	tt.Advance(50 * time.Millisecond)
	tt.Key(ui.Ctrl, ui.KeySpace)
	if w.Max != console.ID {
		t.Fatalf("maximized %q, want the console %q", w.Max, console.ID)
	}
}

func abs(v float32) float32 { return max(v, -v) }

// tabAt is a point on the title of the tab titled title, just before its
// close button: the kind's button can have the same name.
func tabAt(t *testing.T, tt *ui.Tester, title string) (x, y float32) {
	t.Helper()
	r, ok := tt.Find("Close " + title)
	if !ok {
		t.Fatalf("no tab %q", title)
	}
	return r.X - 12, r.Y + r.H/2
}

// dragTab drags the tab titled title to x, y, a few points at a time, and
// lets it go there.
func dragTab(t *testing.T, tt *ui.Tester, title string, x, y float32) {
	t.Helper()
	fx, fy := tabAt(t, tt, title)
	tt.Press(fx, fy)
	for i := 1; i <= 12; i++ {
		f := float32(i) / 12
		tt.Move(fx+(x-fx)*f, fy+(y-fy)*f)
		tt.Advance(16 * time.Millisecond)
	}
	tt.Advance(100 * time.Millisecond)
	tt.Release(x, y)
	tt.Advance(800 * time.Millisecond)
}

// A tab chosen shows its pane; Ctrl+PageDown goes to the next.
func TestDockTabChoose(t *testing.T) {
	s, tt := workspaceTester()
	editor := s.Work.Layout.Windows[0].Root.Children[1].Children[0]
	if err := tt.Click("theme.go"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(300 * time.Millisecond)
	if editor.Active != 2 {
		t.Fatalf("active %d, want theme.go (2)", editor.Active)
	}
	tt.Key(ui.Ctrl, ui.KeyPageDown)
	tt.Advance(300 * time.Millisecond)
	if editor.Active != 0 {
		t.Fatalf("Ctrl+PageDown: active %d, want report.go (0) after the last", editor.Active)
	}
}

// Dragged along its stack, a tab takes its new place there.
func TestDockTabReorder(t *testing.T) {
	s, tt := workspaceTester()
	editor := s.Work.Layout.Windows[0].Root.Children[1].Children[0]
	r, _ := tt.Find("report.go")
	dragTab(t, tt, "theme.go", r.X+4, r.Y+r.H/2)
	if got := paneTitles(editor); !slices.Equal(got, []string{"theme.go", "report.go", "layout.go"}) {
		t.Fatalf("the stack is %v", got)
	}
	if editor.Active != 0 {
		t.Fatalf("active %d, want the tab moved", editor.Active)
	}
}

// Dropped on another area's tabs, a tab joins its stack there; dropped on
// its right half, it docks in a new area on its right.
func TestDockTabDock(t *testing.T) {
	s, tt := workspaceTester()
	root := s.Work.Layout.Windows[0].Root
	console := root.Children[1].Children[1]
	o, _ := tt.Find("Output")
	dragTab(t, tt, "layout.go", o.X, o.Y+o.H/2)
	if got := paneTitles(console); !slices.Equal(got, []string{"Console", "layout.go", "Output"}) {
		t.Fatalf("the console's stack is %v", got)
	}
	f, _ := tt.Find("Files")
	files := root.Sizes[0]
	dragTab(t, tt, "theme.go", f.X+170, f.Y+300)
	// the root is a row: it takes the new area in, beside Files
	if len(root.Children) != 4 {
		t.Fatalf("Files did not split side by side: %d children", len(root.Children))
	}
	if got := paneTitles(root.Children[1]); !slices.Equal(got, []string{"theme.go"}) {
		t.Fatalf("the new area shows %v", got)
	}
	if d := root.Sizes[0] - root.Sizes[1]; d > 0.01 || d < -0.01 || root.Sizes[0]+root.Sizes[1]-files > 0.01 {
		t.Fatalf("Files was not shared in halves: %v, was %v", root.Sizes, files)
	}
	if !tt.HasText("theme.go") || !s.Work.Layout.CanUndo() {
		t.Fatal("theme.go is not shown, or can't be undone")
	}
}

// An area whose last tab leaves folds away, its neighbor taking its room;
// Undo brings it back.
func TestDockTabEmptyFolds(t *testing.T) {
	s, tt := workspaceTester()
	root := s.Work.Layout.Windows[0].Root
	p, _ := tt.Find("Properties")
	dragTab(t, tt, "Outline", p.X+120, p.Y+p.H/2)
	right := root.Children[2]
	if !right.IsArea() || !slices.Equal(paneTitles(right), []string{"Properties", "Outline"}) {
		t.Fatalf("the right column is %+v, want one area of Properties and Outline", right)
	}
	if err := tt.Click("Undo"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(600 * time.Millisecond)
	right = s.Work.Layout.Windows[0].Root.Children[2]
	if right.IsArea() || len(right.Children) != 2 {
		t.Fatalf("undo: the right column is %+v", right)
	}
}

// A tab's close button closes its pane; Undo brings it back.
func TestDockTabClose(t *testing.T) {
	s, tt := workspaceTester()
	console := s.Work.Layout.Windows[0].Root.Children[1].Children[1]
	if err := tt.Click("Close Console"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(400 * time.Millisecond)
	if got := paneTitles(console); !slices.Equal(got, []string{"Output"}) {
		t.Fatalf("the stack is %v", got)
	}
	// the Output pane is of the console's kind: its menu is named Console
	if _, ok := tt.Find("Close Console"); ok {
		t.Fatal("the console's tab still shows")
	}
	if err := tt.Click("Undo"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(400 * time.Millisecond)
	console = s.Work.Layout.Windows[0].Root.Children[1].Children[1]
	if got := paneTitles(console); !slices.Equal(got, []string{"Console", "Output"}) {
		t.Fatalf("undo: the stack is %v", got)
	}
}

func paneTitles(n *mz.Node) []string {
	var out []string
	for _, p := range n.Panes {
		out = append(out, p.Title)
	}
	return out
}

// The compass in the middle of the area under a tab dragged: its right
// target docks the tab in a new area on the area's right.
func TestDockCompass(t *testing.T) {
	s, tt := workspaceTester()
	mid := s.Work.Layout.Windows[0].Root.Children[1]
	a := editorArea(tt)
	// the compass's middle is the body's, under the header; its right
	// target one tile and a gap to the right
	cx, cy := a.X+a.W/2, a.Y+28+(a.H-28)/2
	dragTab(t, tt, "Output", cx+34, cy)
	row := mid.Children[0]
	if row.IsArea() || row.Dir != mz.DockRow || len(row.Children) != 2 {
		t.Fatalf("the editor did not split side by side: %+v", row)
	}
	if got := paneTitles(row.Children[1]); !slices.Equal(got, []string{"Output"}) {
		t.Fatalf("the new area on the right shows %v", got)
	}
}

// The target in the middle of the window's left edge docks the tab along
// the whole of that edge; the area it left folds away.
func TestDockEdgeTarget(t *testing.T) {
	s, tt := workspaceTester()
	f, _ := tt.Find("Files")
	// the Dock's top left is the Files area's, 4 DIPs out from its kind
	// button; the Dock is 560 tall
	x, y := f.X-4+12+15, f.Y-4+280
	dragTab(t, tt, "Outline", x, y)
	tt.Advance(600 * time.Millisecond)
	root := s.Work.Layout.Windows[0].Root
	if len(root.Children) != 4 || !slices.Equal(paneTitles(root.Children[0]), []string{"Outline"}) {
		t.Fatalf("Outline is not along the left edge: %d children, first %v", len(root.Children), paneTitles(root.Children[0]))
	}
	if right := root.Children[3]; !right.IsArea() || !slices.Equal(paneTitles(right), []string{"Properties"}) {
		t.Fatalf("the area Outline left did not fold away: %+v", right)
	}
	if w := root.Sizes[0]; w < 0.2 || w > 0.3 {
		t.Fatalf("Outline takes %v of the width, want a quarter", w)
	}
}

// Dragged near an edge whose end meets its end, along the same line, an
// edge settles onto it, lining them up.
func TestDockEdgeSnap(t *testing.T) {
	s, tt := workspaceTester()
	pane := func(title string) *mz.Node { return mz.DockArea(&mz.Pane{Kind: "console", Title: title}) }
	s.Work.Layout = mz.NewLayout(mz.DockSplit(mz.DockRow, nil,
		mz.DockSplit(mz.DockColumn, []float32{0.5, 0.5}, pane("A1"), pane("B1")),
		mz.DockSplit(mz.DockColumn, []float32{0.6, 0.4}, pane("C1"), pane("D1")),
	))
	tt.Frame()
	tt.Advance(300 * time.Millisecond)
	root := s.Work.Layout.Windows[0].Root
	left, right := root.Children[0], root.Children[1]
	a, _ := tt.Find("Close A1")
	b, _ := tt.Find("Close B1")
	c, _ := tt.Find("Close C1")
	d, _ := tt.Find("Close D1")
	// the edges are 2 DIPs over the headers' tops, less the 4 DIPs the close
	// buttons sit under the areas' tops
	y := b.Y - 6 - 2
	want := d.Y - 6 - 2
	// short of the right column's edge by 5
	dragEdge(tt, a.X, y, 0, want-y-5)
	if dd := left.Sizes[0] - right.Sizes[0]; dd > 0.002 || dd < -0.002 {
		t.Fatalf("the left edge did not settle onto the right one: %v and %v (from %v, to %v, c %v)", left.Sizes, right.Sizes, y, want, c.Y)
	}
}
