# Progress

An inset well, as a slider's track, filling with the accent gradient on a
spring as the value grows from 0 to 1. A negative value is work of unknown
length: a band of the fill sweeps across the well, again and again. As the
value reaches 1, a light passes once over the full bar; press +10% until
it does.

<!-- preview progress 160 -->

```go
mizan.Progress(c, app.uploaded, mizan.ProgressOptions{Label: "Uploading Q3-report.pdf", ShowValue: true})
mizan.Progress(c, -1, mizan.ProgressOptions{Label: "Syncing contacts", Size: mizan.Sm})
```

`Label` stands above the bar and names it to assistive technology,
`ShowValue` writes the percentage above its end, `Color` fills it in place
of the accent, and `Size` sets its thickness.
