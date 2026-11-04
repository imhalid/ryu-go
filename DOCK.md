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
- **Cutting an edge** (after the user tried it on Linux): the tree makes an edge between two splits one edge all along (a row of two columns whose edges line up: the edge between the columns runs the whole height), so Option alone couldn't move a part of it. Now, Option-pressed where the edges across it meet it from both sides, it is cut there (`cutPlan`, `Layout.cutEdge`): the two columns become a column of rows, one per part, laid out as before, and the part under the pointer moves. With Option held over such an edge, its grip shows in the middle of the part that would move. What takes the pointer on the edges is now built over the window (`hits`), from the last frame's layout, so that a press outlives the change of the tree; the gap and its grip stay in the flow. The nodes a cut makes take their sizes at once (`settled`).
- **Settling** (after the user asked for a magnet): an edge dragged settles onto an edge it would line up with (one whose end meets its end, along the same line) within 8px, holds there until the pointer is 16px away, and glides on and off over 240 ms (`springGuide`); as it lands, the split line's light runs along it from the pointer. The edge follows the pointer by where it is from where it pressed, not by adding up its moves: what takes the pointer moves with the edge.
- **What will move shows first:** under the pointer, the grips show on every edge that would move (on all of a line, or only the one with Option held), before it is pressed.
- Tried by the user in the real window (edges, lined-up edges, crossings, Option): no problem seen.

### 2. Split and join from corners

- [x] Corner zones on every area; a drag inward splits, along the axis the pointer leaves the corner by.
- [x] The split line follows the pointer and springs into place when let go.
- [x] A drag from a corner into a neighbor joins: the area to be taken in darkens, an arrow shows the direction; let go, the neighbor grows over it on a spring.
- [x] Undo of a split or a join (the tree is data: keep the last few).

Done when: splitting and joining feel like Blender's, measured frame by frame and confirmed in the real window.

**Result (2026-11-04, `dock_corner.go`):** measured frame by frame in the tester, light and dark; tried by the user in the real window ("good for now"; the snap feedback is waiting, above).

- **Corners:** a 14px square in each area's corner (crosshair cursor); under the pointer a short bracket draws itself from its middle out along the corner, concentric with the area's radius.
- **Split:** the axis is the one the pointer leaves the corner by (decided 8px out), and turns only when the other clearly wins (1.6×). The line grows out from the pointer across the area (`springGuide`), follows it, and settles onto the area's middle or a lined-up edge within 8px (a weight on a spring, so it glides on and off). Let go, the edge's middle is where the line was; the new area is on the corner's side with a copy of the pane (its scroll's place, not the `ScrollState`, which is bound to one view). Its surface was already there; only what it shows fades in.
- **Join:** only into a neighbor sharing a whole edge (`tidy` keeps those siblings). The neighbor darkens (`Scrim`), an arrow grows in over it pointing the way; between neighbors the darkening slides on a spring. Let go, the area folds away on `springShift` with its edge, the darkening drawn by the folding area itself (from the guide it would be a frame behind), and the layout changes as the next frame starts.
- **Undo:** `Layout.Undo` and `CanUndo`, the last 16 changes; the gallery has an Undo button. An undone join unfolds on the springs; an area an undo (or a restore) takes away folds away where it was, darkened as a join's, toward its far side from the sibling that takes its room (`vanished`, `ghosts`).
- **Settling** (after the user tried it): the split line settles onto the middle or a lined-up edge within 8px and holds there until the pointer is 16px away; as it lands, a light in the accent (lifted toward white in the dark theme) runs out from the pointer along the line to both ends, once (620 ms, `EaseInOut`).


### 3. Headers and maximize

- [x] An area's header with the kind menu (the app's kinds), and the pane's title.
- [x] Ctrl+Space maximizes the area over the window and back, with a growth from where it was.
- [x] Keyboard: moving the focus between areas.

**Result (2026-11-04, `dock_area.go`):** measured frame by frame in the tester, light and dark; tried by the user in the real window.

