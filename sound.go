package ryugo

import (
	"math"
	"math/rand/v2"
	"sync"
	"testing"

	"github.com/egoist/mygo/audio"
)

// A press heard: a muted tap, as a fingertip on a firm key. A soft knock
// of noise for the touch, under it a short low tone for the body, settling
// a little in pitch as it sounds (a struck thing does: it never rises,
// which reads as a toy), the highs taken off so nothing rings. It is over
// in ~80 ms. It comes from the side of the button that was pressed, a few
// percent higher or lower each time, so quick presses don't sound alike.
// The sounds are made once, in Go: no files.
//
// Sounds are off unless the app asks for them (Options.Sounds): desktop
// buttons are silent, and sounds an app makes on every press tire quickly.
// Reduce Motion doesn't silence them: it is about motion.

// tone is a kind of press sound.
type tone uint8

const (
	tonePlain  tone = iota // a raised button: light
	toneSolid              // a solid fill: lower, fuller
	toneDanger             // a destructive action: lowest and shortest
)

// drop describes a press sound.
type drop struct {
	pitch  float64 // Hz, where the body settles
	sag    float64 // how far above pitch it starts, as a share (0.06: 6%)
	decay  float64 // 1/s, the body's
	knock  float64 // the touch's noise, as a share of the body
	cutoff float64 // Hz: what is above is taken off, gently
	volume float64
}

var drops = [...]drop{
	tonePlain:  {pitch: 430, sag: 0.06, decay: 70, knock: 0.35, cutoff: 2400, volume: 0.3},
	toneSolid:  {pitch: 330, sag: 0.07, decay: 62, knock: 0.3, cutoff: 2000, volume: 0.32},
	toneDanger: {pitch: 240, sag: 0.08, decay: 80, knock: 0.3, cutoff: 1700, volume: 0.32},
}

const soundRate = 48000

// samples makes the tap: 1.5 ms of attack (no click), the body's sine dying
// as e^(−decay·t) with a faint wooden partial over it, the knock a few ms of
// noise, all through two gentle low-pass poles.
func (d drop) samples() []float32 {
	n := int(math.Ceil(6 / d.decay * soundRate)) // to e^−6
	out := make([]float32, n)
	noise := rand.New(rand.NewPCG(1, 2)) // the same knock every time
	lp := 1 - math.Exp(-2*math.Pi*d.cutoff/soundRate)
	var phase, a, b, peak float64
	raw := make([]float64, n)
	for i := range raw {
		t := float64(i) / soundRate
		f := d.pitch * (1 + d.sag*math.Exp(-t/0.012))
		phase += 2 * math.Pi * f / soundRate
		body := math.Exp(-d.decay*t) * (math.Sin(phase) + 0.12*math.Exp(-3*d.decay*t)*math.Sin(2.7*phase))
		knock := d.knock * math.Exp(-t/0.0025) * (2*noise.Float64() - 1)
		v := (body + knock) * min(t/0.0015, 1)
		// the last 5 ms fade to nothing
		if tail := float64(n-i) / soundRate; tail < 0.005 {
			v *= tail / 0.005
		}
		a += lp * (v - a)
		b += lp * (a - b)
		raw[i] = b
		peak = max(peak, math.Abs(b))
	}
	for i, v := range raw {
		out[i] = float32(v / peak)
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
