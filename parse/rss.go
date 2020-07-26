package parse

import (
	"net/http"

	"github.com/mmcdole/gofeed"
)

var FirefoxUserAgent string = "Mozilla/5.0 (X11; Linux x86_64; rv:78.0) Gecko/20100101 Firefox/78.0"

type userAgentTransport struct {
	base      http.RoundTripper
	userAgent string
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.base = http.DefaultTransport
	newReq := req
	newReq.Header = make(http.Header)
	for k, vv := range req.Header {
		newReq.Header[k] = vv
	}
	newReq.Header.Set("User-Agent", t.userAgent)
	return t.base.RoundTrip(newReq)
}

func HTTPClient(client *http.Client, userAgent string) *http.Client {
	c := *client
	c.Transport = &userAgentTransport{base: c.Transport, userAgent: userAgent}
	return &c
}

func (ParseNews ParseNewsStruct) ParseRSS(url string) (rssNews []NewsStruct, err error) {
	httpClient := HTTPClient(&http.Client{}, FirefoxUserAgent)
	fp := gofeed.NewParser()
	fp.Client = httpClient

	feed, err := fp.ParseURL(url)
	if err != nil {
		return rssNews, err
	}
	for _, news := range feed.Items {
		rssNews = append(rssNews, NewsStruct{
			Title: news.Title,
			Link:  news.Link,
		})
		if len(rssNews) >= 5 {
			break
		}
	}
	return rssNews, nil
}
