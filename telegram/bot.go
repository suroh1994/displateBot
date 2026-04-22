package telegram

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"
)

const (
	MaxNumResultsPerQueryResponse = 50
	MaxMediaMessageBatchSize      = 10
)

type Client struct {
	bot    *bot.Bot
	logger *slog.Logger
}

func NewClient(token string, logger *slog.Logger, handlerFunc bot.HandlerFunc) (*Client, error) {
	opts := []bot.Option{
		bot.WithDefaultHandler(handlerFunc),
	}

	b, err := bot.New(token, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Client: %v", err)
	}

	return &Client{
		bot:    b,
		logger: logger,
	}, nil
}

func (c *Client) Serve(ctx context.Context) {
	c.bot.Start(ctx)
}

func (c *Client) SendMessage(ctx context.Context, chatID int64, text string) {
	_, err := c.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
	if err != nil {
		c.logger.Error("failed to send message", "chatID", chatID, "err", err)
	}
}

func (c *Client) SendMediaGroup(ctx context.Context, chatID int64, photos []models.InputMedia) {
	_, err := c.bot.SendMediaGroup(ctx, &bot.SendMediaGroupParams{
		ChatID: chatID,
		Media:  photos,
	})
	if err != nil {
		c.logger.Error("failed to send media group", "chatID", chatID, "err", err)
	}
}
