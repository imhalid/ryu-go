// Package ryugo provides the components of the Ryugo design language for
// MyGo's native UI (package ui): tokens written directly in tokens.go as
// OKLCH values, one surface language (raised control, inset well, card
// edge), three size steps and one type scale.
//
// Behavior (focus, keyboard, accessibility) comes from MyGo's unstyled
// components (ButtonBase, TextInputBase, SelectBase…); this package only
// provides the look.
//
//	func (a *app) view(c *ui.Context) {
//		t := ryugo.Root(c, ryugo.Options{})
//		ryugo.Card(c, ryugo.CardOptions{Title: "Report", Pad: true}, func() {
//			ryugo.Input(c.Key("q"), &a.query, ryugo.InputOptions{Placeholder: "Search"})
//			if ryugo.Button(c.Key("go"), "Apply", ryugo.ButtonOptions{Color: t.Colors.Accent, Variant: ryugo.Solid}).Clicked() {
//				...
//			}
//		})
//	}
//
// The theme (Light, Dark), size (Sm, Md, Lg) and accent color are given
// to the whole window with Root and to a region with Scope; the nearest
// scope wins. Components read the values with From(c).
package ryugo
