package main

import (
	"strings"
	"time"

	"github.com/DeKoniX/telenews/models"
	"github.com/DeKoniX/telenews/parse"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const telegramHelpMessage = "" +
	"Бот отправляет последние новости\n" +
	"/start - запускает бот\n" +
	"/stop - останавливает бота и удаляет пользователя\n" +
	"/source - показать список доступных источников\n" +
	"/add - добавляет источник в бота - <code>/add rss https://example.com/feed.xml</code>\n" +
	"/list - показать ваш список слежения\n" +
	"/del - удалить ваш источник - <code>/del rss https://example.com/feed.xml</code>\n" +
	"/help - этот текст\n" +
	"/version - текущая версия бота"

const telegramListSource = "vk_wall, rss"

func (teleNews *teleNewsStruct) telegramInit(token string) (err error) {
	teleNews.bot, err = tgbotapi.NewBotAPI(token)
	return err
}

func (teleNews teleNewsStruct) telegramUpdate() {
	var err error

	updateCfg := tgbotapi.NewUpdate(0)
	updateCfg.Timeout = 60

	updates := teleNews.bot.GetUpdatesChan(updateCfg)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				var user models.User
				user.SelectByChatId(update.Message.Chat.ID)
				if user.UserName == "" {
					user.UserName = update.Message.From.UserName
					user.ChatID = update.Message.Chat.ID
					err = user.Insert()
					if err != nil {
						teleNews.logger.Println("[ERR][DB] Error create user: ", err)
						teleNews.telegramSendMessage(update.Message.Chat.ID, "Произошла ошибка, сообщите о ней пожалуйста admin@dekonix.ru\n"+err.Error(), false)
					} else {
						teleNews.telegramSendMessage(update.Message.Chat.ID, "Добро пожаловать "+update.Message.From.UserName, false)
					}
				}
			case "stop":
				var user models.User
				user.SelectByChatId(update.Message.Chat.ID)
				if user.UserName != "" {
					err = user.Delete()
					if err != nil {
						teleNews.logger.Println("[ERR][DB] Error delete user: ", err)
					}
				}
			case "source":
				sources := telegramListSource
				if teleNews.parser.TwitterEnabled() {
					sources += ", twitter"
				}
				teleNews.telegramSendMessage(update.Message.Chat.ID, "Доступные источники: "+sources, false)
			case "add":
				sourceType, _, _ := strings.Cut(strings.TrimSpace(update.Message.CommandArguments()), " ")
				if sourceType == string(models.Twitter) && !teleNews.parser.TwitterEnabled() {
					teleNews.telegramSendMessage(update.Message.Chat.ID, "Поиск X не настроен администратором бота.", false)
					continue
				}
				if update.Message.CommandArguments() == "" {
					teleNews.telegramSendMessage(update.Message.Chat.ID, "Для добавления источника, ипользуйте синтаксис команды:\n <code>/add source url</code> \n Например: <code>/add vk_wall golang</code>\n/source - покажет список доступных для использования источников", true)
				} else {
					var user models.User
					err = user.SelectByChatId(update.Message.Chat.ID)
					if user.UserName != "" {
						sou, query, err := addSource(update.Message.Chat.ID, update.Message.CommandArguments())
						if err != nil {
							teleNews.telegramSendMessage(update.Message.Chat.ID, "Не смог добавить источник, поправте команду или обратитесь к администратору", false)
						} else {
							teleNews.telegramSendMessage(update.Message.Chat.ID, "Добавлен источник: "+sou+" - "+query, false)
						}
					}
				}
			case "list":
				var user models.User
				user.SelectByChatId(update.Message.Chat.ID)
				if user.UserName != "" {
					message, err := listSource(update.Message.Chat.ID)
					if err != nil {
						teleNews.logger.Println("[ERR][DB] Get list source: ", err)
					} else {
						teleNews.telegramSendMessage(update.Message.Chat.ID, "Добавленные вами источники:\n"+message, false)
					}
				}
			case "del":
				var user models.User
				user.SelectByChatId(update.Message.Chat.ID)
				if user.UserName != "" {
					sou, query, err := deleteSource(update.Message.Chat.ID, update.Message.CommandArguments())
					if err != nil {
						teleNews.telegramSendMessage(update.Message.Chat.ID, "Не смог удалить источник, поправте команду или обратитесь к администратору", false)
					} else {
						teleNews.telegramSendMessage(update.Message.Chat.ID, "Удален источник: "+sou+" - "+query, false)
					}
				}
			case "help":
				teleNews.telegramSendMessage(update.Message.Chat.ID, telegramHelpMessage, true)
			case "version":
				teleNews.telegramSendMessage(update.Message.Chat.ID, "Версия "+version+". Разработчик: DeKoniX (admin@dekonix.ru)", false)
			}
		}

		time.Sleep(time.Second)
	}
}

func (teleNews *teleNewsStruct) telegramSendMessage(chatID int64, text string, modeHTML bool) {
	msg := tgbotapi.NewMessage(chatID, text)
	if modeHTML {
		msg.ParseMode = tgbotapi.ModeHTML
	}
	_, err := teleNews.bot.Send(msg)
	if err != nil {
		teleNews.logger.Println("[ERR][Telegram] Error send message, chat ID: ", chatID, ": ", err, "\nMSG: ", text)
	}
}

func (teleNews *teleNewsStruct) telegramSendPhotos(chatID int64, news parse.NewsStruct) {
	msg := tgbotapi.NewMessage(chatID, news.Title)
	_, err := teleNews.bot.Send(msg)
	if err != nil {
		teleNews.logger.Println("[ERR][Telegram] Error send message, chat ID: ", chatID, ": ", err, "\nMSG: ", news.Title)
	}

	for start := 0; start < len(news.PhotoURL); start += 10 {
		urls := news.PhotoURL[start:min(start+10, len(news.PhotoURL))]
		var err error
		if len(urls) == 1 {
			_, err = teleNews.bot.Send(tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(urls[0])))
		} else {
			media := make([]any, 0, len(urls))
			for _, url := range urls {
				media = append(media, tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(url)))
			}
			_, err = teleNews.bot.SendMediaGroup(tgbotapi.NewMediaGroup(chatID, media))
		}
		if err != nil {
			teleNews.logger.Printf("[ERR][Telegram] Error send photos, chat ID: %d: %s", chatID, err)
		}
	}
}
