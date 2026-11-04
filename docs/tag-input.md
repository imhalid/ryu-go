# Tag input

Words as neutral pills in Input's well, before a bare input. Enter or a
comma adds what was typed, unless it is there already (ignoring case);
each pill's button takes it out, and Backspace in the empty input takes
out the last. The well grows a line as the tags fill one.

<!-- preview tag-input 250 -->

```go
mizan.TagInput(c.Key("tags"), &app.tags, mizan.TagInputOptions{Label: "Tags", Placeholder: "Add a tag"})
```

`Max` stops it taking more.
