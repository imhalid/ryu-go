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

// The panes as tabs, one shown, one closed: an area left without panes folds away, and Undo brings it back.
func TestTabsShowAndClose(t *testing.T) {
	l := NewLayout(DockArea(&Pane{ID: "a", Kind: "k"}, &Pane{ID: "b", Kind: "k"}))
	l.AddPane("b", DockRow, &Pane{ID: "c", Kind: "k"})
	tabs := l.Tabs("")
	var ids []string
	for _, tab := range tabs {
		ids = append(ids, tab.Pane.ID)
	}
	if !slices.Equal(ids, []string{"a", "b", "c"}) || !tabs[0].Shown || tabs[1].Shown || !tabs[2].Shown {
		t.Fatalf("tabs %+v", tabs)
	}
	if tabs[0].Area != tabs[1].Area || tabs[0].Area == tabs[2].Area {
		t.Fatalf("areas %+v", tabs)
	}
	if !l.ShowPane("b") || l.Tabs("")[0].Shown || !l.Tabs("")[1].Shown || l.focusArea != tabs[1].Area {
		t.Fatalf("b not shown: %+v, focus %q", l.Tabs(""), l.focusArea)
	}
	if l.ShowPane("x") || l.ClosePane("x") {
		t.Fatal("a pane the layout has not")
	}
	if !l.ClosePane("c") || l.FindPane("c") != nil || l.Find(tabs[2].Area) != nil {
		t.Fatalf("c not closed with its area: %v", l.PaneIDs())
	}
	l.ClosePane("a")
	l.ClosePane("b")
	if len(l.PaneIDs()) != 0 || l.Window("main").Root == nil {
		t.Fatal("the window's last area must stay")
	}
	l.Undo()
	l.Undo()
	l.Undo()
	if got := l.PaneIDs(); !slices.Equal(got, []string{"a", "b", "c"}) {
		t.Fatalf("after Undo: %v", got)
	}
}
