// Package docs is repoview's in-TUI manual: markdown pages compiled into the binary and
// browsed from Actions ▸ Docs. The parse/render/index machinery is shared (it lives in
// bubblestack/components, used by gdaddon too); this package owns only repoview's pages.
//
// Adding a page is dropping a numbered .md into pages/ — no code change. The filename
// orders it, the first "# " heading is its title, and the first line under that heading
// is its one-line description in the index.
package docs

import (
	"embed"

	"github.com/brohd11/bubblestack/components"
)

//go:embed pages/*.md
var pagesFS embed.FS

// Index is the docs menu: one self-dispatching row per page, each pushing its own reader.
func Index() *components.PickerScreen {
	return components.DocsIndex("Docs", "Docs", components.ParseDocPages(pagesFS, "pages"))
}
