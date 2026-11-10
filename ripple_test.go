package ryugo

import (
	"image"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// changed counts the pixels of r that differ between a and b.
func changed(a, b *image.RGBA, r image.Rectangle) int {
	n := 0
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			if a.RGBAAt(x, y) != b.RGBAAt(x, y) {
				n++
			}
		}
	}
	return n
}

// A press sends a wave through the button, which then settles back exactly
// as it was and stops asking for frames; a press by keyboard sends one too,
// and with less motion, or without Options.Waves, none goes.
func TestButtonRipple(t *testing.T) {
	rippleDraws = true
	defer func() { rippleDraws = false }()

	waves := true
	run := func(t *testing.T, reduce, focus bool, press func(tt *ui.Tester, r ui.Rect)) (moved, after int) {
		tt := ui.NewTester(func(c *ui.Context) {
			Root(c, Options{Waves: waves})
			ui.Column(c).Padding(20).Children(func() { Button(c.Key("b"), "Save changes", ButtonOptions{}) })
		}, 240, 80)
		tt.SetPreferences(ui.Preferences{ReduceMotion: reduce})
		tt.Advance(100 * time.Millisecond)
		r, _ := tt.Find("Save changes")
		tt.Move(r.X+r.W*0.25, r.Y+r.H/2)
		if focus {
			tt.Key(0, ui.KeyTab)
		}
		tt.Advance(300 * time.Millisecond)
		rest := image.NewRGBA(tt.Image().Bounds())
		copy(rest.Pix, tt.Image().Pix)
		box := image.Rect(int(r.X)-8, int(r.Y)-8, int(r.X+r.W)+8, int(r.Y+r.H)+8)

		press(tt, r)
		tt.Advance(80 * time.Millisecond)
		moved = changed(rest, tt.Image(), box)
		tt.Advance(2 * time.Second)
		if tt.FramePending() {
			t.Error("the button still asks for frames after the wave")
		}
		return moved, changed(rest, tt.Image(), box)
	}
	pointer := func(tt *ui.Tester, r ui.Rect) {
		x, y := r.X+r.W*0.25, r.Y+r.H/2
		tt.Press(x, y)
		tt.Advance(40 * time.Millisecond)
		tt.Release(x, y)
	}
	keyboard := func(tt *ui.Tester, r ui.Rect) {
		tt.Key(0, ui.KeySpace)
	}

	t.Run("pointer", func(t *testing.T) {
		moved, after := run(t, false, false, pointer)
		if moved < 40 {
			t.Errorf("the wave moved %d pixels", moved)
		}
		if after != 0 {
			t.Errorf("%d pixels differ once the wave settled", after)
		}
	})
	t.Run("keyboard", func(t *testing.T) {
		moved, after := run(t, false, true, keyboard)
		if moved < 40 {
			t.Errorf("the wave moved %d pixels", moved)
		}
		if after != 0 {
			t.Errorf("%d pixels differ once the wave settled", after)
		}
	})
	t.Run("less motion", func(t *testing.T) {
		moved, after := run(t, true, false, pointer)
		if moved != 0 || after != 0 {
			t.Errorf("with less motion %d pixels moved, %d after", moved, after)
		}
	})
	t.Run("waves off", func(t *testing.T) {
		waves = false
		defer func() { waves = true }()
		moved, after := run(t, false, false, pointer)
		if moved != 0 || after != 0 {
			t.Errorf("without Options.Waves %d pixels moved, %d after", moved, after)
		}
	})
}
