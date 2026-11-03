// Gallery, Mizan bileşenlerinin vitrinini bir pencerede açar.
//
//	go run ./cmd/gallery
package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/imhalid/mizan-go/gallery"
)

func main() {
	s := gallery.New()
	mygo.App.WhenReady(func() {
		mygo.NewWindow(mygo.WindowOptions{
			Title:   "Mizan",
			Width:   1180,
			Height:  860,
			Content: ui.View(s.View),
		})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
