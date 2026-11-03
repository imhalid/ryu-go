package gallery

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/mizan-go"
)

// formsPage is the state of the forms page: Textarea, NumberInput,
// Combobox, DatePicker and TagInput.
type formsPage struct {
	ready bool

	Notes, Bio      string
	Quantity, Price float64
	// what is wrong with the number being typed, for the fields
	QuantityErr, PriceErr string
	Dose                  float64
	Assignee              string
	Country               string
	Due, Start            time.Time
	Tags, Labels          []string
}

// init gives the page its sample values, the first time it shows.
func (f *formsPage) init(now time.Time) {
	f.ready = true
	f.Notes = "Follow up with the cardiology team on Thursday."
	f.Quantity, f.Price, f.Dose = 2, 24.5, 5
	f.Assignee = "leo"
	// a day of this month other than today, so that both show
	d := now.AddDate(0, 0, 5)
	if now.Day() > 20 {
		d = now.AddDate(0, 0, -5)
	}
	f.Start = time.Date(d.Year(), d.Month(), d.Day(), 9, 0, 0, 0, time.Local)
	f.Tags = []string{"cardiology", "follow-up", "urgent"}
}

var countries = []mz.Option[string]{
	{Value: "de", Label: "Germany"},
	{Value: "fr", Label: "France"},
	{Value: "jp", Label: "Japan"},
	{Value: "nl", Label: "Netherlands"},
	{Value: "no", Label: "Norway"},
	{Value: "tr", Label: "Türkiye"},
	{Value: "uk", Label: "United Kingdom"},
	{Value: "us", Label: "United States"},
}

