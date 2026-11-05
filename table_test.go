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

// A table wider than its room fades at the edge with more beyond it, so that
// what is cut reads as going on: a red cell there is paler than one inside.
func TestTableSideScrollFades(t *testing.T) {
	type rec struct{ id int }
	rows := []rec{{1}, {2}, {3}}
	var cols []Column[rec]
	for _, id := range []string{"a", "b", "c", "d", "e"} {
		cols = append(cols, Column[rec]{ID: id, Title: id, Width: 220, Cell: func(c *ui.Context, r rec) {
			ui.Box(c).FillWidth().Height(24).Background(ui.Hex("#ff0000"))
		}})
	}
	st := &TableState{}
	tt := ui.NewTester(func(c *ui.Context) {
		Root(c, Options{})
		ui.Column(c).Padding(20).Children(func() {
			Table(c.Key("t"), st, cols, rows, TableOptions[rec]{Key: func(r rec) any { return r.id }, Height: 260})
		})
	}, 520, 360)
	tt.Advance(600 * time.Millisecond)
	if st.scroll.MaxX < 100 {
		t.Fatalf("the table does not scroll sideways: %v", st.scroll.MaxX)
	}
	img := tt.Image()
	y := 20 + 36 + 20 // the first row's middle: padding, header, half a row
	edge := img.RGBAAt(520-20-4, y)
	inner := img.RGBAAt(200, y)
	if inner.R < 200 || inner.G > 40 {
		t.Fatalf("inner cell %v is not red", inner)
	}
	if edge.G < inner.G+25 {
		t.Fatalf("no fade at the right edge: edge %v, inside %v", edge, inner)
	}
}
