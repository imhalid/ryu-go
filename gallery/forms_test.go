package gallery

import (
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/ryu-go"
)

// formsTester opens the forms page.
func formsTester(mode mz.Mode) (*State, *ui.Tester) {
	s := New()
	s.Mode, s.Page = mode, 3
	tt := ui.NewTester(s.View, 1180, 1100)
	tt.Advance(300 * time.Millisecond)
	return s, tt
}

// clickField clicks the control below the Field labeled label: the
// control's own accessibility name is the same, so the label's text is
// found first.
func clickField(t *testing.T, tt *ui.Tester, label string) {
	t.Helper()
	r, ok := tt.Find(label)
	if !ok {
		t.Fatalf("no field %q", label)
	}
	tt.ClickAt(r.X+40, r.Y+r.H+mz.GapField+20)
	tt.Advance(300 * time.Millisecond)
}

func TestTextarea(t *testing.T) {
	s, tt := formsTester(mz.Light)
	clickField(t, tt, "Bio")
	tt.Type("Cardiologist")
	tt.Frame()
	if s.Forms.Bio != "Cardiologist" {
		t.Fatalf("bio %q", s.Forms.Bio)
	}
	if !tt.HasText("12 / 80") {
		t.Fatalf("no count of 12 / 80 in %v", tt.Texts())
	}
	// Enter breaks the line, and the field grows with it
	r0, _ := tt.Find("12 / 80")
	tt.Key(0, ui.KeyEnter)
	tt.Type("in Oslo")
	tt.Key(0, ui.KeyEnter)
	tt.Type("and Bergen")
	tt.Advance(300 * time.Millisecond)
	if s.Forms.Bio != "Cardiologist\nin Oslo\nand Bergen" {
		t.Fatalf("bio %q", s.Forms.Bio)
	}
	r1, ok := tt.Find("31 / 80")
	if !ok {
		t.Fatalf("no count of 31 / 80 in %v", tt.Texts())
	}
	if r1.Y <= r0.Y {
		t.Fatalf("the field did not grow: its count at %.0f, before %.0f", r1.Y, r0.Y)
	}
}

// A stepper held steps once, then again and again after a pause, up to
// the maximum; letting go does not step once more.
func TestNumberInputHold(t *testing.T) {
	s, tt := formsTester(mz.Light)
	x, y := center(t, tt, "Increase")
	tt.Press(x, y)
	tt.Advance(200 * time.Millisecond)
	if s.Forms.Quantity != 3 {
		t.Fatalf("pressed: quantity %v, want 3", s.Forms.Quantity)
	}
	tt.Advance(500 * time.Millisecond)
	if q := s.Forms.Quantity; q < 5 || q >= 10 {
		t.Fatalf("held 0.7 s: quantity %v, want it repeating, under 10", q)
	}
	tt.Advance(time.Second)
	tt.Release(x, y)
	tt.Frame()
	if s.Forms.Quantity != 10 {
		t.Fatalf("held: quantity %v, want 10 (the maximum)", s.Forms.Quantity)
	}
	x, y = center(t, tt, "Decrease")
	tt.Press(x, y)
	tt.Frame()
	tt.Release(x, y)
	tt.Frame()
	if s.Forms.Quantity != 9 {
		t.Fatalf("a click: quantity %v, want 9", s.Forms.Quantity)
	}
}

