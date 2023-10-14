package main

import (
	"context"
	"displateBot/backend"
	"displateBot/displate"
	"displateBot/telegram"
	"log/slog"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	botTokenEnvKey = "TELEGRAM_BOT_TOKEN"
)

var botToken string

func init() {
	botToken = os.Getenv(botTokenEnvKey)
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	displateClient := displate.NewClient(logger.With("component", "displateClient"))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	store := backend.NewStore(logger.With("component", "backend"))
	go store.UpdateDatabase(displateClient, ctx)

	b, err := telegram.NewClient(botToken, logger.With("component", "telegramBot"), handleMessage(store))
	if err != nil {
		logger.Error("failed to initialize telegram client", "err", err)
		return
	}
	go b.Serve(ctx)

	// ToDo implement graceful shutdown
	sigchan := make(chan os.Signal)
	signal.Notify(sigchan, os.Interrupt)
	select {
	case <-sigchan:
		return
	}
}

func handleMessage(be backend.Store) func(context.Context, *bot.Bot, *models.Update) {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		// TODO add logging
		switch update.Message.Text {
		case "/available":
			photos := make([]models.InputMedia, 0)
			for _, availableDisplate := range be.AvailableDisplates() {
				photo := models.InputMediaPhoto{
					Media:   availableDisplate.Images.Main.URL,
					Caption: availableDisplate.Title,
				}
				photos = append(photos, &photo)
			}
			b.SendMediaGroup(ctx, &bot.SendMediaGroupParams{
				ChatID:              update.Message.Chat.ID,
				Media:               photos,
				DisableNotification: false,
				ProtectContent:      false,
			})
		case "/upcoming":
			photos := make([]models.InputMedia, 0)
			for _, availableDisplate := range be.UpcomingDisplates() {
				photo := models.InputMediaPhoto{
					Media:   availableDisplate.Images.Main.URL,
					Caption: availableDisplate.Title,
				}
				photos = append(photos, &photo)
			}
			b.SendMediaGroup(ctx, &bot.SendMediaGroupParams{
				ChatID:              update.Message.Chat.ID,
				Media:               photos,
				DisableNotification: false,
				ProtectContent:      false,
			})
		case "/help":
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "This bot currently supports two commands: /available and /upcoming.",
			})
		default:
			// TODO log error message? or return a help message?
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Sorry, I don't know how to handle this message. Please try /help for a list of valid messages.",
			})

		}
	}
}
