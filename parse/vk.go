package parse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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
	type attachment struct {
		Type  string
		Photo struct {
			Sizes []struct {
				Height int
				URL    string
				Width  int
			}
		}
		Video struct {
			ID      int
			OwnerID int `json:"owner_id"`
			Player  string
			Files   map[string]string
		}
	}
	type post struct {
		ID          int
		OwnerID     int `json:"owner_id"`
		Text        string
		Attachments []attachment
		CopyHistory []post `json:"copy_history"`
	}
	type vkJSON struct {
		Response struct {
			Items []post
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
		link := fmt.Sprintf("https://vk.com/%s?w=wall%v_%v", query, news.OwnerID, news.ID)
		newsStruct := NewsStruct{
			Title: query,
			MSG:   news.Text,
			Link:  link,
		}
		var collectAttachments func(post)
		collectAttachments = func(post post) {
			for _, attach := range post.Attachments {
				switch attach.Type {
				case "photo":
					var photoURL string
					var area int
					for _, photo := range attach.Photo.Sizes {
						if photo.URL != "" && (photoURL == "" || photo.Width*photo.Height > area) {
							area = photo.Width * photo.Height
							photoURL = photo.URL
						}
					}
					if photoURL != "" {
						newsStruct.PhotoURL = append(newsStruct.PhotoURL, photoURL)
					}
				case "video":
					video := attach.Video
					var file string
					var resolution int
					for key, candidate := range video.Files {
						if suffix, ok := strings.CutPrefix(key, "mp4_"); ok {
							height, err := strconv.Atoi(suffix)
							if err == nil && height > 0 && candidate != "" && (file == "" || height < resolution) {
								file, resolution = candidate, height
							}
						}
					}
					if file != "" {
						newsStruct.VideoURL = append(newsStruct.VideoURL, file)
					}
					if video.OwnerID != 0 && video.ID != 0 {
						newsStruct.VideoLinks = append(newsStruct.VideoLinks, fmt.Sprintf("https://vk.com/video%d_%d", video.OwnerID, video.ID))
					} else if video.Player != "" {
						newsStruct.VideoLinks = append(newsStruct.VideoLinks, video.Player)
					}
				}
			}
			for _, copy := range post.CopyHistory {
				collectAttachments(copy)
			}
		}
		collectAttachments(news)
		vkWallNews = append(vkWallNews, newsStruct)
	}

	return vkWallNews, nil
}
