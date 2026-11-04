# Table

Rows of any type under a header of columns, built only as far as they show,
so a table of thousands scrolls as one of ten. A click on a header sorts by
it (Shift-click adds a key), and the rows slide to their new places; a
header dragged lifts its column, the others sliding aside for it; its end
edge resizes it. Choose rows with the checkboxes (Shift-click for a range)
and a selection bar springs in over the toolbar. Try the search, the View
menu, a row's chevron, and archiving a few rows.

<!-- preview table 520 -->

```go
cols := []mizan.Column[Order]{
	{Title: "Customer", Kind: mizan.KindPerson, Value: func(o Order) any { return o.Customer }, Sortable: true},
	{Title: "Status", Kind: mizan.KindStatus, Value: func(o Order) any { return o.Status }, Sortable: true},
	{Title: "Total", Kind: mizan.KindNumber, Value: func(o Order) any { return o.Total }, Sortable: true, Decimals: 2, Unit: "USD"},
}
mizan.Table(c.Key("orders"), &app.table, cols, app.orders, mizan.TableOptions[Order]{
	Key:        func(o Order) any { return o.ID },
	Selectable: true, Search: true, ViewMenu: true,
	Actions:    []mizan.TableAction{{Label: "Archive", Danger: true, Do: app.archive}},
})
```

## Columns

A column's `Kind` says how its cells look and sort: `KindText`,
`KindNumber` (tabular, with `Decimals` and a `Unit`), `KindChange` (a signed
change in its tone), `KindDate`, `KindStatus` (a [status dot](status-dot.md)
of a `CellStatus`), `KindPerson` (an [avatar](avatar.md) and the name) and
`KindBar` (a share from 0 to 1). `Cell` builds a cell of your own instead.
`Width` sets a column's width; without one it shares the room the others
leave, no narrower than `MinWidth`. `Fixed` keeps the user from moving or
resizing it.

## State

`TableState` is yours to keep, and to save: the sort (`Sort`), the chosen
rows (`Selected`, by key), the columns' order, widths and hidden ones, and
the search's `Query`. The table sorts the rows itself; filtering by the
query is up to the app.

## Options

- `Expand` builds a row's details, opening under it from a chevron.
- `Pin` keeps the first column in place as the table scrolls sideways.
- `PageSize` shows a page at a time, with [pagination](pagination.md) under it.
- `Loading` shows placeholders in the rows' place, `Error` a message with
  `Retry`, and `Empty` what an empty table says.
- `Toolbar` builds more of the toolbar; `Size` is the density.

A table of pages, in the small density:

<!-- preview table-pages 340 -->

With the focus on the table, Up and Down move the cursor row, Space chooses
it, Shift with the arrows chooses a range, and Enter opens it.
