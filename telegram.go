package main

import (
	"time"

	"github.com/DeKoniX/telenews/models"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const telegramHelpMessage = "" +
	"Бот отправляет последние новости\n" +
	"/start - запускает бот\n" +
	"/stop - останавливает бота и удаляет пользователя\n" +
	"/source - показать список доступных источников\n" +
	"/add - добавляет источник в бота - <code>/add twitter #golang</code>\n" +
	"/list - показать ваш список слежения\n" +
	"/del - удалить ваш источник по его id - <code>/del twitter #golang</code>\n" +
	"/help - этот текст\n" +
	"/version - текущая версия бота"

const telegramListSource = "twitter, vk_wall, rss"

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
		if update.Message.Command() == "start" {
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
		} else if update.Message.Command() == "stop" {
			var user models.User
			user.SelectByChatId(update.Message.Chat.ID)
			if user.UserName != "" {
				err = user.Delete()
				if err != nil {
					teleNews.logger.Println("[ERR][DB] Error delete user: ", err)
				}
			}
		} else if update.Message.Command() == "source" {
			teleNews.telegramSendMessage(update.Message.Chat.ID, "Доступный список источников с которыми может работать бот: "+telegramListSource, false)
		} else if update.Message.Command() == "add" {
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
		} else if update.Message.Command() == "list" {
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
		} else if update.Message.Command() == "del" {
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
		} else if update.Message.Command() == "help" {
			teleNews.telegramSendMessage(update.Message.Chat.ID, telegramHelpMessage, true)
		} else if update.Message.Command() == "version" {
			teleNews.telegramSendMessage(update.Message.Chat.ID, "Версия "+version+". Разработчик: DeKoniX (admin@dekonix.ru)", false)
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
