package ryugo

import (
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/egoist/mygo/ui"
)

// A Table is Ryugo's data table: rows on the table surface under a header
// in its tint, divided by hairlines, numbers tabular and at the end. Its
// rows are built only as far as they show, at a height set by the size
// step, so that thousands stay quick. It moves as the rest of Ryugo does,
// never by a flash of color: rows slide to their new places as the sort
// changes and as rows come and go, a column picked up by its header lifts
// and the others slide aside for it, a selection bar springs in over the
// toolbar.

var (
	iconSortArrow = Svg(16, 12, "M4.5 9.5L8 6l3.5 3.5")
	iconView      = Svg(16, 15, "M2.5 4.5h7M12.5 4.5h1M2.5 11.5h1M6.5 11.5h7M11 3v3M5 10v3")
	iconAlert     = Svg(16, 18, "M8 5.5v3.5M8 11.5v.01M7.1 2.6l-5.4 9.6A1 1 0 002.6 13.7h10.8a1 1 0 00.9-1.5L8.9 2.6a1 1 0 00-1.8 0z")
	iconSearch    = Svg(16, 15, "M7 12.5a5.5 5.5 0 1 0 0-11 5.5 5.5 0 0 0 0 11zM14.5 14.5l-3.6-3.6")
)

// SortKey is one key of a Table's sort: the column's ID, descending or not.
type SortKey struct {
	Column string
	Desc   bool
}

// TableState is what a Table keeps from frame to frame that an app may
// want to read, set or save: the sort, the rows chosen, the columns'
// order, widths and visibility, the rows opened, the search, the page and
// the density. The zero value is a table as its columns declare it.
type TableState struct {
	// Sort is the sort, first key first; a click on a header sorts by it,
	// a Shift-click adds it after the others.
	Sort []SortKey
	// Selected holds the keys of the rows chosen.
	Selected map[any]bool
	// Order is the columns' order by ID; columns it does not name follow
	// in the order declared.
	Order []string
	// Widths are the widths the user gave columns, by ID.
	Widths map[string]float32
	// Hidden are the columns the user hid, by ID.
	Hidden map[string]bool
	// Expanded holds the keys of the rows opened (TableOptions.Expand).
	Expanded map[any]bool
	// Query is the text in the toolbar's search box; the app filters the
	// rows it gives the table by it.
	Query string
	// Page is the page shown, from 0, with TableOptions.PageSize.
	Page int
	// Density, when set, is the size step the user chose in the view
	// menu, over TableOptions.Size.
	Density Size
	// Collapsed holds the groups folded (TableOptions.Group), by key.
	Collapsed map[string]bool

	scroll ui.ScrollState
	cursor int
	anchor any
}

// TableAction is an action of the selection bar, on the rows chosen.
type TableAction struct {
	Label  string
	Icon   *ui.SVG
	Danger bool
	// Do runs the action on the keys of the rows chosen, in the order
	// shown.
	Do func(keys []any)
}

// TableOptions are the options of Table.
type TableOptions[R any] struct {
	// Key is a row's identity, comparable and unique: selection, opened
	// rows and the rows' motion follow it.
	Key func(r R) any
	// Label names the table to assistive technology.
	Label string
	// Selectable adds a column of checkboxes, and a box in the header
	// choosing every row. Shift-click chooses a range.
	Selectable bool
	// Actions are the selection bar's buttons, shown while rows are
	// chosen.
	Actions []TableAction
	// Search puts a search box in the toolbar, writing TableState.Query.
	Search bool
	// Toolbar builds more of the toolbar, after the search box.
	Toolbar func()
	// ViewMenu puts a menu at the toolbar's end showing and hiding the
	// columns and choosing the density.
	ViewMenu bool
	// Expand, if set, builds a row's details, which open under it from a
	// chevron at its start.
	Expand func(c *ui.Context, r R)
	// Pin keeps the first column (with the checkboxes and chevrons) in
	// place as the table scrolls sideways.
	Pin bool
	// PageSize, if set, shows the rows a page at a time, with Pagination
	// under the table.
	PageSize int
	// Height is the table's height; 0 fills its parent (Grow).
	Height float32
	// Loading shows placeholders in the rows' place; Error, a message and
	// Retry; Empty is what an empty table shows.
	Loading bool
	Error   string
	Retry   func()
	Empty   EmptyStateOptions
	Size    Size
	// Open, if set, runs when a row is clicked or Enter is pressed on the row
	// with the cursor: the app opens what the row is. Rows show a pointer then.
	Open func(r R)
	// Row, if set, is called with each row's element as it is built, for what the
	// app adds to it: a context menu, a handler.
	Row func(row ui.Element, r R)
	// HideSelectionBar leaves the bar of the rows chosen to the app, which shows its own.
	HideSelectionBar bool
	// Group, if set, puts each row in the group it names, under a heading
	// row: the groups go in the order they first come in the rows given,
	// and a header's sort sorts the rows within each group. A click on a
	// heading folds the group, or opens it (TableState.Collapsed). With
	// PageSize, pages count the rows alone, folded ones too: a page shows
	// the heading of each group it has rows of, and a folded group's
	// heading alone.
	//
	//	Group: func(o Order) string { return o.Region },
	Group func(r R) string
	// GroupLabel, if set, is a group's heading, given its key and its rows;
	// the key by default.
	GroupLabel func(key string, rows []R) string
	// Labels are the table's words, for an app in another language.
	Labels TableLabels
}

// tableRT is what a Table keeps for itself between frames.
type tableRT[R any] struct {
	ready   bool
	prevY   map[any]float32
	prevKey []any
	known   map[any]bool
	last    map[any]R
	ghosts  []tableGhost[R]
	detailH map[any]float32
	detailP map[any]float32
	availW  float32
	viewH   float32
	shift   map[string]float32
	col     colDrag
	order   []string // a new column order, applied in the next frame
	orderAt time.Time
	headX   float32
	colGen  int     // counts column moves, for fresh shift channels
	prevTop float32 // the scroll the rows last showed at
	page    int     // the page last shown
	turn    float32 // a page turn: 1 forward, -1 back, 0 none
	count   int     // the rows chosen the selection bar last showed
	reveal  bool    // the cursor row moved by the keys: scroll it into view
	fadeMax float32 // how far the table scrolled sideways when its edge fades were last worked out
}

// tableGhost is a row taken out, fading where it was.
type tableGhost[R any] struct {
	row  R
	key  any
	y    float32
	at   time.Time
	turn float32 // the page turn it went with: it leaves the other way
}

// colDrag is a column being moved by its header.
type colDrag struct {
	id      string
	moved   float32
	active  bool
	grab    ui.Point
	pos     ui.Point
	w       float32
	at      int
	headTop float32
}

// tableLine is a line of a table's body: row n of the rows shown, or
// with n < 0, the heading of group g.
type tableLine struct {
	n int
	g string
}

// groupKey keys a group's heading apart from the rows.
type groupKey struct{ g string }

// tableMetrics are a table's sizes at a size step.
type tableMetrics struct {
	row, head, fs, pad float32
}

func tableSizes(s Size) tableMetrics {
	switch s {
	case Sm:
		return tableMetrics{row: 34, head: 32, fs: FsSm, pad: 10}
	case Lg:
		return tableMetrics{row: 48, head: 40, fs: FsMd, pad: 14}
	}
	return tableMetrics{row: 40, head: 36, fs: FsMd, pad: 12}
}

const (
	selectW = 44
	expandW = 36
)

// tableFrame is a Table being built: what its parts share in a frame.
type tableFrame[R any] struct {
	t      *Tokens
	c      *ui.Context
	s      *TableState
	o      TableOptions[R]
	cols   []Column[R]
	rows   []R
	m      tableMetrics
	rt     *tableRT[R]
	outer  ui.Element
	order  []int
	widths map[int]float32
	lead   float32
	total  float32
	pinned int
	base   map[string]float32 // where each column starts, unmoved
	shift  map[string]float32 // how far each column shows moved
	all    []int              // the rows' indexes, sorted
	idx    []int              // those on the page shown, and not folded
	keys   []any              // their keys
	lines  []tableLine        // the lines shown: the rows of idx, and the groups' headings
	group  map[string][]int   // the rows of each group, by key, in the order of all
	l      TableLabels        // the words, in English where the app gave none
	chosen []any              // the keys chosen, in the order shown
	focus  bool
	now    time.Time
}

