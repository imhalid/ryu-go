package mizan

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
	// history holds the layouts before the last splits and joins, for Undo.
	history [][]byte
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
		return fmt.Errorf("mizan: unknown dock direction %q", b)
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
		return nil, fmt.Errorf("mizan: a layout needs a window")
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
	n := &Node{}
	if len(a.Panes) > 0 {
		p := a.Panes[a.Active]
		q := &Pane{Kind: p.Kind, Title: p.Title, State: maps.Clone(p.State)}
		// the scroll's place, not the state itself: it is bound to one view
		q.Scroll.X, q.Scroll.Y = p.Scroll.X, p.Scroll.Y
		n.Panes = []*Pane{q}
	}
	pair, sizes := []*Node{a, n}, []float32{keep, made}
	if first {
		pair, sizes = []*Node{n, a}, []float32{made, keep}
	}
	if _, p, i := l.where(id); p != nil && p.Dir == dir && room > 0 {
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

// undoDepth is how many splits and joins Undo goes back.
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
}

// CanUndo reports whether there is a split or a join to undo.
func (l *Layout) CanUndo() bool { return len(l.history) > 0 }

// Undo takes the layout back to before its last split or join, and reports
// whether there was one. The panes come back with the scroll and state they
// had then.
//
//	if layout.Undo() {
//		app.saveLayout()
//	}
func (l *Layout) Undo() bool {
	for len(l.history) > 0 {
		b := l.history[len(l.history)-1]
		h := l.history[:len(l.history)-1]
		back, err := ParseLayout(b)
		if err != nil {
			l.history = h
			continue
		}
		next := l.next
		*l = *back
		// IDs made from here on are new, never those of the nodes undone
		l.history, l.next = h, max(l.next, next)
		return true
	}
	return false
}
