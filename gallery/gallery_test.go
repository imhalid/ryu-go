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
	bx, by := center(t, tt, "Done")
	drag(tt, ax, ay, bx, by+40)
	if got := s.Cards[0].col; got != 2 {
		t.Fatalf("card column %d, want 2", got)
	}
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