// Table shows rows of type R under a header of columns. It sorts the rows
// by s.Sort itself; the app filters them (by s.Query with Search). See
// TableOptions for selection, the selection bar, the toolbar, opened
// rows, the pinned column, groups, pages and the loading, error and empty
// states, and Column.Foot for a footer row.
// The user sorts by a click on a header (Shift-click adds a key), moves a
// column by dragging its header, and resizes it by the header's end edge
// (a double click there gives it back its width). With the keyboard focus
// on the table, Up and Down move the cursor row, Space chooses it, Shift
// with the arrows chooses a range, and Enter opens it.
//
//	cols := []ryugo.Column[Order]{
//		{Title: "Customer", Kind: ryugo.KindPerson, Value: func(o Order) any { return o.Customer }, Sortable: true},
//		{Title: "Total", Kind: ryugo.KindNumber, Value: func(o Order) any { return o.Total }, Sortable: true, Decimals: 2},
//	}
//	ryugo.Table(c.Key("orders"), &app.table, cols, app.orders, ryugo.TableOptions[Order]{
//		Key: func(o Order) any { return o.ID }, Selectable: true,
//	})
func Table[R any](c *ui.Context, s *TableState, cols []Column[R], rows []R, o TableOptions[R]) ui.Element {
	if o.Key == nil {
		panic("ryugo: Table needs a Key")
	}
	t := sizedFrom(c, or(s.Density, o.Size))
	f := &tableFrame[R]{t: t, c: c, s: s, o: o, cols: cols, rows: rows, m: tableSizes(t.Size), now: c.Now(), l: o.Labels.fill(),
		shift: map[string]float32{}}
	f.outer = ui.Column(c).FillWidth().MinWidth(0).Gap(10)
	if o.Height > 0 {
		f.outer.Height(o.Height)
	} else {
		f.outer.Grow(1).MinHeight(0)
	}
	f.rt = *ui.Local(f.outer, "rt", func() *tableRT[R] {
		return &tableRT[R]{prevY: map[any]float32{}, known: map[any]bool{}, last: map[any]R{},
			detailH: map[any]float32{}, detailP: map[any]float32{}, shift: map[string]float32{}}
	})
	f.placeColumns()
	f.sortRows()
	f.columnShifts()

	var body ui.Element
	f.outer.Children(func() {
		f.bar()
		// the border is drawn over the header and rows: under them, their
		// fills would cover it in the rounded corners
		fadeL, fadeR := f.edgeFade()
		box := ui.Column(c).FillWidth().Grow(1).MinHeight(0).Radius(t.R(RadiusControl)).Background(t.Colors.TableBg).
			Clip().Role(ui.RoleTable).Focusable().FocusRing(false).
			DrawOver(func(p *ui.Painter, r ui.Rect) {
				f.paintEdgeFade(p, r, fadeL, fadeR)
				p.Stroke(r, t.Colors.Hairline, t.R(RadiusControl), 1)
			})
		if o.Label != "" {
			box.Label(o.Label)
		}
		f.keyboard(box)
		f.focus = box.FocusVisible()
		box.Children(func() {
			head := ui.Row(c).Height(f.m.head).Shrink(0).Role(ui.RoleRow).Clip().
				Background(f.headBg()).InsetShadow(0, -1, 0, 0, t.Colors.Hairline)
			body = ui.ScrollBoth(c).Grow(1).MinHeight(0).TrackScroll(&s.scroll)
			head.ScrollWith(body)
			head.Children(f.header)
			f.sortLight(head)
			body.Children(f.body)
			if f.hasFoot() {
				// under the rows' view, so it stays at the bottom as they
				// scroll, and goes sideways with them as the header does
				foot := ui.Row(c).Height(f.m.head).Shrink(0).Role(ui.RoleRow).Clip().
					Background(f.headBg()).InsetShadow(0, 1, 0, 0, t.Colors.Hairline)
				foot.ScrollWith(body)
				foot.Children(f.footer)
			}
			if b := head.LayoutBounds(); b.W > 0 {
				f.rt.col.headTop = b.Y
				f.rt.headX = b.X
			}
		})
		if o.PageSize > 0 && len(f.all) > o.PageSize {
			Pagination(c.Key("pages"), &s.Page, len(f.all), PaginationOptions{PageSize: o.PageSize, Size: t.Size, Labels: f.l.Pagination})
		}
	})
	if b := body.LayoutBounds(); b.W > 0 {
		f.rt.availW, f.rt.viewH = b.W, b.H
	}
	f.rt.ready = true
	return f.outer
}

func (f *tableFrame[R]) headBg() ui.Color { return over(f.t.Colors.Fill, f.t.Colors.TableBg) }

// placeColumns applies a column move let go in the last frame, every
// column going on from where it showed to its new place, and lays out the
// visible columns.
func (f *tableFrame[R]) placeColumns() {
	rt, s := f.rt, f.s
	if rt.order != nil && f.now.After(rt.orderAt) {
		before := f.columnStarts(tableOrder(f.cols, s))
		s.Order = rt.order
		rt.order = nil
		after := f.columnStarts(tableOrder(f.cols, s))
		// new channels, at rest: the pointer's jumps would otherwise
		// carry their speed into the slide and fling the column
		rt.colGen++
		for id, x := range after {
			f.outer.Jump(f.colKey(id), before[id]+rt.shift[id]-x)
		}
	}
	f.order = tableOrder(f.cols, s)
	f.widths, f.lead, f.total = tableWidths(f.cols, s, f.order, f.o, f.rt.availW)
	f.base = f.columnStarts(f.order)
	if f.o.Pin && len(f.order) > 0 {
		f.pinned = 1
	}
}

// columnStarts is where each column of order starts in a row, by ID.
func (f *tableFrame[R]) columnStarts(order []int) map[string]float32 {
	widths, lead, _ := tableWidths(f.cols, f.s, order, f.o, f.rt.availW)
	xs := map[string]float32{}
	x := lead
	for _, i := range order {
		xs[f.cols[i].id()] = x
		x += widths[i]
	}
	return xs
}

// sortRows sorts the rows by the table's sort, takes the page shown, and
// finds the rows chosen.
func (f *tableFrame[R]) sortRows() {
	s, rows := f.s, f.rows
	idx := make([]int, len(rows))
	for i := range idx {
		idx[i] = i
	}
	if len(s.Sort) > 0 {
		values := map[string]func(R) any{}
		for i := range f.cols {
			if f.cols[i].Value != nil {
				values[f.cols[i].id()] = f.cols[i].Value
			}
		}
		slices.SortStableFunc(idx, func(a, b int) int {
			for _, k := range s.Sort {
				v := values[k.Column]
				if v == nil {
					continue
				}
				if d := compareValues(v(rows[a]), v(rows[b])); d != 0 {
					if k.Desc {
						return -d
					}
					return d
				}
			}
			return 0
		})
	}
	var groups []string // the group of each row, by index
	if g := f.o.Group; g != nil {
		// the groups in the order they first come in the rows given; the
		// sort, stable, keeps its order within each
		groups = make([]string, len(rows))
		rank := map[string]int{}
		f.group = map[string][]int{}
		for i, r := range rows {
			groups[i] = g(r)
			if _, ok := rank[groups[i]]; !ok {
				rank[groups[i]] = len(rank)
			}
		}
		slices.SortStableFunc(idx, func(a, b int) int { return rank[groups[a]] - rank[groups[b]] })
		for _, i := range idx {
			f.group[groups[i]] = append(f.group[groups[i]], i)
		}
	}
	f.all, f.idx = idx, idx
	page := idx
	if ps := f.o.PageSize; ps > 0 {
		pages := max(1, (len(idx)+ps-1)/ps)
		s.Page = min(max(s.Page, 0), pages-1)
		page = idx[s.Page*ps : min(len(idx), (s.Page+1)*ps)]
	}
	f.idx, f.keys, f.lines = nil, nil, nil
	for p, i := range page {
		if groups != nil {
			g := groups[i]
			if p == 0 || groups[page[p-1]] != g {
				f.lines = append(f.lines, tableLine{n: -1, g: g})
			}
			if s.Collapsed[g] {
				continue
			}
		}
		f.lines = append(f.lines, tableLine{n: len(f.idx)})
		f.idx = append(f.idx, i)
		f.keys = append(f.keys, f.o.Key(rows[i]))
	}
	f.chosen = nil
	for _, i := range f.all {
		if k := f.o.Key(rows[i]); s.Selected[k] {
			f.chosen = append(f.chosen, k)
		}
	}
	s.cursor = min(max(s.cursor, 0), max(len(f.keys)-1, 0))
}

