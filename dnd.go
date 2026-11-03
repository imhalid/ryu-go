package mizan

import (
	"slices"
	"time"

	"github.com/egoist/mygo/ui"
)

// Sürükle-bırakın dili iki parçadır: sürüklenen şey kabarık malzemedir
// (tutulabilir), bırakılan yer gömük kuyudur. Kuyu, üstüne kendi aldığı
// bir şey gelince Mizan'ın dosya bölgesindeki gibi düz vurgu kenarı ve
// %8 vurguya çalan zemin alır; sürüklenenin yeri %40 soluk kalır.

var (
	iconUpload = Svg(16, 20, "M8 10.5V2.5M5 5.5l3-3 3 3M2.5 10v2A1.5 1.5 0 004 13.5h8a1.5 1.5 0 001.5-1.5v-2")
	iconFile   = Svg(16, 14, "M9 1.5H4.5a1.5 1.5 0 0 0-1.5 1.5v10a1.5 1.5 0 0 0 1.5 1.5h7a1.5 1.5 0 0 0 1.5-1.5V5.5zM9 1.5v4h4")
)

// dropWell, bırakma yerini boyar: normalde kuyu, üstüne bir şey gelince
// vurgu kenarı ve vurguya çalan zemin.
func (t *Tokens) dropWell(e ui.Element, over bool) {
	c := t.Colors
	if over {
		e.Background(Mix(c.Well, c.ColorAccent500, 0.08)).Border(1, c.ColorAccent500)
	} else {
		e.Background(c.Well).Border(1, c.WellBorder)
	}
	e.Transition(ui.ElementTransition{Duration: 150 * time.Millisecond, Colors: true})
}

// grip, tutamak: iki sütun üç satır nokta.
func grip(c *ui.Context, color ui.Color) {
	ui.Row(c).Gap(3).Shrink(0).Children(func() {
		for range 2 {
			ui.Column(c).Gap(3).Children(func() {
				for range 3 {
					ui.Box(c).Size(2.5, 2.5).Radius(1.25).Background(color)
				}
			})
		}
	})
}

// tile, sürüklenen kabarık karodur: düğmenin malzemesi, kontrol köşesi.
func (t *Tokens) tile(e ui.Element, hovered bool) {
	cl := t.Colors
	e.Radius(RadiusControl).Cursor(ui.CursorGrab)
	t.raised(e, t.controlGradient(cl.ControlTop, cl.ControlBottom), cl.ControlBorder, t.Shadows.ControlShadow,
		raisedState{hovered: hovered})
}

// SortableOptions, Sortable'ın seçenekleridir.
type SortableOptions struct {
	// Label, satırın metnidir: erişilebilirlik adı ve klavyeyle bulma.
	Label func(i int) string
	// Key, satırın kimliğidir; yoksa sırası. Satırların durumu (yerel
	// değerler, odak) taşınırken korunsun diye verin.
	Key func(i int) any
}

// Sortable, satırları sürüklenerek sıralanan listedir: her satır tutamaklı
// kabarık bir karo, bırakılacak yeri vurgu renginde çizgi gösterir; liste
// *items'ı kendisi yeniden dizer. s çerçeveler boyunca aynı kalmalıdır.
//
//	mizan.Sortable(c, &app.list, &app.tasks, mizan.SortableOptions{}, func(t Task) {
//		ui.Text(c, t.Title)
//	})
func Sortable[T any](c *ui.Context, s *ui.ListState, items *[]T, o SortableOptions, row func(T)) ui.Element {
	t := From(c)
	cl := t.Colors
	s.Label, s.Key = o.Label, o.Key
	s.Reorder = func(rows []int, to int) { *items = move(*items, rows, to) }
	return ui.List(c, s, len(*items), func(i int) {
		e := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(12).Padding(10, 14, 10, 12).MinHeight(t.Metrics.ControlH + 8)
		t.tile(e, e.Hovered())
		e.TextColor(cl.Text).FontSize(FsMd).Children(func() {
			grip(c, cl.Muted)
			ui.Row(c).Grow(1).MinWidth(0).AlignItems(ui.Center).Gap(8).Children(func() { row((*items)[i]) })
		})
	}).Gap(6).Padding(2)
}

// move, rows satırlarını to'nun önüne taşır (sıraları korunur).
func move[T any](items []T, rows []int, to int) []T {
	moved := make([]T, 0, len(rows))
	rest := make([]T, 0, len(items))
	at := to
	for i, v := range items {
		if slices.Contains(rows, i) {
			moved = append(moved, v)
			if i < to {
				at--
			}
		} else {
			rest = append(rest, v)
		}
	}
	return slices.Insert(rest, at, moved...)
}

// DragCard, başka bir yere (DropArea) sürüklenen kabarık karodur; value
// bırakılan yere gider. Sürüklenirken yeri soluk kalır, kopyası işaretçiyi
// izler.
//
//	mizan.DragCard(c.Key(task.ID), task, func() { ui.Text(c, task.Title) })
func DragCard(c *ui.Context, value any, body func()) ui.Element {
	t := From(c)
	e := ui.Column(c).FillWidth().Gap(4).Padding(10, 12).MinWidth(0).Drag(value)
	t.tile(e, e.Hovered())
	e.TextColor(t.Colors.Text).FontSize(FsMd).Children(body)
	if e.Dragging() {
		e.Opacity(0.4)
	}
	return e
}

// DropAreaOptions, DropArea'nın seçenekleridir.
type DropAreaOptions struct {
	Title string
	// Count, başlığın yanındaki nötr sayıdır (boşsa gösterilmez).
	Count string
	// Accept, sürüklenen değeri alıp almayacağını söyler (yoksa T'nin
	// her değeri): aynı sütundan gelen kartı almamak gibi.
	Accept func(v any) bool
	// Fill, sütunu kabının boyuna uzatır: yan yana sütunlar eşit boyda
	// durur, boş sütunun da bırakılacak yeri olur.
	Fill bool
}

