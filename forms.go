package mizan

import (
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
)

// The form fields beyond Input and Select: the ones you type into are
// wells, as Input is (Textarea, NumberInput, Combobox, TagInput); the one
// you only open is raised, as Select's trigger is (DatePicker). Their
// lists and calendars float on Select's frosted pop surface.

var (
	iconPlus     = Svg(16, 12, "M8 3.5v9M3.5 8h9")
	iconMinus    = Svg(16, 12, "M3.5 8h9")
	iconPrev     = Svg(16, 14, "M10 3.5L5.5 8l4.5 4.5")
	iconNext     = Svg(16, 14, "M6 3.5l4.5 4.5L6 12.5")
	iconCalendar = Svg(16, 15, "M4.5 3h7A2 2 0 0 1 13.5 5v6.5a2 2 0 0 1-2 2h-7a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2zM2.5 6.5h11M5.5 1.5v3M10.5 1.5v3")
)

// fieldFs is the text size of a field one types into: InputFs, body text at md.
func fieldFs(m Metrics) float32 {
	if m.InputFs == 0 {
		return FsBase
	}
	return m.InputFs
}

// TextareaOptions are the options of Textarea.
type TextareaOptions struct {
	Placeholder string
	// Rows is how many lines high the field is while short (3 if unset);
	// it grows with its text up to MaxRows (8 if unset), then scrolls.
	Rows, MaxRows int
	// MaxLength, if set, shows the count of characters against it; past it
	// the count and the field turn the Up color. The text is not cut.
	MaxLength int
	// Count shows the count of characters without a limit.
	Count    bool
	Label    string
	Size     Size
	Disabled bool
	Invalid  bool
}

// Textarea is a multi-line text field: Input's inset well, as high as
// its text from Rows lines up to MaxRows. With MaxLength or Count, the
// count of characters stands in its bottom right corner. It returns the
// text area itself (for Changed, Focus).
//
//	mizan.Textarea(c.Key("bio"), &app.bio, mizan.TextareaOptions{Placeholder: "About you", MaxLength: 280})
func Textarea(c *ui.Context, value *string, o TextareaOptions) ui.Element {
	t := sizedFrom(c, o.Size)
	m := t.Metrics
	cl := t.Colors
	fs := fieldFs(m)
	rows := or(o.Rows, 3)
	maxRows := max(or(o.MaxRows, 8), rows)
	n := utf8.RuneCountInString(*value)
	over := o.MaxLength > 0 && n > o.MaxLength

	// the first line sits where an Input's text does
	py := max(6, (m.FieldH-fs*1.4)/2)
	box := ui.Column(c).Radius(m.FieldR).Padding(py, m.FieldPs).Gap(2).MinWidth(0).Cursor(ui.CursorText)
	var area ui.Element
	box.Children(func() {
		area = ui.TextAreaBase(c, value).FillWidth().Lines(rows, maxRows).
			Placeholder(o.Placeholder).PlaceholderColor(cl.Muted).
			TextColor(cl.Text).FontSize(fs).Disabled(o.Disabled).FocusRing(false)
		if o.Label != "" {
			area.Label(o.Label)
		}
		if o.MaxLength > 0 || o.Count {
			s := strconv.Itoa(n)
			if o.MaxLength > 0 {
				s += " / " + strconv.Itoa(o.MaxLength)
			}
			color := cl.Muted
			if over {
				color = cl.Up
			}
			ui.Text(c, s).AlignSelf(ui.End).FontSize(FsXs).TextColor(color).FontFeatures("tnum").NoWrap().Unselectable()
		}
	})
	t.well(box, area.Focused(), o.Invalid || over)
	if o.Disabled {
		box.Opacity(0.4).Cursor(ui.CursorDefault)
	} else if box.Clicked() {
		area.Focus()
	}
	return area
}

// NumberInputOptions are the options of NumberInput.
type NumberInputOptions struct {
	// Min and Max bound the value; when both are zero it has no bounds.
	Min, Max float64
	// Step is how far the buttons and the arrows move the value (1 if
	// unset); the value shows with as many decimals as Step has.
	Step float64
	// Prefix and Suffix are muted units around the digits ("$", "kg").
	Prefix, Suffix string
	Label          string
	Size           Size
	Disabled       bool
	Invalid        bool
	// Error, if set, is given what is wrong with the text being typed, as
	// "At most 10", for a Field around the input to show; "" while it is
	// a number the input takes.
	Error *string
}

