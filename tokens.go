// Ryugo's design tokens, written directly as Go values. Colors are OKLCH
// (oklch(l, c, h, alpha), with l and alpha from 0 to 1 and h in degrees),
// and derived colors mix others in Oklab (Mix), as CSS color-mix does.
// Edit the values here: this file is the source of the tokens.

package ryugo

import "github.com/egoist/mygo/ui"

// Type scale, spacing and corner radii (DIPs). They do not depend on the size step.
const (
	// Axis and calendar weekday header, counter pill. The smallest step: anything smaller is unreadable.
	Fs2xs = 11
	// Footnote, tooltip, table header, badge, field help. Short monospaced code
	// is this size too (mono looks larger).
	FsXs = 12
	// Labels and dense rows: field label, chip, pagination, table of contents, step name.
	FsSm = 13
	// Controls and secondary body: button, menu, popover, alert text.
	FsMd = 14
	// Body text (html/body).
	FsBase = 15
	// Lower bound for touch list rows and text fields: iOS zooms the page when
	// you focus a field under 16px (max(var(--fs-touch), …)).
	FsTouch = 16
	// Intro paragraph (lede); card title in a narrow card.
	FsLg = 17
	// Card, dialog and sheet title.
	FsXl = 20
	// Large number field, ring chart total, value of a wide Stat.
	Fs2xl = 24
	// Stat value.
	Fs3xl = 28
	// Page title (h1); fluid, 32px on narrow screens.
	FsDisplay = 48

	// Page side gutter (.wrap, max-width 1280px).
	Gutter = 40
	// Page side gutter below 760px.
	GutterSm = 16
	// Between cards and rows (.grid gap, row margin-bottom, .gap height).
	GapSection = 16
	// Between Stat tiles, in an auto-fill minmax(210–238px, 1fr) grid.
	GapTiles = 14
	// Inner padding of the Card panel (pad=true).
	PadCard = 28
	// Card vertical padding below 760px (18px horizontally).
	PadCardSm = 20
	// Between side-by-side controls and card actions; between icon and label in a button.
	GapControl = 8
	// Between label and control (.field); between chips; distance of a popover from its trigger.
	GapField = 6
	// Inner padding of a dropdown list; options stop 5px from the edge.
	PadPop = 5

	// Card panel.
	RadiusLg = 24
	// Stat tile, results panel.
	Radius = 18
	// Popover menu, number field (md), mini KPI boxes, decision box.
	RadiusPop = 14
	// Input, Select/DatePicker trigger, segmented track, tooltip.
	RadiusControl = 12
	// Table header ends and row hover; month/year buttons in the calendar.
	RadiusRow = 10
	// Segmented pill and buttons, stepper group: 3px inside the track, so
	// concentric with it (radius-control − 3px = 9px). Derived, not fixed:
	// changing radius-control keeps the nested corner concentric.
	RadiusInner = 9
	// List option, calendar header button, logo box.
	RadiusItem = 8
	// Horizontal bar track and heatmap cell.
	RadiusTrack = 6
	// Button, chip, badge, nav link, calendar day: standalone clickable items.
	RadiusPill = 999
)

