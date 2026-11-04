# Meter

How full something is, as a disk or a quota: a progress bar's well, filled
in a tone that turns from Down (fine) to Warn to Up as the value passes
`Warning` and `Critical`. Assistive technology sees a level indicator.

<!-- preview meter 190 -->

```go
mizan.Meter(c, app.used, 0, 64, mizan.MeterOptions{Label: "Storage", Warning: 48, Critical: 60,
	Format: func(v float64) string { return fmt.Sprintf("%.0f of 64 GB", v) }})
```

Unlike a progress bar, a full meter plays no light: full doesn't mean done.
