package ryugo

import (
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/shader"
)

// A press sends a wave through the button from where it landed: a front
// leaves the pointer at wavePress.speed and, as it reaches a point, the point
// swings out and back along the line from the press, a sine dying away
// (sin(f·t)·e^(−d·t), t counted from when the front came by). The surface,
// its outline and its label ripple together. Where the wave swells out, a
// very faint white light catches the bump (waveLook.glow), as a light from
// above would; the troughs stay as they are, so the button never darkens.
// A press by keyboard sends it from the middle. Other elements send the
// same wave, tuned to their size (waveLook), as something lands in them.

// waveLook tunes a wave to what it runs through: a button's is quick and
// fine; a panel's front runs fast and its swells are long and low, so a
// large surface moves as one, not as a jelly.
type waveLook struct {
	speed  float32 // DIPs a second
	freq   float32 // rad/s
	decay  float32 // 1/s
	amp    float32 // DIPs, at most
	glow   float32 // white mixed in at a crest's top, at most: barely there
	margin float32 // how far the effect draws around the element: the swing, and its shadows and rings, whole
}

var (
	// a pressed button: ~36 DIPs from crest to crest, e^−5 left after 0.7 s
	wavePress = waveLook{speed: 240, freq: 42, decay: 7, amp: 1.8, glow: 0.1, margin: 8}
	// a Dialog or a Sheet settling: one wide, slow swell across the panel
	wavePanel = waveLook{speed: 900, freq: 24, decay: 6, amp: 2.2, glow: 0.07, margin: 56}
	// a dropped card landing, its lifted shadow with it
	waveLand = waveLook{speed: 300, freq: 36, decay: 8, amp: 1.6, glow: 0.08, margin: 40}
	// a chat bubble growing out of its tail corner: its row grows open
	// under it, so the margin holds the whole bubble, or it would be cut
	waveBubble = waveLook{speed: 260, freq: 40, decay: 8, amp: 1.4, glow: 0.08, margin: 72}
	// a Switch's track as the knob lands, Segmented's track as the pill goes
	waveSmall = waveLook{speed: 160, freq: 56, decay: 10, amp: 1, glow: 0.1, margin: 8}
)

// waveMax is how many fronts run at once; one more takes the oldest's place.
const waveMax = 4

// rippleDraws is false in tests: the tester draws on the CPU, where the
// shader takes seconds for a click, and the gallery's tests press hundreds
// of buttons. The fronts are kept all the same; ripple_test.go turns it on.
var rippleDraws = !testing.Testing()

var rippleShader = shader.MustCompile(`package main

var Fronts [4]vec4 // x, y, seconds since the press, 1 for a live front
var Speed float
var Freq float
var Decay float
var Amp float
var Glow float

func LayerEffect(pos vec2) vec4 {
	d := vec2(0)
	light := 0.0
	for i := 0; i < 4; i++ {
		f := Fronts[i]
		v := pos - f.xy
		dist := length(v)
		local := max(0.0, f.z-dist/Speed)
		swell := sin(Freq*local) * exp(-Decay*local)
		// the point under the press has no direction to swing in: it holds still
		calm := smoothstep(0.0, 8.0, dist)
		d += Amp * swell * calm * f.w * v / max(dist, 0.001)
		// only the bumps catch the light, softest at their feet; the light
		// needs no direction, so the point under the press is lit too
		crest := max(swell, 0.0)
		light += crest * crest * f.w
	}
	c := layerSample(pos - d)
	// premultiplied: the white goes only where the button is, as much as it is
	c.rgb = mix(c.rgb, vec3(c.a), Glow*min(light, 1.0))
	return c
}
`)

// waveFront is one wave: where it left from, as shares of the element's
// box (so a wave sent before the element is laid out, or one that grows,
// still leaves from its corner), and when.
type waveFront struct {
	fx, fy float32
	at     time.Time
}

// waveState is an element's waves.
type waveState struct {
	fronts []waveFront
}

