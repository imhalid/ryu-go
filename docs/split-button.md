# Split button

An action joined to an arrow that opens a menu of its variants, as New page beside the pages of the other
templates, or Save beside Save As…: one pill with a Button's face, color, fill and size, each segment lighting up
apart as it is hovered and pressed, sending its own wave and sound, a line between them.

<!-- preview split-button 110 -->

```go
sb := ryugo.SplitButton(c.Key("new"), "New page", ryugo.SplitButtonOptions{Icon: icons.Plus, Color: t.Colors.Accent, MenuLabel: "From a template"})
sb.Arrow.PopoverMenuButton(func(m *ui.Menu) {
	if m.Item("Meeting notes").Chosen() { app.newFrom("meeting") }
})
if sb.Button.Clicked() {
	app.newPage()
}
```

It is MyGo's `SplitButton` underneath: one stop of Tab, Left and Right moving between the segments, Enter and Space
clicking the action, Enter, Space, Up, Down and Alt+Down opening the menu; assistive technology hears a group of a
button and a menu button, the arrow by `MenuLabel` ("More options" without one). `Disabled` disables both segments,
`ActionDisabled` the action alone, the menu still opening. The arrow takes a menu with `Menu` (the system's),
`PopoverMenuButton` (one drawn in the window), or opens one of the app's on `Arrow.Clicked()`.
