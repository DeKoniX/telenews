package parse

//func (ParseNews *ParseNewsStruct) parseRSS() {
//	var fp *gofeed.Parser
//	var feed *gofeed.Feed
//	var err error
//
//	for _, uri := range ParseNews.config.List.Rss {
//		fp = gofeed.NewParser()
//		feed, err = fp.ParseURL(uri)
//		if err != nil {
//			ParseNews.logger.Println("[ERR] Ошибка чтения RSS ленты ", uri, ": ", err)
//			return
//		}
//
//		for _, item := range feed.Items {
//			if ParseNews.testFeed("", item.Title, *item.PublishedParsed) == true {
//				_, err = ParseNews.dataBase.InsertInfo("", item.Title, *item.PublishedParsed)
//				if err != nil {
//					ParseNews.logger.Println("[ERR] Ошибка добаления новости в БД ", err)
//					return
//				} else {
//					err = ParseNews.sendMSG(feed.Title, item.Title, item.Link)
//					if err != nil {
//						ParseNews.logger.Println("[ERR] Ошибка отправления сообщения в TG ", err)
//						return
//					}
//				}
//			}
//		}
//	}
//}
