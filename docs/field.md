# Field

A label, a control and a line under it, help or error. An invalid field's
label turns the Up color; the control itself is marked with its own
`Invalid`. A field asked to (`Shake` counting up) swings sideways a few
times, each smaller: press Check form.

<!-- preview field 270 -->

```go
ryugo.Field(c, ryugo.FieldOptions{Label: "Email", Error: app.emailErr, Required: true, Shake: app.checks}, func() {
	ryugo.Input(c.Key("email"), &app.email, ryugo.InputOptions{Label: "Email", Invalid: app.emailErr != ""})
})
```

`Help` shows under the control while there is no `Error`. `Required` adds
a mark to the label. `Shake` is a counter: each time it grows, the field
shakes once, so a form's check can say where the problem is without a
word.
