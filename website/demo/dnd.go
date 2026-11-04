package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/ryu-go"
)

// card is a task on the board; col is the column it is in.
type card struct {
	Title, Tag string
	col        int
}

func init() {
	steps := []string{"Import data", "Map columns", "Validate", "Publish"}
	var list ui.ListState
	demo("sortable", func(c *ui.Context, t *mz.Tokens) {
		ui.Column(c).Width(320).Children(func() {
			mz.Sortable(c.Key("steps"), &list, &steps, mz.SortableOptions{
				Label: func(i int) string { return steps[i] },
				Key:   func(i int) any { return steps[i] },
			}, func(step string) {
				ui.Text(c, step).Grow(1)
			})
		})
		status(c, t, strings.Join(steps, " → "))
	})

	cards := []*card{
		{Title: "Sign-in screen", Tag: "Design", col: 0},
		{Title: "Report export", Tag: "Web", col: 0},
		{Title: "Notification permissions", Tag: "iOS", col: 1},
		{Title: "Dark theme", Tag: "Design", col: 2},
	}
	// move puts k in column col, before the index-th of the column's other
	// cards (at its end past them).
	move := func(k *card, col, index int) {
		rest := slices.DeleteFunc(slices.Clone(cards), func(x *card) bool { return x == k })
		at, n := len(rest), 0
		for i, x := range rest {
			if x.col != col {
				continue
			}
			if n == index {
				at = i
				break
			}
			n++
			at = i + 1
		}
		k.col = col
		cards = slices.Insert(rest, at, k)
	}
	demo("drag-and-drop", func(c *ui.Context, t *mz.Tokens) {
		ui.Row(c).Gap(10).Width(620).AlignItems(ui.Stretch).Children(func() {
			for col, title := range []string{"To do", "In progress", "Done"} {
				n := 0
				for _, k := range cards {
					if k.col == col {
						n++
					}
				}
				ui.Column(c.Key(col)).Grow(1).Basis(0).MinWidth(0).Children(func() {
					mz.DropArea(c, mz.DropAreaOptions{Title: title, Count: fmt.Sprint(n), Fill: true}, func() {
						for _, k := range cards {
							if k.col != col {
								continue
							}
							mz.DragCard(c.Key(k.Title), k, func() {
								ui.Text(c, k.Title).Wrap()
								ui.Row(c).Children(func() { mz.Badge(c, k.Tag, mz.ToneNeutral) })
							})
						}
					}, func(k *card, index int) { move(k, col, index) })
				})
			}
		})
	})

	type file struct{ Name, Size string }
	files := []file{{"Q3-report.pdf", "2.4 MB"}, {"patient-list-october.xlsx", "860 KB"}}
	demo("drop-zone", func(c *ui.Context, t *mz.Tokens) {
		ui.Column(c).Gap(8).Width(380).Children(func() {
			for _, p := range mz.DropZone(c.Key("drop"), mz.DropZoneOptions{Title: "Drop files here", Hint: "PDF, XLSX · up to 10 MB"}) {
				files = append(files, file{Name: p[strings.LastIndexAny(p, `/\`)+1:]})
			}
			for i := 0; i < len(files); i++ {
				if mz.FileItem(c.Key(files[i].Name), files[i].Name, files[i].Size) {
					files = append(files[:i], files[i+1:]...)
					i--
				}
			}
		})
	})
}
