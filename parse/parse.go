package parse

import (
	"log"
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
	httpClient *http.Client
	logger     *log.Logger
}

func InitParse() (parseNews *ParseNewsStruct) {
	return parseNews
}
