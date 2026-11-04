package ryugo

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/egoist/mygo/ui"
)

// The model of a Dock (see DOCK.md): plain data the app owns, saves and
// restores. A Layout is a set of windows, each holding a tree whose nodes
// are splits (a row or a column of children, with their sizes) or areas (a
// stack of panes). A pane is a view of the app's data, never its holder: it
// keeps only what it shows (its scroll, the app's own view state), so it
// can go anywhere without losing anything.

// Layout is the arrangement of a Dock: its windows, the main one first.
type Layout struct {
	Windows []*DockWindow `json:"windows"`
	// next numbers the nodes and panes the layout makes.
	next int
	// history holds the layouts before the last changes, for Undo, and
	// future those Undo went back from, for Redo.
	history, future [][]byte

	// What the Docks of a layout's windows share, not saved: a tab dragged
	// from one window over another, and the areas a change made, which the
	// Dock that shows them opens from nothing (opening) or fades in (born).
	travel  *dockTravel
	opening map[string]bool
	born    map[string]bool
}

// dockTravel is a tab dragged out of its window: where the pointer is on
// the screen and over which window, and what letting go there does, which
// that window's Dock says.
type dockTravel struct {
	pane, from string // the pane, and the window it is dragged from
	title      string
	w, h       float32
	grab       ui.Point
	over       string   // the window under the pointer; none when outside every one
	at         ui.Point // the pointer in that window's content
	op         func()   // letting go there; nil when it takes nothing
}

// mark notes that the Docks show area id as new: opening from nothing, as
// its neighbors make room, and fading in what it shows.
func (l *Layout) mark(id string, open bool) {
	if l.born == nil {
		l.born, l.opening = map[string]bool{}, map[string]bool{}
	}
	l.born[id] = true
	if open {
		l.opening[id] = true
	}
}

