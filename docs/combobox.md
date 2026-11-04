# Combobox

A select you type into. The field shows the chosen option's label; typing
filters the options, those starting with the text first, and a click,
Down or the chevron shows them all. Up and Down move, Enter or a click
chooses, Escape closes, and text that matched nothing goes back to the
choice as the field loses the focus.

<!-- preview combobox 340 -->

```go
mizan.Combobox(c.Key("owner"), &app.owner, people, mizan.ComboboxOptions{Label: "Assignee", Placeholder: "Assign to…"})
```

Its options are a [Select](select.md)'s, `[]mizan.Option[T]`, with their
descriptions, dots and disabled ones. It returns the text input.
