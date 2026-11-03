// Package mizan, Mizan tasarım dilinin MyGo native UI (package ui) için
// bileşenleridir. Web kütüphanesinin kodu taşınmaz; görsel dili taşınır:
// aynı token'lar (tokens.json'dan üretilir), aynı yüzey dili (kabarık
// kontrol, gömük alan, kart kenarı), aynı boyut kademeleri ve yazı ölçeği.
//
// Davranış (odak, klavye, erişilebilirlik) MyGo'nun görünümsüz bileşenlerinden
// gelir (ButtonBase, TextInputBase, SelectBase…); bu paket yalnız görünüşü verir.
//
//	func (a *app) view(c *ui.Context) {
//		t := mizan.Root(c, mizan.Options{})
//		mizan.Card(c, mizan.CardOptions{Title: "Rapor", Pad: true}, func() {
//			mizan.Input(c.Key("q"), &a.query, mizan.InputOptions{Placeholder: "Ara"})
//			if mizan.Button(c.Key("go"), "Uygula", mizan.ButtonOptions{Color: t.Colors.Accent, Variant: mizan.Solid}).Clicked() {
//				...
//			}
//		})
//	}
//
// Tema (Light, Dark), malzeme (SkinMizan, SkinBevel) ve boyut (Sm, Md, Lg)
// Root'ta pencerenin bütününe, Scope ile bir bölgeye verilir; en yakın
// kapsam kazanır. Bileşenler değerleri From(c) ile okur.
package mizan
