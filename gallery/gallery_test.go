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

// TestSnapshots, MIZAN_SNAPSHOTS klasörüne vitrinin dört hâlini çizer:
// açık/koyu × mizan/bevel. Görsel karşılaştırma için.
func TestSnapshots(t *testing.T) {
	dir := os.Getenv("MIZAN_SNAPSHOTS")
	if dir == "" {
		t.Skip("MIZAN_SNAPSHOTS yok")
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

// TestSelectOpen, açık Select listesini çizer.
func TestSelectOpen(t *testing.T) {
	dir := os.Getenv("MIZAN_SNAPSHOTS")
	if dir == "" {
		t.Skip("MIZAN_SNAPSHOTS yok")
	}
	for _, mode := range []mz.Mode{mz.Light, mz.Dark} {
		s := New()
		s.Mode = mode
		tt := ui.NewTester(s.View, 1180, 1000)
		tt.Click("Son 30 gün")
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

// page2, ikinci sayfayı açan bir test ortamı kurar.
func page2(t *testing.T, mode mz.Mode, skin mz.Skin) (*State, *ui.Tester) {
	s := New()
	s.Mode, s.Skin, s.Page = mode, skin, 1
	return s, ui.NewTester(s.View, 1180, 1180)
}

// TestPage2Snapshots, ikinci sayfayı (sohbet, sürükle-bırak, göstergeler)
// dört hâlde çizer.
func TestPage2Snapshots(t *testing.T) {
	dir := os.Getenv("MIZAN_SNAPSHOTS")
	if dir == "" {
		t.Skip("MIZAN_SNAPSHOTS yok")
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

// drag, işaretçiyi a'dan b'ye adım adım sürükler.
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
		t.Fatalf("%q yok", s)
	}
	return r.X + r.W/2, r.Y + r.H/2
}

func TestBoardDrag(t *testing.T) {
	s, tt := page2(t, mz.Light, mz.SkinMizan)
	ax, ay := center(t, tt, "Giriş ekranı")
	bx, by := center(t, tt, "Bitti")
	drag(tt, ax, ay, bx, by+40)
	if got := s.Cards[0].col; got != 2 {
		t.Fatalf("kart sütunu %d, 2 olmalı", got)
	}
}

func TestSortableDrag(t *testing.T) {
	s, tt := page2(t, mz.Light, mz.SkinMizan)
	ax, ay := center(t, tt, "Doğrula")
	r, _ := tt.Find("Veriyi içe aktar")
	drag(tt, ax, ay, ax, r.Y+2)
	if s.Steps[0] != "Doğrula" {
		t.Fatalf("sıra %v", s.Steps)
	}
}

func TestDropFiles(t *testing.T) {
	s, tt := page2(t, mz.Light, mz.SkinMizan)
	x, y := center(t, tt, "Dosyaları buraya bırakın")
	if !tt.DropFiles(x, y, "/tmp/a/rapor.csv") {
		t.Fatal("bölge dosyayı almadı")
	}
	tt.Frame()
	if n := len(s.Files); n != 3 || s.Files[2].Name != "rapor.csv" {
		t.Fatalf("dosyalar %v", s.Files)
	}
	if err := tt.Click("Kaldır"); err != nil {
		t.Fatal(err)
	}
	if len(s.Files) != 2 {
		t.Fatalf("kaldırılmadı: %v", s.Files)
	}
}

func TestComposerSend(t *testing.T) {
	s, tt := page2(t, mz.Light, mz.SkinMizan)
	if err := tt.Click("Mesaj"); err != nil {
		t.Fatal(err)
	}
	tt.Type("Merhaba")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	last := s.Msgs[len(s.Msgs)-1]
	if !last.Mine || last.Text != "Merhaba" || s.Draft != "" {
		t.Fatalf("son mesaj %+v, taslak %q", last, s.Draft)
	}
}

// TestDragOverSnapshot, kart "Bitti" sütununun üstündeyken çizer.
func TestDragOverSnapshot(t *testing.T) {
	dir := os.Getenv("MIZAN_SNAPSHOTS")
	if dir == "" {
		t.Skip("MIZAN_SNAPSHOTS yok")
	}
	_, tt := page2(t, mz.Light, mz.SkinMizan)
	ax, ay := center(t, tt, "Giriş ekranı")
	bx, by := center(t, tt, "Bitti")
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
		{"Anlık bildirim", func() bool { return s.Push }},
		{"Bülten", func() bool { return !s.Mail }},
		{"Aylık", func() bool { return s.Plan == "month" }},
		{"Kurumsal", func() bool { return s.Tier == "team" }},
		{"Eşitleme", func() bool { return !s.Sync }},
		{"Takvim", func() bool { return s.ViewMode == 2 }},
	} {
		if err := tt.Click(step.click); err != nil {
			t.Fatal(err)
		}
		tt.Advance(300 * time.Millisecond)
		if !step.ok() {
			t.Fatalf("%q tıklanınca durum değişmedi", step.click)
		}
	}
	if err := tt.Click("Kapalı"); err != nil {
		t.Fatal(err)
	}
	// slider: sağ ok bir adım artırır, sürükleme sona götürür
	r, ok := tt.Find("Ses")
	if !ok {
		t.Fatal("Ses kaydırıcısı yok")
	}
	tt.ClickAt(r.X+r.W/2, r.Y+r.H/2)
	tt.Key(0, ui.KeyRight)
	before := s.Volume
	tt.Key(0, ui.KeyRight)
	if s.Volume != before+1 {
		t.Fatalf("ses %v, %v olmalı", s.Volume, before+1)
	}
	drag(tt, r.X+r.W/2, r.Y+r.H/2, r.X+r.W+40, r.Y+r.H/2)
	if s.Volume != 100 {
		t.Fatalf("ses %v, 100 olmalı", s.Volume)
	}
}
