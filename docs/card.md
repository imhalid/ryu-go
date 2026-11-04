# Card

A surface whose title stands outside the panel, above it. The panel is in
the card color with large corners, and its border fades from the top down;
in the dark theme, where shadows don't show, the border sets it apart.

<!-- preview card 210 -->

```go
mizan.Card(c, mizan.CardOptions{Title: "Weekly report", Subtitle: "Updated 5 minutes ago", Pad: true, Actions: func() {
	mizan.Button(c.Key("share"), "Share", mizan.ButtonOptions{Icon: shareIcon, IconOnly: true, Size: mizan.Sm})
}}, func() {
	...
})
```

`Pad` gives the panel its padding; leave it out for content that goes to
the edges, as a table. `Actions` builds buttons on the title's row.