// columnShifts moves the columns for the one being dragged: it follows
// the pointer, and the others slide aside where it would go and close the
// room it left, as cards do in a list.
func (f *tableFrame[R]) columnShifts() {
	d, s := &f.rt.col, f.s
	if f.rt.order != nil {
		// let go, waiting for the next frame to take the new order: the
		// columns hold still where they show
		for id, v := range f.rt.shift {
			f.shift[id] = v
		}
		return
	}
	target := map[string]float32{}
	dragged := ""
	if dk := f.colIndex(d.id); d.active && dk >= 0 {
		dragged = d.id
		w := f.widths[f.order[dk]]
		origin := f.rt.headX - s.scroll.X
		left := d.pos.X - d.grab.X - origin
		mid := left + w/2
		at := 0
		for k, i := range f.order {
			if k == dk {
				continue
			}
			x := f.base[f.cols[i].id()]
			if k > dk {
				x -= w
			}
			if x+f.widths[i]/2 < mid {
				at++
			}
		}
		d.at = max(at, f.pinned)
		n := 0
		for k, i := range f.order {
			id := f.cols[i].id()
			if k == dk {
				target[id] = left - f.base[id]
				continue
			}
			var sh float32
			if k > dk {
				sh -= w
			}
			if n >= d.at {
				sh += w
			}
			target[id] = sh
			n++
		}
	} else {
		d.active = false
	}
	for _, i := range f.order {
		id := f.cols[i].id()
		if id == dragged {
			f.shift[id] = f.outer.Jump(f.colKey(id), target[id])
		} else {
			f.shift[id] = f.outer.AnimateTo(f.colKey(id), target[id], springShift)
		}
	}
	f.rt.shift = f.shift
}

// colKey is the animation channel of column id's shift, new after each
// move of a column.
func (f *tableFrame[R]) colKey(id string) string { return fmt.Sprintf("col:%s:%d", id, f.rt.colGen) }

func (f *tableFrame[R]) colIndex(id string) int {
	for k, i := range f.order {
		if f.cols[i].id() == id {
			return k
		}
	}
	return -1
}

// bar is the bar over the table: the toolbar, and the selection bar that
// springs in over it while rows are chosen, as a raised strip with the
// count, the actions and Clear. Without a toolbar, the selection bar
// opens room for itself.
func (f *tableFrame[R]) bar() {
	t, c, s, o := f.t, f.c, f.s, f.o
	cl := t.Colors
	tools := o.Search || o.Toolbar != nil || o.ViewMenu
	if n := len(f.chosen); n > 0 {
		f.rt.count = n
	}
	p := clamp01(f.outer.AnimateTo("selbar", b2f(len(f.chosen) > 0 && !o.HideSelectionBar), springSelBar))
	if !tools && p < 0.001 {
		return
	}
	h := t.Metrics.ControlH + 4
	slot := ui.Box(c).FillWidth().Shrink(0).Height(h)
	if !tools {
		// it opens room as it comes, the gap under it too
		slot.Height(h*p).Margin(0, 0, -10*(1-p), 0).ClipY()
	}
	slot.Children(func() {
		if tools {
			// the two bars hand over rather than cross: the toolbar is
			// mostly gone before the selection bar is much there
			tb := ui.Row(c).Absolute().Left(0).Right(0).Top(0).Bottom(0).AlignItems(ui.Center).Gap(8).
				Opacity(clamp01(1-p/0.55)).Translate(0, -6*p).IgnoreInput(p > 0.5)
			tb.Children(func() {
				if o.Search {
					ui.Box(c).Width(260).Shrink(1).MinWidth(140).Children(func() {
						Input(c.Key("search"), &s.Query, InputOptions{Placeholder: f.l.Search, PrefixIcon: iconSearch, Clearable: true, Label: f.l.Search, Size: Sm})
					})
				}
				if o.Toolbar != nil {
					o.Toolbar()
				}
				ui.Spacer(c)
				if o.ViewMenu {
					f.viewMenu()
				}
			})
		}
		in := p
		if tools {
			in = clamp01((p - 0.3) / 0.7)
		}
		sel := ui.Row(c).Absolute().Left(0).Right(0).Top(2).Bottom(2).AlignItems(ui.Center).Gap(8).
			Padding(0, 6, 0, 14).Radius(t.R(RadiusControl)).Opacity(in).Translate(0, 8*(1-p)).IgnoreInput(p < 0.5).
			LinearGradient(t.controlGradient(cl.ControlTop, cl.ControlBottom)).Border(0.5, cl.ControlBorder)
		t.Shadows.ControlShadow.Apply(sel)
		if o.HideSelectionBar {
			sel.Invisible()
		}
		sel.Children(func() {
			ui.Text(c, f.l.selected(f.rt.count)).FontSize(FsMd).FontWeight(600).TextColor(cl.Text).FontFeatures("tnum").NoWrap()
			if len(o.Actions) > 0 {
				Separator(c, SeparatorOptions{Vertical: true})
			}
			for n, a := range o.Actions {
				if Button(c.Key("act"+strconv.Itoa(n)), a.Label, ButtonOptions{Icon: a.Icon, Danger: a.Danger, Size: Sm}).Clicked() && a.Do != nil {
					a.Do(slices.Clone(f.chosen))
				}
			}
			ui.Spacer(c)
			if Button(c.Key("clear"), f.l.Clear, ButtonOptions{Size: Sm}).Clicked() {
				s.Selected = nil
			}
		})
	})
}

// springSelBar is how the selection bar comes and goes.
var springSelBar = ui.Animation{Spring: ui.SpringOptions{Duration: 320 * time.Millisecond, Bounce: 0.1}}

// viewMenu is the toolbar's menu showing and hiding the columns and
// choosing the density.
func (f *tableFrame[R]) viewMenu() {
	s := f.s
	MenuButton(f.c.Key("view"), f.l.View, ButtonOptions{Icon: iconView, Size: Sm}, func(m *MenuList) {
		m.Label(f.l.Columns)
		shown := len(f.order)
		for i := range f.cols {
			col := &f.cols[i]
			id := col.id()
			on := !s.Hidden[id]
			// the last column shown stays
			if m.Item(col.Title, ItemOptions{Checkable: true, Checked: on, Disabled: on && shown == 1}) {
				if s.Hidden == nil {
					s.Hidden = map[string]bool{}
				}
				s.Hidden[id] = on
			}
		}
		m.Separator()
		m.Label(f.l.Density)
		for _, d := range []struct {
			size Size
			name string
		}{{Sm, f.l.Compact}, {Md, f.l.Default}, {Lg, f.l.Comfortable}} {
			if m.Item(d.name, ItemOptions{Checkable: true, Checked: f.t.Size == d.size}) {
				s.Density = d.size
			}
		}
	})
}

