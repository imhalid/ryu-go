// Package palettewin floats a board's palettes in windows of their own,
// as Pixelmator Classic's, on MyGo's windows: a ryugo.PaletteHost. Each
// palette is a borderless, transparent child window of the app's, which
// the palette draws itself in; it stays over the app's window and closes
// with it. Where the system doesn't put a window where asked, nor say
// where the pointer is (Wayland, unless GTK runs on X11 with
// GDK_BACKEND=x11), Free reports false and the palettes float over their
// board in the app's window instead. Experimental, as the palettes are.
//
// It is a package of its own so that apps whose palettes float over their
// board only don't take in MyGo's window package with it.
//
//	app.board.Host = palettewin.New(win)
package palettewin

import (
	"os"
	"runtime"
	"sync"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// Host shows palettes in MyGo's windows.
type Host struct {
	main *mygo.Window

	mu      sync.Mutex
	wins    map[string]*mygo.Window
	opening map[string]bool
	gone    map[string]bool // hidden while their windows opened
	rects   map[string]ui.Rect
}

// New returns the host of the palettes of a board in main.
func New(main *mygo.Window) *Host {
	return &Host{main: main, wins: map[string]*mygo.Window{}, opening: map[string]bool{},
		gone: map[string]bool{}, rects: map[string]ui.Rect{}}
}

// Free reports whether the palettes float in windows of their own: where
// the system puts a window where asked and says where the pointer is.
func (h *Host) Free() bool {
	return runtime.GOOS != "linux" || os.Getenv("WAYLAND_DISPLAY") == "" || os.Getenv("GDK_BACKEND") == "x11"
}

// Board is where the main window's content is on the screen.
func (h *Host) Board() ui.Rect { return rect(h.main.ContentBounds()) }

// Work is the work area of the display the main window is on.
func (h *Host) Work() ui.Rect {
	b := h.main.ContentBounds()
	return rect(mygo.Screen.DisplayNearestPoint(mygo.Point{X: b.X + b.Width/2, Y: b.Y + b.Height/2}).WorkArea)
}

// Pointer is where the pointer is on the screen.
func (h *Host) Pointer() ui.Point {
	p := mygo.Screen.CursorScreenPoint()
	return ui.Point{X: float32(p.X), Y: float32(p.Y)}
}

func rect(r mygo.Rectangle) ui.Rect {
	return ui.Rect{X: float32(r.X), Y: float32(r.Y), W: float32(r.Width), H: float32(r.Height)}
}

func bounds(r ui.Rect) mygo.Rectangle {
	x, y := int(r.X), int(r.Y)
	if x == 0 && y == 0 {
		// zero is "centered" to a new window
		x = 1
	}
	return mygo.Rectangle{X: x, Y: y, Width: int(r.W + 0.5), Height: int(r.H + 0.5)}
}

// Show shows palette id in its window, its content at r on the screen,
// opening the window the first time, from a goroutine (never in a frame).
func (h *Host) Show(id string, r ui.Rect, view func(*ui.Context)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.gone, id)
	if w := h.wins[id]; w != nil {
		if h.rects[id] != r {
			h.rects[id] = r
			w.SetContentBounds(bounds(r))
		}
		return
	}
	h.rects[id] = r
	if h.opening[id] || view == nil {
		return
	}
	h.opening[id] = true
	go h.open(id, view)
}

func (h *Host) open(id string, view func(*ui.Context)) {
	h.mu.Lock()
	r := h.rects[id]
	h.mu.Unlock()
	b := bounds(r)
	// no frame, no shadow of the system's: the palette draws its own, in a
	// margin of its window
	w := mygo.NewWindow(mygo.WindowOptions{
		Title: id, X: b.X, Y: b.Y, Width: b.Width, Height: b.Height, UseContentSize: true,
		Frameless: true, Transparent: true, DisableShadow: true, DisableResize: true,
		Parent: h.main, SkipTaskbar: true,
		Content: ui.View(view),
	})
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.opening, id)
	if h.gone[id] {
		delete(h.gone, id)
		go w.Destroy()
		return
	}
	h.wins[id] = w
	// moved or resized while it opened
	if now := h.rects[id]; now != r {
		w.SetContentBounds(bounds(now))
	}
	h.main.Invalidate()
}

// Hide closes palette id's window.
func (h *Host) Hide(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if w := h.wins[id]; w != nil {
		delete(h.wins, id)
		go w.Destroy()
		return
	}
	if h.opening[id] {
		h.gone[id] = true
	}
}

// Redraw redraws the main window and the palettes' windows, but except.
func (h *Host) Redraw(except string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if except != "main" {
		h.main.Invalidate()
	}
	for id, w := range h.wins {
		if id != except {
			w.Invalidate()
		}
	}
}
