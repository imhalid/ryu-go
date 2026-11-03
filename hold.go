package mizan

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// holdState is where a press held on a stepping button is.
type holdState struct {
	down    bool      // pressed in the last frame
	stepped bool      // this press stepped already, so its click does not
	next    time.Time // when it steps again
	n       int       // how many times it repeated
}

// hold makes the button e step as a native stepper does: once as it is
// pressed, then, held, again after 400 ms and faster and faster, down to
// every 40 ms. A click without a press (assistive technology's) steps
// once. A button turned off, at Min or Max, stops.
func hold(c *ui.Context, e ui.Element, off bool, step func()) {
	const delay, first, fastest = 400 * time.Millisecond, 120 * time.Millisecond, 40 * time.Millisecond
	st := ui.Local(e, "hold", func() holdState { return holdState{} })
	clicked := e.Clicked()
	if off || !e.Pressed() {
		if clicked && !st.stepped && !off {
			step()
		}
		*st = holdState{}
		return
	}
	now := c.Now()
	switch {
	case !st.down:
		*st = holdState{down: true, stepped: true, next: now.Add(delay)}
		step()
	case !now.Before(st.next):
		st.n++
		st.next = now.Add(max(fastest, first-time.Duration(st.n)*8*time.Millisecond))
		step()
	}
	c.After(st.next.Sub(now))
}
