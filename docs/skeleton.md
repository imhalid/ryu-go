# Skeleton

Placeholders where content is still loading: fills of the strong well's
tone, rounded as a field, with a sheen sweeping across them. `SkeletonText`
is a paragraph of bars as tall as a line, the last one shorter. Press Load
to swap them for the content.

<!-- preview skeleton 170 -->

```go
ryugo.Skeleton(c, 36, 36).Radius(ryugo.RadiusPill) // an avatar's
ryugo.Skeleton(c, 120, 12)
ryugo.SkeletonText(c, 2)
```

A width of zero takes the row's width. Round it otherwise with `Radius`;
the sheen keeps to its corners.
