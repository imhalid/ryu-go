# Status dot

A dot of a state and its text. A live dot spreads outward and fades every
1.8 seconds, which stops when motion is reduced.

<!-- preview status-dot 90 -->

```go
mizan.StatusDot(c, mizan.DotLive, "Live")
mizan.StatusDot(c, mizan.DotDelayed, "Delayed")
mizan.StatusDot(c, mizan.DotOff, "Offline")
```

A [table](table.md)'s `KindStatus` column shows one in each cell.