// keyboard moves the cursor row with the arrows, Home and End, chooses
// with Space (a range with Shift and the arrows) and opens with Enter.
func (f *tableFrame[R]) keyboard(box ui.Element) {
	s, n := f.s, len(f.keys)
	if n == 0 {
		return
	}
	move := func(to int, extend bool) {
		from := s.cursor
		s.cursor = min(max(to, 0), n-1)
		f.rt.reveal = true
		if extend && f.o.Selectable {
			if s.anchor == nil {
				s.anchor = f.keys[from]
			}
			f.chooseRange(s.anchor, f.keys[s.cursor], true)
		}
	}
	switch {
	case box.Shortcut(0, ui.KeyDown):
		move(s.cursor+1, false)
	case box.Shortcut(0, ui.KeyUp):
		move(s.cursor-1, false)
	case box.Shortcut(ui.Shift, ui.KeyDown):
		move(s.cursor+1, true)
	case box.Shortcut(ui.Shift, ui.KeyUp):
		move(s.cursor-1, true)
	case box.Shortcut(0, ui.KeyHome):
		move(0, false)
	case box.Shortcut(0, ui.KeyEnd):
		move(n-1, false)
	case box.Shortcut(0, ui.KeyPageDown):
		move(s.cursor+10, false)
	case box.Shortcut(0, ui.KeyPageUp):
		move(s.cursor-10, false)
	case box.Shortcut(0, ui.KeySpace) && f.o.Selectable:
		k := f.keys[s.cursor]
		f.choose(k, !s.Selected[k])
		s.anchor = k
	case box.Shortcut(0, ui.KeyEnter) && f.o.Open != nil:
		f.o.Open(f.rows[f.idx[s.cursor]])
	case box.Shortcut(0, ui.KeyEnter) && f.o.Expand != nil:
		f.toggleOpen(f.keys[s.cursor])
	}
}

func (f *tableFrame[R]) choose(k any, on bool) {
	s := f.s
	if on {
		if s.Selected == nil {
			s.Selected = map[any]bool{}
		}
		s.Selected[k] = true
	} else {
		delete(s.Selected, k)
	}
}

// chooseRange sets the rows shown from a to b, in either order, to on.
func (f *tableFrame[R]) chooseRange(a, b any, on bool) {
	ia, ib := slices.Index(f.keys, a), slices.Index(f.keys, b)
	if ia < 0 || ib < 0 {
		f.choose(b, on)
		return
	}
	for n := min(ia, ib); n <= max(ia, ib); n++ {
		f.choose(f.keys[n], on)
	}
}

func (f *tableFrame[R]) toggleOpen(k any) {
	s := f.s
	if s.Expanded == nil {
		s.Expanded = map[any]bool{}
	}
	if s.Expanded[k] {
		delete(s.Expanded, k)
	} else {
		s.Expanded[k] = true
	}
}

// header builds the header row: the box choosing every row, the column
// headers, and the pinned ones over them.
func (f *tableFrame[R]) header() {
	c := f.c
	row := ui.Row(c).Width(f.total).Height(f.m.head).AlignItems(ui.Stretch).Shrink(0)
	row.Children(func() {
		if f.pinned > 0 {
			ui.Box(c).Width(f.lead + f.widths[f.order[0]]).Shrink(0)
		} else {
			f.headLead()
		}
		for k, i := range f.order {
			if k < f.pinned {
				continue
			}
			f.headCell(k, i)
		}
	})
	if f.pinned > 0 {
		// ScrollWith leaves the header's absolute cells in place already
		pin := f.pinGroup(ui.Row(c), f.m.head, f.headBg(), false).InsetShadow(0, -1, 0, 0, f.t.Colors.Hairline)
		pin.Children(func() {
			f.headLead()
			f.headCell(0, f.order[0])
		})
	}
}

// hasFoot is whether the table shows a footer: a column has a Foot, and
// the rows are there to sum.
func (f *tableFrame[R]) hasFoot() bool {
	if f.o.Loading || f.o.Error != "" {
		return false
	}
	for _, i := range f.order {
		if f.cols[i].Foot != nil {
			return true
		}
	}
	return false
}

// footer builds the footer row: each column's Foot under it, in the
// column's place and width, muted, the pinned one held over the rest as
// the header's is.
func (f *tableFrame[R]) footer() {
	c := f.c
	row := ui.Row(c).Width(f.total).Height(f.m.head).AlignItems(ui.Stretch).Shrink(0)
	row.Children(func() {
		lead := f.lead
		if f.pinned > 0 {
			lead += f.widths[f.order[0]]
		}
		ui.Box(c).Width(lead).Shrink(0)
		for k, i := range f.order {
			if k >= f.pinned {
				f.footCell(i)
			}
		}
	})
	if f.pinned > 0 {
		pin := f.pinGroup(ui.Row(c), f.m.head, f.headBg(), false).InsetShadow(0, 1, 0, 0, f.t.Colors.Hairline)
		pin.Children(func() {
			ui.Box(c).Width(f.lead).Shrink(0)
			f.footCell(f.order[0])
		})
	}
}

// footCell is column i's cell in the footer: its Foot over every row,
// aligned as its cells are, numbers tabular; it goes with the column
// as the column moves.
func (f *tableFrame[R]) footCell(i int) {
	c := f.c
	col := &f.cols[i]
	id := col.id()
	e := ui.Row(c.Key("f:"+id)).Width(f.widths[i]).Shrink(0).Padding(0, f.m.pad).AlignItems(ui.Center).Clip().Role(ui.RoleCell)
	if col.align() == ui.End {
		e.Justify(ui.End)
	}
	if d := &f.rt.col; d.active && d.id == id {
		e.Opacity(0.45)
	}
	e.Translate(f.shift[id], 0).Children(func() {
		if col.Foot != nil {
			ui.Text(c, col.Foot(f.rows)).FontSize(f.m.fs).FontWeight(500).TextColor(f.t.Colors.Muted).
				FontFeatures("tnum").SingleLine().Ellipsis("…")
		}
	})
}

// sortLight lights the top of the header of the column the rows sort by
// first: a thin line of the accent along its top edge, brightest in its
// middle and fading to its ends (lifted toward white in the dark theme),
// and under it a soft glow fading down into the header. It slides to the
// column a new sort picks, as Tabs' underline does, and fades as the sort
// goes.
func (f *tableFrame[R]) sortLight(head ui.Element) {
	cl := f.t.Colors
	var x, w float32
	on := false
	if len(f.s.Sort) > 0 {
		if k := f.colIndex(f.s.Sort[0].Column); k >= 0 {
			id := f.s.Sort[0].Column
			x, w, on = f.base[id]+f.shift[id], f.widths[f.order[k]], true
			if k >= f.pinned {
				x -= f.s.scroll.X
			}
		}
	}
	placed := ui.Local(head, "light-placed", func() bool { return false })
	a := head.AnimateTo("light", b2f(on), ui.Animation{Duration: 220 * time.Millisecond, Ease: ui.EaseOut})
	if !on && a < 0.01 {
		*placed = false
		return
	}
	if on {
		if !*placed || f.rt.col.active {
			x, w = head.Jump("light-x", x), head.Jump("light-w", w)
			*placed = true
		} else {
			x, w = head.AnimateTo("light-x", x, springPill), head.AnimateTo("light-w", w, springPill)
		}
	} else {
		x, w = head.AnimateTo("light-x", x, springPill), head.AnimateTo("light-w", w, springPill)
	}
	tone := cl.ColorAccent500
	core := tone
	if f.t.Dark {
		core = Mix(tone, ui.RGB(255, 255, 255), 0.45)
	}
	band := func(c ui.Color, alpha float32) ui.LinearGradient {
		return ui.LinearGradient{Angle: 90, Oklab: true, Stops: []ui.GradientStop{
			{Color: c.Alpha(0), At: 0}, {Color: c.Alpha(alpha * 0.6), At: 0.25}, {Color: c.Alpha(alpha), At: 0.5},
			{Color: c.Alpha(alpha * 0.6), At: 0.75}, {Color: c.Alpha(0), At: 1},
		}}
	}
	glow := float32(0.09)
	if f.t.Dark {
		glow = 0.16
	}
	head.DrawOver(func(p *ui.Painter, r ui.Rect) {
		// under the table's border, which is drawn over the header
		x0, y0 := r.X+x, r.Y+1
		// the glow: strips fading and narrowing as they go down
		const strips, step = 9, 2
		for i := range strips {
			k := float32(i) / strips
			inset := w * (0.1 + 0.2*k)
			p.FillGradient(ui.Rect{X: x0 + inset, Y: y0 + float32(i*step), W: w - 2*inset, H: step}, band(tone, glow*a*(1-k)*(1-k)), 0)
		}
		p.FillGradient(ui.Rect{X: x0 + 8, Y: y0, W: w - 16, H: 1.5}, band(core, a), 0)
	})
}

