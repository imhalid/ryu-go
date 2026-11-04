# Menu

A list of actions on glass, below its anchor. The arrows move between the
items, Enter or a click chooses one, and Escape or a press outside closes
it. `MenuButton` is a button that opens one below itself; `Menu` opens
below any element you give it.

<!-- preview menu 290 -->

```go
ryugo.MenuButton(c.Key("sort"), "Sort", ryugo.ButtonOptions{}, func(m *ryugo.MenuList) {
	m.Label("Sort by")
	for _, by := range []string{"Name", "Date", "Size"} {
		if m.Item(by, ryugo.ItemOptions{Checkable: true, Checked: app.sort == by}) {
			app.sort = by
		}
	}
})

ryugo.Menu(c.Key("more"), moreButton, &app.menu, func(m *ryugo.MenuList) {
	if m.Item("Rename", ryugo.ItemOptions{Icon: editIcon, Shortcut: "⌘R"}) {
		app.rename()
	}
	m.Separator()
	if m.Item("Delete", ryugo.ItemOptions{Icon: trashIcon, Danger: true}) {
		app.delete()
	}
})
```

`Item` returns true in the frame it is chosen. An item has an `Icon`, a
`Shortcut` shown on its end side, a check (`Checkable`, `Checked`), and
can be `Danger` or `Disabled`. A menu too long for the window scrolls.
