# Workspace

**Experimental:** what `Dock` is for isn't settled yet; its API may change
or go without notice.

`Dock` fills its place with areas as Blender does: edge to edge, every edge
shared. Drag the gaps between areas, split an area by dragging its corner
inward, join two by dragging a corner into a neighbor, and drag a tab onto
another area's tabs, middle or sides to dock it there. Undo and Redo go back
and forth through every change.

<!-- preview workspace 560 -->

```go
layout := ryugo.NewLayout(ryugo.DockSplit(ryugo.DockRow, []float32{0.25, 0.75},
	ryugo.DockArea(&ryugo.Pane{Kind: "files", Title: "Files"}),
	ryugo.DockArea(&ryugo.Pane{Kind: "editor", Title: "main.go"}),
))
ryugo.Dock(c.Key("dock"), layout, ryugo.DockOptions{
	Kinds: []ryugo.DockKind{
		{Kind: "files", Title: "Files", Build: app.files},
		{Kind: "editor", Title: "Editor", Build: app.editor},
	},
	OnChange: app.saveLayout,
}).Fill()
```

## Layout

The layout is plain data the app owns (`Layout`), saved as JSON
(`MarshalJSON`) and read back as it was (`ParseLayout`). A pane is a view
of the app's data, never its holder, so it goes anywhere without losing
what it shows: a `DockKind` builds every pane of its kind, given the pane.
`OnChange` runs after every change, to save it.

## Edges, corners and tabs

- **Edges** drag, and the edges lined up with them move along; Option moves
  one alone. An edge dragged settles onto an edge it would line up with, and
  a light runs along it. Where edges meet, a drag moves both ways; a double
  click shares the room out.
- **Corners** split an area (dragged inward) or join it to a neighbor
  (dragged into it).
- **Tabs:** every pane is a tab of its area's stack, the chosen one on a
  raised pill that slides between them. Dragged, a tab lifts; over an area,
  a compass in its middle and targets on the edges say where it would dock,
  and a region shows it. A right click splits the area with it or closes
  it; a double click maximizes the area (Ctrl+Space too).

## Windows

On a desktop, with a `DockHost` (`dockwin.New`, on MyGo's windows), a tab
dragged out of the window tears off into a window of its own, and docks
into another window it is let go over; windows keep their places in the
layout. In a browser, as here, there are no windows to tear off into, and
tabs stay in the page.
