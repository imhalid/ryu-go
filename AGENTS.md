# ryu-go: rules for agents

Rules for every agent (and person) working in this repository. It is a design system for MyGo's native UI (package `ui`): behavior (focus, keyboard, accessibility) comes from MyGo's unstyled `…Base` components, and Ryugo gives the look. When a rule has to be broken, the reason goes in a code comment and this file is updated.

## 1. Structure and writing

- Each family of components is one file at the root: `controls.go`, `forms.go`, `overlays.go`, `dnd.go`, `feedback.go`, `disclosure.go`, `toggle.go`, `table.go`, `palette.go`… Colors, shadows and sizes live in `tokens.go`, written in OKLCH.
- There is one skin: Ryugo. Bevel was removed; don't bring it back.
- **Concentric corners.** A nested surface takes its radius from the one around it: `Concentric(outer, padding) = max(outer − padding, 0)`. The two radii are never scaled apart, and the same radius is never used inside and out: the gap at the corners would look thicker than along the sides. Where it applies: Segmented's track and pill (3px), ToggleGroup and its toggles (3px), NumberInput and its stepper (4px), the panel of a menu, Select or Popover and its items (`PadPop`), TagInput and its chips (4px); a board column goes out from its cards (card + 10px). It does not apply to containers with a large padding (Card).
- Radii are never written as constants: `t.R(RadiusX)` (the scope's `Corners`: Sm 0.5, Md 1, Lg 1.5). Pills (`RadiusPill`) and circles don't scale; what must stay square, as a checkbox, gets a cap (`min(t.R(6), 7)`).
- The surface has two languages: the raised control (`t.raised`, lit from above) and the inset well (`t.well`). A component doesn't write its own gradient; it uses these two.
- Code comments are in English, tell the decision, and match the density of the code around them. Everything exported has a doc comment with a usage example.
- The package is an early beta: the API changes freely, but when it does, the README and the gallery are updated in the same turn.
- Every new component shows in the gallery (`gallery/`) and gets an interaction test in `gallery/*_test.go`.
- The gate: `gofmt -l .` empty, `go vet ./...`, `go test ./...`, and a build without cgo for `linux`, `darwin`, `windows` and `js/wasm` (the same as CI).

## 2. Motion and micro-interactions

Small feedback and animation matter a great deal to the user. An interaction isn't done when it works; it has to feel right.

### Principles

- **Every change of state comes with motion.** Nothing that opens, closes, moves, is added or is removed appears or vanishes in one frame. Closing is animated too (the `presence` pattern, below).
- **Springs in, curves out.** Arrivals on a spring (`ui.SpringOptions`) with a little bounce; departures on a shorter curve (`EaseOut`, 150–220 ms) with none. Leaving is quicker than arriving.
- **Bounce in measure.** `Bounce` 0.04–0.16 for things that move; 0 for things that show a value (a Progress filling, cards making room): a bar past its value reads as more than there is.
- **No wind-up.** An element never goes back before it goes toward its target. When position and width change together, they ride one spring (the Switch knob: `x` and `w` on the same spring).
- **No jumps.** When a value is written two ways (a layout change and a transform), their sum must stay continuous frame to frame. In the frame that takes the new layout, take the transform back by the same amount (`Jump`), then let it run to zero (FLIP).
- **A physical feel.** What is held lifts (grows a little, its shadow becomes `ShadowPop`), leans with its speed (`DragVelocity`, ±2–2.5°), and glides into place when let go. What is around it makes room; a blue line is not enough.
- **Feedback on press.** A pressed button goes down 1px or shrinks to 96–98%; Segmented's pill squeezes in by 2px a side and springs open on release (`springSqueeze`). A press changes no color (no fill, no darkening): the feedback is motion alone. A Button pressed sends a wave through itself from the pointer (`ripple.go`, after MorphRipple.metal): a front runs out at 240 DIPs/s and each point it reaches swings out and back along the line from the press, `sin(42·t)·e^(−7·t)`, 1.8px at most, label and outline with the surface; by keyboard it starts from the middle. A distortion shader (`Element.Effect`), set only while a wave runs; off in tests (`rippleDraws`), whose CPU renderer takes seconds a click. A held stepper button steps once, then repeats after 400 ms, faster and faster.
- **Alive but calm.** A target saying "I'll take this" can be something slowly alive rather than a flat color: soft blots drifting on `noise`, like the shadows of clouds (`cloud`). Very faint, very slow (a period of seconds), never distracting.
- **Marks are drawn.** A check doesn't appear; it flows from the short arm to the long one as a pen would (`checkMark`, `Path.Trim` 0→1, 260 ms), and draws itself back the way it came, quicker, fading with the fill. In lists that open (Select, Combobox, Menu) the chosen item's check draws ~90 ms after the list opens: choosing closes the list, so the change itself is seldom seen.
- **The radio dot opens.** Chosen, the fill comes first and the dot opens from its middle ~40 ms later, without a bounce (`dotMark`, 240 ms); left, it shrinks away in 90 ms. No springy "pop".
- **A finishing sheen.** When a Progress reaches 100%, once the fill settles (~380 ms) a soft white band crosses it once (`finish`, 1.1 s); a bar full from the start doesn't play it. Not on Meter: a full meter doesn't mean "done".
- **A toast lands with the border beam.** A conic band in a 1px ring along the border (a long band fading at both ends, a bright head at 66%; in the dark theme the head is the tone lifted 65% toward white), and under it the same band, wider and fainter, drawn *under* the glass as its glow, so only what spills past the edge shows. On arrival it runs as two lights: one from the tone's dot (the middle of the start edge) along the top to the right, the other from across it along the bottom to the left; each goes only half way round, slowly (3 s `EaseInOut`), and fades where the other started. Drawn with `StrokePathConicGradient`; a layered line trimmed along the border (`Trim`) was found "amateur", and a halo and a band crossing the glass "simple". An absolute child paints over the content: what has to sit under, as a glow, is drawn in the parent's own `Draw`.
- **A toast takes its time.** It arrives rising from below, growing from 96% (`toastIn` 700 ms, bounce 0.14); the stack makes room on a 560 ms spring; it leaves over 340 ms, sinking, shrinking and fading. Swiped, it follows the pointer; past a third of its width, or flung, it goes on to the right as it fades, then closes (MyGo's exit copy doesn't know the swipe, so this is by hand); short of that it springs back, and pulled left it gives like a rubber band. A `ToastLoading` toast shows a spinner; when the result comes with the same ID, the spinner fades, the dot opens, the words cross-fade, and the beam runs again.
- **Lights of completion are thin.** The light running along an edge when something completes (`rise`): a 1px core, a narrow glow, ~1.5 s, `EaseInOut`, one curve (don't bend the animation's curve a second time along the path).
- **No blur.** Dialogs and Sheets have no blurred halo; a `Scrim` fades in instead. Glass (`glass`) only in popovers, menus, toasts, and the palettes' HUD (`palette.go`: a darker glass, dark in both themes, as Pixelmator Classic's).
- **A conversation moves as one.** In `Chat`, a new message grows in from its bubble's tail corner (the right for yours, the left for theirs) as the list makes room (`Collapse`, `springLand`); the typing bubble folds away as the message from the one typing grows out of its corner; the group's tail corner rounds off and passes to the new last bubble, its avatar and time row with it; a delivery check draws itself, and read, a second one after it as both take the accent. The list stays at the latest while it is there (following a message as it grows, not jumping to where it ends); scrolled up, a "Latest" button springs in with the number not seen.
- **Small things move too.** CopyButton's icon gives way to a drawn check, and comes back 1.5 s later; a held Slider's knob follows the pointer between steps and settles onto its step on a spring, a label of its value rising over it; a tooltip (a raised surface, as a button's) follows the pointer along its element on `springFollow`, its middle never past the element's ends; the pointer come onto it, it fades at once (a tooltip never stands in the pointer's way); it lingers 350 ms after the pointer leaves (fading over 120 ms), and going from one element to the next, slides there, the last one holding until the next is placed, so a row of buttons never flickers; a field's `Shake` swings it sideways a few times, each smaller (420 ms), motion only; a pop-up list that scrolls fades at its edges where there is more.
- **Less motion.** MyGo's `AnimateTo` and `Keyframes` go straight to their targets under `ReduceMotion`; drawings that count time by hand (`p.Now()`) must honor it too.
- **Rejected:** a tooltip following the pointer up and down too, as a callout with a beak pointing at it, was tried and taken back: it follows along its element only, and looks raised as a button. A countdown line showing a toast's time left was found "classic" and removed; don't show countdowns. NumberFlow-style rolling digits in NumberInput were not wanted (removed); don't suggest "spinning counter" ideas. A tailed, haloed glow on the Spinner was tried and not liked; the spinner stays plain. A Dialog growing out of the button that opened it (and shrinking back into it) was tried and removed; dialogs rise in where they are.

### Table (`table.go`)

- MyGo's `ui.Table` is not used: its look is fixed, moving a column changes color on press, and its neighbors jump. Ours: the header, moving and resizing columns, and our own virtualization at a fixed row height (only the rows in view, with room above and below).
- When the sort changes, rows **that showed** slide to their new places by FLIP; rows coming from out of view fade in at their place (they don't fly in from afar). A new row fades in, a removed row fades where it was (a ghost), and a page turns sideways.
- Moving a column is moving a card: a copy of the header lifts, the column's cells go along with it, faded, the neighbors make room on a spring, the new order applies in the next frame, and every column glides from where it showed.
- The column sorted first has a thin accent line along the top of its header and a glow fading down under it; in the dark theme the line goes toward white. When the sort moves to another column, the light slides there on a spring, as Tabs' underline does.
- No resize edge at the end of the last column (the table's edge).
- Moving a row (`Reorder`) is moving a card too: a copy lifts off its grip and follows the pointer up and down, the rows of its group slide apart, and let go it glides from where the copy was. The app reorders its own slice, called once the frame is built; the frame after, the rows' shift channels start fresh (`rowGen`) and every row glides from where it showed. No grip while a header sorts: the two orders would fight.
- Groups keep the order they come in; the table never reorders groups or rows itself beyond a header's sort within each group. A folded group's rows leave as removed rows do.
- The selection bar springs in over the toolbar; the two never cross but hand over (one is mostly gone before the other is much there).

### Tuned springs (reuse them)

| Name | Where | Value |
| --- | --- | --- |
| `springPanel` | Dialog arriving | 300 ms, bounce 0.12 |
| `springSheet` | Sheet arriving | 380 ms, bounce 0.06 |
| `springPop` | Popover, Menu | 220 ms, bounce 0.1 |
| `springPill` | Segmented's pill, Tabs' underline | 280 ms, bounce 0.14 |
| `springSqueeze` | A pressed pill giving and springing back | 220 ms, bounce 0.25 |
| `springShift` | Cards making room | 320 ms, bounce 0.06 |
| `springLand` | A dropped card landing | 380 ms, bounce 0.16 |
| `springLift` | A card lifting | 260 ms, bounce 0.2 |
| `stackSpring` | The toast stack | 560 ms, bounce 0.08 |
| `toastIn` | A toast arriving | 700 ms, bounce 0.14 |
| `springBar` | A Progress filling | 500 ms, bounce 0 |
| `springSelBar` | The table's selection bar | 320 ms, bounce 0.1 |
| `springOpen` | A table row's details | 320 ms, bounce 0.04 |
| `springGuide` | The Dock's split line: growing, settling onto an edge | 200 ms, bounce 0.1 |
| `springFollow` | A tooltip following the pointer | 260 ms, bounce 0.04 |

A new motion first tries one of these; when it needs a new one, name it and add it here.

## 3. MyGo pitfalls (learned)

- **`Transition.Enter` only plays for an element whose parent the last frame showed.** A Dialog's panel is born in the same frame as its backdrop, so its entrance never plays, and closing, it vanishes with the backdrop at once. The fix is `overlays.go` › `presence`: the state lives on a "host" element that takes no room, the progress is driven by hand with `Jump(0)` + `AnimateTo(1)`, the layer keeps being built until its exit ends, and while it closes it lets the pointer through to the page (`PassThrough`, `IgnoreInput`).
- **`AnimateTo` starts at its target on the first call.** To animate something from 0, `Jump(key, 0)` in the first frame (once, with a `ui.Local` flag).
- **Two `Translate`s on one element: the last one wins.** Motion from two sources (a list's shift and a landing) goes on two elements: the place outside, the card inside (`dnd.go` › `dragSource`).
- **`LayoutBounds` is the last frame's layout.** An operation that reorders is measured in the frame *after* the one that builds the new order. A change of state (as a drop) applies when every list will build the new order in the same frame: at the start of the next frame, never in the frame it was let go in (`settleDrop`).
- **The last frame's `ui.Element` can't be used** ("element from pass N used in pass M"). To carry a measure across frames, keep the number (`dropList.heights`), not the element.
- **A drag can leave the window.** While pressed, `Dragged` and `PointerPosition` keep reporting outside the window; the screen point is `Window.ContentBounds()` + `LayoutBounds` + `PointerPosition`. Open windows with `mygo.NewWindow` from a goroutine, never inside a frame (DOCK.md, step 0).
- **A `ui.ScrollState` is bound to one view.** Copying the struct copies its handle, and two scroll views built with it panic ("Handle bound to two controls"). To give a copy the same place, set `X` and `Y` on a fresh one (`Layout.splitArea`).
- **We have our own drag layer.** `Element.Drag`'s copy and `ui.List` Reorder's blue line can't be styled; instead the pointer is followed with `Dragged`/`PointerPosition` and the copy drawn in `ui.Overlay`. The pointer's position is read as `LayoutBounds + PointerPosition`, not by adding up deltas.
- **`ScrollWith` doesn't move absolute children.** A pinned cell in the header needs no counter-translate; one in the body does.
- **`ButtonBase.Selected(true)` paints MyGo's theme accent fill.** Draw states such as the current page with our own surface, and put the selection in the label.
- **An absolute child paints over its siblings in the flow.** What has to stay under (a glow, glass) is drawn in the parent's own `Draw`.
- **A value driven by `Jump` builds up a speed.** A channel `Jump`ed with the pointer, then `Jump`ed to another value and let go with `AnimateTo`, counts the difference between the two jumps as a speed, and the spring flings the element. When a fresh start is needed, use a new channel key (the table: `colKey`, a counter that grows with every column move).
- **The frame between a drop and its applying must hold still too.** If the new layout applies in the next frame, values must not start sliding to their targets in the frame between; they hold where they are.
- **A clipped box's border sits under its children.** In a rounded corner the header's fill covers the border; the border is drawn over the content with `DrawOver`.
- **`slices.Delete` zeroes the slots it frees.** A handler that removes from a slice while a loop over it still builds (a tab's close button inside its stack's loop) leaves nil behind for the rest of the loop: build the frame from a copy (`slices.Clone`).
- **A press ends when its element isn't built.** MyGo forgets the pressed element (and so the drag) in the first frame that doesn't build it, and keys are scoped to the parent: a drag that changes the tree must be taken by an element whose place doesn't change (the Dock's edges take the pointer over the window, with a stand-in for the edge held when it went).
- **`Dragged` loses moves when the element moves with the pointer.** Read where the pointer is (`LayoutBounds + PointerPosition`) against where it pressed.
- **A file named `*_windows.go` (or `_linux`, `_darwin`, `_js`…) builds on that system only.** Go reads the suffix as a build constraint: name a file about windows `dock_travel.go`, not `dock_windows.go`.
- **Overlay elements share one parent.** What two components build in `ui.Overlay` under the same key collides (two Docks in one window): key it with what tells them apart (`dockView.over`).
- **The tester's pointer carries no modifiers.** `Press`, `Move` and `Release` send none, and each pointer event sets the held ones, so `HoldModifiers` before a press is lost: an Option-drag can't be tested in the tester.
- **`Context.After` answers once, for the next frame.** An animation's frame in between answers it: ask again each frame until the time comes (CopyButton).
- **An element belongs to the parent open as it is made.** Made before its parent's `Children`, it is that parent's sibling, not its child: make it inside the callback (a palette's body, folded in a clip).
- **`Tester.Find` returns the first element showing or labeled s.** Names repeat (an area's kind button and its tab, a file list and a tab): find an element by a label only it has ("Close " + the tab's title), or give the gallery names that don't repeat.
- **`PassThrough` is the element's alone.** Its children (its words too) still take the pointer, and hide what is under them from hover: what floats over the page without taking the pointer, as a tooltip, takes `IgnoreInput(true)`, which reaches its children.
- **Paint-only animation** is done in `Draw` with `p.Now()` and `p.AnimationFrame()` (it doesn't rebuild the view): the cloud glow, the spinner.

## 4. Verifying motion

Passing tests don't prove the feel is right.

- **Measure frame by frame.** Play the motion in `ui.Tester` with a temporary `gallery/zz_*_tmp_test.go`, log positions with `tt.Find(...)` each frame or save `tt.Image()` as PNG; delete the temporary file when done. Look for continuity: nothing may jump at the moment of a drop.
- Take screenshots in **light and dark**.
- **The tester is not the real app.** It may handle frames in another order or pass; some bugs only show in the real window (the drop timing was found this way). When the user sends a screen recording, extract frames with `ffmpeg` (`fps=30,crop=…,tile=6x5`) and find the moment in the frames.
- Don't report something as fixed that you couldn't try in the real window; say what you verified and what you couldn't.

## 5. Bring suggestions

The user shouldn't have to think of everything. While touching a component, or when a job is done, if you see a micro-interaction that would suit it, suggest it in a sentence or two ("we could also add: …"). Make it concrete (what, where, how it feels), two or three at most at a time, and get a yes before doing it.

Pending ideas (delete as they're done):

- Tabs: the content slides in from the side the tab changed toward; the underline stretches toward its target as it goes, and settles to its width.
- Sortable, the board and the table: the list scrolls by itself as a card nears its edge, slowly, faster the nearer it gets.
- Accordion: as a section opens, the ones under it slide down on a spring rather than jump.

## 6. To do

Delete items as they're done; add a gap or an idea here as it comes up.

### Components

- **Data:** charts (LineChart, BarChart, Sparkline; MyGo's `chart` package), Gauge, RingChart, BulletChart; Stat (a number, its change, a small chart): as the value changes, the sparkline draws itself (as `checkMark` does) and the change badge goes to its new tone, with no rolling digits; DescriptionList; Tree.
- **App skeleton:**
  - Header, the top bar of an app (title, navigation, actions): as the page scrolls it shrinks and turns to glass, and a thin groove appears under it.
  - Sidebar / NavList: the chosen item's pill slides on a spring; collapsed, it goes down to icons.
  - Breadcrumb, ContextMenu.
  - CommandPalette (⌘K): a glass panel that opens growing; as one types, the results slide to their places as the table's rows do, and the chosen row's pill follows the arrows on a spring.
  - SplitView: the handle's panes follow the pointer; let go near an edge, the pane springs shut; a double click gives back its size.
  - **Palettes** (experimental, `palette.go`, the gallery's Studio page): tool panels floating over a canvas as Pixelmator Classic's, in the window (Wayland places no windows). Given a `PaletteHost` (`palettewin`), they are windows of their own (borderless, transparent children of the app's window, drawing their shadow in a margin), carried by their title bar and settling onto the work area's, the app window's and each other's edges; on Wayland, where windows can't be placed, they float in the window. A palette's window redraws the others while the pointer is over it, and the app's window the palettes' while it is over it: no loop. Waiting: docking palettes together into tabbed groups, a palette's own resizing, the keyboard (moving the focus between palettes, moving one with the arrows), the macOS HUD material (`Vibrancy`) behind a palette's window.
  - **Dock** (experimental: what it is for isn't settled; keep it marked so), a Blender-style workspace: areas that split and join from their corners, shared edges, tabs that dock, and panes that tear off into windows of their own with no data lost. Its plan and step-by-step to-do list are in [DOCK.md](DOCK.md); read it before working on it, and tick items off there.
  - A dashboard page in the gallery built of these, with the table and charts: the first look at the library as a whole.
- **Special form controls:** OTPField, Rating, ColorPicker, DateRangePicker, TimePicker, Stepper (a wizard); HoldButton, a dangerous action held to confirm: while held, a light in the BorderBeam's language fills round its edge, let go it draws back, and done, the button confirms.
- **Other:** Timeline, Carousel, AvatarGroup, Callout / Alert, HoverCard; a family of empty states ("No results", "No access", "Offline") with line drawings that draw themselves as they show.

### Known gaps

- Board cards can't be moved with the keyboard (Sortable has Alt+Up/Down, the board doesn't).
- Table columns can't be moved with the keyboard.
- While dragging, a list or a table doesn't scroll by itself near its edge.
- The spinner is plain; the tailed, haloed glow was rejected, and no other direction was tried.

### Not tried in the real window

Verified frame by frame in the tester, not yet confirmed by the user in the real window:

- Table (moving columns, the pinned column, opening rows, page turns, the sorted header's light, moving rows by their grip, folding groups).
- The toast's two-light beam, swiping away, loading → result.
- `Corners` S/M/L and concentric corners.
- The Button's press wave (`ripple.go`): its feel on a real GPU, and whether the label shifts as the effect turns on and off.
- Dock: tabs (picking up, reordering, closing), docking with the region preview, the compass and the window's edge targets, an emptied area folding away, the split line settling with its light, an undone split folding away, an edge cut with Option, edges settling onto lined-up edges; tabs torn off into windows and docked from one window into another (only with a fake host in the tester), windows closed giving their panes back, the layout saved and opened with its windows; on Wayland, tabs dragged by the system between windows (not testable in the tester at all).
- Palettes: dragging, settling onto edges and neighbors with the light, folding, closing and opening, the HUD; as windows of their own (only with a fake host in the tester).
