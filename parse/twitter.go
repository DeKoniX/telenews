package parse

import (
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

func (ParseNews *ParseNewsStruct) InitTwitter(consumerKey, consumerSecret, token, tokenSecret string) {
	ParseNews.twitter.config = oauth1.NewConfig(consumerKey, consumerSecret)
	ParseNews.twitter.token = oauth1.NewToken(token, tokenSecret)
	ParseNews.httpClient = ParseNews.twitter.config.Client(oauth1.NoContext, ParseNews.twitter.token)

	ParseNews.twitter.client = twitter.NewClient(ParseNews.httpClient)
}

//func (ParseNews *ParseNewsStruct) parseTwitter() {
//	for _, searchTweet := range ParseNews.config.List.Twitter {
//		search, _, err := ParseNews.twitter.client.Search.Tweets(&twitter.SearchTweetParams{Query: searchTweet, Lang: "ru", Count: 10})
//		if err != nil {
//			ParseNews.logger.Println("[ERR] Ошибка запроса поиска Twitter: ", err)
//			return
//		}
//
//		for _, tweet := range search.Statuses {
//			twDate, _ := time.Parse(time.RubyDate, tweet.CreatedAt)
//			if ParseNews.testFeed(tweet.IDStr, tweet.Text, twDate) {
//				if tweet.Retweeted == false {
//					link := fmt.Sprintf("https://twitter.com/%s/status/%s\n", tweet.User.ScreenName, tweet.IDStr)
//					_, err = ParseNews.dataBase.InsertInfo(tweet.IDStr, tweet.Text, twDate)
//					if err != nil {
//						ParseNews.logger.Println("[ERR] Ошибка добаления новости в БД ", err)
//						return
//					} else {
//						err = ParseNews.sendMSG(tweet.User.Name, tweet.Text, link)
//						if err != nil {
//							ParseNews.logger.Println("[ERR] Ошибка отправления сообщения в TG ", err)
//							return
//						}
//					}
//
//				}
//			}
//		}
//	}
//}