// NumberInput is a number field: Input's inset well holding bold tabular
// digits, with the raised − / + stepper on its right, the two halves of
// one small button. The buttons and Up and Down step the value by Step
// (ten steps with Shift), within Min and Max; a button held steps on and
// on, faster and faster.
//
// Only a number can be typed: digits, one decimal point (a comma is
// taken as one) with no more decimals than Step has, and a leading minus
// when Min is below zero; anything else is left out as it is typed or
// pasted. What is typed applies as soon as it is a number in range on a
// step; anything else marks the field invalid, says why in Error, and is
// settled as the field loses the focus or Enter is pressed: brought into
// range, to the nearest step, or, when it is no number, back to the
// value. It returns the text input (for Changed, Focus).
//
//	mizan.NumberInput(c.Key("qty"), &app.qty, mizan.NumberInputOptions{Min: 1, Max: 99, Label: "Quantity"})
func NumberInput(c *ui.Context, value *float64, o NumberInputOptions) ui.Element {
	t := sizedFrom(c, o.Size)
	m := t.Metrics
	cl := t.Colors
	step := o.Step
	if step <= 0 {
		step = 1
	}
	lo, hi := o.Min, o.Max
	bounded := lo != 0 || hi != 0
	if !bounded {
		lo, hi = math.Inf(-1), math.Inf(1)
	}
	decimals := 0
	if s := strconv.FormatFloat(step, 'f', -1, 64); strings.Contains(s, ".") {
		decimals = len(s) - strings.IndexByte(s, '.') - 1
	}
	format := func(v float64) string { return strconv.FormatFloat(v, 'f', decimals, 64) }
	round := func(v float64) float64 {
		p := math.Pow(10, float64(decimals))
		return math.Round(v*p) / p
	}
	clamp := func(v float64) float64 { return math.Max(lo, math.Min(hi, round(v))) }
	// the steps count from Min, or from zero without bounds
	base := 0.0
	if bounded {
		base = lo
	}
	onStep := func(v float64) bool {
		n := (v - base) / step
		return math.Abs(n-math.Round(n)) < 1e-9
	}
	snap := func(v float64) float64 { return clamp(base + math.Round((v-base)/step)*step) }

	box := ui.Row(c).AlignItems(ui.Stretch).Height(m.NumH).Gap(4).Padding(4, 4, 4, m.NumPs).
		Radius(m.NumR).MinWidth(0).Cursor(ui.CursorText)
	text := ui.Local(box, "text", func() string { return format(*value) })
	focused := ui.Local(box, "focused", func() bool { return false })
	stepBy := func(n float64) {
		*value = clamp(*value + n*step)
		*text = format(*value)
	}
	// settle brings what was typed to a value the input takes
	settle := func() {
		if v, err := strconv.ParseFloat(*text, 64); err == nil {
			*value = snap(v)
		}
		*text = format(*value)
	}
	var field ui.Element
	msg := ""
	box.Children(func() {
		ui.Row(c).Grow(1).MinWidth(0).AlignItems(ui.Center).Gap(m.InputGap).Children(func() {
			if o.Prefix != "" {
				unit(c, o.Prefix, cl, m.NumPrefixFs)
			}
			field = ui.TextInputBase(c, text).Grow(1).MinWidth(0).
				TextColor(cl.Text).FontSize(m.NumFs).FontWeight(600).LetterSpacing(m.NumLs * m.NumFs).
				FontFeatures("tnum").Disabled(o.Disabled).FocusRing(false)
			if o.Label != "" {
				field.Label(o.Label)
			}
			if bounded {
				field.Range(lo, hi, *value).Step(step)
			}
			switch {
			case field.Shortcut(0, ui.KeyUp):
				stepBy(1)
			case field.Shortcut(0, ui.KeyDown):
				stepBy(-1)
			case field.Shortcut(ui.Shift, ui.KeyUp):
				stepBy(10)
			case field.Shortcut(ui.Shift, ui.KeyDown):
				stepBy(-10)
			}
			if field.Changed() {
				at, _ := field.TextSelection()
				if s, caret := numberText(*text, at, lo < 0, decimals); s != *text {
					*text = s
					field.SetTextSelection(caret, caret)
				}
				if v, err := strconv.ParseFloat(*text, 64); err == nil && v >= lo && v <= hi && onStep(v) {
					*value = round(v)
				}
			}
			switch {
			case field.Focused() && field.Submitted():
				settle()
			case field.Focused():
				v, err := strconv.ParseFloat(*text, 64)
				switch {
				case err != nil:
					msg = "Enter a number"
				case v < lo:
					msg = "At least " + format(lo)
				case v > hi:
					msg = "At most " + format(hi)
				case !onStep(v):
					msg = "In steps of " + strconv.FormatFloat(step, 'f', -1, 64)
				}
			case *focused:
				settle()
			default:
				// what the app set, shown in full
				*text = format(*value)
			}
			*focused = field.Focused()
			if o.Suffix != "" {
				unit(c, o.Suffix, cl, m.NumPrefixFs)
			}
		})
		// the stepper: one raised button cut in two, + above −
		// 4px inside the field, concentric with it
		st := ui.Column(c).Width(m.NumStepW).Shrink(0).Radius(Concentric(m.NumR, 4)).Clip().
			LinearGradient(t.controlGradient(cl.ControlTop, cl.ControlBottom)).Border(0.5, cl.ControlBorder)
		t.Shadows.ShadowStepper.Apply(st)
		st.Children(func() {
			for i, b := range []struct {
				icon  *ui.SVG
				label string
				dir   float64
			}{{iconPlus, "Increase", 1}, {iconMinus, "Decrease", -1}} {
				off := o.Disabled || b.dir > 0 && *value >= hi || b.dir < 0 && *value <= lo
				// the buttons take no focus: the arrows step from the field
				e := ui.Box(c.Key(i)).Grow(1).Basis(0).Center().Role(ui.RoleButton).Label(b.label).
					KeepFocus().Cursor(ui.CursorPointer).Disabled(off)
				if i > 0 {
					// the divider is a shadow, not a border: both halves stay the same height
					e.InsetShadow(0, 0.5, 0, 0, cl.ControlBorder)
				}
				switch {
				case off:
					e.Opacity(0.35).Cursor(ui.CursorDefault)
				case e.Pressed():
					e.Background(Mix(ui.Color{}, cl.Text, 0.1))
				case e.Hovered():
					e.Background(cl.Fill)
				}
				e.TextColor(cl.Text).Transition(ui.ElementTransition{Duration: 150 * time.Millisecond, Colors: true})
				e.Children(func() { ui.Icon(c, b.icon).Size(10, 10) })
				hold(c, e, off, func() { stepBy(b.dir) })
			}
		})
	})
	field.Error(msg)
	if o.Error != nil {
		*o.Error = msg
	}
	t.well(box, field.Focused(), o.Invalid || msg != "")
	if o.Disabled {
		box.Opacity(0.4).Cursor(ui.CursorDefault)
	} else if box.Clicked() {
		field.Focus()
	}
	return field
}

