package gallery

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/ryu-go"
)

var inbox = mz.Svg(16, 18, "M1.5 9.5h3.5l1 2h4l1-2h3.5M3.5 3.5h9l2 6v3a1 1 0 0 1-1 1h-11a1 1 0 0 1-1-1v-3z")

// more is the second page: chat, drag and drop, and indicators.
func (s *State) more(c *ui.Context, t *mz.Tokens) {
	ui.Row(c).Gap(mz.GapSection).AlignItems(ui.Start).Grow(1).MinHeight(0).Children(func() {
		ui.Column(c).Grow(1).Basis(0).Gap(mz.GapSection).Children(func() {
			s.chat(c, t)
			s.indicators(c, t)
		})
		ui.Column(c).Grow(1).Basis(0).Gap(mz.GapSection).Children(func() {
			s.board(c, t)
			s.files(c, t)
		})
	})
}

func (s *State) chat(c *ui.Context, t *mz.Tokens) {
	mz.Card(c, mz.CardOptions{Title: "Chat", Subtitle: "Your messages in the solid material, theirs on the second surface. Send one: it grows in from its corner, its check draws itself, and a reply grows out of the typing bubble.", Pad: true}, func() {
		ui.Column(c).Gap(8).Children(func() {
			s.converse(c)
			typing := ""
			if s.typing(c) {
				typing = "Leo Park"
			}
			ui.Box(c).FillWidth().Height(380).Children(func() {
				mz.Chat(c.Key("chat"), s.Msgs, mz.ChatOptions{Typing: typing, Separator: func(i int) string {
					if i == 0 {
						return "Today"
					}
					return ""
				}}).Fill()
			})
			if mz.Composer(c.Key("composer"), &s.Draft, mz.ComposerOptions{Placeholder: "Type a message", Attach: func() {}}) {
				s.sent++
				s.Msgs = append(s.Msgs, mz.ChatMessage{ID: fmt.Sprint("mine-", s.sent), Mine: true, Text: strings.TrimSpace(s.Draft), Time: "09:46", Status: mz.Sending})
				s.Draft = ""
				s.sentAt = c.Now()
			}
		})
	})
}

// The other side of the gallery's chat: a message sent is sent at 0.5 s,
// Leo types from 0.9 s, and replies at 2.4 s, as the message is read.
const (
	chatSent  = 500 * time.Millisecond
	chatType  = 900 * time.Millisecond
	chatReply = 2400 * time.Millisecond
)

// converse plays the other side of the chat as time goes on.
func (s *State) converse(c *ui.Context) {
	if s.sentAt.IsZero() {
		return
	}
	since := c.Now().Sub(s.sentAt)
	last := &s.Msgs[len(s.Msgs)-1]
	switch {
	case since >= chatReply:
		for i := range s.Msgs {
			if s.Msgs[i].Mine {
				s.Msgs[i].Status = mz.Read
			}
		}
		s.Msgs = append(s.Msgs, mz.ChatMessage{ID: fmt.Sprint("leo-", s.sent), Author: "Leo Park", Text: "Got it, thanks!", Time: "09:46"})
		s.sentAt = time.Time{}
		return
	case since >= chatSent && last.Mine && last.Status == mz.Sending:
		last.Status = mz.Sent
	}
	for _, d := range []time.Duration{chatSent, chatType, chatReply} {
		if since < d {
			c.After(d - since)
			break
		}
	}
}

// typing reports whether Leo is typing.
func (s *State) typing(c *ui.Context) bool {
	return !s.sentAt.IsZero() && c.Now().Sub(s.sentAt) >= chatType
}

