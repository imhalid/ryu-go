# Dock: a Blender-style workspace

**Experimental** (2026-11-04, the user's call): what the Dock is for isn't settled yet. It is marked so in its doc comment, the README and the gallery's Workspace page; its API may change or go.

The plan for Ryugo's area layout: a window filled edge to edge with areas that split, join, resize, hold tabs, and tear off into windows of their own, as Blender's do, with no data ever lost. This file is the plan and the to-do list; tick items off as they land, and keep the rules of [AGENTS.md](AGENTS.md) (motion, verification, no color on press).

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

- **Ryugo's buttons** (after the user asked): the chosen tab sits on a raised pill (`paintRaised`, `ShadowRaised`, the button's pill shape), which slides to a new choice on `springPill` and gives under a press (`springSqueeze`), as Segmented's; the others are Muted, Text under the pointer, with no fill. The header's buttons (the kind's menu, maximize, a window's close) are Button's raised pill, small. This replaces the underline below.
- **The header** is now the kind's icon and a chevron (its menu, as Blender's editor type; `DockKind.Icon`, a generic one when nil), the tabs, and the maximize button. Every pane is a tab, one alone too, so that it can be dragged. The tab chosen is in Text (Text2 when its area isn't focused), the others Muted; with more than one, a 2px underline slides to the choice (`springPill`), as Tabs' does. The close button shows on the tab chosen and the one under the pointer. Ctrl+PageUp and Ctrl+PageDown choose the tab before and after.
- **Picking up:** a tab dragged 4px lifts off as a raised tile (`springLift`, `ShadowPop`, leaning with its speed); its place stays empty, and the tabs after it slide over to close the room (`springShift`).
- **Where it goes:** over an area's tabs, they slide apart where it would go (a tab in that stack); over the rest of an area, the side the pointer is nearest within 30% of it docks it beside the area (left, right, top, bottom; only when the area is large enough for two), else its middle puts it at the end of the stack. Over its own area, alone there, it goes nowhere. The region is a wash of the accent with a hairline of it (`ColorAccent500`, 10% and 55%; 16% and 70% in the dark theme), inset 3px with a concentric radius, sliding between zones and areas on `springPop`, fading in 140 ms.
- **Let go:** the move applies as the next frame starts (as a drop does), every stack building the new order together; the tab flies from where its copy was to its place once a frame laid it out (`springLand`), and the underline slides to it. Docked beside an area, the new area opens from nothing on `springShift`, its neighbor making room, and what it shows fades in. Nowhere, the tab flies back.
- **The compass** (after the user tried it, from Visual Studio's docking guides, in Ryugo's language; `dock_compass.go`): over the area under the pointer, a plus of five raised targets (30px, the control's surface) on a plate (the popover's color, `ShadowPop`, its corners concentric with the targets'), in the middle of the area's body: the middle puts the tab in the stack, the arms dock it beside the area. Each target draws what it does: a small pane, the part it takes filled in the accent. A target that can't apply (the middle over the tab's own area, an arm where the area is too small to split) shows faded. The compass grows in on `springPop` (from 92%), glides from area to area on it, and goes in 140 ms; it shows only over an area large enough for it, and keeps clear of the window's edge targets. Under the pointer a target grows 8% (`springSqueeze`) and fills in, the region shows behind, and the tab's copy lets it show through (55%). The compass is built above the copy.
- **The window's edges:** four targets in the middle of the window's edges (the top one under the headers, so that it hides no tab) dock the tab along the whole of that edge, a quarter of the window long (`dockPane` into the root). They fade in with the drag.
- Targets win over the nearness of an area's side, which stays as the quick way.
- **Closing:** the pane goes at once (Undo brings it back); where its tab was, a copy narrows to nothing as it fades (200 ms) and the tabs after it slide over.
- **An area left empty** folds away as a join does, without the darkening, its neighbor before it (else after it) taking its room. A window's last area stays, saying "Drag a tab here".
- **The model:** `movePane`, `dockPane`, `closePane`, `removeArea`; every change by the user is one step of Undo.

### 5. Windows and saving

- [x] Tear-off of a pane or a tab into its own window (from step 0), and docking back from it.
- [x] Moving a whole area to a new window: from a tab's menu ("Move area to new window"); a Shift-drag from a corner is not done.
- [x] Closing a torn-off window docks its panes back into the main window.
- [x] Saving the layout on every change; restoring windows, positions and panes on open; a window placed off every display comes back onto one.

**Result (2026-11-04, `dock_travel.go`, `dockwin/`):** measured in the tester with two Docks of one layout side by side, a fake host making them two windows; not tried in real windows yet.

- **The core stays free of windows.** The Dock knows windows through `DockHost`: `Content(window)` (where its content is on the screen, and whether the system says) and `Sync()` (open the app's windows for the layout's, close those it no longer holds, redraw them all). `ryu-go/dockwin` is the host on MyGo's windows, a package of its own so that an app using a Dock without windows doesn't take in MyGo's window package. The layout's windows are the truth: every Dock notices when they change (a tear-off, an Undo, a layout restored) and calls `Sync`.
- **A tab dragged out of its window** (its pointer outside the window's content, read from the window's own coordinates, so it works everywhere) is no longer the window's to show: its edge targets fade, and the layout's `travel` says where it is on the screen and over which window. That window's Dock shows it as its own tab dragged: the copy, the stack making room, the region, the compass; and says what letting it go there does. Let go there, the move applies as the source's next frame starts, and the tab lands in the other window from where its copy was.
- **Let go over no window,** the tab tears off into a new window where its copy was, the size of the area it left (320×240 to 900×700). A window left without panes closes (`pruneWindows`).
- **Closing a window** gives its tree back to the main window, along its right edge, a quarter of it, opening from nothing (`Layout.DockBack`, which `dockwin` calls from `OnClose`).
- **Places:** `dockwin` writes a window's content bounds into its `DockWindow` as it moves and resizes, and calls its `OnChange`; opening, a window whose title bar would be on no display's work area opens where the system puts it.
- **Wayland** says nowhere where windows are: there a tab dragged out only tears off (the compositor places the new window), and the tab's menu has "Move to main window". With `GDK_BACKEND=x11` GTK runs on XWayland, and all of it works.
- **No title bar** (after the user tried it): windows torn off open with MyGo's hidden title bar and no window buttons (`TitleBarHidden`, `HideWindowButtons`), which keeps their shadow, corners and resizing edges where the system draws them. Linux rounds no window's corners and outlined them with a black band (the content's padding on the dark background): there the window is transparent, and `dockwin` draws its surface, rounded as the Dock's areas are, with a hairline and a soft shadow in a 16px margin whose pixels resize the window; the layout keeps the surface's place and size, without the margin. The Dock's areas fill a torn-off window to its edges, their corners its own. The room after an area's tabs moves the window (`DragWindow`), and the top right area's header ends with the window's close button. A window's only tab carries the window: it follows the pointer (`DockHost.Pointer`, `Move`), fades to 55% over another window of the layout, which shows where it would dock, and let go there, docks into it as the window closes. Where the system doesn't say where the pointer is (Wayland), that tab moves the window by the system, as a title bar does.
- **Redrawing** (after the user tried it): a window is drawn only when its own input or animations ask, so a theme changed in the main window left a torn-off one in the old theme until the pointer went over it. A Dock now redraws the layout's other windows (`DockHost.Sync`) when its layout changes and when its theme does; `dockwin.Host.Redraw` is there for the app's own data, and the gallery calls it when its theme changes on any page.
- **Merging windows** where no drag can (Wayland): a tab's menu lists the other windows torn off ("Move to Properties window"), and moves the tab into the target's top right area; the window left empty closes. - **Dragging on Wayland** (`dock_native.go`): where the system says nowhere where windows or the pointer are, tabs are dragged by the system's drag and drop (MyGo's `DragData` with a typed value, in a format of our own that other apps don't take). The copy is the system's (a picture of the tab), the tab's place stays empty; the Dock under the pointer, in whichever window, takes the value (`DragOver`, `Drop` on its root) and shows it as its own tab dragged: its stack making room, the region, the compass. Let go, it docks there and lands from where its copy was; let go over no window (not given up with Escape), it tears off into a window of its own, which the system places. A window torn off moves by the room after its tabs. Not testable in the tester (it has no way to drag the system's way); to try in real windows.
- **The gallery** (`cmd/gallery`) saves the workspace's layout in the user's config directory (`ryugo-gallery/layout.json`) on every change, and opens it as it was, windows too.

### 6. Gallery and docs

- [x] The Workspace page showing all of it (files, editors, a console, an outline and properties as kinds; the table and charts as kinds wait for the charts).
- [x] README: a section on Dock, with a screenshot.
- [x] Interaction tests for split, join, resize, docking, maximize, and the save and restore round trip.

**Also in the core (2026-11-04):** Redo (`Layout.Redo`, `CanRedo`; a new change clears what Undo went back from); a tab's menu on a right click (split right or down, open in a new window, move the area to a new window, move to the main window, close, close the others); a double click on a tab maximizes its area; a layout put in place of another (restored, reset) starts afresh, every node where it is, instead of gliding from the springs of nodes that had the same IDs.

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