// Colors are the color roles of one theme.
type Colors struct {
	// Page background. Cards sit on top of it; pure black in the dark theme.
	Bg ui.Color
	// Fill of a focused field (input :focus) and the base of the tooltip glass.
	Surface ui.Color
	// Card panel and Stat surface. Very close to the background in the dark
	// theme; a gradient border (gradientBorder-*) separates them, not a shadow.
	Card ui.Color
	// Opaque secondary surface: the base of the color mix in heatmap cells (color-mix(..., var(--surface-2))).
	Surface2 ui.Color
	// Primary text and values, on bg, card and surface (lowest 15.4:1).
	Text ui.Color
	// Secondary body text: insight lists, tooltip row labels, description paragraphs; on bg and card (≥9.1:1).
	Text2 ui.Color
	// Labels, subtitles, units, inactive tabs. Passes 4.5:1 on bg, card and fill
	// (chip, segmented track, field addon); 4.8:1 on fill over bg in the light
	// theme.
	Muted ui.Color
	// Single-pixel dividers: the top bar's bottom line, the footer's top line,
	// the tooltip border, the number field border.
	Hairline ui.Color
	// Inset, semi-transparent tint: table header, row hover, active nav link,
	// hovered button. Not the background of controls (that's the opaque --well).
	Fill ui.Color
	// Inset control background, opaque: Input, Textarea, NumberInput, Checkbox,
	// Radio, Slider track, Segmented track, Chips. The surface beneath (card,
	// background, pattern) doesn't show through. On it, the placeholder (muted)
	// is 4.97:1 / 5.86:1 and the box border (axis) 3.15:1 / 5.22:1.
	Well ui.Color
	// Darker inset background, opaque: the track of an off Switch.
	WellStrong ui.Color
	// Resting border of an inset field (Input, Textarea, OTP cell): a thin
	// line.
	WellBorder ui.Color
	// Fill of a selected table row (a transparent tint of accent); the selected
	// counterpart of the row hover fill (fill).
	FillSelected ui.Color
	// Background the table sits on: sticky header and pinned column cells cover
	// the rows beneath with it. Redefined on the container for tables outside a
	// card.
	TableBg ui.Color
	// Alias of hairline; components use hairline directly.
	Border ui.Color
	// Not for borders: small marks that aren't opaque. Carousel dots, Sheet handle.
	BorderStrong ui.Color
	// Top (dark) line of a grooved divider: table rows, the facts/kpis columns
	// in the results panel. Always paired with sep-light.
	SepDark ui.Color
	// Bottom (light) line of a grooved divider; drawn 1px below or to the right
	// of sep-dark, which gives the line an embossed depth.
	SepLight ui.Color
	// 10% stop of the card's gradient border (top, most visible). The
	// .edge::before mask shows it only as a 1px ring.
	GradientBorder10 ui.Color
	// 25% stop of the card's gradient border.
	GradientBorder25 ui.Color
	// 75% stop of the card's gradient border.
	GradientBorder75 ui.Color
	// 90% stop of the card's gradient border (bottom, faintest). The border
	// fades from top to bottom: light comes from above.
	GradientBorder90 ui.Color
	// Light step of the accent scale; no component uses it right now.
	ColorAccent400 ui.Color
	// Focus border and ring, top stop of the selected-item gradient, drag-to-
	// zoom selection. White text on the selected gradient is 3.71:1 / 3.41:1
	// (below 4.5:1 for bold 14px).
	ColorAccent500 ui.Color
	// Bottom stop of the selected-item gradient; the value of accent in the light theme. White text 4.75:1.
	ColorAccent600 ui.Color
	// Shadow tint of the selected day button (40% via color-mix).
	ColorAccent700 ui.Color
	// The system's single accent color: today's date, selection check, insight
	// dots, link hover, default series color. As text on card it's 4.75:1
	// (light) / 5.51:1 (dark); on bg in the light theme 4.35:1, below the
	// threshold.
	Accent ui.Color
	// Top color of the raised control gradient (--control = linear-
	// gradient(180deg, control-top, control-bottom)): buttons, selected chip,
	// segmented pill, stepper.
	ControlTop ui.Color
	// Bottom color of the raised control gradient. It sits 5.5 (light) / 7
	// (dark) lightness points from the top, so the gradient is visible (at 4
	// points it looked flat, at 7 the bottom edge looked cut off); the darkest
	// band is at 72% of the bottom color (base.css --control). Text on this
	// background is also ≥ 12:1.
	ControlBottom ui.Color
	// 1px border of raised controls (0.5px on chips and the pill).
	ControlBorder ui.Color
	// Switch and Slider knob.
	Thumb ui.Color
	// Text and marks on accent and solid colors: selected day, highlighted list
	// option, checkbox check, solid button label, avatar initials.
	OnAccent ui.Color
	// Secondary text and marks on accent: the highlighted option's description,
	// shortcut and check; the ring around a series dot.
	OnAccentMuted ui.Color
	// Modal backdrop (Dialog, CommandPalette ::backdrop).
	Scrim ui.Color
	// Drawer backdrop (Sheet ::backdrop): the page stays visible beside it.
	ScrimSoft ui.Color
	// Dimming of a layer that has a modal open inside it (Dialog, Sheet ::after).
	Dim ui.Color
	// Popover menu glass, used with backdrop-filter: saturate(180%) blur(24px). Select list and calendar.
	Pop ui.Color
	// Increase (bad by default: error rate, response time, cost). As text on
	// card 6.4:1, on its own 10% badge background (over bg) 4.9:1. The badge
	// always carries a ▲; color alone carries no meaning.
	Up ui.Color
	// Decrease (good by default). In the light theme 5.7:1 on card, 4.5:1 on its
	// own 10% badge background (over bg). Little brightness difference from up
	// (1.13:1): the ▲/▼ mark is required.
	Down ui.Color
	// Warning (medium risk, pending): the readable text counterpart of c3 (same
	// hue; darker in light, lighter in dark). Badge tone="warn". On card 6.4:1 /
	// 9.1:1; on its own 10% badge background (over bg) 5.1:1 / 9.1:1.
	Warn ui.Color
	// Chart grid: dotted horizontal lines made of zero-length, round-capped dashes; x-axis ticks.
	Grid ui.Color
	// Axis labels (11px, tabular). 3.65:1 on card in the light theme (below 4.5:1); 5.81:1 in the dark theme.
	Axis ui.Color
	// Series 1; the first color of the categorical palette and the default
	// series. All eight series share one lightness (light L 58%, dark L 70%; hue
	// fixed, chroma as high as sRGB allows), so none outweighs the others. On
	// card ≥4:1 in light, ≥6.4:1 in dark (above 3:1 for graphical objects). In
	// the light theme only c4, c5 and c7 pass 4.5:1 as text.
	C1 ui.Color
	// Series 2.
	C2 ui.Color
	// Series 3. Not used as text in the light theme (4.4:1); lines, dots, fills.
	C3 ui.Color
	// Series 4.
	C4 ui.Color
	// Series 5.
	C5 ui.Color
	// Series 6. Not used as text in the light theme (4.1:1); lines, dots, fills.
	C6 ui.Color
	// Series 7.
	C7 ui.Color
	// Series 8. Not used as text in the light theme (4.2:1); lines, dots, fills.
	C8 ui.Color
}

