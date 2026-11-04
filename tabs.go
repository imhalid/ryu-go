package ryugo

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// Tab is one tab of Tabs.
type Tab struct {
	Label string
	Icon  *ui.SVG
	// Badge is a small count beside the label, as "12".
	Badge string
}

// TabsOptions are the options of Tabs.
type TabsOptions struct {
	// Label is the tab list's accessible name.
	Label string
	Size  Size
}

// Tabs is a row of tabs over a groove: the chosen tab's label is in the
// text color with a 2px underline that slides to a new choice on a spring,
// and a soft fill follows the pointer from tab to tab. The arrows, Home and
// End move the choice while a tab has the focus (straight there, without
// the slide). Build the chosen panel below it, in TabPanel.
//
//	ryugo.Tabs(c.Key("tabs"), &app.tab, []ryugo.Tab{{Label: "Overview"}, {Label: "Activity", Badge: "3"}}, ryugo.TabsOptions{})
//	ryugo.TabPanel(c, app.tab, func() { ... })
func Tabs(c *ui.Context, selected *int, tabs []Tab, o TabsOptions) ui.Element {
	t := sizedFrom(c, o.Size)
	m := t.Metrics
	cl := t.Colors
	parts := ui.TabsBase(c, selected, len(tabs))
	list := parts.List.Gap(2).AlignItems(ui.Stretch).
		InsetShadow(0, -2, 0, 0, cl.SepLight).InsetShadow(0, -1, 0, 0, cl.SepDark)
	if o.Label != "" {
		list.Label(o.Label)
	}
	els := make([]ui.Element, len(tabs))
	hovered, keys := -1, false
	list.Children(func() {
		for i, tb := range tabs {
			e := parts.Tab(i).Height(m.TabH).Padding(0, m.TabPx).AlignItems(ui.Center).Gap(6).
				Radius(t.R(RadiusItem)).Cursor(ui.CursorPointer).FocusRing(false).Border(2, ui.Color{})
			els[i] = e
			text := cl.Muted
			if *selected == i || e.Hovered() {
				text = cl.Text
			}
			if e.Hovered() {
				hovered = i
			}
			if e.FocusVisible() {
				keys = true
				e.Border(2, Mix(ui.Color{}, cl.ColorAccent500, 0.6))
			}
			if e.Pressed() {
				e.Scale(0.97, 0.97)
			}
			e.TextColor(text).Transition(ui.ElementTransition{Duration: 200 * time.Millisecond, Colors: true}).Children(func() {
				if tb.Icon != nil {
					ui.Icon(c, tb.Icon)
				}
				ui.Text(c, tb.Label).FontSize(m.TabFs).FontWeight(500).NoWrap()
				if tb.Badge != "" {
					ui.Box(c).Padding(0, 6).Radius(RadiusPill).Background(cl.Fill).Children(func() {
						ui.Text(c, tb.Badge).TextColor(cl.Muted).FontSize(Fs2xs).FontWeight(600).FontFeatures("tnum")
					})
				}
			})
		}
	})

	// The underline and the hover fill are painted by the list, under the
	// tabs, placed from where the last frame laid them out.
	lb := list.LayoutBounds()
	at := func(i int) (ui.Rect, bool) {
		if i < 0 || i >= len(els) || lb.W == 0 {
			return ui.Rect{}, false
		}
		b := els[i].LayoutBounds()
		return ui.Rect{X: b.X - lb.X, Y: b.Y - lb.Y, W: b.W, H: b.H}, b.W > 0
	}
	sel, hasSel := at(*selected)
	placed := ui.Local(list, "placed", func() bool { return false })
	var ux, uw float32
	if hasSel {
		if keys || !*placed {
			ux, uw = list.Jump("line-x", sel.X), list.Jump("line-w", sel.W)
			*placed = true
		} else {
			ux, uw = list.AnimateTo("line-x", sel.X, springPill), list.AnimateTo("line-w", sel.W, springPill)
		}
	}
	// the fill stays on the last tab hovered as it fades out
	hv, hasHover := at(hovered)
	last := ui.Local(list, "hovered", func() ui.Rect { return ui.Rect{} })
	fade := list.AnimateTo("hover-a", b2f(hasHover), ui.Animation{Duration: 200 * time.Millisecond})
	var hx, hw float32
	switch {
	case hasHover && fade < 0.05:
		// a fill appearing starts on its tab, rather than sliding in from the last
		*last = hv
		hx, hw = list.Jump("hover-x", hv.X), list.Jump("hover-w", hv.W)
	default:
		if hasHover {
			*last = hv
		}
		glide := ui.Animation{Duration: 250 * time.Millisecond, Ease: ui.EaseOut}
		hx, hw = list.AnimateTo("hover-x", last.X, glide), list.AnimateTo("hover-w", last.W, glide)
	}
	h := m.TabH
	list.Draw(func(p *ui.Painter, r ui.Rect) {
		if fade > 0 && hw > 0 {
			p.Fill(ui.Rect{X: r.X + hx, Y: r.Y + 4, W: hw, H: h - 8}, cl.Fill.Alpha(fade), t.R(RadiusItem))
		}
		if hasSel && uw > 16 {
			p.Fill(ui.Rect{X: r.X + ux + 8, Y: r.Y + r.H - 2, W: uw - 16, H: 2}, cl.Text, 1)
		}
	})
	return list
}

// TabPanel builds the panel of the chosen tab: a new panel fades in,
// rising 4px.
func TabPanel(c *ui.Context, selected int, body func()) ui.Element {
	return ui.Column(c.Key(selected)).FillWidth().MinWidth(0).Margin(16, 0, 0, 0).Children(body).
		Transition(ui.ElementTransition{Duration: 180 * time.Millisecond, Ease: ui.EaseOut, Enter: &ui.Motion{Y: 4}})
}
