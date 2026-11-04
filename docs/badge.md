# Badge

A pill on a tint of its own text color, so the background always matches
the text, whatever the tone. Numbers in it are tabular. `Tag` is the larger
neutral pill beside a page's title.

<!-- preview badge 130 -->

```go
mizan.Badge(c, "+4.2%", mizan.ToneDown)
mizan.Badge(c, "Pending", mizan.ToneWarn)
mizan.Tag(c, "v2.4")
```

The tones are `ToneDefault` (the text color), `ToneNeutral` (a gray
label), and the three tones of a change: `ToneDown` reads as good, `ToneWarn`
as a warning and `ToneUp` as bad, as a cost going down or up.
