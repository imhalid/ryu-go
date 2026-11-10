package ryugo

import (
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/shader"
)

// A press sends a wave through the button from where it landed: a front
// leaves the pointer at rippleSpeed and, as it reaches a point, the point
// swings out and back along the line from the press, a sine dying away
// (sin(f·t)·e^(−d·t), t counted from when the front came by). The surface,
// its outline and its label ripple together. Where the wave swells out, a
// very faint white light catches the bump (rippleGlow), as a light from
// above would; the troughs stay as they are, so the button never darkens.
// A press by keyboard sends it from the middle.
const (
	rippleSpeed = 240 // DIPs a second
	rippleFreq  = 42  // rad/s: ~36 DIPs from crest to crest
	rippleDecay = 7   // 1/s: e^−5 left after rippleTail
	rippleAmp   = 1.8 // DIPs, at most
	rippleMax   = 4   // fronts at once; a fifth press takes the oldest's place
	rippleGlow  = 0.1 // white mixed in at a crest's top, at most: barely there
)

// rippleTail is how long a point keeps swinging once the front has passed.
const rippleTail = 5 * time.Second / rippleDecay

// rippleMargin is how far the effect draws around the button: the swing,
// and the focus ring's 5px, which the effect's bitmap must hold whole.
const rippleMargin = 8

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
		// only the bumps catch the light, softest at their feet
		crest := max(swell, 0.0)
		light += crest * crest * calm * f.w
	}
	c := layerSample(pos - d)
	// premultiplied: the white goes only where the button is, as much as it is
	c.rgb = mix(c.rgb, vec3(c.a), Glow*min(light, 1.0))
	return c
}
`)

type rippleFront struct {
	x, y float32
	at   time.Time
}

// rippleState is a button's waves, and whether the pointer held it last frame.
type rippleState struct {
	fronts []rippleFront
	held   bool
}

// ripple sends a wave through e as it is pressed, pressed being whether the
// pointer holds it now and clicked whether it was clicked this frame (a click
// with no press before it came from the keyboard). Call it as the element is
// built; with nothing moving it sets no effect. It returns where a press
// began this frame, across the button, x < 0 for the keyboard's, and ok
// false when none did (with Reduce Motion or without Options.Waves too: the
// press is still heard).
func ripple(c *ui.Context, e ui.Element, pressed, clicked bool) (x, w float32, ok bool) {
	st := ui.Local(e, "ripple", func() rippleState { return rippleState{} })
	now := c.Now()
	// no wave without Options.Waves, nor with less motion: the press is
	// still reported, and heard with Options.Sounds
	still := c.Preferences().ReduceMotion || !From(c).waves
	b := e.LayoutBounds()
	switch {
	case pressed && !st.held:
		px, py, _ := e.PointerPosition()
		x, w, ok = px, b.W, true
		if !still {
			st.add(rippleFront{px, py, now})
		}
	case clicked && !st.held:
		x, w, ok = -1, b.W, true
		if !still {
			st.add(rippleFront{b.W / 2, b.H / 2, now})
		}
	}
	st.held = pressed
	if len(st.fronts) == 0 {
		return
	}

	// a front is done once it has crossed to the farthest corner and the
	// last point it reached has settled
	live := st.fronts[:0]
	for _, f := range st.fronts {
		far := math.Hypot(float64(max(f.x, b.W-f.x)), float64(max(f.y, b.H-f.y)))
		end := time.Duration(far/rippleSpeed*float64(time.Second)) + rippleTail
		if !still && now.Sub(f.at) < end {
			live = append(live, f)
		}
	}
	st.fronts = live
	if len(live) == 0 || !rippleDraws {
		return
	}

	var fronts [4 * rippleMax]float32
	for i, f := range live {
		fronts[4*i] = f.x
		fronts[4*i+1] = f.y
		fronts[4*i+2] = float32(now.Sub(f.at).Seconds())
		fronts[4*i+3] = 1
	}
	e.Effect(rippleShader.Effect(shader.Options{
		Uniforms: map[string]any{
			"Fronts": fronts[:],
			"Speed":  rippleSpeed,
			"Freq":   rippleFreq,
			"Decay":  rippleDecay,
			"Amp":    rippleAmp,
			"Glow":   rippleGlow,
		},
		MaxSampleOffset: ui.Point{X: rippleMargin, Y: rippleMargin},
	}))
	c.AnimationFrame()
	return
}

func (st *rippleState) add(f rippleFront) {
	if len(st.fronts) == rippleMax {
		st.fronts = append(st.fronts[:0], st.fronts[1:]...)
	}
	st.fronts = append(st.fronts, f)
}