// numberText keeps of s, typed into a NumberInput, what makes a number:
// digits, a minus first when neg, one decimal point (a comma is one) with
// no more than decimals digits after it. It returns the text and where
// the caret, at at in s, is in it.
func numberText(s string, at int, neg bool, decimals int) (string, int) {
	var b strings.Builder
	caret, n, point, after := 0, 0, false, 0
	for i, r := range []rune(s) {
		keep := true
		switch {
		case r >= '0' && r <= '9':
			if point {
				keep = after < decimals
				after++
			}
		case r == '.' || r == ',':
			keep = decimals > 0 && !point
			point, r = point || keep, '.'
		case r == '-' || r == '−':
			keep = neg && n == 0
			r = '-'
		default:
			keep = false
		}
		if keep {
			b.WriteRune(r)
			n++
		}
		if i < at {
			caret = n
		}
	}
	return b.String(), caret
}

// unit is a NumberInput's muted unit, at the stepper's size step.
func unit(c *ui.Context, s string, cl Colors, fs float32) {
	ui.Text(c, s).TextColor(cl.Muted).FontSize(fs).NoWrap().Unselectable()
}

// ComboboxOptions are the options of Combobox.
type ComboboxOptions struct {
	Placeholder string
	Label       string
	Size        Size
	Disabled    bool
	Invalid     bool
}

