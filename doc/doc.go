// Package doc embeds repoview's manual pages (embedded/*.md), browsed via Actions ▸ Docs.
// Add a page by dropping in a numbered .md: the first "# " heading is its title and the
// line under it its description.
package doc

import (
	"embed"

	"github.com/brohd11/bubblestack/components"
)

//go:embed embedded/*.md
var pagesFS embed.FS

// Pages returns the embedded manual pages in filename order.
func Pages() []components.DocPage { return components.ParseDocPages(pagesFS, "embedded") }
