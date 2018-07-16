package parse

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func (ParseNews ParseNewsStruct) parseVKWall(query string) (vkWallNews []newsStruct, err error) {
	type vkJSON struct {
		Response struct {
			Items []struct {
				Id      int    `json:"id"`
				FromID  int    `json:"from_id"`
				OwnerID int    `json:"owner_id"`
				Date    int64  `json:"date"`
				Text    string `json:"text"`
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
		vkWallNews = append(vkWallNews, newsStruct{
			Title: query,
			MSG:   news.Text,
			Link:  link,
		})
	}
	return vkWallNews, nil
}
