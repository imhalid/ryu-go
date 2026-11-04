# Link

A link inside text, in an accent deep enough to read as text, with an
underline a third as strong that turns full under the pointer. A click or
Enter opens its URL in the browser; in a Router, a path goes there. With
no URL it is a button that reads as a link: `Clicked` reports it.

<!-- preview link 130 -->

```go
ryugo.Link(c, "Read the changelog", "https://github.com/imhalid/ryu-go")
if ryugo.Link(c, "Show details", "").Clicked() {
	app.details = !app.details
}
```

Inside a `ui.RichText` it is a link within the paragraph, wrapping with it.
