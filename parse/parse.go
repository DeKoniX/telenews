package parse

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type ParseNewsStruct struct {
	SourceID uint
	twitter  struct {
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
	Hash  string
}

func (newsStruct *NewsStruct) GenHash(sourceID uint) (_ string) {
	h := md5.New()
	io.WriteString(h, strconv.Itoa(int(sourceID)))
	io.WriteString(h, newsStruct.Title)
	io.WriteString(h, newsStruct.MSG)
	io.WriteString(h, newsStruct.Link)
	newsStruct.Hash = fmt.Sprintf("%x", h.Sum(nil))

	return newsStruct.Hash
}
