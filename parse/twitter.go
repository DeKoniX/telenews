package parse

import (
	"fmt"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

func (ParseNews *ParseNewsStruct) InitTwitter(consumerKey, consumerSecret, token, tokenSecret string) {
	ParseNews.twitter.config = oauth1.NewConfig(consumerKey, consumerSecret)
	ParseNews.twitter.token = oauth1.NewToken(token, tokenSecret)
	ParseNews.httpClient = ParseNews.twitter.config.Client(oauth1.NoContext, ParseNews.twitter.token)

	ParseNews.twitter.client = twitter.NewClient(ParseNews.httpClient)
}

func (ParseNews ParseNewsStruct) ParseTwitter(query, lang string) (twitterNews []NewsStruct, err error) {
	search, _, err := ParseNews.twitter.client.Search.Tweets(&twitter.SearchTweetParams{Query: query, Count: 10, Lang: lang})
	if err != nil {
		return twitterNews, err
	}

	for _, tweet := range search.Statuses {
		if !tweet.Retweeted && tweet.InReplyToStatusID == 0 {
			link := fmt.Sprintf("https://twitter.com/%s/status/%s\n", tweet.User.ScreenName, tweet.IDStr)
			twitterNews = append(twitterNews, NewsStruct{
				Title: tweet.User.Name,
				MSG:   tweet.Text,
				Link:  link,
			})
		}
	}
	return twitterNews, nil
}
