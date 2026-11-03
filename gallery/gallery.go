// Package gallery, Mizan bileşenlerinin vitrinidir: her bileşen her
// boyutta, iki temada ve iki malzemede yan yana.
package gallery

import (
	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/mizan-go"
)

// State, vitrinin durumudur.
type State struct {
	Mode  mz.Mode
	Skin  mz.Skin
	Query string
	Email string
	Pass  string
	Range int
	Team  string
	Count int

	Mail, Push, Digest bool
	Plan, Tier         string
	Sync, LED          bool
	ViewMode           int
	Volume, Zoom       float64

	// Page, 0: kontroller; 1: sohbet, sürükle-bırak ve göstergeler.
	Page  int
	Msgs  []mz.ChatMessage
	Draft string
	Steps []string
	steps ui.ListState
	Cards []*card
	Files []file
}

// card, panodaki bir iştir; col, bulunduğu sütundur.
type card struct {
	Title, Tag string
	col        int
}

type file struct{ Name, Size string }

// New, örnek değerlerle bir vitrin durumu verir.
func New() *State {
	return &State{
		Range: 30, Email: "ada@", Pass: "gizli",
		Mail: true, Plan: "week", Tier: "pro", Sync: true, Volume: 64, Zoom: 1.25,
		Msgs: []mz.ChatMessage{
			{Author: "Ayşe Yılmaz", Text: "Haftalık rapor hazır mı?", Time: "09:41"},
			{Author: "Ayşe Yılmaz", Text: "Toplantıdan önce bir göz atmak istiyorum.", Time: "09:41"},
			{Mine: true, Text: "Neredeyse; son grafikleri ekliyorum.", Time: "09:43", Status: mz.Read},
			{Mine: true, Text: "On dakikaya paylaşırım.", Time: "09:43", Status: mz.Sent},
			{Author: "Mert Kaya", Text: "Ben de veri kaynaklarını kontrol ettim, sorun yok 👍", Time: "09:45"},
		},
		Steps: []string{"Veriyi içe aktar", "Sütunları eşle", "Doğrula", "Yayınla"},
		Cards: []*card{
			{Title: "Giriş ekranı", Tag: "Tasarım", col: 0},
			{Title: "Rapor dışa aktarma", Tag: "Web", col: 0},
			{Title: "Bildirim izinleri", Tag: "iOS", col: 1},
			{Title: "Karanlık tema", Tag: "Tasarım", col: 2},
		},
		Files: []file{{"Q3-rapor.pdf", "2,4 MB"}, {"hasta-listesi-ekim-2026-son-hali.xlsx", "860 KB"}},
	}
}

var search = mz.Svg(16, 15, "M7 12.5a5.5 5.5 0 1 0 0-11 5.5 5.5 0 0 0 0 11zM14.5 14.5l-3.6-3.6")

// View, vitrini kurar.
func (s *State) View(c *ui.Context) {
	t := mz.Root(c, mz.Options{Mode: s.Mode, Skin: s.Skin})
	ui.Scroll(c).Fill().Children(func() { s.page(c, t) })
}

func (s *State) page(c *ui.Context, t *mz.Tokens) {
	ui.Column(c).FillWidth().Padding(mz.Gutter-8, mz.Gutter).Gap(mz.GapSection).Children(func() {
		ui.Row(c).AlignItems(ui.End).Justify(ui.SpaceBetween).Gap(16).Children(func() {
			ui.Column(c).Gap(4).Children(func() {
				mz.Text(c, "Mizan", mz.TDisplay)
				mz.Text(c, "Sakin, veri yoğun arayüzler için native bileşenler.", mz.TLede).TextColor(t.Colors.Muted)
			})
			ui.Row(c).Gap(mz.GapControl).Children(func() {
				for i, name := range []string{"Kontroller", "Sohbet ve sürükle"} {
					o := mz.ButtonOptions{Size: mz.Sm}
					if s.Page == i {
						o.Color = t.Colors.Accent
					}
					if mz.Button(c.Key(name), name, o).Clicked() {
						s.Page = i
					}
				}
				ui.Box(c).Size(8, 0)
				mode := "Koyu"
				if t.Dark {
					mode = "Açık"
				}
				if mz.Button(c.Key("mode"), mode, mz.ButtonOptions{Size: mz.Sm}).Clicked() {
					s.Mode = mz.Dark
					if t.Dark {
						s.Mode = mz.Light
					}
				}
				skin := "Bevel"
				if t.Skin == mz.SkinBevel {
					skin = "Mizan"
				}
				if mz.Button(c.Key("skin"), skin, mz.ButtonOptions{Size: mz.Sm}).Clicked() {
					s.Skin = mz.SkinBevel
					if t.Skin == mz.SkinBevel {
						s.Skin = mz.SkinMizan
					}
				}
			})
		})
		if s.Page == 1 {
			s.more(c, t)
			return
		}
		s.buttons(c, t)
		ui.Row(c).Gap(mz.GapSection).AlignItems(ui.Start).Children(func() {
			ui.Column(c).Grow(1).Basis(0).Children(func() { s.fields(c, t) })
			ui.Column(c).Grow(1).Basis(0).Gap(mz.GapSection).Children(func() {
				s.sizes(c, t)
				s.choices(c, t)
			})
		})
	})
}

