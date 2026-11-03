package gallery

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/mizan-go"
)

// TestSnapshots draws the showcase's four variants into the MIZAN_SNAPSHOTS
// directory: light/dark × mizan/bevel. For visual comparison.
func TestSnapshots(t *testing.T) {
	dir := os.Getenv("MIZAN_SNAPSHOTS")
	if dir == "" {
		t.Skip("MIZAN_SNAPSHOTS not set")
	}
	for _, v := range []struct {
		name string
		mode mz.Mode
		skin mz.Skin
	}{
		{"light-mizan", mz.Light, mz.SkinMizan},
		{"dark-mizan", mz.Dark, mz.SkinMizan},
		{"light-bevel", mz.Light, mz.SkinBevel},
		{"dark-bevel", mz.Dark, mz.SkinBevel},
	} {
		s := New()
		s.Mode, s.Skin = v.mode, v.skin
		img := ui.Render(s.View, 1180, 1500, 1)
		f, err := os.Create(filepath.Join(dir, v.name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, img); err != nil {
			t.Fatal(err)
		}
		f.Close()
	}
}

// TestSelectOpen draws an open Select list.
func TestSelectOpen(t *testing.T) {
	dir := os.Getenv("MIZAN_SNAPSHOTS")
	if dir == "" {
		t.Skip("MIZAN_SNAPSHOTS not set")
	}
	for _, mode := range []mz.Mode{mz.Light, mz.Dark} {
		s := New()
		s.Mode = mode
		tt := ui.NewTester(s.View, 1180, 1000)
		tt.Click("Last 30 days")
		tt.Advance(500 * time.Millisecond)
		tt.Key(0, ui.KeyDown)
		tt.Advance(500 * time.Millisecond)
		name := "select-light.png"
		if mode == mz.Dark {
			name = "select-dark.png"
		}
		f, err := os.Create(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		png.Encode(f, tt.Image())
		f.Close()
	}
}

// page2 sets up a tester showing the second page.
func page2(t *testing.T, mode mz.Mode, skin mz.Skin) (*State, *ui.Tester) {
	s := New()
	s.Mode, s.Skin, s.Page = mode, skin, 1
	return s, ui.NewTester(s.View, 1180, 1180)
}

// TestPage2Snapshots draws the second page (chat, drag and drop, indicators)
// in four variants.
func TestPage2Snapshots(t *testing.T) {
	dir := os.Getenv("MIZAN_SNAPSHOTS")
	if dir == "" {
		t.Skip("MIZAN_SNAPSHOTS not set")
	}
	for _, v := range []struct {
		name string
		mode mz.Mode
		skin mz.Skin
	}{
		{"more-light-mizan", mz.Light, mz.SkinMizan},
		{"more-dark-mizan", mz.Dark, mz.SkinMizan},
		{"more-light-bevel", mz.Light, mz.SkinBevel},
		{"more-dark-bevel", mz.Dark, mz.SkinBevel},
	} {
		_, tt := page2(t, v.mode, v.skin)
		tt.Advance(400 * time.Millisecond)
		f, err := os.Create(filepath.Join(dir, v.name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		png.Encode(f, tt.Image())
		f.Close()
	}
}

// drag drags the pointer from a to b in steps.
func drag(tt *ui.Tester, ax, ay, bx, by float32) {
	tt.Press(ax, ay)
	for i := 1; i <= 10; i++ {
		f := float32(i) / 10
		tt.Move(ax+(bx-ax)*f, ay+(by-ay)*f)
		tt.Advance(16 * time.Millisecond)
	}
	tt.Release(bx, by)
	tt.Advance(300 * time.Millisecond)
}

func center(t *testing.T, tt *ui.Tester, s string) (float32, float32) {
	r, ok := tt.Find(s)
	if !ok {
		t.Fatalf("no %q", s)
	}
	return r.X + r.W/2, r.Y + r.H/2
}

func TestBoardDrag(t *testing.T) {
	s, tt := page2(t, mz.Light, mz.SkinMizan)
	ax, ay := center(t, tt, "Sign-in screen")
	bx, _ := center(t, tt, "Done")
	r, _ := tt.Find("Dark theme")
	drag(tt, ax, ay, bx, r.Y+r.H+30)
	if k := s.cardNamed("Sign-in screen"); k.col != 2 {
		t.Fatalf("card column %d, want 2", k.col)
	}
	// below Dark theme, the column's only card
	if s.Cards[len(s.Cards)-1].Title != "Sign-in screen" {
		t.Fatalf("order %v, want Sign-in screen after Dark theme", titles(s.Cards))
	}
}

// Within a column, a card dragged above another goes before it.
func TestBoardReorder(t *testing.T) {
	s, tt := page2(t, mz.Light, mz.SkinMizan)
	ax, ay := center(t, tt, "Report export")
	r, _ := tt.Find("Sign-in screen")
	drag(tt, ax, ay, ax, r.Y)
	if got := titles(s.Cards)[:2]; got[0] != "Report export" || got[1] != "Sign-in screen" {
		t.Fatalf("order %v, want Report export first", titles(s.Cards))
	}
	if k := s.cardNamed("Report export"); k.col != 0 {
		t.Fatalf("column %d, want 0", k.col)
	}
}

func (s *State) cardNamed(title string) *card {
	for _, k := range s.Cards {
		if k.Title == title {
			return k
		}
	}
	return nil
}

func titles(cards []*card) []string {
	var out []string
	for _, k := range cards {
		out = append(out, k.Title)
	}
	return out
}

func TestSortableDrag(t *testing.T) {
	s, tt := page2(t, mz.Light, mz.SkinMizan)
	ax, ay := center(t, tt, "Validate")
	r, _ := tt.Find("Import data")
	drag(tt, ax, ay, ax, r.Y+2)
	if s.Steps[0] != "Validate" {
		t.Fatalf("order %v", s.Steps)
	}
}

// Alt+Down moves the focused row one place down, and the focus with it.
func TestSortableKeys(t *testing.T) {
	s, tt := page2(t, mz.Light, mz.SkinMizan)
	if err := tt.Click("Import data"); err != nil {
		t.Fatal(err)
	}
	tt.Key(ui.Alt, ui.KeyDown)
	tt.Advance(50 * time.Millisecond)
	if s.Steps[0] != "Map columns" || s.Steps[1] != "Import data" {
		t.Fatalf("Alt+Down: order %v", s.Steps)
	}
	tt.Key(ui.Alt, ui.KeyDown)
	tt.Advance(400 * time.Millisecond)
	if s.Steps[2] != "Import data" {
		t.Fatalf("Alt+Down again: order %v, want the focus to follow the row", s.Steps)
	}
}

func TestDropFiles(t *testing.T) {
	s, tt := page2(t, mz.Light, mz.SkinMizan)
	x, y := center(t, tt, "Drop files here")
	if !tt.DropFiles(x, y, "/tmp/a/report.csv") {
		t.Fatal("the zone did not take the file")
	}
	tt.Frame()
	if n := len(s.Files); n != 3 || s.Files[2].Name != "report.csv" {
		t.Fatalf("files %v", s.Files)
	}
	if err := tt.Click("Remove"); err != nil {
		t.Fatal(err)
	}
	if len(s.Files) != 2 {
		t.Fatalf("not removed: %v", s.Files)
	}
}

func TestComposerSend(t *testing.T) {
	s, tt := page2(t, mz.Light, mz.SkinMizan)
	if err := tt.Click("Message"); err != nil {
		t.Fatal(err)
	}
	tt.Type("Hello")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	last := s.Msgs[len(s.Msgs)-1]
	if !last.Mine || last.Text != "Hello" || s.Draft != "" {
		t.Fatalf("last message %+v, draft %q", last, s.Draft)
	}
}

// TestDragOverSnapshot draws a card held over the "Done" column.
func TestDragOverSnapshot(t *testing.T) {
	dir := os.Getenv("MIZAN_SNAPSHOTS")
	if dir == "" {
		t.Skip("MIZAN_SNAPSHOTS not set")
	}
	_, tt := page2(t, mz.Light, mz.SkinMizan)
	ax, ay := center(t, tt, "Sign-in screen")
	bx, by := center(t, tt, "Done")
	tt.Press(ax, ay)
	for i := 1; i <= 10; i++ {
		f := float32(i) / 10
		tt.Move(ax+(bx-ax)*f, ay+(by+60-ay)*f)
		tt.Advance(16 * time.Millisecond)
	}
	tt.Advance(200 * time.Millisecond)
	f, _ := os.Create(filepath.Join(dir, "drag-over.png"))
	png.Encode(f, tt.Image())
	f.Close()
}

func TestChoices(t *testing.T) {
	s := New()
	tt := ui.NewTester(s.View, 1180, 1500)
	for _, step := range []struct {
		click string
		ok    func() bool
	}{
		{"Push notifications", func() bool { return s.Push }},
		{"Newsletter", func() bool { return !s.Mail }},
		{"Monthly", func() bool { return s.Plan == "month" }},
		{"Enterprise", func() bool { return s.Tier == "team" }},
		{"Sync", func() bool { return !s.Sync }},
		{"Calendar", func() bool { return s.ViewMode == 2 }},
	} {
		if err := tt.Click(step.click); err != nil {
			t.Fatal(err)
		}
		tt.Advance(300 * time.Millisecond)
		if !step.ok() {
			t.Fatalf("clicking %q did not change the state", step.click)
		}
	}
	if err := tt.Click("Disabled"); err != nil {
		t.Fatal(err)
	}
	// slider: the right arrow adds a step, dragging goes to the end
	r, ok := tt.Find("Volume")
	if !ok {
		t.Fatal("no Volume slider")
	}
	tt.ClickAt(r.X+r.W/2, r.Y+r.H/2)
	tt.Key(0, ui.KeyRight)
	before := s.Volume
	tt.Key(0, ui.KeyRight)
	if s.Volume != before+1 {
		t.Fatalf("volume %v, want %v", s.Volume, before+1)
	}
	drag(tt, r.X+r.W/2, r.Y+r.H/2, r.X+r.W+40, r.Y+r.H/2)
	if s.Volume != 100 {
		t.Fatalf("volume %v, want 100", s.Volume)
	}
}

// TestAccentSnapshots draws the first page with brand accents.
func TestAccentSnapshots(t *testing.T) {
	dir := os.Getenv("MIZAN_SNAPSHOTS")
	if dir == "" {
		t.Skip("MIZAN_SNAPSHOTS not set")
	}
	for _, v := range []struct {
		name string
		mode mz.Mode
		skin mz.Skin
		pick string
	}{
		{"accent-orange-light", mz.Light, mz.SkinMizan, "Orange"},
		{"accent-violet-dark", mz.Dark, mz.SkinBevel, "Violet"},
	} {
		s := New()
		s.Mode, s.Skin = v.mode, v.skin
		tt := ui.NewTester(s.View, 1180, 1500)
		if err := tt.Click(v.pick); err != nil {
			t.Fatal(err)
		}
		tt.Advance(500 * time.Millisecond)
		f, err := os.Create(filepath.Join(dir, v.name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		png.Encode(f, tt.Image())
		f.Close()
	}
}

// page3 opens the layers page.
func page3(mode mz.Mode, skin mz.Skin) (*State, *ui.Tester) {
	s := New()
	s.Mode, s.Skin, s.Page = mode, skin, 2
	return s, ui.NewTester(s.View, 1180, 860)
}

func TestDialog(t *testing.T) {
	slow(t)
	s, tt := page3(mz.Light, mz.SkinMizan)
	if err := tt.Click("Open dialog"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(400 * time.Millisecond)
	if !tt.HasText("Delete the report?") {
		t.Fatal("the dialog did not open")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Advance(400 * time.Millisecond)
	if s.Dialog || tt.HasText("Delete the report?") {
		t.Fatal("Escape did not close the dialog")
	}
	tt.Click("Open dialog")
	tt.Advance(400 * time.Millisecond)
	if err := tt.Click("Delete report"); err != nil {
		t.Fatal(err)
	}
	if s.Dialog || s.Last != "Deleted" {
		t.Fatalf("confirm: open %v, last %q", s.Dialog, s.Last)
	}
	tt.Click("Open dialog")
	tt.Advance(400 * time.Millisecond)
	if err := tt.Click("Close"); err != nil {
		t.Fatal(err)
	}
	if s.Dialog {
		t.Fatal("the close button did not close the dialog")
	}
}

func TestSheet(t *testing.T) {
	slow(t)
	s, tt := page3(mz.Light, mz.SkinMizan)
	tt.Click("Open sheet")
	tt.Advance(500 * time.Millisecond)
	if err := tt.Click("Save"); err != nil {
		t.Fatal(err)
	}
	if s.Sheet || s.Last != "Saved" {
		t.Fatalf("sheet: open %v, last %q", s.Sheet, s.Last)
	}
}

func TestMenu(t *testing.T) {
	s, tt := page3(mz.Light, mz.SkinMizan)
	tt.Click("More")
	tt.Advance(300 * time.Millisecond)
	if err := tt.Click("Share"); err != nil {
		t.Fatal(err)
	}
	if s.Menu || s.Last != "Share" {
		t.Fatalf("menu: open %v, last %q", s.Menu, s.Last)
	}
	// the keys: the first item has the focus, Down moves to the next
	tt.Click("More")
	tt.Advance(300 * time.Millisecond)
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if s.Last != "Share" || s.Menu {
		t.Fatalf("keys: open %v, last %q", s.Menu, s.Last)
	}
	tt.Click("Sort")
	tt.Advance(300 * time.Millisecond)
	if err := tt.Click("Date"); err != nil {
		t.Fatal(err)
	}
	if s.Sort != "Date" {
		t.Fatalf("sort %q", s.Sort)
	}
}

func TestToast(t *testing.T) {
	s, tt := page3(mz.Light, mz.SkinMizan)
	tt.Click("Archive")
	tt.Advance(300 * time.Millisecond)
	if !tt.HasText("3 reports archived") {
		t.Fatal("no toast")
	}
	if err := tt.Click("Undo"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(500 * time.Millisecond)
	if s.Last != "Undo" || tt.HasText("3 reports archived") {
		t.Fatalf("undo: last %q", s.Last)
	}
	tt.Click("Notify")
	tt.Advance(5 * time.Second)
	if tt.HasText("Report shared") {
		t.Fatal("the toast did not time out")
	}
}

// A toast swiped past a third of its width goes; a short swipe springs
// back.
func TestToastSwipe(t *testing.T) {
	_, tt := page3(mz.Light, mz.SkinMizan)
	tt.Click("Succeed")
	tt.Advance(800 * time.Millisecond)
	x, y := center(t, tt, "Export finished")
	swipe := func(by float32) {
		tt.Press(x, y)
		for i := 1; i <= 8; i++ {
			tt.Move(x+by*float32(i)/8, y)
			tt.Advance(16 * time.Millisecond)
		}
		tt.Release(x+by, y)
		tt.Advance(600 * time.Millisecond)
	}
	swipe(40)
	if r, ok := tt.Find("Export finished"); !ok || r.X-x > 20 {
		t.Fatalf("a short swipe: the toast at %v (%v), want it back", r, ok)
	}
	swipe(200)
	if tt.HasText("Export finished") {
		t.Fatal("a long swipe left the toast")
	}
}

// A loading toast turns into its result in place.
func TestToastLoading(t *testing.T) {
	_, tt := page3(mz.Light, mz.SkinMizan)
	tt.Click("Upload")
	tt.Advance(800 * time.Millisecond)
	if !tt.HasText("Uploading Q3-report.pdf") {
		t.Fatal("no loading toast")
	}
	tt.Advance(1600 * time.Millisecond)
	tt.Advance(600 * time.Millisecond)
	if tt.HasText("Uploading Q3-report.pdf") || !tt.HasText("Q3-report.pdf uploaded") {
		t.Fatalf("the upload did not turn into its result: %v", tt.Texts())
	}
}

func TestTooltip(t *testing.T) {
	_, tt := page3(mz.Light, mz.SkinMizan)
	x, y := center(t, tt, "Share link")
	tt.Move(x, y)
	tt.Advance(2 * time.Second)
	if !tt.HasText("Share the report") {
		t.Fatal("no tooltip")
	}
}

// TestLayerSnapshots draws the layers open.
func TestLayerSnapshots(t *testing.T) {
	dir := os.Getenv("MIZAN_SNAPSHOTS")
	if dir == "" {
		t.Skip("MIZAN_SNAPSHOTS not set")
	}
	save := func(tt *ui.Tester, name string) {
		f, err := os.Create(filepath.Join(dir, name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		png.Encode(f, tt.Image())
		f.Close()
	}
	for _, v := range []struct {
		name string
		mode mz.Mode
		skin mz.Skin
	}{{"light", mz.Light, mz.SkinMizan}, {"dark", mz.Dark, mz.SkinBevel}} {
		_, tt := page3(v.mode, v.skin)
		tt.Click("Open dialog")
		tt.Advance(600 * time.Millisecond)
		save(tt, "dialog-"+v.name)
		tt.Key(0, ui.KeyEscape)
		tt.Advance(600 * time.Millisecond)
		tt.Click("Open sheet")
		tt.Advance(800 * time.Millisecond)
		save(tt, "sheet-"+v.name)
		tt.Key(0, ui.KeyEscape)
		tt.Advance(600 * time.Millisecond)
		tt.Click("Succeed")
		tt.Click("Archive")
		tt.Click("Fail")
		tt.Advance(600 * time.Millisecond)
		tt.Click("More")
		tt.Advance(300 * time.Millisecond)
		tt.Key(0, ui.KeyDown)
		x, y := center(t, tt, "Share link")
		_ = x
		tt.Move(x+300, y)
		tt.Advance(300 * time.Millisecond)
		save(tt, "menu-toast-"+v.name)
	}
}

func TestDialogDrag(t *testing.T) {
	slow(t)
	_, tt := page3(mz.Light, mz.SkinMizan)
	tt.Click("Open dialog")
	tt.Advance(600 * time.Millisecond)
	r0, ok := tt.Find("Delete the report?")
	if !ok {
		t.Fatal("no dialog")
	}
	x, y := r0.X+20, r0.Y+r0.H/2
	tt.Press(x, y)
	for i := 1; i <= 20; i++ {
		tt.Move(x+float32(i)*8, y+float32(i)*3)
		tt.Advance(16 * time.Millisecond)
	}
	tt.Advance(100 * time.Millisecond) // hold still, so it is not thrown
	tt.Release(x+160, y+60)
	tt.Advance(time.Second)
	r1, _ := tt.Find("Delete the report?")
	if dx, dy := r1.X-r0.X, r1.Y-r0.Y; dx < 150 || dx > 170 || dy < 50 || dy > 70 {
		t.Fatalf("the dialog moved by %.0f, %.0f; want about 160, 60", dx, dy)
	}
	// dragged far past the window, it springs back inside it
	x, y = r1.X+20, r1.Y+r1.H/2
	tt.Press(x, y)
	for i := 1; i <= 20; i++ {
		tt.Move(x+float32(i)*60, y)
		tt.Advance(16 * time.Millisecond)
	}
	tt.Release(x+1200, y)
	tt.Advance(2 * time.Second)
	panel, _ := tt.Find("Delete report")
	if panel.X+panel.W > 1180 {
		t.Fatalf("the dialog rests outside the window: its button ends at %.0f", panel.X+panel.W)
	}
	if r2, _ := tt.Find("Delete the report?"); r2.X <= r1.X {
		t.Fatalf("the dialog did not move right: %.0f, before %.0f", r2.X, r1.X)
	}
}

// TestPolishSnapshots draws a scrolled dialog's faded edges, a dialog
// being dragged, and the glass of a menu over the page.
func TestPolishSnapshots(t *testing.T) {
	dir := os.Getenv("MIZAN_SNAPSHOTS")
	if dir == "" {
		t.Skip("MIZAN_SNAPSHOTS not set")
	}
	save := func(tt *ui.Tester, name string) {
		f, _ := os.Create(filepath.Join(dir, name+".png"))
		png.Encode(f, tt.Image())
		f.Close()
	}
	for _, v := range []struct {
		name string
		mode mz.Mode
		skin mz.Skin
	}{{"light", mz.Light, mz.SkinMizan}, {"dark", mz.Dark, mz.SkinMizan}} {
		_, tt := page3(v.mode, v.skin)
		tt.Click("Open dialog")
		tt.Advance(600 * time.Millisecond)
		r, _ := tt.Find("Delete the report?")
		tt.Wheel(r.X+100, r.Y+300, 0, 160)
		tt.Advance(600 * time.Millisecond)
		save(tt, "dialog-scrolled-"+v.name)
		x, y := r.X+20, r.Y+r.H/2
		tt.Press(x, y)
		for i := 1; i <= 8; i++ {
			tt.Move(x+float32(i)*22, y+float32(i)*4)
			tt.Advance(16 * time.Millisecond)
		}
		save(tt, "dialog-dragging-"+v.name)
		tt.Release(x+176, y+32)
	}
}

// Toasts stack behind the newest, and spread out under the pointer.
func TestToastStack(t *testing.T) {
	_, tt := page3(mz.Light, mz.SkinMizan)
	tt.Click("Succeed")
	tt.Click("Archive")
	tt.Click("Fail")
	tt.Advance(time.Second)
	old, _ := tt.Find("Export finished")
	front, _ := tt.Find("Upload failed")
	if gap := front.Y - old.Y; gap > 40 {
		t.Fatalf("collapsed, the oldest toast is %.0f above the newest; want it tucked behind", gap)
	}
	tt.Move(front.X+40, front.Y+4)
	tt.Advance(time.Second)
	old, _ = tt.Find("Export finished")
	front, _ = tt.Find("Upload failed")
	if gap := front.Y - old.Y; gap < 100 {
		t.Fatalf("hovered, the oldest toast is %.0f above the newest; want the stack spread out", gap)
	}
}

func TestTabs(t *testing.T) {
	s, tt := page3(mz.Light, mz.SkinMizan)
	if err := tt.Click("Activity"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(400 * time.Millisecond)
	if s.Tab != 1 || !tt.HasText("12 comments and 3 exports since Monday.") {
		t.Fatalf("tab %d", s.Tab)
	}
	tt.Key(0, ui.KeyRight)
	tt.Advance(400 * time.Millisecond)
	if s.Tab != 2 {
		t.Fatalf("Right: tab %d, want 2", s.Tab)
	}
}

// slow skips a test under -short: the layers draw glass and shadows,
// which the Tester draws on the CPU.
func slow(t *testing.T) {
	if testing.Short() {
		t.Skip("blurs on the CPU; skipped with -short")
	}
}
