// Package docs is repoview's in-TUI manual adapter: the Docs index opened from
// Actions ▸ Docs. The manual's pages live in the repo's doc folder
// (repoview/doc/embedded, exposed by the doc package) so they're easy to find and
// edit; the parse/render/index machinery is shared bubblestack machinery
// (components).
package docs

import (
	"github.com/brohd11/repoview/doc"

	"github.com/brohd11/bubblestack/components"
)

// Pages returns the embedded manual pages in filename order (see repoview/doc).
func Pages() []components.DocPage { return doc.Pages() }

// Index is the docs menu: one self-dispatching row per page, each pushing its own reader.
func Index() *components.PickerScreen {
	return components.DocsIndex("Docs", "Docs", Pages())
}
