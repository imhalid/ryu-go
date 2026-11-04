package mizan

import (
	"bytes"
	"testing"
)

func sampleLayout() *Layout {
	return NewLayout(DockSplit(DockRow, []float32{0.25, 0.75},
		DockArea(&Pane{Kind: "files", Title: "Files"}),
		DockSplit(DockColumn, nil,
			DockArea(&Pane{Kind: "editor", Title: "a.go"}, &Pane{Kind: "editor", Title: "b.go"}),
			DockArea(&Pane{Kind: "made-up", Title: "Unknown"}),
		),
	))
}

// A layout saved and read back is the same layout, scroll and state too.
func TestLayoutRoundTrip(t *testing.T) {
	l := sampleLayout()
	p := l.Windows[0].Root.Children[0].Panes[0]
	p.Scroll.Y = 240
	p.State = map[string]string{"selected": "dock.go"}
	a, err := l.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	back, err := ParseLayout(a)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := back.MarshalJSON()
	if !bytes.Equal(a, b) {
		t.Fatalf("round trip changed the layout:\n%s\n%s", a, b)
	}
	q := back.Windows[0].Root.Children[0].Panes[0]
	if q.Scroll.Y != 240 || q.State["selected"] != "dock.go" {
		t.Fatalf("pane state lost: scroll %v, state %v", q.Scroll.Y, q.State)
	}
	if back.Find(l.Windows[0].Root.Children[1].ID) == nil {
		t.Fatal("a node's ID did not survive")
	}
}

