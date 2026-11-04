# Collapsible

A title beside a chevron, opening what is under it. The chevron turns a
quarter, and the body grows into view as it fades in; closing, it shrinks
away. A click, Enter or Space opens and closes it.

<!-- preview collapsible 200 -->

```go
mizan.Collapsible(c.Key("advanced"), "Advanced options", &app.advanced, mizan.CollapsibleOptions{}, func() {
	mizan.Checkbox(c.Key("verbose"), &app.verbose, "Verbose logging", mizan.CheckOptions{})
})
```

`Description` adds a muted line under the title. It returns the header,
whose `Changed` reports a click.
