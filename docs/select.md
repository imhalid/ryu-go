# Select

A choice from a list, without search: the trigger is a raised field (a
button's skin, a field's size), the list a panel of glass. The highlighted
option fills with the accent gradient, and a check draws itself by the
chosen one as the list opens.

<!-- preview select 300 -->

```go
ryugo.Select(c.Key("range"), &app.days, []ryugo.Option[int]{
	{Value: 7, Label: "Last 7 days"},
	{Value: 30, Label: "Last 30 days", Description: "default"},
	{Value: 365, Label: "Last year", Disabled: true},
}, ryugo.SelectOptions{Label: "Period"})
```

The value is any comparable type: the option whose `Value` equals it is
the chosen one. An option has a `Description`, muted beside its label, a
`Dot` of a series color before it, and can be `Disabled`. With nothing
chosen, the trigger shows `Placeholder`. A list too long for the window
scrolls, fading at its edges where there is more. To type and filter, use
a [Combobox](combobox.md).
