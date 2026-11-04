# Sheet

A panel along the window's end edge, over a scrim as a dialog's. It
springs in from past the edge, tilted a little, and straightens as it
lands; closing, it speeds back out. It closes as a [dialog](dialog.md)
does, and takes the same options.

<!-- preview sheet 440 -->

```go
mizan.Sheet(c.Key("settings"), &app.settings, mizan.DialogOptions{
	Title:       "Report settings",
	Description: "Changes apply to the next export.",
	Footer:      func() { ... },
}, func() {
	...
})
```
