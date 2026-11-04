# Popover

A panel of frosted glass below its anchor, on the surface a select's list
uses. A press outside it or Escape closes it. Build the anchor first, in
the same frame.

<!-- preview popover 220 -->

```go
b := ryugo.Button(c.Key("filter"), "Filter", ryugo.ButtonOptions{})
if b.Clicked() {
	app.filtering = !app.filtering
}
ryugo.Popover(c.Key("filters"), b, &app.filtering, func() {
	ryugo.Text(c, "Filters", ryugo.TLabel)
	ryugo.Checkbox(c.Key("mine"), &app.mine, "Only mine", ryugo.CheckOptions{})
})
```
