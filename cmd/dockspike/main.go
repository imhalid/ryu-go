// Dockspike is step 0 of DOCK.md: it finds out whether a pane dragged by
// its bar can leave the window, open in a window of its own where it is let
// go, come back by a drag onto another window, and come back when its
// window closes, without losing what it holds (its count and its scroll).
// It logs what the drag sees, to read after a try in the real window.
//
//	go run ./cmd/dockspike
package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/mizan-go"
)

// pane is a pane of the model: what it holds lives here, never in a
// window, so a pane changing windows loses nothing.
type pane struct {
	ID, Title string
	Count     int
	Scroll    ui.ScrollState
	Win       string // the window it shows in: "main", or a torn-off one's ID
}

// drag is a pane being dragged by its bar, in screen points.
type drag struct {
	id     string
	from   string
	moved  float32
	active bool
	grab   ui.Point // where the bar is held, from the bar's top left
	screen ui.Point // the pointer on the screen
	inside bool     // whether the pointer is over the window it started in
	logged time.Time
}

type spike struct {
	panes []*pane
	mu    sync.Mutex
	wins  map[string]*mygo.Window
	next  int
	drag  drag
}

func main() {
	s := &spike{wins: map[string]*mygo.Window{}}
	for i, t := range []string{"Orders", "Console", "Notes"} {
		s.panes = append(s.panes, &pane{ID: fmt.Sprint(i), Title: t, Win: "main"})
	}
	mygo.App.WhenReady(func() {
		w := mygo.NewWindow(mygo.WindowOptions{Title: "Dock spike", Width: 760, Height: 640, Content: ui.View(s.view("main"))})
		s.setWin("main", w)
		logf("main window: bounds %+v, content %+v", w.Bounds(), w.ContentBounds())
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

func logf(format string, args ...any) {
	fmt.Printf("%s  %s\n", time.Now().Format("15:04:05.000"), fmt.Sprintf(format, args...))
}

func (s *spike) win(id string) *mygo.Window {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.wins[id]
}

func (s *spike) setWin(id string, w *mygo.Window) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if w == nil {
		delete(s.wins, id)
	} else {
		s.wins[id] = w
	}
}

// redraw asks every window for a frame: the model changed.
func (s *spike) redraw() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, w := range s.wins {
		w.Invalidate()
	}
}

// view builds window id: its panes, side by side.
func (s *spike) view(id string) func(c *ui.Context) {
	return func(c *ui.Context) {
		t := mz.Root(c, mz.Options{})
		cl := t.Colors
		ui.Column(c).Fill().Padding(16).Gap(12).Background(cl.Bg).Children(func() {
			mz.Text(c, "Window "+id+": drag a pane by its bar out of the window, or onto another one", mz.TBody).TextColor(cl.Muted)
			ui.Row(c).Grow(1).Gap(12).AlignItems(ui.Stretch).Children(func() {
				for _, p := range s.panes {
					if p.Win == id {
						s.pane(c, t, id, p)
					}
				}
			})
		})
	}
}

// pane builds p in window id: a bar to drag it by, a counter, and a list
// to scroll, the two things it must not lose.
func (s *spike) pane(c *ui.Context, t *mz.Tokens, win string, p *pane) {
	cl := t.Colors
	ui.Column(c.Key(p.ID)).Grow(1).Basis(0).MinWidth(0).Radius(12).Background(cl.Card).Clip().Children(func() {
		bar := ui.Row(c).Height(34).Padding(0, 12).AlignItems(ui.Center).Background(cl.Fill).Cursor(ui.CursorGrab)
		bar.Children(func() { ui.Text(c, p.Title).FontWeight(600) })
		s.dragBar(c, win, p, bar)
		ui.Column(c).Grow(1).MinHeight(0).Padding(12).Gap(8).Children(func() {
			if mz.Button(c.Key("count"), fmt.Sprintf("Count: %d", p.Count), mz.ButtonOptions{Size: mz.Sm}).Clicked() {
				p.Count++
			}
			ui.Scroll(c).Grow(1).TrackScroll(&p.Scroll).Children(func() {
				for i := range 80 {
					ui.Text(c, fmt.Sprintf("%s line %d", p.Title, i+1)).FontSize(13)
				}
			})
		})
	})
}

// dragBar follows a press on p's bar: once it moves a few points, the pane
// is dragged; let go, it goes to the window under the pointer, or to a new
// window there when the pointer is over none.
func (s *spike) dragBar(c *ui.Context, win string, p *pane, bar ui.Element) {
	d := &s.drag
	dx, dy, pressed := bar.Dragged()
	w := s.win(win)
	if w == nil {
		return
	}
	cb := w.ContentBounds()
	b := bar.LayoutBounds()
	px, py, over := bar.PointerPosition()
	screen := ui.Point{X: float32(cb.X) + b.X + px, Y: float32(cb.Y) + b.Y + py}
	switch {
	case pressed && !d.active:
		if d.moved += abs(dx) + abs(dy); d.moved > 4 {
			*d = drag{id: p.ID, from: win, active: true, grab: ui.Point{X: px, Y: py}, screen: screen}
			logf("drag %s from %s: start at screen %v", p.Title, win, screen)
		}
	case pressed && d.active && d.id == p.ID:
		d.screen = screen
		d.inside = screen.X >= float32(cb.X) && screen.X < float32(cb.X+cb.Width) &&
			screen.Y >= float32(cb.Y) && screen.Y < float32(cb.Y+cb.Height)
		if time.Since(d.logged) > 120*time.Millisecond {
			d.logged = time.Now()
			logf("drag %s: local (%.0f, %.0f) over %v, screen (%.0f, %.0f), inside %s: %v",
				p.Title, b.X+px, b.Y+py, over, screen.X, screen.Y, win, d.inside)
		}
	case !pressed && d.active && d.id == p.ID:
		d.active = false
		logf("drag %s: let go at screen %v, inside %s: %v", p.Title, d.screen, win, d.inside)
		s.drop(p, d.screen, d.grab)
	case !pressed:
		d.moved = 0
	}
}

// drop puts p in the window under at, or in a new one opened there.
func (s *spike) drop(p *pane, at, grab ui.Point) {
	s.mu.Lock()
	var to string
	for id, w := range s.wins {
		cb := w.ContentBounds()
		if at.X >= float32(cb.X) && at.X < float32(cb.X+cb.Width) && at.Y >= float32(cb.Y) && at.Y < float32(cb.Y+cb.Height) {
			to = id
		}
	}
	s.mu.Unlock()
	from := p.Win
	switch {
	case to == from:
		logf("drop %s: back in %s, nothing to do", p.Title, from)
		return
	case to != "":
		p.Win = to
		logf("drop %s: moved from %s to %s (count %d, scroll %.0f)", p.Title, from, to, p.Count, p.Scroll.Y)
	default:
		s.next++
		to = fmt.Sprintf("w%d", s.next)
		p.Win = to
		logf("drop %s: outside every window, opening %s at %v (count %d, scroll %.0f)", p.Title, to, at, p.Count, p.Scroll.Y)
		go s.open(to, p.Title, at, grab)
	}
	s.closeIfEmpty(from)
	s.redraw()
}

// open opens window id with its content's top left where the bar was held.
func (s *spike) open(id, title string, at, grab ui.Point) {
	m := s.win("main")
	// how far the content is from the window's top left: the title bar
	off := 0
	if m != nil {
		off = m.ContentBounds().Y - m.Bounds().Y
	}
	w := mygo.NewWindow(mygo.WindowOptions{
		Title: title, Width: 420, Height: 380, UseContentSize: true,
		X: int(at.X - grab.X - 12), Y: int(at.Y-grab.Y-16) - off,
		Content: ui.View(s.view(id)),
	})
	s.setWin(id, w)
	logf("%s open: bounds %+v, content %+v", id, w.Bounds(), w.ContentBounds())
	w.OnClose(func(e *mygo.CloseEvent) {
		// closing loses nothing: its panes go back to the main window
		var back []string
		for _, p := range s.panes {
			if p.Win == id {
				p.Win = "main"
				back = append(back, fmt.Sprintf("%s (count %d, scroll %.0f)", p.Title, p.Count, p.Scroll.Y))
			}
		}
		s.setWin(id, nil)
		logf("%s closing: back to main: %s", id, strings.Join(back, ", "))
		s.redraw()
	})
}

// closeIfEmpty closes a torn-off window left with no pane.
func (s *spike) closeIfEmpty(id string) {
	if id == "main" {
		return
	}
	for _, p := range s.panes {
		if p.Win == id {
			return
		}
	}
	if w := s.win(id); w != nil {
		s.setWin(id, nil)
		logf("%s empty: closing it", id)
		go w.Destroy()
	}
}

func abs(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}
