package parse

import "github.com/mmcdole/gofeed"

func (ParseNews ParseNewsStruct) parseRSS(url string) (rssNews []newsStruct, err error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)
	if err != nil {
		return rssNews, err
	}
	for _, news := range feed.Items {
		rssNews = append(rssNews, newsStruct{
			Title: news.Title,
			Link:  news.Link,
		})
		if len(rssNews) >= 5 {
			break
		}
	}
	return rssNews, nil
}