// pinGroup styles e as the pinned cells of a row: over the row's start,
// kept in place as the rest scrolls sideways (follow: by moving back as
// far as the row scrolls), on the row's background, with a soft shadow at
// its end once something scrolls under it.
func (f *tableFrame[R]) pinGroup(e ui.Element, h float32, bg ui.Color, follow bool) ui.Element {
	x := f.s.scroll.X
	e.Absolute().Left(0).Top(0).Height(h).AlignItems(ui.Stretch).Background(bg)
	if follow {
		e.Translate(x, 0)
	}
	if a := clamp01(x / 24); a > 0 {
		e.Shadow(6, 0, 12, -6, ui.Color{A: uint8(60 * a)})
	}
	return e
}

// headLead is the header's leading cells: the box choosing every row and
// the room over the chevrons.
func (f *tableFrame[R]) headLead() {
	c, s := f.c, f.s
	if f.o.Selectable {
		ui.Box(c).Width(selectW).Shrink(0).Center().Children(func() {
			n := len(f.chosen)
			all := n > 0 && n == len(f.all)
			if f.t.tableCheck(c.Key("all"), all, n > 0 && !all, f.l.SelectAll).Clicked() {
				if all {
					s.Selected = nil
				} else {
					for _, i := range f.all {
						f.choose(f.o.Key(f.rows[i]), true)
					}
				}
			}
		})
	}
	if f.o.Expand != nil {
		ui.Box(c).Width(expandW).Shrink(0)
	}
}

// headCell is column order[k]'s header: its title, and the arrow of the
// sort, which turns as the sort turns and shows faintly under the pointer
// on a column that sorts. A click sorts by it, a Shift-click adds it to
// the sort; a drag moves the column, and its end edge resizes it.
func (f *tableFrame[R]) headCell(k, i int) {
	t, c, s := f.t, f.c, f.s
	cl := t.Colors
	col := &f.cols[i]
	id := col.id()
	d := &f.rt.col
	h := ui.Row(c.Key("h:"+id)).Width(f.widths[i]).Shrink(0).Padding(0, f.m.pad).AlignItems(ui.Center).Gap(4).
		Role(ui.RoleColumnHeader).Label(col.Title)
	if col.align() == ui.End {
		h.Justify(ui.End)
	}
	at, desc := sortPlace(s.Sort, id)
	hovered := h.Hovered() && !d.active
	moves := !col.Fixed && k >= f.pinned
	dragged := d.active && d.id == id
	if moves {
		f.dragHeader(h, id, f.widths[i])
	}
	if col.Sortable {
		h.Cursor(ui.CursorPointer)
		if h.Clicked() && !dragged {
			s.Sort = sortClick(s.Sort, id, h.ClickModifiers()&ui.Shift != 0)
		}
	}
	// content builds the header's title and arrow, in it or in the copy
	// held up while it moves
	content := func(lifted bool) {
		color := cl.Muted
		if col.Sortable && (hovered || lifted) || at >= 0 {
			color = cl.Text2
		}
		ui.Text(c, col.Title).FontSize(FsSm).FontWeight(500).TextColor(color).SingleLine().Ellipsis("…")
		if !col.Sortable {
			return
		}
		show := float32(0)
		switch {
		case at >= 0:
			show = 1
		case hovered:
			show = 0.35
		}
		a := ui.Icon(c, iconSortArrow).Size(12, 12).TextColor(cl.Text2).Shrink(0)
		rot := a.AnimateTo("rot", 180*b2f(desc), springPill)
		a.Rotate(rot).Opacity(a.AnimateTo("show", show, ui.Animation{Duration: 160 * time.Millisecond, Ease: ui.EaseOut}))
		if at >= 0 && len(s.Sort) > 1 {
			ui.Text(c, strconv.Itoa(at+1)).FontSize(Fs2xs).FontWeight(600).TextColor(cl.Muted).FontFeatures("tnum")
		}
	}
	h.Children(func() {
		content(false)
		// no edge at the table's end, where it would only crowd the corner
		if moves && k < len(f.order)-1 {
			f.resizeEdge(col, f.widths[i])
		}
	})
	if dragged {
		// its place stays empty while the header is held
		h.Opacity(0)
		f.headGhost(content, f.widths[i])
	} else {
		h.Translate(f.shift[id], 0)
	}
}

// dragHeader picks up a column by its header once the pointer pressing it
// moved a few DIPs, follows the pointer, and lets the column go to where
// it was taken: the new order applies in the next frame.
func (f *tableFrame[R]) dragHeader(h ui.Element, id string, w float32) {
	d := &f.rt.col
	moved := ui.Local(h, "moved", func() float32 { return 0 })
	dx, dy, pressed := h.Dragged()
	switch {
	case pressed && d.active && d.id == id:
		b := h.LayoutBounds()
		px, py, _ := h.PointerPosition()
		d.pos = ui.Point{X: b.X + px, Y: b.Y + py}
	case pressed && !d.active:
		if *moved += abs32(dx) + abs32(dy); *moved > 4 {
			b := h.LayoutBounds()
			px, py, _ := h.PointerPosition()
			*d = colDrag{id: id, active: true, grab: ui.Point{X: px, Y: py}, pos: ui.Point{X: b.X + px, Y: b.Y + py},
				w: w, at: f.colIndex(id), headTop: b.Y}
		}
	case !pressed:
		*moved = 0
		if d.active && d.id == id {
			d.active = false
			f.rt.order = f.moveColumn(id, d.at)
			f.rt.orderAt = f.now
			f.c.AnimationFrame()
		}
	}
}

// moveColumn is every column's ID with id taken to place at among the
// visible columns but itself, hidden ones keeping theirs.
func (f *tableFrame[R]) moveColumn(id string, at int) []string {
	var others []string
	for _, i := range f.order {
		if oid := f.cols[i].id(); oid != id {
			others = append(others, oid)
		}
	}
	full := slices.DeleteFunc(fullOrder(f.cols, f.s), func(x string) bool { return x == id })
	pos := len(full)
	switch {
	case at < len(others):
		pos = slices.Index(full, others[at])
	case len(others) > 0:
		pos = slices.Index(full, others[len(others)-1]) + 1
	}
	return slices.Insert(full, pos, id)
}

// headGhost is the header being moved, lifted off the table: it follows
// the pointer on the header's line, a little larger, on the raised skin.
func (f *tableFrame[R]) headGhost(content func(lifted bool), w float32) {
	t, c, d := f.t, f.c, &f.rt.col
	cl := t.Colors
	ui.Overlay(c, func() {
		x := d.pos.X - d.grab.X
		g := ui.Row(c).Absolute().Left(x).Top(d.headTop).Width(w).Height(f.m.head).Padding(0, f.m.pad).
			AlignItems(ui.Center).Gap(4).Radius(t.R(RadiusItem)).PassThrough().Cursor(ui.CursorGrabbing).
			LinearGradient(t.controlGradient(cl.ControlTop, cl.ControlBottom)).Border(0.5, cl.ControlBorder)
		t.Shadows.ShadowPop.Apply(g)
		picked := ui.Local(g, "picked", func() bool { return false })
		if !*picked {
			*picked = true
			g.Jump("lift", 0)
		}
		lift := g.AnimateTo("lift", 1, springLift)
		sc := 1 + 0.04*lift
		g.Scale(sc, sc).Children(func() { content(true) })
	})
}

