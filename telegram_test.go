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