// themeColors[theme]: theme 0 is light, 1 is dark.
var themeColors = [2]Colors{
	{
		Bg:               oklch(0.955, 0.004, 286, 1),
		Surface:          oklch(1, 0, 0, 1),
		Card:             oklch(1, 0, 0, 1),
		Surface2:         oklch(0.965, 0.004, 286, 1),
		Text:             oklch(0.23, 0.004, 286, 1),
		Text2:            oklch(0.36, 0.005, 286, 1),
		Muted:            oklch(0.48, 0.008, 286, 1),
		Hairline:         oklch(0, 0, 0, 0.09),
		Fill:             oklch(0, 0, 0, 0.055),
		Well:             oklch(0.93, 0.005, 286, 1),
		WellStrong:       oklch(0.88, 0.004, 286, 1),
		WellBorder:       oklch(0, 0, 0, 0.1),
		FillSelected:     oklch(0.56, 0.19, 256, 0.11),
		TableBg:          oklch(1, 0, 0, 1),
		Border:           oklch(0, 0, 0, 0.09),
		BorderStrong:     oklch(0, 0, 0, 0.16),
		SepDark:          oklch(0, 0, 0, 0.09),
		SepLight:         oklch(1, 0, 0, 0.8),
		GradientBorder10: oklch(0, 0, 0, 0.12),
		GradientBorder25: oklch(0, 0, 0, 0.08),
		GradientBorder75: oklch(0, 0, 0, 0.07),
		GradientBorder90: oklch(0, 0, 0, 0.06),
		ColorAccent400:   oklch(0.68, 0.17, 256, 1),
		ColorAccent500:   oklch(0.62, 0.2, 256, 1),
		ColorAccent600:   oklch(0.56, 0.19, 256, 1),
		ColorAccent700:   oklch(0.48, 0.16, 256, 1),
		Accent:           oklch(0.56, 0.19, 256, 1),
		ControlTop:       oklch(1, 0, 0, 1),
		ControlBottom:    oklch(0.935, 0.005, 286, 1),
		ControlBorder:    oklch(0, 0, 0, 0.12),
		Thumb:            oklch(1, 0, 0, 1),
		OnAccent:         oklch(1, 0, 0, 1),
		OnAccentMuted:    oklch(1, 0, 0, 0.75),
		Scrim:            oklch(0, 0, 0, 0.2),
		ScrimSoft:        oklch(0, 0, 0, 0.15),
		Dim:              oklch(0, 0, 0, 0.06),
		Pop:              oklch(1, 0, 0, 0.88),
		Up:               oklch(0.51, 0.207, 27, 1),
		Down:             oklch(0.5, 0.13, 148, 1),
		Warn:             oklch(0.49, 0.11, 68, 1),
		Grid:             oklch(0, 0, 0, 0.07),
		Axis:             oklch(0.62, 0.007, 286, 1),
		C1:               oklch(0.58, 0.198, 256, 1),
		C2:               oklch(0.58, 0.103, 182, 1),
		C3:               oklch(0.58, 0.127, 68, 1),
		C4:               oklch(0.58, 0.21, 12, 1),
		C5:               oklch(0.58, 0.21, 306, 1),
		C6:               oklch(0.58, 0.156, 130, 1),
		C7:               oklch(0.58, 0.21, 29, 1),
		C8:               oklch(0.58, 0.124, 236, 1),
	},
	{
		Bg:               oklch(0, 0, 0, 1),
		Surface:          oklch(0.23, 0.004, 286, 1),
		Card:             oklch(0.18, 0.004, 286, 1),
		Surface2:         oklch(0.29, 0.004, 286, 1),
		Text:             oklch(0.97, 0.003, 286, 1),
		Text2:            oklch(0.83, 0.007, 286, 1),
		Muted:            oklch(0.68, 0.007, 286, 1),
		Hairline:         oklch(1, 0, 0, 0.1),
		Fill:             oklch(1, 0, 0, 0.06),
		Well:             oklch(0.23, 0.004, 286, 1),
		WellStrong:       oklch(0.33, 0.004, 286, 1),
		WellBorder:       oklch(1, 0, 0, 0.1),
		FillSelected:     oklch(0.62, 0.2, 256, 0.16),
		TableBg:          oklch(0.18, 0.004, 286, 1),
		Border:           oklch(1, 0, 0, 0.1),
		BorderStrong:     oklch(1, 0, 0, 0.16),
		SepDark:          oklch(0, 0, 0, 0.7),
		SepLight:         oklch(1, 0, 0, 0.07),
		GradientBorder10: oklch(0.31, 0, 0, 0.51),
		GradientBorder25: oklch(0.32, 0, 0, 0.3),
		GradientBorder75: oklch(0.19, 0, 0, 0.3),
		GradientBorder90: oklch(0.3, 0, 0, 0.25),
		ColorAccent400:   oklch(0.7, 0.159, 256, 1),
		ColorAccent500:   oklch(0.64, 0.19, 256, 1),
		ColorAccent600:   oklch(0.56, 0.19, 256, 1),
		ColorAccent700:   oklch(0.49, 0.16, 256, 1),
		Accent:           oklch(0.64, 0.19, 256, 1),
		ControlTop:       oklch(0.35, 0.003, 286, 1),
		ControlBottom:    oklch(0.28, 0.004, 286, 1),
		ControlBorder:    oklch(1, 0, 0, 0.08),
		Thumb:            oklch(0.97, 0.003, 286, 1),
		OnAccent:         oklch(1, 0, 0, 1),
		OnAccentMuted:    oklch(1, 0, 0, 0.75),
		Scrim:            oklch(0, 0, 0, 0.55),
		ScrimSoft:        oklch(0, 0, 0, 0.5),
		Dim:              oklch(0, 0, 0, 0.35),
		Pop:              oklch(0.29, 0.004, 286, 0.9),
		Up:               oklch(0.66, 0.22, 27, 1),
		Down:             oklch(0.76, 0.21, 148, 1),
		Warn:             oklch(0.78, 0.15, 68, 1),
		Grid:             oklch(1, 0, 0, 0.08),
		Axis:             oklch(0.65, 0.007, 286, 1),
		C1:               oklch(0.7, 0.159, 256, 1),
		C2:               oklch(0.7, 0.125, 182, 1),
		C3:               oklch(0.7, 0.153, 68, 1),
		C4:               oklch(0.7, 0.195, 12, 1),
		C5:               oklch(0.7, 0.2, 306, 1),
		C6:               oklch(0.7, 0.188, 130, 1),
		C7:               oklch(0.7, 0.191, 29, 1),
		C8:               oklch(0.7, 0.15, 236, 1),
	},
}

