# Copy button

A round raised button that copies text to the clipboard: its icon shrinks
away as a check draws itself in its place, in the Down color, and 1.5 s
later the check draws back and the icon returns.

<!-- preview copy-button 110 -->

```go
mizan.CopyButton(c.Key("copy"), app.link, mizan.CopyOptions{Label: "Copy link"})
```

`Label` names it to assistive technology; give it a `Tooltip` to say it to
everyone else too. `Size` and `Disabled` are a button's.
