# Avatar

A person's initials, or a picture. The background is a tint of a color
over the card and the letters a deeper shade of it, legible in both
themes; without a `Color`, one of the series colors is picked from the
name, so the same name has the same color everywhere.

<!-- preview avatar 150 -->

```go
mizan.Avatar(c, "Maya Chen", mizan.AvatarOptions{Online: true})
mizan.Avatar(c, "Ivy Brooks", mizan.AvatarOptions{Solid: true, Size: 40})
```

`Image` shows a picture in place of the initials, `Online` a green dot at
its corner, and `Solid` makes the background the color itself with white
letters. `Size` is the diameter, 32 by default.
