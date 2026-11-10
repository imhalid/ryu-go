package ryugo

import (
	"math"
	"math/rand/v2"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/shader"
)

// A Dialog closed by its Danger button burns away: a front of blue fire
// runs out from where the button was pressed across the card, ragged on
// noise, the paper ahead of it paling and the paper behind it charring to a
// clean band edged with a fine line of pale ash, past which the card is
// gone. Flames in separate tongues rise from the front, a light glows on
// the page around it, the paper lifts and curls back as the heat nears,
// and sparks and flakes of ash leave as the front reaches them. The fire's
// colors come from the accent, so a brand's fire is its own blue. It takes
// burnFor; the front crosses in its first 1.15 s.
//
// It plays only with Options.Waves, never with Reduce Motion, and only for
// a closing a Danger button started: Cancel, Escape, the scrim and the
// close button keep the quick fade. What deleting does is the app's and
// happens at once; the burn is only the picture of the card leaving, and
// the pointer goes through it to the page.
const burnFor = 2 * time.Second

// burnParts is how many sparks and flakes leave the card.
const burnParts = 64

var burnShader = shader.MustCompile(`package main

var Size vec2   // the card, DIPs
var Radius float
var Click vec2  // where the Danger button was pressed, in the card
var MaxD float  // from Click to the farthest corner
var K float     // the card's size against 432 x 130
var T float     // seconds since the burn began
var Dark float

var Glow vec4   // the light on the page around the front
var Pale vec4   // the paper paling ahead of the heat
var Scorch vec4 // scorched, before it chars
var Char vec4
var Band vec4   // the char's band, warm black to this
var Ash vec4
var Shade vec4  // the paper lifting
var Heat vec4   // the outer heat, and the flames' feet
var Gold vec4   // the inner heat, and the flames' middle
var Line vec4   // the burning line, and the flames' tips
var Stray vec4  // points catching ahead of the front

var Parts [64]vec4 // x, y, radius, alpha (0: none)
var Kinds [64]vec4 // 1 for a spark, 0 for a flake; a seed

func hash(p vec2) float {
	return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453)
}

func vnoise(p vec2) float {
	i := floor(p)
	f := fract(p)
	u := f * f * (3.0 - 2.0*f)
	a := hash(i)
	b := hash(i + vec2(1.0, 0.0))
	c := hash(i + vec2(0.0, 1.0))
	d := hash(i + vec2(1.0, 1.0))
	return mix(mix(a, b, u.x), mix(c, d, u.x), u.y)
}

func fbm(p vec2) float {
	v := 0.0
	w := 0.5
	sum := 0.0
	for i := 0; i < 5; i++ {
		v += w * vnoise(p)
		sum += w
		p = p*2.03 + 17.1
		w *= 0.5
	}
	return v / sum
}

func coarse(p vec2) float {
	return length((p-Click)*vec2(1.0, 1.15))/MaxD + (fbm(p*0.018)-0.5)*0.42
}

func field(p vec2) float {
	pocket := smoothstep(0.6, 0.78, fbm(p*0.03+40.0)) * 0.12
	return coarse(p) + (fbm(p*0.085)-0.5)*0.05 - pocket
}

// sdBox is the distance from p to the card's rounded edge, < 0 inside.
func sdBox(p vec2) float {
	h := Size * 0.5
	q := abs(p-h) - h + Radius
	return length(max(q, 0.0)) + min(max(q.x, q.y), 0.0) - Radius
}

func band(d, ahead, behind float) float {
	w := behind
	if d > 0.0 {
		w = ahead
	}
	return exp(-(d / w) * (d / w))
}

func over(top, under vec4) vec4 {
	return top + under*(1.0-top.a)
}

func LayerEffect(pos vec2) vec4 {
	level := T/1.15*1.4 - 0.1
	sd := sdBox(pos)
	// off the card (its shadow), the paper nearest it says what is left
	pe := clamp(pos, vec2(0.0), Size)
	fe := field(pe)
	d := fe - level
	charD := d + (vnoise(pe*0.05)-0.5)*0.014
	// past the ash line the paper is gone, with a crisp edge
	left := smoothstep(-0.0535-0.0025, -0.0535, charD)

	// the light behind, on the page around the front
	near := exp(-max(sd, 0.0) / (18.0 * K))
	lightAmt := 0.3 - 0.15*Dark
	light := exp(-abs(d)/0.07) * near * lightAmt * (1.0 - smoothstep(1.1, 1.3, level))
	out := vec4(Glow.rgb, 1.0) * light

	// the paper: it pulls back from the click and rises as the heat nears
	lift := clamp(1.0-d/0.1, 0.0, 1.0)
	lift *= lift
	dir := (pos - Click) / max(length(pos-Click), 0.001)
	// it moves within the card only: lifted past the card's edge, columns
	// of the edge's paper smeared out in stripes
	q := pos
	if sd < 0.0 {
		q = pos - dir*9.0*K*lift + vec2(0.0, 12.0*K*lift*lift)
	}
	c := layerSample(q)
	if sd < 1.0 && c.a > 0.0 {
		rgb := c.rgb / c.a
		rgb = mix(rgb, Shade.rgb, 0.35*lift)
		rgb = mix(rgb, Pale.rgb, 0.55*smoothstep(0.17, 0.05, d))
		rgb = mix(rgb, Scorch.rgb, smoothstep(0.06, 0.008, d))
		rgb = mix(rgb, Char.rgb, smoothstep(0.012, -0.006, d))
		if charD < 0.0 {
			rgb = mix(Char.rgb, Band.rgb, clamp(-charD/0.05, 0.0, 1.0))
		}
		rgb = mix(rgb, Ash.rgb, 0.85*(1.0-smoothstep(0.0025, 0.0045, abs(charD+0.05))))

		// the heat sits on the unburnt side, so the char stays dark
		flick := 0.72 + 0.28*vnoise(pos*0.02+vec2(T*1.5, 0.0))
		rgb = mix(rgb, Heat.rgb, clamp(band(d, 0.05, 0.012)*0.92*flick, 0.0, 1.0))
		rgb = mix(rgb, Gold.rgb, clamp(band(d, 0.024, 0.006)*0.95*flick, 0.0, 1.0))
		rgb = mix(rgb, Line.rgb, clamp(band(d, 0.008, 0.008)*flick, 0.0, 1.0))

		// stray points catch ahead of the front
		stray := smoothstep(0.78, 0.84, vnoise(pos*0.17)) * smoothstep(0.02, 0.03, d) * (1.0 - smoothstep(0.13, 0.15, d))
		stray *= 0.6 + 0.4*sin(18.0*T+vnoise(pos*0.05)*30.0)
		rgb = mix(rgb, Stray.rgb, clamp(stray, 0.0, 1.0))
		c = vec4(rgb, 1.0) * c.a
	}
	out = over(c*left, out)

	// flames rise from the front, in separate tongues
	n1 := vnoise(vec2(pos.x*0.045+0.6*T, pos.y*0.02+6.0*T))
	n2 := vnoise(vec2(pos.x*0.11+5.0, pos.y*0.05+9.0*T))
	tall := (16.0 + 72.0*smoothstep(0.32, 0.8, n1)) * K
	flame := 0.0
	for i := 0; i < 5; i++ {
		s := tall * float(i) / 4.0
		sp := pos + vec2(0.0, s)
		on := smoothstep(2.0, -6.0, sdBox(sp))
		cf := (coarse(sp) - level) / 0.042
		flame += exp(-cf*cf) * (1.0 - smoothstep(0.0, tall, s)) * 0.5 * on
	}
	flame = min(flame, 1.0) * smoothstep(0.2, 0.62, n1*0.7+n2*0.3)
	fc := mix(Heat.rgb, Gold.rgb, smoothstep(0.2, 0.55, flame))
	fc = mix(fc, Line.rgb, smoothstep(0.55, 0.9, flame))
	fa := smoothstep(0.03, 0.62, flame) * 0.95
	out = over(vec4(fc, 1.0)*fa, out)

	// sparks and flakes
	for i := 0; i < 64; i++ {
		pt := Parts[i]
		if pt.w > 0.0 {
			v := pos - pt.xy
			r := pt.z
			if Kinds[i].x < 0.5 {
				// a flake's outline is irregular
				r *= 0.7 + 0.5*vnoise(v/max(length(v), 0.001)*2.0+Kinds[i].y)
			}
			dist := length(v)
			a := (1.0 - smoothstep(r-0.8, r+0.4, dist)) * pt.w
			if a > 0.0 {
				col := Ash.rgb
				if Kinds[i].x > 0.5 {
					// a bright core in the fire's color: a white one reads as a
					// ring on a light page
					col = mix(Gold.rgb, Heat.rgb, smoothstep(0.0, r, dist))
				}
				out = over(vec4(col, 1.0)*a, out)
			}
		}
	}
	return out
}
`)

