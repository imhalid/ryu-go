package mizan

import (
	"slices"
	"time"

	"github.com/egoist/mygo/ui"
)

// Drag and drop speaks in two parts: what is dragged is the raised skin (it
// can be held), where it is dropped is an inset well. When something it
// accepts comes over it, the well takes a solid accent border and a
// background tinted 8% toward the accent, as in Mizan's file zone; the
// dragged item's place stays faded to 40%.

var (
	iconUpload = Svg(16, 20, "M8 10.5V2.5M5 5.5l3-3 3 3M2.5 10v2A1.5 1.5 0 004 13.5h8a1.5 1.5 0 001.5-1.5v-2")
	iconFile   = Svg(16, 14, "M9 1.5H4.5a1.5 1.5 0 0 0-1.5 1.5v10a1.5 1.5 0 0 0 1.5 1.5h7a1.5 1.5 0 0 0 1.5-1.5V5.5zM9 1.5v4h4")
)

// dropWell paints a drop target: normally a well; when something comes over
// it, an accent border and a background tinted toward the accent.
func (t *Tokens) dropWell(e ui.Element, over bool) {
	c := t.Colors
	if over {
		e.Background(Mix(c.Well, c.ColorAccent500, 0.08)).Border(1, c.ColorAccent500)
	} else {
		e.Background(c.Well).Border(1, c.WellBorder)
	}
	e.Transition(ui.ElementTransition{Duration: 150 * time.Millisecond, Colors: true})
}

// grip is the grip: dots in two columns of three rows.
func grip(c *ui.Context, color ui.Color) {
	ui.Row(c).Gap(3).Shrink(0).Children(func() {
		for range 2 {
			ui.Column(c).Gap(3).Children(func() {
				for range 3 {
					ui.Box(c).Size(2.5, 2.5).Radius(1.25).Background(color)
				}
			})
		}
	})
}

// tile is the raised tile being dragged: the button's skin, the control's corner radius.
func (t *Tokens) tile(e ui.Element, hovered bool) {
	cl := t.Colors
	e.Radius(RadiusControl).Cursor(ui.CursorGrab)
	t.raised(e, t.controlGradient(cl.ControlTop, cl.ControlBottom), cl.ControlBorder, t.Shadows.ControlShadow,
		raisedState{hovered: hovered})
}

// SortableOptions are the options of Sortable.
type SortableOptions struct {
	// Label is a row's text: its accessible name, and for finding it by keyboard.
	Label func(i int) string
	// Key is a row's identity; its index if unset. Provide it so that rows'
	// state (local values, focus) is kept as they move.
	Key func(i int) any
}

// Sortable is a list whose rows are reordered by dragging: each row is a
// raised tile with a grip, and a line in the accent color shows where it
// will drop; the list reorders *items itself. s must stay the same across
// frames.
//
//	mizan.Sortable(c, &app.list, &app.tasks, mizan.SortableOptions{}, func(t Task) {
//		ui.Text(c, t.Title)
//	})
func Sortable[T any](c *ui.Context, s *ui.ListState, items *[]T, o SortableOptions, row func(T)) ui.Element {
	t := From(c)
	cl := t.Colors
	s.Label, s.Key = o.Label, o.Key
	s.Reorder = func(rows []int, to int) { *items = move(*items, rows, to) }
	return ui.List(c, s, len(*items), func(i int) {
		e := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(12).Padding(10, 14, 10, 12).MinHeight(t.Metrics.ControlH + 8)
		t.tile(e, e.Hovered())
		e.TextColor(cl.Text).FontSize(FsMd).Children(func() {
			grip(c, cl.Muted)
			ui.Row(c).Grow(1).MinWidth(0).AlignItems(ui.Center).Gap(8).Children(func() { row((*items)[i]) })
		})
	}).Gap(6).Padding(2)
}

// move moves the rows at the indices rows to before index to (keeping their order).
func move[T any](items []T, rows []int, to int) []T {
	moved := make([]T, 0, len(rows))
	rest := make([]T, 0, len(items))
	at := to
	for i, v := range items {
		if slices.Contains(rows, i) {
			moved = append(moved, v)
			if i < to {
				at--
			}
		} else {
			rest = append(rest, v)
		}
	}
	return slices.Insert(rest, at, moved...)
}

// DragCard is a raised tile dragged to another place (a DropArea); value
// goes to where it is dropped. While dragged, its place stays faded and a
// copy follows the pointer.
//
//	mizan.DragCard(c.Key(task.ID), task, func() { ui.Text(c, task.Title) })
func DragCard(c *ui.Context, value any, body func()) ui.Element {
	t := From(c)
	e := ui.Column(c).FillWidth().Gap(4).Padding(10, 12).MinWidth(0).Drag(value)
	t.tile(e, e.Hovered())
	e.TextColor(t.Colors.Text).FontSize(FsMd).Children(body)
	if e.Dragging() {
		e.Opacity(0.4)
	}
	return e
}

// DropAreaOptions are the options of DropArea.
type DropAreaOptions struct {
	Title string
	// Count is the neutral count next to the title (not shown if empty).
	Count string
	// Accept says whether to take the dragged value (every value of T if
	// unset): to refuse a card from the same column, for example.
	Accept func(v any) bool
	// Fill stretches the column to its container's height: columns side by
	// side stay the same height, and an empty column still has room to drop on.
	Fill bool
}

