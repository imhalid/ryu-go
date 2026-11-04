package mizan

import (
	"fmt"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// The copy button copies, says so, and turns back 1.5 s later.
func TestCopyButton(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		Root(c, Options{})
		ui.Column(c).Padding(20).Children(func() { CopyButton(c.Key("copy"), "the link", CopyOptions{Label: "Copy link"}) })
	}, 300, 200)
	tt.Advance(100 * time.Millisecond)
	if err := tt.Click("Copy link"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(100 * time.Millisecond)
	if tt.Clipboard() != "the link" {
		t.Fatalf("clipboard %q", tt.Clipboard())
	}
	if _, ok := tt.Find("Copied"); !ok {
		t.Fatal("the button doesn't say it copied")
	}
	tt.Advance(1600 * time.Millisecond)
	if _, ok := tt.Find("Copy link"); !ok {
		t.Fatal("the button didn't turn back")
	}
}

// Held, a slider shows its value over the knob.
func TestSliderValueLabel(t *testing.T) {
	v := 50.0
	tt := ui.NewTester(func(c *ui.Context) {
		Root(c, Options{})
		ui.Column(c).Padding(40).Width(300).Children(func() { Slider(c.Key("s"), &v, 0, 100, SliderOptions{Label: "Level", Step: 25}) })
	}, 400, 200)
	tt.Advance(100 * time.Millisecond)
	if tt.HasText("50") {
		t.Fatal("the value shows before the slider is held")
	}
	r, _ := tt.Find("Level")
	tt.Press(r.X+r.W/2, r.Y+r.H/2)
	tt.Advance(200 * time.Millisecond)
	if !tt.HasText("50") {
		t.Fatalf("held, the slider shows %v", tt.Texts())
	}
	tt.Release(r.X+r.W/2, r.Y+r.H/2)
	tt.Advance(300 * time.Millisecond)
	if tt.HasText("50") {
		t.Fatal("let go, the value still shows")
	}
}

// A field shaken swings sideways, and comes back where it was.
func TestFieldShake(t *testing.T) {
	n := 0
	tt := ui.NewTester(func(c *ui.Context) {
		Root(c, Options{})
		ui.Column(c).Padding(20).Children(func() {
			Field(c, FieldOptions{Label: "Email", Shake: n}, func() { ui.Text(c, "control") })
		})
	}, 300, 200)
	tt.Advance(100 * time.Millisecond)
	at, _ := tt.Find("Email")
	n++
	tt.Frame()
	moved := false
	for range 10 {
		tt.Advance(30 * time.Millisecond)
		if r, _ := tt.Find("Email"); r.X != at.X {
			moved = true
		}
	}
	tt.Advance(500 * time.Millisecond)
	if r, _ := tt.Find("Email"); !moved || r.X != at.X {
		t.Fatalf("moved %v, at %v, was %v", moved, r.X, at.X)
	}
}

// A chat keeps the latest message in view; scrolled up, a new one shows
// on the Latest button, which takes the list down to it.
func TestChatLatest(t *testing.T) {
	var msgs []ChatMessage
	for i := range 20 {
		msgs = append(msgs, ChatMessage{ID: fmt.Sprint(i), Author: "Maya", Text: fmt.Sprint("Message ", i), Time: "09:00"})
	}
	tt := ui.NewTester(func(c *ui.Context) {
		Root(c, Options{})
		ui.Box(c).Size(400, 300).Children(func() { Chat(c.Key("chat"), msgs, ChatOptions{}).Fill() })
	}, 400, 300)
	tt.Advance(800 * time.Millisecond)
	if r, _ := tt.Find("Message 19"); r.Y > 300 || r.Y < 0 {
		t.Fatalf("the latest message is at %v, out of view", r.Y)
	}
	// up, a new message doesn't pull the list down
	for range 10 {
		tt.Wheel(200, 150, 0, -120)
		tt.Advance(50 * time.Millisecond)
	}
	msgs = append(msgs, ChatMessage{ID: "new", Author: "Maya", Text: "Message new", Time: "09:01"})
	tt.Frame()
	tt.Advance(600 * time.Millisecond)
	if !tt.HasText("1 new") {
		t.Fatalf("no button for the message not seen: %v", tt.Texts())
	}
	if err := tt.Click("1 new"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(900 * time.Millisecond)
	if r, _ := tt.Find("Message new"); r.Y > 300 || r.Y < 0 {
		t.Fatalf("the new message is at %v, out of view", r.Y)
	}
}
