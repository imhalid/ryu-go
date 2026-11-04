# Dialog

A modal panel over a scrim that fades in. It springs in from 94% and a
little below, and fades out on the way down as it closes. Its header drags
it around the window: it leans with the pointer's speed, a throw glides
on, and it springs back inside the window when let go; a double click on
the header brings it back to the middle. Content that scrolls fades into
its header and footer.

<!-- preview dialog 480 -->

```go
ryugo.Dialog(c.Key("delete"), &app.confirm, ryugo.DialogOptions{
	Title:       "Delete the report?",
	Description: "Q3-report.pdf will be removed for everyone.",
	Footer: func() {
		if ryugo.Button(c.Key("cancel"), "Cancel", ryugo.ButtonOptions{}).Clicked() {
			app.confirm = false
		}
		if ryugo.Button(c.Key("ok"), "Delete report", ryugo.ButtonOptions{Danger: true}).Clicked() {
			app.delete()
		}
	},
}, func() {
	ui.Text(c, "This cannot be undone.").Wrap()
})
```

It shows while `*open` is true. Escape, a click on the scrim or the close
button set it to false, which the returned panel's `Dismissed` reports;
`NoClose` leaves the close button out (Escape and the scrim still close
it), `Fixed` keeps it where it opens, and `Width` sets its width, 480 by
default.
