package ryugo

import (
	"image/png"
	"os"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// A split button: a click on the action is the action's, one on the arrow the arrow's; Tab reaches it once, and
// Right moves to the arrow, whose Enter is its own; assistive technology hears the action by its label and the arrow
// by the menu's; disabled, neither does anything.
func TestSplitButton(t *testing.T) {
	var acts, menus int
	disabled := false
	tt := ui.NewTester(func(c *ui.Context) {
		Root(c, Options{})
		ui.Column(c).Padding(20).Children(func() {
			sb := SplitButton(c.Key("new"), "New page", SplitButtonOptions{Color: From(c).Colors.Accent, Disabled: disabled, MenuLabel: "From a template"})
			if sb.Button.Clicked() {
				acts++
			}
			if sb.Arrow.Clicked() {
				menus++
			}
		})
	}, 320, 100)
	tt.Advance(100 * time.Millisecond)
	// by the keys: one stop of Tab, Right to the arrow, Enter its own
	tt.Key(0, ui.KeyTab)
	if !tt.Focused("New page") {
		t.Fatal("Tab did not reach the action")
	}
	tt.Key(0, ui.KeyRight)
	if !tt.Focused("From a template") {
		t.Fatal("Right did not move to the arrow")
	}
	tt.Key(0, ui.KeyEnter)
	tt.Advance(100 * time.Millisecond)
	if acts != 0 || menus != 1 {
		t.Fatalf("Enter on the arrow: the action %d, the arrow %d", acts, menus)
	}
	if err := tt.Click("New page"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("From a template"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(100 * time.Millisecond)
	if acts != 1 || menus != 2 {
		t.Fatalf("clicks: the action %d, the arrow %d", acts, menus)
	}
	act, _ := tt.Find("New page")
	arrow, _ := tt.Find("From a template")
	if arrow.X < act.X+act.W-1 || arrow.H != act.H {
		t.Fatalf("the arrow is not beside the action: %v, %v", act, arrow)
	}
	if os.Getenv("RYUGO_SHOT") != "" {
		f, _ := os.Create(os.Getenv("RYUGO_SHOT"))
		png.Encode(f, tt.Image())
		f.Close()
	}
	disabled = true
	tt.Advance(100 * time.Millisecond)
	tt.Click("New page")
	tt.Click("From a template")
	tt.Advance(100 * time.Millisecond)
	if acts != 1 || menus != 2 {
		t.Fatalf("disabled, clicked: the action %d, the arrow %d", acts, menus)
	}
}
