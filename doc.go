// Package mizan provides the components of the Mizan design language for
// MyGo's native UI (package ui). It ports the web library's visual language,
// not its code: the same tokens (written directly in tokens.go as OKLCH
// values), the same surface language (raised control, inset well, card
// edge), the same size steps and type scale.
//
// Behavior (focus, keyboard, accessibility) comes from MyGo's unstyled
// components (ButtonBase, TextInputBase, SelectBase…); this package only
// provides the look.
//
//	func (a *app) view(c *ui.Context) {
//		t := mizan.Root(c, mizan.Options{})
//		mizan.Card(c, mizan.CardOptions{Title: "Report", Pad: true}, func() {
//			mizan.Input(c.Key("q"), &a.query, mizan.InputOptions{Placeholder: "Search"})
//			if mizan.Button(c.Key("go"), "Apply", mizan.ButtonOptions{Color: t.Colors.Accent, Variant: mizan.Solid}).Clicked() {
//				...
//			}
//		})
//	}
//
// The theme (Light, Dark), size (Sm, Md, Lg) and accent color are given
// to the whole window with Root and to a region with Scope; the nearest
// scope wins. Components read the values with From(c).
package mizan
