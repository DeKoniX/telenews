package parse

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func (parseNews *ParseNewsStruct) InitVK(secureKey string) {
	parseNews.vk.secureKey = secureKey
}

func (ParseNews ParseNewsStruct) ParseVKWall(query string) (vkWallNews []NewsStruct, err error) {
	type vkJSON struct {
		Response struct {
			Items []struct {
				Id          int    `json:"id"`
				FromID      int    `json:"from_id"`
				OwnerID     int    `json:"owner_id"`
				Date        int64  `json:"date"`
				Text        string `json:"text"`
				Attachments []struct {
					TypeAttach string `json:"type"`
					Photo      struct {
						Photo1280 string `json:"photo_1280"`
					}
				}
			}
		}
	}

	var vkjson vkJSON

	uri, _ := url.Parse("https://api.vk.com/method/wall.get")
	q := uri.Query()
	q.Add("domain", query)
	q.Add("count", "5")
	q.Add("filter", "owner")
	q.Add("access_token", ParseNews.vk.secureKey)
	q.Add("v", "5.44")
	uri.RawQuery = q.Encode()

	resp, err := http.Get(uri.String())
	if err != nil {
		return vkWallNews, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return vkWallNews, err
	}

	err = json.Unmarshal(body, &vkjson)
	if err != nil {
		return vkWallNews, err
	}

	for _, news := range vkjson.Response.Items {
		link := fmt.Sprintf("https://vk.com/%s?w=wall%v_%v", query, news.OwnerID, news.Id)
		vkWallNews = append(vkWallNews, NewsStruct{
			Title: query,
			MSG:   news.Text,
			Link:  link,
		})
		for id, attach := range news.Attachments {
			if id != 0 {
				if attach.TypeAttach == "photo" {
					if attach.Photo.Photo1280 != "" {
						vkWallNews = append(vkWallNews, NewsStruct{
							Title: attach.Photo.Photo1280,
							Link:  attach.Photo.Photo1280,
						})
					}
				}
			}
		}
	}
	return vkWallNews, nil
}
