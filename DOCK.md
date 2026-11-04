# Dock: a Blender-style workspace

The plan for Mizan's area layout: a window filled edge to edge with areas that split, join, resize, hold tabs, and tear off into windows of their own, as Blender's do, with no data ever lost. This file is the plan and the to-do list; tick items off as they land, and keep the rules of [AGENTS.md](AGENTS.md) (motion, verification, no color on press).

The working name of the component is `Dock`. It is written by us, on MyGo's native windows and our own drag layer (`dnd.go`); MyGo has no docking component.

## What it should feel like

- **The window is all areas.** No gaps, no floating panels: the window is split into areas from edge to edge. The edges between areas are shared lines: dragging one resizes every area along it.
- **Split from a corner.** Dragging inward from an area's corner splits it in two. The split line follows the pointer, and springs into place when let go.
- **Join toward a neighbor.** Dragging a corner toward a neighbor joins the two: the area to be taken in darkens, and an arrow over it shows the direction, as in Blender.
- **An area's header.** Each area has a header with a menu choosing what it shows (a table, a chart, a console…).
- **Maximize.** Ctrl+Space grows an area over the whole window, and the same key brings it back, both with a soft growth.
- **Tabs.** An area can hold several panes as tabs; a tab can be moved within its stack, onto another area's edge or middle (docking), or out of the window.
- **Tear-off.** A pane or tab dragged out of the window opens a new window where it is let go, and moves there. It can be dragged back into a window and docked.

## Data never lost

This is a hard requirement.

- **Data lives in the model.** A pane's data is never held by a window: it is the app's, and a pane is a view of it. A pane changing windows doesn't touch it.
- **Pane state travels with the pane.** What a pane shows that isn't the app's data (its scroll, its chosen row, its open tab) lives in the layout tree with the pane, not in the window's element state, so it goes wherever the pane goes.
- **Closing a window loses nothing.** As a torn-off window closes, its close is cancelled long enough to dock its panes back into the main window. A user can't lose a pane by closing a window.
- **The layout is saved.** The whole layout, windows and their positions included, is plain data, saved on every change and restored as it was when the app opens again.

## Model

- A **layout** is a set of windows, each holding a tree.
- A **tree** node is a **split** (a row or a column of children, with their sizes in fractions) or an **area**.
- An **area** holds a stack of **panes** (tabs) and the one shown.
- A **pane** has a stable ID, a kind (what it shows, which the app builds), and its view state (scroll, selection…).
- The app gives the Dock a builder per kind: `func(c *ui.Context, p *Pane)`.
- The layout serializes to JSON and back without loss; unknown kinds are kept, not dropped.

## Steps

### 0. Spike: out of the window and back

The one unknown: whether MyGo keeps reporting the pointer while a drag leaves the window. Its API doesn't say. Prove it first.

