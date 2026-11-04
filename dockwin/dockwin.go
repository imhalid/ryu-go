// Package dockwin shows the windows of a Dock's layout in MyGo's windows,
// experimental as Dock is:
// a ryugo.DockHost. A tab dragged out of a window tears off into a window
// of its own where it is let go, and docks into another window of the
// layout it is let go over; a window closed gives what it holds back to
// the main window; windows' places and sizes are kept in the layout, so
// that saved, they open again where they were. Windows torn off have no
// title bar: the Dock's headers move them (a window's only tab carries it,
// to dock it into another), and their top right area closes them.
//
// It is a package of its own so that apps that use a Dock without windows
// of its own don't take in MyGo's window package with it.
//
//	host := dockwin.New(win, func() *ryugo.Layout { return app.layout }, func(id string) func(*ui.Context) {
//		return func(c *ui.Context) {
//			ryugo.Root(c, ryugo.Options{})
//			ryugo.Dock(c, app.layout, ryugo.DockOptions{Window: id, Kinds: app.kinds, Host: host}).Fill()
//		}
//	})
//	host.Sync() // the windows saved with the layout open
package dockwin

import (
	"os"
	"runtime"
	"sync"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/imhalid/ryu-go"
)

// Host shows a layout's windows in MyGo's windows.
type Host struct {
	main   *mygo.Window
	layout func() *ryugo.Layout
	view   func(window string) func(*ui.Context)
	// OnChange runs when a window moved, was resized or closed: the layout
	// changed, the moment to save it.
	OnChange func()

	mu      sync.Mutex
	wins    map[string]*mygo.Window
	opening map[string]bool
}

// New returns the host of the layout layout returns (it can change, as a
// layout restored), whose main window is main; view builds the content of
// a window the layout tears off, by its ID: a Dock of that window.
func New(main *mygo.Window, layout func() *ryugo.Layout, view func(window string) func(*ui.Context)) *Host {
	return &Host{main: main, layout: layout, view: view,
		wins: map[string]*mygo.Window{"main": main}, opening: map[string]bool{}}
}

// shaped reports whether the windows torn off draw their own shape: on
// Linux, which doesn't round a window's corners, they are transparent, and
// their content is a rounded surface with a shadow, margin pixels round it
// (which resize the window, as GTK's own shadows do). Elsewhere the system
// rounds them and draws their shadow.
func shaped() bool { return runtime.GOOS == "linux" }

// margin is the transparent room round a shaped window's surface, for its
// shadow.
const margin = 16

// pad is the room the window takes round its surface: the margin where the
// window draws its own shape.
func pad() int {
	if shaped() {
		return margin
	}
	return 0
}

// frame wraps the content of a window torn off, view: where the window
// draws its own shape, a surface with the corners of the Dock's areas (its
// own areas fill it edge to edge), a hairline and a soft shadow, in a
// transparent window.
func frame(view func(*ui.Context)) func(*ui.Context) {
	if !shaped() {
		return view
	}
	return func(c *ui.Context) {
		var t *ryugo.Tokens
		var surface ui.Element
		ui.Box(c).Fill().Padding(margin).Children(func() {
			surface = ui.Box(c).Fill().Clip().Children(func() {
				view(c)
				t = ryugo.From(c)
			})
		})
		r := t.R(ryugo.RadiusControl)
		surface.Radius(r).Shadow(0, 4, 14, 0, ui.RGBA(0, 0, 0, 0.28)).Shadow(0, 0, 0, 0.5, ui.RGBA(0, 0, 0, 0.12)).
			DrawOver(func(p *ui.Painter, b ui.Rect) { p.Stroke(b, t.Colors.Hairline, r, 1) })
		c.Root().Background(ui.Transparent)
	}
}

// placed reports whether the system says where windows are: Wayland's
// compositors don't, unless GTK runs on X11 (GDK_BACKEND=x11).
func placed() bool {
	return runtime.GOOS != "linux" || os.Getenv("WAYLAND_DISPLAY") == "" || os.Getenv("GDK_BACKEND") == "x11"
}

func (h *Host) win(id string) *mygo.Window {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.wins[id]
}

// Content is where window id's content is on the screen; ok is false
// where the system doesn't say, with its size still right.
func (h *Host) Content(id string) (ui.Rect, bool) {
	w := h.win(id)
	if w == nil {
		return ui.Rect{}, false
	}
	b := w.ContentBounds()
	return ui.Rect{X: float32(b.X), Y: float32(b.Y), W: float32(b.Width), H: float32(b.Height)}, placed()
}

// Pointer is where the pointer is on the screen; ok is false where the
// system doesn't say where windows are.
func (h *Host) Pointer() (ui.Point, bool) {
	if !placed() {
		return ui.Point{}, false
	}
	p := mygo.Screen.CursorScreenPoint()
	return ui.Point{X: float32(p.X), Y: float32(p.Y)}, true
}

