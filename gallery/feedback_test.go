package gallery

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	mz "github.com/imhalid/ryu-go"
)

// page5 opens the feedback page.
func page5() (*State, *ui.Tester) {
	s := New()
	s.Mode, s.Page = mz.Light, 4
	return s, ui.NewTester(s.View, 1180, 1100)
}

func TestCollapsible(t *testing.T) {
	s, tt := page5()
	if tt.HasText("Verbose logging") {
		t.Fatal("closed, the collapsible shows its body")
	}
	if err := tt.Click("Advanced options"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(300 * time.Millisecond)
	if !s.Feedback.Advanced || !tt.HasText("Verbose logging") {
		t.Fatal("a click did not open the collapsible")
	}
	tt.Click("Verbose logging")
	if !s.Feedback.Verbose {
		t.Fatal("the body's checkbox took no click")
	}
	tt.Click("Advanced options")
	tt.Advance(300 * time.Millisecond)
	if s.Feedback.Advanced || tt.HasText("Verbose logging") {
		t.Fatal("a second click did not close it")
	}
}

// One section of a Single accordion is open at a time.
func TestAccordionSingle(t *testing.T) {
	s, tt := page5()
	f := &s.Feedback
	if !f.Billing || f.Security {
		t.Fatalf("at first: billing %v, security %v", f.Billing, f.Security)
	}
	tt.Click("Security")
	tt.Advance(300 * time.Millisecond)
	if !f.Security || f.Billing {
		t.Fatalf("Security opened: billing %v, security %v", f.Billing, f.Security)
	}
	if !tt.HasText("Two-factor sign-in is on for every member of the team.") {
		t.Fatal("Security's body does not show")
	}
	tt.Click("Security")
	tt.Advance(300 * time.Millisecond)
	if f.Security || f.Billing {
		t.Fatalf("Security closed: billing %v, security %v; want both closed", f.Billing, f.Security)
	}
}

func TestToggle(t *testing.T) {
	s, tt := page5()
	f := &s.Feedback
	tt.Click("Italic")
	if !f.Italic || !f.Bold {
		t.Fatalf("Italic: italic %v, bold %v", f.Italic, f.Bold)
	}
	tt.Click("Bold")
	if f.Bold {
		t.Fatal("a click did not let Bold go")
	}
	// the alignments are exclusive, by the app
	tt.Click("Align center")
	tt.Frame()
	if !f.Center || f.Left {
		t.Fatalf("center: left %v, center %v", f.Left, f.Center)
	}
	// Space turns the focused toggle over
	tt.Key(0, ui.KeySpace)
	tt.Frame()
	if f.Center {
		t.Fatal("Space did not let Align center go")
	}
}

func TestProgressAndSkeleton(t *testing.T) {
	s, tt := page5()
	if !tt.HasText("64%") {
		t.Fatalf("no 64%% in %v", tt.Texts())
	}
	tt.Click("+10%")
	tt.Advance(600 * time.Millisecond)
	if !tt.HasText("74%") {
		t.Fatalf("+10%%: no 74%% in %v", tt.Texts())
	}
	if tt.HasText("Maya Chen") {
		t.Fatal("the placeholders show the content already")
	}
	tt.Click("Load")
	tt.Frame()
	if !s.Feedback.Loaded || !tt.HasText("Maya Chen") {
		t.Fatal("Load did not show the content")
	}
}
