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
		"expansions":   {"author_id"},
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
			ID       string `json:"id"`
			Text     string `json:"text"`
			AuthorID string `json:"author_id"`
		} `json:"data"`
		Includes struct {
			Users []author `json:"users"`
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
		news = append(news, NewsStruct{
			Title: title,
			MSG:   post.Text,
			Link:  "https://x.com/" + handle + "/status/" + post.ID,
		})
	}
	return news, nil
}
