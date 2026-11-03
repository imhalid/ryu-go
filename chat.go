package mizan

import (
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

// Mizan has no chat; its language is set here: your own message wears the
// solid button's skin (accent gradient, white text), the other side's the
// second surface and a hairline border. Labels look like field labels
// (fs-sm 500 muted), times are tabular. Consecutive messages from the same
// person form a group: the name only on the first, the avatar and the tail
// corner only on the last.

var (
	iconSend        = Svg(16, 16, "M8 12.5v-9M4.5 7L8 3.5 11.5 7")
	iconAttach      = Svg(16, 16, "M13 7.5l-5.2 5.2a3.2 3.2 0 0 1-4.5-4.5l5.5-5.5a2.1 2.1 0 0 1 3 3L6.3 11.2a1 1 0 0 1-1.5-1.5L9.5 5")
	iconSent        = Svg(16, 13, "M3.5 8.5l3 3 6-7")
	iconRead        = Svg(16, 13, "M1.5 8.5l3 3 6-7M8 11l.5.5 6-7")
	iconPending     = Svg(16, 12, "M8 4.5V8l2 1.5M14 8A6 6 0 1 1 2 8a6 6 0 0 1 12 0z")
	bubbleR, tailR  = float32(18), float32(6)
	bubbleAvatarDia = float32(28)
)

// Delivery is the delivery status of your own message.
type Delivery uint8

const (
	DeliveryNone Delivery = iota
	// Sending shows a clock icon.
	Sending
	// Sent shows a single check.
	Sent
	// Read shows a double check in the accent color.
	Read
)

// ChatMessage is a message in a chat.
type ChatMessage struct {
	Author string
	Mine   bool
	Text   string
	// Body, if set, builds the bubble's contents instead of Text (Markdown, images…).
	Body   func()
	Time   string
	Status Delivery
	// Avatar is the other side's picture and color (from the name if unset).
	Avatar AvatarOptions
}

// Group is a message's place in its group: the first message carries the
// name, the last the avatar and the tail. GroupOf computes it.
type Group struct{ First, Last bool }

// GroupOf gives msgs[i]'s place in its group: consecutive messages from the
// same author form a group.
func GroupOf(msgs []ChatMessage, i int) Group {
	same := func(j int) bool {
		return j >= 0 && j < len(msgs) && msgs[j].Mine == msgs[i].Mine && msgs[j].Author == msgs[i].Author
	}
	return Group{First: !same(i - 1), Last: !same(i + 1)}
}

// Message builds a message with its bubble, name and time. Your own message
// sits on the right, the other side's on the left with its avatar.
//
//	for i, m := range app.msgs {
//		mizan.Message(c.Key(i), m, mizan.GroupOf(app.msgs, i))
//	}
func Message(c *ui.Context, m ChatMessage, g Group) ui.Element {
	t := From(c)
	cl := t.Colors
	row := ui.Row(c).FillWidth().AlignItems(ui.End).Gap(8)
	if m.Mine {
		row.Justify(ui.End)
	}
	if g.Last {
		row.Margin(0, 0, 10, 0)
	}
	row.Children(func() {
		if !m.Mine {
			if g.Last {
				a := m.Avatar
				a.Size = or(a.Size, bubbleAvatarDia)
				// the avatar sits at the bubble's bottom edge, above the time row
				Avatar(c, m.Author, a).Margin(0, 0, 18, 0)
			} else {
				ui.Box(c).Size(bubbleAvatarDia, 0).Shrink(0)
			}
		}
		col := ui.Column(c).Gap(3).MaxWidthPercent(72).MinWidth(0)
		if m.Mine {
			col.AlignItems(ui.End)
		} else {
			col.AlignItems(ui.Start)
		}
		col.Children(func() {
			if g.First && !m.Mine && m.Author != "" {
				ui.Text(c, m.Author).TextColor(cl.Muted).FontSize(FsSm).FontWeight(500).Padding(0, 12)
			}
			b := bubble(c, t, m.Mine, g.Last)
			b.Children(func() {
				if m.Body != nil {
					m.Body()
					return
				}
				ui.Text(c, m.Text).Wrap().FontSize(FsBase).LineHeight(1.4)
			})
			if g.Last && (m.Time != "" || m.Status != DeliveryNone) {
				ui.Row(c).AlignItems(ui.Center).Gap(4).Padding(0, 6).Children(func() {
					ui.Text(c, m.Time).TextColor(cl.Muted).FontSize(FsXs).FontFeatures("tnum").NoWrap()
					switch m.Status {
					case Sending:
						ui.Icon(c, iconPending).TextColor(cl.Muted).Label("Sending")
					case Sent:
						ui.Icon(c, iconSent).TextColor(cl.Muted).Label("Sent")
					case Read:
						ui.Icon(c, iconRead).TextColor(cl.Accent).Label("Read")
					}
				})
			}
		})
	})
	return row
}

// bubble is a bubble's surface. The bottom corner of a group's last bubble
// that faces the speaker is the tail (a small radius).
func bubble(c *ui.Context, t *Tokens, mine, tail bool) ui.Element {
	cl := t.Colors
	b := ui.Column(c).Padding(8, 12).MinWidth(0)
	bl, br := bubbleR, bubbleR
	if tail && mine {
		br = tailR
	} else if tail {
		bl = tailR
	}
	b.Radius(bubbleR, bubbleR, br, bl)
	if mine {
		// same as the solid button: lightness capped at 0.52/0.46, so white text stays legible
		top := capLightness(cl.Accent, 0.52)
		bottom := capLightness(lighten(cl.Accent, -0.06), 0.46)
		b.LinearGradient(ui.LinearGradient{Angle: 180, Oklab: true, From: top, To: bottom}).TextColor(cl.OnAccent)
		t.Shadows.Sheen.With(ShadowLayer{Y: 1, Blur: 2, Color: Mix(cl.Accent, ui.Color{}, 0.55)}).Apply(b)
	} else {
		b.Background(cl.Surface2).Border(1, cl.Hairline).TextColor(cl.Text)
	}
	return b
}

// Typing is the bubble showing that the other side is typing: three dots
// rise and fall in turn.
func Typing(c *ui.Context, author string, o AvatarOptions) ui.Element {
	t := From(c)
	cl := t.Colors
	return ui.Row(c).AlignItems(ui.End).Gap(8).Margin(0, 0, 10, 0).Children(func() {
		o.Size = or(o.Size, bubbleAvatarDia)
		Avatar(c, author, o)
		b := bubble(c, t, false, true).Label(author + " is typing")
		b.Padding(0, 14).Height(36).Justify(ui.Center).Children(func() {
			ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
				for i := range 3 {
					d := ui.Box(c.Key(i)).Size(6, 6).Radius(3).Background(cl.Muted)
					v := d.Keyframes("bob", []ui.Keyframe{
						{At: 0, Value: 0},
						{At: 300 * time.Millisecond, Value: 1, Ease: ui.EaseOut},
						{At: 600 * time.Millisecond, Value: 0, Ease: ui.EaseIn},
						{At: 1200 * time.Millisecond, Value: 0},
					}, ui.KeyframeOptions{Repeat: ui.RepeatForever, Delay: time.Duration(i) * 150 * time.Millisecond})
					d.Translate(0, -3*v).Opacity(0.45 + 0.55*v)
				}
			})
		})
	})
}

