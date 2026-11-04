package mizan

import (
	"encoding/json"
	"fmt"

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
	}
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