// Combobox is a Select one can type into: the field is Input's inset well
// showing the chosen option's label, and typing filters the options, those
// whose label starts with the text first; a click, Down or the chevron
// shows them all. The list is Select's: the pop surface, the accent
// gradient on the highlighted option, a check mark by the chosen one.
// Up and Down move, Enter or a click chooses, Escape closes; text that
// matched nothing goes back to the choice as the field loses the focus.
// It returns the text input.
//
//	mizan.Combobox(c.Key("owner"), &app.owner, people, mizan.ComboboxOptions{Placeholder: "Assign to…"})
func Combobox[T comparable](c *ui.Context, value *T, options []Option[T], o ComboboxOptions) ui.Element {
	t := sizedFrom(c, o.Size)
	m := t.Metrics
	cl := t.Colors
	current := -1
	for i, op := range options {
		if op.Value == *value {
			current = i
		}
	}
	label := ""
	if current >= 0 {
		label = options[current].Label
	}

	// the input is the well itself, so that the list lines up with it
	wrap := ui.Box(c).MinWidth(0)
	text := ui.Local(wrap, "text", func() string { return label })
	wasOpen := ui.Local(wrap, "open", func() bool { return false })
	var cb ui.ComboboxParts
	wrap.Children(func() {
		cb = ui.ComboboxBase(c, text)
		in := cb.Input
		ps := m.FieldPs
		showDot := current >= 0 && options[current].Dot != (ui.Color{}) && *text == label
		if showDot {
			ps += 8 + m.InputGap
		}
		in.FillWidth().Height(m.FieldH).Padding(0, m.FieldPe+m.InputBtn+2, 0, ps).Radius(m.FieldR).
			Placeholder(o.Placeholder).PlaceholderColor(cl.Muted).TextColor(cl.Text).FontSize(fieldFs(m)).
			Disabled(o.Disabled).FocusRing(false)
		if o.Label != "" {
			in.Label(o.Label)
		}
		t.well(in, in.Focused(), o.Invalid)
		if o.Disabled {
			in.Opacity(0.4)
		}
		if showDot {
			ui.Box(c).Absolute().Left(m.FieldPs).Top(0).Bottom(0).Center().PassThrough().Children(func() {
				dot(c, options[current].Dot)
			})
		}
		// the chevron toggles the whole list, keeping the focus in the input
		ch := ui.Box(c).Absolute().Right(m.FieldPe-4).Top((m.FieldH-m.InputBtn)/2).Size(m.InputBtn, m.InputBtn).
			Center().Radius(RadiusPill).KeepFocus().Cursor(ui.CursorPointer).TextColor(cl.Muted)
		if o.Disabled {
			ch.Opacity(0.4).PassThrough()
		} else if ch.Hovered() {
			ch.Background(cl.Fill).TextColor(cl.Text)
		}
		ch.Children(func() { ui.Icon(c, iconChevrons) })
		if ch.Clicked() && !o.Disabled {
			cb.SetOpen(!cb.Open())
			in.Focus()
			if cb.Open() {
				in.SetTextSelection(0, utf8.RuneCountInString(*text))
			}
		}
		// a click opening the list selects the text, so that typing replaces it
		if in.Clicked() && cb.Open() && !*wasOpen && !cb.Filtering() {
			in.SetTextSelection(0, utf8.RuneCountInString(*text))
		}
	})

	q := ""
	if cb.Filtering() {
		q = *text
	}
	shown := matchOptions(options, q)
	cb.Popup(func(panel ui.Element) {
		popover(t, panel)
		panel.Gap(0)
		ui.Scroll(c).MaxHeight(300 - 2*PadPop).Children(func() {
			for _, i := range shown {
				op := options[i]
				item := cb.Item(strconv.Itoa(i)).Disabled(op.Disabled)
				optionItem(c, t, item, op, i == current)
			}
		})
	})
	if v, ok := cb.Chosen(); ok {
		if i, err := strconv.Atoi(v); err == nil && i < len(options) && !options[i].Disabled {
			*value = options[i].Value
			*text = options[i].Label
			cb.Input.SetTextSelection(0, utf8.RuneCountInString(*text))
		}
	} else if !cb.Input.Focused() && !cb.Open() && *text != label {
		// what was typed matched no option, or the app chose another
		*text = label
	}
	*wasOpen = cb.Open()
	return cb.Input
}

// matchOptions returns the indexes of the options whose label contains q,
// ignoring case, those starting with it first; all of them for a blank q.
func matchOptions[T comparable](options []Option[T], q string) []int {
	q = strings.ToLower(strings.TrimSpace(q))
	var starts, contains []int
	for i, op := range options {
		switch l := strings.ToLower(op.Label); {
		case q == "" || strings.HasPrefix(l, q):
			starts = append(starts, i)
		case strings.Contains(l, q):
			contains = append(contains, i)
		}
	}
	return append(starts, contains...)
}

