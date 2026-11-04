# Radio

One of a few. Chosen, the fill comes first and the white dot opens from its
middle a moment later; left, the dot shrinks away. Put the radios of one
choice in a `ui.RadioGroup`, which the arrows move within.

<!-- preview radio 220 -->

```go
ui.RadioGroup(c, func() {
	ryugo.Radio(c.Key("day"), &app.plan, "day", "Daily", ryugo.CheckOptions{})
	ryugo.Radio(c.Key("week"), &app.plan, "week", "Weekly", ryugo.CheckOptions{})
}).Label("Digest frequency")
```

The value is any comparable type: a radio is chosen while `*selected`
equals its value. `Description` and `Disabled` are a checkbox's.
