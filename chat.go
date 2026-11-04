package ryugo

import (
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

// Ryugo has no chat; its language is set here: your own message wears the
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
	// ID tells the message apart as messages come (its place if unset):
	// Chat animates the new ones in.
	ID     string
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
//		ryugo.Message(c.Key(i), m, ryugo.GroupOf(app.msgs, i))
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
			// the avatar goes with the group's last message: as a new one comes,
			// it fades from the last and grows in on the new
			ui.Box(c.Key("avatar")).Size(bubbleAvatarDia, 0).Shrink(0).Children(func() {
				if g.Last {
					a := m.Avatar
					a.Size = or(a.Size, bubbleAvatarDia)
					// the avatar sits at the bubble's bottom edge, above the time row
					Avatar(c.Key("face"), m.Author, a).Absolute().Left(0).Bottom(18).
						TransformOrigin(0.5, 1).Transition(ui.ElementTransition{Duration: 200 * time.Millisecond, Ease: ui.EaseOut,
						Enter: &ui.Motion{Scale: 0.6}, Exit: &ui.Motion{Scale: 0.8}})
				}
			})
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
			b := bubble(c.Key("bubble"), t, m.Mine, g.Last)
			b.Children(func() {
				if m.Body != nil {
					m.Body()
					return
				}
				ui.Text(c, m.Text).Wrap().FontSize(FsBase).LineHeight(1.4)
			})
			if g.Last && (m.Time != "" || m.Status != DeliveryNone) {
				// the time row goes with the group's last message, folding away
				ui.Row(c.Key("time")).AlignItems(ui.Center).Gap(4).Padding(0, 6).
					Transition(ui.ElementTransition{Duration: 200 * time.Millisecond, Ease: ui.EaseOut, Enter: &ui.Motion{Collapse: true}, Exit: &ui.Motion{Collapse: true}}).
					Children(func() {
						ui.Text(c, m.Time).TextColor(cl.Muted).FontSize(FsXs).FontFeatures("tnum").NoWrap()
						if m.Status != DeliveryNone {
							delivery(c, t, m.Status)
						}
					})
			}
		})
	})
	return row
}