func TestNumberInput(t *testing.T) {
	s, tt := formsTester(mz.Light)
	if s.Forms.Quantity != 2 {
		t.Fatalf("quantity %v, want 2", s.Forms.Quantity)
	}
	// the first stepper on the page is Quantity's, from 1 to 10
	if err := tt.Click("Increase"); err != nil {
		t.Fatal(err)
	}
	if s.Forms.Quantity != 3 {
		t.Fatalf("+: quantity %v, want 3", s.Forms.Quantity)
	}
	tt.Click("Decrease")
	tt.Click("Decrease")
	tt.Click("Decrease")
	if s.Forms.Quantity != 1 {
		t.Fatalf("−: quantity %v, want 1 (the minimum)", s.Forms.Quantity)
	}
	clickField(t, tt, "Quantity")
	for range 4 {
		tt.Key(0, ui.KeyUp)
	}
	if s.Forms.Quantity != 5 {
		t.Fatalf("Up: quantity %v, want 5", s.Forms.Quantity)
	}
	tt.Key(ui.Shift, ui.KeyUp)
	if s.Forms.Quantity != 10 {
		t.Fatalf("Shift+Up: quantity %v, want 10 (the maximum)", s.Forms.Quantity)
	}
	tt.Key(0, ui.KeyDown)
	if s.Forms.Quantity != 9 {
		t.Fatalf("Down: quantity %v, want 9", s.Forms.Quantity)
	}
	// what is typed applies once it is a number in range
	tt.Key(0, ui.KeyBackspace)
	tt.Type("7")
	tt.Frame()
	if s.Forms.Quantity != 7 {
		t.Fatalf("typed: quantity %v, want 7", s.Forms.Quantity)
	}
	tt.Type("0")
	tt.Frame()
	if s.Forms.Quantity != 7 {
		t.Fatalf("70 is out of range, yet quantity is %v", s.Forms.Quantity)
	}
	if !tt.HasText("At most 10") {
		t.Fatalf("no message for 70 in %v", tt.Texts())
	}
	// leaving the field brings what was typed into range
	tt.ClickAt(600, 20)
	tt.Advance(300 * time.Millisecond)
	if s.Forms.Quantity != 10 || tt.HasText("At most 10") {
		t.Fatalf("70 left: quantity %v, want 10 and no message", s.Forms.Quantity)
	}
	// back in the field, Down steps from the value
	clickField(t, tt, "Quantity")
	tt.Key(0, ui.KeyDown)
	if s.Forms.Quantity != 9 {
		t.Fatalf("Down after leaving: quantity %v, want 9", s.Forms.Quantity)
	}
	// the price steps by halves and shows two decimals
	clickField(t, tt, "Price")
	tt.Key(0, ui.KeyUp)
	tt.Frame()
	if s.Forms.Price != 25 {
		t.Fatalf("price %v, want 25", s.Forms.Price)
	}
	// only a number can be typed: no letters, a comma is a point, no
	// more decimals than the step's
	tt.Key(0, ui.KeyEnd)
	for range 6 {
		tt.Key(0, ui.KeyBackspace)
	}
	tt.Type("a1,77x")
	tt.Frame()
	if !tt.HasText("In steps of 0.5") {
		t.Fatalf("typed a1,77x (1.7): no message in %v", tt.Texts())
	}
	// Enter brings it to the nearest step
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if s.Forms.Price != 1.5 || tt.HasText("In steps of 0.5") {
		t.Fatalf("Enter: price %v, want 1.5", s.Forms.Price)
	}
}