// Shadows are the layered shadows of one theme (CSS box-shadow).
type Shadows struct {
	// Raised button: a 1px inner highlight on top + a contact shadow below.
	ControlShadow Shadow
	// Selected item lifted inside a track: selected chip, segmented pill.
	ShadowRaised Shadow
	// The +/− group on the right of the number field.
	ShadowStepper Shadow
	// Chart tooltip.
	Shadow Shadow
	// Popover menu: 0.5px outline + two layers of spread.
	ShadowPop Shadow
	// Soft ring of a focused or open field (color-accent-500 at 20%): input
	// :focus, open Select/DatePicker, number field :focus-within. Used together
	// with a color-accent-500 border.
	Ring Shadow
	// Keyboard focus on a button: a 3px gap in the background color, then 2px of color-accent-500 at 60%.
	FocusRing Shadow
	// Top highlight line of a solid-colored surface: solid button and chip, BulletChart and RatioBar bars.
	Sheen Shadow
	// Top highlight line of an accent-gradient selection mark: checkbox, radio, selected day, current step.
	SheenAccent Shadow
	// Top highlight line of a soft tinted band (RangeBar normal range).
	SheenTint Shadow
	// Inner edge of an inset track: Slider, Progress, RangeBar, BulletChart track.
	WellEdge Shadow
	// Selected Segmented button before it's measured (server render).
	ShadowRaisedSm Shadow
	// Switch knob.
	KnobShadow Shadow
	// Drop shadow of the Slider knob (also used together with the focus ring).
	KnobDrop Shadow
	// Slider knob: thin border + drop shadow.
	KnobShadowLg Shadow
}