// bubble is a bubble's surface. The bottom corner of a group's last bubble
// that faces the speaker is the tail (a small radius): as a new message
// joins the group, it rounds off, and the new last bubble's sharpens.
func bubble(c *ui.Context, t *Tokens, mine, tail bool) ui.Element {
	cl := t.Colors
	b := ui.Column(c).Padding(8, 12).MinWidth(0)
	k := b.AnimateTo("tail", b2f(tail), ui.Animation{Duration: 220 * time.Millisecond, Ease: ui.EaseOut})
	r := bubbleR + (tailR-bubbleR)*k
	bl, br := bubbleR, bubbleR
	if mine {
		br = r
	} else {
		bl = r
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

// delivery is a message's delivery mark: a clock while it is sent; then a
// check that draws itself as a pen would, and, read, a second one drawn
// after it as the marks take the accent.
func delivery(c *ui.Context, t *Tokens, d Delivery) {
	cl := t.Colors
	e := ui.Box(c.Key("delivery")).Size(16, 13).Shrink(0)
	switch d {
	case Sending:
		e.Label("Sending")
	case Sent:
		e.Label("Sent")
	case Read:
		e.Label("Read")
	}
	pen := ui.CubicBezier(0.65, 0, 0.35, 1)
	one := e.AnimateTo("one", b2f(d >= Sent), ui.Animation{Duration: 260 * time.Millisecond, Ease: pen})
	two := e.AnimateTo("two", b2f(d == Read), ui.Animation{Duration: 260 * time.Millisecond, Ease: pen, Delay: 80 * time.Millisecond})
	read := e.AnimateTo("read", b2f(d == Read), ui.Animation{Duration: 240 * time.Millisecond, Ease: ui.EaseOut})
	color := Mix(cl.Muted, cl.Accent, read)
	e.Children(func() {
		if one < 0.99 {
			ui.Icon(c, iconPending).TextColor(cl.Muted).Opacity(1 - one).Absolute().Left(0).Top(0)
		}
	})
	e.Draw(func(p *ui.Painter, r ui.Rect) {
		// the marks of iconRead, in its 16 by 13 box
		at := func(x, y float32) (float32, float32) { return r.X + x*r.W/16, r.Y + y*r.H/13 }
		if one > 0.001 {
			var a ui.Path
			a.MoveTo(at(1.5+2*(1-two), 8.5))
			a.LineTo(at(4.5+2*(1-two), 11.5))
			a.LineTo(at(10.5+2*(1-two), 4.5))
			a.Trim(0, one)
			p.StrokePath(&a, 1.5, color)
		}
		if two > 0.001 {
			var b ui.Path
			b.MoveTo(at(8, 11))
			b.LineTo(at(8.5, 11.5))
			b.LineTo(at(14.5, 4.5))
			b.Trim(0, two)
			p.StrokePath(&b, 1.5, color)
		}
	})
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
//	if ryugo.Composer(c.Key("composer"), &app.draft, ryugo.ComposerOptions{Placeholder: "Message"}) {
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

var iconDown = Svg(16, 14, "M8 3.5v9M4 8.5l4 4 4-4")

// ChatOptions are the options of Chat.
type ChatOptions struct {
	// Typing, set, is who is typing: their bubble shows under the messages,
	// and as their message comes it folds away as the message grows out of
	// its corner.
	Typing       string
	TypingAvatar AvatarOptions
	// Separator, set, is the label of a day separator before msgs[i]; ""
	// for none.
	Separator func(i int) string
}

// chatState is a Chat's: its scroll, and the messages as the last frame
// had them, to tell the new ones and count those not seen.
type chatState struct {
	scroll  ui.ScrollState
	keys    []any
	unseen  int
	started bool
	// follow: until then the list stays at its end, as what came grows
	follow time.Time
	typing bool // someone was typing in the last frame
}

// chatFollow is how long the list stays at its end as a message grows in.
const chatFollow = 700 * time.Millisecond

// Chat is a conversation: the messages in a list that scrolls, kept at the
// latest. A new message grows in from its bubble's tail corner (the right
// for yours, the left for theirs) as the list makes room, the group's tail
// and avatar passing to it; a message from the one typing grows out of
// where the typing bubble was. Scrolled up, the latest messages don't pull
// the list down: a "Latest" button springs in at the bottom, with the
// number of messages not seen, and takes the list down to them. Give it a
// size: Fill it, or Grow it in its parent.
//
//	ryugo.Chat(c.Key("chat"), app.msgs, ryugo.ChatOptions{Typing: app.typing}).Fill()
func Chat(c *ui.Context, msgs []ChatMessage, o ChatOptions) ui.Element {
	t := From(c)
	cl := t.Colors
	root := ui.Box(c).MinHeight(0).MinWidth(0)
	st := *ui.Local(root, "chat", func() *chatState { return &chatState{} })
	sc := &st.scroll
	keys := make([]any, len(msgs))
	for i, m := range msgs {
		keys[i] = any(i)
		if m.ID != "" {
			keys[i] = m.ID
		}
	}
	// the messages that came since the last frame
	fresh := 0
	if st.started {
		old := map[any]bool{}
		for _, k := range st.keys {
			old[k] = true
		}
		for _, k := range keys {
			if !old[k] {
				fresh++
			}
		}
	}
	// at the latest, or near it: what comes keeps the list there
	atEnd := sc.MaxY-sc.Y < 48
	mine := len(msgs) > 0 && msgs[len(msgs)-1].Mine
	switch {
	case !st.started:
		st.started, st.follow = true, c.Now().Add(chatFollow)
	case (fresh > 0 || o.Typing != "" && !st.typing) && atEnd, fresh > 0 && mine:
		// as the new message (or the typing bubble) grows, the list follows
		// it down
		st.follow, st.unseen = c.Now().Add(chatFollow), 0
	case fresh > 0:
		st.unseen += fresh
	case atEnd:
		st.unseen = 0
	}
	if c.Now().Before(st.follow) && !sc.Scrolling() {
		sc.Y = 1 << 30
		c.AnimationFrame()
	}
	st.keys, st.typing = keys, o.Typing != ""

	enter := ui.ElementTransition{Animation: &springLand, Enter: &ui.Motion{Collapse: true, Scale: 0.82, Y: 6},
		Exit: &ui.Motion{Collapse: true, Scale: 0.9}}
	root.Children(func() {
		ui.Scroll(c).Fill().TrackScroll(sc).Children(func() {
			list := ui.Column(c.Key("list")).FillWidth().Gap(2).Padding(4, 2)
			list.Children(func() {
				for i, m := range msgs {
					if o.Separator != nil {
						if label := o.Separator(i); label != "" {
							ui.Box(c.Key(chatSep{keys[i]})).FillWidth().Transition(ui.ElementTransition{Duration: 200 * time.Millisecond, Enter: &ui.Motion{Collapse: true}}).
								Children(func() { DaySeparator(c, label) })
						}
					}
					// a new message grows from its tail's corner, the list
					// making room for it
					row := ui.Box(c.Key(keys[i])).FillWidth().Transition(enter)
					if m.Mine {
						row.TransformOrigin(1, 1)
					} else {
						row.TransformOrigin(0, 1)
					}
					row.Children(func() { Message(c, m, GroupOf(msgs, i)) })
				}
				if o.Typing != "" {
					// it folds away as the message comes, which grows out of
					// its corner
					ui.Box(c.Key(chatTyping{})).FillWidth().TransformOrigin(0, 1).
						Transition(ui.ElementTransition{Animation: &springLand, Enter: &ui.Motion{Collapse: true, Scale: 0.8},
							Exit: &ui.Motion{Collapse: true, Scale: 0.92}}).
						Children(func() { Typing(c, o.Typing, o.TypingAvatar) })
				}
			})
		})
		// scrolled up, the way down to the latest
		away := sc.MaxY-sc.Y > 120
		shown := root.AnimateTo("latest", b2f(away), ui.Animation{Duration: 140 * time.Millisecond, Ease: ui.EaseOut})
		if away {
			shown = root.AnimateTo("latest", 1, *springPop)
		}
		if shown > 0.01 {
			label := "Latest"
			if st.unseen > 0 {
				label = strconv.Itoa(st.unseen) + " new"
			}
			ui.Row(c).Absolute().Left(0).Right(0).Bottom(10).Justify(ui.Center).PassThrough().Children(func() {
				b := ui.ButtonBase(c.Key("latest")).Label(label).Height(28).Padding(0, 12).Gap(6).AlignItems(ui.Center).
					Radius(RadiusPill).Cursor(ui.CursorPointer).Opacity(min(shown, 1)).
					Translate(0, 10*(1-shown)).Scale(0.9+0.1*shown, 0.9+0.1*shown)
				t.raised(b, t.controlGradient(cl.ControlTop, cl.ControlBottom), cl.ControlBorder, t.Shadows.ShadowPop,
					raisedState{hovered: b.Hovered(), pressed: b.Pressed(), focus: b.FocusVisible()})
				if b.Clicked() {
					sc.ScrollTo(0, 1<<30, ui.Animation{Spring: ui.SpringOptions{Duration: 520 * time.Millisecond}})
					st.unseen = 0
				}
				b.TextColor(cl.Text).Children(func() {
					ui.Icon(c, iconDown).Size(12, 12)
					ui.Text(c, label).FontSize(FsSm).FontWeight(600).FontFeatures("tnum").NoWrap()
				})
			})
		}
	})
	return root
}

type (
	chatSep    struct{ key any }
	chatTyping struct{}
)
