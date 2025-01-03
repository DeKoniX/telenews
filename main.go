package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/DeKoniX/telenews/models"
	"github.com/DeKoniX/telenews/parse"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const version = "2.3.0"

type teleNewsStruct struct {
	bot    *tgbotapi.BotAPI
	logger *log.Logger
	config *configStruct
	parser parse.ParseNewsStruct
}

func main() {
	var teleNews teleNewsStruct
	var err error

	// Get config
	teleNews.config, err = getConfig("telenews.yml")
	if err != nil {
		teleNews.logger.Println("[ERR][CFG] Error read config file: ", err)
		os.Exit(1)
	}

	// Create logger
	logFile, err := os.Create(teleNews.config.General.LogDir + "telenews.log")
	if err != nil {
		fmt.Println("[ERR][LOG] Error logging file: ", err)
		os.Exit(1)
	}
	defer logFile.Close()

	teleNews.logger = log.New(logFile, "TeleNews: ", log.LstdFlags)

	// DB Initial
	err = models.Init(
		teleNews.config.DB.Address,
		teleNews.config.DB.UserName,
		teleNews.config.DB.Password,
		teleNews.config.DB.DBName,
	)
	if err != nil {
		teleNews.logger.Println("[ERR][DB] Error connect DB: ", err)
		os.Exit(1)
	}

	// Parse Initial
	teleNews.parser.InitTwitter(
		teleNews.config.Twitter.ConsumerKey,
		teleNews.config.Twitter.ConsumerSecret,
		teleNews.config.Twitter.Token,
		teleNews.config.Twitter.TokenSecret,
	)
	teleNews.parser.InitVK(teleNews.config.Vk.SecureKey)

	err = teleNews.telegramInit(teleNews.config.Telegram.Token)
	if err != nil {
		teleNews.logger.Println("[ERR][Telegram] Initial telegram error: ", err)
		os.Exit(1)
	}

	if teleNews.config.General.Debug {
		teleNews.bot.Debug = true
	}

	go teleNews.telegramUpdate()
	go teleNews.workNews()
	for {
		time.Sleep(time.Second * 30)
	}
}

func (teleNews *teleNewsStruct) workNews() {
	var timeNow time.Time
	isWork := false

	go teleNews.parseNews()
	time.Sleep(time.Minute * 1)
	for {
		timeNow = time.Now()
		if timeNow.Minute()%2 == 0 {
			if !isWork {
				go teleNews.parseNews()
				isWork = true
				time.Sleep(time.Second * 30)
			}
		} else {
			isWork = false
		}
		time.Sleep(time.Second * 20)
	}
}
