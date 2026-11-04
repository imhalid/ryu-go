# Input

The inset text field: an opaque well with a thin border, which lightens to
the surface and takes an accent ring as it gets the focus. It returns the
text input, for `Changed`, `Submitted` and `Focus`.

<!-- preview input 260 -->

```go
mizan.Input(c.Key("q"), &app.query, mizan.InputOptions{Label: "Search", Placeholder: "Search", PrefixIcon: searchIcon, Clearable: true})
```

## Options

- `Label` names it to assistive technology; around it, a `Field` shows one.
- `PrefixIcon`, `Prefix` and `Suffix` sit inside the well, before and after
  the text: an icon, or a unit such as `$` and `USD`.
- `Clearable` adds a button clearing the text while there is some.
- `Password` hides the text, with a button showing it.
- `Invalid` marks it in the Up tone; `Disabled` dims it.
