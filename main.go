package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/kelseyhightower/envconfig"
	"github.com/soljarka/book-club/config"
	"github.com/soljarka/book-club/hosts"
)

func main() {
	var c config.Config
	err := envconfig.Process("bookclub", &c)
	if err != nil {
		log.Fatal(err.Error())
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{}

	b, err := bot.New(c.BotToken, opts...)
	if err != nil {
		panic(err)
	}

	b.RegisterHandlerMatchFunc(nextHostMatchFunc, nextHostHandler)
	b.RegisterHandlerMatchFunc(nextNthHostMatchFunc, nextNthHostHandler)

	b.Start(ctx)
}

func nextHostMatchFunc(update *models.Update) bool {
	if update.Message == nil {
		return false
	}
	return update.Message.Text == "/next"
}

func nextNthHostMatchFunc(update *models.Update) bool {
	if update.Message == nil {
		return false
	}
	return strings.HasPrefix(update.Message.Text, "/next ")
}

func nextNthHostHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	n, err := strconv.Atoi(strings.TrimPrefix(update.Message.Text, "/next "))

	if err != nil {
		return
	}

	host := hosts.GetNthHost(n)
	message := fmt.Sprintf("Ближайшее %d-е собрание клуба: %s, ведущий: %s", n, host.Date, host.Name)
	if host.Book != "" {
		message += fmt.Sprintf(", книга: %s", host.Book)
	}
	message += "."

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   message,
	})
}

func nextHostHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	host := hosts.GetNthHost(1)
	message := fmt.Sprintf("Cледующее собрание клуба: %s, ведущий: %s", host.Date, host.Name)
	if host.Book != "" {
		message += fmt.Sprintf(", книга: %s", host.Book)
	}
	message += "."

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   message,
	})
}
