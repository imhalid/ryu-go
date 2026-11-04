# Tooltip

A small raised label, lit from above as a button is, above an element the
pointer rests on or the focus is on. It rises as it fades in and follows
the pointer along its element, lagging a little behind. Going from one
element to the next it doesn't close and open again but slides there, so
a row of buttons never flickers; it lingers a moment after the pointer
leaves, and goes at once when the pointer comes onto it.

<!-- preview tooltip 170 -->

```go
b := ryugo.Button(c.Key("share"), "Share", ryugo.ButtonOptions{Icon: shareIcon, IconOnly: true}).Description("Share the report")
ryugo.Tooltip(c.Key("shareTip"), b, "Share the report")
```

Give the anchor a `Description` too: assistive technology does not read
tooltips. A long tooltip wraps.
