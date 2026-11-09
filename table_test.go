package ryugo

import (
	"fmt"
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

// A column with a Foot puts a footer under the rows: each cell under its
// column's header, a number at the end as the column's cells are, and the
// rows' view no taller than its room.
func TestTableFooter(t *testing.T) {
	type rec struct {
		id    int
		name  string
		total float64
	}
	rows := []rec{{1, "Ada", 10}, {2, "Bora", 20.5}, {3, "Cem", 4}}
	cols := []Column[rec]{
		{Title: "Name", Value: func(r rec) any { return r.name }, Foot: func(rs []rec) string { return fmt.Sprintf("%d people", len(rs)) }},
		{Title: "Note", Value: func(r rec) any { return "-" }},
		{Title: "Total", Kind: KindNumber, Value: func(r rec) any { return r.total }, Decimals: 1,
			Foot: func(rs []rec) string {
				var sum float64
				for _, r := range rs {
					sum += r.total
				}
				return fmt.Sprintf("%.1f", sum)
			}},
	}
	st := &TableState{}
	tt := ui.NewTester(func(c *ui.Context) {
		Root(c, Options{})
		ui.Column(c).Padding(20).Children(func() {
			Table(c.Key("t"), st, cols, rows, TableOptions[rec]{Key: func(r rec) any { return r.id }, Selectable: true, Height: 300})
		})
	}, 700, 400)
	tt.Advance(500 * time.Millisecond)
	pad := tableSizes(Md).pad
	name, ok1 := tt.Find("Name")
	total, ok2 := tt.Find("Total")
	people, ok3 := tt.Find("3 people")
	sum, ok4 := tt.Find("34.5")
	if !ok1 || !ok2 || !ok3 || !ok4 {
		t.Fatalf("missing header or footer cells: %v", tt.Texts())
	}
	if d := people.X - (name.X + pad); abs32(d) > 1 {
		t.Fatalf("the Name footer starts at %v, its column's text at %v", people.X, name.X+pad)
	}
	if d := (sum.X + sum.W) - (total.X + total.W - pad); abs32(d) > 1 {
		t.Fatalf("the Total footer ends at %v, its column's end at %v", sum.X+sum.W, total.X+total.W-pad)
	}
	// the footer sits at the table's bottom, inside its height
	if bottom := float32(20 + 300); people.Y > bottom || people.Y < bottom-tableSizes(Md).head {
		t.Fatalf("the footer is at %v, not in the table's last %v DIPs above %v", people.Y, tableSizes(Md).head, bottom)
	}
	// hidden, a column takes its footer cell with it
	st.Hidden = map[string]bool{"Name": true}
	tt.Frame()
	tt.Advance(300 * time.Millisecond)
	if tt.HasText("3 people") {
		t.Fatal("a hidden column's footer still shows")
	}
	if !tt.HasText("34.5") {
		t.Fatal("the shown column's footer went")
	}
}

// With Group, rows show under a heading per group, the groups in the order
// they first come in the rows; a header's sort sorts within each group, and
// a click on a heading folds it, taking its rows out, and opens it again.
func TestTableGroups(t *testing.T) {
	type rec struct {
		id         int
		name, team string
	}
	rows := []rec{{1, "Zeki", "Mobile"}, {2, "Bora", "Growth"}, {3, "Ada", "Mobile"}, {4, "Cem", "Growth"}, {5, "Eda", "Payments"}}
	cols := []Column[rec]{{Title: "Name", Value: func(r rec) any { return r.name }, Sortable: true}}
	st := &TableState{}
	tt := ui.NewTester(func(c *ui.Context) {
		Root(c, Options{})
		ui.Column(c).Padding(20).Children(func() {
			Table(c.Key("t"), st, cols, rows, TableOptions[rec]{
				Key: func(r rec) any { return r.id }, Height: 460,
				Group:      func(r rec) string { return r.team },
				GroupLabel: func(g string, rs []rec) string { return "Team " + g },
			})
		})
	}, 600, 520)
	tt.Advance(500 * time.Millisecond)
	ys := func(names ...string) []float32 {
		t.Helper()
		var out []float32
		for _, n := range names {
			r, ok := tt.Find(n)
			if !ok {
				t.Fatalf("no %q in %v", n, tt.Texts())
			}
			out = append(out, r.Y)
		}
		return out
	}
	inOrder := func(names ...string) {
		t.Helper()
		y := ys(names...)
		for i := 1; i < len(y); i++ {
			if y[i] <= y[i-1] {
				t.Fatalf("%q (y %v) is not under %q (y %v)", names[i], y[i], names[i-1], y[i-1])
			}
		}
	}
	inOrder("Team Mobile", "Zeki", "Ada", "Team Growth", "Bora", "Cem", "Team Payments", "Eda")
	// sorted by name: within the groups, which keep their order
	tt.Click("Name")
	tt.Advance(600 * time.Millisecond)
	inOrder("Team Mobile", "Ada", "Zeki", "Team Growth", "Bora", "Cem", "Team Payments", "Eda")
	// folded, a group keeps its heading and loses its rows
	if err := tt.Click("Team Mobile"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(600 * time.Millisecond)
	if !st.Collapsed["Mobile"] {
		t.Fatalf("Collapsed is %v after a click on Mobile's heading", st.Collapsed)
	}
	if tt.HasText("Ada") || tt.HasText("Zeki") {
		t.Fatal("a folded group's rows still show")
	}
	inOrder("Team Mobile", "Team Growth", "Bora", "Cem", "Team Payments", "Eda")
	// the keys go past the folded rows
	tt.Click("Bora")
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyDown)
	tt.Advance(100 * time.Millisecond)
	if st.cursor != 2 {
		t.Fatalf("the cursor is on row %d, want 2 (Eda)", st.cursor)
	}
	tt.Click("Team Mobile")
	tt.Advance(600 * time.Millisecond)
	inOrder("Team Mobile", "Ada", "Zeki", "Team Growth")
}

// With pages, a page counts the rows alone and shows the heading of each
// group it has rows of.
func TestTableGroupPages(t *testing.T) {
	type rec struct {
		id   int
		team string
	}
	var rows []rec
	for i := range 6 {
		rows = append(rows, rec{i, []string{"A", "A", "A", "B", "B", "B"}[i]})
	}
	cols := []Column[rec]{{Title: "ID", Value: func(r rec) any { return fmt.Sprintf("row %d", r.id) }}}
	st := &TableState{}
	tt := ui.NewTester(func(c *ui.Context) {
		Root(c, Options{})
		ui.Column(c).Padding(20).Children(func() {
			Table(c.Key("t"), st, cols, rows, TableOptions[rec]{
				Key: func(r rec) any { return r.id }, Height: 400, PageSize: 4,
				Group: func(r rec) string { return "group " + r.team },
			})
		})
	}, 600, 520)
	tt.Advance(500 * time.Millisecond)
	for _, s := range []string{"group A", "row 0", "row 2", "group B", "row 3"} {
		if !tt.HasText(s) {
			t.Fatalf("page 1 lacks %q: %v", s, tt.Texts())
		}
	}
	if tt.HasText("row 4") {
		t.Fatal("page 1 shows a fifth row")
	}
	st.Page = 1
	tt.Frame()
	tt.Advance(500 * time.Millisecond)
	if !tt.HasText("group B") || !tt.HasText("row 5") || tt.HasText("group A") {
		t.Fatalf("page 2 shows %v", tt.Texts())
	}
}
