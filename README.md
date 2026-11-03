# Mizan for MyGo

Mizan is a design system for calm, data-dense interfaces. This is its
native edition: components for [MyGo](https://github.com/egoist/mygo)'s
native UI (package `ui`), drawn by MyGo itself on macOS, Linux, Windows and
in the browser, in pure Go without cgo.

It ports Mizan's visual language, not its web code: the same OKLCH tokens,
the same two surfaces (a raised control that takes the light from above, an
inset well you type into), the same size steps and type scale. Behavior
(focus, keyboard, accessibility) comes from MyGo's unstyled `…Base` widgets;
Mizan gives them the look.

![Controls in the light theme](docs/controls-light.png)

## Install

```sh
go get github.com/imhalid/mizan-go
```

```go
import (
	"github.com/egoist/mygo/ui"
	mizan "github.com/imhalid/mizan-go"
)

func (a *app) view(c *ui.Context) {
	t := mizan.Root(c, mizan.Options{})
	mizan.Card(c, mizan.CardOptions{Title: "Report", Pad: true}, func() {
		mizan.Field(c, mizan.FieldOptions{Label: "Search"}, func() {
			mizan.Input(c.Key("q"), &a.query, mizan.InputOptions{Label: "Search", Clearable: true})
		})
		if mizan.Button(c.Key("go"), "Apply", mizan.ButtonOptions{Color: t.Colors.Accent, Variant: mizan.Solid}).Clicked() {
			a.apply()
		}
	})
	mizan.Toaster(c)
}
```

Run the gallery, a tour of every component:

```sh
go run github.com/imhalid/mizan-go/cmd/gallery@latest
```

## Theming

`Root` themes the window and `Scope` a region; fields left unset are
inherited, and the nearest scope wins.

- **Mode:** `Light`, `Dark`, or the system's appearance.
- **Skin:** `SkinMizan`, or `SkinBevel` (raised edges, metal knobs, an LED switch).
- **Size:** `Sm`, `Md`, `Lg`. A button and a field of the same step are the same height.
- **Accent:** a brand color. The accent ramp, selection fills and rings take its hue and chroma at Mizan's lightness steps, so contrast holds in both themes.

```go
mizan.Root(c, mizan.Options{Skin: mizan.SkinBevel, Accent: ui.Oklch(0.56, 0.2, 295)})
mizan.Scope(c, mizan.Options{Size: mizan.Sm}).Children(func() { ... })
```

The tokens are Go values in [tokens.go](tokens.go), written in OKLCH.

![Bevel skin with a violet accent, dark theme](docs/accent-violet-bevel-dark.png)

## Components

| Group | Components |
| --- | --- |
| Controls | `Button`, `Input`, `Select`, `Field`, `Card` |
| Selection | `Checkbox`, `Radio`, `Choice`, `Switch`, `Segmented`, `Slider` |
| Forms | `Textarea`, `NumberInput`, `Combobox`, `DatePicker`, `TagInput` |
| Navigation | `Tabs`, `TabPanel`, `Link` |
| Disclosure | `Collapsible`, `Accordion` |
| Toolbars | `Toggle`, `ToggleGroup`, `Separator` |
| Layers | `Dialog`, `Sheet`, `Popover`, `Menu`, `MenuButton`, `Tooltip`, `Toaster` |
| Chat | `Message`, `Typing`, `DaySeparator`, `Composer` |
| Drag and drop | `Sortable`, `DragCard`, `DropArea`, `DropZone`, `FileItem` |
| Indicators | `Badge`, `Tag`, `StatusDot`, `Kbd`, `Avatar`, `EmptyState` |
| Feedback | `Progress`, `Meter`, `Spinner`, `Skeleton`, `SkeletonText` |

![Chat and drag and drop, dark theme](docs/chat-drag-dark.png)

![Form fields with the date picker open, dark theme](docs/forms-dark.png)

## Layers

Layers move on quick springs: they spring in, and leave faster than they came.

- **Dialogs** drag by their header: they lean into the motion, glide on when thrown and spring back inside the window.
- **Dialogs and sheets** come in over a soft scrim that fades in with them: a dialog rises from 94%, a sheet slides in from the edge, tilted a little. Closing, they fade or slide out, and let the pointer through to the page at once.
- **Scroll edges:** content scrolling inside a dialog or sheet fades into its header and footer.
- **Glass:** popovers, menus, tooltips and toasts are frosted glass, from MyGo's `plugins/glass`.
- **Toasts:** they come from MyGo's `Context.AddToast`. A new one rises from below as it fades in; older ones tuck behind it and spread out under the pointer, all on springs.

![A dialog over its scrim](docs/dialog-light.png)

![A menu and a stack of toasts on glass](docs/menu-toasts-light.png)

## Development

```sh
go vet ./...
go test ./...         # interaction tests through ui.Tester
go test -short ./...  # skips the tests that draw glass on the CPU
MIZAN_SNAPSHOTS=/tmp/mizan go test -run 'Snapshot|SelectOpen' ./gallery/   # PNGs of every page
```

Everything builds with `CGO_ENABLED=0` for Linux, macOS, Windows and
`js/wasm`. CI runs all of it and keeps the snapshots as an artifact.

## License

[MIT](LICENSE)