// burnState is a Dialog's burn: when and where its Danger button was
// pressed, and the burn under way.
type burnState struct {
	armedAt time.Time
	at      ui.Point // the press, in the window
	wasOpen bool
	burning bool
	click   ui.Point // the press, in the card, once burning
	box     ui.Rect  // the card as it began to burn
	parts   []burnPart
}

// burnPart is a spark or a flake: where it starts, in the card, and when.
type burnPart struct {
	x, y, start, life float64
	size, rise, drift float64
	spark             bool
	seed              float64
}

// burnArm is the burn of the Dialog being built: a Danger button pressed
// inside it arms it (Button).
var burnArm *burnState

// arm marks the burn as started by a press at (x, y), in the window.
func (b *burnState) arm(c *ui.Context, x, y float32) {
	b.armedAt, b.at = c.Now(), ui.Point{X: x, Y: y}
}

// closing tells, as the panel starts to close, whether it burns: it does
// when a Danger button pressed just before closed it, and waves play.
func (b *burnState) closing(c *ui.Context, open bool) {
	switch {
	case open:
		b.burning = false
	case b.wasOpen:
		b.burning = wavesOn(c) && !b.armedAt.IsZero() && c.Now().Sub(b.armedAt) < 250*time.Millisecond
		b.armedAt, b.parts = time.Time{}, nil
	}
	b.wasOpen = open
}