// wavesOn says whether waves play: only with Options.Waves, and never with
// less motion.
func wavesOn(c *ui.Context) bool {
	return From(c).waves && !c.Preferences().ReduceMotion
}

// send starts a wave from (fx, fy), shares of the element's box: (0.5, 0.5)
// is its middle, (1, 1) its bottom right corner. It does nothing when waves
// don't play.
func (st *waveState) send(c *ui.Context, fx, fy float32) {
	if !wavesOn(c) {
		return
	}
	if len(st.fronts) == waveMax {
		st.fronts = append(st.fronts[:0], st.fronts[1:]...)
	}
	st.fronts = append(st.fronts, waveFront{fx, fy, c.Now()})
	waveSent(fx, fy)
	c.AnimationFrame()
}

// waveSent hears every wave sent, for tests (waves_test.go).
var waveSent = func(fx, fy float32) {}

// run draws e through its waves while any runs, as l tunes them, and asks
// for frames until the last has settled; with none it sets no effect. Call
// it every frame e is built, after any send.
func (st *waveState) run(c *ui.Context, e ui.Element, l waveLook) {
	if len(st.fronts) == 0 {
		return
	}
	b := e.LayoutBounds()
	now := c.Now()
	still := !wavesOn(c)
	// a front is done once it has crossed to the farthest corner and the
	// last point it reached has settled
	tail := time.Duration(5 / float64(l.decay) * float64(time.Second))
	live := st.fronts[:0]
	for _, f := range st.fronts {
		x, y := f.fx*b.W, f.fy*b.H
		far := math.Hypot(float64(max(x, b.W-x)), float64(max(y, b.H-y)))
		end := time.Duration(far/float64(l.speed)*float64(time.Second)) + tail
		if !still && (b.W == 0 || now.Sub(f.at) < end) {
			live = append(live, f)
		}
	}
	st.fronts = live
	if len(live) == 0 {
		return
	}
	c.AnimationFrame()
	if !rippleDraws || b.W == 0 {
		return
	}
	var fronts [4 * waveMax]float32
	for i, f := range live {
		fronts[4*i] = f.fx * b.W
		fronts[4*i+1] = f.fy * b.H
		fronts[4*i+2] = float32(now.Sub(f.at).Seconds())
		fronts[4*i+3] = 1
	}
	e.Effect(rippleShader.Effect(shader.Options{
		Uniforms: map[string]any{
			"Fronts": fronts[:],
			"Speed":  l.speed,
			"Freq":   l.freq,
			"Decay":  l.decay,
			"Amp":    l.amp,
			"Glow":   l.glow,
		},
		MaxSampleOffset: ui.Point{X: l.margin, Y: l.margin},
	}))
}

// rippleState is a button's waves, and whether the pointer held it last frame.
type rippleState struct {
	waveState
	held bool
}

// ripple sends a wave through e as it is pressed, pressed being whether the
// pointer holds it now and clicked whether it was clicked this frame (a click
// with no press before it came from the keyboard). Call it as the element is
// built; with nothing moving it sets no effect. It returns where a press
// began this frame, across the button, x < 0 for the keyboard's, and ok
// false when none did. The press is reported with Reduce Motion and without
// Options.Waves too, when no wave goes: it is still heard.
func ripple(c *ui.Context, e ui.Element, pressed, clicked bool) (x, w float32, ok bool) {
	st := ui.Local(e, "ripple", func() rippleState { return rippleState{} })
	b := e.LayoutBounds()
	switch {
	case pressed && !st.held:
		px, py, _ := e.PointerPosition()
		x, w, ok = px, b.W, true
		if b.W > 0 && b.H > 0 {
			st.send(c, px/b.W, py/b.H)
		}
	case clicked && !st.held:
		x, w, ok = -1, b.W, true
		st.send(c, 0.5, 0.5)
	}
	st.held = pressed
	st.run(c, e, wavePress)
	return
}
