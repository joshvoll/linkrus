package crawler

import (
	"net/url"
	"regexp"
)

var (
	exclusionRegex = regexp.MustCompile(`(?i)\.(?:jpg|jpeg|png|gif|ico|css|js)$`)
	baesHrefRegex  = regexp.MustCompile(`(?i)<base.*?href\s*?=\s*?"(.*?)\s*?"`)
	findLinkRegex  = regexp.MustCompile(`(?i)<a.*?href\s*?=\s*?"\s*?(.*?)\s*?".*?>`)
	nofollowRegex  = regexp.MustCompile(`(?i)rel\s*?=\s*?"?nofollow"?`)
)

// resolveURL expands target into an absolute URL using the following rules:
// - targets starting with '//' are treated as absolute URLs that inherit the
//   protocol from relTo.
// - targets starting with '/' are absolute URLs that are appended to the host
//   from relTo.
// - all other targets are assumed to be relative to relTo.
//
// If the target URL cannot be parsed, an nil URL wil be returned
func resolvedURL(relTo *url.URL, target string) *url.URL {
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
