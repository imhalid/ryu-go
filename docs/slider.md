# Slider

A value in a range: a 6 px inset track filled with the accent gradient up
to the knob. Held, the knob grows a little and follows the pointer between
steps, a label of its value rising over it; let go, it settles onto its
step on a spring.

<!-- preview slider 150 -->

```go
mizan.Slider(c.Key("vol"), &app.volume, 0, 100, mizan.SliderOptions{Label: "Volume", Step: 1,
	Format: func(v float64) string { return fmt.Sprintf("%.0f%%", v) }})
```

`Step` rounds the value; the arrows move it a step at a time. With
`Format`, the value also stands on the slider's right, as it says.
