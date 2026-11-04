# Date picker

A raised trigger, as Select's, that opens a month's calendar below it. The
chosen day fills with the accent gradient, today is written in the accent
color, and the days of the months around are muted.

<!-- preview date-picker 420 -->

```go
ryugo.DatePicker(c.Key("due"), &app.due, ryugo.DatePickerOptions{Label: "Due date", Min: app.start})
```

A click, or the arrows and Enter, choose a day and close it; Page Up and
Page Down move by months, Home and End to the month's first and last day.
The calendar is in the window's language, its weeks starting on the first
day of the user's regional formats. `Min` and `Max` bound the days that
can be chosen; a zero `*date` shows `Placeholder`, and `Format` writes the
chosen day otherwise. A chosen day keeps the time of day and location of
`*date`.
