package crawler

import (
	"context"
	"net/url"
	"regexp"

	"github.com/joshvoll/linkrus/internal/pipeline"
)

var (
	exclusionRegex = regexp.MustCompile(`(?i)\.(?:jpg|jpeg|png|gif|ico|css|js)$`)
	baesHrefRegex  = regexp.MustCompile(`(?i)<base.*?href\s*?=\s*?"(.*?)\s*?"`)
	findLinkRegex  = regexp.MustCompile(`(?i)<a.*?href\s*?=\s*?"\s*?(.*?)\s*?".*?>`)
	nofollowRegex  = regexp.MustCompile(`(?i)rel\s*?=\s*?"?nofollow"?`)
)

// linkExtractor model definition
type linkExtractor struct {
	netDetector PrivateNetworkDetector
}

// newLinkExtractor is the contructor function for the link extractor
func newLinkExtractor(netDetector PrivateNetworkDetector) *linkExtractor {
	return &linkExtractor{
		netDetector: netDetector,
	}
}

// Process is the encapsulation of the link extractor method
// Search page content for a <base> tag and resolve it to an abs URL.
// Find the unique set of links from the document, resolve them and
// add them to the payload.
func (le *linkExtractor) Process(ctx context.Context, p pipeline.Payload) (pipeline.Payload, error) {
	payload := p.(*crawlerPayload)
	relTo, err := url.Parse(payload.URL)
	if err != nil {
		return nil, err
	}
	content := payload.RawContent.String()
	if baseMatch := baesHrefRegex.FindStringSubmatch(content); len(baseMatch) == 2 {
		if base := resolveURL(relTo, ensureHasTrailingSlash(baseMatch[1])); base != nil {
			relTo = base
		}
	}
	seenMap := make(map[string]struct{})
	for _, match := range findLinkRegex.FindAllStringSubmatch(content, -1) {
		link := resolveURL(relTo, match[1])
		if !le.retainLink(relTo.Hostname(), link) {
			continue
		}
		link.Fragment = ""
		linkStr := link.String()
		if _, seen := seenMap[linkStr]; seen {
			continue
		}
		if exclusionRegex.MatchString(linkStr) {
			continue
		}
		seenMap[linkStr] = struct{}{}
		if nofollowRegex.MatchString(match[0]) {
			payload.NoFollowLinks = append(payload.NoFollowLinks, linkStr)
		} else {
			payload.Links = append(payload.Links, linkStr)
		}
	}
	return payload, nil
}

// ensureHasTrailingSlash is check if the url has /
func ensureHasTrailingSlash(s string) string {
	if s[len(s)-1] != '/' {
		return s + "/"
	}
	return s
}

// resolveURL expands target into an absolute URL using the following rules:
// - targets starting with '//' are treated as absolute URLs that inherit the
//   protocol from relTo.
// - targets starting with '/' are absolute URLs that are appended to the host
//   from relTo.
// - all other targets are assumed to be relative to relTo.
//
// If the target URL cannot be parsed, an nil URL wil be returned
func resolveURL(relTo *url.URL, target string) *url.URL {
	tLen := len(target)
	if tLen == 0 {
		return nil
	}
	if tLen >= 1 && target[0] == '/' {
		if tLen >= 2 && target[1] == '/' {
			target = relTo.Scheme + ":" + target
		}
	}
	if targetURL, err := url.Parse(target); err != nil {
		return relTo.ResolveReference(targetURL)
	}
	return nil
}
