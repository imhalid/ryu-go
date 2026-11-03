package gallery

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/mizan-go"
)

// order is a row of the orders table.
type order struct {
	ID       int
	Customer string
	Region   string
	Status   mz.CellStatus
	Items    int
	Total    float64
	Change   float64
	Placed   time.Time
	Packed   float64
}

// member is a row of the team table.
type member struct {
	Name, Role, Team string
	Load             float64
}

// tablePage is the state of the table page.
type tablePage struct {
	ready         bool
	Orders        []order
	Orders2       mz.TableState
	Team          []member
	TeamState     mz.TableState
	Loading, Fail bool
	loadAt        time.Time
	next          int
}

var (
	customers = []string{"Maya Chen", "Leo Park", "Sam Rivera", "Noah Kim", "Ivy Brooks", "Ada Lovelace", "Omar Haddad",
		"Lena Fischer", "Kenji Sato", "Priya Nair", "Tom Becker", "Elif Yılmaz", "Lucas Moreau", "Sofia Rossi"}
	regions  = []string{"Europe", "North America", "Asia Pacific", "Middle East", "Latin America"}
	statuses = []mz.CellStatus{{State: mz.DotLive, Text: "Paid"}, {State: mz.DotDelayed, Text: "Pending"},
		{State: mz.DotOff, Text: "Refunded"}, {State: mz.DotDelayed, Text: "On hold"}}
)

// init fills the page with sample rows the first time it shows: the
// same rows every time.
func (p *tablePage) init() {
	p.ready = true
	seed := uint32(7)
	rnd := func(n int) int {
		seed = seed*1664525 + 1013904223
		return int(seed>>8) % n
	}
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	for i := range 240 {
		p.Orders = append(p.Orders, order{
			ID: 1040 + i, Customer: customers[rnd(len(customers))], Region: regions[rnd(len(regions))],
			Status: statuses[rnd(len(statuses))], Items: 1 + rnd(12),
			Total: float64(rnd(480000))/100 + 12, Change: float64(rnd(400)-160) / 10,
			Placed: start.Add(time.Duration(rnd(33*24)) * time.Hour), Packed: float64(rnd(101)) / 100,
		})
	}
	p.next = 1040 + 240
	p.Orders2.Sort = []mz.SortKey{{Column: "Placed", Desc: true}}
	roles := []string{"Engineer", "Designer", "Product", "Research", "Support"}
	teams := []string{"Payments", "Growth", "Platform", "Mobile"}
	for i, n := range customers {
		p.Team = append(p.Team, member{Name: n, Role: roles[i%len(roles)], Team: teams[rnd(len(teams))], Load: float64(rnd(101)) / 100})
	}
}

