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
	r, _ := tt.Find("report.go")
	// the editor's title sits 12 DIPs into its area; the edge is just left of it
	x, y := r.X-12-2, r.Y+100
	dragEdge(tt, x, y, 80, 0)
	if root.Sizes[0] <= before+0.03 {
		t.Fatalf("Files' size %v, was %v: the edge did not move", root.Sizes[0], before)
	}
	if sum := root.Sizes[0] + root.Sizes[1] + root.Sizes[2]; sum < 0.999 || sum > 1.001 {
		t.Fatalf("sizes add up to %v", sum)
	}
	// dragged far, it stops where Files is as small as an area goes
	r, _ = tt.Find("report.go")
	dragEdge(tt, r.X-14, r.Y+100, -600, 0)
	if w := root.Sizes[0] * (1180 - 2*68); w < 80 {
		t.Fatalf("Files shrank to about %v DIPs, under the least size", w)
	}
	// a double click on it shares the room out
	r, _ = tt.Find("report.go")
	tt.ClickAt(r.X-14, r.Y+100)
	tt.ClickAt(r.X-14, r.Y+100)
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
	return r.X + 40, r.Y + r.H/2 - 14 - 2, o.X - 12 - 2
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