// DropArea is an inset column that takes values of type T (a board column,
// a folder): it highlights when a value it accepts comes over it, and
// returns the dropped value.
//
//	if task, ok := mizan.DropArea[*Task](c.Key("done"), mizan.DropAreaOptions{Title: "Done"}, func() {
//		...
//	}); ok {
//		task.Done = true
//	}
func DropArea[T any](c *ui.Context, o DropAreaOptions, body func()) (T, bool) {
	t := From(c)
	cl := t.Colors
	e := ui.Column(c).Gap(8).Padding(10).Radius(RadiusLg).MinWidth(0)
	if o.Fill {
		e.Grow(1)
	}
	v, over := ui.DragOver[T](e)
	if over && o.Accept != nil && !o.Accept(v) {
		over = false
	}
	t.dropWell(e, over)
	t.Shadows.WellShadow.Apply(e)
	e.Children(func() {
		if o.Title != "" {
			ui.Row(c).AlignItems(ui.Center).Gap(8).Padding(2, 4, 4, 4).Children(func() {
				ui.Text(c, o.Title).TextColor(cl.Text2).FontSize(FsSm).FontWeight(600)
				if o.Count != "" {
					Badge(c, o.Count, ToneNeutral)
				}
			})
		}
		body()
	})
	dropped, ok := ui.Drop[T](e)
	if ok && o.Accept != nil && !o.Accept(dropped) {
		var zero T
		return zero, false
	}
	return dropped, ok
}

// DropZoneOptions are the options of DropZone.
type DropZoneOptions struct {
	// Title is the bold line ("Drop files here"); Hint the muted description
	// below it ("PNG, PDF · up to 10 MB").
	Title string
	Hint  string
	// Browse, if set, makes the zone clickable and is called when it is
	// clicked (for example, a file picker with mygo.Dialog.Open).
	Browse   func()
	Disabled bool
}

// DropZone is a well with a dashed border that takes files from other apps
// (Finder, Explorer): when files come over it, the border turns solid and
// takes the accent color. It returns the dropped files' paths, or nil.
//
//	if paths := mizan.DropZone(c.Key("upload"), mizan.DropZoneOptions{Title: "Drop files"}); paths != nil {
//		app.add(paths)
//	}
func DropZone(c *ui.Context, o DropZoneOptions) []string {
	t := From(c)
	cl := t.Colors
	var e ui.Element
	if o.Browse != nil {
		e = ui.ButtonBase(c).Label(or(o.Title, "Choose files")).Cursor(ui.CursorPointer).Disabled(o.Disabled).FocusRing(false)
	} else {
		e = ui.Column(c)
	}
	e.FillWidth().AlignItems(ui.Center).Gap(4).Padding(24, 16).Radius(RadiusControl)
	over := !o.Disabled && e.FileDragOver()
	var files []string
	if !o.Disabled {
		files = e.DroppedFiles()
	}
	switch {
	case over:
		e.Background(Mix(cl.Well, cl.ColorAccent500, 0.08)).Border(1, cl.ColorAccent500)
	case o.Browse != nil && e.FocusVisible():
		e.Background(cl.Well).Border(1, cl.ColorAccent500).BorderStyle(ui.BorderDashed)
		t.Shadows.Ring.Apply(e)
	case e.Hovered() && !o.Disabled:
		e.Background(cl.Well).Border(1, cl.Muted).BorderStyle(ui.BorderDashed)
	default:
		e.Background(cl.Well).Border(1, Mix(ui.Color{}, cl.Muted, 0.45)).BorderStyle(ui.BorderDashed)
	}
	e.Transition(ui.ElementTransition{Duration: 150 * time.Millisecond, Colors: true})
	if o.Disabled {
		e.Opacity(0.4)
	}
	e.TextColor(cl.Muted).FontSize(FsSm).Children(func() {
		ui.Icon(c, iconUpload).Size(20, 20).TextColor(cl.Muted).Margin(0, 0, 4, 0)
		if o.Title != "" {
			ui.Text(c, o.Title).TextColor(cl.Text).FontSize(FsMd).FontWeight(600).TextAlign(ui.Center)
		}
		if o.Hint != "" {
			ui.Text(c, o.Hint).TextAlign(ui.Center).Wrap()
		}
	})
	if o.Browse != nil && !o.Disabled && e.Clicked() {
		o.Browse()
	}
	return files
}

// FileItem is the row of an attached file: a well background, the name
// truncated, the size tabular, a remove button on the right. It returns
// true when Remove is pressed.
func FileItem(c *ui.Context, name, size string) bool {
	t := From(c)
	cl := t.Colors
	removed := false
	ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(10).MinHeight(36).Padding(0, 4, 0, 12).Radius(RadiusItem).
		Background(cl.Well).FontSize(FsSm).Children(func() {
		ui.Icon(c, iconFile).TextColor(cl.Muted)
		ui.Text(c, name).TextColor(cl.Text).Grow(1).MinWidth(0).SingleLine().Ellipsis("…")
		if size != "" {
			ui.Text(c, size).TextColor(cl.Muted).FontFeatures("tnum").NoWrap()
		}
		removed = inputButton(c.Key("remove"), t, iconClear, "Remove").Margin(0).Clicked()
	})
	return removed
}
