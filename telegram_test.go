package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/DeKoniX/telenews/parse"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestTelegramPhotoBatches(t *testing.T) {
	for _, count := range []int{1, 2, 10, 11, 20, 21} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			// Given: a local Bot API endpoint that enforces album limits.
			var received []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := r.ParseForm(); err != nil {
					t.Error(err)
				}
				result := `{}`
				switch r.URL.Path {
				case "/bottest/getMe":
					result = `{"id":1,"is_bot":true,"first_name":"Test"}`
				case "/bottest/sendMessage":
					t.Error("unexpected separate text message")
				case "/bottest/sendPhoto":
					if len(received) == 0 && !strings.Contains(r.Form.Get("caption"), "News") {
						t.Error("missing photo caption")
					}
					received = append(received, r.Form.Get("photo"))
				case "/bottest/sendMediaGroup":
					var media []struct {
						Caption string `json:"caption"`
						Type    string `json:"type"`
						Media   string `json:"media"`
					}
					if err := json.Unmarshal([]byte(r.Form.Get("media")), &media); err != nil {
						t.Error(err)
					}
					if len(media) < 2 || len(media) > 10 {
						t.Errorf("invalid album size: %d", len(media))
					}
					for _, photo := range media {
						if len(received) == 0 && !strings.Contains(photo.Caption, "News") {
							t.Error("missing album caption")
						}
						if photo.Type != "photo" {
							t.Errorf("unexpected media type: %s", photo.Type)
						}
						received = append(received, photo.Media)
					}
					result = `[]`
				default:
					t.Errorf("unexpected endpoint: %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				if _, err := fmt.Fprintf(w, `{"ok":true,"result":%s}`, result); err != nil {
					t.Error(err)
				}
			}))
			defer server.Close()
			bot, err := tgbotapi.NewBotAPIWithClient("test", server.URL+"/bot%s/%s", server.Client())
			if err != nil {
				t.Fatal(err)
			}
			news := parse.NewsStruct{Title: "News"}
			for i := range count {
				news.PhotoURL = append(news.PhotoURL, fmt.Sprintf("https://example.com/%d.jpg", i))
			}
			app := teleNewsStruct{bot: bot, logger: log.New(io.Discard, "", 0)}

			// When
			app.telegramSendNews(42, news)

			// Then: every photo is delivered exactly once, in order.
			if !reflect.DeepEqual(received, news.PhotoURL) {
				t.Fatalf("received %v, want %v", received, news.PhotoURL)
			}
		})
	}
}

func TestRSSAnnouncementDelivery(t *testing.T) {
	for _, count := range []int{0, 1, 2, 11} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			var delivered string
			var calls []string
			var photos int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := r.ParseForm(); err != nil {
					t.Error(err)
				}
				switch r.URL.Path {
				case "/feed":
					io.WriteString(w, `<rss version="2.0"><channel><title>Feed</title><item><title>News</title><link>https://example.com/article</link><description><![CDATA[<p>Игра A &amp; B</p>]]></description>`)
					for i := range count {
						fmt.Fprintf(w, `<enclosure url="https://example.com/%d.jpg" type="image/jpeg"/>`, i)
					}
					io.WriteString(w, `</item></channel></rss>`)
				case "/bottest/getMe":
					io.WriteString(w, `{"ok":true,"result":{"id":1,"is_bot":true}}`)
				case "/bottest/sendMessage":
					calls = append(calls, "text")
					delivered = r.Form.Get("text")
					var preview struct {
						Disabled bool   `json:"is_disabled"`
						URL      string `json:"url"`
					}
					if err := json.Unmarshal([]byte(r.Form.Get("link_preview_options")), &preview); err != nil || preview.Disabled || preview.URL != "https://example.com/article" {
						t.Errorf("missing explicit article preview: %q", r.Form.Get("link_preview_options"))
					}
					io.WriteString(w, `{"ok":true,"result":{}}`)
				case "/bottest/sendPhoto":
					calls = append(calls, "media")
					photos++
					if r.Form.Get("caption") != "" {
						t.Error("RSS photo repeats text")
					}
					io.WriteString(w, `{"ok":true,"result":{}}`)
				case "/bottest/sendMediaGroup":
					calls = append(calls, "media")
					var album []struct {
						Caption string `json:"caption"`
					}
					if err := json.Unmarshal([]byte(r.Form.Get("media")), &album); err != nil {
						t.Error(err)
					}
					for _, item := range album {
						if item.Caption != "" {
							t.Error("RSS album repeats text")
						}
					}
					photos += len(album)
					io.WriteString(w, `{"ok":true,"result":[]}`)
				default:
					t.Errorf("unexpected request: %s", r.URL.Path)
				}
			}))
			defer server.Close()
			news, err := (parse.ParseNewsStruct{}).ParseRSS(server.URL + "/feed")
			if err != nil || len(news) != 1 {
				t.Fatalf("RSS parse: %v, items: %d", err, len(news))
			}
			bot, err := tgbotapi.NewBotAPIWithClient("test", server.URL+"/bot%s/%s", server.Client())
			if err != nil {
				t.Fatal(err)
			}
			app := teleNewsStruct{bot: bot, logger: log.New(io.Discard, "", 0)}
			app.telegramSendNews(42, news[0])
			if want := "<b>News</b>\n\nИгра A &amp; B\n\n<a href=\"https://example.com/article\">Источник</a>"; delivered != want {
				t.Fatalf("delivered %q, want %q", delivered, want)
			}
			if len(calls) != 1 || calls[0] != "text" || photos != 0 {
				t.Fatalf("calls=%v, photos=%d; want one text message and no photos", calls, photos)
			}
			textCalls := 0
			for _, call := range calls {
				if call == "text" {
					textCalls++
				}
			}
			if textCalls != 1 {
				t.Fatalf("sent text %d times", textCalls)
			}
		})
	}
}

