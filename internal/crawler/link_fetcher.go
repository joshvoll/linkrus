package crawler

import (
	"context"
	"io"
	"net/url"
	"strings"

	"github.com/joshvoll/linkrus/internal/pipeline"
)

// linkFetcher definition
type linkFetcher struct {
	urlGetter   URLGetter
	netDetector PrivateNetworkDetector
}

// newLinkFetcher is the private constructor method to get the LinkFetcher struct
func newLinkFetcher(urlGetter URLGetter, netDetector PrivateNetworkDetector) *linkFetcher {
	return &linkFetcher{
		urlGetter:   urlGetter,
		netDetector: netDetector,
	}
}

// Process implementes the pipeline.Payload interface
// Skip the url that point to a file that cannot contain html content.
// never crawl link in private network (e.g. local address), this is a security risk!
// skip payloads for invalid http status code.
// skip payloads for non-html page headers
func (lf *linkFetcher) Process(ctx context.Context, p pipeline.Payload) (pipeline.Payload, error) {
	payload := p.(*crawlerPayload)
	if exclusionRegex.MatchString(payload.URL) {
		return nil, nil
	}
	if isPrivate, err := lf.isPrivate(payload.URL); err != nil || isPrivate {
		return nil, nil
	}
	res, err := lf.urlGetter.Get(payload.URL)
	if err != nil {
		return nil, nil
	}
	_, err = io.Copy(&payload.RawContent, res.Body)
	_ = res.Body.Close()
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, nil
	}
	if contentType := res.Header.Get("Content-Type"); !strings.Contains(contentType, "html") {
		return nil, nil
	}
	return payload, nil
}

// isPrivate check if the network is private or not
// parse the url
func (lf *linkFetcher) isPrivate(URL string) (bool, error) {
	u, err := url.Parse(URL)
	if err != nil {
		return false, err
	}
	return lf.netDetector.IsPrivate(u.Hostname())
}
