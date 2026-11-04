package main

import "testing"

func TestSplitFirstWord(t *testing.T) {
	for _, tt := range []struct {
		name string
		text string
		word string
		rest string
	}{
		{"regular space", "rss https://www.noob-club.ru/rss2.xml", "rss", "https://www.noob-club.ru/rss2.xml"},
		{"nonbreaking space", "rss\u00a0https://www.noob-club.ru/rss2.xml", "rss", "https://www.noob-club.ru/rss2.xml"},
		{"nonbreaking space after command", "/add\u00a0rss https://www.noob-club.ru/rss2.xml", "/add", "rss https://www.noob-club.ru/rss2.xml"},
		{"narrow nonbreaking space", "rss\u202fhttps://www.noob-club.ru/rss2.xml", "rss", "https://www.noob-club.ru/rss2.xml"},
		{"tab before query", "twitter\tfinal fantasy xiv", "twitter", "final fantasy xiv"},
		{"missing query", "rss\u00a0", "rss", ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			word, rest := splitFirstWord(tt.text)
			if word != tt.word || rest != tt.rest {
				t.Fatalf("splitFirstWord(%q) = (%q, %q), want (%q, %q)", tt.text, word, rest, tt.word, tt.rest)
			}
		})
	}
}
