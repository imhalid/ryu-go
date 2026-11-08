package ryugo

// What an app that owns its panes needs of a Layout: to know them, to add one beside another and to
// take one away, as the user does by a split and a close, but from the app's own code (a page opened
// "to the side"). The Dock shows these changes as it shows the user's: an area opens or folds away.

// PaneIDs are the IDs of the layout's panes, the main window's first, each window's areas in order.
func (l *Layout) PaneIDs() []string {
	var ids []string
	for _, w := range l.Windows {
		walk(w.Root, func(n *Node) {
			for _, p := range n.Panes {
				ids = append(ids, p.ID)
			}
		})
	}
	return ids
}

// FindPane returns the pane id, nil when the layout has none.
func (l *Layout) FindPane(id string) *Pane {
	if a, i := l.paneAt(id); a != nil {
		return a.Panes[i]
	}
	return nil
}

// AddPane puts pane in an area of its own beside the area that holds the pane beside, which keeps half
// of its room: after it for DockRow and DockColumn alike. Beside a pane the layout does not have, the
// main window's first area is split. It reports whether the pane went in. Not a step of Undo: the app
// asked for it, the user did not.
func (l *Layout) AddPane(beside string, dir DockDir, pane *Pane) bool {
	if pane == nil || pane.ID == "" || l.FindPane(pane.ID) != nil {
		return false
	}
	a, _ := l.paneAt(beside)
	if a == nil {
		if len(l.Windows) == 0 {
			return false
		}
		walk(l.Windows[0].Root, func(n *Node) {
			if a == nil && n.IsArea() {
				a = n
			}
		})
	}
	if a == nil {
		return false
	}
	l.beside(a, pane, dir, false, 0.5, 0.5, 0)
	return true
}

// RemovePane takes pane id out of the layout, and its area with it when it was alone there (a window's
// last area stays, empty). It reports whether the pane was there. Not a step of Undo.
func (l *Layout) RemovePane(id string) bool {
	_, a := l.takePane(id)
	if a == nil {
		return false
	}
	if len(a.Panes) == 0 {
		l.removeArea(a.ID)
	}
	return true
}
