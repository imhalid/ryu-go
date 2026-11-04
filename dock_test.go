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