func TestCombobox(t *testing.T) {
	s, tt := formsTester(mz.Light)
	clickField(t, tt, "Country")
	if !tt.HasText("Japan") || !tt.HasText("Norway") {
		t.Fatal("a click did not show all the countries")
	}
	tt.Type("n")
	tt.Advance(300 * time.Millisecond)
	// those starting with "n" first, then those holding it
	if !tt.HasText("Netherlands") || !tt.HasText("Norway") || !tt.HasText("Japan") || tt.HasText("Türkiye") {
		t.Fatalf("filtering by n: %v", tt.Texts())
	}
	tt.Type("o")
	tt.Advance(300 * time.Millisecond)
	if tt.HasText("Netherlands") || !tt.HasText("Norway") {
		t.Fatalf("filtering by no: %v", tt.Texts())
	}
	tt.Key(0, ui.KeyEnter)
	tt.Advance(300 * time.Millisecond)
	if s.Forms.Country != "no" {
		t.Fatalf("Enter chose %q, want no", s.Forms.Country)
	}
	if tt.HasText("Japan") {
		t.Fatal("the list stayed open")
	}
	// a click on an option chooses it
	clickField(t, tt, "Country")
	if err := tt.Click("Japan"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(300 * time.Millisecond)
	if s.Forms.Country != "jp" {
		t.Fatalf("the click chose %q, want jp", s.Forms.Country)
	}
	// text that matches nothing goes back to the choice when the field is left
	clickField(t, tt, "Assignee")
	tt.Type("zzz")
	tt.ClickAt(600, 20)
	tt.Advance(300 * time.Millisecond)
	if s.Forms.Assignee != "leo" {
		t.Fatalf("assignee %q, want leo", s.Forms.Assignee)
	}
	// a disabled option is not chosen
	clickField(t, tt, "Assignee")
	tt.Click("Ivy Brooks")
	tt.Advance(300 * time.Millisecond)
	if s.Forms.Assignee != "leo" {
		t.Fatalf("the disabled option was chosen: %q", s.Forms.Assignee)
	}
}

// dayLabel is how the calendar names a day for assistive technology.
func dayLabel(d time.Time) string {
	return i18n.FormatDate("en", d, i18n.DateOptions{Skeleton: "yMMMMd"})
}

func TestDatePicker(t *testing.T) {
	s, tt := formsTester(mz.Light)
	start := s.Forms.Start
	clickField(t, tt, "Start date")
	month := i18n.FormatDate("en", start, i18n.DateOptions{Skeleton: "yMMMM"})
	if !tt.HasText(month) {
		t.Fatalf("no calendar of %s: %v", month, tt.Texts())
	}
	// a click on a day chooses it and closes the calendar
	pick := time.Date(start.Year(), start.Month(), 15, 0, 0, 0, 0, time.Local)
	if err := tt.Click(dayLabel(pick)); err != nil {
		t.Fatal(err)
	}
	tt.Advance(400 * time.Millisecond)
	if !sameDate(s.Forms.Start, pick) || s.Forms.Start.Hour() != 9 {
		t.Fatalf("start %v, want the 15th at 9:00", s.Forms.Start)
	}
	if tt.HasText(month) {
		t.Fatal("the calendar stayed open")
	}
	// the month buttons and the keys: the next month's 1st, then Right and Enter
	clickField(t, tt, "Start date")
	tt.Click("Next month")
	tt.Advance(300 * time.Millisecond)
	next := time.Date(pick.Year(), pick.Month()+1, 1, 0, 0, 0, 0, time.Local)
	if !tt.HasText(i18n.FormatDate("en", next, i18n.DateOptions{Skeleton: "yMMMM"})) {
		t.Fatalf("Next month did not show %v: %v", next.Month(), tt.Texts())
	}
	tt.Key(0, ui.KeyHome)
	tt.Key(0, ui.KeyRight)
	tt.Key(0, ui.KeyEnter)
	tt.Advance(400 * time.Millisecond)
	if want := next.AddDate(0, 0, 1); !sameDate(s.Forms.Start, want) {
		t.Fatalf("start %v, want %v", s.Forms.Start, want)
	}
	// Escape closes it without choosing
	clickField(t, tt, "Due date")
	tt.Key(0, ui.KeyLeft)
	tt.Key(0, ui.KeyEscape)
	tt.Advance(400 * time.Millisecond)
	if !s.Forms.Due.IsZero() {
		t.Fatalf("Escape chose %v", s.Forms.Due)
	}
	if tt.HasText("Next month") {
		t.Fatal("Escape did not close the calendar")
	}
	// days before Min (the start date) are not chosen
	clickField(t, tt, "Due date")
	tt.Key(0, ui.KeyPageDown)
	tt.Key(0, ui.KeyPageUp)
	before := s.Forms.Start.AddDate(0, 0, -1)
	tt.Click(dayLabel(before))
	tt.Frame()
	if !s.Forms.Due.IsZero() {
		t.Fatalf("a day before Min was chosen: %v", s.Forms.Due)
	}
}

func sameDate(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

func TestTagInput(t *testing.T) {
	s, tt := formsTester(mz.Light)
	clickField(t, tt, "Labels")
	tt.Type("billing")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if !slices.Equal(s.Forms.Labels, []string{"billing"}) {
		t.Fatalf("Enter: labels %v", s.Forms.Labels)
	}
	// a comma adds too; a tag held already, in any case, is not added
	tt.Type("ops,Billing,")
	tt.Frame()
	if !slices.Equal(s.Forms.Labels, []string{"billing", "ops"}) {
		t.Fatalf("comma: labels %v", s.Forms.Labels)
	}
	// Max is three
	tt.Type("qa,legal,")
	tt.Frame()
	if !slices.Equal(s.Forms.Labels, []string{"billing", "ops", "qa"}) {
		t.Fatalf("max: labels %v", s.Forms.Labels)
	}
	// Backspace in the empty input takes out the last tag
	tt.Key(0, ui.KeyBackspace)
	tt.Frame()
	if !slices.Equal(s.Forms.Labels, []string{"billing", "ops"}) {
		t.Fatalf("Backspace: labels %v", s.Forms.Labels)
	}
	// but while the input holds text it deletes a character
	tt.Type("ab")
	tt.Key(0, ui.KeyBackspace)
	tt.Frame()
	if len(s.Forms.Labels) != 2 {
		t.Fatalf("Backspace in text took a tag out: %v", s.Forms.Labels)
	}
	// a chip's button takes it out
	if err := tt.Click("Remove follow-up"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !slices.Equal(s.Forms.Tags, []string{"cardiology", "urgent"}) {
		t.Fatalf("remove: tags %v", s.Forms.Tags)
	}
}

// TestFormsSnapshots draws the forms page in both themes, and with the
// date picker and the combobox open.
func TestFormsSnapshots(t *testing.T) {
	dir := os.Getenv("RYUGO_SNAPSHOTS")
	if dir == "" {
		t.Skip("RYUGO_SNAPSHOTS not set")
	}
	save := func(tt *ui.Tester, name string) {
		f, err := os.Create(filepath.Join(dir, name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, tt.Image()); err != nil {
			t.Fatal(err)
		}
		f.Close()
	}
	for _, v := range []struct {
		name string
		mode mz.Mode
	}{{"light", mz.Light}, {"dark", mz.Dark}} {
		_, tt := formsTester(v.mode)
		save(tt, "forms-"+v.name)
		clickField(t, tt, "Start date")
		tt.Advance(400 * time.Millisecond)
		save(tt, "forms-date-"+v.name)
		tt.Key(0, ui.KeyEscape)
		tt.Advance(400 * time.Millisecond)
		clickField(t, tt, "Assignee")
		tt.Key(0, ui.KeyDown)
		tt.Advance(400 * time.Millisecond)
		save(tt, "forms-combo-"+v.name)
	}
}
