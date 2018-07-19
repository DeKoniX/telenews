package main

import (
	"io/ioutil"

	"errors"
	"strings"

	"fmt"

	"github.com/DeKoniX/telenews/models"
	"github.com/DeKoniX/telenews/parse"
	"gopkg.in/yaml.v1"
)

type configStruct struct {
	DB struct {
		Address  string `yaml:"address"`
		UserName string `yaml:"username"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
	}
	Telegram struct {
		Token string `yaml:"token"`
	}
	Twitter struct {
		ConsumerKey    string `yaml:"consumerKey"`
		ConsumerSecret string `yaml:"consumerSecret"`
		Token          string `yaml:"token"`
		TokenSecret    string `yaml:"tokenSecret"`
	}
	Vk struct {
		SecureKey string `yaml:"secureKey"`
	}
}

func getConfig(configPath string) (config *configStruct, err error) {
	dat, err := ioutil.ReadFile(configPath)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(dat, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

func addSource(chatID int64, message string) (sou, query string, _ error) {
	var user models.User
	var source models.Source

	user.SelectByChatId(chatID)
	messageSplit := strings.Split(message, " ")
	if len(messageSplit) < 3 {
		return sou, query, errors.New("Wrong message :" + message)
	}
	switch messageSplit[1] {
	case "twitter":
		source.Type = models.Twitter
	case "vk_wall":
		source.Type = models.VKWall
	case "rss":
		source.Type = models.RSS
	default:
		return sou, query, errors.New("No source type")
	}

	source.Query = messageSplit[2]
	_, err := source.Insert(user)

	return messageSplit[1], source.Query, err
}

func listSource(chatID int64) (msg string, err error) {
	var user models.User

	user.SelectByChatId(chatID)
	sources, err := models.Source{}.SelectByUser(user)
	if err != nil {
		return msg, err
	}

	for _, source := range sources {
		msg += fmt.Sprintf("- %s -- %s\n", source.Type, source.Query)
	}

	return msg, nil
}

func deleteSource(chatID int64, message string) (sou, query string, err error) {
	var user models.User
	var source models.Source

	user.SelectByChatId(chatID)
	messageSplit := strings.Split(message, " ")
	if len(messageSplit) < 2 {
		return sou, query, errors.New("Wrong message :" + message)
	}
	switch messageSplit[1] {
	case "twitter":
		source.Type = models.Twitter
	case "vk_wall":
		source.Type = models.VKWall
	case "rss":
		source.Type = models.RSS
	default:
		return sou, query, errors.New("No source type")
	}

	err = source.SelectByQueryAndType(user, messageSplit[2], source.Type)
	if err != nil {
		return sou, query, err
	}
	err = source.Delete()

	return messageSplit[1], source.Query, nil
}

func (teleNews *teleNewsStruct) parseNews() {
	var parseNews []parse.NewsStruct
	sources, err := models.Source{}.SelectAll()
	if err != nil {
		teleNews.logger.Println("[ERR][DB] Error select all")
	}
	for _, source := range sources {
		switch source.Type {
		case models.RSS:
			parseNews, err = teleNews.parser.ParseRSS(source.Query)
			if err != nil {
				teleNews.logger.Println("[ERR][RSS] Error parse RSS: ", err)
			}
		case models.Twitter:
			parseNews, err = teleNews.parser.ParseTwitter(source.Query, "ru")
			if err != nil {
				teleNews.logger.Println("[ERR][TW] Error parse Twitter: ", err)
			}
		case models.VKWall:
			parseNews, err = teleNews.parser.ParseVKWall(source.Query)
			if err != nil {
				teleNews.logger.Println("[ERR][VKW] Error parse VKWall: ", err)
			}
		}
		for _, news := range parseNews {
			var item models.Item
			item.Title = news.Title
			item.Text = news.MSG
			item.Link = news.Link
			_, err = item.Insert(source)
			if err != nil && err != models.ErrAlreadyExists {
				teleNews.logger.Println("[ERR][DB] Error insert item: ", err)
			} else {
				if err != models.ErrAlreadyExists {
					var user models.User
					err = user.SelectById(source.UserID)
					if err != nil {
						teleNews.logger.Println("[ERR][DB] Select User: ", err)
					}
					teleNews.telegramSendMessage(
						user.ChatID,
						"<b>"+item.Title+"</b>\n"+
							item.Text+"\n"+
							"<a href=\""+item.Link+"\">"+item.Title+"</a>",
						true,
					)
				}
			}
		}
	}
}
