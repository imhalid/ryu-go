# Collapsible

A title beside a chevron, opening what is under it. The chevron turns a
quarter, and the body grows into view as it fades in; closing, it shrinks
away. A click, Enter or Space opens and closes it.

<!-- preview collapsible 200 -->

```go
ryugo.Collapsible(c.Key("advanced"), "Advanced options", &app.advanced, ryugo.CollapsibleOptions{}, func() {
	ryugo.Checkbox(c.Key("verbose"), &app.verbose, "Verbose logging", ryugo.CheckOptions{})
})
```

`Description` adds a muted line under the title. It returns the header,
whose `Changed` reports a click.
