package ryugo

import (
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// crossings counts the zero crossings of s in [from, to) seconds.
func crossings(s []float32, from, to float64) int {
	n := 0
	for i := int(from*soundRate) + 1; i < int(to*soundRate) && i < len(s); i++ {
		if (s[i-1] < 0) != (s[i] < 0) {
			n++
		}
	}
	return n
}

// Each drop starts and ends at silence, stays in range, and rises in pitch.
func TestDropSamples(t *testing.T) {
	for k, d := range drops {
		s := d.samples()
		if len(s) == 0 || len(s) > soundRate/4 {
			t.Fatalf("tone %d: %d samples", k, len(s))
		}
		for i, v := range s {
			if math.IsNaN(float64(v)) || v > 1 || v < -1 {
				t.Fatalf("tone %d: sample %d is %v", k, i, v)
			}
		}
		if a, z := math.Abs(float64(s[0])), math.Abs(float64(s[len(s)-1])); a > 0.01 || z > 0.001 {
			t.Errorf("tone %d: starts at %v, ends at %v: it would click", k, a, z)
		}
		if early, late := crossings(s, 0, 0.01), crossings(s, 0.05, 0.06); late <= early {
			t.Errorf("tone %d: %d crossings early, %d late: the pitch doesn't rise", k, early, late)
		}
	}
}

// A pressed Button plays its tone from the side pressed, by keyboard from
// the middle, with Reduce Motion too; without Sounds it is silent.
func TestButtonSound(t *testing.T) {
	type played struct {
		k   tone
		pan float64
	}
	var got []played
	defer func(f func(tone, float64, float64)) { playSound = f }(playSound)
	playSound = func(k tone, pan, speed float64) {
		if speed < 0.96 || speed > 1.04 {
			t.Errorf("speed %v", speed)
		}
		got = append(got, played{k, pan})
	}

	run := func(sounds, reduce bool, o ButtonOptions, act func(tt *ui.Tester, r ui.Rect)) []played {
		got = nil
		tt := ui.NewTester(func(c *ui.Context) {
			Root(c, Options{Sounds: sounds})
			ui.Column(c).Padding(20).Children(func() { Button(c.Key("b"), "Save changes", o) })
		}, 240, 80)
		tt.SetPreferences(ui.Preferences{ReduceMotion: reduce})
		tt.Advance(100 * time.Millisecond)
		r, _ := tt.Find("Save changes")
		act(tt, r)
		tt.Advance(100 * time.Millisecond)
		return got
	}
	at := func(f float32) func(tt *ui.Tester, r ui.Rect) {
		return func(tt *ui.Tester, r ui.Rect) {
			x, y := r.X+r.W*f, r.Y+r.H/2
			tt.Press(x, y)
			tt.Advance(40 * time.Millisecond)
			tt.Release(x, y)
		}
	}
	keyboard := func(tt *ui.Tester, r ui.Rect) {
		tt.Key(0, ui.KeyTab)
		tt.Advance(50 * time.Millisecond)
		tt.Key(0, ui.KeySpace)
	}

	if p := run(false, false, ButtonOptions{}, at(0.5)); len(p) != 0 {
		t.Errorf("without Sounds it played %v", p)
	}
	if p := run(true, false, ButtonOptions{}, at(0.1)); len(p) != 1 || p[0].k != tonePlain || p[0].pan >= -0.2 {
		t.Errorf("pressed on the left: %v", p)
	}
	if p := run(true, false, ButtonOptions{Color: themeColors[0].Accent, Variant: Solid}, at(0.9)); len(p) != 1 || p[0].k != toneSolid || p[0].pan <= 0.2 {
		t.Errorf("solid pressed on the right: %v", p)
	}
	if p := run(true, false, ButtonOptions{Danger: true}, at(0.5)); len(p) != 1 || p[0].k != toneDanger {
		t.Errorf("danger: %v", p)
	}
	if p := run(true, false, ButtonOptions{}, keyboard); len(p) != 1 || p[0].pan != 0 {
		t.Errorf("by keyboard: %v", p)
	}
	if p := run(true, true, ButtonOptions{}, at(0.5)); len(p) != 1 {
		t.Errorf("with Reduce Motion: %v", p)
	}
	if p := run(true, false, ButtonOptions{Disabled: true}, at(0.5)); len(p) != 0 {
		t.Errorf("disabled: %v", p)
	}
}
