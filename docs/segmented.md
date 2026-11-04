# Segmented

A choice between a few views. The track is inset and as tall as a field
beside it; the chosen option is a raised pill inside it, sliding to a new
choice on a spring, and squeezing in a little as it is pressed. Grooves
divide the other options, fading on either side of the pill.

<!-- preview segmented 130 -->

```go
ryugo.Segmented(c.Key("view"), &app.view, []string{"List", "Grid", "Calendar", "Board"}, ryugo.SegmentedOptions{Label: "View"})
```

`*selected` is the index of the chosen label. `Fill` stretches it across
its parent, the options sharing the width; `Size` and `Disabled` are a
button's.