- **Kinds** are now `[]DockKind{Kind, Title, Build}`: the order is the menu's. The header shows the kind's title and a chevron (a menu of the kinds, the current one checked), then the pane's title when it differs, and at the end a maximize button that shows under the pointer. The header's buttons are quiet: a fill under the pointer, a 4% shrink when pressed, corners concentric with the area's. A kind chosen replaces the pane with a new one of it; Undo brings the old one back.
- **Maximize:** `DockWindow.Max` (saved with the layout). The area keeps its place in the tree as a bare surface while its content shows in a copy over the window that grows from that place (`springOpen`), lifting with a shadow while it moves; brought back, it shrinks to the same place, then the content goes back into the tree. Ctrl+Space acts on the area under the pointer, else the one focused. **On macOS, Ctrl+Space is the system's input-source switch** by default and may never reach the app; the header's button always works.
- **Focus:** an area takes the keyboard focus as a whole (a press in it that no control takes gives it the focus); its kind's title goes from Text2 to Text. Ctrl+Alt+arrows move the focus to the nearest area that way.
- The corners' squares are 10px now, so that the header's buttons near them stay pressable.


### 4. Tabs and docking

- [x] Tab stacks on Tabs, with close buttons; tabs reorder within their stack (as Sortable does, the others sliding aside).
- [x] A tab dragged over an area shows where it would go: its left, right, top or bottom half, or its middle (a new tab). The preview is a translucent region that slides between zones on a spring; no blue line.
- [x] Let go, the pane docks there: the split opens on a spring, its neighbors making room.
- [x] An area left empty folds away, its neighbor taking its room.

**Result (2026-11-04, `dock_tabs.go`):** measured frame by frame in the tester, light and dark; not yet tried in the real window.

- **The header** is now the kind's icon and a chevron (its menu, as Blender's editor type; `DockKind.Icon`, a generic one when nil), the tabs, and the maximize button. Every pane is a tab, one alone too, so that it can be dragged. The tab chosen is in Text (Text2 when its area isn't focused), the others Muted; with more than one, a 2px underline slides to the choice (`springPill`), as Tabs' does. The close button shows on the tab chosen and the one under the pointer. Ctrl+PageUp and Ctrl+PageDown choose the tab before and after.
- **Picking up:** a tab dragged 4px lifts off as a raised tile (`springLift`, `ShadowPop`, leaning with its speed); its place stays empty, and the tabs after it slide over to close the room (`springShift`).
- **Where it goes:** over an area's tabs, they slide apart where it would go (a tab in that stack); over the rest of an area, the side the pointer is nearest within 30% of it docks it beside the area (left, right, top, bottom; only when the area is large enough for two), else its middle puts it at the end of the stack. Over its own area, alone there, it goes nowhere. The region is a wash of the accent with a hairline of it (`ColorAccent500`, 10% and 55%; 16% and 70% in the dark theme), inset 3px with a concentric radius, sliding between zones and areas on `springPop`, fading in 140 ms.
- **Let go:** the move applies as the next frame starts (as a drop does), every stack building the new order together; the tab flies from where its copy was to its place once a frame laid it out (`springLand`), and the underline slides to it. Docked beside an area, the new area opens from nothing on `springShift`, its neighbor making room, and what it shows fades in. Nowhere, the tab flies back.
- **The compass** (after the user tried it, from Visual Studio's docking guides, in Mizan's language; `dock_compass.go`): over the area under the pointer, a plus of five raised targets (30px, the control's surface) on a plate (the popover's color, `ShadowPop`, its corners concentric with the targets'), in the middle of the area's body: the middle puts the tab in the stack, the arms dock it beside the area. Each target draws what it does: a small pane, the part it takes filled in the accent. A target that can't apply (the middle over the tab's own area, an arm where the area is too small to split) shows faded. The compass grows in on `springPop` (from 92%), glides from area to area on it, and goes in 140 ms; it shows only over an area large enough for it, and keeps clear of the window's edge targets. Under the pointer a target grows 8% (`springSqueeze`) and fills in, the region shows behind, and the tab's copy lets it show through (55%). The compass is built above the copy.
- **The window's edges:** four targets in the middle of the window's edges (the top one under the headers, so that it hides no tab) dock the tab along the whole of that edge, a quarter of the window long (`dockPane` into the root). They fade in with the drag.
- Targets win over the nearness of an area's side, which stays as the quick way.
- **Closing:** the pane goes at once (Undo brings it back); where its tab was, a copy narrows to nothing as it fades (200 ms) and the tabs after it slide over.
- **An area left empty** folds away as a join does, without the darkening, its neighbor before it (else after it) taking its room. A window's last area stays, saying "Drag a tab here".
- **The model:** `movePane`, `dockPane`, `closePane`, `removeArea`; every change by the user is one step of Undo.

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

Polish waiting (from the user's tries; do them in a polish pass): none now.

## Limits

- Moving panes between windows depends on step 0; if MyGo loses the pointer outside the window, tear-off is an action, not a drag.
- Some timing bugs may show only in the real window, as with drag and drop.