// optionItem styles an option of a pop-up list as Select's: the accent
// gradient while highlighted, a check mark by the chosen one, the
// series-color dot and the muted description.
func optionItem[T comparable](c *ui.Context, t *Tokens, item ui.Element, op Option[T], chosen bool) {
	cl := t.Colors
	item.AlignItems(ui.Center).Gap(6).Padding(6, 14, 6, 8).Radius(Concentric(t.R(RadiusPop), PadPop)).FontSize(FsMd).Cursor(ui.CursorPointer)
	hl := item.Highlighted() && !op.Disabled
	muted, check := cl.Muted, cl.Accent
	if hl {
		item.LinearGradient(t.accentFill()).TextColor(cl.OnAccent)
		muted, check = cl.OnAccentMuted, cl.OnAccentMuted
	} else {
		item.TextColor(cl.Text)
	}
	if op.Disabled {
		item.Opacity(0.35).Cursor(ui.CursorDefault)
	}
	item.Children(func() {
		checkMark(c, 14, IconStroke, chosen, false, true, check)
		if op.Dot != (ui.Color{}) {
			d := dot(c, op.Dot)
			if hl {
				d.Shadow(0, 0, 0, 1.5, cl.OnAccentMuted)
			}
		}
		ui.Text(c, op.Label).NoWrap()
		if op.Description != "" {
			ui.Text(c, op.Description).TextColor(muted).FontSize(FsSm).NoWrap()
		}
	})
}

// DatePickerOptions are the options of DatePicker.
type DatePickerOptions struct {
	// Placeholder is the muted text shown while no date is chosen (the
	// date is zero); "Pick a date" if unset.
	Placeholder string
	// Min and Max, if set, are the first and the last day one can choose.
	Min, Max time.Time
	// Format writes the chosen date in the field; the user's regional
	// format of medium length ("Mar 14, 2026") if unset.
	Format   func(time.Time) string
	Label    string
	Size     Size
	Disabled bool
	Invalid  bool
}

// DatePicker is a date field: a raised trigger, as Select's, that opens a
// month's calendar on the pop surface below it. The chosen day fills with
// the accent gradient and its sheen, as a checked Checkbox; today is
// written in the accent color, the days of the months around muted. A
// click, or the arrows and Enter, choose a day, which closes it; Page Up
// and Page Down, or the buttons by the title, move by months, Home and
// End to the month's first and last day, and Escape closes it. The
// calendar is in the window's language, its weeks starting on the first
// day of the user's regional formats. A chosen day keeps the time of day
// and location of *date. It returns the trigger.
//
//	mizan.DatePicker(c.Key("due"), &app.due, mizan.DatePickerOptions{Label: "Due date"})
func DatePicker(c *ui.Context, date *time.Time, o DatePickerOptions) ui.Element {
	t := sizedFrom(c, o.Size)
	m := t.Metrics
	cl := t.Colors
	region, locale := formats(c)

	tr := ui.ButtonBase(c).Role(ui.RolePopUpButton).FillWidth().Height(m.FieldH).Padding(0, m.FieldPe, 0, m.FieldPs).
		Radius(m.FieldR).Justify(ui.SpaceBetween).AlignItems(ui.Center).Gap(6).Cursor(ui.CursorPointer).Disabled(o.Disabled)
	if o.Label != "" {
		tr.Label(o.Label)
	}
	open := ui.Local(tr, "open", func() bool { return false })
	// cursor is the day the keys are on, in the month the calendar shows
	cursor := ui.Local(tr, "cursor", func() time.Time { return time.Time{} })
	if tr.Clicked() && !o.Disabled {
		*open = !*open
		*cursor = *date
		if date.IsZero() {
			*cursor = c.Now()
		}
	}
	tr.Expanded(*open)
	s := raisedState{hovered: tr.Hovered() && !*open, focus: tr.FocusVisible(), disabled: o.Disabled}
	t.raised(tr, t.controlGradient(cl.ControlTop, cl.ControlBottom), cl.ControlBorder, t.Shadows.ControlShadow, s)
	switch {
	case o.Invalid:
		tr.BorderColor(cl.Up)
	case *open:
		tr.BorderColor(cl.ColorAccent500)
		t.Shadows.Ring.Apply(tr)
	}
	if tr.Pressed() && !o.Disabled {
		tr.Scale(0.99, 0.99)
	}
	text := ""
	if !date.IsZero() {
		if o.Format != nil {
			text = o.Format(*date)
		} else {
			text = i18n.FormatDate(locale, *date, i18n.DateOptions{Date: i18n.MediumStyle})
		}
		tr.Value(text)
	}
	tr.FontSize(m.FieldFs).FontWeight(500).TextColor(cl.Text).Children(func() {
		if text == "" {
			ui.Text(c, or(o.Placeholder, "Pick a date")).TextColor(cl.Muted).FontWeight(400).SingleLine().Ellipsis("…").Shrink(1)
		} else {
			ui.Text(c, text).SingleLine().Ellipsis("…").Shrink(1).FontFeatures("tnum")
		}
		ui.Icon(c, iconCalendar).TextColor(cl.Muted)
	})

	ui.PopoverBase(c, tr, open, func(panel ui.Element) {
		popover(t, panel)
		panel.Width(288).Padding(10)
		t.calendar(c, panel, date, cursor, o, region, locale, func(d time.Time) {
			*date = d
			*open = false
			tr.Focus()
		})
	})
	return tr
}

