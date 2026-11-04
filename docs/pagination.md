# Pagination

Pages of rows. On the start side, the range shown and the total; on the
end side, previous and next around the page numbers, the first and last
always there and the pages around the current one between, gaps shown as
"…". The current page sits on a raised pill that slides from number to
number, the numbers keeping their places as it moves.

<!-- preview pagination 100 -->

```go
ryugo.Pagination(c.Key("pages"), &app.page, len(app.rows), ryugo.PaginationOptions{PageSize: 20})
```

`*page` counts from 0. A [Table](table.md) with `PageSize` has one of its
own.
