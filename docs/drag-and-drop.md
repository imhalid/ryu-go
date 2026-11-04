# Drag and drop

Cards between the columns of a board. A `DragCard` picked up lifts and
follows the pointer; the `DropArea` it is held over glows, a cloud's shadow
of the accent drifting across it, its cards sliding apart where it would
go. As the card lands, a light runs up both sides of the column.

<!-- preview drag-and-drop 270 -->

```go
ryugo.DropArea(c.Key("done"), ryugo.DropAreaOptions{Title: "Done", Count: "3"}, func() {
	for _, task := range app.done {
		ryugo.DragCard(c.Key(task.ID), task, func() { ui.Text(c, task.Title).Wrap() })
	}
}, func(task *Task, index int) {
	app.move(task, "done", index)
})
```

A card carries a value; the area's function gets the value let go on it,
and its index among the column's other cards. Move it there: it applies
before any column builds the next frame. `Accept` turns some values away,
and `Fill` stretches the column to its row's height.
