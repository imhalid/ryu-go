package mizan

import (
	"time"

	"github.com/egoist/mygo/ui"
)

var iconChevronEnd = Svg(16, 14, "M6 3.5l4.5 4.5-4.5 4.5")

// CollapsibleOptions are the options of Collapsible.
type CollapsibleOptions struct {
	// Description is a muted line under the title, shown open or closed.
	Description string
	Disabled    bool
}

// Collapsible is a disclosure: a title beside a chevron, which a click,
// Enter or Space opens and closes, and below it what body builds while
// *open is true. The chevron turns a quarter, the body grows into view
// and fades in as it opens, and shrinks away as it closes. The header
// takes a soft fill under the pointer. It returns the header, whose
// Changed reports a click.
//
//	mizan.Collapsible(c.Key("advanced"), "Advanced", &app.advanced, mizan.CollapsibleOptions{}, func() { ... })
func Collapsible(c *ui.Context, title string, open *bool, o CollapsibleOptions, body func()) ui.Element {
	t := From(c)
	var head ui.Element
	ui.Column(c).FillWidth().MinWidth(0).Children(func() {
		head = t.disclosure(c, title, open, o, body, false)
	})
	return head
}

// AccordionOptions are the options of Accordion.
type AccordionOptions struct {
	// Single keeps one section open at a time: opening one closes the
	// others.
	Single bool
}

// AccordionList builds an Accordion's sections.
type AccordionList struct {
	c      *ui.Context
	t      *Tokens
	n      int
	opens  []*bool
	opened *bool
}

// Item adds a section: its title, open while *open, and what body builds
// in it. It returns its header, whose Changed reports a click.
func (a *AccordionList) Item(title string, open *bool, body func()) ui.Element {
	return a.ItemWith(title, open, CollapsibleOptions{}, body)
}

// ItemWith is Item with the options of a Collapsible.
func (a *AccordionList) ItemWith(title string, open *bool, o CollapsibleOptions, body func()) ui.Element {
	c := a.c.Key(a.n)
	first := a.n == 0
	a.n++
	a.opens = append(a.opens, open)
	var head ui.Element
	sec := ui.Column(c).FillWidth().MinWidth(0).Padding(4, 0)
	if !first {
		// a groove above every section but the first
		sec.InsetShadow(0, 1, 0, 0, a.t.Colors.SepDark).InsetShadow(0, 2, 0, 0, a.t.Colors.SepLight)
	}
	sec.Children(func() {
		head = a.t.disclosure(c, title, open, o, body, true)
	})
	if head.Changed() && *open {
		a.opened = open
	}
	return head
}

// Accordion is a stack of Collapsible sections divided by grooves, as a
// settings page's; build them with Item. Each opens and closes on its own,
// or, with Single, opening one closes the others.
//
//	mizan.Accordion(c.Key("faq"), mizan.AccordionOptions{Single: true}, func(a *mizan.AccordionList) {
//		a.Item("Billing", &app.billing, func() { ... })
//		a.Item("Security", &app.security, func() { ... })
//	})
func Accordion(c *ui.Context, o AccordionOptions, build func(a *AccordionList)) ui.Element {
	t := From(c)
	a := &AccordionList{c: c, t: t}
	e := ui.Column(c).FillWidth().MinWidth(0).Children(func() { build(a) })
	if o.Single && a.opened != nil {
		for _, open := range a.opens {
			if open != a.opened {
				*open = false
			}
		}
	}
	return e
}

// disclosure builds a collapsible's header and body; in an accordion
// (wide), the header spans the row and the title is a little larger.
func (t *Tokens) disclosure(c *ui.Context, title string, open *bool, o CollapsibleOptions, body func(), wide bool) ui.Element {
	cl := t.Colors
	p := ui.CollapsibleBase(c, open)
	prog := p.Progress()
	h := p.Trigger.FillWidth().AlignItems(ui.Center).Gap(10).Padding(8, 10).Radius(t.R(RadiusItem)).
		Cursor(ui.CursorPointer).FocusRing(false).Disabled(o.Disabled)
	if wide {
		h.Padding(10, 10)
	}
	if h.Hovered() && !o.Disabled {
		h.Background(cl.Fill)
	}
	if h.FocusVisible() {
		t.Shadows.FocusRing.Apply(h)
	}
	if o.Disabled {
		h.Opacity(0.4).Cursor(ui.CursorDefault)
	}
	fs := float32(FsMd)
	if wide {
		fs = FsBase
	}
	h.Transition(ui.ElementTransition{Duration: 150 * time.Millisecond, Colors: true}).Children(func() {
		ui.Icon(c, iconChevronEnd).TextColor(cl.Muted).Rotate(90 * prog)
		ui.Column(c).Grow(1).MinWidth(0).Gap(2).Children(func() {
			ui.Text(c, title).TextColor(cl.Text).FontSize(fs).FontWeight(500).Wrap()
			if o.Description != "" {
				ui.Text(c, o.Description).TextColor(cl.Muted).FontSize(FsSm).Wrap()
			}
		})
	})
	if panel := p.Panel(body); panel != (ui.Element{}) {
		// under the title, past the chevron; it rises 4px as it fades in
		panel.FillWidth().Padding(4, 10, 12, 34).Gap(10).TextColor(cl.Text2).FontSize(FsMd).
			Opacity(prog).Translate(0, -4*(1-prog))
	}
	return h
}
