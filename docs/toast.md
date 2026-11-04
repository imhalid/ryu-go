# Toast

Notices in the window's bottom corner, on frosted glass: a tint and a dot
of their tone, a title, a description, an action and a close button. A new
one rises from below, growing to its size, as two lights of Ryugo's border
beam run round it from opposite sides. Older ones tuck behind it and
spread out under the pointer; a swipe to the side sends one away.

<!-- preview toast 330 -->

```go
ryugo.Toaster(c) // once, anywhere in the view

c.AddToast(ui.Toast{ID: "export", Title: "Export finished", Description: "Q3-report.pdf, 2.4 MB", Type: ryugo.ToastSuccess})
```

Toasts are MyGo's own, `Context.AddToast`; `Toaster` draws them. `Type` is
`ToastInfo`, `ToastSuccess`, `ToastWarning`, `ToastError` or
`ToastLoading`. A toast added with the ID of one showing replaces it in
place: a `ToastLoading` toast turns into its result (press Upload), the
spinner giving way to a dot. `Action` and `OnAction` add a button, and a
negative `Timeout` keeps it until it is closed.
