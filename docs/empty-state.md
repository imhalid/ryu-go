# Empty state

What an empty list, a failure or a first visit says: a centered column with
an icon in a neutral well, a title, a description and one action. It has
no color accent; the state is told in words.

<!-- preview empty-state 220 -->

```go
mizan.EmptyState(c, mizan.EmptyStateOptions{Icon: inboxIcon, Title: "No files", Description: "Files you add are listed here.", Action: func() {
	if mizan.Button(c.Key("add"), "Add a file", mizan.ButtonOptions{Size: mizan.Sm}).Clicked() {
		app.browse()
	}
}})
```

A [table](table.md) shows one of these when it has no rows (`Empty`).
