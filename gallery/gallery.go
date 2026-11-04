// Package gallery is the showcase of Ryugo's components: each component in
// every size, side by side in both themes and both materials.
package gallery

import (
	"time"

	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/ryu-go"
)

// State is the state of the showcase.
type State struct {
	Mode mz.Mode
	// Accent is the brand color picked in the header, zero for Ryugo's blue.
	Accent ui.Color
	// Density and Round are the size step and the corners picked in the
	// header: 0 small, 1 medium, 2 large.
	Density, Round int
	Query          string
	Email          string
	Pass           string
	Range          int
	Team           string
	Count          int

	Mail, Push, Digest bool
	Plan, Tier         string
	Sync, LED          bool
	ViewMode           int
	Volume, Zoom       float64

	// Page is 0: controls; 1: chat, drag and drop, and indicators; 2: layers;
	// 3: more form controls; 4: feedback, disclosure and toggles; 5: tables;
	// 6: the workspace; 7: the studio, palettes over a canvas.
	Page  int
	Msgs  []mz.ChatMessage
	Draft string
	// the gallery's chat: how many messages were sent, and when the last
	sent   int
	sentAt time.Time
	// shake counts the form's checks, which shake its invalid field
	shake int
	Steps []string
	steps ui.ListState
	Cards []*card
	Files []file

	// The layers page: what is open, and what its controls set.
	Dialog, Sheet, Pop, Menu bool
	uploadAt                 time.Time
	Sort                     string
	Last                     string
	Tab                      int

	// Forms is the state of the forms page.
	Forms formsPage
	// Feedback is the state of the feedback page.
	Feedback feedbackPage
	// Tables is the state of the table page.
	Tables tablePage
	// Work is the state of the workspace page.
	Work   workspacePage
	Studio studioPage
}

// card is a task on the board; col is the column it is in.
type card struct {
	Title, Tag string
	col        int
}

type file struct{ Name, Size string }

// New returns a showcase state with sample values.
func New() *State {
	return &State{
		Density: 1, Round: 1,
		Range: 30, Email: "ada@", Pass: "secret",
		Mail: true, Plan: "week", Tier: "pro", Sync: true, Volume: 64, Zoom: 1.25,
		Msgs: []mz.ChatMessage{
			{Author: "Maya Chen", Text: "Is the weekly report ready?", Time: "09:41"},
			{Author: "Maya Chen", Text: "I'd like to take a look before the meeting.", Time: "09:41"},
			{Mine: true, Text: "Almost; I'm adding the last charts.", Time: "09:43", Status: mz.Read},
			{Mine: true, Text: "I'll share it in ten minutes.", Time: "09:43", Status: mz.Sent},
			{Author: "Leo Park", Text: "I checked the data sources too, all good 👍", Time: "09:45"},
		},
		Steps: []string{"Import data", "Map columns", "Validate", "Publish"},
		Cards: []*card{
			{Title: "Sign-in screen", Tag: "Design", col: 0},
			{Title: "Report export", Tag: "Web", col: 0},
			{Title: "Notification permissions", Tag: "iOS", col: 1},
			{Title: "Dark theme", Tag: "Design", col: 2},
		},
		Files:    []file{{"Q3-report.pdf", "2.4 MB"}, {"patient-list-october-2026-final-version.xlsx", "860 KB"}},
		Feedback: newFeedback(),
	}
}

var search = mz.Svg(16, 15, "M7 12.5a5.5 5.5 0 1 0 0-11 5.5 5.5 0 0 0 0 11zM14.5 14.5l-3.6-3.6")

// View builds the showcase.
// steps and corners are the header's size steps and corners.
var (
	steps   = [...]mz.Size{mz.Sm, mz.Md, mz.Lg}
	corners = [...]mz.Corners{mz.CornersSm, mz.CornersMd, mz.CornersLg}
)

func (s *State) View(c *ui.Context) {
	t := mz.Root(c, mz.Options{Mode: s.Mode, Accent: s.Accent, Size: steps[s.Density], Corners: corners[s.Round]})
	ui.Scroll(c).Fill().Children(func() { s.page(c, t) })
	mz.Toaster(c)
}

