# Checkbox

An 18 px square, inset while empty; checked, it fills with the accent
gradient and its check flows from the short arm to the long one as a pen
would draw it. Unchecked, the check draws itself back the way it came. The
whole row, label and description, takes the click.

<!-- preview checkbox 200 -->

```go
ryugo.Checkbox(c.Key("mail"), &app.mail, "Newsletter", ryugo.CheckOptions{})
ryugo.Checkbox(c.Key("all"), &all, "All notifications", ryugo.CheckOptions{Indeterminate: app.mail != app.push})
```

`Indeterminate` shows a dash for a parent whose children differ; `Changed`
on the returned element reports a click, to set the children. `Description`
adds a muted line under the label, and `Disabled` dims it.
