package telegram

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"log/slog"
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
