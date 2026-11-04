package main

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
	Total    float64
	Change   float64
	Placed   time.Time
	Packed   float64
}

// orders are the table's sample rows: the same rows every time.
func orders() []order {
	customers := []string{"Maya Chen", "Leo Park", "Sam Rivera", "Noah Kim", "Ivy Brooks", "Ada Lovelace", "Omar Haddad",
		"Lena Fischer", "Kenji Sato", "Priya Nair", "Tom Becker", "Elif Yılmaz", "Lucas Moreau", "Sofia Rossi"}
	regions := []string{"Europe", "North America", "Asia Pacific", "Middle East", "Latin America"}
	statuses := []mz.CellStatus{{State: mz.DotLive, Text: "Paid"}, {State: mz.DotDelayed, Text: "Pending"},
		{State: mz.DotOff, Text: "Refunded"}, {State: mz.DotDelayed, Text: "On hold"}}
	seed := uint32(7)
	rnd := func(n int) int {
		seed = seed*1664525 + 1013904223
		return int(seed>>8) % n
	}
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	var rows []order
	for i := range 120 {
		rows = append(rows, order{
			ID: 1040 + i, Customer: customers[rnd(len(customers))], Region: regions[rnd(len(regions))],
			Status: statuses[rnd(len(statuses))], Total: float64(rnd(480000))/100 + 12, Change: float64(rnd(400)-160) / 10,
			Placed: start.Add(time.Duration(rnd(33*24)) * time.Hour), Packed: float64(rnd(101)) / 100,
		})
	}
	return rows
}

func init() {
	rows := orders()
	state := mz.TableState{Sort: []mz.SortKey{{Column: "Placed", Desc: true}}}
	cols := []mz.Column[order]{
		{Title: "Customer", Kind: mz.KindPerson, Value: func(o order) any { return o.Customer }, Sortable: true, MinWidth: 170},
		{Title: "Order", Value: func(o order) any { return fmt.Sprintf("#%d", o.ID) }, Sortable: true, Width: 90},
		{Title: "Status", Kind: mz.KindStatus, Value: func(o order) any { return o.Status }, Sortable: true, Width: 120},
		{Title: "Region", Value: func(o order) any { return o.Region }, Sortable: true, MinWidth: 130},
		{Title: "Placed", Kind: mz.KindDate, Value: func(o order) any { return o.Placed }, Sortable: true, Width: 120},
		{Title: "Total", Kind: mz.KindNumber, Value: func(o order) any { return o.Total }, Sortable: true, Decimals: 2, Unit: "USD", Width: 130},
		{Title: "vs. last month", ID: "Change", Kind: mz.KindChange, Value: func(o order) any { return o.Change }, Sortable: true, Width: 120},
		{Title: "Packed", Kind: mz.KindBar, Value: func(o order) any { return o.Packed }, Sortable: true, Width: 120},
	}
	demo("table", func(c *ui.Context, t *mz.Tokens) {
		shown := rows
		if q := strings.ToLower(strings.TrimSpace(state.Query)); q != "" {
			shown = slices.DeleteFunc(slices.Clone(rows), func(o order) bool {
				return !strings.Contains(strings.ToLower(o.Customer+" "+o.Region+" "+o.Status.Text+" "+fmt.Sprint(o.ID)), q)
			})
		}
		mz.Table(c.Key("orders"), &state, cols, shown, mz.TableOptions[order]{
			Key:        func(o order) any { return o.ID },
			Label:      "Orders",
			Selectable: true, Search: true, ViewMenu: true, Pin: true, Height: 440,
			Empty: mz.EmptyStateOptions{Icon: iconInbox, Title: "No orders", Description: "Nothing matches the search."},
			Actions: []mz.TableAction{
				{Label: "Archive", Danger: true, Do: func(keys []any) {
					rows = slices.DeleteFunc(rows, func(o order) bool { return slices.Contains(keys, any(o.ID)) })
					state.Selected = nil
				}},
			},
			Expand: func(c *ui.Context, o order) {
				ui.Row(c).Gap(32).Children(func() {
					for _, kv := range [][2]string{{"Order", fmt.Sprintf("#%d", o.ID)}, {"Placed", o.Placed.Format("Mon, Jan 2 · 15:04")}, {"Ships from", o.Region}} {
						ui.Column(c.Key(kv[0])).Gap(2).Children(func() {
							ui.Text(c, kv[0]).FontSize(mz.FsXs).TextColor(t.Colors.Muted)
							ui.Text(c, kv[1]).FontSize(mz.FsMd).TextColor(t.Colors.Text)
						})
					}
				})
			},
		}).AlignSelf(ui.Stretch)
	})

	type member struct {
		Name, Role string
		Load       float64
	}
	team := []member{{"Maya Chen", "Designer", 0.72}, {"Leo Park", "Engineer", 0.4}, {"Sam Rivera", "Engineer", 0.91},
		{"Noah Kim", "Product", 0.55}, {"Ivy Brooks", "Research", 0.18}, {"Ada Lovelace", "Engineer", 0.63},
		{"Omar Haddad", "Support", 0.33}, {"Lena Fischer", "Designer", 0.8}, {"Kenji Sato", "Engineer", 0.47}}
	var teamState mz.TableState
	demo("table-pages", func(c *ui.Context, t *mz.Tokens) {
		mz.Table(c.Key("team"), &teamState, []mz.Column[member]{
			{Title: "Name", Kind: mz.KindPerson, Value: func(m member) any { return m.Name }, Sortable: true},
			{Title: "Role", Value: func(m member) any { return m.Role }, Sortable: true},
			{Title: "Load", Kind: mz.KindBar, Value: func(m member) any { return m.Load }, Sortable: true, Width: 140},
		}, team, mz.TableOptions[member]{
			Key: func(m member) any { return m.Name }, Label: "Team", PageSize: 4, Height: 260, Size: mz.Sm,
		}).AlignSelf(ui.Stretch)
	})
}