func TestRSSFailedTextDoesNotSendMedia(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/bottest/getMe":
			io.WriteString(w, `{"ok":true,"result":{"id":1,"is_bot":true}}`)
		case "/bottest/sendMessage":
			io.WriteString(w, `{"ok":false,"error_code":400,"description":"fixture rejected text"}`)
		default:
			t.Errorf("media sent after text failure: %s", r.URL.Path)
		}
	}))
	defer server.Close()
	bot, err := tgbotapi.NewBotAPIWithClient("test", server.URL+"/bot%s/%s", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	app := teleNewsStruct{bot: bot, logger: log.New(io.Discard, "", 0)}
	app.telegramSendNews(42, parse.NewsStruct{IsRSS: true, Title: "News", MSG: "Summary", Link: "https://example.com/article", PhotoURL: []string{"https://example.com/a.jpg"}})
}

func TestTelegramRejectedMixedAlbumKeepsCaption(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Error(err)
		}
		method := strings.TrimPrefix(r.URL.Path, "/bottest/")
		methods = append(methods, method)
		switch method {
		case "getMe":
			io.WriteString(w, `{"ok":true,"result":{"id":1,"is_bot":true}}`)
			return
		case "sendMediaGroup":
			var album []struct {
				Type    string `json:"type"`
				Caption string `json:"caption"`
			}
			if err := json.Unmarshal([]byte(r.Form.Get("media")), &album); err != nil {
				t.Error(err)
			}
			if len(album) != 2 || album[0].Type != "photo" || album[1].Type != "video" ||
				!strings.Contains(album[0].Caption, "A &lt; B") || album[1].Caption != "" {
				t.Errorf("unexpected album: %+v", album)
			}
			fallthrough
		case "sendPhoto":
			io.WriteString(w, `{"ok":false,"error_code":400,"description":"fixture rejected media"}`)
			return
		case "sendVideo":
			if !strings.Contains(r.Form.Get("caption"), "A &lt; B") || r.Form.Get("parse_mode") != "HTML" {
				t.Error("video did not receive the caption")
			}
		default:
			t.Errorf("unexpected method %q", method)
		}
		io.WriteString(w, `{"ok":true,"result":{}}`)
	}))
	defer server.Close()
	bot, err := tgbotapi.NewBotAPIWithClient("test", server.URL+"/bot%s/%s", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	app := teleNewsStruct{bot: bot, logger: log.New(io.Discard, "", 0)}
	app.telegramSendNews(42, parse.NewsStruct{
		Title: "Test", MSG: "A < B", Link: "https://example.com/news",
		PhotoURL: []string{"https://example.com/missing.jpg"},
		VideoURL: []string{"https://example.com/clip.mp4"},
	})
	if want := []string{"getMe", "sendMediaGroup", "sendPhoto", "sendVideo"}; !reflect.DeepEqual(methods, want) {
		t.Fatalf("methods = %v, want %v", methods, want)
	}
}
