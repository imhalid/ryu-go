package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/mizan-go"
)

// The other side of the chat: a message sent is sent at 0.5 s, Leo types
// from 0.9 s, and replies at 2.4 s, as the message is read.
const (
	chatSent  = 500 * time.Millisecond
	chatType  = 900 * time.Millisecond
	chatReply = 2400 * time.Millisecond
)

func init() {
	msgs := []mz.ChatMessage{
		{Author: "Maya Chen", Text: "Is the weekly report ready?", Time: "09:41"},
		{Author: "Maya Chen", Text: "I'd like to take a look before the meeting.", Time: "09:41"},
		{Mine: true, Text: "Almost; I'm adding the last charts.", Time: "09:43", Status: mz.Read},
		{Author: "Leo Park", Text: "I checked the data sources too, all good 👍", Time: "09:45"},
	}
	draft, sent := "", 0
	var sentAt time.Time
	demo("chat", func(c *ui.Context, t *mz.Tokens) {
		// play the other side as time goes on
		typing := ""
		if !sentAt.IsZero() {
			since := c.Now().Sub(sentAt)
			last := &msgs[len(msgs)-1]
			switch {
			case since >= chatReply:
				for i := range msgs {
					if msgs[i].Mine {
						msgs[i].Status = mz.Read
					}
				}
				msgs = append(msgs, mz.ChatMessage{ID: fmt.Sprint("leo-", sent), Author: "Leo Park", Text: "Got it, thanks!", Time: "09:46"})
				sentAt = time.Time{}
			case since >= chatSent && last.Mine && last.Status == mz.Sending:
				last.Status = mz.Sent
			}
			if !sentAt.IsZero() {
				if since >= chatType {
					typing = "Leo Park"
				}
				for _, d := range []time.Duration{chatSent, chatType, chatReply} {
					if since < d {
						c.After(d - since)
						break
					}
				}
			}
		}
		ui.Column(c).Gap(8).Width(440).Grow(1).MinHeight(0).Children(func() {
			ui.Box(c).FillWidth().Grow(1).MinHeight(0).Children(func() {
				mz.Chat(c.Key("chat"), msgs, mz.ChatOptions{Typing: typing, Separator: func(i int) string {
					if i == 0 {
						return "Today"
					}
					return ""
				}}).Fill()
			})
			if mz.Composer(c.Key("composer"), &draft, mz.ComposerOptions{Placeholder: "Type a message"}) {
				sent++
				msgs = append(msgs, mz.ChatMessage{ID: fmt.Sprint("mine-", sent), Mine: true, Text: strings.TrimSpace(draft), Time: "09:46", Status: mz.Sending})
				draft = ""
				sentAt = c.Now()
			}
		})
	})
}
