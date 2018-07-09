package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/DeKoniX/telenews/db"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jinzhu/gorm"
)

const version = "2.0.0"

type teleNewsStruct struct {
	bot    *tgbotapi.BotAPI
	db     *gorm.DB
	logger *log.Logger
	config *configStruct
}

func main() {
	var TeleNews teleNewsStruct
	var err error

	logFile, err := os.Create("telenews.log")
	if err != nil {
		fmt.Println("[ERR][LOG] Ошибка создания/чтения лог файла: ", err)
		os.Exit(0)
	}
	defer logFile.Close()
	TeleNews.logger = log.New(logFile, "TeleNews: ", log.LstdFlags)

	TeleNews.config, err = getConfig("telenews.yml")
	if err != nil {
		TeleNews.logger.Println("[ERR][CFG] Ошибка чтения конфигурации: ", err)
		os.Exit(0)
	}
	TeleNews.logger.Printf("%+v\n", TeleNews.config)

	TeleNews.db, err = db.Init(
		TeleNews.config.DB.Host,
		TeleNews.config.DB.Port,
		TeleNews.config.DB.UserName,
		TeleNews.config.DB.Password,
		TeleNews.config.DB.DBName,
	)
	if err != nil {
		TeleNews.logger.Println("[ERR][DB] Ошибка чтения БД: ", err)
		os.Exit(0)
	}
	defer TeleNews.db.Close()

	TeleNews.bot, err = tgbotapi.NewBotAPI(TeleNews.config.Telegram.Token)
	if err != nil {
		TeleNews.logger.Println("[ERR][Telegram] Ошибка подключения к боту: ", err)
		os.Exit(0)
	}

	TeleNews.bot.Debug = true

	fmt.Println("Начал работу!")

	go TeleNews.tgUpdate()
	for {
		time.Sleep(time.Second * 30)
	}
	//TeleNews.workNews()
}

func (TeleNews *teleNewsStruct) tgUpdate() {
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
			var user db.User
			user.UserName = update.Message.From.UserName
			user.ChatID = update.Message.Chat.ID
			TeleNews.db.Create(&user)
			if err != nil {
				TeleNews.logger.Println("[ERR][DB] Ошибка добавления пользователя: ", err)
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пользователь "+update.Message.From.UserName+" добавлен")
			_, err = TeleNews.bot.Send(msg)
			if err != nil {
				TeleNews.logger.Println("[ERR][Telegram] Ошибка отправления сообщения пользователю ", update.Message.From.UserName, " чат ID ", update.Message.Chat.ID, ": ", err)
			}
		}

		if update.Message.Text == "/stop" {
			TeleNews.db.Where("chat_id = ?", update.Message.Chat.ID).Delete(&db.User{})
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

//func (TeleNews *TeleNewsStruct) workNews() {
//	var timeNow time.Time
//	isWork := false
//
//	TeleNews.parseRSS()
//	TeleNews.parseTwitter()
//	TeleNews.parseVK()
//	time.Sleep(time.Minute * 1)
//	for {
//		timeNow = time.Now()
//		if timeNow.Minute()%5 == 0 {
//			if isWork == false {
//				TeleNews.parseRSS()
//				TeleNews.parseTwitter()
//				TeleNews.parseVK()
//				isWork = true
//				time.Sleep(time.Minute * 1)
//			}
//		} else {
//			isWork = false
//		}
//		time.Sleep(time.Second * 30)
//	}
//}
//
//func (TeleNews *TeleNewsStruct) testFeed(id, title string, date time.Time) bool {
//	var hash string
//
//	if id == "" {
//		hash = GenHash("", title, date)
//	} else {
//		hash = GenHash(id, "", time.Now())
//	}
//	rows, err := TeleNews.dataBase.SelectInfo(hash)
//	if err != nil {
//		return false
//	}
//
//	if len(rows) == 0 {
//		timeNow := time.Now()
//		timeD := timeNow.Add(-(time.Hour * 24 * 3))
//		if timeD.After(date) {
//			return false
//		}
//
//		return true
//	}
//
//	return false
//}
//
//func (TeleNews *TeleNewsStruct) sendMSG(ch string, title string, link string) error {
//	// database, bot, logger
//	users, err := TeleNews.dataBase.SelectUsers()
//	if err != nil {
//		return err
//	}
//
//	for _, user := range users {
//		msg := tgbotapi.NewMessage(user.ChatID, "*"+ch+"*\n"+title+"\n["+link+"]("+link+")")
//		msg.ParseMode = "markdown"
//		TeleNews.bot.Send(msg)
//		TeleNews.logger.Println("MSG: " + user.Username + " -> " + "*" + ch + "*\n" + title + "\n" + link)
//	}
//
//	return nil
//}
