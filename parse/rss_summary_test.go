package parse

import (
	"strings"
	"testing"

	"github.com/mmcdole/gofeed"
)

func TestRSSAnnouncement(t *testing.T) {
	for _, tc := range []struct {
		name, description, want string
	}{
		{"plain", "  Две игры   сегодня  ", "Две игры сегодня"},
		{"html", `<p>Игра <b>A</b> &amp; B&nbsp;сегодня.</p><p>Xbox<br>PC</p><script>bad()</script><style>.ad{}</style>`, "Игра A & B сегодня. Xbox PC"},
		{"empty", "", ""},
		{"media only", `<p><img src="/cover.jpg"></p>`, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			feed, err := gofeed.NewParser().Parse(strings.NewReader(`<rss version="2.0" xmlns:content="http://purl.org/rss/1.0/modules/content/"><channel><title>Feed</title><item><title>News</title><link>https://example.com/article</link><description><![CDATA[` + tc.description + `]]></description><content:encoded><![CDATA[Full article must not become an announcement]]></content:encoded></item></channel></rss>`))
			if err != nil {
				t.Fatal(err)
			}
			news := rssItem(feed.Items[0], "https://example.com/feed")
			if news.MSG != tc.want {
				t.Fatalf("announcement = %q, want %q", news.MSG, tc.want)
			}
			// Existing RSS records were hashed with an empty message.
			legacy := NewsStruct{Title: news.Title, Link: news.Link}
			if news.GenHash(42) != legacy.GenHash(42) {
				t.Fatal("adding an announcement changes the existing RSS deduplication key")
			}
		})
	}
}

func TestTextPostsStillHashMessage(t *testing.T) {
	a := NewsStruct{Title: "Post", MSG: "first", Link: "https://example.com/post"}
	b := a
	b.MSG = "edited"
	if a.GenHash(42) == b.GenHash(42) {
		t.Fatal("non-RSS message edits must change the hash")
	}
}
