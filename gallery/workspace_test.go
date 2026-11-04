package gallery

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/mizan-go"
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
