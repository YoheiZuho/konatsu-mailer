package api

import (
	"sync"

	"github.com/microcosm-cc/bluemonday"
)

var (
	htmlPolicyOnce sync.Once
	htmlPolicy     *bluemonday.Policy
)

// sanitizeHTML strips scripts and other dangerous markup from email HTML on the
// server (defense in depth alongside the client's DOMPurify; design §11).
func sanitizeHTML(html string) string {
	if html == "" {
		return ""
	}
	htmlPolicyOnce.Do(func() {
		// UGCPolicy permits common formatting, links, and images while removing
		// scripts, event handlers, iframes, objects, etc.
		htmlPolicy = bluemonday.UGCPolicy()
		htmlPolicy.AllowAttrs("style").OnElements("span", "div", "p", "td", "table", "font")
		htmlPolicy.AllowAttrs("align").Globally()
	})
	return htmlPolicy.Sanitize(html)
}
