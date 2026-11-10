package ryugo

import (
	"math"
	"math/rand/v2"
	"sync"
	"testing"

	"github.com/egoist/mygo/audio"
)

// A press heard: the wave's sine made audible, a drop's "plip". Its pitch
// rises a little as it sounds (a falling drop's cavity closing), and it dies
// away in ~120 ms. It comes from the side of the button that was pressed,
// a few percent higher or lower each time, so quick presses don't sound
// alike. The sounds are made once, in Go: no files.
//
// Sounds are off unless the app asks for them (Options.Sounds): desktop
// buttons are silent, and sounds an app makes on every press tire quickly.
// Reduce Motion doesn't silence them: it is about motion.

// tone is a kind of press sound.
type tone uint8

const (
	tonePlain  tone = iota // a raised button: light and high
	toneSolid              // a solid fill: lower, fuller
	toneDanger             // a destructive action: low and short
)

// drop describes a press sound.
type drop struct {
	from, to float64 // Hz, at the start and as it settles
	rise     float64 // s: how quickly it goes from `from` to `to`
	decay    float64 // 1/s
	body     float64 // the octave under it, as a share of the tone
	volume   float64
}

var drops = [...]drop{
	tonePlain:  {from: 1000, to: 1400, rise: 0.025, decay: 48, body: 0, volume: 0.13},
	toneSolid:  {from: 760, to: 1060, rise: 0.03, decay: 40, body: 0.3, volume: 0.15},
	toneDanger: {from: 560, to: 680, rise: 0.02, decay: 60, body: 0.35, volume: 0.15},
}

const soundRate = 48000

// samples makes the drop's sound: 2 ms of attack (no click), then the
// sine dying as e^(−decay·t), its pitch gliding from `from` to `to`.
func (d drop) samples() []float32 {
	n := int(math.Ceil(6 / d.decay * soundRate)) // to e^−6
	out := make([]float32, n)
	var phase float64
	for i := range out {
		t := float64(i) / soundRate
		f := d.to + (d.from-d.to)*math.Exp(-t/d.rise)
		phase += 2 * math.Pi * f / soundRate
		env := math.Exp(-d.decay*t) * min(t/0.002, 1)
		// the last 5 ms fade to nothing
		if tail := float64(n-i) / soundRate; tail < 0.005 {
			env *= tail / 0.005
		}
		v := math.Sin(phase) + d.body*math.Sin(phase/2)
		out[i] = float32(env * v / (1 + d.body))
	}
	return out
}

var (
	soundsOnce sync.Once
	sounds     [len(drops)]*audio.Sound
)

// soundPlays is false in tests: no audio device is opened there.
// sound_test.go replaces playSound to hear what would play.
var soundPlays = !testing.Testing()

// playSound plays the press sound k, pan from −1 (left) to 1 (right) and
// speed its pitch (1 as made).
var playSound = func(k tone, pan, speed float64) {
	if !soundPlays {
		return
	}
	soundsOnce.Do(func() {
		for i, d := range drops {
			sounds[i] = audio.NewSound(soundRate, 1, d.samples())
		}
	})
	sounds[k].NewPlayer().SetVolume(drops[k].volume).SetPan(pan).SetSpeed(speed).Play()
}

// pressSound plays k for a press at x across a button w wide (x < 0: by
// keyboard, from the middle).
func pressSound(k tone, x, w float32) {
	pan := 0.0
	if x >= 0 && w > 0 {
		pan = 0.35 * (2*float64(x/w) - 1)
		pan = max(-0.35, min(0.35, pan))
	}
	playSound(k, pan, 0.97+0.06*rand.Float64())
}
