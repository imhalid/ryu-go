# Popover

A panel of frosted glass below its anchor, on the surface a select's list
uses. A press outside it or Escape closes it. Build the anchor first, in
the same frame.

<!-- preview popover 220 -->

```go
b := mizan.Button(c.Key("filter"), "Filter", mizan.ButtonOptions{})
if b.Clicked() {
	app.filtering = !app.filtering
}
mizan.Popover(c.Key("filters"), b, &app.filtering, func() {
	mizan.Text(c, "Filters", mizan.TLabel)
	mizan.Checkbox(c.Key("mine"), &app.mine, "Only mine", mizan.CheckOptions{})
})
```