var themeShadows = [2]Shadows{
	{
		ControlShadow:  Shadow{{X: 0, Y: 1, Blur: 0, Spread: 0, Color: oklch(1, 0, 0, 0.5), Inset: true}, {X: 0, Y: 1, Blur: 2, Spread: 0, Color: oklch(0, 0, 0, 0.06), Inset: false}},
		ShadowRaised:   Shadow{{X: 0, Y: 1, Blur: 2, Spread: 0, Color: oklch(0, 0, 0, 0.08), Inset: false}, {X: 0, Y: 3, Blur: 8, Spread: 0, Color: oklch(0, 0, 0, 0.06), Inset: false}},
		ShadowStepper:  Shadow{{X: 0, Y: 1, Blur: 2, Spread: 0, Color: oklch(0, 0, 0, 0.08), Inset: false}, {X: 0, Y: 2, Blur: 6, Spread: 0, Color: oklch(0, 0, 0, 0.06), Inset: false}},
		Shadow:         Shadow{{X: 0, Y: 1, Blur: 2, Spread: 0, Color: oklch(0, 0, 0, 0.04), Inset: false}, {X: 0, Y: 8, Blur: 24, Spread: 0, Color: oklch(0, 0, 0, 0.06), Inset: false}},
		ShadowPop:      Shadow{{X: 0, Y: 0, Blur: 0, Spread: 0.5, Color: oklch(0, 0, 0, 0.06), Inset: false}, {X: 0, Y: 2, Blur: 6, Spread: 0, Color: oklch(0, 0, 0, 0.08), Inset: false}, {X: 0, Y: 14, Blur: 36, Spread: 0, Color: oklch(0, 0, 0, 0.16), Inset: false}},
		Ring:           Shadow{{X: 0, Y: 0, Blur: 0, Spread: 4, Color: oklch(0.62, 0.2, 256, 0.2), Inset: false}},
		FocusRing:      Shadow{{X: 0, Y: 0, Blur: 0, Spread: 3, Color: oklch(0.97, 0.003, 286, 1), Inset: false}, {X: 0, Y: 0, Blur: 0, Spread: 5, Color: oklch(0.62, 0.2, 256, 0.6), Inset: false}},
		Sheen:          Shadow{{X: 0, Y: 1, Blur: 0, Spread: 0, Color: oklch(1, 0, 0, 0.25), Inset: true}},
		SheenAccent:    Shadow{{X: 0, Y: 1, Blur: 0, Spread: 0, Color: oklch(1, 0, 0, 0.28), Inset: true}},
		SheenTint:      Shadow{{X: 0, Y: 1, Blur: 0, Spread: 0, Color: oklch(1, 0, 0, 0.3), Inset: true}},
		WellEdge:       Shadow{{X: 0, Y: 0, Blur: 0, Spread: 1, Color: oklch(0, 0, 0, 0.08), Inset: true}},
		ShadowRaisedSm: Shadow{{X: 0, Y: 1, Blur: 2, Spread: 0, Color: oklch(0, 0, 0, 0.08), Inset: false}},
		KnobShadow:     Shadow{{X: 0, Y: 0, Blur: 0, Spread: 0.5, Color: oklch(0, 0, 0, 0.04), Inset: false}, {X: 0, Y: 2, Blur: 4, Spread: 0, Color: oklch(0, 0, 0, 0.15), Inset: false}, {X: 0, Y: 1, Blur: 1, Spread: 0, Color: oklch(0, 0, 0, 0.06), Inset: false}},
		KnobDrop:       Shadow{{X: 0, Y: 2, Blur: 6, Spread: 0, Color: oklch(0, 0, 0, 0.2), Inset: false}},
		KnobShadowLg:   Shadow{{X: 0, Y: 0, Blur: 0, Spread: 0.5, Color: oklch(0, 0, 0, 0.1), Inset: false}, {X: 0, Y: 2, Blur: 6, Spread: 0, Color: oklch(0, 0, 0, 0.2), Inset: false}},
	},
	{
		ControlShadow:  Shadow{{X: 0, Y: 1, Blur: 0, Spread: 0, Color: oklch(1, 0, 0, 0.08), Inset: true}, {X: 0, Y: 1, Blur: 2, Spread: 0, Color: oklch(0, 0, 0, 0.4), Inset: false}},
		ShadowRaised:   Shadow{{X: 0, Y: 1, Blur: 2, Spread: 0, Color: oklch(0, 0, 0, 0.08), Inset: false}, {X: 0, Y: 3, Blur: 8, Spread: 0, Color: oklch(0, 0, 0, 0.06), Inset: false}},
		ShadowStepper:  Shadow{{X: 0, Y: 1, Blur: 2, Spread: 0, Color: oklch(0, 0, 0, 0.08), Inset: false}, {X: 0, Y: 2, Blur: 6, Spread: 0, Color: oklch(0, 0, 0, 0.06), Inset: false}},
		Shadow:         Shadow{{X: 0, Y: 1, Blur: 2, Spread: 0, Color: oklch(0, 0, 0, 0.4), Inset: false}, {X: 0, Y: 8, Blur: 24, Spread: 0, Color: oklch(0, 0, 0, 0.4), Inset: false}},
		ShadowPop:      Shadow{{X: 0, Y: 0, Blur: 0, Spread: 0.5, Color: oklch(1, 0, 0, 0.08), Inset: false}, {X: 0, Y: 2, Blur: 6, Spread: 0, Color: oklch(0, 0, 0, 0.4), Inset: false}, {X: 0, Y: 14, Blur: 36, Spread: 0, Color: oklch(0, 0, 0, 0.67), Inset: false}},
		Ring:           Shadow{{X: 0, Y: 0, Blur: 0, Spread: 4, Color: oklch(0.64, 0.19, 256, 0.2), Inset: false}},
		FocusRing:      Shadow{{X: 0, Y: 0, Blur: 0, Spread: 3, Color: oklch(0, 0, 0, 1), Inset: false}, {X: 0, Y: 0, Blur: 0, Spread: 5, Color: oklch(0.64, 0.19, 256, 0.6), Inset: false}},
		Sheen:          Shadow{{X: 0, Y: 1, Blur: 0, Spread: 0, Color: oklch(1, 0, 0, 0.25), Inset: true}},
		SheenAccent:    Shadow{{X: 0, Y: 1, Blur: 0, Spread: 0, Color: oklch(1, 0, 0, 0.28), Inset: true}},
		SheenTint:      Shadow{{X: 0, Y: 1, Blur: 0, Spread: 0, Color: oklch(1, 0, 0, 0.3), Inset: true}},
		WellEdge:       Shadow{{X: 0, Y: 0, Blur: 0, Spread: 1, Color: oklch(1, 0, 0, 0.1), Inset: true}},
		ShadowRaisedSm: Shadow{{X: 0, Y: 1, Blur: 2, Spread: 0, Color: oklch(0, 0, 0, 0.08), Inset: false}},
		KnobShadow:     Shadow{{X: 0, Y: 0, Blur: 0, Spread: 0.5, Color: oklch(0, 0, 0, 0.04), Inset: false}, {X: 0, Y: 2, Blur: 4, Spread: 0, Color: oklch(0, 0, 0, 0.15), Inset: false}, {X: 0, Y: 1, Blur: 1, Spread: 0, Color: oklch(0, 0, 0, 0.06), Inset: false}},
		KnobDrop:       Shadow{{X: 0, Y: 2, Blur: 6, Spread: 0, Color: oklch(0, 0, 0, 0.2), Inset: false}},
		KnobShadowLg:   Shadow{{X: 0, Y: 0, Blur: 0, Spread: 0.5, Color: oklch(0, 0, 0, 0.1), Inset: false}, {X: 0, Y: 2, Blur: 6, Spread: 0, Color: oklch(0, 0, 0, 0.2), Inset: false}},
	},
}

