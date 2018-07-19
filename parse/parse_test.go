package parse

import (
	"fmt"
	"os"
	"testing"
)

func TestTwitter(t *testing.T) {
	const query = "#golang"
	const lang = "ru"

	var parseNews ParseNewsStruct

	var consumerKey = os.Getenv("TwitterConsumerKey")
	var consumerSecret = os.Getenv("TwitterConsumerSecret")
	var token = os.Getenv("TwitterToken")
	var tokenSecret = os.Getenv("TwitterTokenSecret")

	parseNews.InitTwitter(consumerKey, consumerSecret, token, tokenSecret)
	twitterNews, err := parseNews.ParseTwitter(query, lang)
	if err != nil {
		t.Error("[ERR] Error parse Twitter: ", err)
	}
	for _, news := range twitterNews {
		fmt.Printf("\nTitle: %s\nMSG: %s\nLink: %s\n", news.Title, news.MSG, news.Link)
	}
}

func TestRSS(t *testing.T) {
	const parseURL = "https://habrahabr.ru/rss/feed/posts/all/6b25bce297b3816483def22b2404be8e/?with_hubs=true?with_hubs=true"

	var parseNews ParseNewsStruct

	rssNews, err := parseNews.ParseRSS(parseURL)
	if err != nil {
		t.Error("[ERR] Error parse RSS: ", err)
	}
	for _, news := range rssNews {
		fmt.Printf("\nTitle: %s\nMSG: %s\nLink: %s\n", news.Title, news.MSG, news.Link)
	}
}

func TestVkWall(t *testing.T) {
	const query = "golang"

	var parseNews ParseNewsStruct
	parseNews.vk.secureKey = os.Getenv("VkSecureKey")

	vkWallNews, err := parseNews.ParseVKWall(query)
	if err != nil {
		t.Error("[ERR] Error parse VkWall: ", err)
	}

	for _, news := range vkWallNews {
		fmt.Printf("\nTitle: %s\nMSG: %s\nLink: %s\n", news.Title, news.MSG, news.Link)
	}
}
