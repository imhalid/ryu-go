# Textarea

Many lines of text in Input's inset well, as high as its text: from `Rows`
lines up to `MaxRows`, then it scrolls. With `MaxLength` or `Count`, the
number of characters stands in its bottom corner.

<!-- preview textarea 290 -->

```go
ryugo.Textarea(c.Key("bio"), &app.bio, ryugo.TextareaOptions{Label: "Bio", Placeholder: "A few words about you", Rows: 2, MaxLength: 80})
```

It returns the text area, for `Changed` and `Focus`.