// formats returns the user's regional formats and the locale dates are
// written in, as MyGo's own date controls find them.
func formats(c *ui.Context) (i18n.Region, string) {
	lang := or(c.Locale(), "en")
	r := c.Region()
	if r.HourCycle != 12 && r.HourCycle != 24 {
		// the system named none: the language's
		r = i18n.RegionFor(lang)
		r.Locale = ""
	}
	return r, r.FormatLocale(lang)
}

// calendar builds a month's grid in panel: the title and the month
// buttons, the weekdays, and six weeks of days around *cursor. choose
// gets the day chosen, at *date's time of day.
func (t *Tokens) calendar(c *ui.Context, panel ui.Element, date, cursor *time.Time, o DatePickerOptions, region i18n.Region, locale string, choose func(time.Time)) {
	cl := t.Colors
	allowed := func(d time.Time) bool {
		return (o.Min.IsZero() || !d.Before(dayOf(o.Min, d.Location()))) &&
			(o.Max.IsZero() || !d.After(dayOf(o.Max, d.Location())))
	}
	pick := func(d time.Time) {
		if !allowed(d) {
			return
		}
		loc, clock := time.Local, time.Duration(0)
		if !date.IsZero() {
			loc = date.Location()
			clock = date.Sub(dayOf(*date, loc))
		}
		y, mo, dd := d.Date()
		choose(time.Date(y, mo, dd, 0, 0, 0, 0, loc).Add(clock))
	}

	panel.Children(func() {
		// one stop of Tab: the keys move a cursor among the days
		grid := ui.Column(c).Focusable().AutoFocus().Role(ui.RoleTable).FocusRing(false)
		cur := *cursor
		keys := [...]bool{
			grid.Shortcut(0, ui.KeyLeft), grid.Shortcut(0, ui.KeyRight),
			grid.Shortcut(0, ui.KeyUp), grid.Shortcut(0, ui.KeyDown),
			grid.Shortcut(0, ui.KeyPageUp), grid.Shortcut(0, ui.KeyPageDown),
			grid.Shortcut(0, ui.KeyHome), grid.Shortcut(0, ui.KeyEnd),
			grid.Shortcut(0, ui.KeyEnter) || grid.Shortcut(0, ui.KeySpace),
		}
		switch {
		case keys[0]:
			cur = cur.AddDate(0, 0, -1)
		case keys[1]:
			cur = cur.AddDate(0, 0, 1)
		case keys[2]:
			cur = cur.AddDate(0, 0, -7)
		case keys[3]:
			cur = cur.AddDate(0, 0, 7)
		case keys[4]:
			cur = addMonths(cur, -1)
		case keys[5]:
			cur = addMonths(cur, 1)
		case keys[6]:
			cur = time.Date(cur.Year(), cur.Month(), 1, 0, 0, 0, 0, cur.Location())
		case keys[7]:
			cur = time.Date(cur.Year(), cur.Month()+1, 0, 0, 0, 0, 0, cur.Location())
		case keys[8]:
			pick(cur)
		}
		*cursor = cur
		month := time.Date(cur.Year(), cur.Month(), 1, 0, 0, 0, 0, cur.Location())
		title := i18n.FormatDate(locale, month, i18n.DateOptions{Skeleton: "yMMMM"})
		grid.Label(title)
		focus := grid.FocusVisible()
		grid.Children(func() {
			ui.Row(c).AlignItems(ui.Center).Justify(ui.SpaceBetween).Margin(0, 0, 8, 0).Children(func() {
				ui.Text(c, title).FontSize(FsBase).FontWeight(600).LetterSpacing(-0.02*FsBase).Padding(0, 8).NoWrap()
				ui.Row(c).Gap(6).Children(func() {
					if ghostButton(c.Key("prev"), t, iconPrev, "Previous month", 30).KeepFocus().Clicked() {
						*cursor = addMonths(cur, -1)
					}
					if ghostButton(c.Key("next"), t, iconNext, "Next month", 30).KeepFocus().Clicked() {
						*cursor = addMonths(cur, 1)
					}
				})
			})
			names := i18n.Weekdays(locale, i18n.Short, true)
			first := int(region.FirstWeekday)
			ui.Row(c).Height(22).AlignItems(ui.Center).Children(func() {
				for i := range 7 {
					ui.Text(c, names[(first+i)%7]).Grow(1).Basis(0).TextAlign(ui.Center).
						FontSize(Fs2xs).FontWeight(500).TextColor(cl.Muted).Role(ui.RoleColumnHeader)
				}
			})
			// six weeks, from the first day of the week on or before the 1st
			start := month.AddDate(0, 0, -((int(month.Weekday()) - first + 7) % 7))
			today := c.Now()
			ui.Column(c).Gap(2).Children(func() {
				for w := range 6 {
					ui.Row(c).Role(ui.RoleRow).Children(func() {
						for d := range 7 {
							day := start.AddDate(0, 0, w*7+d)
							ui.Box(c).Grow(1).Basis(0).Center().Children(func() {
								cell := t.day(c, dayState{
									chosen:  !date.IsZero() && sameDay(day, *date),
									today:   sameDay(day, today),
									outside: day.Month() != month.Month(),
									cursor:  focus && sameDay(day, cur),
									off:     !allowed(day),
									label:   i18n.FormatDate(locale, day, i18n.DateOptions{Skeleton: "yMMMMd"}),
									number:  i18n.FormatNumber(locale, day.Day(), i18n.NumberOptions{}),
								})
								if sameDay(day, cur) {
									// assistive technology reads the day the keys are on
									grid.ActiveDescendant(cell)
								}
								if cell.Clicked() {
									*cursor = day
									pick(day)
								}
							})
						}
					})
				}
			})
		})
	})
}