// DockWindow is a window of a Layout: the main one ("main"), or one a pane
// was torn off into, with where it is on the screen.
type DockWindow struct {
	ID     string `json:"id"`
	X      int    `json:"x,omitempty"`
	Y      int    `json:"y,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
	Root   *Node  `json:"root"`
	// Max is the area maximized over the window, if one is.
	Max string `json:"max,omitempty"`
}

// DockDir is the direction a split lays its children out in.
type DockDir uint8

const (
	// DockRow lays its children side by side.
	DockRow DockDir = iota
	// DockColumn stacks them.
	DockColumn
)

func (d DockDir) MarshalText() ([]byte, error) {
	if d == DockColumn {
		return []byte("column"), nil
	}
	return []byte("row"), nil
}

func (d *DockDir) UnmarshalText(b []byte) error {
	switch string(b) {
	case "row", "":
		*d = DockRow
	case "column":
		*d = DockColumn
	default:
		return fmt.Errorf("ryugo: unknown dock direction %q", b)
	}
	return nil
}

// Node is a node of a window's tree: a split, with Children and their
// Sizes, fractions of the room adding up to 1; or an area, with Panes and
// the one shown.
type Node struct {
	ID       string    `json:"id"`
	Dir      DockDir   `json:"dir,omitempty"`
	Children []*Node   `json:"children,omitempty"`
	Sizes    []float32 `json:"sizes,omitempty"`
	Panes    []*Pane   `json:"panes,omitempty"`
	Active   int       `json:"active,omitempty"`
}

// IsArea reports whether n is an area rather than a split.
func (n *Node) IsArea() bool { return len(n.Children) == 0 }

// Pane is a pane of an area: what it shows (its Kind, which the app builds)
// and how it shows it. Scroll is where the pane is scrolled, for a Kind
// that scrolls (TrackScroll); State holds the app's own view state of the
// pane. Both travel with the pane and are saved with the layout.
type Pane struct {
	ID     string            `json:"id"`
	Kind   string            `json:"kind"`
	Title  string            `json:"title"`
	Scroll ui.ScrollState    `json:"-"`
	State  map[string]string `json:"state,omitempty"`
}

// paneJSON is how a Pane is saved: its scroll as two numbers.
type paneJSON struct {
	ID      string            `json:"id"`
	Kind    string            `json:"kind"`
	Title   string            `json:"title"`
	ScrollX float32           `json:"scrollX,omitempty"`
	ScrollY float32           `json:"scrollY,omitempty"`
	State   map[string]string `json:"state,omitempty"`
}

func (p *Pane) MarshalJSON() ([]byte, error) {
	return json.Marshal(paneJSON{p.ID, p.Kind, p.Title, p.Scroll.X, p.Scroll.Y, p.State})
}

func (p *Pane) UnmarshalJSON(b []byte) error {
	var j paneJSON
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}
	*p = Pane{ID: j.ID, Kind: j.Kind, Title: j.Title, State: j.State}
	p.Scroll.X, p.Scroll.Y = j.ScrollX, j.ScrollY
	return nil
}

// NewLayout returns a layout of one main window holding root.
func NewLayout(root *Node) *Layout {
	l := &Layout{Windows: []*DockWindow{{ID: "main", Root: root}}}
	l.fix()
	return l
}

// DockSplit returns a split of children in direction dir; sizes are their
// fractions of the room (equal when nil).
func DockSplit(dir DockDir, sizes []float32, children ...*Node) *Node {
	return &Node{Dir: dir, Sizes: sizes, Children: children}
}

// DockArea returns an area holding panes, the first shown.
func DockArea(panes ...*Pane) *Node { return &Node{Panes: panes} }

// ParseLayout reads a layout saved as JSON (encoding/json of a Layout). It
// mends what it can rather than fail: sizes that don't match the children
// are shared out again, missing IDs are made, and an empty split goes.
func ParseLayout(b []byte) (*Layout, error) {
	var l Layout
	if err := json.Unmarshal(b, &l); err != nil {
		return nil, err
	}
	if len(l.Windows) == 0 {
		return nil, fmt.Errorf("ryugo: a layout needs a window")
	}
	l.fix()
	return &l, nil
}

// MarshalJSON saves the layout, for ParseLayout to read back.
func (l *Layout) MarshalJSON() ([]byte, error) {
	type plain Layout
	return json.Marshal((*plain)(l))
}

// Window returns the window id, nil when there is none.
func (l *Layout) Window(id string) *DockWindow {
	for _, w := range l.Windows {
		if w.ID == id {
			return w
		}
	}
	return nil
}

// Find returns the node id in any window, nil when there is none.
func (l *Layout) Find(id string) *Node {
	var found *Node
	for _, w := range l.Windows {
		walk(w.Root, func(n *Node) {
			if n.ID == id {
				found = n
			}
		})
	}
	return found
}

func walk(n *Node, fn func(*Node)) {
	if n == nil {
		return
	}
	fn(n)
	for _, ch := range n.Children {
		walk(ch, fn)
	}
}

// fix gives every node and pane an ID, and every split sizes that match its
// children and add up to 1; it keeps the numbering past the IDs it finds.
func (l *Layout) fix() {
	seen := map[string]bool{}
	for _, w := range l.Windows {
		walk(w.Root, func(n *Node) {
			seen[n.ID] = true
			for _, p := range n.Panes {
				seen[p.ID] = true
			}
		})
	}
	id := func() string {
		for {
			l.next++
			if s := fmt.Sprintf("n%d", l.next); !seen[s] {
				seen[s] = true
				return s
			}
		}
	}
	for _, w := range l.Windows {
		walk(w.Root, func(n *Node) {
			if n.ID == "" {
				n.ID = id()
			}
			for _, p := range n.Panes {
				if p.ID == "" {
					p.ID = id()
				}
			}
			if n.IsArea() {
				n.Sizes = nil
				n.Active = min(max(n.Active, 0), max(len(n.Panes)-1, 0))
				return
			}
			normalize(n)
		})
		w.Root = tidy(w.Root)
	}
}

// tidy folds away what a split or a join can leave: a split of one child
// becomes the child, and a split inside a split of the same direction gives
// its children to it, so that two areas sharing a whole edge are always
// siblings. It returns what takes n's place.
func tidy(n *Node) *Node {
	if n == nil || n.IsArea() {
		return n
	}
	var kids []*Node
	var sizes []float32
	for k, ch := range n.Children {
		ch = tidy(ch)
		if !ch.IsArea() && ch.Dir == n.Dir {
			for j, g := range ch.Children {
				kids = append(kids, g)
				sizes = append(sizes, n.Sizes[k]*ch.Sizes[j])
			}
			continue
		}
		kids = append(kids, ch)
		sizes = append(sizes, n.Sizes[k])
	}
	n.Children, n.Sizes = kids, sizes
	if len(n.Children) == 1 {
		return n.Children[0]
	}
	return n
}

// normalize makes n's sizes match its children and add up to 1.
func normalize(n *Node) {
	var sum float32
	ok := len(n.Sizes) == len(n.Children)
	for _, s := range n.Sizes {
		if s <= 0 {
			ok = false
		}
		sum += s
	}
	if !ok || sum <= 0 {
		n.Sizes = make([]float32, len(n.Children))
		for i := range n.Sizes {
			n.Sizes[i] = 1 / float32(len(n.Children))
		}
		return
	}
	for i := range n.Sizes {
		n.Sizes[i] /= sum
	}
}

// where returns the window holding node id, its parent (nil for a window's
// root) and its place there.
func (l *Layout) where(id string) (*DockWindow, *Node, int) {
	for _, w := range l.Windows {
		if w.Root != nil && w.Root.ID == id {
			return w, nil, 0
		}
		var parent *Node
		at := 0
		walk(w.Root, func(n *Node) {
			for i, ch := range n.Children {
				if ch.ID == id {
					parent, at = n, i
				}
			}
		})
		if parent != nil {
			return w, parent, at
		}
	}
	return nil, nil, 0
}

// replace puts by in old's place, in its parent or as its window's root.
func (l *Layout) replace(old, by *Node) {
	w, p, i := l.where(old.ID)
	switch {
	case p != nil:
		p.Children[i] = by
	case w != nil:
		w.Root = by
	}
}

// splitArea splits area id in direction dir into itself and a new area
// showing a copy of its pane, keep and made DIPs long; the new one comes
// first when first. room is the length the parent shares out, for a parent
// of the same direction, which takes the new area in. It returns the new
// area.
func (l *Layout) splitArea(id string, dir DockDir, first bool, keep, made, room float32) *Node {
	a := l.Find(id)
	if a == nil || !a.IsArea() {
		return nil
	}
	l.remember()
	var q *Pane
	if len(a.Panes) > 0 {
		p := a.Panes[a.Active]
		q = &Pane{Kind: p.Kind, Title: p.Title, State: maps.Clone(p.State)}
		// the scroll's place, not the state itself: it is bound to one view
		q.Scroll.X, q.Scroll.Y = p.Scroll.X, p.Scroll.Y
	}
	return l.beside(a, q, dir, first, keep, made, room)
}

// beside puts a new area holding pane (none when nil) beside area a, as
// splitArea does, and returns it.
func (l *Layout) beside(a *Node, pane *Pane, dir DockDir, first bool, keep, made, room float32) *Node {
	n := &Node{}
	if pane != nil {
		n.Panes = []*Pane{pane}
	}
	pair, sizes := []*Node{a, n}, []float32{keep, made}
	if first {
		pair, sizes = []*Node{n, a}, []float32{made, keep}
	}
	if _, p, i := l.where(a.ID); p != nil && p.Dir == dir && room > 0 {
		// the parent's other children keep their lengths: the new edge
		// comes out of the area split
		var kids []*Node
		var px []float32
		for k, ch := range p.Children {
			if k == i {
				kids, px = append(kids, pair...), append(px, sizes...)
				continue
			}
			kids, px = append(kids, ch), append(px, p.Sizes[k]*room)
		}
		p.Children, p.Sizes = kids, px
	} else {
		l.replace(a, &Node{Dir: dir, Children: pair, Sizes: sizes})
	}
	l.fix()
	return n
}

// paneAt returns the area holding pane id, and its place in the stack.
func (l *Layout) paneAt(id string) (*Node, int) {
	var area *Node
	at := -1
	for _, w := range l.Windows {
		walk(w.Root, func(n *Node) {
			for i, p := range n.Panes {
				if p.ID == id {
					area, at = n, i
				}
			}
		})
	}
	return area, at
}

// takePane takes pane id out of its area, which shows its neighbor in the
// stack if it showed it, and returns the pane and the area.
func (l *Layout) takePane(id string) (*Pane, *Node) {
	a, i := l.paneAt(id)
	if a == nil {
		return nil, nil
	}
	p := a.Panes[i]
	a.Panes = slices.Delete(a.Panes, i, i+1)
	if a.Active > i || a.Active >= len(a.Panes) {
		a.Active--
	}
	a.Active = max(a.Active, 0)
	return p, a
}

// movePane moves pane id into area to's stack at index (among the panes
// other than it; at the end when out of range), showing it there.
func (l *Layout) movePane(id, to string, index int) bool {
	t := l.Find(to)
	if t == nil || !t.IsArea() {
		return false
	}
	p, _ := l.takePane(id)
	if p == nil {
		return false
	}
	index = min(max(index, 0), len(t.Panes))
	t.Panes = slices.Insert(t.Panes, index, p)
	t.Active = index
	return true
}

// dockPane moves pane id into a new area beside node to (an area, or a
// window's root to dock along its edge), as splitArea makes one, and
// returns it.
func (l *Layout) dockPane(id, to string, dir DockDir, first bool, keep, made, room float32) *Node {
	t := l.Find(to)
	if t == nil {
		return nil
	}
	p, _ := l.takePane(id)
	if p == nil {
		return nil
	}
	return l.beside(t, p, dir, first, keep, made, room)
}

// closePane closes pane id, for Undo to bring back.
func (l *Layout) closePane(id string) bool {
	if a, _ := l.paneAt(id); a == nil {
		return false
	}
	l.remember()
	l.takePane(id)
	return true
}

// removeArea takes area id, left empty, out of its split: the neighbor
// before it (else after it) takes its room. A window's root stays.
func (l *Layout) removeArea(id string) bool {
	_, p, i := l.where(id)
	if p == nil {
		return false
	}
	k := i - 1
	if k < 0 {
		k = i + 1
	}
	p.Sizes[k] += p.Sizes[i]
	p.Children = slices.Delete(p.Children, i, i+1)
	p.Sizes = slices.Delete(p.Sizes, i, i+1)
	l.fix()
	return true
}

// joinArea joins area gone into its neighbor keep: keep takes its room and
// gone goes, with its panes (Undo brings them back).
func (l *Layout) joinArea(keep, gone string) bool {
	if !l.joinable(keep, gone) {
		return false
	}
	l.remember()
	_, p, i := l.where(gone)
	_, _, k := l.where(keep)
	p.Sizes[k] += p.Sizes[i]
	p.Children = slices.Delete(p.Children, i, i+1)
	p.Sizes = slices.Delete(p.Sizes, i, i+1)
	l.fix()
	return true
}

// joinable reports whether areas a and b sit side by side in one split,
// sharing a whole edge.
func (l *Layout) joinable(a, b string) bool {
	_, pa, i := l.where(a)
	_, pb, j := l.where(b)
	na, nb := l.Find(a), l.Find(b)
	return pa != nil && pa == pb && (i-j == 1 || j-i == 1) && na.IsArea() && nb.IsArea()
}

// undoDepth is how many changes Undo goes back.
const undoDepth = 16

func (l *Layout) remember() {
	b, err := l.MarshalJSON()
	if err != nil {
		return
	}
	l.history = append(l.history, b)
	if len(l.history) > undoDepth {
		l.history = l.history[1:]
	}
	// a new change: what was undone can't be redone over it
	l.future = nil
}

// CanUndo reports whether there is a change to undo: a split, a join, a
// pane docked, moved, torn off or closed, a kind chosen.
func (l *Layout) CanUndo() bool { return len(l.history) > 0 }

// CanRedo reports whether there is a change Undo went back from.
func (l *Layout) CanRedo() bool { return len(l.future) > 0 }

// Undo takes the layout back to before its last change (see CanUndo), and
// reports whether there was one. The panes come back with the scroll and
// state they had then; windows torn off since close, and those closed
// open again, when a Dock with a DockHost shows the layout.
//
//	if layout.Undo() {
//		app.saveLayout()
//	}
func (l *Layout) Undo() bool { return l.step(&l.history, &l.future) }

// Redo makes again the last change Undo went back from, and reports
// whether there was one.
func (l *Layout) Redo() bool { return l.step(&l.future, &l.history) }

// step takes the layout to the last of from, keeping what it was in to.
func (l *Layout) step(from, to *[][]byte) bool {
	now, err := l.MarshalJSON()
	if err != nil {
		return false
	}
	for len(*from) > 0 {
		b := (*from)[len(*from)-1]
		*from = (*from)[:len(*from)-1]
		back, err := ParseLayout(b)
		if err != nil {
			continue
		}
		*to = append(*to, now)
		// IDs made from here on are new, never those of the nodes undone;
		// what the Docks share stays
		next, history, future, travel, opening, born := l.next, l.history, l.future, l.travel, l.opening, l.born
		*l = *back
		l.next = max(l.next, next)
		l.history, l.future, l.travel, l.opening, l.born = history, future, travel, opening, born
		return true
	}
	return false
}

// newWindowID returns an ID no window of l has.
func (l *Layout) newWindowID() string {
	for i := len(l.Windows); ; i++ {
		if id := fmt.Sprintf("w%d", i); l.Window(id) == nil {
			return id
		}
	}
}

// tearOff moves pane id out into a new window, its content at r on the
// screen (r's place zero: where the system puts it), and returns it.
func (l *Layout) tearOff(id string, r ui.Rect) *DockWindow {
	p, _ := l.takePane(id)
	if p == nil {
		return nil
	}
	w := &DockWindow{ID: l.newWindowID(), X: int(r.X), Y: int(r.Y), Width: int(r.W), Height: int(r.H), Root: &Node{Panes: []*Pane{p}}}
	l.Windows = append(l.Windows, w)
	l.fix()
	return w
}

// tearOffArea moves area id, all its panes, out into a new window, as
// tearOff does.
func (l *Layout) tearOffArea(id string, r ui.Rect) *DockWindow {
	a := l.Find(id)
	if a == nil || !a.IsArea() || len(a.Panes) == 0 {
		return nil
	}
	panes, active := a.Panes, a.Active
	a.Panes, a.Active = nil, 0
	w := &DockWindow{ID: l.newWindowID(), X: int(r.X), Y: int(r.Y), Width: int(r.W), Height: int(r.H), Root: &Node{Panes: panes, Active: active}}
	l.Windows = append(l.Windows, w)
	l.fix()
	return w
}

// DockBack moves what window id holds into the main window, along its
// right edge, a quarter of it, and drops the window: what a window torn
// off holds comes back when it closes. It reports whether there was such
// a window.
//
//	win.OnClose(func(*mygo.CloseEvent) { layout.DockBack(id) })
func (l *Layout) DockBack(id string) bool {
	w, main := l.Window(id), l.Window("main")
	if w == nil || main == nil || w == main {
		return false
	}
	l.Windows = slices.DeleteFunc(l.Windows, func(o *DockWindow) bool { return o == w })
	if w.Root != nil && countPanes(w.Root) > 0 {
		l.graft(w.Root)
	}
	l.fix()
	return true
}

// moveToMain moves pane id into a new area along the main window's right
// edge, and returns it.
func (l *Layout) moveToMain(id string) *Node {
	if l.Window("main") == nil {
		return nil
	}
	p, _ := l.takePane(id)
	if p == nil {
		return nil
	}
	n := &Node{Panes: []*Pane{p}}
	l.graft(n)
	l.fix()
	return n
}

// graft puts n along the main window's right edge, a quarter of it, or in
// its place when the window holds nothing; the Docks open it from nothing.
func (l *Layout) graft(n *Node) {
	main := l.Window("main")
	if main.Root == nil || main.Root.IsArea() && len(main.Root.Panes) == 0 {
		main.Root = n
	} else {
		main.Root = &Node{Dir: DockRow, Children: []*Node{main.Root, n}, Sizes: []float32{0.75, 0.25}}
	}
	l.fix()
	walk(n, func(m *Node) {
		if m.IsArea() {
			l.mark(m.ID, false)
		}
	})
	l.mark(n.ID, true)
}

// pruneWindows drops the windows other than the main one left without a
// pane, and reports whether it dropped any.
func (l *Layout) pruneWindows() bool {
	n := len(l.Windows)
	l.Windows = slices.DeleteFunc(l.Windows, func(w *DockWindow) bool {
		return w.ID != "main" && (w.Root == nil || countPanes(w.Root) == 0)
	})
	return len(l.Windows) != n
}

func countPanes(n *Node) int {
	c := 0
	walk(n, func(m *Node) { c += len(m.Panes) })
	return c
}
