# Mizan

Mizan is a design system with weight: controls lit from above, wells
pressed into the surface, cards that lift, lean and land, checks that draw
themselves. This is its native edition: components for [MyGo](https://mygo.egoist.dev/docs/ui)'s native UI
(package `ui`), drawn by MyGo itself on macOS, Linux, Windows and in the
browser, in pure Go without cgo. Every preview on this site is the real
component, compiled to WebAssembly: click, type and drag in it.

It ports Mizan's visual language, not its web code: the same OKLCH tokens,
the same two surfaces (a raised control that takes the light from above, an
inset well you type into), the same size steps and type scale. Behavior
(focus, keyboard, accessibility) comes from MyGo's unstyled `…Base` widgets;
Mizan gives them the look.

## Install

Coming soon: Mizan's package isn't published yet. Its first release will
say here how to add it to an app.

## Use

Call `Root` at the start of your view, every frame, and `Toaster` once
wherever in it. Everything else is a function you call where the component
goes:

```go
import (
	"github.com/egoist/mygo/ui"
	mizan "github.com/imhalid/mizan-go"
)

func (a *app) view(c *ui.Context) {
	t := mizan.Root(c, mizan.Options{})
	mizan.Card(c, mizan.CardOptions{Title: "Report", Pad: true}, func() {
		mizan.Field(c, mizan.FieldOptions{Label: "Search"}, func() {
			mizan.Input(c.Key("q"), &a.query, mizan.InputOptions{Label: "Search", Clearable: true})
		})
		if mizan.Button(c.Key("go"), "Apply", mizan.ButtonOptions{Color: t.Colors.Accent, Variant: mizan.Solid}).Clicked() {
			a.apply()
		}
	})
	mizan.Toaster(c)
}
```

A component keeps no state of yours: it reads and writes the value you
pass a pointer to (`&a.query`), and reports what happened in the frame it
happened (`Clicked`, `Changed`). Give repeated components keys with
`c.Key`, as everywhere in MyGo.

## Guides

- [Theming](theming.md): light and dark, size steps, corners and the accent, for the window or a region.

## Controls

- [Button](button.md): one kind of raised pill, in any color, fill and size.
- [Copy button](copy-button.md): a round button that copies text, a check drawing itself in reply.
- [Input](input.md): the inset text field, with icons, affixes, clearing and passwords.
- [Select](select.md): a choice from a list on glass, without search.
- [Field](field.md): a label, a control and its help or error, shaking when asked.
- [Card](card.md): a surface whose title stands above it.

## Selection

- [Checkbox](checkbox.md): checked, unchecked or mixed, its check drawn as a pen would.
- [Radio](radio.md): one of a few, its dot opening from the middle.
- [Choice](choice.md): radio buttons shaped as cards.
- [Switch](switch.md): on and off, its knob stretching while held.
- [Segmented](segmented.md): a choice between views, on a sliding pill.
- [Slider](slider.md): a value in a range, its knob settling onto steps.

## Forms

- [Textarea](textarea.md): many lines, growing with what is typed.
- [Number input](number-input.md): digits only, with a stepper that repeats while held.
- [Combobox](combobox.md): a select you type into to filter.
- [Date picker](date-picker.md): a month's calendar below a field.
- [Tag input](tag-input.md): words as pills in a field.

## Navigation

- [Tabs](tabs.md): an underline that slides to the chosen tab.
- [Link](link.md): a link in text, or a button that reads as one.
- [Pagination](pagination.md): pages of rows, a pill sliding between numbers.

## Disclosure

- [Collapsible](collapsible.md): a section that grows into view.
- [Accordion](accordion.md): a stack of sections, one or many open.

## Toolbars

- [Toggle](toggle.md): a button that stays pressed, alone or in a toolbar.
- [Separator](separator.md): a groove between content, across or along.

## Layers

- [Dialog](dialog.md): a modal panel over a scrim, dragged by its header.
- [Sheet](sheet.md): a panel along the window's edge.
- [Popover](popover.md): a panel of glass below its anchor.
- [Menu](menu.md): a list of actions, from any button or a menu button.
- [Tooltip](tooltip.md): a raised label following the pointer along its element.
- [Toast](toast.md): notices in the corner, landing with a beam of light.

## Chat

- [Chat](chat.md): a conversation that grows message by message, and its composer.

## Drag and drop

- [Sortable](sortable.md): rows reordered by dragging, or with the keyboard.
- [Drag and drop](drag-and-drop.md): cards between the columns of a board.
- [Drop zone](drop-zone.md): files from other apps, and the rows of what was added.

## Indicators

- [Badge](badge.md): a tinted pill for a status or a change, and the tag beside a title.
- [Status dot](status-dot.md): live, delayed or off.
- [Kbd](kbd.md): keyboard keys and shortcuts.
- [Avatar](avatar.md): initials or a picture, and whether someone is online.
- [Empty state](empty-state.md): what an empty list says, and what to do.

## Feedback

- [Progress](progress.md): work of known or unknown length.
- [Meter](meter.md): how full something is, turning toward warning.
- [Spinner](spinner.md): a ring for a short wait.
- [Skeleton](skeleton.md): placeholders where content is still loading.

## Data

- [Table](table.md): sorting, moving and resizing columns, selection, details, pages, all virtualized.

## Layout

- [Workspace](workspace.md): experimental: a Dock of areas that split, join and take docked tabs, as Blender's.
