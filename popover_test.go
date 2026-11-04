package mizan

import (
	"fmt"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// A Select longer than its panel scrolls inside it: the last option, gone
// to with End, shows within the panel, not past it.
func TestLongSelectScrolls(t *testing.T) {
	var v int
	var opts []Option[int]
	for i := range 30 {
		opts = append(opts, Option[int]{Value: i, Label: fmt.Sprint("Option ", i)})
	}
	tt := ui.NewTester(func(c *ui.Context) {
		Root(c, Options{})
		ui.Column(c).Padding(20).Children(func() { Select(c.Key("s"), &v, opts, SelectOptions{}) })
	}, 400, 800)
	tt.Advance(200 * time.Millisecond)
	if err := tt.Click("Option 0"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(400 * time.Millisecond)
	first, _ := tt.Find("Option 1")
	tt.Key(0, ui.KeyEnd)
	tt.Advance(400 * time.Millisecond)
	last, ok := tt.Find("Option 29")
	if !ok || last.Y > first.Y+300 {
		t.Fatalf("the last option is at %v, past the panel (its first options at %v)", last.Y, first.Y)
	}
}

// A menu longer than its panel scrolls inside it too.
func TestLongMenuScrolls(t *testing.T) {
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		Root(c, Options{})
		ui.Column(c).Padding(20).Children(func() {
			b := Button(c.Key("b"), "More", ButtonOptions{})
			Menu(c, b, &open, func(m *MenuList) {
				for i := range 30 {
					m.Item(fmt.Sprint("Item ", i), ItemOptions{})
				}
			})
		})
	}, 400, 900)
	tt.Advance(400 * time.Millisecond)
	first, _ := tt.Find("Item 0")
	last, _ := tt.Find("Item 29")
	if last.Y-first.Y < menuMaxH {
		t.Fatalf("the items span %v: the menu is not long enough to test", last.Y-first.Y)
	}
	tt.Key(0, ui.KeyEnd)
	tt.Advance(400 * time.Millisecond)
	if last, _ = tt.Find("Item 29"); last.Y > first.Y+menuMaxH {
		t.Fatalf("the last item is at %v, past the panel (its first at %v)", last.Y, first.Y)
	}
}
