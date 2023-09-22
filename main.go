package main

import (
	"displateBot/displateApi"
	"displateBot/telegram"
	"fmt"
	"os"
	"time"
)

const (
	BotTokenEnvKey = "TELEGRAM_BOT_TOKEN"
)

var botToken string

func init() {
	botToken = os.Getenv(BotTokenEnvKey)
}

func main() {
	backend := displateApi.NewBackend()
	fetcher := displateApi.NewFetcher()

	chTerminate := make(chan int)
	go periodicallyUpdateDatabase(fetcher, backend, chTerminate)

	bot := telegram.NewBot(botToken, backend, chTerminate)
	go bot.Serve()

	// ToDo implement graceful shutdown
	select {}
}

// ToDo move to store/backend?
func periodicallyUpdateDatabase(fetcher displateApi.Fetcher, db displateApi.Database, chTerminate chan int) {
	ticker := time.NewTicker(time.Second * 5)

	for {
		select {
		case tick := <-ticker.C:
			fmt.Printf("fetching at %s\n", tick.Format(time.RFC3339))

			// fetch data
			displates, err := fetcher.GetLimitedEditionDisplates()
			if err != nil {
				fmt.Printf("failed to get displates: %v\n", err)
				continue
			}
			// write to db
			err = db.StoreDisplates(displates)
			if err != nil {
				fmt.Printf("failed to store displates: %v\n", err)
				continue
			}
		case <-chTerminate:
			return
		}
	}
}
