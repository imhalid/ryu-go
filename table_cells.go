package ryugo

import (
	"cmp"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

// ColumnKind is how a Table column shows its values: the kind picks the
// cell's look, its alignment and how the column sorts.
type ColumnKind uint8

const (
	// KindText is a line of text, cut with an ellipsis.
	KindText ColumnKind = iota
	// KindNumber is a number with thousands separators, tabular, at the
	// end of the cell, with the column's Unit muted beside it.
	KindNumber
	// KindChange is a change in percent points as a toned badge: a rise
	// in the Down color, a fall in the Up color, as Ryugo reads money.
	KindChange
	// KindDate is a time.Time as "Nov 3, 2026".
	KindDate
	// KindStatus is a CellStatus: a StatusDot and its text.
	KindStatus
	// KindPerson is a name with its avatar.
	KindPerson
	// KindBar is a fraction from 0 to 1 as a short bar and its percent.
	KindBar
)

// CellStatus is the value of a KindStatus column.
type CellStatus struct {
	State DotState
	Text  string
}

// Column is a column of a Table showing rows of type R.
type Column[R any] struct {
	// ID identifies the column in the TableState (order, widths, sort,
	// hidden); the Title when empty.
	ID, Title string
	Kind      ColumnKind
	// Value is the row's value in this column, for the kind to show and
	// for sorting: a string, a number, a time.Time, a CellStatus.
	Value func(r R) any
	// Cell, if set, builds the cell's content instead of the kind.
	Cell func(c *ui.Context, r R)
	// Width is the column's width; 0 shares the room the others leave, no
	// narrower than MinWidth (120 by default). The user resizes a column
	// by the edge of its header, which then keeps its width.
	Width, MinWidth float32
	// Align aligns the cells; by default numbers, changes and bars go to
	// the end, the rest to the start.
	Align ui.Align
	// Sortable lets a click on the header sort the rows by Value.
	Sortable bool
	// Decimals of a KindNumber or KindChange (KindChange: 1 by default).
	Decimals int
	// Unit follows a KindNumber's value, muted ("kg", "ms").
	Unit string
	// Fixed keeps the column where it is: the user neither moves nor
	// resizes it.
	Fixed bool
}

func (col *Column[R]) id() string { return or(col.ID, col.Title) }

// align is where the column's cells go.
func (col *Column[R]) align() ui.Align {
	if col.Align != 0 {
		return col.Align
	}
	switch col.Kind {
	case KindNumber, KindChange, KindBar:
		return ui.End
	}
	return ui.Start
}

func (col *Column[R]) minWidth() float32 { return or(col.MinWidth, 120) }

// cell builds a cell's content by the column's kind.
func (t *Tokens) cell(c *ui.Context, kind ColumnKind, v any, decimals int, unit string, fs float32) {
	cl := t.Colors
	text := func(s string) ui.Element {
		return ui.Text(c, s).FontSize(fs).TextColor(cl.Text).SingleLine().Ellipsis("…")
	}
	switch kind {
	case KindNumber:
		f, ok := number(v)
		if !ok {
			return
		}
		ui.Row(c).AlignItems(ui.Center).Gap(4).Shrink(1).MinWidth(0).Children(func() {
			text(grouped(f, decimals)).FontFeatures("tnum")
			if unit != "" {
				ui.Text(c, unit).FontSize(fs - 1).TextColor(cl.Muted).NoWrap()
			}
		})
	case KindChange:
		f, ok := number(v)
		if !ok {
			return
		}
		d := decimals
		if d == 0 {
			d = 1
		}
		s := strconv.FormatFloat(math.Abs(f), 'f', d, 64) + "%"
		tone := ToneNeutral
		switch {
		case f > 0:
			s, tone = "+"+s, ToneDown
		case f < 0:
			s, tone = "−"+s, ToneUp
		}
		Badge(c, s, tone)
	case KindDate:
		if tm, ok := v.(time.Time); ok && !tm.IsZero() {
			text(tm.Format("Jan 2, 2006")).FontFeatures("tnum")
		}
	case KindStatus:
		if st, ok := v.(CellStatus); ok {
			StatusDot(c, st.State, st.Text)
		}
	case KindPerson:
		if name, ok := v.(string); ok && name != "" {
			ui.Row(c).AlignItems(ui.Center).Gap(8).MinWidth(0).Children(func() {
				Avatar(c, name, AvatarOptions{Size: 22})
				text(name)
			})
		}
	case KindBar:
		f, ok := number(v)
		if !ok {
			return
		}
		f = math.Max(0, math.Min(1, f))
		ui.Row(c).AlignItems(ui.Center).Gap(8).Children(func() {
			well := ui.Box(c).Size(56, 6).Radius(t.Pill(6)).Background(cl.Well).Clip()
			t.Shadows.WellEdge.Apply(well)
			well.Children(func() {
				ui.Box(c).Absolute().Left(0).Top(0).Bottom(0).WidthPercent(float32(f * 100)).Radius(t.Pill(6)).LinearGradient(t.accentFill())
			})
			ui.Text(c, strconv.Itoa(int(f*100+0.5))+"%").FontSize(fs - 1).TextColor(cl.Text2).FontFeatures("tnum").NoWrap().MinWidth(32).TextAlign(ui.End)
		})
	default:
		switch x := v.(type) {
		case nil:
		case string:
			text(x)
		case CellStatus:
			text(x.Text)
		default:
			if f, ok := number(v); ok {
				text(grouped(f, decimals)).FontFeatures("tnum")
			}
		}
	}
}

// number reads v as a float64, for numbers of any type.
func number(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case int32:
		return float64(x), true
	case uint:
		return float64(x), true
	case uint64:
		return float64(x), true
	}
	return 0, false
}

// grouped writes f with d decimals and commas between groups of three
// digits: 1234567.5 → "1,234,567.5".
func grouped(f float64, d int) string {
	s := strconv.FormatFloat(math.Abs(f), 'f', d, 64)
	whole, frac, _ := strings.Cut(s, ".")
	var b strings.Builder
	if f < 0 {
		b.WriteString("−")
	}
	for i, r := range whole {
		if i > 0 && (len(whole)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(r)
	}
	if frac != "" {
		b.WriteByte('.')
		b.WriteString(frac)
	}
	return b.String()
}

// compareValues orders two values of a column: numbers by size, times in
// time, statuses and text by their text, ignoring case; nothing last.
func compareValues(a, b any) int {
	if a == nil || b == nil {
		switch {
		case a == nil && b == nil:
			return 0
		case a == nil:
			return 1
		}
		return -1
	}
	if fa, ok := number(a); ok {
		if fb, ok := number(b); ok {
			return cmp.Compare(fa, fb)
		}
	}
	if ta, ok := a.(time.Time); ok {
		if tb, ok := b.(time.Time); ok {
			return ta.Compare(tb)
		}
	}
	str := func(v any) string {
		switch x := v.(type) {
		case string:
			return x
		case CellStatus:
			return x.Text
		case bool:
			return strconv.FormatBool(x)
		}
		return ""
	}
	return cmp.Compare(strings.ToLower(str(a)), strings.ToLower(str(b)))
}