// resizeEdge is the end edge of a column's header, which resizes the
// column as it is dragged, and gives it back its width on a double click:
// a short hairline, darker under the pointer.
func (f *tableFrame[R]) resizeEdge(col *Column[R], w float32) {
	c, s := f.c, f.s
	cl := f.t.Colors
	id := col.id()
	{
		edge := ui.Box(c.Key("edge")).Absolute().Top(0).Bottom(0).Right(-4).Width(9).Cursor(ui.CursorResizeEW)
		if edge.DoubleClicked() {
			delete(s.Widths, id)
		}
		if dx, _, held := edge.Dragged(); held && dx != 0 {
			if s.Widths == nil {
				s.Widths = map[string]float32{}
			}
			cur, ok := s.Widths[id]
			if !ok {
				cur = w
			}
			s.Widths[id] = max(col.minWidth()*0.5, 48, cur+dx)
		}
		hot := edge.Hovered() || edge.Pressed()
		a := edge.AnimateTo("hot", b2f(hot), ui.Animation{Duration: 120 * time.Millisecond})
		line := Mix(cl.Hairline, cl.Muted, a)
		edge.Draw(func(p *ui.Painter, r ui.Rect) {
			hh := r.H * (0.4 + 0.25*a)
			p.Fill(ui.Rect{X: r.X + r.W/2 - 0.5, Y: r.Y + (r.H-hh)/2, W: 1, H: hh}, line, 0.5)
		})
	}
}

// sortPlace is where column id is in the sort, -1 when it is not, and
// whether it sorts descending.
func sortPlace(sort []SortKey, id string) (int, bool) {
	for n, k := range sort {
		if k.Column == id {
			return n, k.Desc
		}
	}
	return -1, false
}

// sortClick is the sort after a click on column id's header: it sorts by
// it ascending, then descending, then not at all; with add (Shift), the
// other keys stay and it goes after them.
func sortClick(sort []SortKey, id string, add bool) []SortKey {
	at, desc := sortPlace(sort, id)
	if !add {
		switch {
		case at < 0:
			return []SortKey{{Column: id}}
		case !desc:
			return []SortKey{{Column: id, Desc: true}}
		}
		return nil
	}
	sort = slices.Clone(sort)
	switch {
	case at < 0:
		return append(sort, SortKey{Column: id})
	case !desc:
		sort[at].Desc = true
		return sort
	}
	return slices.Delete(sort, at, at+1)
}

// tableCheck is a 16px checkbox of a table, as Checkbox's box: a well,
// or the accent fill with its check drawn (a dash, partial).
func (t *Tokens) tableCheck(c *ui.Context, on, partial bool, label string) ui.Element {
	b := ui.ButtonBase(c).Label(label).Role(ui.RoleCheckBox).Checked(on).Size(16, 16).Radius(min(t.R(5), 6)).Center().Shrink(0).
		Cursor(ui.CursorPointer).FocusRing(false)
	if partial {
		b.Mixed()
	}
	t.mark(b, on || partial, b.Pressed(), b.FocusVisible())
	b.Children(func() { checkMark(c, 11, 1.6, on || partial, partial && !on, false, t.Colors.OnAccent) })
	return b
}

// body builds the rows in view, between room for those above and below
// it: each row as high as the size step, and its details as far as they
// are open. Rows slide from where they showed to their new places as the
// sort changes or rows come and go; a row new to the table fades in, one
// taken out fades where it was.
func (f *tableFrame[R]) body() {
	c, s, rt, m, o := f.c, f.s, f.rt, f.m, f.o
	viewW, viewH := rt.availW, rt.viewH
	if viewH <= 0 {
		viewH = 600
	}
	if viewW <= 0 {
		viewW = f.total
	}
	switch {
	case o.Loading:
		f.placeholders(viewH)
		return
	case o.Error != "":
		f.state(viewW, viewH, EmptyStateOptions{Icon: iconAlert, Title: f.l.LoadError, Description: o.Error, Action: func() {
			if Button(c.Key("retry"), f.l.TryAgain, ButtonOptions{}).Clicked() && o.Retry != nil {
				o.Retry()
			}
		}})
		return
	case len(f.lines) == 0:
		e := o.Empty
		if e.Title == "" {
			e.Title = f.l.NoRows
		}
		f.state(viewW, viewH, e)
		f.forget()
		return
	}

	// where each line starts: a row with its details as far as they are
	// open, a group's heading
	ys := make([]float32, len(f.lines)+1)
	lineKeys := make([]any, len(f.lines))
	at := 0 // the cursor row's line
	for l, ln := range f.lines {
		h := m.head
		if ln.n < 0 {
			lineKeys[l] = groupKey{ln.g}
		} else {
			k := f.keys[ln.n]
			lineKeys[l] = k
			if ln.n == s.cursor {
				at = l
			}
			h = m.row
			if o.Expand != nil {
				p, ok := rt.detailP[k]
				if !ok && s.Expanded[k] {
					p = 1
				}
				h += rt.detailH[k] * p
			}
		}
		ys[l+1] = ys[l] + h
	}
	if rt.reveal {
		rt.reveal = false
		y0, y1 := ys[at], ys[at]+m.row
		switch {
		case y0 < s.scroll.Y:
			s.scroll.Y = y0
		case y1 > s.scroll.Y+viewH:
			s.scroll.Y = y1 - viewH
		}
	}
	over := 6 * m.row
	first, last := 0, len(f.lines)
	for first < len(f.lines) && ys[first+1] < s.scroll.Y-over {
		first++
	}
	for last > first && ys[last-1] > s.scroll.Y+viewH+over {
		last--
	}
	moved := rt.ready && !slices.Equal(lineKeys, rt.prevKey)
	rt.turn = 0
	if rt.ready && o.PageSize > 0 && s.Page != rt.page {
		rt.turn = 1
		if s.Page < rt.page {
			rt.turn = -1
		}
	}
	rt.page = s.Page

	content := ui.Column(c).Width(f.total).Shrink(0)
	content.Children(func() {
		ui.Box(c).Height(ys[first]).Shrink(0)
		for l := first; l < last; l++ {
			if ln := f.lines[l]; ln.n < 0 {
				f.heading(ln.g, ys[l], moved, viewW, viewH)
			} else {
				f.row(ln.n, ys[l], moved, viewW, viewH)
			}
		}
		ui.Box(c).Height(ys[len(f.lines)] - ys[last]).Shrink(0)
		f.ghosts(moved, viewH)
	})

	rt.prevTop = s.scroll.Y
	rt.prevY = make(map[any]float32, len(lineKeys))
	for l, k := range lineKeys {
		rt.prevY[k] = ys[l]
		rt.known[k] = true
	}
	rt.prevKey = lineKeys
}

// forget lets the rows go: the next rows shown come in as new.
func (f *tableFrame[R]) forget() {
	f.rt.prevKey, f.rt.prevY = nil, map[any]float32{}
}

// rowBg is a row's background: the table's, a little toward the text under
// the pointer, toward the accent while chosen. Opaque, for the pinned
// cells to cover what scrolls under them.
func (f *tableFrame[R]) rowBg(chosen, hovered bool) ui.Color {
	cl := f.t.Colors
	bg := cl.TableBg
	if chosen {
		bg = Mix(bg, cl.ColorAccent500, 0.08)
	}
	if hovered {
		bg = Mix(bg, cl.Text, 0.03)
	}
	return bg
}

// row builds row n at y: its line of cells, and its details under it.
func (f *tableFrame[R]) row(n int, y float32, moved bool, viewW, viewH float32) {
	c, s, rt, m, o := f.c, f.s, f.rt, f.m, f.o
	cl := f.t.Colors
	k := f.keys[n]
	r := f.rows[f.idx[n]]
	rt.last[k] = r
	wrap := ui.Column(c.Key(k)).Width(f.total).Shrink(0)
	f.arrive(wrap, k, y, moved, viewH)
	open := s.Expanded[k]
	ep := float32(0)
	if o.Expand != nil {
		ep = wrap.AnimateTo("open", b2f(open), springOpen)
		rt.detailP[k] = ep
	}
	chosen := s.Selected[k]
	wrap.Children(func() {
		line := ui.Row(c).Width(f.total).Height(m.row).Shrink(0).AlignItems(ui.Stretch).Role(ui.RoleRow).Selected(chosen)
		if o.Open != nil {
			line.Cursor(ui.CursorPointer)
		}
		if o.Row != nil {
			o.Row(line, r)
		}
		bg := f.rowBg(chosen, line.Hovered())
		line.Background(bg).InsetShadow(0, -1, 0, 0, cl.Hairline)
		if f.focus && n == s.cursor {
			line.InsetShadow(0, 0, 0, 1.5, cl.ColorAccent500.Alpha(0.55))
		}
		if line.Clicked() {
			s.cursor = n
			if o.Open != nil {
				o.Open(r)
			}
		}
		line.Children(func() {
			if f.pinned > 0 {
				ui.Box(c).Width(f.lead + f.widths[f.order[0]]).Shrink(0)
			} else {
				f.rowLead(k, n, chosen, ep)
			}
			for j, i := range f.order {
				if j < f.pinned {
					continue
				}
				f.cell(i, r)
			}
			if f.pinned > 0 {
				pin := f.pinGroup(ui.Row(c), m.row, bg, true).InsetShadow(0, -1, 0, 0, cl.Hairline)
				pin.Children(func() {
					f.rowLead(k, n, chosen, ep)
					f.cell(f.order[0], r)
				})
			}
		})
		if o.Expand != nil && (open || ep > 0.001) {
			f.details(k, r, ep, viewW)
		}
	})
}

