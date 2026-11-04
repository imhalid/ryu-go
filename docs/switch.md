# Switch

A 38 by 22 switch: the track is a dark well that fills with the accent
gradient when on, and the knob slides on a spring, stretching into a short
capsule while it is held.

<!-- preview switch 170 -->

```go
ryugo.Switch(c.Key("sync"), &app.sync, ryugo.SwitchOptions{Label: "Sync", Description: "Refresh the data every hour."})
```

A switch applies at once; for a choice that waits for a Save, a
[checkbox](checkbox.md) reads better.
