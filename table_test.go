package ryugo

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// A click on a row, or Enter on the cursor row, opens it; the app sees every
// row's element; the words are the app's; and the app can keep the selection bar.
func TestTableOpenRowLabels(t *testing.T) {
	type rec struct {
		id   int
		name string
	}
	rows := []rec{{1, "Ada"}, {2, "Bora"}, {3, "Cem"}}
	cols := []Column[rec]{{Title: "Name", Value: func(r rec) any { return r.name }, Sortable: true}}
	st := &TableState{}
	var opened []int
	seen := map[int]bool{}
	empty := false
	tt := ui.NewTester(func(c *ui.Context) {
		Root(c, Options{})
		ui.Column(c).Padding(20).Children(func() {
			rs := rows
			if empty {
				rs = nil
			}
			Table(c.Key("t"), st, cols, rs, TableOptions[rec]{
				Key: func(r rec) any { return r.id }, Selectable: true, Search: true, Height: 360,
				Open:             func(r rec) { opened = append(opened, r.id) },
				Row:              func(row ui.Element, r rec) { seen[r.id] = true },
				HideSelectionBar: true,
				Labels:           TableLabels{Search: "Ara", SelectAll: "Hepsini seç", NoRows: "Satır yok"},
			})
		})
	}, 700, 520)
	tt.Advance(500 * time.Millisecond)
	if !seen[1] || !seen[2] || !seen[3] {
		t.Fatalf("the app saw the rows %v", seen)
	}
	if !tt.HasText("Ara") && !tt.HasText("Hepsini seç") {
		t.Fatalf("the app's words are not shown: %v", tt.Texts())
	}
	if tt.HasText("Search") || tt.HasText("Select all rows") {
		t.Fatalf("English words remain: %v", tt.Texts())
	}
	if err := tt.Click("Bora"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(100 * time.Millisecond)
	if len(opened) != 1 || opened[0] != 2 {
		t.Fatalf("a click opened %v, want row 2", opened)
	}
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	tt.Advance(100 * time.Millisecond)
	if len(opened) != 2 || opened[1] != 3 {
		t.Fatalf("Down then Enter opened %v, want row 3 after row 2", opened)
	}
	// a row chosen: no bar of the table's own
	st.Selected = map[any]bool{1: true}
	tt.Advance(600 * time.Millisecond)
	if tt.HasText("1 selected") {
		t.Fatal("the selection bar shows though the app hid it")
	}
	empty = true
	tt.Frame()
	tt.Advance(500 * time.Millisecond)
	if !tt.HasText("Satır yok") {
		t.Fatalf("no empty state in the app's words: %v", tt.Texts())
	}
}