// arrive moves wrap, the line of key k at y, from where it showed: a
// line that showed slides from where it was; one that comes from out of
// view fades in at its place instead of flying in; on a page turn, the
// new page comes in from its side.
func (f *tableFrame[R]) arrive(wrap ui.Element, k any, y float32, moved bool, viewH float32) {
	rt := f.rt
	if moved {
		py, ok := rt.prevY[k]
		switch {
		case rt.turn != 0:
			wrap.Jump("in", 0)
			wrap.Jump("px", 28*rt.turn)
		case ok && py > rt.prevTop-f.m.row && py < rt.prevTop+viewH:
			if off := py - y; abs32(off) > 0.5 {
				wrap.Jump("fy", off)
			}
		case !ok || abs32(py-y) > 0.5:
			wrap.Jump("in", 0)
		}
	} else if rt.ready && !rt.known[k] {
		wrap.Jump("in", 0)
	}
	in := wrap.AnimateTo("in", 1, ui.Animation{Duration: 240 * time.Millisecond, Ease: ui.EaseOut})
	fy := wrap.AnimateTo("fy", 0, springShift)
	px := wrap.AnimateTo("px", 0, springShift)
	rise := float32(6)
	if px != 0 {
		rise = 0
	}
	wrap.Translate(px, fy-rise*(1-in)).Opacity(in)
}

// heading builds group g's heading at y: a band across the table with
// a chevron, the group's name and its count, kept in view as the table
// scrolls sideways. A click folds the group, or opens it.
func (f *tableFrame[R]) heading(g string, y float32, moved bool, viewW, viewH float32) {
	c, s, m := f.c, f.s, f.m
	cl := f.t.Colors
	wrap := ui.Column(c.Key(groupKey{g})).Width(f.total).Shrink(0)
	f.arrive(wrap, groupKey{g}, y, moved, viewH)
	label := g
	if f.o.GroupLabel != nil {
		rows := make([]R, len(f.group[g]))
		for n, i := range f.group[g] {
			rows[n] = f.rows[i]
		}
		label = f.o.GroupLabel(g, rows)
	}
	open := !s.Collapsed[g]
	wrap.Children(func() {
		b := ui.ButtonBase(c.Key("g")).Label(label).Expanded(open).Width(f.total).Height(m.head).Shrink(0).
			Cursor(ui.CursorPointer).FocusRing(false)
		bg := Mix(cl.TableBg, cl.Text, 0.025)
		if b.Hovered() {
			bg = Mix(cl.TableBg, cl.Text, 0.045)
		}
		b.Background(bg).InsetShadow(0, -1, 0, 0, cl.Hairline)
		if b.FocusVisible() {
			b.InsetShadow(0, 0, 0, 1.5, cl.ColorAccent500.Alpha(0.55))
		}
		if b.Clicked() {
			if s.Collapsed == nil {
				s.Collapsed = map[string]bool{}
			}
			if open {
				s.Collapsed[g] = true
			} else {
				delete(s.Collapsed, g)
			}
		}
		ep := b.AnimateTo("open", b2f(open), springOpen)
		b.Children(func() {
			ui.Row(c).Width(viewW).Height(m.head).AlignItems(ui.Center).Gap(8).Padding(0, m.pad).
				Translate(s.scroll.X, 0).Children(func() {
				ui.Icon(c, iconChevronEnd).Size(12, 12).TextColor(cl.Muted).Rotate(90 * ep).Shrink(0)
				ui.Text(c, label).FontSize(FsSm).FontWeight(600).TextColor(cl.Text2).SingleLine().Ellipsis("…").Shrink(1).MinWidth(0)
				ui.Text(c, strconv.Itoa(len(f.group[g]))).FontSize(FsSm).TextColor(cl.Muted).FontFeatures("tnum").NoWrap()
			})
		})
	})
}

// rowLead is a row's leading cells: its checkbox and its chevron.
func (f *tableFrame[R]) rowLead(k any, n int, chosen bool, ep float32) {
	c, s := f.c, f.s
	cl := f.t.Colors
	if f.o.Selectable {
		ui.Box(c).Width(selectW).Shrink(0).Center().Children(func() {
			b := f.t.tableCheck(c.Key("sel"), chosen, false, f.l.SelectRow)
			if b.Clicked() {
				if b.ClickModifiers()&ui.Shift != 0 && s.anchor != nil {
					f.chooseRange(s.anchor, k, !chosen)
				} else {
					f.choose(k, !chosen)
				}
				s.anchor, s.cursor = k, n
			}
		})
	}
	if f.o.Expand != nil {
		ui.Box(c).Width(expandW).Shrink(0).Center().Children(func() {
			b := ui.ButtonBase(c.Key("open")).Label(f.l.Details).Expanded(s.Expanded[k]).Size(24, 24).Radius(RadiusPill).
				Center().Cursor(ui.CursorPointer).FocusRing(false)
			color := cl.Muted
			if b.Hovered() {
				color = cl.Text
			}
			if b.Pressed() {
				b.Scale(0.9, 0.9)
			}
			if b.FocusVisible() {
				f.t.Shadows.FocusRing.Apply(b)
			}
			b.Children(func() { ui.Icon(c, iconChevronEnd).Size(12, 12).TextColor(color).Rotate(90 * ep) })
			if b.Clicked() {
				f.toggleOpen(k)
				s.cursor = n
			}
		})
	}
}

// cell is row r's cell of column i: the column's own, or its kind's.
func (f *tableFrame[R]) cell(i int, r R) {
	c := f.c
	col := &f.cols[i]
	id := col.id()
	e := ui.Row(c.Key("c:"+id)).Width(f.widths[i]).Shrink(0).Padding(0, f.m.pad).AlignItems(ui.Center).Clip().Role(ui.RoleCell)
	if col.align() == ui.End {
		e.Justify(ui.End)
	}
	if d := &f.rt.col; d.active && d.id == id {
		// the column being moved goes with its header, faded
		e.Opacity(0.45)
	}
	e.Translate(f.shift[id], 0).Children(func() {
		switch {
		case col.Cell != nil:
			col.Cell(c, r)
		case col.Value != nil:
			f.t.cell(c, col.Kind, col.Value(r), col.Decimals, col.Unit, f.m.fs)
		}
	})
}

// details is row k's details under it, opening as far as ep: they grow
// into view and fade in, rising 4px, and keep to the table's view as it
// scrolls sideways.
func (f *tableFrame[R]) details(k any, r R, ep, viewW float32) {
	c, rt := f.c, f.rt
	cl := f.t.Colors
	clip := ui.Box(c).Width(f.total).Height(rt.detailH[k] * ep).Shrink(0).ClipY()
	clip.Children(func() {
		in := ui.Column(c).Width(viewW).Padding(8, 16, 16, f.lead+f.m.pad).Gap(8).FontSize(FsMd).TextColor(cl.Text2).
			Background(Mix(cl.TableBg, cl.Text, 0.02)).InsetShadow(0, -1, 0, 0, cl.Hairline).
			Translate(f.s.scroll.X, -4*(1-ep)).Opacity(ep)
		in.Children(func() { f.o.Expand(c, r) })
		if h := in.LayoutBounds().H; h > 0 {
			rt.detailH[k] = h
		}
	})
}

