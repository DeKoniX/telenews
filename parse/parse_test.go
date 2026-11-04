package parse

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

type testTransport func(*http.Request) (*http.Response, error)

func (f testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestTwitterSearchMapsAuthor(t *testing.T) {
	var parser ParseNewsStruct
	parser.InitTwitter("test-token")
	client := &http.Client{Transport: testTransport(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/2/tweets/search/recent" || req.Header.Get("Authorization") != "Bearer test-token" ||
			req.URL.Query().Get("query") != "(from:golang #go) -is:retweet -is:reply" ||
			req.URL.Query().Get("expansions") != "author_id,attachments.media_keys" {
			t.Fatalf("unexpected X request: %s %s", req.Method, req.URL)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{
			"data":[{"id":"100","text":"A < B","author_id":"42","attachments":{"media_keys":["photo","video"]}}],
			"includes":{"users":[{"id":"42","name":"The Go Team","username":"golang"}],
            "media":[{"media_key":"photo","type":"photo","url":"https://example.com/photo.jpg"},
                     {"media_key":"video","type":"video","variants":[{"content_type":"application/x-mpegURL","url":"https://example.com/live.m3u8"},{"content_type":"video/mp4","bit_rate":100,"url":"https://example.com/clip.mp4"}]}]}
		}`))}, nil
	})}

	news, err := parser.parseTwitter(client, "from:golang #go")

	if err != nil || len(news) != 1 || news[0].Title != "The Go Team" || news[0].MSG != "A < B" ||
		news[0].Link != "https://x.com/golang/status/100" ||
		len(news[0].PhotoURL) != 1 || news[0].PhotoURL[0] != "https://example.com/photo.jpg" ||
		len(news[0].VideoURL) != 1 || news[0].VideoURL[0] != "https://example.com/clip.mp4" {
		t.Fatalf("unexpected X news: %+v, %v", news, err)
	}
}

func TestVKWallIncludesFirstPhoto(t *testing.T) {
	var parser ParseNewsStruct
	parser.InitVK("test-token")
	client := &http.Client{Transport: testTransport(func(req *http.Request) (*http.Response, error) {
		if err := req.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if req.Method != http.MethodPost || req.PostForm.Get("access_token") != "test-token" ||
			req.PostForm.Get("v") != "5.199" || req.URL.RawQuery != "" {
			t.Fatalf("unexpected VK request: %s %s", req.Method, req.URL)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{
			"response":{"items":[{"id":12,"owner_id":-34,"attachments":[{"type":"photo","photo":{"sizes":[
				{"width":100,"height":100,"url":"https://example.com/small.jpg"},
				{"width":200,"height":100,"url":"https://example.com/large.jpg"}
			]}}],
            "copy_history":[{"attachments":[{"type":"video","video":{"owner_id":-34,"id":7,"files":{"mp4_360":"https://example.com/clip.mp4"}}},
              {"type":"video","video":{"owner_id":-34,"id":8}}]}]}]}
		}`))}, nil
	})}

	news, err := parser.parseVKWall(client, "golang")

	if err != nil || len(news) != 1 || len(news[0].PhotoURL) != 1 || news[0].PhotoURL[0] != "https://example.com/large.jpg" ||
		len(news[0].VideoURL) != 1 || news[0].VideoURL[0] != "https://example.com/clip.mp4" ||
		len(news[0].VideoLinks) != 2 || news[0].VideoLinks[1] != "https://vk.com/video-34_8" {
		t.Fatalf("unexpected VK news: %+v, %v", news, err)
	}
}

func TestRSS(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.UserAgent() != FirefoxUserAgent {
			t.Errorf("unexpected User-Agent: %q", req.UserAgent())
		}
		if _, err := io.WriteString(w, `<rss version="2.0"><channel><title>Test</title>
			<item><title>Go news</title><link>https://example.com/news</link><enclosure url="/photo.jpg" type="image/jpeg"/><enclosure url="/clip.mp4" type="video/mp4"/></item>
		</channel></rss>`); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()

	news, err := (ParseNewsStruct{}).ParseRSS(server.URL)

	if err != nil || len(news) != 1 || news[0].Title != "Go news" || len(news[0].PhotoURL) != 1 || news[0].PhotoURL[0] != server.URL+"/photo.jpg" || len(news[0].VideoURL) != 1 || news[0].VideoURL[0] != server.URL+"/clip.mp4" || news[0].Link != "https://example.com/news" {
		t.Fatalf("unexpected RSS news: %+v, %v", news, err)
	}
}

func TestTwitterLive(t *testing.T) {
	token := os.Getenv("X_BEARER_TOKEN")
	if token == "" {
		t.Skip("set X_BEARER_TOKEN to run the live X search")
	}
	var parser ParseNewsStruct
	parser.InitTwitter(token)
	if _, err := parser.ParseTwitter("#golang"); err != nil {
		t.Fatal(err)
	}
}

func TestVKWallLive(t *testing.T) {
	token := os.Getenv("VkSecureKey")
	if token == "" {
		t.Skip("set VkSecureKey to run the live VK wall request")
	}
	var parser ParseNewsStruct
	parser.InitVK(token)
	if _, err := parser.ParseVKWall("golang"); err != nil {
		t.Fatal(err)
	}
}