- [x] Drag an element past the window's edge; log the pointer's position (window position + local pointer) while outside.
- [x] On release outside, open a native window (`mygo.NewWindow` with `Content`) at the release point, showing a pane built from the shared model.
- [x] Close that window: move the pane back into the main window as it closes.
- [x] Drag between two windows of the app and drop (by our own geometry, not MyGo's typed drag).
- [x] ~~If the pointer is lost outside the window: fall back to an "Open in new window" action~~: not needed, the pointer is not lost.

Done when: a pane goes out to its own window and back without losing its scroll or selection, in the real window (not only the tester).

**Result (2026-11-04, macOS, `cmd/dockspike`): it works.** Tried by the user in the real window, who saw no problem; from the log:

- **The pointer is followed outside the window.** While the bar is pressed, `Dragged` and `PointerPosition` keep reporting past the window's edges, with negative or larger-than-the-window local coordinates. The pointer on the screen is `Window.ContentBounds()` + the element's `LayoutBounds` + `PointerPosition`.
- **Coordinates:** window and content bounds are in screen points with the origin at the top left; the content starts 32 points under the window's top (the title bar). macOS may nudge a new window's position a little.
- **Opening:** `mygo.NewWindow` is called from a goroutine (it waits for the main thread), never from inside a frame; the pane's model says which window it is in before the window exists, and its view shows it once it does.
- **Between windows:** the drop finds the window under the pointer by its content bounds; the pane moves there and an emptied torn-off window closes itself.
- **Closing:** `OnClose` moves the window's panes back to the main window; nothing was lost (counts and scrolls in the log matched before and after every move).
- **Redrawing:** a change of the model redraws every window with `Window.Invalidate()`.
- **Not yet tried:** Windows and Linux; multiple displays; dropping onto a window behind another.

The spike stays in `cmd/dockspike` as a reference until step 5 replaces it.

### 1. Areas: splits and shared edges

- [x] The layout tree and its JSON form, with tests for round trips.
- [x] Areas filling the window edge to edge; splits laid out from fractions.
- [x] Edges between areas: dragging one resizes all the areas along the shared line; minimum sizes; a double click shares the room equally.
- [x] Edge hover and drag show motion only (the edge grows a little), no color fill.
- [x] Gallery: a "Workspace" page with a fixed layout (files on the left, editors in the middle, a console below).

Done when: edges move smoothly with no jump on release, and the layout survives a save and a restore.

**Result (2026-11-04):** `dock_layout.go` holds the model (`Layout`, `DockWindow`, `Node`, `Pane`; `ParseLayout` mends sizes, IDs and the active pane rather than fail), `dock.go` the view, and the gallery's Workspace page shows it with Save layout, Restore and Reset. Tested in `dock_test.go` (round trip, mending) and `gallery/workspace_test.go` (a drag, the least size, a double click, lined-up edges moving as one).

- **Lined-up edges** are found at the end of the frame the edge is pressed in, from that frame's edges: the last frame's are not laid out yet in a frame that didn't rebuild.
- **Sizes** show through springs (a double click glides), but under the pointer they follow it at once (`Jump`), and the spring takes over from where the pointer left them: no jump on release.
- **Crossings** (after the user tried it): where edges meet, a 12px square takes the pointer and drags both ways at once, the vertical edge across, the edges across it along, each with the edges lined up with it. A small cross grows in it under the pointer. Crossings are placed from the last frame's edges; when those moved, the dock asks for one more frame, or a crossing would stay where the layout was.
- **One edge alone:** with Option (Alt) held as it is pressed, an edge (or a crossing) moves only itself, not the edges lined up with it. Lined-up edges stay linked by default, as in Blender. Not testable in the tester: its pointer events carry no modifiers.
- **What will move shows first:** under the pointer, the grips show on every edge that would move (on all of a line, or only the one with Option held), before it is pressed.
- Tried by the user in the real window (edges, lined-up edges, crossings, Option): no problem seen.

### 2. Split and join from corners

- [ ] Corner zones on every area; a drag inward splits, along the axis the pointer leaves the corner by.
- [ ] The split line follows the pointer and springs into place when let go.
- [ ] A drag from a corner into a neighbor joins: the area to be taken in darkens, an arrow shows the direction; let go, the neighbor grows over it on a spring.
- [ ] Undo of a split or a join (the tree is data: keep the last few).

Done when: splitting and joining feel like Blender's, measured frame by frame and confirmed in the real window.

### 3. Headers and maximize

- [ ] An area's header with the kind menu (the app's kinds), and the pane's title.
- [ ] Ctrl+Space maximizes the area over the window and back, with a growth from where it was.
- [ ] Keyboard: moving the focus between areas.

### 4. Tabs and docking

- [ ] Tab stacks on Tabs, with close buttons; tabs reorder within their stack (as Sortable does, the others sliding aside).
- [ ] A tab dragged over an area shows where it would go: its left, right, top or bottom half, or its middle (a new tab). The preview is a translucent region that slides between zones on a spring; no blue line.
- [ ] Let go, the pane docks there: the split opens on a spring, its neighbors making room.
- [ ] An area left empty folds away, its neighbor taking its room.

### 5. Windows and saving

- [ ] Tear-off of a pane or a tab into its own window (from step 0), and docking back from it.
- [ ] Moving a whole area to a new window (Shift-drag from a corner, as Blender's duplicate).
- [ ] Closing a torn-off window docks its panes back into the main window.
- [ ] Saving the layout on every change; restoring windows, positions and panes on open; a window placed off every display comes back onto one.

### 6. Gallery and docs

- [ ] The Workspace page showing all of it, with the table, charts and a console as kinds.
- [ ] README: a section on Dock, with a screenshot.
- [ ] Interaction tests for split, join, resize, docking, maximize, and the save and restore round trip.

## Polish to get right

The previews are where it feels like Blender or not; give them the most time.

- The split line, the join arrow and the docking region move on springs, never jump between positions.
- What is held lifts (as cards do); what makes room slides aside.
- Nothing appears or vanishes in a frame: areas open and fold, windows' panes arrive and leave with motion.
- Measure every gesture frame by frame in the tester, then confirm in the real window: drag-and-drop taught that some bugs only show there.

## Limits

- Moving panes between windows depends on step 0; if MyGo loses the pointer outside the window, tear-off is an action, not a drag.
- Some timing bugs may show only in the real window, as with drag and drop.
