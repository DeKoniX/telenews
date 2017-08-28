package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"time"

	"io/ioutil"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/mmcdole/gofeed"
	"gopkg.in/yaml.v2"
)

const version = "0.0.9"

type TeleNewsStruct struct {
	dataBase DB
	bot      *tgbotapi.BotAPI
	twitter  struct {
		config *oauth1.Config
		token  *oauth1.Token
		client *twitter.Client
	}
	httpClient *http.Client
	config     ConfigStruct
	logger     *log.Logger
}

type ConfigStruct struct {
	Telegram struct {
		Token string
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
	List struct {
		Rss     []string
		Vk      []string
		Twitter []string
	}
}

func main() {
	var TeleNews TeleNewsStruct

	var err error
	var logFile *os.File

	logFile, err = os.Create("telenews.log")
	if err != nil {
		fmt.Println("[ERR][LOG] Ошибка создания/чтения лог файла: ", err)
		os.Exit(0)
	}
	TeleNews.logger = log.New(logFile, "TeleNews: ", log.LstdFlags)

	TeleNews.config, err = getConfig("telenews.yml")
	if err != nil {
		TeleNews.logger.Println("[ERR][CFG] Ошибка чтения конфигурации: ", err)
		os.Exit(0)
	}
	TeleNews.logger.Printf("%+v\n", TeleNews.config)

	TeleNews.dataBase, err = dbInit()
	if err != nil {
		TeleNews.logger.Println("[ERR][DB] Ошибка чтения БД: ", err)
		os.Exit(0)
	}

	TeleNews.twitter.config = oauth1.NewConfig(TeleNews.config.Twitter.ConsumerKey, TeleNews.config.Twitter.ConsumerSecret)
	TeleNews.twitter.token = oauth1.NewToken(TeleNews.config.Twitter.Token, TeleNews.config.Twitter.TokenSecret)
	TeleNews.httpClient = TeleNews.twitter.config.Client(oauth1.NoContext, TeleNews.twitter.token)

	TeleNews.twitter.client = twitter.NewClient(TeleNews.httpClient)

	TeleNews.bot, err = tgbotapi.NewBotAPI(TeleNews.config.Telegram.Token)
	if err != nil {
		TeleNews.logger.Println("[ERR][Telegram] Ошибка подключения к боту: ", err)
		os.Exit(0)
	}

	TeleNews.bot.Debug = true

	fmt.Println("Начал работу!")

	go TeleNews.tgUpdate()
	TeleNews.workNews()
}

func (TeleNews *TeleNewsStruct) tgUpdate() {
	var updateCfg tgbotapi.UpdateConfig
	var updates tgbotapi.UpdatesChannel
	var err error

	updateCfg = tgbotapi.NewUpdate(0)
	updateCfg.Timeout = 60

	updates, err = TeleNews.bot.GetUpdatesChan(updateCfg)
	if err != nil {
		TeleNews.logger.Println("[ERR][Telegram] Ошибка обновления TG бота: ", err)
		os.Exit(0)
	}

	for update := range updates {
		if update.Message.Text == "/start" {
			_, err = TeleNews.dataBase.InsertUser(update.Message.From.UserName, update.Message.Chat.ID)
			if err != nil {
				TeleNews.logger.Println("[ERR][DB] Ошибка добавления пользователя: ", err)
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пользователь "+update.Message.From.UserName+" добавлен для рассылки")
			_, err = TeleNews.bot.Send(msg)
			if err != nil {
				TeleNews.logger.Println("[ERR][Telegram] Ошибка отправления сообщения пользователю ", update.Message.From.UserName, " чат ID ", update.Message.Chat.ID, ": ", err)
			}
		}

		if update.Message.Text == "/stop" {
			err = TeleNews.dataBase.DeleteUser(update.Message.Chat.ID)
			if err != nil {
				TeleNews.logger.Println("[ERR][DB] Ошибка удаления пользователя: ", err)
			}
		}

		if update.Message.Text == "/help" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Бот отправляет последние новости на русском языке\n/start - Запускает пост новостей\n/stop - Останавливает бота\n/help - Этот текст\n/version - Текущая версия бота\n")
			msg.ParseMode = "markdown"
			TeleNews.bot.Send(msg)
		}

		if update.Message.Text == "/version" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Версия "+version+". Разработчик: DeKoniX (admin@dekonix.ru)")
			TeleNews.bot.Send(msg)
		}

		time.Sleep(time.Second * 2)
	}
}

func (TeleNews *TeleNewsStruct) workNews() {
	var timeNow time.Time
	isWork := false

	TeleNews.parseRSS()
	TeleNews.parseTwitter()
	TeleNews.parseVK()
	time.Sleep(time.Minute * 1)
	for {
		timeNow = time.Now()
		if timeNow.Minute()%5 == 0 {
			if isWork == false {
				TeleNews.parseRSS()
				TeleNews.parseTwitter()
				TeleNews.parseVK()
				isWork = true
				time.Sleep(time.Minute * 1)
			}
		} else {
			isWork = false
		}
		time.Sleep(time.Second * 30)
	}
}