func (p *tablePage) view(c *ui.Context, t *mz.Tokens) {
	if !p.ready {
		p.init()
	}
	cl := t.Colors
	ui.Column(c).Gap(mz.GapSection).Children(func() {
		mz.Card(c, mz.CardOptions{Title: "Orders", Subtitle: "Sort by a header, drag one to move its column, open a row for its details.", Pad: true}, func() {
			rows := p.Orders
			if q := strings.ToLower(strings.TrimSpace(p.Orders2.Query)); q != "" {
				rows = slices.DeleteFunc(slices.Clone(rows), func(o order) bool {
					return !strings.Contains(strings.ToLower(o.Customer+" "+o.Region+" "+o.Status.Text+" "+fmt.Sprint(o.ID)), q)
				})
			}
			mz.Table(c.Key("orders"), &p.Orders2, orderColumns(), rows, mz.TableOptions[order]{
				Key:        func(o order) any { return o.ID },
				Label:      "Orders",
				Selectable: true, Search: true, ViewMenu: true, Pin: true, Height: 520,
				Loading: p.Loading, Error: errText(p.Fail), Retry: func() { p.Fail = false },
				Empty: mz.EmptyStateOptions{Icon: inbox, Title: "No orders", Description: "Nothing matches the search."},
				Actions: []mz.TableAction{
					{Label: "Export", Do: func([]any) {}},
					{Label: "Archive", Danger: true, Do: func(keys []any) { p.archive(keys) }},
				},
				Toolbar: func() {
					if mz.Button(c.Key("add"), "New order", mz.ButtonOptions{Size: mz.Sm}).Clicked() {
						p.add()
					}
					if mz.Button(c.Key("load"), "Reload", mz.ButtonOptions{Size: mz.Sm, Loading: p.Loading}).Clicked() {
						p.Loading, p.loadAt = true, c.Now()
					}
					if mz.Button(c.Key("fail"), "Fail", mz.ButtonOptions{Size: mz.Sm}).Clicked() {
						p.Fail = true
					}
					if p.Loading {
						// a stand-in for the rows arriving
						const wait = 1600 * time.Millisecond
						if left := wait - c.Now().Sub(p.loadAt); left > 0 {
							c.After(left)
						} else {
							p.Loading = false
						}
					}
				},
				Expand: func(c *ui.Context, o order) {
					ui.Row(c).Gap(32).Children(func() {
						detail(c, cl, "Order", fmt.Sprintf("#%d", o.ID))
						detail(c, cl, "Items", fmt.Sprint(o.Items))
						detail(c, cl, "Placed", o.Placed.Format("Mon, Jan 2 · 15:04"))
						detail(c, cl, "Ships from", o.Region)
					})
					ui.Text(c, "Two parcels; the second leaves the warehouse once the last item is packed.").TextColor(cl.Muted).FontSize(mz.FsSm)
				},
			})
		})
		mz.Card(c, mz.CardOptions{Title: "Team", Subtitle: "A page at a time, in the compact density.", Pad: true}, func() {
			mz.Table(c.Key("team"), &p.TeamState, []mz.Column[member]{
				{Title: "Name", Kind: mz.KindPerson, Value: func(m member) any { return m.Name }, Sortable: true},
				{Title: "Role", Value: func(m member) any { return m.Role }, Sortable: true},
				{Title: "Team", Value: func(m member) any { return m.Team }, Sortable: true},
				{Title: "Load", Kind: mz.KindBar, Value: func(m member) any { return m.Load }, Sortable: true, Width: 140},
			}, p.Team, mz.TableOptions[member]{
				Key: func(m member) any { return m.Name }, Label: "Team", PageSize: 5, Height: 300, Size: mz.Sm,
			})
		})
	})
}

func errText(fail bool) string {
	if fail {
		return "The orders service didn't answer in time."
	}
	return ""
}

func detail(c *ui.Context, cl mz.Colors, label, value string) {
	ui.Column(c).Gap(2).Children(func() {
		ui.Text(c, label).FontSize(mz.FsXs).TextColor(cl.Muted)
		ui.Text(c, value).FontSize(mz.FsMd).TextColor(cl.Text).FontFeatures("tnum")
	})
}

func orderColumns() []mz.Column[order] {
	return []mz.Column[order]{
		{Title: "Customer", Kind: mz.KindPerson, Value: func(o order) any { return o.Customer }, Sortable: true, MinWidth: 180},
		{Title: "Order", Value: func(o order) any { return fmt.Sprintf("#%d", o.ID) }, Sortable: true, Width: 96},
		{Title: "Status", Kind: mz.KindStatus, Value: func(o order) any { return o.Status }, Sortable: true, Width: 130},
		{Title: "Region", Value: func(o order) any { return o.Region }, Sortable: true, MinWidth: 140},
		{Title: "Placed", Kind: mz.KindDate, Value: func(o order) any { return o.Placed }, Sortable: true, Width: 130},
		{Title: "Total", Kind: mz.KindNumber, Value: func(o order) any { return o.Total }, Sortable: true, Decimals: 2, Unit: "USD", Width: 140},
		{Title: "vs. last month", ID: "Change", Kind: mz.KindChange, Value: func(o order) any { return o.Change }, Sortable: true, Width: 130},
		{Title: "Packed", Kind: mz.KindBar, Value: func(o order) any { return o.Packed }, Sortable: true, Width: 140},
	}
}

// archive takes the orders out.
func (p *tablePage) archive(keys []any) {
	gone := map[any]bool{}
	for _, k := range keys {
		gone[k] = true
	}
	p.Orders = slices.DeleteFunc(p.Orders, func(o order) bool { return gone[o.ID] })
	p.Orders2.Selected = nil
}

// add puts a new order first.
func (p *tablePage) add() {
	o := order{ID: p.next, Customer: customers[p.next%len(customers)], Region: regions[p.next%len(regions)],
		Status: statuses[1], Items: 3, Total: 249.5, Change: 2.4, Placed: time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC), Packed: 0}
	p.next++
	p.Orders = append([]order{o}, p.Orders...)
}
