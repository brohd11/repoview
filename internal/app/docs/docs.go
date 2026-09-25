// Package docs opens repoview's manual (the doc package) in the shared docs index.
package docs

import (
	"github.com/brohd11/repoview/doc"

	"github.com/brohd11/bubblestack/components"
)

// Pages returns the embedded manual pages in filename order (see repoview/doc).
func Pages() []components.DocPage { return doc.Pages() }