func (TeleNews *TeleNewsStruct) parseVK() {
	// config,logger
	type vkJSON struct {
		Response struct {
			Items []struct {
				Id      int    `json:"id"`
				FromID  int    `json:"from_id"`
				OwnerID int    `json:"owner_id"`
				Date    int64  `json:"date"`
				Text    string `json:"text"`
			}
		}
	}

	var vkjson vkJSON

	for _, groupVkName := range TeleNews.config.List.Vk {
		uri, _ := url.Parse("https://api.vk.com/method/wall.get")
		q := uri.Query()
		q.Add("domain", groupVkName)
		q.Add("count", "5")
		q.Add("filter", "owner")
		q.Add("access_token", TeleNews.config.Vk.SecureKey)
		q.Add("v", "5.44")
		uri.RawQuery = q.Encode()

		resp, err := http.Get(uri.String())
		if err != nil {
			TeleNews.logger.Println("[ERR] Ошибка получения информации от VK: ", err)
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			TeleNews.logger.Println("[ERR] Ошибка чтения информации от VK: ", err)
			return
		}

		err = json.Unmarshal(body, &vkjson)
		if err != nil {
			TeleNews.logger.Println("[ERR] Ошибка чтения информации от VK: ", err)
			return
		}

		for _, postVk := range vkjson.Response.Items {
			postVkDate := time.Unix(postVk.Date, 0)
			if TeleNews.testFeed(postVk.Text, postVkDate) {
				link := fmt.Sprintf("https://vk.com/%s?w=wall%v_%v", groupVkName, postVk.OwnerID, postVk.Id)
				err = TeleNews.sendMSG(groupVkName, postVk.Text, link)
				if err != nil {
					TeleNews.logger.Println("[ERR] Ошибка отправления сообщения в TG ", err)
				} else {
					_, err = TeleNews.dataBase.InsertInfo(postVk.Text, postVkDate)
					if err != nil {
						TeleNews.logger.Println("[ERR] Ошибка добаления новости в БД ", err)
					}
				}

			}
		}
	}
}

func (TeleNews *TeleNewsStruct) parseTwitter() {
	for _, searchTweet := range TeleNews.config.List.Twitter {
		search, _, err := TeleNews.twitter.client.Search.Tweets(&twitter.SearchTweetParams{Query: searchTweet, Lang: "ru", Count: 10})
		if err != nil {
			TeleNews.logger.Println("[ERR] Ошибка запроса поиска Twitter: ", err)
			return
		}

		for _, tweet := range search.Statuses {
			twDate, _ := time.Parse(time.RubyDate, tweet.CreatedAt)
			if TeleNews.testFeed(tweet.Text, twDate) {
				if tweet.Retweeted == false {
					link := fmt.Sprintf("https://twitter.com/%s/status/%s\n", tweet.User.ScreenName, tweet.IDStr)
					_, err = TeleNews.dataBase.InsertInfo(tweet.Text, twDate)
					if err != nil {
						TeleNews.logger.Println("[ERR] Ошибка добаления новости в БД ", err)
						return
					} else {
						err = TeleNews.sendMSG(tweet.User.Name, tweet.Text, link)
						if err != nil {
							TeleNews.logger.Println("[ERR] Ошибка отправления сообщения в TG ", err)
							return
						}
					}

				}
			}
		}
	}
}

func (TeleNews *TeleNewsStruct) parseRSS() {
	var fp *gofeed.Parser
	var feed *gofeed.Feed
	var err error

	for _, uri := range TeleNews.config.List.Rss {
		fp = gofeed.NewParser()
		feed, err = fp.ParseURL(uri)
		if err != nil {
			TeleNews.logger.Println("[ERR] Ошибка чтения RSS ленты ", uri, ": ", err)
			return
		}

		for _, item := range feed.Items {
			if TeleNews.testFeed(item.Title, *item.PublishedParsed) == true {
				_, err = TeleNews.dataBase.InsertInfo(item.Title, *item.PublishedParsed)
				if err != nil {
					TeleNews.logger.Println("[ERR] Ошибка добаления новости в БД ", err)
					return
				} else {
					err = TeleNews.sendMSG(feed.Title, item.Title, item.Link)
					if err != nil {
						TeleNews.logger.Println("[ERR] Ошибка отправления сообщения в TG ", err)
						return
					}
				}
			}
		}
	}
}

func (TeleNews *TeleNewsStruct) testFeed(title string, date time.Time) bool {
	hash := GenHash(title, date)
	rows, err := TeleNews.dataBase.SelectInfo(hash)
	if err != nil {
		return false
	}

	if len(rows) == 0 {
		timeNow := time.Now()
		timeD := timeNow.Add(-(time.Hour * 24 * 3))
		if timeD.After(date) {
			return false
		}

		return true
	}

	return false
}

func (TeleNews *TeleNewsStruct) sendMSG(ch string, title string, link string) error {
	// database, bot, logger
	users, err := TeleNews.dataBase.SelectUsers()
	if err != nil {
		return err
	}

	for _, user := range users {
		msg := tgbotapi.NewMessage(user.ChatID, "*"+ch+"*\n"+title+"\n["+link+"]("+link+")")
		msg.ParseMode = "markdown"
		TeleNews.bot.Send(msg)
		TeleNews.logger.Println("MSG: " + user.Username + " -> " + "*" + ch + "*\n" + title + "\n" + link)
	}

	return nil
}

func getConfig(configPath string) (config ConfigStruct, err error) {
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
