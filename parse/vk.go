package parse

//func (ParseNews *ParseNewsStruct) parseVK() {
//	// config,logger
//	type vkJSON struct {
//		Response struct {
//			Items []struct {
//				Id      int    `json:"id"`
//				FromID  int    `json:"from_id"`
//				OwnerID int    `json:"owner_id"`
//				Date    int64  `json:"date"`
//				Text    string `json:"text"`
//			}
//		}
//	}
//
//	var vkjson vkJSON
//
//	for _, groupVkName := range ParseNews.config.List.Vk {
//		uri, _ := url.Parse("https://api.vk.com/method/wall.get")
//		q := uri.Query()
//		q.Add("domain", groupVkName)
//		q.Add("count", "5")
//		q.Add("filter", "owner")
//		q.Add("access_token", ParseNews.config.Vk.SecureKey)
//		q.Add("v", "5.44")
//		uri.RawQuery = q.Encode()
//
//		resp, err := http.Get(uri.String())
//		if err != nil {
//			ParseNews.logger.Println("[ERR] Ошибка получения информации от VK: ", err)
//			return
//		}
//		defer resp.Body.Close()
//
//		body, err := ioutil.ReadAll(resp.Body)
//		if err != nil {
//			ParseNews.logger.Println("[ERR] Ошибка чтения информации от VK: ", err)
//			return
//		}
//
//		err = json.Unmarshal(body, &vkjson)
//		if err != nil {
//			ParseNews.logger.Println("[ERR] Ошибка чтения информации от VK: ", err)
//			return
//		}
//
//		for _, postVk := range vkjson.Response.Items {
//			postVkDate := time.Unix(postVk.Date, 0)
//			if ParseNews.testFeed(strconv.Itoa(postVk.Id), postVk.Text, postVkDate) {
//				link := fmt.Sprintf("https://vk.com/%s?w=wall%v_%v", groupVkName, postVk.OwnerID, postVk.Id)
//				err = ParseNews.sendMSG(groupVkName, postVk.Text, link)
//				if err != nil {
//					ParseNews.logger.Println("[ERR] Ошибка отправления сообщения в TG ", err)
//				} else {
//					_, err = ParseNews.dataBase.InsertInfo(strconv.Itoa(postVk.Id), postVk.Text, postVkDate)
//					if err != nil {
//						ParseNews.logger.Println("[ERR] Ошибка добаления новости в БД ", err)
//					}
//				}
//
//			}
//		}
//	}
//}
