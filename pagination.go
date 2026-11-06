package ryugo

import (
	"strconv"

	"github.com/egoist/mygo/ui"
)

// PaginationOptions are the options of Pagination.
type PaginationOptions struct {
	// PageSize is how many items a page holds.
	PageSize int
	Size     Size
	// Labels are its words, for an app in another language.
	Labels PaginationLabels
}

// Pagination moves between the pages of total items, *page being the page
// shown from 0: the range shown and the total on the start side ("21–40
// of 200"), and on the end side the previous and next buttons around the
// page numbers, the first and last always there and the run around the
// page in between, gaps shown as "…". The page shown sits on a raised pill
// that slides from number to number on a spring; the numbers keep their
// places as it moves.
//
//	ryugo.Pagination(c.Key("pages"), &app.page, len(app.rows), ryugo.PaginationOptions{PageSize: 20})
func Pagination(c *ui.Context, page *int, total int, o PaginationOptions) ui.Element {
	t := sizedFrom(c, o.Size)
	m := t.Metrics
	cl := t.Colors
	l := o.Labels.fill()
	ps := max(o.PageSize, 1)
	pages := max(1, (total+ps-1)/ps)
	*page = min(max(*page, 0), pages-1)
	from, to := *page*ps+1, min(total, (*page+1)*ps)
	bar := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(8).Role(ui.RoleGroup).Label(l.Group)
	bar.Children(func() {
		ui.Text(c, strconv.Itoa(from)+"–"+strconv.Itoa(to)+" "+l.Of+" "+grouped(float64(total), 0)).
			FontSize(FsSm).TextColor(cl.Muted).FontFeatures("tnum").NoWrap()
		ui.Spacer(c)
		step := func(key, label string, icon *ui.SVG, to int, off bool) {
			b := pageButton(c.Key(key), t, label, off)
			b.Children(func() { ui.Icon(c, icon).Size(14, 14) })
			if b.Clicked() && !off {
				*page = to
			}
		}
		step("prev", l.Previous, iconPrev, *page-1, *page == 0)
		slots := pageSlots(*page, pages)
		track := ui.Row(c).AlignItems(ui.Center).Gap(2)
		var at ui.Element
		track.Children(func() {
			for n, p := range slots {
				if p < 0 {
					ui.Box(c.Key(n)).Size(m.BtnH-4, m.BtnH).Center().Children(func() {
						ui.Text(c, "…").FontSize(FsSm).TextColor(cl.Muted)
					})
					continue
				}
				label := l.Page + " " + strconv.Itoa(p+1)
				if p == *page {
					label += ", " + l.Current
				}
				b := pageButton(c.Key(n), t, label, false)
				color := cl.Muted
				if p == *page {
					color = cl.Text
					at = b
				} else if b.Hovered() {
					color = cl.Text2
				}
				b.TextColor(color).Children(func() {
					ui.Text(c, strconv.Itoa(p+1)).FontSize(FsSm).FontWeight(600).FontFeatures("tnum")
				})
				if b.Clicked() {
					*page = p
				}
			}
		})
		// the pill under the page shown, painted by the track under the
		// numbers, placed from where the last frame laid them out
		if at != (ui.Element{}) {
			ab, tb := at.LayoutBounds(), track.LayoutBounds()
			if ab.W > 0 && tb.W > 0 {
				placed := ui.Local(track, "placed", func() bool { return false })
				x, w := ab.X-tb.X, ab.W
				if !*placed {
					x, w = track.Jump("pill-x", x), track.Jump("pill-w", w)
					*placed = true
				} else {
					x, w = track.AnimateTo("pill-x", x, springPill), track.AnimateTo("pill-w", w, springPill)
				}
				y, h := ab.Y-tb.Y, ab.H
				track.Draw(func(p *ui.Painter, r ui.Rect) {
					t.paintRaised(p, ui.Rect{X: r.X + x, Y: r.Y + y, W: w, H: h}, t.Pill(h), t.Shadows.ShadowRaised)
				})
			}
		}
		step("next", l.Next, iconNext, *page+1, *page == pages-1)
	})
	return bar
}

// pageButton is a round button of Pagination: no fill of its own (the
// pill under the page shown is the track's), a little smaller as it is
// pressed.
func pageButton(c *ui.Context, t *Tokens, label string, off bool) ui.Element {
	h := t.Metrics.BtnH
	b := ui.ButtonBase(c).Label(label).Center().MinWidth(h).Height(h).Padding(0, 6).Radius(t.Pill(h)).
		Cursor(ui.CursorPointer).FocusRing(false).Disabled(off).TextColor(t.Colors.Text2)
	if off {
		b.Opacity(0.35).Cursor(ui.CursorDefault)
	}
	if b.Pressed() && !off {
		b.Scale(0.94, 0.94)
	}
	if b.FocusVisible() {
		t.Shadows.FocusRing.Apply(b)
	}
	return b
}

// pageSlots are the pages Pagination shows, by slot, -1 for a gap: all of
// them up to 7, else the first and the last with the run around page.
func pageSlots(page, pages int) []int {
	if pages <= 7 {
		s := make([]int, pages)
		for i := range s {
			s[i] = i
		}
		return s
	}
	switch {
	case page < 4:
		return []int{0, 1, 2, 3, 4, -1, pages - 1}
	case page > pages-5:
		return []int{0, -1, pages - 5, pages - 4, pages - 3, pages - 2, pages - 1}
	}
	return []int{0, -1, page - 1, page, page + 1, -1, pages - 1}
}
