package gallery

import (
	"strings"

	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/mizan-go"
)

var inbox = mz.Svg(16, 18, "M1.5 9.5h3.5l1 2h4l1-2h3.5M3.5 3.5h9l2 6v3a1 1 0 0 1-1 1h-11a1 1 0 0 1-1-1v-3z")

// more, ikinci sayfadır: sohbet, sürükle-bırak ve göstergeler.
func (s *State) more(c *ui.Context, t *mz.Tokens) {
	ui.Row(c).Gap(mz.GapSection).AlignItems(ui.Start).Grow(1).MinHeight(0).Children(func() {
		ui.Column(c).Grow(1).Basis(0).Gap(mz.GapSection).Children(func() {
			s.chat(c, t)
			s.indicators(c, t)
		})
		ui.Column(c).Grow(1).Basis(0).Gap(mz.GapSection).Children(func() {
			s.board(c, t)
			s.files(c, t)
		})
	})
}

func (s *State) chat(c *ui.Context, t *mz.Tokens) {
	mz.Card(c, mz.CardOptions{Title: "Sohbet", Subtitle: "Kendi mesajın solid malzeme, karşınınki ikinci yüzey.", Pad: true}, func() {
		ui.Column(c).Gap(2).Children(func() {
			mz.DaySeparator(c, "Bugün")
			for i, m := range s.Msgs {
				mz.Message(c.Key(i), m, mz.GroupOf(s.Msgs, i))
			}
			mz.Typing(c, "Mert Kaya", mz.AvatarOptions{})
			ui.Box(c).Size(0, 6)
			if mz.Composer(c.Key("composer"), &s.Draft, mz.ComposerOptions{Placeholder: "Mesaj yazın", Attach: func() {}}) {
				s.Msgs = append(s.Msgs, mz.ChatMessage{Mine: true, Text: strings.TrimSpace(s.Draft), Time: "09:46", Status: mz.Sending})
				s.Draft = ""
			}
		})
	})
}

func (s *State) indicators(c *ui.Context, t *mz.Tokens) {
	cl := t.Colors
	mz.Card(c, mz.CardOptions{Title: "Göstergeler", Subtitle: "Rozet, durum, tuş ve avatar.", Pad: true}, func() {
		ui.Column(c).Gap(14).Children(func() {
			ui.Row(c).Gap(8).Wrap().AlignItems(ui.Center).Children(func() {
				mz.Badge(c, "+4,2%", mz.ToneDown)
				mz.Badge(c, "−1,8%", mz.ToneUp)
				mz.Badge(c, "Beklemede", mz.ToneWarn)
				mz.Badge(c, "Taslak", mz.ToneNeutral)
				mz.Tag(c, "v2.4")
			})
			ui.Row(c).Gap(18).Wrap().AlignItems(ui.Center).Children(func() {
				mz.StatusDot(c, mz.DotLive, "Canlı")
				mz.StatusDot(c, mz.DotDelayed, "Gecikmeli")
				mz.StatusDot(c, mz.DotOff, "Kapalı")
				ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
					ui.Text(c, "Ara").TextColor(cl.Muted).FontSize(mz.FsSm)
					mz.Kbd(c, "⌘", "K")
				})
			})
			ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
				for _, n := range []string{"Ayşe Yılmaz", "Mert Kaya", "Deniz Şahin", "Can Öztürk"} {
					mz.Avatar(c.Key(n), n, mz.AvatarOptions{Online: n == "Mert Kaya"})
				}
				mz.Avatar(c.Key("solid"), "Ece Arslan", mz.AvatarOptions{Solid: true})
			})
		})
	})
}

func (s *State) board(c *ui.Context, t *mz.Tokens) {
	mz.Card(c, mz.CardOptions{Title: "Sürükle-bırak", Subtitle: "Sürüklenen kabarık, bırakılan yer gömük.", Pad: true}, func() {
		ui.Column(c).Gap(16).Children(func() {
			mz.Sortable(c.Key("steps"), &s.steps, &s.Steps, mz.SortableOptions{
				Label: func(i int) string { return s.Steps[i] },
				Key:   func(i int) any { return s.Steps[i] },
			}, func(step string) {
				ui.Text(c, step).Grow(1)
			})
			ui.Row(c).Gap(10).AlignItems(ui.Stretch).Children(func() {
				for col, title := range []string{"Yapılacak", "Sürüyor", "Bitti"} {
					n := 0
					for _, k := range s.Cards {
						if k.col == col {
							n++
						}
					}
					ui.Column(c.Key(col)).Grow(1).Basis(0).MinWidth(0).Children(func() {
						k, ok := mz.DropArea[*card](c, mz.DropAreaOptions{
							Title: title, Count: itoa(n), Fill: true,
							Accept: func(v any) bool { return v.(*card).col != col },
						}, func() {
							for _, k := range s.Cards {
								if k.col != col {
									continue
								}
								mz.DragCard(c.Key(k.Title), k, func() {
									ui.Text(c, k.Title).Wrap()
									ui.Row(c).Children(func() { mz.Badge(c, k.Tag, mz.ToneNeutral) })
								})
							}
						})
						if ok {
							k.col = col
						}
					})
				}
			})
		})
	})
}

func (s *State) files(c *ui.Context, t *mz.Tokens) {
	mz.Card(c, mz.CardOptions{Title: "Dosyalar", Subtitle: "Finder ya da Explorer'dan sürükleyin.", Pad: true}, func() {
		ui.Column(c).Gap(8).Children(func() {
			for _, p := range mz.DropZone(c.Key("drop"), mz.DropZoneOptions{Title: "Dosyaları buraya bırakın", Hint: "PDF, XLSX · en çok 10 MB"}) {
				s.Files = append(s.Files, file{Name: p[strings.LastIndexAny(p, `/\`)+1:]})
			}
			for i := 0; i < len(s.Files); i++ {
				if mz.FileItem(c.Key(s.Files[i].Name), s.Files[i].Name, s.Files[i].Size) {
					s.Files = append(s.Files[:i], s.Files[i+1:]...)
					i--
				}
			}
			if len(s.Files) == 0 {
				mz.EmptyState(c, mz.EmptyStateOptions{Icon: inbox, Title: "Dosya yok", Description: "Eklediğiniz dosyalar burada listelenir."})
			}
		})
	})
}
