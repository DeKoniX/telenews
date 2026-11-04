package parse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const twitterSearchURL = "https://api.x.com/2/tweets/search/recent"

var twitterHTTPClient = &http.Client{Timeout: 15 * time.Second}

func (p *ParseNewsStruct) InitTwitter(bearerToken string) {
	p.twitterBearerToken = strings.TrimSpace(bearerToken)
}

func (p ParseNewsStruct) TwitterEnabled() bool {
	return p.twitterBearerToken != ""
}

func (p ParseNewsStruct) ParseTwitter(query string) ([]NewsStruct, error) {
	return p.parseTwitter(twitterHTTPClient, query)
}

func (p ParseNewsStruct) parseTwitter(client *http.Client, query string) ([]NewsStruct, error) {
	params := url.Values{
		"query":        {"(" + query + ") -is:retweet -is:reply"},
		"expansions":   {"author_id,attachments.media_keys"},
		"media.fields": {"type,url,variants"},
		"tweet.fields": {"author_id"},
		"user.fields":  {"name,username"},
		"max_results":  {"10"},
	}
	req, err := http.NewRequest(http.MethodGet, twitterSearchURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("create X search request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.twitterBearerToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request X search: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("X search returned HTTP %d", resp.StatusCode)
	}

	type author struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Username string `json:"username"`
	}
	var result struct {
		Data []struct {
			ID          string `json:"id"`
			Text        string `json:"text"`
			AuthorID    string `json:"author_id"`
			Attachments struct {
				MediaKeys []string `json:"media_keys"`
			}
		} `json:"data"`
		Includes struct {
			Users []author `json:"users"`
			Media []struct {
				Key      string `json:"media_key"`
				Type     string
				URL      string
				Variants []struct {
					URL         string
					BitRate     int    `json:"bit_rate"`
					ContentType string `json:"content_type"`
				}
			}
		} `json:"includes"`
		Errors []struct {
			Title string `json:"title"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode X search response: %w", err)
	}
	if len(result.Data) == 0 && len(result.Errors) > 0 {
		return nil, fmt.Errorf("X search failed: %s", result.Errors[0].Title)
	}

	authors := make(map[string]author, len(result.Includes.Users))
	for _, user := range result.Includes.Users {
		authors[user.ID] = user
	}
	attachments := make(map[string]NewsStruct, len(result.Includes.Media))
	for _, media := range result.Includes.Media {
		item := NewsStruct{}
		if media.Type == "photo" && media.URL != "" {
			item.PhotoURL = []string{media.URL}
		}
		var video string
		var rate int
		for _, variant := range media.Variants {
			if variant.ContentType == "video/mp4" && variant.URL != "" && (video == "" || variant.BitRate < rate) {
				video, rate = variant.URL, variant.BitRate
			}
		}
		if video != "" {
			item.VideoURL = []string{video}
		}
		attachments[media.Key] = item
	}
	news := make([]NewsStruct, 0, len(result.Data))
	for _, post := range result.Data {
		user := authors[post.AuthorID]
		title := user.Name
		if title == "" {
			title = "X"
		}
		handle := user.Username
		if handle == "" {
			handle = "i/web"
		}
		item := NewsStruct{
			Title: title,
			MSG:   post.Text,
			Link:  "https://x.com/" + handle + "/status/" + post.ID,
		}
		for _, key := range post.Attachments.MediaKeys {
			media := attachments[key]
			item.PhotoURL = append(item.PhotoURL, media.PhotoURL...)
			item.VideoURL = append(item.VideoURL, media.VideoURL...)
		}
		news = append(news, item)
	}
	return news, nil
}
