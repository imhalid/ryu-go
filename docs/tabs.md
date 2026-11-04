# Tabs

A row of tabs over a groove. The chosen tab's label is in the text color,
with a 2 px underline that slides to a new choice on a spring, and a soft
fill follows the pointer from tab to tab. Build the chosen tab's panel in
`TabPanel` under it: a new panel fades in, rising a little.

<!-- preview tabs 160 -->

```go
ryugo.Tabs(c.Key("tabs"), &app.tab, []ryugo.Tab{{Label: "Overview"}, {Label: "Activity", Badge: "12"}}, ryugo.TabsOptions{Label: "Report"})
ryugo.TabPanel(c, app.tab, func() {
	...
})
```

A tab has an `Icon` and a `Badge`, a count beside its label. The arrows,
Home and End move the choice while a tab has the focus.