func (s *State) page(c *ui.Context, t *mz.Tokens) {
	ui.Column(c).FillWidth().Padding(mz.Gutter-8, mz.Gutter).Gap(mz.GapSection).Children(func() {
		ui.Row(c).AlignItems(ui.End).Justify(ui.SpaceBetween).Gap(16).Children(func() {
			ui.Column(c).Gap(4).Children(func() {
				mz.Text(c, "Ryugo", mz.TDisplay)
				mz.Text(c, "Native components you can feel.", mz.TLede).TextColor(t.Colors.Muted)
			})
			// the pages, and under them how the page looks: the accent, the
			// size step, the corners and the theme
			ui.Column(c).Gap(10).AlignItems(ui.End).Children(func() {
				ui.Row(c).Gap(mz.GapControl).Children(func() {
					for i, name := range []string{"Controls", "Chat and drag", "Layers", "Forms", "Feedback", "Table", "Workspace", "Studio"} {
						o := mz.ButtonOptions{Size: mz.Sm}
						if s.Page == i {
							o.Color = t.Colors.Accent
						}
						if mz.Button(c.Key(name), name, o).Clicked() {
							s.Page = i
						}
					}
				})
				ui.Row(c).Gap(mz.GapControl).AlignItems(ui.Center).Children(func() {
					s.accents(c, t)
					ui.Box(c).Size(10, 0)
					setting(c, t, "Size", &s.Density)
					setting(c, t, "Radius", &s.Round)
					ui.Box(c).Size(6, 0)
					mode := "Dark"
					if t.Dark {
						mode = "Light"
					}
					if mz.Button(c.Key("mode"), mode, mz.ButtonOptions{Size: mz.Sm}).Clicked() {
						s.Mode = mz.Dark
						if t.Dark {
							s.Mode = mz.Light
						}
					}
				})
			})
		})
		if s.Page == 7 {
			s.Studio.view(c, t)
			return
		}
		// the studio's palettes in windows of their own go with its page
		if s.Studio.Board.Host != nil {
			s.Studio.Board.Hide()
		}
		if s.Page == 6 {
			s.Work.view(c, t)
			return
		}
		if s.Page == 5 {
			s.Tables.view(c, t)
			return
		}
		if s.Page == 4 {
			s.Feedback.view(c, t)
			return
		}
		if s.Page == 3 {
			s.Forms.view(c, t)
			return
		}
		if s.Page == 2 {
			s.layers(c, t)
			return
		}
		if s.Page == 1 {
			s.more(c, t)
			return
		}
		s.buttons(c, t)
		ui.Row(c).Gap(mz.GapSection).AlignItems(ui.Start).Children(func() {
			ui.Column(c).Grow(1).Basis(0).Children(func() { s.fields(c, t) })
			ui.Column(c).Grow(1).Basis(0).Gap(mz.GapSection).Children(func() {
				s.sizes(c, t)
				s.choices(c, t)
			})
		})
	})
}

func (s *State) buttons(c *ui.Context, t *mz.Tokens) {
	cl := t.Colors
	mz.Card(c, mz.CardOptions{Title: "Button", Subtitle: "One kind of raised pill; color, fill and size.", Pad: true}, func() {
		ui.Column(c).Gap(14).Children(func() {
			ui.Row(c).Gap(mz.GapControl).Wrap().AlignItems(ui.Center).Children(func() {
				if mz.Button(c.Key("default"), "Default", mz.ButtonOptions{}).Clicked() {
					s.Count++
				}
				mz.Button(c.Key("icon"), "Search", mz.ButtonOptions{Icon: search})
				mz.Button(c.Key("iconOnly"), "Search", mz.ButtonOptions{Icon: search, IconOnly: true})
				mz.Button(c.Key("soft"), "Accent", mz.ButtonOptions{Color: cl.Accent})
				mz.Button(c.Key("solid"), "Export", mz.ButtonOptions{Color: cl.Accent, Variant: mz.Solid})
				mz.Button(c.Key("danger"), "Delete", mz.ButtonOptions{Danger: true})
				mz.Button(c.Key("dangerSoft"), "Remove", mz.ButtonOptions{Danger: true, Variant: mz.Soft})
				mz.Button(c.Key("loading"), "Saving", mz.ButtonOptions{Loading: true})
				mz.Button(c.Key("disabled"), "Disabled", mz.ButtonOptions{Disabled: true})
			})
			ui.Row(c).Gap(mz.GapControl).Wrap().Children(func() {
				for i, col := range []ui.Color{cl.C1, cl.C2, cl.C3, cl.C4, cl.C5, cl.C6, cl.C7, cl.C8} {
					mz.Button(c.Key(i), "Series", mz.ButtonOptions{Color: col, Size: mz.Sm})
				}
			})
			mz.Text(c, "Default button presses: "+itoa(s.Count), mz.TCaption).TextColor(cl.Muted)
		})
	})
}

