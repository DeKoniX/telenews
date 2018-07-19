package parse

import (
	"net/http"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type ParseNewsStruct struct {
	twitter struct {
		config *oauth1.Config
		token  *oauth1.Token
		client *twitter.Client
	}
	vk struct {
		secureKey string
	}
	httpClient *http.Client
}

type NewsStruct struct {
	Title string
	MSG   string
	Link  string
}
