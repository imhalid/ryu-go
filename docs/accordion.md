# Accordion

A stack of collapsible sections divided by grooves, as a settings page's.
Each opens and closes on its own, or, with `Single`, opening one closes the
others.

<!-- preview accordion 270 -->

```go
mizan.Accordion(c.Key("settings"), mizan.AccordionOptions{Single: true}, func(a *mizan.AccordionList) {
	a.Item("Billing", &app.billing, func() { ... })
	a.ItemWith("Security", &app.security, mizan.CollapsibleOptions{Description: "Two-factor sign-in, sessions"}, func() { ... })
})
```

`Item` adds a section; `ItemWith` one with a collapsible's options.