// burnK is a card's size against the 432 x 130 the burn was tuned on.
func burnK(w, h float32) float64 {
	return min(max(math.Sqrt(float64(w*h)/(432*130)), 0.6), 1.4)
}

// burnColors are the fire's, from the accent's hue: as the fire of paper
// would be, with blue in place of orange and a cool char.
func (t *Tokens) burnColors() map[string]any {
	_, a, bb := toOklab(t.Colors.Accent)
	h := float32(math.Atan2(bb, a) * 180 / math.Pi)
	char := oklch(0.13, 0.015, h, 1)
	if t.Dark {
		// char vanishes on a dark page: lifted
		char = oklch(0.3, 0.015, h, 1)
	}
	return map[string]any{
		"Glow":   oklch(0.72, 0.17, h, 1),
		"Pale":   oklch(0.88, 0.06, h, 1),
		"Scorch": oklch(0.32, 0.09, h, 1),
		"Char":   char,
		"Band":   oklch(0.24, 0.02, h, 1),
		"Ash":    oklch(0.77, 0.012, h, 1),
		"Shade":  oklch(0.82, 0.02, h, 1),
		"Heat":   oklch(0.56, 0.21, h, 1),
		"Gold":   oklch(0.8, 0.13, h-25, 1),
		"Line":   oklch(0.97, 0.03, h, 1),
		"Stray":  oklch(0.82, 0.12, h-15, 1),
		"Dark":   b2f(t.Dark),
	}
}

// burnHash, burnNoise and burnFbm are the shader's noise, in Go, so a
// particle leaves as the front reaches it.
func burnHash(x, y float64) float64 {
	v := math.Sin(x*127.1+y*311.7) * 43758.5453
	return v - math.Floor(v)
}

func burnNoise(x, y float64) float64 {
	ix, iy := math.Floor(x), math.Floor(y)
	fx, fy := x-ix, y-iy
	ux, uy := fx*fx*(3-2*fx), fy*fy*(3-2*fy)
	a, b := burnHash(ix, iy), burnHash(ix+1, iy)
	c, d := burnHash(ix, iy+1), burnHash(ix+1, iy+1)
	top, bottom := a+(b-a)*ux, c+(d-c)*ux
	return top + (bottom-top)*uy
}

func burnFbm(x, y float64) float64 {
	v, w, sum := 0.0, 0.5, 0.0
	for range 5 {
		v += w * burnNoise(x, y)
		sum += w
		x, y = x*2.03+17.1, y*2.03+17.1
		w *= 0.5
	}
	return v / sum
}