// dayState is how a calendar's day shows.
type dayState struct {
	chosen, today, outside, cursor, off bool
	label, number                       string
}

// day is a calendar's 36px round day: the accent gradient and its sheen
// when chosen, the accent color today, muted outside the month; a fill
// while hovered, and a ring where the keys are.
func (t *Tokens) day(c *ui.Context, s dayState) ui.Element {
	cl := t.Colors
	e := ui.Box(c).Size(36, 36).Center().Radius(RadiusPill).Role(ui.RoleButton).Label(s.label).
		Selected(s.chosen).Cursor(ui.CursorPointer).Disabled(s.off)
	fg, weight := cl.Text, 400
	switch {
	case s.chosen:
		e.LinearGradient(t.accentFill())
		t.Shadows.SheenAccent.With(ShadowLayer{Y: 1, Blur: 2, Color: Mix(ui.Color{}, cl.ColorAccent700, 0.4)}).Apply(e)
		fg, weight = cl.OnAccent, 600
	case s.today:
		fg, weight = cl.Accent, 600
	case s.outside:
		fg = cl.Muted
	}
	if !s.chosen && !s.off && e.Hovered() {
		e.Background(cl.Fill)
	}
	if s.cursor {
		e.InsetShadow(0, 0, 0, 2, Mix(ui.Color{}, cl.ColorAccent500, 0.6))
	}
	if s.off {
		e.Opacity(0.3).Cursor(ui.CursorDefault)
	} else if e.Pressed() {
		e.Scale(0.94, 0.94)
	}
	e.Transition(ui.ElementTransition{Duration: 150 * time.Millisecond, Colors: true, Ease: ui.EaseOut})
	e.Children(func() {
		ui.Text(c, s.number).TextColor(fg).FontSize(FsMd).FontWeight(weight).FontFeatures("tnum").Unselectable()
	})
	return e
}

// sameDay reports whether a and b fall on the same date.
func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

// dayOf returns the start of d's date in loc.
func dayOf(d time.Time, loc *time.Location) time.Time {
	y, m, dd := d.Date()
	return time.Date(y, m, dd, 0, 0, 0, 0, loc)
}

