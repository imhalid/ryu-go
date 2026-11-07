package ryugo

import (
	"slices"
	"testing"
)

func TestAddAndRemovePane(t *testing.T) {
	l := NewLayout(DockArea(&Pane{ID: "a", Kind: "k"}))
	if !l.AddPane("a", DockRow, &Pane{ID: "b", Kind: "k"}) || !l.AddPane("b", DockRow, &Pane{ID: "c", Kind: "k"}) {
		t.Fatal("not added")
	}
	if got := l.PaneIDs(); !slices.Equal(got, []string{"a", "b", "c"}) {
		t.Fatalf("panes %v", got)
	}
	if l.AddPane("a", DockRow, &Pane{ID: "b", Kind: "k"}) {
		t.Fatal("the same pane went in twice")
	}
	if !l.RemovePane("b") || l.FindPane("b") != nil {
		t.Fatal("not removed")
	}
	if got := l.PaneIDs(); !slices.Equal(got, []string{"a", "c"}) {
		t.Fatalf("after: %v", got)
	}
	l.RemovePane("a")
	l.RemovePane("c")
	if len(l.PaneIDs()) != 0 || l.Window("main").Root == nil {
		t.Fatalf("the last area must stay: %v", l.PaneIDs())
	}
	// and the layout still saves and reads
	b, err := l.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseLayout(b); err != nil {
		t.Fatal(err)
	}
}
