package parse

import "github.com/mmcdole/gofeed"

func (ParseNews ParseNewsStruct) ParseRSS(url string) (rssNews []NewsStruct, err error) {
	fp := gofeed.NewParser()
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
