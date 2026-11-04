package main

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/ryu-go"
)

func init() {
	notes, bio := "Follow up with the cardiology team on Thursday.", ""
	demo("textarea", func(c *ui.Context, t *mz.Tokens) {
		ui.Column(c).Gap(16).Width(360).Children(func() {
			mz.Field(c, mz.FieldOptions{Label: "Notes", Help: "Grows up to six lines, then scrolls."}, func() {
				mz.Textarea(c.Key("notes"), &notes, mz.TextareaOptions{Label: "Notes", Placeholder: "Add a note…", MaxRows: 6})
			})
			mz.Field(c, mz.FieldOptions{Label: "Bio"}, func() {
				mz.Textarea(c.Key("bio"), &bio, mz.TextareaOptions{Label: "Bio", Placeholder: "A few words about you", Rows: 2, MaxLength: 80, Count: true})
			})
		})
	})

	qty, price, dose := 2.0, 24.5, 5.0
	qtyErr, priceErr := "", ""
	demo("number-input", func(c *ui.Context, t *mz.Tokens) {
		ui.Column(c).Gap(16).Width(400).Children(func() {
			ui.Row(c).Gap(12).AlignItems(ui.Start).Children(func() {
				ui.Column(c).Grow(1).Basis(0).MinWidth(0).Children(func() {
					mz.Field(c, mz.FieldOptions{Label: "Quantity", Help: "1 to 10", Error: qtyErr}, func() {
						mz.NumberInput(c.Key("qty"), &qty, mz.NumberInputOptions{Label: "Quantity", Min: 1, Max: 10, Error: &qtyErr})
					})
				})
				ui.Column(c).Grow(1).Basis(0).MinWidth(0).Children(func() {
					mz.Field(c, mz.FieldOptions{Label: "Price", Error: priceErr}, func() {
						mz.NumberInput(c.Key("price"), &price, mz.NumberInputOptions{Label: "Price", Min: 0, Max: 1000, Step: 0.5, Prefix: "$", Error: &priceErr})
					})
				})
			})
			ui.Row(c).Gap(12).AlignItems(ui.End).Children(func() {
				for _, z := range []mz.Size{mz.Sm, mz.Md, mz.Lg} {
					ui.Column(c.Key(z)).Grow(1).Basis(0).MinWidth(0).Children(func() {
						mz.NumberInput(c, &dose, mz.NumberInputOptions{Label: "Dose", Min: 0, Max: 50, Suffix: "mg", Size: z})
					})
				}
			})
		})
	})

	assignee, country := "leo", ""
	demo("combobox", func(c *ui.Context, t *mz.Tokens) {
		cl := t.Colors
		people := []mz.Option[string]{
			{Value: "maya", Label: "Maya Chen", Description: "Design", Dot: cl.C1},
			{Value: "leo", Label: "Leo Park", Description: "Web", Dot: cl.C2},
			{Value: "sam", Label: "Sam Rivera", Description: "iOS", Dot: cl.C3},
			{Value: "noah", Label: "Noah Kim", Description: "Android", Dot: cl.C4},
			{Value: "ivy", Label: "Ivy Brooks", Description: "On leave", Disabled: true},
		}
		ui.Column(c).Gap(16).Width(320).Children(func() {
			mz.Field(c, mz.FieldOptions{Label: "Assignee"}, func() {
				mz.Combobox(c.Key("assignee"), &assignee, people, mz.ComboboxOptions{Label: "Assignee", Placeholder: "Assign to…"})
			})
			mz.Field(c, mz.FieldOptions{Label: "Country"}, func() {
				mz.Combobox(c.Key("country"), &country, countries, mz.ComboboxOptions{Label: "Country", Placeholder: "Search countries"})
			})
		})
	})

	var start, due time.Time
	demo("date-picker", func(c *ui.Context, t *mz.Tokens) {
		if start.IsZero() {
			n := c.Now()
			start = time.Date(n.Year(), n.Month(), n.Day(), 9, 0, 0, 0, time.Local)
		}
		ui.Row(c).Gap(12).Width(440).AlignItems(ui.Start).Children(func() {
			ui.Column(c).Grow(1).Basis(0).MinWidth(0).Children(func() {
				mz.Field(c, mz.FieldOptions{Label: "Start date"}, func() {
					mz.DatePicker(c.Key("start"), &start, mz.DatePickerOptions{Label: "Start date"})
				})
			})
			ui.Column(c).Grow(1).Basis(0).MinWidth(0).Children(func() {
				mz.Field(c, mz.FieldOptions{Label: "Due date", Help: "From the start date on."}, func() {
					mz.DatePicker(c.Key("due"), &due, mz.DatePickerOptions{Label: "Due date", Min: start})
				})
			})
		})
	})

	tags, labels := []string{"cardiology", "follow-up", "urgent"}, []string(nil)
	demo("tag-input", func(c *ui.Context, t *mz.Tokens) {
		ui.Column(c).Gap(16).Width(380).Children(func() {
			mz.Field(c, mz.FieldOptions{Label: "Tags", Help: "Enter or a comma adds a tag; Backspace takes the last out."}, func() {
				mz.TagInput(c.Key("tags"), &tags, mz.TagInputOptions{Label: "Tags", Placeholder: "Add a tag"})
			})
			mz.Field(c, mz.FieldOptions{Label: "Labels"}, func() {
				mz.TagInput(c.Key("labels"), &labels, mz.TagInputOptions{Label: "Labels", Placeholder: "Up to three labels", Max: 3})
			})
		})
		status(c, t, fmt.Sprintf("%d tags, %d labels", len(tags), len(labels)))
	})
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