func (s *State) fields(c *ui.Context, t *mz.Tokens) {
	mz.Card(c, mz.CardOptions{Title: "Form", Subtitle: "Inset fields and a raised trigger.", Pad: true}, func() {
		ui.Column(c).Gap(16).Children(func() {
			mz.Field(c, mz.FieldOptions{Label: "Search"}, func() {
				mz.Input(c.Key("q"), &s.Query, mz.InputOptions{Label: "Search", Placeholder: "Patients, files, reports…", PrefixIcon: search, Clearable: true})
			})
			mz.Field(c, mz.FieldOptions{Label: "Email", Error: "Enter a valid email address.", Required: true, Shake: s.shake}, func() {
				mz.Input(c.Key("email"), &s.Email, mz.InputOptions{Label: "Email", Invalid: true})
			})
			mz.Field(c, mz.FieldOptions{Label: "Password", Help: "At least 12 characters."}, func() {
				mz.Input(c.Key("pass"), &s.Pass, mz.InputOptions{Label: "Password", Password: true})
			})
			mz.Field(c, mz.FieldOptions{Label: "Period"}, func() {
				mz.Select(c.Key("range"), &s.Range, []mz.Option[int]{
					{Value: 7, Label: "Last 7 days"},
					{Value: 30, Label: "Last 30 days", Description: "default"},
					{Value: 90, Label: "Last 90 days"},
					{Value: 365, Label: "Last year", Disabled: true},
				}, mz.SelectOptions{Label: "Period"})
			})
			mz.Field(c, mz.FieldOptions{Label: "Team"}, func() {
				mz.Select(c.Key("team"), &s.Team, []mz.Option[string]{
					{Value: "web", Label: "Web", Dot: t.Colors.C1},
					{Value: "ios", Label: "iOS", Dot: t.Colors.C2},
					{Value: "and", Label: "Android", Dot: t.Colors.C3},
				}, mz.SelectOptions{Label: "Team", Placeholder: "Choose a team"})
			})
			// a check that finds the email invalid shakes its field
			if mz.Button(c.Key("check"), "Check form", mz.ButtonOptions{Size: mz.Sm}).Clicked() {
				s.shake++
			}
		})
	})
}

func (s *State) sizes(c *ui.Context, t *mz.Tokens) {
	mz.Card(c, mz.CardOptions{Title: "Size", Subtitle: "sm · md · lg: a button and a field of the same step are the same height.", Pad: true}, func() {
		ui.Column(c).Gap(14).Children(func() {
			for _, z := range []mz.Size{mz.Sm, mz.Md, mz.Lg} {
				mz.Scope(c.Key(z), mz.Options{Size: z}).Children(func() {
					ui.Row(c).Gap(mz.GapControl).AlignItems(ui.Center).Children(func() {
						ui.Column(c).Grow(1).Children(func() {
							mz.Input(c.Key("in"), &s.Query, mz.InputOptions{Label: "Search", Placeholder: "Search"})
						})
						mz.Button(c.Key("b"), "Apply", mz.ButtonOptions{Color: t.Colors.Accent, Variant: mz.Solid})
					})
				})
			}
		})
	})
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for ; n > 0; n /= 10 {
		b = append([]byte{byte('0' + n%10)}, b...)
	}
	return string(b)
}

// accents are the brand colors the header offers, Ryugo's blue first.
var accents = []struct {
	Name  string
	Color ui.Color
}{
	{"Blue", ui.Color{}},
	{"Violet", ui.Oklch(0.56, 0.2, 295)},
	{"Green", ui.Oklch(0.6, 0.15, 155)},
	{"Orange", ui.Oklch(0.68, 0.18, 50)},
	{"Rose", ui.Oklch(0.6, 0.2, 10)},
}

// accents builds the brand color swatches: a ring marks the one chosen.
// setting is a small, muted label and a compact S / M / L choice.
func setting(c *ui.Context, t *mz.Tokens, label string, at *int) {
	ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
		ui.Text(c, label).FontSize(mz.FsXs).FontWeight(500).TextColor(t.Colors.Muted).NoWrap()
		mz.Segmented(c.Key(label), at, []string{"S", "M", "L"}, mz.SegmentedOptions{Label: label, Size: mz.Sm})
	})
}

func (s *State) accents(c *ui.Context, t *mz.Tokens) {
	ui.RadioGroup(c, func() {
		ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
			for _, a := range accents {
				sw := ui.RadioBase(c.Key(a.Name), &s.Accent, a.Color).Label(a.Name).Size(22, 22).Radius(11).Center().
					Cursor(ui.CursorPointer).FocusRing(false)
				// the swatch shows the accent the color gives, not the color itself
				shown := mz.Swatch(t, a.Color)
				sw.Children(func() { ui.Box(c).Size(16, 16).Radius(8).Background(shown) })
				if s.Accent == a.Color {
					sw.Border(1.5, shown)
				} else if sw.Hovered() {
					sw.Border(1.5, t.Colors.Hairline)
				}
			}
		})
	}).Label("Accent")
}
