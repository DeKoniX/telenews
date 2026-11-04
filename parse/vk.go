package parse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (parseNews *ParseNewsStruct) InitVK(secureKey string) {
	parseNews.vk.secureKey = secureKey
}

var vkHTTPClient = &http.Client{Timeout: 15 * time.Second}

func (p ParseNewsStruct) ParseVKWall(query string) ([]NewsStruct, error) {
	return p.parseVKWall(vkHTTPClient, query)
}

func (p ParseNewsStruct) parseVKWall(client *http.Client, query string) ([]NewsStruct, error) {
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

	params := url.Values{
		"domain":       {query},
		"count":        {"5"},
		"filter":       {"owner"},
		"access_token": {p.vk.secureKey},
		"v":            {"5.199"},
	}
	req, err := http.NewRequest(http.MethodPost, "https://api.vk.com/method/wall.get", strings.NewReader(params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create VK wall request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request VK wall: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("VK wall returned HTTP %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(&vkjson); err != nil {
		return nil, fmt.Errorf("decode VK wall response: %w", err)
	}
	if vkjson.Error.ErrorCode != 0 {
		return nil, fmt.Errorf("VK error %d: %s", vkjson.Error.ErrorCode, vkjson.Error.ErrorMsg)
	}

	vkWallNews := make([]NewsStruct, 0, len(vkjson.Response.Items))
	for _, news := range vkjson.Response.Items {
		link := fmt.Sprintf("https://vk.com/%s?w=wall%v_%v", query, news.OwnerID, news.Id)
		newsStruct := NewsStruct{
			Title: query,
			MSG:   news.Text,
			Link:  link,
		}
		for _, attach := range news.Attachments {
			if attach.TypeAttach != "photo" {
				continue
			}
			var photoURL string
			var area int
			for _, photo := range attach.Photo.Sizes {
				if photo.Url != "" && (photoURL == "" || photo.Width*photo.Height > area) {
					area = photo.Width * photo.Height
					photoURL = photo.Url
				}
			}
			if photoURL != "" {
				newsStruct.PhotoURL = append(newsStruct.PhotoURL, photoURL)
			}
		}
		vkWallNews = append(vkWallNews, newsStruct)
	}

	return vkWallNews, nil
}