// Metrics are the measurements of one size step (DIPs). A zero font size is
// inherited from the surroundings (initial in CSS); letter spacing (…Ls) is
// in em.
type Metrics struct {
	// Single height for every control that sits on a line: Button, Input,
	// Select, DatePicker / DateRangePicker trigger, TagInput, Segmented rail.
	// Buttons and fields next to each other are always the same height.
	ControlH float32
	// Button height (= control-h); also the width of an icon button.
	BtnH float32
	// Horizontal padding of a button.
	BtnPx float32
	// Button label.
	BtnFs float32
	// Between icon and label in a button.
	BtnGap float32
	// Field height (= control-h): Input, Select, DatePicker and DateRangePicker trigger.
	FieldH float32
	// Left padding of a field.
	FieldPs float32
	// Right padding of a field (the arrow or icon side).
	FieldPe float32
	// Field corner.
	FieldR float32
	// Text of the Select and date triggers.
	FieldFs float32
	// Input text; inherited from the surroundings at md (initial).
	InputFs float32
	// Between addon, field and buttons in an Input with addons.
	InputGap float32
	// Clear / show button inside an Input.
	InputBtn float32
	// Chip height.
	ChipH float32
	// Chip horizontal padding.
	ChipPx float32
	// Chip label.
	ChipFs float32
	// Between dot / icon and label inside a chip.
	ChipGap float32
	// Between chips.
	ChipsGap float32
	// Corner of a rounded chip.
	ChipR float32
	// The × button on a filter chip.
	ChipX float32
	// Vertical padding of a Segmented option.
	SegPy float32
	// Horizontal padding of a Segmented option.
	SegPx float32
	// Text of a Segmented option.
	SegFs float32
	// NumberInput height.
	NumH float32
	// NumberInput left padding.
	NumPs float32
	// NumberInput corner.
	NumR float32
	// NumberInput digits.
	NumFs float32
	// NumberInput letter spacing (tighter at larger sizes).
	NumLs float32
	// NumberInput unit.
	NumPrefixFs float32
	// Width of the NumberInput +/− group.
	NumStepW float32
	// Tab height.
	TabH float32
	// Tab horizontal padding.
	TabPx float32
	// Tab text.
	TabFs float32
	// Table cell vertical padding.
	CellPy float32
	// Table cell horizontal padding; the row divider starts this far in.
	CellPx float32
	// Vertical padding of the table header row.
	HeadPy float32
	// Width of the table control column (selection box, expand button); also added to pinned column offsets.
	CellControlW float32
	// Clickable width of the column resize handle.
	ResizeHit float32
	// Pagination button height and minimum width.
	PgH float32
	// NavList row height (at least 44px on a coarse pointer).
	NavH float32
	// NavList row horizontal padding; the group heading starts this far in.
	NavPx float32
	// NavList row text.
	NavFs float32
	// Between icon, label and badge in a NavList row.
	NavGap float32
}

