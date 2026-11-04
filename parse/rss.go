package parse

import (
	"net/http"
	"net/url"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	ext "github.com/mmcdole/gofeed/extensions"
	"golang.org/x/net/html"
)

var FirefoxUserAgent string = "Mozilla/5.0 (X11; Linux x86_64; rv:78.0) Gecko/20100101 Firefox/78.0"

type userAgentTransport struct {
	base      http.RoundTripper
	userAgent string
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	newReq := req.Clone(req.Context())
	newReq.Header.Set("User-Agent", t.userAgent)
	return t.base.RoundTrip(newReq)
}

func HTTPClient(client *http.Client, userAgent string) *http.Client {
	c := *client
	if c.Transport == nil {
		c.Transport = http.DefaultTransport
	}
	c.Transport = &userAgentTransport{base: c.Transport, userAgent: userAgent}
	return &c
}

func (ParseNews ParseNewsStruct) ParseRSS(feedURL string) (rssNews []NewsStruct, err error) {
	httpClient := HTTPClient(&http.Client{Timeout: 15 * time.Second}, FirefoxUserAgent)
	fp := gofeed.NewParser()
	fp.Client = httpClient

	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		return rssNews, err
	}
	for _, news := range feed.Items {
		rssNews = append(rssNews, rssItem(news, feedURL))
		if len(rssNews) >= 5 {
			break
		}
	}
	return rssNews, nil
}

func rssItem(item *gofeed.Item, feedURL string) NewsStruct {
	news := NewsStruct{Title: item.Title, Link: item.Link}
	base, err := url.Parse(feedURL)
	if err != nil {
		return news
	}
	add := func(rawURL, mediaType string) {
		if strings.TrimSpace(rawURL) == "" {
			return
		}
		link, err := base.Parse(strings.TrimSpace(rawURL))
		if err != nil || (link.Scheme != "https" && link.Scheme != "http") || link.Host == "" {
			return
		}
		ext := strings.ToLower(path.Ext(link.Path))
		kind := strings.ToLower(strings.TrimSpace(strings.SplitN(mediaType, ";", 2)[0]))
		switch {
		case strings.HasPrefix(kind, "image/") || kind == "image" ||
			(kind == "" && slices.Contains([]string{".jpg", ".jpeg", ".png", ".webp"}, ext)):
			if !slices.Contains(news.PhotoURL, link.String()) {
				news.PhotoURL = append(news.PhotoURL, link.String())
			}
		case kind == "video/mp4" || ((kind == "" || kind == "video") && ext == ".mp4"):
			if !slices.Contains(news.VideoURL, link.String()) {
				news.VideoURL = append(news.VideoURL, link.String())
			}
		}
	}
	for _, enclosure := range item.Enclosures {
		add(enclosure.URL, enclosure.Type)
	}
	var addMediaRSS func(map[string][]ext.Extension)
	addMediaRSS = func(elements map[string][]ext.Extension) {
		for _, content := range elements["content"] {
			kind := content.Attrs["type"]
			if kind == "" {
				kind = content.Attrs["medium"]
			}
			add(content.Attrs["url"], kind)
		}
		for _, group := range elements["group"] {
			addMediaRSS(group.Children)
		}
	}
	addMediaRSS(item.Extensions["media"])
	if len(news.PhotoURL) == 0 && len(news.VideoURL) == 0 {
		if item.Image != nil {
			add(item.Image.URL, "image")
		}
		for _, thumbnail := range item.Extensions["media"]["thumbnail"] {
			add(thumbnail.Attrs["url"], "image")
		}
	}
	for _, content := range []string{item.Content, item.Description} {
		tokens := html.NewTokenizer(strings.NewReader(content))
		for tokenType := tokens.Next(); tokenType != html.ErrorToken; tokenType = tokens.Next() {
			if tokenType != html.StartTagToken && tokenType != html.SelfClosingTagToken {
				continue
			}
			token := tokens.Token()
			var src, kind string
			for _, attr := range token.Attr {
				switch attr.Key {
				case "src":
					src = attr.Val
				case "type":
					kind = attr.Val
				}
			}
			switch token.Data {
			case "img":
				add(src, "image")
			case "video", "source":
				add(src, kind)
			}
		}
	}
	return news
}
