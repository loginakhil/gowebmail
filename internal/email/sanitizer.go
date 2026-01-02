package email

import (
	"github.com/microcosm-cc/bluemonday"
)

// Sanitizer handles HTML sanitization for safe email rendering
type Sanitizer struct {
	policy *bluemonday.Policy
}

// NewSanitizer creates a new HTML sanitizer
func NewSanitizer() *Sanitizer {
	// Create a policy that allows safe HTML
	p := bluemonday.UGCPolicy()

	// Allow additional safe elements
	p.AllowElements("p", "br", "strong", "em", "u", "s", "del", "ins")
	p.AllowElements("h1", "h2", "h3", "h4", "h5", "h6")
	p.AllowElements("ul", "ol", "li", "blockquote", "pre", "code")
	p.AllowElements("table", "thead", "tbody", "tfoot", "tr", "th", "td")
	p.AllowElements("div", "span", "hr")

	// Allow links with safe attributes
	p.AllowAttrs("href").OnElements("a")
	p.RequireNoReferrerOnLinks(true)
	p.RequireNoFollowOnLinks(true)

	// Allow images but block external resources
	p.AllowAttrs("alt", "title").OnElements("img")
	p.AllowDataURIImages()

	// Allow safe styling attributes
	p.AllowAttrs("class").Globally()
	p.AllowAttrs("style").OnElements("p", "div", "span", "td", "th")

	// Allow table attributes
	p.AllowAttrs("colspan", "rowspan").OnElements("td", "th")

	return &Sanitizer{policy: p}
}

// Sanitize sanitizes HTML content
func (s *Sanitizer) Sanitize(html string) string {
	return s.policy.Sanitize(html)
}

// SanitizeBytes sanitizes HTML content from bytes
func (s *Sanitizer) SanitizeBytes(html []byte) []byte {
	return s.policy.SanitizeBytes(html)
}
