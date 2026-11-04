# Choice

Radio buttons shaped as cards, for a choice that needs a line of
explanation, as a plan. Each sits in a well, and the chosen one rises to
the raised skin, standing out with its shadow and border.

<!-- preview choice 130 -->

```go
ui.RadioGroup(c, func() {
	ui.Row(c).Gap(8).Children(func() {
		ryugo.Choice(c.Key("free"), &app.tier, "free", "Free", "3 projects", ryugo.CheckOptions{})
		ryugo.Choice(c.Key("pro"), &app.tier, "pro", "Pro", "Unlimited projects", ryugo.CheckOptions{})
	})
}).Label("Plan")
```

Give the cards of a row the same width (`Grow(1).Basis(0)` on a column
around each) so they read as a grid.
