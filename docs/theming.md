# Theming

`Root` themes the window and `Scope` a region; fields left unset are
inherited, and the nearest scope wins. Choose a size, corners and an accent
here: the card below is in a `Scope` made of them.

<!-- preview theming 460 -->

```go
ryugo.Root(c, ryugo.Options{Accent: ui.Oklch(0.56, 0.2, 295), Corners: ryugo.CornersLg})
ryugo.Scope(c, ryugo.Options{Size: ryugo.Sm}).Children(func() { ... })
```

## Mode

`Light`, `Dark`, or, left unset (`ModeInherit`), the system's appearance.
On this site the previews follow the page's theme: switch it in the header.

## Size

`Sm`, `Md` (the default) or `Lg`. A button and a field of the same step are
the same height, so a row of them lines up whatever the step.

## Corners

`CornersSm`, `CornersMd` or `CornersLg`: the radii of cards, fields, menus
and tables at half, at, or one and a half times Ryugo's. Pills and circles
keep their shape, and a surface nested inside another keeps its corners
concentric with it (`Concentric`: the outer radius less the padding).

## Accent

A brand color. The accent ramp, selection fills and rings take its hue and
chroma at Ryugo's lightness steps, so contrast holds in both themes. `Swatch`
gives the accent a color turns into, for a picker of brand colors.

## Tokens

`From(c)` returns the tokens of the nearest scope: `Colors` (text, surfaces,
the accent, the series colors `C1`…`C8`, and the tones `Down`, `Warn` and
`Up`), `Shadows`, `Metrics` and `R`, which scales a radius by the scope's
corners. They are Go values in
[tokens.go](https://github.com/imhalid/ryu-go/blob/main/tokens.go),
written in OKLCH. `Text` writes in Ryugo's type scale:

```go
ryugo.Text(c, "Weekly active users", ryugo.TTitle)
ryugo.Text(c, "Updated 5 minutes ago", ryugo.TCaption).TextColor(t.Colors.Muted)
```