// sizeMetrics[step]: 0 is sm, 1 md, 2 lg.
var sizeMetrics = [3]Metrics{
	{
		ControlH:     30,
		BtnH:         30,
		BtnPx:        12,
		BtnFs:        13,
		BtnGap:       6,
		FieldH:       30,
		FieldPs:      10,
		FieldPe:      8,
		FieldR:       10,
		FieldFs:      13,
		InputFs:      13,
		InputGap:     6,
		InputBtn:     20,
		ChipH:        26,
		ChipPx:       10,
		ChipFs:       12,
		ChipGap:      6,
		ChipsGap:     4,
		ChipR:        8,
		ChipX:        18,
		SegPy:        3,
		SegPx:        10,
		SegFs:        12,
		NumH:         36,
		NumPs:        12,
		NumR:         12,
		NumFs:        14,
		NumLs:        0,
		NumPrefixFs:  13,
		NumStepW:     26,
		TabH:         34,
		TabPx:        10,
		TabFs:        13,
		CellPy:       8,
		CellPx:       10,
		HeadPy:       8,
		CellControlW: 32,
		ResizeHit:    8,
		PgH:          28,
		NavH:         30,
		NavPx:        8,
		NavFs:        13,
		NavGap:       8,
	},
	{
		ControlH:     40,
		BtnH:         40,
		BtnPx:        16,
		BtnFs:        14,
		BtnGap:       8,
		FieldH:       40,
		FieldPs:      12,
		FieldPe:      10,
		FieldR:       12,
		FieldFs:      14,
		InputFs:      0,
		InputGap:     8,
		InputBtn:     24,
		ChipH:        30,
		ChipPx:       12,
		ChipFs:       13,
		ChipGap:      7,
		ChipsGap:     6,
		ChipR:        10,
		ChipX:        22,
		SegPy:        4,
		SegPx:        12,
		SegFs:        13,
		NumH:         44,
		NumPs:        14,
		NumR:         14,
		NumFs:        15,
		NumLs:        0,
		NumPrefixFs:  15,
		NumStepW:     30,
		TabH:         40,
		TabPx:        12,
		TabFs:        14,
		CellPy:       12,
		CellPx:       12,
		HeadPy:       10,
		CellControlW: 40,
		ResizeHit:    10,
		PgH:          32,
		NavH:         36,
		NavPx:        10,
		NavFs:        14,
		NavGap:       10,
	},
	{
		ControlH:     44,
		BtnH:         44,
		BtnPx:        18,
		BtnFs:        14,
		BtnGap:       8,
		FieldH:       44,
		FieldPs:      14,
		FieldPe:      12,
		FieldR:       12,
		FieldFs:      15,
		InputFs:      15,
		InputGap:     8,
		InputBtn:     26,
		ChipH:        34,
		ChipPx:       14,
		ChipFs:       14,
		ChipGap:      8,
		ChipsGap:     6,
		ChipR:        10,
		ChipX:        24,
		SegPy:        6,
		SegPx:        14,
		SegFs:        14,
		NumH:         56,
		NumPs:        16,
		NumR:         16,
		NumFs:        24,
		NumLs:        -0.02,
		NumPrefixFs:  17,
		NumStepW:     34,
		TabH:         44,
		TabPx:        14,
		TabFs:        15,
		CellPy:       14,
		CellPx:       14,
		HeadPy:       12,
		CellControlW: 44,
		ResizeHit:    12,
		PgH:          36,
		NavH:         40,
		NavPx:        12,
		NavFs:        15,
		NavGap:       12,
	},
}

