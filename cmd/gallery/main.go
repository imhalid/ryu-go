// Gallery opens the showcase of Mizan's components in a window. The
// workspace's layout is saved in the user's config directory on every
// change, and opens again as it was, its windows torn off too.
//
//	go run ./cmd/gallery
package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/mizan-go"
	"github.com/imhalid/mizan-go/dockwin"
	"github.com/imhalid/mizan-go/gallery"
)

func main() {
	s := gallery.New()
	file := layoutFile()
	if b, err := os.ReadFile(file); err == nil {
		if l, err := mz.ParseLayout(b); err == nil {
			s.Work.Layout = l
		}
	}
	save := func() {
		b, err := s.Layout().MarshalJSON()
		if err != nil || file == "" {
			return
		}
		os.MkdirAll(filepath.Dir(file), 0o755)
		os.WriteFile(file, b, 0o644)
	}
	var host *dockwin.Host
	// the windows torn off take the theme chosen in the main one at once
	var theme [4]any
	view := func(c *ui.Context) {
		s.View(c)
		if now := [4]any{s.Mode, s.Accent, s.Density, s.Round}; now != theme {
			theme = now
			if host != nil {
				host.Redraw()
			}
		}
	}
	mygo.App.WhenReady(func() {
		w := mygo.NewWindow(mygo.WindowOptions{
			Title:   "Mizan",
			Width:   1180,
			Height:  860,
			Content: ui.View(view),
		})
		host = dockwin.New(w, s.Layout, s.WindowView)
		host.OnChange = save
		s.Work.Host, s.Work.OnChange = host, save
		// the windows saved with the layout open again
		host.Sync()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

// layoutFile is where the workspace's layout is saved.
func layoutFile() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "mizan-gallery", "layout.json")
}
