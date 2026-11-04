# Toggle

A button that stays pressed while on, as Bold in an editor. Off, it is
flat and muted, with a soft fill under the pointer; on, it is pressed into
the surface, its icon in the accent. `ToggleGroup` puts toggles and buttons
on one raised bar, which the arrows move within; a vertical
[separator](separator.md) divides it.

<!-- preview toggle 130 -->

```go
mizan.ToggleGroup(c.Key("style"), "Text style", func() {
	mizan.Toggle(c.Key("b"), &app.bold, "Bold", mizan.ToggleOptions{Icon: boldIcon, IconOnly: true})
	mizan.Toggle(c.Key("i"), &app.italic, "Italic", mizan.ToggleOptions{Icon: italicIcon, IconOnly: true})
})
```

`Changed` reports a click, as to make two toggles exclusive (the
alignments above). Assistive technology sees a toggle button.
