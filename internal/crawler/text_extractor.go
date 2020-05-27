package crawler

import (
	"context"
	"html"
	"regexp"
	"strings"
	"sync"

	"github.com/joshvoll/linkrus/internal/pipeline"
	"github.com/microcosm-cc/bluemonday"
)

var (
	titleRegex         = regexp.MustCompile(`(?i)<title.*?>(.*?)</title>`)
	repeatedSpaceRegex = regexp.MustCompile(`\s+`)
)

// textExtractor definition
type textExtractor struct {
	policyPool sync.Pool
}

// newTextExtractor constructor method
func newTextExtactor() *textExtractor {
	return &textExtractor{
		policyPool: sync.Pool{
			New: func() interface{} {
				return bluemonday.StrictPolicy()
			},
		},
	}
}

// Process encapsulation of the link extractor method
func (te *textExtractor) Process(ctx context.Context, p pipeline.Payload) (pipeline.Payload, error) {
	payload := p.(*crawlerPayload)
	policy := te.policyPool.Get().(*bluemonday.Policy)
	if titleMatch := titleRegex.FindStringSubmatch(payload.RawContent.String()); len(titleMatch) == 2 {
		payload.Title = strings.TrimSpace(html.UnescapeString(repeatedSpaceRegex.ReplaceAllString(policy.Sanitize(titleMatch[1]), "")))
	}
	payload.TextContext = strings.TrimSpace(html.UnescapeString(repeatedSpaceRegex.ReplaceAllString(
		policy.SanitizeReader(&payload.RawContent).String(), " ",
	)))
	te.policyPool.Put(policy)
	return payload, nil
}
