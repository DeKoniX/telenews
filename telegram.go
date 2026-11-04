package main

import (
	"encoding/json"
	"errors"
	"html"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"

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
	var requestErr *url.Error
	if errors.As(err, &requestErr) {
		return requestErr.Err
	}
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

type newsMedia struct {
	url   string
	video bool
}

func (teleNews *teleNewsStruct) telegramSendNews(chatID int64, news parse.NewsStruct) {
	title := strings.TrimSpace(news.Title)
	body := strings.TrimSpace(news.MSG)
	var lines, visibleLines []string
	if title != "" {
		lines = append(lines, "<b>"+html.EscapeString(title)+"</b>")
		visibleLines = append(visibleLines, title)
	}
	if body != "" && body != title {
		lines = append(lines, html.EscapeString(body))
		visibleLines = append(visibleLines, body)
	}
	for _, link := range news.VideoLinks {
		lines = append(lines, `<a href="`+html.EscapeString(link)+`">Видео</a>`)
		visibleLines = append(visibleLines, "Видео")
	}
	if news.Link != "" {
		lines = append(lines, `<a href="`+html.EscapeString(news.Link)+`">Источник</a>`)
		visibleLines = append(visibleLines, "Источник")
	}
	separator := "\n"
	if body != "" && body != title {
		separator = "\n\n"
	}
	caption := strings.Join(lines, separator)
	if news.IsRSS {
		preview, _ := json.Marshal(struct {
			IsDisabled bool   `json:"is_disabled"`
			URL        string `json:"url"`
		}{URL: news.Link})
		_, err := teleNews.bot.MakeRequest("sendMessage", tgbotapi.Params{
			"chat_id": strconv.FormatInt(chatID, 10), "text": caption,
			"parse_mode": tgbotapi.ModeHTML, "link_preview_options": string(preview),
		})
		if err != nil {
			teleNews.logger.Printf("[ERR][Telegram] Error send RSS text, chat ID: %d: %s", chatID, err)
		}
		// RSS uses a single text message; Telegram supplies the article preview.
		return
	}

	var media []newsMedia
	for _, url := range news.PhotoURL {
		media = append(media, newsMedia{url: url})
	}
	for _, url := range news.VideoURL {
		media = append(media, newsMedia{url: url, video: true})
	}
	if len(media) == 0 {
		if caption != "" {
			teleNews.telegramSendMessage(chatID, caption, true)
		}
		return
	}

	// Measure visible text, excluding HTML markup and escaped entities.
	visible := strings.Join(visibleLines, separator)
	if caption != "" && len(utf16.Encode([]rune(visible))) > 1024 {
		teleNews.telegramSendMessage(chatID, caption, true)
		caption = ""
	}
	for start := 0; start < len(media); start += 10 {
		batch := media[start:min(start+10, len(media))]
		if len(batch) > 1 {
			album := make([]any, 0, len(batch))
			for i, item := range batch {
				file := tgbotapi.FileURL(item.url)
				if item.video {
					video := tgbotapi.NewInputMediaVideo(file)
					video.SupportsStreaming = true
					if i == 0 {
						video.Caption, video.ParseMode = caption, tgbotapi.ModeHTML
					}
					album = append(album, video)
				} else {
					photo := tgbotapi.NewInputMediaPhoto(file)
					if i == 0 {
						photo.Caption, photo.ParseMode = caption, tgbotapi.ModeHTML
					}
					album = append(album, photo)
				}
			}
			_, err := teleNews.bot.SendMediaGroup(tgbotapi.NewMediaGroup(chatID, album))
			if err == nil {
				caption = ""
				continue
			}
			teleNews.logger.Printf("[ERR][Telegram] Error send album, chat ID: %d: %s", chatID, err)
			var apiErr *tgbotapi.Error
			if !errors.As(err, &apiErr) || apiErr.Code != 400 {
				return
			}
		}

		for _, item := range batch {
			var err error
			file := tgbotapi.FileURL(item.url)
			if item.video {
				video := tgbotapi.NewVideo(chatID, file)
				video.Caption, video.ParseMode = caption, tgbotapi.ModeHTML
				video.SupportsStreaming = true
				_, err = teleNews.bot.Send(video)
			} else {
				photo := tgbotapi.NewPhoto(chatID, file)
				photo.Caption, photo.ParseMode = caption, tgbotapi.ModeHTML
				_, err = teleNews.bot.Send(photo)
			}
			if err == nil {
				caption = ""
				continue
			}
			teleNews.logger.Printf("[ERR][Telegram] Error send media, chat ID: %d: %s", chatID, err)
			var apiErr *tgbotapi.Error
			if !errors.As(err, &apiErr) || apiErr.Code != 400 {
				return
			}
		}
	}
	if caption != "" {
		teleNews.telegramSendMessage(chatID, caption, true)
	}
}
