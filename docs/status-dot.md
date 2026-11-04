# Status dot

A dot of a state and its text. A live dot spreads outward and fades every
1.8 seconds, which stops when motion is reduced.

<!-- preview status-dot 90 -->

```go
ryugo.StatusDot(c, ryugo.DotLive, "Live")
ryugo.StatusDot(c, ryugo.DotDelayed, "Delayed")
ryugo.StatusDot(c, ryugo.DotOff, "Offline")
```

A [table](table.md)'s `KindStatus` column shows one in each cell.