// ParseLayout mends what it can: sizes, IDs, an active pane out of range;
// a kind it doesn't know is kept.
func TestParseLayoutMends(t *testing.T) {
	l, err := ParseLayout([]byte(`{"windows":[{"id":"main","root":{"dir":"column","sizes":[3],"children":[
		{"panes":[{"kind":"made-up","title":"X"}],"active":5},
		{"panes":[{"id":"p1","kind":"files","title":"Files"}]}]}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	root := l.Windows[0].Root
	if root.Dir != DockColumn || len(root.Sizes) != 2 || root.Sizes[0] != 0.5 {
		t.Fatalf("sizes not mended: %v", root.Sizes)
	}
	a := root.Children[0]
	if a.ID == "" || a.Panes[0].ID == "" || a.Active != 0 || a.Panes[0].Kind != "made-up" {
		t.Fatalf("area not mended: %+v %+v", a, a.Panes[0])
	}
	if root.Children[1].Panes[0].ID != "p1" {
		t.Fatal("an ID given was replaced")
	}
	if _, err := ParseLayout([]byte(`{"windows":[]}`)); err == nil {
		t.Fatal("a layout with no window parsed")
	}
}

// A split in a split of the other direction makes a new split; one in a
// split of its direction joins it, the others keeping their lengths. A
// join takes the room back, folds a split of one child away, and Undo
// brings everything back.
func TestSplitJoinUndo(t *testing.T) {
	l := sampleLayout()
	root := l.Windows[0].Root // row [files 0.25 | column [a.go, unknown]]
	files := root.Children[0]
	before, _ := l.MarshalJSON()
	// files is 250 of a 1000 DIP room: split it side by side, the new area first
	made := l.splitArea(files.ID, DockRow, true, 150, 96, 1000)
	if made == nil || len(root.Children) != 3 || root.Children[0] != made || root.Children[1] != files {
		t.Fatalf("split side by side: %+v", root.Children)
	}
	if d := root.Sizes[2]*(1000-dockGap) - 750; d > 0.01 || d < -0.01 {
		t.Fatalf("the other child's length changed: %v", root.Sizes)
	}
	if made.Panes[0].Kind != "files" || made.Panes[0].ID == files.Panes[0].ID {
		t.Fatalf("the new area's pane: %+v", made.Panes[0])
	}
	// stacked: files becomes a column of itself and a new area
	low := l.splitArea(files.ID, DockColumn, false, 300, 200, 0)
	col := root.Children[1]
	if col.Dir != DockColumn || col.Children[0] != files || col.Children[1] != low || col.Sizes[0] != 0.6 {
		t.Fatalf("split stacked: %+v", col)
	}
	// joined back: the column of one folds away, files takes its place
	if !l.joinArea(files.ID, low.ID) || root.Children[1] != files {
		t.Fatalf("join: %+v", root.Children)
	}
	if l.joinArea(files.ID, root.Children[2].Children[0].ID) {
		t.Fatal("joined areas that don't share a whole edge")
	}
	for l.Undo() {
	}
	after, _ := l.MarshalJSON()
	if !bytes.Equal(before, after) {
		t.Fatalf("undo did not bring the layout back:\n%s\n%s", before, after)
	}
}

// Panes move between stacks, dock beside an area and close; an area left
// empty goes, its neighbor taking its room; Undo goes back step by step.
func TestLayoutPanes(t *testing.T) {
	l := sampleLayout()
	root := l.Windows[0].Root
	files, col := root.Children[0], root.Children[1]
	editor, other := col.Children[0], col.Children[1]
	b := editor.Panes[1]
	editor.Active = 1

	l.remember()
	if !l.movePane(b.ID, files.ID, 0) || len(editor.Panes) != 1 || files.Panes[0] != b || files.Active != 0 || editor.Active != 0 {
		t.Fatalf("move: files %v active %d, editor %v active %d", files.Panes, files.Active, editor.Panes, editor.Active)
	}
	u := other.Panes[0]
	l.remember()
	n := l.dockPane(u.ID, editor.ID, DockRow, false, 100, 100, 0)
	if n == nil || n.Panes[0] != u || len(other.Panes) != 0 {
		t.Fatalf("dock: new area %+v, the area left %+v", n, other)
	}
	if !l.removeArea(other.ID) || l.Find(other.ID) != nil {
		t.Fatal("the empty area did not go")
	}
	// the column is left with the editor's row alone, which tidy gives to the
	// root, a row too: files, the editor, the new area
	if r := l.Windows[0].Root; len(r.Children) != 3 || r.Children[1] != editor || r.Children[2] != n {
		t.Fatalf("after the empty area went: %+v", r.Children)
	}
	if !l.closePane(b.ID) || len(files.Panes) != 1 {
		t.Fatalf("close: files %v", files.Panes)
	}
	for range 3 {
		if !l.Undo() {
			t.Fatal("nothing to undo")
		}
	}
	back := l.Windows[0].Root.Children[1].Children[0]
	if len(back.Panes) != 2 || back.Panes[1].Title != "b.go" {
		t.Fatalf("undo: the editor holds %v", back.Panes)
	}
}

// Cut where the columns' edges line up, a row of two columns becomes a
// column of two rows: the layout shows the same, and each part of the
// edge between the columns can move alone.
func TestLayoutCutEdge(t *testing.T) {
	pane := func(s string) *Node { return DockArea(&Pane{Kind: "k", Title: s}) }
	l := NewLayout(DockSplit(DockRow, []float32{0.6, 0.4},
		DockSplit(DockColumn, []float32{0.5, 0.5}, pane("editor"), pane("output")),
		DockSplit(DockColumn, []float32{0.5, 0.5}, pane("outline"), pane("files")),
	))
	root := l.Windows[0].Root
	row := l.cutEdge(root.ID, 0, [][2]int{{0, 0}}, 1, 556)
	root = l.Windows[0].Root
	if row == nil || root.Dir != DockColumn || len(root.Children) != 2 {
		t.Fatalf("not a column of two rows: %+v", root)
	}
	if row != root.Children[1] || row.Dir != DockRow || row.Children[0].Panes[0].Title != "output" || row.Children[1].Panes[0].Title != "files" {
		t.Fatalf("the part cut is not output and files: %+v", row)
	}
	top := root.Children[0]
	if top.Children[0].Panes[0].Title != "editor" || top.Sizes[0] != 0.6 || row.Sizes[1] != 0.4 {
		t.Fatalf("the rows' sizes or children changed: %+v %+v", top, row)
	}
	if d := root.Sizes[0] - 0.5; d > 0.001 || d < -0.001 {
		t.Fatalf("the rows are %v, want halves", root.Sizes)
	}
}