// addMonths moves d by n months, to the last day of the month it lands in
// when that is shorter (January 31 and one month is February 28).
func addMonths(d time.Time, n int) time.Time {
	first := time.Date(d.Year(), d.Month()+time.Month(n), 1, 0, 0, 0, 0, d.Location())
	last := first.AddDate(0, 1, -1).Day()
	return first.AddDate(0, 0, min(d.Day(), last)-1)
}

// TagInputOptions are the options of TagInput.
type TagInputOptions struct {
	// Placeholder shows while there are no tags.
	Placeholder string
	// Max, if set, is the most tags it takes.
	Max      int
	Label    string
	Size     Size
	Disabled bool
	Invalid  bool
}

// TagInput is a field of tags: Input's inset well holding the tags as
// neutral pills, each with a small button taking it out, before a bare
// input. Enter or a comma adds what was typed, unless the tags hold it
// already (ignoring case), and Backspace in the empty input takes out the
// last tag. The well grows a line as the tags fill one. It returns the
// text input.
//
//	mizan.TagInput(c.Key("tags"), &app.tags, mizan.TagInputOptions{Placeholder: "Add a tag"})
func TagInput(c *ui.Context, tags *[]string, o TagInputOptions) ui.Element {
	t := sizedFrom(c, o.Size)
	m := t.Metrics
	cl := t.Colors
	chipH := m.FieldH - 10

	box := ui.Row(c).Wrap().AlignItems(ui.Center).Gap(4).MinHeight(m.FieldH).Padding(4, m.FieldPe, 4, 4).
		Radius(m.FieldR).MinWidth(0).Cursor(ui.CursorText)
	text := ui.Local(box, "text", func() string { return "" })
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" || o.Max > 0 && len(*tags) >= o.Max {
			return
		}
		for _, have := range *tags {
			if strings.EqualFold(have, s) {
				return
			}
		}
		*tags = append(*tags, s)
	}
	var field ui.Element
	box.Children(func() {
		for i, tag := range *tags {
			ui.Row(c.Key(tag)).Height(chipH).AlignItems(ui.Center).Gap(2).Padding(0, 3, 0, m.ChipPx-2).
				Radius(Concentric(m.FieldR, 4)).Background(cl.Fill).Shrink(0).Children(func() {
				ui.Text(c, tag).TextColor(cl.Text).FontSize(m.ChipFs).FontWeight(500).NoWrap()
				d := chipH - 8
				// not a stop of Tab: Backspace takes tags out from the keyboard
				x := ui.Box(c).Size(d, d).Center().Radius(RadiusPill).Role(ui.RoleButton).Label("Remove " + tag).
					KeepFocus().Cursor(ui.CursorPointer).TextColor(cl.Muted).Disabled(o.Disabled)
				if x.Hovered() && !o.Disabled {
					x.Background(Mix(cl.Fill, cl.Text, 0.12)).TextColor(cl.Text)
				}
				x.Children(func() { ui.Icon(c, iconClear).Size(d*0.55, d*0.55) })
				if x.Clicked() {
					*tags = append((*tags)[:i:i], (*tags)[i+1:]...)
				}
			})
		}
		// the input keeps its identity, whatever tags come before it
		field = ui.TextInputBase(c.Key("input"), text).Grow(1).Basis(64).MinWidth(64).Height(chipH).Padding(0, 0, 0, 6).
			PlaceholderColor(cl.Muted).TextColor(cl.Text).FontSize(fieldFs(m)).Disabled(o.Disabled).FocusRing(false)
		if len(*tags) == 0 {
			field.Placeholder(o.Placeholder)
		}
		if o.Label != "" {
			field.Label(o.Label)
		}
		field.OnTextKey(func(k ui.TextKey) bool {
			if k.Key != ui.KeyBackspace || k.Mods != 0 || k.Text() != "" || len(*tags) == 0 {
				return false
			}
			*tags = (*tags)[:len(*tags)-1]
			k.SetText("")
			return true
		})
		if strings.Contains(*text, ",") {
			parts := strings.Split(*text, ",")
			for _, s := range parts[:len(parts)-1] {
				add(s)
			}
			*text = parts[len(parts)-1]
		}
		if field.Submitted() {
			add(*text)
			*text = ""
		}
	})
	t.well(box, field.Focused(), o.Invalid)
	if o.Disabled {
		box.Opacity(0.4).Cursor(ui.CursorDefault)
	} else if box.Clicked() {
		field.Focus()
	}
	return field
}
