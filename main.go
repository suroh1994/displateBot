package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"log/slog"
	"os"
	"strings"
)

const (
	BotTokenEnvKey = "TELEGRAM_BOT_TOKEN"
)

var botToken string

func init() {
	botToken = os.Getenv(BotTokenEnvKey)
}

func main() {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("failed to create Bot: %v", err)
	}

	log.Printf("Authorized on account %s (@%s)", bot.Self.FirstName, bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	displateClient := NewDisplateClient()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	for update := range updates {
		updateLogger := logger.With("username", update.SentFrom().UserName)
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.ReplyToMessageID = update.Message.MessageID

			if _, err := bot.Send(msg); err != nil {
				log.Println(err)
			}
		} else if update.InlineQuery != nil { // If we got an inline query
			updateLogger.Info("query", update.InlineQuery.Query)

			displates, err := displateClient.GetAllLimitedEditionDisplates()
			if err != nil {
				break
			}

			matches := make([]Displate, 0, len(displates))
			for _, displate := range displates {
				if strings.Contains(displate.Title, update.InlineQuery.Query) {
					matches = append(matches, displate)
				}
			}

			response := make([]interface{}, 0, len(matches))
			for _, matchedDisplate := range matches {
				image := tgbotapi.NewInlineQueryResultPhotoWithThumb(matchedDisplate.Title, matchedDisplate.Images.Main.URL, matchedDisplate.Images.Main.URL)
				image.Width = 560
				image.Height = 784
				image.Caption = update.InlineQuery.Query
				response = append(response, image)
			}

			inlineConf := tgbotapi.InlineConfig{
				InlineQueryID: update.InlineQuery.ID,
				IsPersonal:    true,
				CacheTime:     0,
				Results:       response,
			}

			if _, err = bot.Send(inlineConf); err != nil {
				log.Println(err)
			}

		}
	}
}
