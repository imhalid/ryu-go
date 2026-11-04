# Chat

A conversation, kept at the latest message. A new message grows in from
its bubble's tail corner as the list makes room, the group's tail and
avatar passing to it; a delivery check draws itself, and read, a second one
after it. A message from someone typing grows out of where the typing
bubble was. Send one to see it all: Leo answers.

<!-- preview chat 480 -->

```go
mizan.Chat(c.Key("chat"), app.msgs, mizan.ChatOptions{Typing: app.typing}).Fill()
if mizan.Composer(c.Key("composer"), &app.draft, mizan.ComposerOptions{Placeholder: "Type a message"}) {
	app.send(app.draft)
	app.draft = ""
}
```

A `ChatMessage` has an `Author`, `Mine` for yours, its `Text` (or a `Body`
that builds it), a `Time` and a delivery `Status`: `Sending`, `Sent` or
`Read`. Give each message an `ID` so new ones animate in. `Separator`
returns a day's label to show before a message, and `Typing` names who is
typing. Scrolled up, the list doesn't follow new messages: a Latest button
springs in, with the number not seen. Give the chat a size (`Fill`, or
`Grow` in its parent).

`Message`, `Typing` and `DaySeparator` build the same parts one by one, for
a list of your own.