func (s *State) indicators(c *ui.Context, t *mz.Tokens) {
	cl := t.Colors
	mz.Card(c, mz.CardOptions{Title: "Indicators", Subtitle: "Badge, status, key and avatar.", Pad: true}, func() {
		ui.Column(c).Gap(14).Children(func() {
			ui.Row(c).Gap(8).Wrap().AlignItems(ui.Center).Children(func() {
				mz.Badge(c, "+4.2%", mz.ToneDown)
				mz.Badge(c, "−1.8%", mz.ToneUp)
				mz.Badge(c, "Pending", mz.ToneWarn)
				mz.Badge(c, "Draft", mz.ToneNeutral)
				mz.Tag(c, "v2.4")
			})
			ui.Row(c).Gap(18).Wrap().AlignItems(ui.Center).Children(func() {
				mz.StatusDot(c, mz.DotLive, "Live")
				mz.StatusDot(c, mz.DotDelayed, "Delayed")
				mz.StatusDot(c, mz.DotOff, "Offline")
				ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
					ui.Text(c, "Search").TextColor(cl.Muted).FontSize(mz.FsSm)
					mz.Kbd(c, "⌘", "K")
				})
			})
			ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
				for _, n := range []string{"Maya Chen", "Leo Park", "Sam Rivera", "Noah Kim"} {
					mz.Avatar(c.Key(n), n, mz.AvatarOptions{Online: n == "Leo Park"})
				}
				mz.Avatar(c.Key("solid"), "Ivy Brooks", mz.AvatarOptions{Solid: true})
			})
		})
	})
}

func (s *State) board(c *ui.Context, t *mz.Tokens) {
	mz.Card(c, mz.CardOptions{Title: "Drag and drop", Subtitle: "What you drag is raised, where you drop it inset.", Pad: true}, func() {
		ui.Column(c).Gap(16).Children(func() {
			mz.Sortable(c.Key("steps"), &s.steps, &s.Steps, mz.SortableOptions{
				Label: func(i int) string { return s.Steps[i] },
				Key:   func(i int) any { return s.Steps[i] },
			}, func(step string) {
				ui.Text(c, step).Grow(1)
			})
			ui.Row(c).Gap(10).AlignItems(ui.Stretch).Children(func() {
				for col, title := range []string{"To do", "In progress", "Done"} {
					n := 0
					for _, k := range s.Cards {
						if k.col == col {
							n++
						}
					}
					ui.Column(c.Key(col)).Grow(1).Basis(0).MinWidth(0).Children(func() {
						mz.DropArea(c, mz.DropAreaOptions{Title: title, Count: itoa(n), Fill: true}, func() {
							for _, k := range s.Cards {
								if k.col != col {
									continue
								}
								mz.DragCard(c.Key(k.Title), k, func() {
									ui.Text(c, k.Title).Wrap()
									ui.Row(c).Children(func() { mz.Badge(c, k.Tag, mz.ToneNeutral) })
								})
							}
						}, func(k *card, index int) { s.moveCard(k, col, index) })
					})
				}
			})
		})
	})
}

// moveCard puts k in column col, before the index-th of the column's
// other cards (at its end past them).
func (s *State) moveCard(k *card, col, index int) {
	rest := slices.DeleteFunc(slices.Clone(s.Cards), func(x *card) bool { return x == k })
	at, n := len(rest), 0
	for i, x := range rest {
		if x.col != col {
			continue
		}
		if n == index {
			at = i
			break
		}
		n++
		at = i + 1
	}
	k.col = col
	s.Cards = slices.Insert(rest, at, k)
}

func (s *State) files(c *ui.Context, t *mz.Tokens) {
	mz.Card(c, mz.CardOptions{Title: "Files", Subtitle: "Drag them from Finder or Explorer.", Pad: true}, func() {
		ui.Column(c).Gap(8).Children(func() {
			for _, p := range mz.DropZone(c.Key("drop"), mz.DropZoneOptions{Title: "Drop files here", Hint: "PDF, XLSX · up to 10 MB"}) {
				s.Files = append(s.Files, file{Name: p[strings.LastIndexAny(p, `/\`)+1:]})
			}
			for i := 0; i < len(s.Files); i++ {
				if mz.FileItem(c.Key(s.Files[i].Name), s.Files[i].Name, s.Files[i].Size) {
					s.Files = append(s.Files[:i], s.Files[i+1:]...)
					i--
				}
			}
			if len(s.Files) == 0 {
				mz.EmptyState(c, mz.EmptyStateOptions{Icon: inbox, Title: "No files", Description: "Files you add are listed here."})
			}
		})
	})
}