func (s *State) buttons(c *ui.Context, t *mz.Tokens) {
	cl := t.Colors
	mz.Card(c, mz.CardOptions{Title: "Button", Subtitle: "Tek tür kabarık hap; renk, dolgu ve boyut.", Pad: true}, func() {
		ui.Column(c).Gap(14).Children(func() {
			ui.Row(c).Gap(mz.GapControl).Wrap().AlignItems(ui.Center).Children(func() {
				if mz.Button(c.Key("default"), "Varsayılan", mz.ButtonOptions{}).Clicked() {
					s.Count++
				}
				mz.Button(c.Key("icon"), "Ara", mz.ButtonOptions{Icon: search})
				mz.Button(c.Key("iconOnly"), "Ara", mz.ButtonOptions{Icon: search, IconOnly: true})
				mz.Button(c.Key("soft"), "Vurgu", mz.ButtonOptions{Color: cl.Accent})
				mz.Button(c.Key("solid"), "Dışa aktar", mz.ButtonOptions{Color: cl.Accent, Variant: mz.Solid})
				mz.Button(c.Key("danger"), "Sil", mz.ButtonOptions{Danger: true})
				mz.Button(c.Key("dangerSoft"), "Kaldır", mz.ButtonOptions{Danger: true, Variant: mz.Soft})
				mz.Button(c.Key("loading"), "Kaydediliyor", mz.ButtonOptions{Loading: true})
				mz.Button(c.Key("disabled"), "Kapalı", mz.ButtonOptions{Disabled: true})
			})
			ui.Row(c).Gap(mz.GapControl).Wrap().Children(func() {
				for i, col := range []ui.Color{cl.C1, cl.C2, cl.C3, cl.C4, cl.C5, cl.C6, cl.C7, cl.C8} {
					mz.Button(c.Key(i), "Seri", mz.ButtonOptions{Color: col, Size: mz.Sm})
				}
			})
			mz.Text(c, "Varsayılan düğmeye basıldı: "+itoa(s.Count), mz.TCaption).TextColor(cl.Muted)
		})
	})
}

func (s *State) fields(c *ui.Context, t *mz.Tokens) {
	mz.Card(c, mz.CardOptions{Title: "Form", Subtitle: "Gömük alanlar ve kabarık tetikleyici.", Pad: true}, func() {
		ui.Column(c).Gap(16).Children(func() {
			mz.Field(c, mz.FieldOptions{Label: "Arama"}, func() {
				mz.Input(c.Key("q"), &s.Query, mz.InputOptions{Label: "Arama", Placeholder: "Hasta, dosya, rapor…", PrefixIcon: search, Clearable: true})
			})
			mz.Field(c, mz.FieldOptions{Label: "E-posta", Error: "Geçerli bir e-posta adresi girin.", Required: true}, func() {
				mz.Input(c.Key("email"), &s.Email, mz.InputOptions{Label: "E-posta", Invalid: true})
			})
			mz.Field(c, mz.FieldOptions{Label: "Parola", Help: "En az 12 karakter."}, func() {
				mz.Input(c.Key("pass"), &s.Pass, mz.InputOptions{Label: "Parola", Password: true})
			})
			mz.Field(c, mz.FieldOptions{Label: "Dönem"}, func() {
				mz.Select(c.Key("range"), &s.Range, []mz.Option[int]{
					{Value: 7, Label: "Son 7 gün"},
					{Value: 30, Label: "Son 30 gün", Description: "varsayılan"},
					{Value: 90, Label: "Son 90 gün"},
					{Value: 365, Label: "Son bir yıl", Disabled: true},
				}, mz.SelectOptions{Label: "Dönem"})
			})
			mz.Field(c, mz.FieldOptions{Label: "Ekip"}, func() {
				mz.Select(c.Key("team"), &s.Team, []mz.Option[string]{
					{Value: "web", Label: "Web", Dot: t.Colors.C1},
					{Value: "ios", Label: "iOS", Dot: t.Colors.C2},
					{Value: "and", Label: "Android", Dot: t.Colors.C3},
				}, mz.SelectOptions{Label: "Ekip", Placeholder: "Ekip seçin"})
			})
		})
	})
}

func (s *State) sizes(c *ui.Context, t *mz.Tokens) {
	mz.Card(c, mz.CardOptions{Title: "Boyut", Subtitle: "sm · md · lg: aynı kademede düğme ve alan aynı boyda.", Pad: true}, func() {
		ui.Column(c).Gap(14).Children(func() {
			for _, z := range []mz.Size{mz.Sm, mz.Md, mz.Lg} {
				mz.Scope(c.Key(z), mz.Options{Size: z}).Children(func() {
					ui.Row(c).Gap(mz.GapControl).AlignItems(ui.Center).Children(func() {
						ui.Column(c).Grow(1).Children(func() {
							mz.Input(c.Key("in"), &s.Query, mz.InputOptions{Label: "Arama", Placeholder: "Ara"})
						})
						mz.Button(c.Key("b"), "Uygula", mz.ButtonOptions{Color: t.Colors.Accent, Variant: mz.Solid})
					})
				})
			}
		})
	})
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for ; n > 0; n /= 10 {
		b = append([]byte{byte('0' + n%10)}, b...)
	}
	return string(b)
}
