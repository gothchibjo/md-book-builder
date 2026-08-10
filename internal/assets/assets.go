// Package assets embeds the CSS shipped inside every generated HTML book.
package assets

import _ "embed"

//go:embed github-markdown.css
var GitHubMarkdownCSS string

//go:embed book.css
var BookCSS string

// CSSBundle returns the complete stylesheet for the book.
func CSSBundle() string {
	return GitHubMarkdownCSS + "\n" + BookCSS + "\n"
}
