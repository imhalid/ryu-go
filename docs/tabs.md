# Tabs

A row of tabs over a groove. The chosen tab's label is in the text color,
with a 2 px underline that slides to a new choice on a spring, and a soft
fill follows the pointer from tab to tab. Build the chosen tab's panel in
`TabPanel` under it: a new panel fades in, rising a little.

<!-- preview tabs 160 -->

```go
mizan.Tabs(c.Key("tabs"), &app.tab, []mizan.Tab{{Label: "Overview"}, {Label: "Activity", Badge: "12"}}, mizan.TabsOptions{Label: "Report"})
mizan.TabPanel(c, app.tab, func() {
	...
})
```

A tab has an `Icon` and a `Badge`, a count beside its label. The arrows,
Home and End move the choice while a tab has the focus.
