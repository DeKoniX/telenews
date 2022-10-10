package parse

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

func (parseNews *ParseNewsStruct) InitVK(secureKey string) {
	parseNews.vk.secureKey = secureKey
}

func (ParseNews ParseNewsStruct) ParseVKWall(query string, retry bool) (vkWallNews []NewsStruct, err error) {
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
						Sizes []struct {
							Height   int    `json:"height"`
							Url      string `json:"url"`
							TypeSize string `json:"type"`
							Width    int    `json:"width"`
						}
					}
				}
			}
		}
		Error struct {
			ErrorCode int    `json:"error_code"`
			ErrorMsg  string `json:"error_msg"`
		}
	}

	var vkjson vkJSON

	uri, _ := url.Parse("https://api.vk.com/method/wall.get")
	q := uri.Query()
	q.Add("domain", query)
	q.Add("count", "5")
	q.Add("filter", "owner")
	q.Add("access_token", ParseNews.vk.secureKey)
	q.Add("v", "5.131")
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

	if vkjson.Error.ErrorCode != 0 {
		return vkWallNews, fmt.Errorf("VK error code: %d error message: %s", vkjson.Error.ErrorCode, vkjson.Error.ErrorMsg)
	}
	if vkjson.Error.ErrorCode == 6 && retry == true {
		time.Sleep(time.Second * time.Duration(rand.Intn(20)+20))
		return ParseNews.ParseVKWall(query, true)
	}

	for _, news := range vkjson.Response.Items {
		link := fmt.Sprintf("https://vk.com/%s?w=wall%v_%v", query, news.OwnerID, news.Id)
		newsStruct := NewsStruct{
			Title: query,
			MSG:   news.Text,
			Link:  link,
		}
		itemHash := newsStruct.GenHash(ParseNews.SourceID)
		vkWallNews = append(vkWallNews, newsStruct)
		for id, attach := range news.Attachments {
			if id != 0 {
				if attach.TypeAttach == "photo" {
					var height int = 0
					var vkWallPhoto NewsStruct
					for _, photo := range attach.Photo.Sizes {
						if height < photo.Height {
							height = photo.Height
							vkWallPhoto = NewsStruct{
								Title: photo.Url,
								Link:  photo.Url,
								Hash:  itemHash,
							}
						}
					}
					vkWallNews = append(vkWallNews, vkWallPhoto)
				}
			}
		}
	}
	return vkWallNews, nil
}
