package main

import (
	"context"
	"displateBot/backend"
	"displateBot/displate"
	"displateBot/telegram"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"

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

	b, err := telegram.NewClient(botToken, logger.With("component", "telegramBot"), handleMessage(store, logger))
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

func handleMessage(be backend.Store, logger *slog.Logger) func(context.Context, *bot.Bot, *models.Update) {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		// TODO add logging
		if update.Message != nil {
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
				sendAsBatches(ctx, b, update, photos, logger)
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
		} else if update.InlineQuery != nil {
			startOffset, err := strconv.Atoi(update.InlineQuery.Offset)
			if err != nil {
				logger.
					With("offset", update.InlineQuery.Offset).
					With("error", err).
					Error("failed to parse offset sent by inline query")
				// reset to 0 as a safe starting point
				startOffset = 0
			}

			matches := make([]displate.Displate, 0, telegram.MaxNumResultsPerQueryResponse)
			displates := be.LimitedEditionDisplates()

			// this limits the number of results returned to 50, as expected by the telegram API
			endOffset := min(startOffset+telegram.MaxNumResultsPerQueryResponse, len(displates))
			for _, d := range displates[startOffset:endOffset] {
				lowercaseTitle := strings.ToLower(d.Title)
				lowercaseQuery := strings.ToLower(update.InlineQuery.Query)
				if strings.Contains(lowercaseTitle, lowercaseQuery) {
					matches = append(matches, d)
				}
			}

			photos := make([]models.InlineQueryResult, 0, len(matches))
			for _, match := range matches {
				photo := models.InlineQueryResultPhoto{
					ID:           strconv.Itoa(match.ID),
					PhotoURL:     match.Images.Main.URL,
					ThumbnailURL: match.Images.Main.URL,
					Title:        match.Title,
					Caption:      match.Title,
				}
				photos = append(photos, &photo)
			}
			inlineQueryResponse, err := b.AnswerInlineQuery(ctx, &bot.AnswerInlineQueryParams{
				InlineQueryID: update.InlineQuery.ID,
				Results:       photos,
				NextOffset:    strconv.Itoa(endOffset),
			})

			logger.
				With("response", inlineQueryResponse).
				With("error", err).
				Debug("responded to inline query")
		}
	}
}

func sendAsBatches(ctx context.Context, b *bot.Bot, update *models.Update, photos []models.InputMedia, logger *slog.Logger) {
	for i := 0; i <= len(photos)/telegram.MaxMediaMessageBatchSize; i++ {
		batchStart := i * telegram.MaxMediaMessageBatchSize
		batchEnd := min(len(photos), (i+1)*telegram.MaxMediaMessageBatchSize)

		_, err := b.SendMediaGroup(ctx, &bot.SendMediaGroupParams{
			ChatID:              update.Message.Chat.ID,
			Media:               photos[batchStart:batchEnd],
			DisableNotification: false,
			ProtectContent:      false,
		})
		if err != nil {
			logger.With("err", err).Error("failed to send /available message")
		}
	}
}
