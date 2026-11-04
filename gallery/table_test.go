package gallery

import (
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/ryu-go"
)

// tablePageTester opens the table page.
func tablePageTester() (*State, *ui.Tester) {
	s := New()
	s.Mode, s.Page = mz.Light, 5
	tt := ui.NewTester(s.View, 1180, 1500)
	tt.Advance(400 * time.Millisecond)
	return s, tt
}

// rowCheck clicks the checkbox of the row showing text.
func rowCheck(t *testing.T, tt *ui.Tester, text string, mods ui.Modifiers) {
	t.Helper()
	x, _ := center(t, tt, "Select all rows")
	_, y := center(t, tt, text)
	tt.ClickAtWith(mods, x, y)
	tt.Frame()
}

func TestTableSort(t *testing.T) {
	s, tt := tablePageTester()
	st := &s.Tables.Orders2
	tt.Click("Total")
	if len(st.Sort) != 1 || st.Sort[0] != (mz.SortKey{Column: "Total"}) {
		t.Fatalf("one click: sort %v", st.Sort)
	}
	tt.Click("Total")
	if len(st.Sort) != 1 || !st.Sort[0].Desc {
		t.Fatalf("two clicks: sort %v, want Total descending", st.Sort)
	}
	x, y := center(t, tt, "Region")
	tt.ClickAtWith(ui.Shift, x, y)
	if len(st.Sort) != 2 || st.Sort[1].Column != "Region" {
		t.Fatalf("Shift-click: sort %v, want Region after Total", st.Sort)
	}
	tt.Click("Total")
	if len(st.Sort) != 0 {
		t.Fatalf("a click on a descending key: sort %v, want none", st.Sort)
	}
}

func TestTableSelect(t *testing.T) {
	s, tt := tablePageTester()
	st := &s.Tables.Orders2
	rowCheck(t, tt, "#1257", 0)
	if !st.Selected[1257] || len(st.Selected) != 1 {
		t.Fatalf("one row: selected %v", st.Selected)
	}
	tt.Advance(400 * time.Millisecond)
	if !tt.HasText("1 selected") {
		t.Fatal("no selection bar")
	}
	// Shift-click chooses the rows between
	rowCheck(t, tt, "#1185", ui.Shift)
	if len(st.Selected) != 5 {
		t.Fatalf("a range: %d rows chosen, want 5", len(st.Selected))
	}
	tt.Click("Clear")
	if len(st.Selected) != 0 {
		t.Fatalf("Clear left %d rows chosen", len(st.Selected))
	}
	tt.Click("Select all rows")
	if len(st.Selected) != len(s.Tables.Orders) {
		t.Fatalf("all: %d of %d chosen", len(st.Selected), len(s.Tables.Orders))
	}
	tt.Click("Select all rows")
	if len(st.Selected) != 0 {
		t.Fatalf("all again: %d chosen, want none", len(st.Selected))
	}
}

func TestTableArchive(t *testing.T) {
	s, tt := tablePageTester()
	n := len(s.Tables.Orders)
	rowCheck(t, tt, "#1257", 0)
	rowCheck(t, tt, "#1140", 0)
	tt.Advance(400 * time.Millisecond)
	if err := tt.Click("Archive"); err != nil {
		t.Fatal(err)
	}
	if len(s.Tables.Orders) != n-2 {
		t.Fatalf("archived: %d orders, want %d", len(s.Tables.Orders), n-2)
	}
	tt.Advance(500 * time.Millisecond)
	if tt.HasText("#1257") {
		t.Fatal("the archived row still shows")
	}
}

// A column dragged by its header past another goes before it.
func TestTableMoveColumn(t *testing.T) {
	s, tt := tablePageTester()
	ax, ay := center(t, tt, "Region")
	bx, _ := center(t, tt, "Status")
	drag(tt, ax, ay, bx-40, ay)
	order := s.Tables.Orders2.Order
	if r, st := slices.Index(order, "Region"), slices.Index(order, "Status"); r < 0 || r > st {
		t.Fatalf("order %v, want Region before Status", order)
	}
	if len(s.Tables.Orders2.Sort) != 1 || s.Tables.Orders2.Sort[0].Column != "Placed" {
		t.Fatalf("the drag sorted: %v", s.Tables.Orders2.Sort)
	}
}

// The end edge of a header resizes its column; a double click gives back
// its width.
func TestTableResize(t *testing.T) {
	s, tt := tablePageTester()
	r, _ := tt.Find("Region")
	x, y := r.X+r.W-1, r.Y+r.H/2
	tt.Press(x, y)
	for i := 1; i <= 6; i++ {
		tt.Move(x+float32(i)*10, y)
		tt.Advance(16 * time.Millisecond)
	}
	tt.Release(x+60, y)
	tt.Frame()
	w, ok := s.Tables.Orders2.Widths["Region"]
	if !ok || w < r.W+40 {
		t.Fatalf("width %v (%v), want about %v", w, ok, r.W+60)
	}
}

func TestTableExpand(t *testing.T) {
	s, tt := tablePageTester()
	if tt.HasText("Ships from") {
		t.Fatal("details show before a row opens")
	}
	if err := tt.Click("Details"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(500 * time.Millisecond)
	if len(s.Tables.Orders2.Expanded) != 1 || !tt.HasText("Ships from") {
		t.Fatalf("expanded %v", s.Tables.Orders2.Expanded)
	}
}

// Down moves the cursor row; Space chooses it.
func TestTableKeys(t *testing.T) {
	s, tt := tablePageTester()
	if err := tt.Click("#1257"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeySpace)
	tt.Frame()
	if sel := s.Tables.Orders2.Selected; len(sel) != 1 || !sel[1279] {
		t.Fatalf("selected %v, want the second row (1279)", sel)
	}
	tt.Key(ui.Shift, ui.KeyDown)
	tt.Key(ui.Shift, ui.KeyDown)
	tt.Frame()
	if n := len(s.Tables.Orders2.Selected); n != 3 {
		t.Fatalf("Shift+Down twice: %d chosen, want 3", n)
	}
}

func TestTablePages(t *testing.T) {
	s, tt := tablePageTester()
	if !tt.HasText("1–5 of 14") {
		t.Fatalf("no range in %v", tt.Texts())
	}
	if err := tt.Click("Page 2"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(400 * time.Millisecond)
	if s.Tables.TeamState.Page != 1 || !tt.HasText("6–10 of 14") {
		t.Fatalf("page %d", s.Tables.TeamState.Page)
	}
	tt.Click("Next page")
	tt.Click("Next page")
	if s.Tables.TeamState.Page != 2 {
		t.Fatalf("next twice: page %d, want the last (2)", s.Tables.TeamState.Page)
	}
}

func TestTableSearchAndHide(t *testing.T) {
	s, tt := tablePageTester()
	tt.Click("Search")
	tt.Type("zzzz")
	tt.Advance(300 * time.Millisecond)
	if !tt.HasText("No orders") {
		t.Fatal("no empty state for a search matching nothing")
	}
	tt.Click("View")
	tt.Advance(300 * time.Millisecond)
	// the menu's fourth item, Region: the header has a Region too
	for range 3 {
		tt.Key(0, ui.KeyDown)
	}
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if !s.Tables.Orders2.Hidden["Region"] {
		t.Fatalf("hidden %v", s.Tables.Orders2.Hidden)
	}
}
