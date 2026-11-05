package ryugo

import (
	"fmt"

	"github.com/egoist/mygo/ui"
)

// TableLabels are the words a Table says, for an app in another language; a
// field left empty keeps the English one.
//
//	ryugo.Table(c.Key("t"), &st, cols, rows, ryugo.TableOptions[Row]{
//		Key: key, Labels: ryugo.TableLabels{Search: "Ara", Selected: "%d seçili", NoRows: "Satır yok"},
//	})
type TableLabels struct {
	// Search is the search box's placeholder and name; View, the toolbar
	// menu's button, with Columns and Density as its headings and Compact,
	// Default and Comfortable as its densities.
	Search, View, Columns, Density, Compact, Default, Comfortable string
	// Selected is the count in the selection bar, with a %d for the number
	// of rows chosen; Clear empties the selection.
	Selected, Clear string
	// SelectRow and SelectAll name the checkboxes; Details, the chevron
	// that opens a row.
	SelectRow, SelectAll, Details string
	// NoRows is an empty table's title; LoadError, the title of a table whose
	// rows did not load, with TryAgain its button.
	NoRows, LoadError, TryAgain string
	// Pagination is the words of the pages under the table.
	Pagination PaginationLabels
}

// fill is l with the English words where it has none.
func (l TableLabels) fill() TableLabels {
	for _, d := range []struct {
		to   *string
		word string
	}{
		{&l.Search, "Search"}, {&l.View, "View"}, {&l.Columns, "Columns"}, {&l.Density, "Density"},
		{&l.Compact, "Compact"}, {&l.Default, "Default"}, {&l.Comfortable, "Comfortable"},
		{&l.Selected, "%d selected"}, {&l.Clear, "Clear"},
		{&l.SelectRow, "Select row"}, {&l.SelectAll, "Select all rows"}, {&l.Details, "Details"},
		{&l.NoRows, "No rows"}, {&l.LoadError, "Couldn't load the rows"}, {&l.TryAgain, "Try again"},
	} {
		if *d.to == "" {
			*d.to = d.word
		}
	}
	return l
}

// selected is the selection bar's words for n rows chosen.
func (l TableLabels) selected(n int) string { return fmt.Sprintf(l.Selected, n) }

// PaginationLabels are the words Pagination says; a field left empty keeps the
// English one.
//
//	ryugo.PaginationOptions{PageSize: 20, Labels: ryugo.PaginationLabels{Of: "/", Previous: "Önceki sayfa", Next: "Sonraki sayfa"}}
type PaginationLabels struct {
	// Group names the control to assistive technology ("Pages"); Of sits
	// between the range and the total ("21–40 of 200").
	Group, Of string
	// Previous and Next name the buttons; Page names a number ("Page 3",
	// the number follows) and Current marks the page shown.
	Previous, Next, Page, Current string
}

// fill is l with the English words where it has none.
func (l PaginationLabels) fill() PaginationLabels {
	for _, d := range []struct {
		to   *string
		word string
	}{
		{&l.Group, "Pages"}, {&l.Of, "of"}, {&l.Previous, "Previous page"}, {&l.Next, "Next page"},
		{&l.Page, "Page"}, {&l.Current, "current"},
	} {
		if *d.to == "" {
			*d.to = d.word
		}
	}
	return l
}

// TableSize is the height of a Table's header and of a row at size step s, for
// an app that gives a table a Height of its rows.
//
//	head, row := ryugo.TableSize(ryugo.Md)
//	height := head + float32(len(rows))*row
func TableSize(s Size) (head, row float32) {
	m := tableSizes(s)
	return m.head, m.row
}

// PaginationHeight is the height of a Pagination at size step s.
//
//	height += ryugo.PaginationHeight(c, ryugo.Md)
func PaginationHeight(c *ui.Context, s Size) float32 { return sizedFrom(c, s).Metrics.BtnH }
