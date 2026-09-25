package app

import (
	"fmt"

	"github.com/brohd11/bubblestack/core"
	"github.com/brohd11/gitstack/repoui"
)

// Header renders the scanned root (with its status marker when it is a checkout) and the
// repo count.
func Header(sh *core.Shared) string {
	c := Of(sh)
	// When the base is a checkout its status marker shares the line, so RootLineValue
	// takes that out of the budget core.HeaderValueWidth reports.
	valWidth := core.HeaderValueWidth(sh.Width(), "Root:  ")
	body := core.Label("Root:  ") + core.Value(repoui.RootLineValue(c.Root, c.RootRepo, valWidth)) + "\n" +
		core.Label("Repos: ") + core.Value(fmt.Sprintf("%d git checkout(s) · depth ≤ %d", len(c.Repos), c.Depth))
	return core.HeaderBox(sh.Width(), body)
}