// DropArea, T türünde değer alan gömük sütundur (pano sütunu, klasör):
// üstüne aldığı bir değer gelince vurgulanır, bırakılan değeri döndürür.
//
//	if task, ok := mizan.DropArea[*Task](c.Key("done"), mizan.DropAreaOptions{Title: "Bitti"}, func() {
//		...
//	}); ok {
//		task.Done = true
//	}
func DropArea[T any](c *ui.Context, o DropAreaOptions, body func()) (T, bool) {
	t := From(c)
	cl := t.Colors
	e := ui.Column(c).Gap(8).Padding(10).Radius(RadiusLg).MinWidth(0)
	if o.Fill {
		e.Grow(1)
	}
	v, over := ui.DragOver[T](e)
	if over && o.Accept != nil && !o.Accept(v) {
		over = false
	}
	t.dropWell(e, over)
	t.Shadows.WellShadow.Apply(e)
	e.Children(func() {
		if o.Title != "" {
			ui.Row(c).AlignItems(ui.Center).Gap(8).Padding(2, 4, 4, 4).Children(func() {
				ui.Text(c, o.Title).TextColor(cl.Text2).FontSize(FsSm).FontWeight(600)
				if o.Count != "" {
					Badge(c, o.Count, ToneNeutral)
				}
			})
		}
		body()
	})
	dropped, ok := ui.Drop[T](e)
	if ok && o.Accept != nil && !o.Accept(dropped) {
		var zero T
		return zero, false
	}
	return dropped, ok
}

// DropZoneOptions, DropZone'un seçenekleridir.
type DropZoneOptions struct {
	// Title, kalın satır ("Dosyaları buraya bırakın"); Hint altındaki
	// soluk açıklama ("PNG, PDF · en çok 10 MB").
	Title string
	Hint  string
	// Browse verilirse bölge tıklanabilir olur ve tıklanınca çağrılır
	// (örneğin mygo.Dialog.Open ile dosya seçici).
	Browse   func()
	Disabled bool
}

// DropZone, başka uygulamalardan (Finder, Explorer) dosya alan kesikli
// kenarlı kuyudur: üstüne dosya gelince kenar düzleşip vurgu rengini alır.
// Bırakılan dosyaların yollarını döndürür, yoksa nil.
//
//	if paths := mizan.DropZone(c.Key("upload"), mizan.DropZoneOptions{Title: "Dosyaları bırakın"}); paths != nil {
//		app.add(paths)
//	}
func DropZone(c *ui.Context, o DropZoneOptions) []string {
	t := From(c)
	cl := t.Colors
	var e ui.Element
	if o.Browse != nil {
		e = ui.ButtonBase(c).Label(or(o.Title, "Dosya seç")).Cursor(ui.CursorPointer).Disabled(o.Disabled).FocusRing(false)
	} else {
		e = ui.Column(c)
	}
	e.FillWidth().AlignItems(ui.Center).Gap(4).Padding(24, 16).Radius(RadiusControl)
	over := !o.Disabled && e.FileDragOver()
	var files []string
	if !o.Disabled {
		files = e.DroppedFiles()
	}
	switch {
	case over:
		e.Background(Mix(cl.Well, cl.ColorAccent500, 0.08)).Border(1, cl.ColorAccent500)
	case o.Browse != nil && e.FocusVisible():
		e.Background(cl.Well).Border(1, cl.ColorAccent500).BorderStyle(ui.BorderDashed)
		t.Shadows.Ring.Apply(e)
	case e.Hovered() && !o.Disabled:
		e.Background(cl.Well).Border(1, cl.Muted).BorderStyle(ui.BorderDashed)
	default:
		e.Background(cl.Well).Border(1, Mix(ui.Color{}, cl.Muted, 0.45)).BorderStyle(ui.BorderDashed)
	}
	e.Transition(ui.ElementTransition{Duration: 150 * time.Millisecond, Colors: true})
	if o.Disabled {
		e.Opacity(0.4)
	}
	e.TextColor(cl.Muted).FontSize(FsSm).Children(func() {
		ui.Icon(c, iconUpload).Size(20, 20).TextColor(cl.Muted).Margin(0, 0, 4, 0)
		if o.Title != "" {
			ui.Text(c, o.Title).TextColor(cl.Text).FontSize(FsMd).FontWeight(600).TextAlign(ui.Center)
		}
		if o.Hint != "" {
			ui.Text(c, o.Hint).TextAlign(ui.Center).Wrap()
		}
	})
	if o.Browse != nil && !o.Disabled && e.Clicked() {
		o.Browse()
	}
	return files
}

// FileItem, eklenmiş bir dosyanın satırıdır: kuyu zemin, ad kırpılır,
// boyut tabular, sağda kaldır düğmesi. Kaldır'a basıldığında true döner.
func FileItem(c *ui.Context, name, size string) bool {
	t := From(c)
	cl := t.Colors
	removed := false
	ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(10).MinHeight(36).Padding(0, 4, 0, 12).Radius(RadiusItem).
		Background(cl.Well).FontSize(FsSm).Children(func() {
		ui.Icon(c, iconFile).TextColor(cl.Muted)
		ui.Text(c, name).TextColor(cl.Text).Grow(1).MinWidth(0).SingleLine().Ellipsis("…")
		if size != "" {
			ui.Text(c, size).TextColor(cl.Muted).FontFeatures("tnum").NoWrap()
		}
		removed = inputButton(c.Key("remove"), t, iconClear, "Kaldır").Margin(0).Clicked()
	})
	return removed
}