func (b *burnState) field(x, y, maxD float64) float64 {
	dx, dy := x-float64(b.click.X), (y-float64(b.click.Y))*1.15
	f := math.Hypot(dx, dy)/maxD + (burnFbm(x*0.018, y*0.018)-0.5)*0.42 + (burnFbm(x*0.085, y*0.085)-0.5)*0.05
	return f - smoothstep(0.6, 0.78, burnFbm(x*0.03+40, y*0.03+40))*0.12
}

func smoothstep(a, b, x float64) float64 {
	t := min(max((x-a)/(b-a), 0), 1)
	return t * t * (3 - 2*t)
}

// begin starts the burn of the card at box, its first frame.
func (b *burnState) begin(box ui.Rect) {
	b.box = box
	b.click = ui.Point{X: b.at.X - box.X, Y: b.at.Y - box.Y}
	if b.at == (ui.Point{}) {
		b.click = ui.Point{X: box.W / 2, Y: box.H / 2}
	}
	maxD := b.maxD()
	k := burnK(box.W, box.H)
	r := rand.New(rand.NewPCG(uint64(b.click.X*31), uint64(b.click.Y*17)))
	b.parts = make([]burnPart, burnParts)
	for i := range b.parts {
		x, y := r.Float64()*float64(box.W), r.Float64()*float64(box.H)
		// it leaves as the front reaches it: level(t) = field
		start := (b.field(x, y, maxD) + 0.1) / 1.4 * 1.15
		p := burnPart{x: x, y: y, start: start, life: 0.5 + 0.7*r.Float64(),
			rise: (60 + 180*r.Float64()) * k, drift: (r.Float64()*2 - 1) * 20, seed: r.Float64() * 10}
		if r.Float64() < 0.8 {
			p.spark, p.size, p.rise = true, 2.5+2*r.Float64(), p.rise*1.2
		} else {
			p.size, p.rise = 3+5*r.Float64(), p.rise*0.8
		}
		b.parts[i] = p
	}
}

func (b *burnState) maxD() float64 {
	cx, cy := float64(b.click.X), float64(b.click.Y)
	w, h := float64(b.box.W), float64(b.box.H)
	return max(math.Hypot(cx, cy*1.15), math.Hypot(w-cx, cy*1.15), math.Hypot(cx, (h-cy)*1.15), math.Hypot(w-cx, (h-cy)*1.15), 1)
}

// burnMargin is how far around the card the burn draws: the light on the
// page, the flames above the card's top, the sparks rising.
var burnMargin = ui.Point{X: 72, Y: 300}

// draw burns card, t into the burn.
func (b *burnState) draw(c *ui.Context, t *Tokens, card ui.Element, at time.Duration) {
	if b.parts == nil {
		box := card.LayoutBounds()
		if box.W == 0 {
			return
		}
		b.begin(box)
	}
	c.AnimationFrame()
	if !rippleDraws {
		return
	}
	s := at.Seconds()
	k := burnK(b.box.W, b.box.H)
	var parts, kinds [4 * burnParts]float32
	for i, p := range b.parts {
		age := s - p.start
		if age < 0 || age > p.life {
			continue
		}
		u := age / p.life
		// it rises fast and slows, drifting and swaying as it goes
		up := p.rise * (1 - (1-u)*(1-u))
		x := p.x + math.Sin(6*age)*12*age + p.drift*u
		alpha := 1 - smoothstep(0.45, 1, u)
		if !p.spark {
			alpha *= 0.6
		}
		parts[4*i] = float32(x)
		parts[4*i+1] = float32(p.y - up)
		parts[4*i+2] = float32(p.size / 2 * k)
		parts[4*i+3] = float32(alpha)
		kinds[4*i] = b2f(p.spark)
		kinds[4*i+1] = float32(p.seed)
	}
	u := t.burnColors()
	u["Size"] = []float32{b.box.W, b.box.H}
	u["Radius"] = t.R(RadiusLg)
	u["Click"] = []float32{b.click.X, b.click.Y}
	u["MaxD"] = b.maxD()
	u["K"] = k
	u["T"] = s
	u["Parts"] = parts[:]
	u["Kinds"] = kinds[:]
	card.Effect(burnShader.Effect(shader.Options{Uniforms: u, MaxSampleOffset: burnMargin}))
}
