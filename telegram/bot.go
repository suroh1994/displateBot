package telegram

import (
	"displateBot/displateApi"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"log/slog"
	"os"
	"strings"
)

type Bot struct {
	api           *tgbotapi.BotAPI
	displateStore displateApi.Store
	logger        *slog.Logger
	chTerminate   chan int
}

func NewBot(token string, store displateApi.Store, chTerminate chan int) *Bot {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("failed to create Bot: %v", err)
	}

	return &Bot{
		api:           bot,
		displateStore: store,
		logger:        slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		chTerminate:   chTerminate,
	}
}

func (b *Bot) Serve() {
	b.logger.Info("Authorized on account %s (@%s)", b.api.Self.FirstName, b.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case update := <-updates:
			if update.Message != nil { // If we got a message
				b.serveChatMessage(update.Message)
			} else if update.InlineQuery != nil { // If we got an inline query
				b.serveInlineQuery(update.InlineQuery)
			}
		case <-b.chTerminate:
			return
		}
	}
}

func (b *Bot) serveChatMessage(message *tgbotapi.Message) {
	log.Printf("[%s] %s", message.From.UserName, message.Text)

	msg := tgbotapi.NewMessage(message.Chat.ID, message.Text)
	msg.ReplyToMessageID = message.MessageID

	if _, err := b.api.Send(msg); err != nil {
		log.Println(err)
	}
}

func (b *Bot) serveInlineQuery(query *tgbotapi.InlineQuery) {
	b.logger.Info("query", query.Query)

	displates := b.displateStore.GetLimitedEditionDisplates()

	matches := make([]displateApi.Displate, 0, len(displates))
	for _, displate := range displates {
		if strings.Contains(displate.Title, query.Query) {
			matches = append(matches, displate)
		}
	}

	response := make([]interface{}, 0, len(matches))
	for _, matchedDisplate := range matches {
		image := tgbotapi.NewInlineQueryResultPhotoWithThumb(matchedDisplate.Title, matchedDisplate.Images.Main.URL, matchedDisplate.Images.Main.URL)
		image.Width = 560
		image.Height = 784
		image.Caption = query.Query
		response = append(response, image)
	}

	inlineConf := tgbotapi.InlineConfig{
		InlineQueryID: query.ID,
		IsPersonal:    true,
		CacheTime:     0,
		Results:       response,
	}

	if _, err := b.api.Send(inlineConf); err != nil {
		log.Println(err)
	}
}
