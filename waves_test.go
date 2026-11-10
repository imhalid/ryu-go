package ryugo

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// waveCase builds a component, and act does to it what sends a wave.
type waveCase struct {
	name  string
	build func(c *ui.Context)
	act   func(t *testing.T, tt *ui.Tester)
	// from is where the wave should leave from, as shares of the box; nil
	// for anywhere
	from func(fx, fy float32) bool
}

// Dialogs, Sheets, a Switch, Segmented, Chat and a dropped card each send
// one wave as something lands in them, which settles and stops asking for
// frames; without Options.Waves or with Reduce Motion, none.
func TestElementWaves(t *testing.T) {
	var sent [][2]float32
	defer func(f func(fx, fy float32)) { waveSent = f }(waveSent)
	waveSent = func(fx, fy float32) { sent = append(sent, [2]float32{fx, fy}) }

	cases := func() []waveCase {
		var dialog, sheet, on bool
		var view int
		msgs := []ChatMessage{{ID: "1", Author: "Ada", Text: "Hi"}}
		steps := []string{"One", "Two", "Three"}
		var list ui.ListState
		return []waveCase{
			{name: "dialog",
				build: func(c *ui.Context) {
					Dialog(c.Key("d"), &dialog, DialogOptions{Title: "Rename"}, func() { ui.Text(c, "Body") })
				},
				act:  func(t *testing.T, tt *ui.Tester) { dialog = true; tt.Frame() },
				from: func(fx, fy float32) bool { return fx == 0.5 && fy == 0.5 }},
			{name: "sheet",
				build: func(c *ui.Context) {
					Sheet(c.Key("s"), &sheet, DialogOptions{Title: "Details"}, func() { ui.Text(c, "Body") })
				},
				act:  func(t *testing.T, tt *ui.Tester) { sheet = true; tt.Frame() },
				from: func(fx, fy float32) bool { return fx == 0 && fy == 0.5 }},
			{name: "switch",
				build: func(c *ui.Context) { Switch(c.Key("sw"), &on, SwitchOptions{Label: "Sync"}) },
				act: func(t *testing.T, tt *ui.Tester) {
					if err := tt.Click("Sync"); err != nil {
						t.Fatal(err)
					}
				},
				from: func(fx, fy float32) bool { return fx > 0.6 && fy == 0.5 }},
			{name: "segmented",
				build: func(c *ui.Context) {
					Segmented(c.Key("seg"), &view, []string{"List", "Grid"}, SegmentedOptions{Label: "View"})
				},
				act: func(t *testing.T, tt *ui.Tester) {
					if err := tt.Click("Grid"); err != nil {
						t.Fatal(err)
					}
				},
				from: func(fx, fy float32) bool { return fx > 0.5 }},
			{name: "chat",
				build: func(c *ui.Context) {
					ui.Box(c).Size(300, 200).Children(func() { Chat(c.Key("chat"), msgs, ChatOptions{}) })
				},
				act: func(t *testing.T, tt *ui.Tester) {
					msgs = append(msgs, ChatMessage{ID: "2", Mine: true, Text: "Hello"})
				},
				from: func(fx, fy float32) bool { return fx == 1 && fy == 1 }},
			{name: "drop",
				build: func(c *ui.Context) {
					Sortable(c.Key("steps"), &list, &steps, SortableOptions{
						Label: func(i int) string { return steps[i] },
						Key:   func(i int) any { return steps[i] },
					}, func(s string) { ui.Text(c, s).Grow(1) })
				},
				act: func(t *testing.T, tt *ui.Tester) {
					a, _ := tt.Find("Three")
					b, _ := tt.Find("One")
					x, y := a.X+a.W/2, a.Y+a.H/2
					tt.Press(x, y)
					for i := 1; i <= 10; i++ {
						tt.Move(x, y+(b.Y+2-y)*float32(i)/10)
						tt.Advance(16 * time.Millisecond)
					}
					tt.Release(x, b.Y+2)
				}},
		}
	}

	for _, mode := range []struct {
		name          string
		waves, reduce bool
		want          int
	}{{"on", true, false, 1}, {"off", false, false, 0}, {"reduce", true, true, 0}} {
		for _, k := range cases() {
			t.Run(mode.name+"/"+k.name, func(t *testing.T) {
				sent = nil
				tt := ui.NewTester(func(c *ui.Context) {
					Root(c, Options{Waves: mode.waves})
					ui.Column(c).Padding(20).Children(func() { k.build(c) })
				}, 640, 480)
				tt.SetPreferences(ui.Preferences{ReduceMotion: mode.reduce})
				tt.Advance(200 * time.Millisecond)
				if len(sent) != 0 {
					t.Fatalf("%d waves before anything happened", len(sent))
				}
				k.act(t, tt)
				tt.Advance(time.Second)
				if len(sent) != mode.want {
					t.Fatalf("%d waves, want %d", len(sent), mode.want)
				}
				if len(sent) == 1 && k.from != nil && !k.from(sent[0][0], sent[0][1]) {
					t.Errorf("the wave left from %v", sent[0])
				}
				tt.Advance(3 * time.Second)
				if tt.FramePending() {
					t.Error("still asking for frames after the wave")
				}
			})
		}
	}
}

// A Dialog closed by its Danger button burns away over burnFor, the
// pointer going through it to the page; closed by Cancel, without
// Options.Waves or with Reduce Motion, it fades out as quickly as ever.
func TestDialogBurn(t *testing.T) {
	run := func(t *testing.T, waves, reduce bool, button string) (at400, at2500 bool) {
		open := true
		tt := ui.NewTester(func(c *ui.Context) {
			Root(c, Options{Waves: waves})
			ui.Column(c).Padding(20).Children(func() {
				Dialog(c.Key("d"), &open, DialogOptions{Title: "Delete the report?",
					Footer: func() {
						if Button(c.Key("cancel"), "Cancel", ButtonOptions{}).Clicked() {
							open = false
						}
						if Button(c.Key("ok"), "Delete report", ButtonOptions{Danger: true}).Clicked() {
							open = false
						}
					}}, func() { ui.Text(c, "This cannot be undone.") })
			})
		}, 720, 520)
		tt.SetPreferences(ui.Preferences{ReduceMotion: reduce})
		tt.Advance(time.Second)
		if err := tt.Click(button); err != nil {
			t.Fatal(err)
		}
		if open {
			t.Fatal("the dialog didn't close")
		}
		tt.Advance(400 * time.Millisecond)
		_, at400 = tt.Find("Delete the report?")
		tt.Advance(2100 * time.Millisecond)
		_, at2500 = tt.Find("Delete the report?")
		if tt.FramePending() {
			t.Error("still asking for frames once closed")
		}
		return
	}
	for _, v := range []struct {
		name          string
		waves, reduce bool
		button        string
		burns         bool
	}{
		{"danger", true, false, "Delete report", true},
		{"cancel", true, false, "Cancel", false},
		{"no waves", false, false, "Delete report", false},
		{"reduce motion", true, true, "Delete report", false},
	} {
		t.Run(v.name, func(t *testing.T) {
			burning, after := run(t, v.waves, v.reduce, v.button)
			if burning != v.burns {
				t.Errorf("still showing 400 ms after closing: %v, want %v", burning, v.burns)
			}
			if after {
				t.Error("still showing once the burn is over")
			}
		})
	}
}
