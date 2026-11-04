# Button

One kind of raised pill, lit from above: a color, a fill and a size make
the rest. A press moves it, never its color. `Clicked` reports a press, in
the frame it happened.

<!-- preview button 190 -->

```go
if ryugo.Button(c.Key("export"), "Export", ryugo.ButtonOptions{Color: t.Colors.Accent, Variant: ryugo.Solid}).Clicked() {
	app.export()
}
```

## Color and fill

Without a `Color` a button is neutral. With one, it is tinted (`Soft`, the
default for a color) or filled (`Solid`). `Danger` makes it the Up tone,
filled unless `Variant` says `Soft`. The series colors `C1`…`C8` of the
tokens give a row of buttons for a chart's legend.

<!-- preview button-colors 150 -->

## Options

- `Icon` puts an icon before the label; with `IconOnly` the label is for
  assistive technology alone, and the button is round.
- `Size`: `Sm`, `Md` or `Lg`, the height of a field of the same step.
- `Loading` shows a spinner in place of the icon and takes no presses.
- `Disabled` dims it.