// Text styles.
var (
	// Page title (h1). Fluid: clamp(32px, 4vw, 48px); 32px below 760px.
	TDisplay = TypeStyle{Size: 48, LineHeight: 1.05, Weight: 700, Tracking: -0.03, Mono: false}
	// Value of a Stat tile.
	TFigure = TypeStyle{Size: 28, LineHeight: 1.1, Weight: 700, Tracking: -0.03, Mono: false}
	// Text of the large number field (NumberInput size=lg).
	TInputLg = TypeStyle{Size: 24, LineHeight: 1.4, Weight: 600, Tracking: -0.02, Mono: false}
	// Card, dialog and sheet title; 17px (fs-lg) in a narrow card.
	TTitle = TypeStyle{Size: 20, LineHeight: 1.5, Weight: 600, Tracking: -0.02, Mono: false}
	// Page subtitle, muted, at most 70ch; 15px below 760px.
	TLede = TypeStyle{Size: 17, LineHeight: 1.5, Weight: 400, Tracking: 0, Mono: false}
	// Default text; set at the html/body level.
	TBody = TypeStyle{Size: 15, LineHeight: 1.5, Weight: 400, Tracking: -0.01, Mono: false}
	// Button label (.btn). Select/DatePicker triggers use weight 500; nav links fs-md 500.
	TControl = TypeStyle{Size: 14, LineHeight: 1.5, Weight: 600, Tracking: -0.01, Mono: false}
	// Field label (.field > .field-label), Stat label, chips and small segmented, in the muted color.
	TLabel = TypeStyle{Size: 13, LineHeight: 1.5, Weight: 500, Tracking: 0, Mono: false}
	// Footnotes, Stat subline, footer, table header (12px 500). Badges 12px 600.
	TCaption = TypeStyle{Size: 12, LineHeight: 1.5, Weight: 400, Tracking: 0, Mono: false}
	// Chart axis labels and calendar weekday headers; axis color, tabular-nums.
	TTick = TypeStyle{Size: 11, LineHeight: 1.2, Weight: 400, Tracking: 0, Mono: false}
	// Short codes (region, platform) in tables and bar lists; mono has a large
	// x-height, so 12px, aligned with min-width 38px.
	TCode = TypeStyle{Size: 12, LineHeight: 1.5, Weight: 700, Tracking: 0, Mono: true}
)
