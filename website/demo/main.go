// Demo shows the previews of the pages of the docs: the website builds it
// for browsers (website/docs-demo.ts) and shows a demo in a frame wherever a
// page of docs/ has a comment
//
//	<!-- preview button 120 -->
//
// naming a demo and the frame's height in pixels. It runs on desktops too:
//
//	go run ./website/demo          # every demo, in a sidebar
//	go run ./website/demo button   # the button's
package main

import (
	"log"
	"os"
	"slices"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/mizan-go"
)

// demos are the demos by name, the page of docs/ they preview: "button"
// is docs/button.md's.
var demos = map[string]func(c *ui.Context, t *mz.Tokens){}

// demo adds the demo of a page. Demos keep their state in the variables
// their function closes over, as an app keeps its state in a value.
func demo(name string, view func(c *ui.Context, t *mz.Tokens)) {
	if _, ok := demos[name]; ok {
		panic("demo: two demos of " + name)
	}
	demos[name] = view
}

func main() {
	name := ""
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	content := index
	if view, ok := demos[name]; ok {
		content = func(c *ui.Context) { preview(c, view) }
	} else if name != "" {
		log.Fatalf("demo: no demo of %q", name)
	}
	mygo.App.WhenReady(func() {
		mygo.NewWindow(mygo.WindowOptions{
			Title:   "Mizan",
			Width:   720,
			Height:  480,
			Content: ui.View(content),
		})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

// preview shows a demo in the middle of the window, in Mizan's theme, as a
// page of the docs frames it. A demo taking the width stretches itself
// (AlignSelf(Stretch)).
func preview(c *ui.Context, view func(c *ui.Context, t *mz.Tokens)) {
	t := mz.Root(c, mz.Options{})
	ui.Column(c).Fill().Center().Gap(12).Padding(20).Background(t.Colors.Bg).Children(func() { view(c, t) })
	mz.Toaster(c)
}

// chosen is the demo the index shows.
var chosen = 0

// index shows every demo, chosen in a sidebar, to try them all at once.
func index(c *ui.Context) {
	names := make([]string, 0, len(demos))
	for name := range demos {
		names = append(names, name)
	}
	slices.Sort(names)
	t := mz.Root(c, mz.Options{})
	ui.Row(c).Fill().Background(t.Colors.Bg).Children(func() {
		ui.Scroll(c).Width(200).Children(func() {
			ui.Column(c).Padding(8).Gap(2).Children(func() {
				for i, name := range names {
					o := mz.ButtonOptions{Size: mz.Sm}
					if i == chosen {
						o.Color = t.Colors.Accent
					}
					if mz.Button(c.Key(name), name, o).AlignSelf(ui.Stretch).Clicked() {
						chosen = i
					}
				}
			})
		})
		ui.Column(c).Grow(1).Children(func() {
			if chosen < len(names) {
				mz.Text(c, "docs/"+names[chosen]+".md", mz.TCaption).TextColor(t.Colors.Muted).Padding(12)
				ui.Column(c).Grow(1).Center().Gap(12).Padding(20).Children(func() { demos[names[chosen]](c, t) })
			}
		})
	})
	mz.Toaster(c)
}

// status is a line under a demo saying what it did.
func status(c *ui.Context, t *mz.Tokens, text string) ui.Element {
	return mz.Text(c, text, mz.TCaption).TextColor(t.Colors.Muted)
}

// Icons of the demos, 16 wide on a 15 box as Mizan's own.
var (
	iconSearch = mz.Svg(16, 15, "M7 12.5a5.5 5.5 0 1 0 0-11 5.5 5.5 0 0 0 0 11zM14.5 14.5l-3.6-3.6")
	iconMore   = mz.Svg(16, 15, "M3.5 8h.01M8 8h.01M12.5 8h.01")
	iconShare  = mz.Svg(16, 15, "M8 10V2.5M5 5.5l3-3 3 3M3 9.5v3A1.5 1.5 0 0 0 4.5 14h7a1.5 1.5 0 0 0 1.5-1.5v-3")
	iconEdit   = mz.Svg(16, 15, "M10.5 2.5l3 3-8 8h-3v-3z")
	iconTrash  = mz.Svg(16, 15, "M2.5 4.5h11M6 4.5V3h4v1.5M4 4.5l.7 9h6.6l.7-9")
	iconInbox  = mz.Svg(16, 18, "M1.5 9.5h3.5l1 2h4l1-2h3.5M3.5 3.5h9l2 6v3a1 1 0 0 1-1 1h-11a1 1 0 0 1-1-1v-3z")
	iconBold   = mz.Svg(16, 15, "M5 3h3.8a2.5 2.5 0 0 1 0 5H5zM5 8h4.3a2.5 2.5 0 0 1 0 5H5z")
	iconItalic = mz.Svg(16, 15, "M7 3h5M4 13h5M9.5 3l-3 10")
	iconUnder  = mz.Svg(16, 15, "M5 3v4.5a3 3 0 0 0 6 0V3M4 13.5h8")
	iconAlignL = mz.Svg(16, 15, "M3 4h10M3 8h6M3 12h8")
	iconAlignC = mz.Svg(16, 15, "M3 4h10M5 8h6M4 12h8")
	iconPin    = mz.Svg(16, 15, "M6 2.5h4M7 2.5v4L4.5 9h7L9 6.5v-4M8 9v4.5")
)
