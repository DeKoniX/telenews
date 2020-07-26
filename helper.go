package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/DeKoniX/telenews/models"
	"github.com/DeKoniX/telenews/parse"
	"gopkg.in/yaml.v1"
)

type configStruct struct {
	General struct {
		LogDir string `yaml:"log_dir"`
		Debug  bool   `yaml:"debug"`
	}
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
	if len(messageSplit) < 2 {
		return sou, query, errors.New("Wrong message :" + message)
	}
	switch messageSplit[0] {
	case "twitter":
		source.Type = models.Twitter
	case "vk_wall":
		source.Type = models.VKWall
	case "rss":
		source.Type = models.RSS
	default:
		return sou, query, errors.New("No source type")
	}

	source.Query = messageSplit[1]
	_, err := source.Insert(user)

	return messageSplit[0], source.Query, err
}

func listSource(chatID int64) (msg string, err error) {
	var user models.User

	user.SelectByChatId(chatID)
	sources, err := models.Source{}.SelectByUser(user)
	if err != nil {
		return msg, err
	}

	if len(sources) == 0 {
		msg = "На данный момент источников у вас нет, добавте источники с помощью команды /add"
	} else {
		for _, source := range sources {
			msg += fmt.Sprintf("- %s -- %s\n", source.Type, source.Query)
		}
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
	switch messageSplit[0] {
	case "twitter":
		source.Type = models.Twitter
	case "vk_wall":
		source.Type = models.VKWall
	case "rss":
		source.Type = models.RSS
	default:
		return sou, query, errors.New("No source type")
	}

	err = source.SelectByQueryAndType(user, messageSplit[1], source.Type)
	if err != nil {
		return sou, query, err
	}
	err = source.Delete()

	return messageSplit[0], source.Query, nil
}

func (teleNews *teleNewsStruct) parseNews() {
	// TODO: сделать обработку нового источника, отправка последней новости, и/или сообщение о нормальном работе источника
	var parseNews []parse.NewsStruct
	sources, err := models.Source{}.SelectTryAll()
	if err != nil {
		teleNews.logger.Println("[ERR][DB] Error select all")
	}
	for _, source := range sources {
		firstRun := false
		var item models.Item
		var now time.Time

		_, n, err := item.Select(source)
		if n == 0 {
			firstRun = true
		}

		teleNews.parser.SourceID = source.ID
		switch source.Type {
		case models.RSS:
			parseNews, err = teleNews.parser.ParseRSS(source.Query)
			if err != nil {
				now = time.Now()
				source.NextTryAfter = now.Add(10 * time.Minute)
				source.Error = fmt.Sprintf("[ERR][RSS][%s] Error parse RSS: %s\n", source.Query, err)
				source.Save()
				teleNews.logger.Printf("[ERR][RSS][%s] Error parse RSS: %s\n", source.Query, err)
			}
		case models.Twitter:
			parseNews, err = teleNews.parser.ParseTwitter(source.Query, "ru")
			if err != nil {
				teleNews.logger.Printf("[ERR][TW][%s] Error parse Twitter: %s\n", source.Query, err)
			}
		case models.VKWall:
			parseNews, err = teleNews.parser.ParseVKWall(source.Query, false)
			if err != nil {
				teleNews.logger.Printf("[ERR][VKW][%s] Error parse VKWall: %s\n", source.Query, err)
			}
		}
		for _, news := range parseNews {
			var (
				item     models.Item
				itemTest models.Item
			)

			item.Title = news.Title
			item.Text = news.MSG
			item.Link = news.Link
			if news.Hash == "" {
				news.GenHash(source.ID)
			}
			item.Hash = news.Hash

			err = itemTest.SelectByHash(item.Hash)
			if err == models.ErrRecordNotFound {
				// Logging news
				if teleNews.config.General.Debug {
					teleNews.logger.Println("[DEBUG] News: ", news)
				}

				var user models.User
				err = user.SelectById(source.UserID)
				if err != nil {
					teleNews.logger.Println("[ERR][DB] Select User: ", err)
				}

				if !firstRun {
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
		for _, news := range parseNews {
			var item models.Item

			item.Title = news.Title
			item.Text = news.MSG
			item.Link = news.Link
			if news.Hash == "" {
				news.GenHash(source.ID)
			}
			item.Hash = news.Hash

			_, err = item.Insert(source)
			if err != nil && err != models.ErrAlreadyExists {
				teleNews.logger.Println("[ERR][DB] Error insert item: ", err)
			}
		}
	}
}
