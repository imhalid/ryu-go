# Spinner

A ring for a short wait: a faint track, and over it a quarter turn of the
accent that goes round once every 0.8 seconds. A button with `Loading`
shows one in place of its icon.

<!-- preview spinner 100 -->

```go
mizan.Spinner(c, mizan.SpinnerOptions{})
mizan.Spinner(c, mizan.SpinnerOptions{Size: 24, Color: t.Colors.Muted})
```

`Label` names it to assistive technology, which sees a progress indicator.
For work whose length is known, use a [progress bar](progress.md).