// Move puts the content of window id at p on the screen, faded (see
// ryugo.DockHost).
func (h *Host) Move(id string, p ui.Point, faded bool) {
	w := h.win(id)
	if w == nil {
		return
	}
	b := w.ContentBounds()
	if x, y := int(p.X), int(p.Y); x != b.X || y != b.Y {
		w.SetContentBounds(mygo.Rectangle{X: x, Y: y, Width: b.Width, Height: b.Height})
	}
	o := 1.0
	if faded {
		o = 0.55
	}
	if w.Opacity() != o {
		w.SetOpacity(o)
	}
}

// Redraw redraws every window of the layout: call it when what they show
// changed outside a Dock (the app's data). A Dock redraws them itself when
// its layout or its theme changes.
func (h *Host) Redraw() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, w := range h.wins {
		w.Invalidate()
	}
}

// Sync opens a window for each of the layout's windows that has none,
// closes those of windows the layout no longer holds, and redraws them
// all. Call it after restoring a layout, to open its windows.
func (h *Host) Sync() {
	l := h.layout()
	h.mu.Lock()
	defer h.mu.Unlock()
	if l == nil {
		return
	}
	for _, dw := range l.Windows {
		if dw.ID == "main" || h.wins[dw.ID] != nil || h.opening[dw.ID] {
			continue
		}
		h.opening[dw.ID] = true
		// a window opens from a goroutine, never in a frame
		go h.open(*dw)
	}
	for id, w := range h.wins {
		if id != "main" && l.Window(id) == nil {
			delete(h.wins, id)
			go w.Destroy()
		}
	}
	for _, w := range h.wins {
		w.Invalidate()
	}
}

// open opens a window for dw, where it was, or where the system puts it
// when that is on no display.
func (h *Host) open(dw ryugo.DockWindow) {
	// the layout keeps the surface's place and size; the window takes the
	// margin round it too
	x, y, wd, ht := dw.X, dw.Y, dw.Width, dw.Height
	if wd <= 0 || ht <= 0 {
		wd, ht = 480, 360
	}
	m := pad()
	wd, ht = wd+2*m, ht+2*m
	if x != 0 || y != 0 {
		x, y = x-m, y-m
	}
	if !placed() || !onDisplay(x, y, wd) {
		x, y = 0, 0
	}
	id := dw.ID
	// no title bar, but the window's shadow, corners and edges that resize
	// it: the Dock's headers move it, and its top right area closes it
	w := mygo.NewWindow(mygo.WindowOptions{
		Title: title(dw), X: x, Y: y, Width: wd, Height: ht, UseContentSize: true,
		MinWidth: 240, MinHeight: 180,
		TitleBarStyle: mygo.TitleBarHidden, HideWindowButtons: true, Transparent: shaped(),
		Content: ui.View(frame(h.view(id))),
	})
	h.mu.Lock()
	delete(h.opening, id)
	if l := h.layout(); l == nil || l.Window(id) == nil {
		// gone while it opened
		h.mu.Unlock()
		w.Destroy()
		return
	}
	h.wins[id] = w
	h.mu.Unlock()
	keep := func() {
		// where it is and its size go into the layout, to be saved
		l := h.layout()
		if l == nil {
			return
		}
		if dw := l.Window(id); dw != nil {
			b := w.ContentBounds()
			m := pad()
			dw.Width, dw.Height = b.Width-2*m, b.Height-2*m
			if placed() {
				dw.X, dw.Y = b.X+m, b.Y+m
			}
			h.changed()
		}
	}
	w.OnMove(keep)
	w.OnResize(keep)
	w.OnClose(func(*mygo.CloseEvent) {
		// closing loses nothing: what it holds goes back to the main window
		h.mu.Lock()
		delete(h.wins, id)
		h.mu.Unlock()
		if l := h.layout(); l != nil && l.DockBack(id) {
			h.changed()
		}
		h.main.Invalidate()
	})
	keep()
	h.main.Invalidate()
}

func (h *Host) changed() {
	if h.OnChange != nil {
		h.OnChange()
	}
}

// onDisplay reports whether a window at x, y, w wide shows its title bar
// on a display: a window saved on a display since gone opens elsewhere.
func onDisplay(x, y, w int) bool {
	if x == 0 && y == 0 {
		return false
	}
	for _, d := range mygo.Screen.Displays() {
		a := d.WorkArea
		if x+w > a.X+40 && x < a.X+a.Width-40 && y >= a.Y && y < a.Y+a.Height-40 {
			return true
		}
	}
	return false
}

// title is a window's title: the pane its first area shows.
func title(dw ryugo.DockWindow) string {
	n := dw.Root
	for n != nil && !n.IsArea() {
		n = n.Children[0]
	}
	if n == nil || len(n.Panes) == 0 {
		return ""
	}
	return n.Panes[min(n.Active, len(n.Panes)-1)].Title
}
