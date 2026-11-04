# Sortable

A list reordered by dragging: each row is a raised tile with a grip. A row
picked up lifts and follows the pointer, leaning with its speed; the others
slide apart where it would go, and let go, it flies into its place.
Alt+Up and Alt+Down move the focused row.

<!-- preview sortable 280 -->

```go
ryugo.Sortable(c.Key("steps"), &app.list, &app.steps, ryugo.SortableOptions{
	Label: func(i int) string { return app.steps[i] },
}, func(step string) {
	ui.Text(c, step).Grow(1)
})
```

It reorders `*items` itself. The `ui.ListState` identifies the list and
must stay the same across frames. `Label` names a row to assistive
technology, and `Key` tells rows apart when their values repeat.
