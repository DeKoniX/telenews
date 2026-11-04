package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
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
				case "/bottest/sendPhoto":
					received = append(received, r.Form.Get("photo"))
				case "/bottest/sendMediaGroup":
					var media []struct {
						Type  string `json:"type"`
						Media string `json:"media"`
					}
					if err := json.Unmarshal([]byte(r.Form.Get("media")), &media); err != nil {
						t.Error(err)
					}
					if len(media) < 2 || len(media) > 10 {
						t.Errorf("invalid album size: %d", len(media))
					}
					for _, photo := range media {
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
			app.telegramSendPhotos(42, news)

			// Then: every photo is delivered exactly once, in order.
			if !reflect.DeepEqual(received, news.PhotoURL) {
				t.Fatalf("received %v, want %v", received, news.PhotoURL)
			}
		})
	}
}
