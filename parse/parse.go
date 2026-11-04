package parse

import (
	"crypto/md5"
	"fmt"
	"io"
	"strconv"
)

type ParseNewsStruct struct {
	twitterBearerToken string
	vk                 struct {
		secureKey string
	}
}

type NewsStruct struct {
	URL        string
	PhotoURL   []string
	VideoURL   []string
	VideoLinks []string
	Title      string
	MSG        string
	Link       string
	Hash       string
	IsRSS      bool
}

type PhotoStruct struct {
	URL string
}

func (newsStruct *NewsStruct) GenHash(sourceID uint) (_ string) {
	h := md5.New()
	io.WriteString(h, strconv.Itoa(int(sourceID)))
	io.WriteString(h, newsStruct.Title)
	// RSS announcements are presentation-only; preserve existing deduplication keys.
	if !newsStruct.IsRSS {
		io.WriteString(h, newsStruct.MSG)
	}
	io.WriteString(h, newsStruct.Link)
	newsStruct.Hash = fmt.Sprintf("%x", h.Sum(nil))

	return newsStruct.Hash
}