// DaySeparator is the day divider between messages: hairlines on both
// sides, a muted label in the middle.
func DaySeparator(c *ui.Context, label string) ui.Element {
	cl := From(c).Colors
	return ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(12).Padding(8, 0, 14, 0).Children(func() {
		ui.Box(c).Grow(1).Height(1).Background(cl.Hairline)
		ui.Text(c, label).TextColor(cl.Muted).FontSize(FsXs).FontWeight(500).NoWrap()
		ui.Box(c).Grow(1).Height(1).Background(cl.Hairline)
	})
}

// ComposerOptions are the options of Composer.
type ComposerOptions struct {
	Placeholder string
	// Attach, if set, is called when the paperclip button on the left is pressed.
	Attach   func()
	Disabled bool
}

// Composer is the field for writing a message: Input's inset well, with a
// solid round send button on the right. Enter or the button sends: it
// returns true in the frame the message is sent; clearing the text is up to
// the caller.
//
//	if mizan.Composer(c.Key("composer"), &app.draft, mizan.ComposerOptions{Placeholder: "Message"}) {
//		app.send()
//	}
func Composer(c *ui.Context, value *string, o ComposerOptions) bool {
	t := From(c)
	m := t.Metrics
	cl := t.Colors
	empty := strings.TrimSpace(*value) == ""
	sent := false
	h := m.FieldH + 8
	box := ui.Row(c).FillWidth().AlignItems(ui.Center).Height(h).Gap(6).Radius(h/2).Padding(0, 4, 0, 6).Cursor(ui.CursorText)
	var field ui.Element
	box.Children(func() {
		if o.Attach != nil {
			if inputButton(c.Key("attach"), t, iconAttach, "Attach file").Margin(0).Clicked() {
				o.Attach()
			}
		} else {
			ui.Box(c).Size(6, 0)
		}
		field = ui.TextInputBase(c, value).Grow(1).MinWidth(0).Label("Message").
			Placeholder(o.Placeholder).PlaceholderColor(cl.Muted).TextColor(cl.Text).FontSize(FsBase).
			Disabled(o.Disabled).FocusRing(false)
		if field.Submitted() && !empty {
			sent = true
		}
		d := h - 8
		send := ui.ButtonBase(c.Key("send")).Label("Send").Center().Size(d, d).Radius(d / 2).Cursor(ui.CursorPointer).
			Disabled(empty || o.Disabled)
		off := empty || o.Disabled
		top := capLightness(cl.Accent, 0.52)
		bottom := capLightness(lighten(cl.Accent, -0.06), 0.46)
		g := ui.LinearGradient{Angle: 180, Oklab: true, From: top, To: bottom}
		sh := t.Shadows.Sheen.With(ShadowLayer{Y: 1, Blur: 2, Color: Mix(cl.Accent, ui.Color{}, 0.55)})
		t.raised(send, g, Mix(bottom, oklch(0, 0, 0, 1), 0.2), sh,
			raisedState{hovered: send.Hovered() && !off, pressed: send.Pressed() && !off, focus: send.FocusVisible(), disabled: off})
		send.TextColor(cl.OnAccent).Children(func() { ui.Icon(c, iconSend) })
		if send.Clicked() && !off {
			sent = true
		}
	})
	t.well(box, field.Focused(), false)
	if o.Disabled {
		box.Opacity(0.4)
	} else if box.Clicked() {
		field.Focus()
	}
	if sent {
		field.Focus()
	}
	return sent
}
