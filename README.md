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
- **Size:** `Sm`, `Md`, `Lg`. A button and a field of the same step are the same height.
- **Corners:** `CornersSm`, `CornersMd`, `CornersLg`: the radii of cards, fields, menus and tables at half, at, or one and a half times Mizan's. Pills and circles keep their shape, and a surface nested inside another keeps its corners concentric with it (`Concentric`: the outer radius less the padding).
- **Accent:** a brand color. The accent ramp, selection fills and rings take its hue and chroma at Mizan's lightness steps, so contrast holds in both themes.

```go
mizan.Root(c, mizan.Options{Accent: ui.Oklch(0.56, 0.2, 295), Corners: mizan.CornersLg})
mizan.Scope(c, mizan.Options{Size: mizan.Sm}).Children(func() { ... })
```

The tokens are Go values in [tokens.go](tokens.go), written in OKLCH.

![A violet accent, dark theme](docs/accent-violet-dark.png)

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
| Data | `Table`, `Pagination` |
| Workspace (experimental) | `Dock` (with `mizan/dockwin` for its windows) |
| Palettes (experimental) | `Palettes`, `Board`, `PaletteSet.Palette` (with `mizan/palettewin` for windows of their own) |

![Chat and drag and drop, dark theme](docs/chat-drag-dark.png)

![Form fields with the date picker open, dark theme](docs/forms-dark.png)

## Layers

Layers move on quick springs: they spring in, and leave faster than they came.

- **Dialogs** drag by their header: they lean into the motion, glide on when thrown and spring back inside the window.
- **Dialogs and sheets** come in over a soft scrim that fades in with them: a dialog rises from 94%, a sheet slides in from the edge, tilted a little. Closing, they fade or slide out, and let the pointer through to the page at once.
- **Scroll edges:** content scrolling inside a dialog or sheet fades into its header and footer.
- **Drag and drop:** a card picked up lifts and follows the pointer, leaning into the motion; the cards where it would go slide apart, and it flies into its place when let go. A column taking it glows, a cloud's shadow of the accent drifting across it, and a light runs up both its sides as it lands. Sortable rows also move with Alt+Up and Alt+Down.
- **Table:** rows built only as far as they show, sorted by a click on a header (Shift-click adds a key), chosen with checkboxes (Shift-click for a range) under a selection bar that springs in over the toolbar. A column picked up by its header lifts and the others slide aside for it; its end edge resizes it. Rows slide to their new places as the sort changes, new ones fade in, removed ones fade where they were, and a page turns sideways. Search, a view menu for columns and density, opening rows for details, a pinned first column, pages, and loading, empty and error states come with it.
- **Glass:** popovers, menus, tooltips and toasts are frosted glass, from MyGo's `plugins/glass`.
- **Toasts:** they come from MyGo's `Context.AddToast`. A new one rises from below, growing to its size, as two lights of Mizan's border beam run round it from opposite sides; older ones tuck behind it and spread out under the pointer. A swipe to the side sends one away, and a `ToastLoading` toast turns into its result in place.

![A dialog over its scrim](docs/dialog-light.png)

![A menu and a stack of toasts on glass](docs/menu-toasts-light.png)

## Dock (experimental)

**Experimental:** what `Dock` is for isn't settled yet; its API may change
or go without notice.

`Dock` fills a window with areas as Blender does: edge to edge, every
edge shared. Its layout is plain data the app owns (`Layout`), saved as
JSON and read back as it was; a pane is a view of the app's data, never
its holder, so it goes anywhere without losing what it shows.

- **Edges** drag, and the edges lined up with them move along; Option
  moves one alone, cut where edges across it meet it. An edge dragged
  settles onto an edge it would line up with, and a light runs along it.
  Where edges meet, a drag moves both ways.
- **Corners** split an area (dragged inward) or join it to a neighbor
  (dragged into it). Undo and Redo go back and forth through every change.
- **Tabs:** every pane is a tab of its area's stack, the chosen one on a
  raised pill that slides between them, as Segmented's. Dragged, a tab lifts;
  over an area, a compass in its middle and targets on the window's edges
  say where it would dock, and a region shows it. A right click splits the
  area with it, sends it to a window of its own, or closes it; a double
  click maximizes the area (Ctrl+Space too).
- **Windows:** with a `DockHost` (`dockwin.New`, on MyGo's windows), a tab
  dragged out of the window tears off into a window of its own where it is
  let go, and docks into another window it is let go over. Windows torn
  off have no title bar: their headers move them, a window's only tab
  carries it to dock it elsewhere, and their top right area closes them,
  giving what they hold back to the main one. Windows keep their places
  in the layout. On Wayland, where the system doesn't say where windows
  are, tabs tear off but don't dock from one window into another by a
  drag; their menu moves them back.

```go
layout := mizan.NewLayout(mizan.DockSplit(mizan.DockRow, []float32{0.25, 0.75},
	mizan.DockArea(&mizan.Pane{Kind: "files", Title: "Files"}),
	mizan.DockArea(&mizan.Pane{Kind: "editor", Title: "main.go"}),
))
mizan.Dock(c.Key("dock"), layout, mizan.DockOptions{
	Kinds:    []mizan.DockKind{{Kind: "files", Title: "Files", Build: app.files}, {Kind: "editor", Title: "Editor", Build: app.editor}},
	OnChange: app.saveLayout,
}).Fill()
```

![A tab docking onto the right half of an area](docs/dock-light.png)

## Palettes (experimental)

Tool panels floating over a canvas, as Pixelmator Classic's: a dark glass
HUD in both themes, a thin title bar with a round close button. A palette
moves by its title bar, lifting as it goes, and settles onto the board's
edges and onto the palettes near it, as a magnet would; a light runs along
the edge it settles onto. A press brings it to the front, a double click on
its title folds it up, and closed, it shrinks away. Where the palettes are
is plain data (`Board`), to save. Given a host (`palettewin.New`, on
MyGo's windows), they float in windows of their own over the app's, as
Pixelmator Classic's, settling onto the screen's edges, the app's window
and each other; where the system doesn't let windows go where asked
(Wayland), they float over their board in the app's window. Experimental,
as the Dock.

```go
mizan.Palettes(c.Key("studio"), &app.board, func(p *mizan.PaletteSet) {
	app.canvas(c)
	p.Palette("layers", mizan.PaletteOptions{Title: "Layers", At: ui.Point{X: -12, Y: 12}}, app.layers)
}).Fill()
```

![Palettes over a canvas](docs/palettes-dark.png)

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