// view builds the forms page.
func (f *formsPage) view(c *ui.Context, t *mz.Tokens) {
	if !f.ready {
		f.init(c.Now())
	}
	cl := t.Colors
	people := []mz.Option[string]{
		{Value: "maya", Label: "Maya Chen", Description: "Design", Dot: cl.C1},
		{Value: "leo", Label: "Leo Park", Description: "Web", Dot: cl.C2},
		{Value: "sam", Label: "Sam Rivera", Description: "iOS", Dot: cl.C3},
		{Value: "noah", Label: "Noah Kim", Description: "Android", Dot: cl.C4},
		{Value: "ivy", Label: "Ivy Brooks", Description: "On leave", Disabled: true},
	}
	ui.Row(c).Gap(mz.GapSection).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Grow(1).Basis(0).Gap(mz.GapSection).Children(func() {
			mz.Card(c, mz.CardOptions{Title: "Text", Subtitle: "Wells that grow with what is typed.", Pad: true}, func() {
				ui.Column(c).Gap(16).Children(func() {
					mz.Field(c, mz.FieldOptions{Label: "Notes", Help: "Grows up to six lines, then scrolls."}, func() {
						mz.Textarea(c.Key("notes"), &f.Notes, mz.TextareaOptions{Label: "Notes", Placeholder: "Add a note…", MaxRows: 6})
					})
					mz.Field(c, mz.FieldOptions{Label: "Bio"}, func() {
						mz.Textarea(c.Key("bio"), &f.Bio, mz.TextareaOptions{Label: "Bio", Placeholder: "A few words about you", Rows: 2, MaxLength: 80})
					})
					mz.Field(c, mz.FieldOptions{Label: "Tags", Help: "Enter or a comma adds a tag; Backspace takes the last out."}, func() {
						mz.TagInput(c.Key("tags"), &f.Tags, mz.TagInputOptions{Label: "Tags", Placeholder: "Add a tag"})
					})
					mz.Field(c, mz.FieldOptions{Label: "Labels"}, func() {
						mz.TagInput(c.Key("labels"), &f.Labels, mz.TagInputOptions{Label: "Labels", Placeholder: "Up to three labels", Max: 3})
					})
				})
			})
			mz.Card(c, mz.CardOptions{Title: "Numbers", Subtitle: "Bold tabular digits beside a raised stepper.", Pad: true}, func() {
				ui.Column(c).Gap(16).Children(func() {
					ui.Row(c).Gap(12).AlignItems(ui.Start).Children(func() {
						ui.Column(c).Grow(1).Basis(0).MinWidth(0).Children(func() {
							mz.Field(c, mz.FieldOptions{Label: "Quantity", Help: "1 to 10", Error: f.QuantityErr}, func() {
								mz.NumberInput(c.Key("qty"), &f.Quantity, mz.NumberInputOptions{Label: "Quantity", Min: 1, Max: 10, Error: &f.QuantityErr})
							})
						})
						ui.Column(c).Grow(1).Basis(0).MinWidth(0).Children(func() {
							mz.Field(c, mz.FieldOptions{Label: "Price", Error: f.PriceErr}, func() {
								mz.NumberInput(c.Key("price"), &f.Price, mz.NumberInputOptions{Label: "Price", Min: 0, Max: 1000, Step: 0.5, Prefix: "$", Error: &f.PriceErr})
							})
						})
					})
					ui.Row(c).Gap(12).AlignItems(ui.End).Children(func() {
						for _, z := range []mz.Size{mz.Sm, mz.Md, mz.Lg} {
							ui.Column(c.Key(z)).Grow(1).Basis(0).MinWidth(0).Children(func() {
								mz.NumberInput(c, &f.Dose, mz.NumberInputOptions{Label: "Dose", Min: 0, Max: 50, Suffix: "mg", Size: z})
							})
						}
					})
				})
			})
		})
		ui.Column(c).Grow(1).Basis(0).Gap(mz.GapSection).Children(func() {
			mz.Card(c, mz.CardOptions{Title: "Choose", Subtitle: "Type to filter; the list and the calendar float on the pop surface.", Pad: true}, func() {
				ui.Column(c).Gap(16).Children(func() {
					mz.Field(c, mz.FieldOptions{Label: "Assignee"}, func() {
						mz.Combobox(c.Key("assignee"), &f.Assignee, people, mz.ComboboxOptions{Label: "Assignee", Placeholder: "Assign to…"})
					})
					mz.Field(c, mz.FieldOptions{Label: "Country"}, func() {
						mz.Combobox(c.Key("country"), &f.Country, countries, mz.ComboboxOptions{Label: "Country", Placeholder: "Search countries"})
					})
					ui.Row(c).Gap(12).AlignItems(ui.Start).Children(func() {
						ui.Column(c).Grow(1).Basis(0).MinWidth(0).Children(func() {
							mz.Field(c, mz.FieldOptions{Label: "Start date"}, func() {
								mz.DatePicker(c.Key("start"), &f.Start, mz.DatePickerOptions{Label: "Start date"})
							})
						})
						ui.Column(c).Grow(1).Basis(0).MinWidth(0).Children(func() {
							mz.Field(c, mz.FieldOptions{Label: "Due date", Help: "From the start date on."}, func() {
								mz.DatePicker(c.Key("due"), &f.Due, mz.DatePickerOptions{Label: "Due date", Min: f.Start})
							})
						})
					})
				})
			})
			mz.Card(c, mz.CardOptions{Title: "States", Subtitle: "Invalid and disabled, as every field shows them.", Pad: true}, func() {
				ui.Column(c).Gap(16).Children(func() {
					mz.Field(c, mz.FieldOptions{Label: "Reviewer", Error: "Choose a reviewer."}, func() {
						mz.Combobox(c.Key("reviewer"), new(string), people, mz.ComboboxOptions{Label: "Reviewer", Placeholder: "Choose…", Invalid: true})
					})
					ui.Row(c).Gap(12).AlignItems(ui.Start).Children(func() {
						ui.Column(c).Grow(1).Basis(0).MinWidth(0).Children(func() {
							mz.Field(c, mz.FieldOptions{Label: "Seats"}, func() {
								mz.NumberInput(c.Key("seats"), new(float64), mz.NumberInputOptions{Label: "Seats", Disabled: true})
							})
						})
						ui.Column(c).Grow(1).Basis(0).MinWidth(0).Children(func() {
							mz.Field(c, mz.FieldOptions{Label: "Renewal"}, func() {
								mz.DatePicker(c.Key("renewal"), new(time.Time), mz.DatePickerOptions{Label: "Renewal", Disabled: true})
							})
						})
					})
					summary := fmt.Sprintf("%d tags · %g × $%.2f · %s", len(f.Tags), f.Quantity, f.Price, or(f.Assignee, "unassigned"))
					mz.Text(c, summary, mz.TCaption).TextColor(cl.Muted)
				})
			})
		})
	})
}