// springOpen is how a row's details open and close.
var springOpen = ui.Animation{Spring: ui.SpringOptions{Duration: 320 * time.Millisecond, Bounce: 0.04}}

// ghosts paints the rows just taken out, fading and settling where they
// were, over the rows closing the room they left.
func (f *tableFrame[R]) ghosts(moved bool, viewH float32) {
	c, rt := f.c, f.rt
	cl := f.t.Colors
	if moved {
		here := map[any]bool{}
		for _, k := range f.keys {
			here[k] = true
		}
		for _, k := range rt.prevKey {
			y, ok := rt.prevY[k]
			r, seen := rt.last[k]
			if here[k] || !ok || !seen || y < f.s.scroll.Y-f.m.row || y > f.s.scroll.Y+viewH {
				continue
			}
			rt.ghosts = append(rt.ghosts, tableGhost[R]{row: r, key: k, y: y, at: f.now, turn: rt.turn})
		}
	}
	// a page turned goes quicker, for the new page to come in clear
	life := func(g tableGhost[R]) time.Duration {
		if g.turn != 0 {
			return 150 * time.Millisecond
		}
		return 200 * time.Millisecond
	}
	rt.ghosts = slices.DeleteFunc(rt.ghosts, func(g tableGhost[R]) bool { return f.now.Sub(g.at) > life(g) })
	for _, g := range rt.ghosts {
		p := clamp01(float32(f.now.Sub(g.at)) / float32(life(g)))
		e := ui.Row(c.Key(ghostKey{g.key})).Absolute().Left(0).Top(g.y).Width(f.total).Height(f.m.row).
			AlignItems(ui.Stretch).PassThrough().Background(cl.TableBg).Opacity(1-p).Translate(0, 4*p)
		e.Children(func() {
			ui.Box(c).Width(f.lead).Shrink(0)
			for _, i := range f.order {
				f.cell(i, g.row)
			}
		})
	}
	if len(rt.ghosts) > 0 {
		c.AnimationFrame()
	}
}

// ghostKey keys a ghost row apart from the live row of the same key.
type ghostKey struct{ key any }

// placeholders fills the rows' place while they load: rows of skeleton
// bars, each column's a little shorter or longer than the last.
func (f *tableFrame[R]) placeholders(viewH float32) {
	c, m := f.c, f.m
	cl := f.t.Colors
	n := int(viewH/m.row) + 1
	ui.Column(c).Width(f.total).Shrink(0).Children(func() {
		for r := range n {
			ui.Row(c.Key(r)).Width(f.total).Height(m.row).Shrink(0).AlignItems(ui.Center).InsetShadow(0, -1, 0, 0, cl.Hairline).Children(func() {
				if f.o.Selectable {
					ui.Box(c).Width(selectW).Shrink(0).Center().Children(func() { Skeleton(c, 16, 16).Radius(min(f.t.R(5), 6)) })
				}
				if f.o.Expand != nil {
					ui.Box(c).Width(expandW).Shrink(0)
				}
				for j, i := range f.order {
					w := f.widths[i] - 2*m.pad
					frac := []float32{0.7, 0.45, 0.6, 0.8, 0.5}[(r+j)%5]
					row := ui.Row(c).Width(f.widths[i]).Shrink(0).Padding(0, m.pad)
					if f.cols[i].align() == ui.End {
						row.Justify(ui.End)
					}
					row.Children(func() { Skeleton(c, max(16, w*frac), 10) })
				}
			})
		}
	})
}

// state fills the table's view with an empty or error state.
func (f *tableFrame[R]) state(w, h float32, e EmptyStateOptions) {
	ui.Column(f.c).Width(w).Height(h).Center().Children(func() { EmptyState(f.c, e) })
}

// tableOrder is the visible columns' indexes, in the order of s.Order.
func tableOrder[R any](cols []Column[R], s *TableState) []int {
	byID := map[string]int{}
	for i := range cols {
		byID[cols[i].id()] = i
	}
	var order []int
	seen := map[int]bool{}
	for _, id := range s.Order {
		if i, ok := byID[id]; ok && !seen[i] {
			seen[i] = true
			order = append(order, i)
		}
	}
	for i := range cols {
		if !seen[i] {
			order = append(order, i)
		}
	}
	return slices.DeleteFunc(order, func(i int) bool { return s.Hidden[cols[i].id()] })
}

// fullOrder is every column's ID, in the order shown, hidden ones too.
func fullOrder[R any](cols []Column[R], s *TableState) []string {
	shown := *s
	shown.Hidden = nil
	order := tableOrder(cols, &shown)
	ids := make([]string, len(order))
	for k, i := range order {
		ids[k] = cols[i].id()
	}
	return ids
}

// tableWidths gives each visible column its width: the user's, the
// column's, or a share of the room left, no narrower than its least. It
// returns them by column index, the leading cells' width, and the rows'.
func tableWidths[R any](cols []Column[R], s *TableState, order []int, o TableOptions[R], avail float32) (map[int]float32, float32, float32) {
	var lead float32
	if o.Selectable {
		lead += selectW
	}
	if o.Expand != nil {
		lead += expandW
	}
	if avail <= 0 {
		avail = 900
	}
	widths := map[int]float32{}
	fixed, flexMin := lead, float32(0)
	var flex []int
	for _, i := range order {
		col := &cols[i]
		if w, ok := s.Widths[col.id()]; ok {
			widths[i] = w
		} else if col.Width > 0 {
			widths[i] = col.Width
		} else {
			flex = append(flex, i)
			flexMin += col.minWidth()
			continue
		}
		fixed += widths[i]
	}
	room := avail - fixed
	for _, i := range flex {
		w := cols[i].minWidth()
		if room > flexMin {
			w = room * w / flexMin
		}
		widths[i] = w
	}
	total := lead
	for _, i := range order {
		total += widths[i]
	}
	return widths, lead, max(total, avail)
}

// tableFadeW is how wide the fade at a sideways-scrolling table's edge is.
const tableFadeW = 40

// edgeFade is how much each edge of a table that scrolls sideways shows its
// fade: at the right while there is more to the right, at the left once it is
// scrolled, each coming and going on a short curve.
func (f *tableFrame[R]) edgeFade() (left, right float32) {
	sc := &f.s.scroll
	more := sc.MaxX > 1
	if !f.rt.ready || sc.MaxX != f.rt.fadeMax { // a frame builds before it is laid out: the range is known in the next
		f.rt.fadeMax = sc.MaxX
		f.c.Invalidate()
	}
	left = f.outer.AnimateTo("fadeL", b2f(more && sc.X > 1), ui.Animation{Duration: 160 * time.Millisecond, Ease: ui.EaseOut})
	right = f.outer.AnimateTo("fadeR", b2f(more && sc.X < sc.MaxX-1), ui.Animation{Duration: 160 * time.Millisecond, Ease: ui.EaseOut})
	return left, right
}

// paintEdgeFade draws the fades over the table, so that rows cut at its edge read
// as going on: the left one starts after the pinned column.
func (f *tableFrame[R]) paintEdgeFade(p *ui.Painter, r ui.Rect, left, right float32) {
	bg := f.t.Colors.TableBg
	if left > 0.01 {
		x := r.X
		if f.pinned > 0 && len(f.order) > 0 {
			x += f.lead + f.widths[f.order[0]]
		}
		p.FillGradient(ui.Rect{X: x, Y: r.Y, W: tableFadeW, H: r.H}, ui.LinearGradient{Angle: 90, Oklab: true,
			Stops: []ui.GradientStop{{Color: bg.Alpha(left), At: 0}, {Color: bg.Alpha(0), At: 1}}}, 0)
	}
	if right > 0.01 {
		p.FillGradient(ui.Rect{X: r.X + r.W - tableFadeW, Y: r.Y, W: tableFadeW, H: r.H}, ui.LinearGradient{Angle: 90, Oklab: true,
			Stops: []ui.GradientStop{{Color: bg.Alpha(0), At: 0}, {Color: bg.Alpha(right), At: 1}}}, 0)
	}
}
